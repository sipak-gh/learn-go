# Chapter 12 — Command-Line Flags and Configuration

`seek` parses options with a hand-written loop. Every new option means more code, and there are no
`--long-name` spellings or real help text. Go's standard `flag` package does this job. You'll learn to
define flags, use testable `FlagSet`s, write custom flag types and help text, and take defaults from
environment variables. Then `seek` gets a complete command-line interface, including
`--color auto|always|never`.

Previous: [Chapter 11 — I/O and Files](11-io-and-files.md). Next: [Chapter 13 — Generics](13-generics.md).

## The flag package

Define each flag with a name, a default, and a usage string. `flag.Parse()` then reads `os.Args[1:]`
and sets the values.

`examples/ch12/greet/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	// Each call defines a flag and returns a pointer to its value.
	name := flag.String("name", "world", "who to greet")
	times := flag.Int("n", 1, "how many `times` to greet")
	shout := flag.Bool("shout", false, "greet in upper case")
	pause := flag.Duration("pause", 0, "pause between greetings, e.g. 1.5s")

	// Usage is called for -h and for errors. The default prints the
	// program's path, which is a temporary file under go run.
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: greet [flags] [extra ...]")
		flag.PrintDefaults()
	}
	flag.Parse() // reads os.Args[1:]; on error prints usage and exits 2

	msg := "Hello, " + *name + "!"
	if *shout {
		msg = strings.ToUpper(msg)
	}
	for range *times {
		fmt.Println(msg)
	}
	fmt.Println("pause:", *pause)
	fmt.Printf("%d extra args: %q\n", flag.NArg(), flag.Args())
}
```

```console
$ go run ./ch12/greet
Hello, world!
pause: 0s
0 extra args: []
$ go run ./ch12/greet -name=Gopher -n 2 --shout extra args
HELLO, GOPHER!
HELLO, GOPHER!
pause: 0s
2 extra args: ["extra" "args"]
$ go run ./ch12/greet -pause 1.5s -name Ann
Hello, Ann!
pause: 1.5s
0 extra args: []
```

- `flag.String`, `flag.Int`, `flag.Bool`, `flag.Duration`, `flag.Float64`, `flag.Uint`, ... return a
  **pointer**. `Parse` fills it in later, so read it with `*name`.
- `flag.Duration` parses a `time.Duration` like `300ms`, `1.5s`, or `2h45m`: a length of time that
  prints nicely ([Chapter 18](18-context-and-time.md) covers `time`).
- `flag.Args()` returns the *positional* arguments left after the flags; `flag.NArg()` counts them and
  `flag.Arg(i)` returns one.
- The `XxxVar` forms store into an existing variable instead of returning a pointer:
  `flag.StringVar(&cfg.name, "name", "world", "who to greet")`. Use them to fill a struct.

### The syntax the flag package accepts

| You type | Means |
|----------|-------|
| `-name Ann` or `--name Ann` | one dash and two dashes are the same |
| `-name=Ann` | `=` works for every flag |
| `-shout` | a bool flag set to true |
| `-shout=false` | the only way to set a bool flag to false |
| `-n 0x10` | integers accept `0x`, `0o`, and `0b` prefixes |
| `--` | end of flags; everything after it is positional |

Parsing **stops at the first non-flag argument**. Everything after it is positional, even if it
starts with `-`. A lone `-` isn't a flag either, so it can stand for stdin.

```console
$ go run ./ch12/greet Ann -shout
Hello, world!
pause: 0s
2 extra args: ["Ann" "-shout"]
$ go run ./ch12/greet -shout false
HELLO, WORLD!
pause: 0s
1 extra args: ["false"]
```

The second run is a classic surprise: a bool flag never takes the *next* argument as its value, so
`false` became positional. Write `-shout=false`.

### Help and errors

`-h` and `-help` print the usage and exit with status 0. A bad flag prints the error and the usage,
then exits with status 2.

```console
$ go run ./ch12/greet -h
Usage: greet [flags] [extra ...]
  -n times
    	how many times to greet (default 1)
  -name string
    	who to greet (default "world")
  -pause duration
    	pause between greetings, e.g. 1.5s
  -shout
    	greet in upper case
$ go run ./ch12/greet -n two
invalid value "two" for flag -n: parse error
Usage: greet [flags] [extra ...]
  -n times
    	how many times to greet (default 1)
  -name string
    	who to greet (default "world")
  -pause duration
    	pause between greetings, e.g. 1.5s
  -shout
    	greet in upper case
exit status 2
```

- `flag.PrintDefaults` lists flags alphabetically and shows a default only if it isn't the zero value.
- A back-quoted word in the usage string (`` `times` ``) names the argument in the help. Otherwise the
  name is the type: `string`, `int`, `duration`, or `value` for custom types.
- `flag.Usage` is a function variable. Replace it to print your own help, writing to
  `flag.CommandLine.Output()` (stderr by default).
- Error messages always use one dash (`flag -n`), whatever the user typed.

## FlagSet: a parser you can test

`flag.String`, `flag.Parse`, and friends use one global parser, `flag.CommandLine`, which reads
`os.Args` and exits on errors. Handy in `main`, but you can't call it twice or test it.

`flag.NewFlagSet(name, handling)` creates an independent parser with the same methods (`String`,
`IntVar`, `Parse`, `Args`, ...). `handling` picks what happens on an error:

| Mode | On error | On `-h` |
|------|----------|---------|
| `flag.ExitOnError` | print, `os.Exit(2)` | print, `os.Exit(0)` |
| `flag.ContinueOnError` | print, return the error | print, return `flag.ErrHelp` |
| `flag.PanicOnError` | print, panic | print, panic |

With `ContinueOnError`, parsing is an ordinary function that returns an error: errors are values
again ([Chapter 9](09-errors.md)).

`examples/ch12/flagset/main.go`

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type config struct {
	port    int
	verbose bool
	files   []string
}

// parse is an ordinary function: it takes the arguments, returns a
// result and an error, and never exits. That makes it easy to test.
func parse(args []string) (config, error) {
	var c config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we report errors ourselves
	fs.IntVar(&c.port, "port", 8080, "port to listen on")
	fs.BoolVar(&c.verbose, "v", false, "verbose output")
	fs.BoolVar(&c.verbose, "verbose", false, "same as -v") // an alias
	if err := fs.Parse(args); err != nil {
		return c, err
	}
	c.files = fs.Args() // whatever is left after the flags
	return c, nil
}

func main() {
	for _, args := range [][]string{
		{},
		{"-port", "9000", "--verbose", "a.txt", "b.txt"},
		{"-port=9000", "-v=false", "a.txt", "-v"},
		{"-v", "--", "-not-a-flag"},
		{"-port", "http"},
		{"-x"},
		{"-help"},
	} {
		c, err := parse(args)
		fmt.Printf("%q\n  %+v\n", args, c)
		if errors.Is(err, flag.ErrHelp) {
			fmt.Println("  help requested")
		} else if err != nil {
			fmt.Println("  error:", err)
		}
	}
}
```

```console
$ go run ./ch12/flagset
[]
  {port:8080 verbose:false files:[]}
["-port" "9000" "--verbose" "a.txt" "b.txt"]
  {port:9000 verbose:true files:[a.txt b.txt]}
["-port=9000" "-v=false" "a.txt" "-v"]
  {port:9000 verbose:false files:[a.txt -v]}
["-v" "--" "-not-a-flag"]
  {port:8080 verbose:true files:[-not-a-flag]}
["-port" "http"]
  {port:0 verbose:false files:[]}
  error: invalid value "http" for flag -port: parse error
["-x"]
  {port:8080 verbose:false files:[]}
  error: flag provided but not defined: -x
["-help"]
  {port:8080 verbose:false files:[]}
  help requested
```

Things to notice:

- **Aliases.** There's no alias feature. Register one variable twice, as `-v` and `-verbose`; either
  name sets it.
- `fs.SetOutput(io.Discard)` silences the set's own messages; the caller prints errors its own way.
- `flag.ErrHelp` is a sentinel error. Check it with `errors.Is` and print help instead of an error.
- After an error, don't trust the values: `port` became `0` when `http` failed to parse.
- Subcommands (like `go build`, `go test`) use the same pattern: look at `os.Args[1]`, then parse the
  rest with that subcommand's own `FlagSet`.

## Custom flag types: flag.Value

Any type can be a flag if it implements `flag.Value`:

```go
type Value interface {
	String() string   // the current value, for help text
	Set(string) error // parse one command-line value
}
```

`Set` runs **each time** the flag appears. Register the type with `flag.Var` (or `fs.Var`). A `Set`
error is reported the usual way.

`examples/ch12/flagvalue/main.go`

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

// listFlag collects the values of a flag that may be repeated.
type listFlag []string

func (l *listFlag) String() string {
	if l == nil { // the flag package may call String on a zero value
		return ""
	}
	return strings.Join(*l, ",")
}

func (l *listFlag) Set(s string) error {
	*l = append(*l, s)
	return nil
}

// level is an enum flag that accepts only three words.
type level int

const (
	low level = iota
	medium
	high
)

var levelNames = []string{"low", "medium", "high"}

func (l level) String() string { return levelNames[l] }

func (l *level) Set(s string) error {
	for i, name := range levelNames {
		if s == name {
			*l = level(i)
			return nil
		}
	}
	return errors.New("must be low, medium, or high")
}

// Compile-time checks: the pointer types implement flag.Value.
var (
	_ flag.Value = (*listFlag)(nil)
	_ flag.Value = (*level)(nil)
)

func main() {
	var tags listFlag
	lvl := medium // the default
	flag.Var(&tags, "tag", "add a `tag` (repeatable)")
	flag.Var(&lvl, "level", "low, medium, or high")

	// Func is a shortcut: a flag that calls a function with each value.
	var upper []string
	flag.Func("up", "add `text` in upper case", func(s string) error {
		upper = append(upper, strings.ToUpper(s))
		return nil
	})
	flag.Parse()

	fmt.Printf("tags=%q level=%v up=%q\n", tags, lvl, upper)
}
```

```console
$ go run ./ch12/flagvalue -tag a -tag b --level=high -up go -up fast
tags=["a" "b"] level=high up=["GO" "FAST"]
$ go run ./ch12/flagvalue
tags=[] level=medium up=[]
$ go run ./ch12/flagvalue -level extreme
invalid value "extreme" for flag -level: must be low, medium, or high
Usage of /…/flagvalue:
  -level value
    	low, medium, or high (default medium)
  -tag tag
    	add a tag (repeatable)
  -up text
    	add text in upper case
exit status 2
```

(The `Usage of` line shows the path of the binary `go run` built; yours will differ. That's why
`greet` set its own `flag.Usage`.)

About the receivers:

- `Set` **must** have a pointer receiver, because it changes the value.
- So `*level`, not `level`, implements `flag.Value` ([Chapter 7](07-structs-methods-pointers.md)
  explains method sets), and we pass `&lvl` to `flag.Var`.
- The value you pass in is the default: `lvl := medium` gives `(default medium)` in the help, via
  `String`.
- The package may call `String` on a zero value, even a nil pointer, to decide whether to print
  `(default ...)`. Make `String` safe for that, as `listFlag` does.

`flag.Func` (and `flag.BoolFunc` for flags without a value) cover simple cases without a new type.
`flag.TextVar` accepts any type with `UnmarshalText`/`MarshalText` methods, such as `net.IP`.

## Defaults from the environment

Many tools take defaults from environment variables, so users can set them once in their shell
profile. Usual precedence: built-in default, then environment, then command line.

`os.Getenv(key)` returns the value, or `""` if unset. `os.LookupEnv(key)` also reports whether the
variable exists. Use `LookupEnv` when empty and unset mean different things.

`examples/ch12/env/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Getenv can't tell "unset" from "set to the empty string".
	fmt.Printf("Getenv:    %q\n", os.Getenv("GREETING"))
	v, ok := os.LookupEnv("GREETING")
	fmt.Printf("LookupEnv: %q, %v\n", v, ok)

	// Usual precedence: built-in default < environment < command line.
	greeting := "hello"
	if v, ok := os.LookupEnv("GREETING"); ok {
		greeting = v
	}
	flag.StringVar(&greeting, "greeting", greeting, "the greeting (default from $GREETING)")
	flag.Parse()
	fmt.Println("result:   ", greeting)
}
```

```console
$ go run ./ch12/env
Getenv:    ""
LookupEnv: "", false
result:    hello
$ GREETING= go run ./ch12/env
Getenv:    ""
LookupEnv: "", true
result:    
$ GREETING=hi go run ./ch12/env
Getenv:    "hi"
LookupEnv: "hi", true
result:    hi
$ GREETING=hi go run ./ch12/env -greeting hey
Getenv:    "hi"
LookupEnv: "hi", true
result:    hey
```

The trick: the environment value becomes the flag's *default*, which the flag overwrites if given.
(In Unix shells, `VAR=value command` sets a variable for one command. In PowerShell, run
`$env:GREETING="hi"` first.)

## Why the flag package is so small

Coming from GNU tools, `flag` feels strict:

- No combined short flags: `-in` is one flag named `in`, not `-i -n`.
- No flags after positional arguments: `prog file -v` treats `-v` as a file name.
- One and two dashes are the same, so there are no separate long and short names.

This is on purpose. Few rules with no exceptions mean parsing is never ambiguous, and the package stays
small enough to read in one sitting. That's Go's taste: a simple tool in the standard library, and
richer third-party packages (such as `spf13/pflag` for GNU-style flags) for those who need more. `seek`
sticks with the standard library and documents the rule: **options come before PATTERN.**

## Building seek: a real command-line interface

A `FlagSet` replaces the hand-written parser. `seek` gains long names for common options
(`--ignore-case`, `--line-number`, ...), full `--help` text, and a real `--color` option:

- `--color always` / `never` force highlighting on or off.
- `--color auto` (the default) highlights only when stdout is a terminal, so piping into a file or
  program gives clean text.
- The environment variable `SEEK_COLOR` sets the default; the flag overrides it.

Parsing moves to a new file, `cmd/seek/flags.go`, in the same `main` package.

### cmd/seek/flags.go (new)

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usageLine = "Usage: seek [OPTIONS] PATTERN [PATH ...]"

// usage is the text printed by --help.
const usage = usageLine + `

Search for PATTERN in each PATH. Directories are searched recursively.
With no PATH, search the current directory; a PATH of - means standard input.
Options must come before PATTERN.

Options:
  -i, --ignore-case         ignore case distinctions
  -v, --invert-match        select non-matching lines
  -n, --line-number         print line numbers
  -c, --count               print only a count of selected lines per file
  -l, --files-with-matches  print only the names of files with selected lines
  -m, --max-count N         stop after N selected lines per file
  -E                        treat PATTERN as a regular expression
      --hidden              search hidden files and directories
      --glob GLOB           only search files whose name matches GLOB, e.g. '*.go'
      --color WHEN          highlight matches: auto, always, or never
                            (default auto, or the value of $SEEK_COLOR)
      --version             print the version and exit
  -h, --help                show this help and exit

Exit status is 0 if a line was selected, 1 if none was, and 2 if an error occurred.`

// options holds everything the command line controls.
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool      // -i
	invert      bool      // -v
	lineNumbers bool      // -n
	count       bool      // -c
	filesOnly   bool      // -l
	maxCount    int       // -m N; 0 means no limit
	color       colorMode // --color WHEN
	regex       bool      // -E
	hidden      bool      // --hidden
	glob        string    // --glob PATTERN
	version     bool      // --version
}

// colorMode is the value of --color. It implements flag.Value, so the
// flag package can parse it like a built-in flag type.
type colorMode int

const (
	colorAuto colorMode = iota
	colorAlways
	colorNever
)

func (c colorMode) String() string {
	switch c {
	case colorAlways:
		return "always"
	case colorNever:
		return "never"
	default:
		return "auto"
	}
}

// Set needs a pointer receiver because it changes c. That means only
// *colorMode, not colorMode, has both methods and satisfies flag.Value.
func (c *colorMode) Set(s string) error {
	switch s {
	case "auto":
		*c = colorAuto
	case "always":
		*c = colorAlways
	case "never":
		*c = colorNever
	default:
		return errors.New(`must be "auto", "always", or "never"`)
	}
	return nil
}

var _ flag.Value = (*colorMode)(nil)

// A usageError reports a mistake on the command line.
type usageError struct {
	err error
}

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

// usageErrorf formats a message like fmt.Errorf and marks it as a usage
// error. Like fmt.Errorf, it wraps an error passed with the %w verb.
func usageErrorf(format string, args ...any) error {
	return &usageError{fmt.Errorf(format, args...)}
}

// parseArgs turns the command line into options. If the user asks for
// help, it returns flag.ErrHelp.
func parseArgs(args []string) (options, error) {
	var opts options

	// The environment supplies the default; the command line overrides it.
	if v, ok := os.LookupEnv("SEEK_COLOR"); ok {
		if err := opts.color.Set(v); err != nil {
			return opts, usageErrorf("invalid SEEK_COLOR %q: %w", v, err)
		}
	}

	// A FlagSet is a self-contained parser. ContinueOnError makes Parse
	// return errors instead of exiting the program, which keeps parseArgs
	// testable. We print our own messages, so the set's output is discarded.
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	// Register each option under a short and a long name. Both names
	// share one variable, so either spelling sets it.
	boolVar := func(p *bool, short, long string) {
		flags.BoolVar(p, short, false, "")
		if long != "" {
			flags.BoolVar(p, long, false, "")
		}
	}
	boolVar(&opts.ignoreCase, "i", "ignore-case")
	boolVar(&opts.invert, "v", "invert-match")
	boolVar(&opts.lineNumbers, "n", "line-number")
	boolVar(&opts.count, "c", "count")
	boolVar(&opts.filesOnly, "l", "files-with-matches")
	boolVar(&opts.regex, "E", "")
	boolVar(&opts.hidden, "hidden", "")
	boolVar(&opts.version, "version", "")
	flags.IntVar(&opts.maxCount, "m", 0, "")
	flags.IntVar(&opts.maxCount, "max-count", 0, "")
	flags.StringVar(&opts.glob, "glob", "", "")
	flags.Var(&opts.color, "color", "")

	// Parse stops at the first argument that isn't a flag (or after "--").
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, err // -h or --help
		}
		return opts, &usageError{err}
	}
	if opts.version {
		return opts, nil // nothing else matters
	}

	if opts.maxCount < 0 {
		return opts, usageErrorf("invalid value for -m: %d is negative", opts.maxCount)
	}
	// Match reports a malformed pattern no matter what the name is.
	if _, err := filepath.Match(opts.glob, ""); err != nil {
		return opts, usageErrorf("invalid --glob pattern %q: %w", opts.glob, err)
	}

	rest := flags.Args() // the positional arguments
	if len(rest) == 0 {
		return opts, usageErrorf("missing PATTERN")
	}
	opts.pattern, opts.paths = rest[0], rest[1:]
	if len(opts.paths) == 0 {
		opts.paths = []string{"."} // search the current directory
	}
	return opts, nil
}
```

How it fits together:

- **The help text** is a raw string literal ([Chapter 6](06-strings.md)): many lines, no escapes. It's
  hand-written instead of `PrintDefaults` output, so short and long names share a line, GNU style. So
  every usage string passed to the flag package is `""`. Error messages print only `usageLine`, hence
  the separate constant.
- **`options`** moved here from `main.go`. `color` is now a `colorMode`; `gofmt` realigns the comments.
- **`colorMode`** uses the enum pattern from [Chapter 2](02-values-and-types.md) and implements
  `flag.Value`. Its zero value is `colorAuto`, so the default needs no code.
- **`SEEK_COLOR`** is applied with `opts.color.Set` *before* parsing; `--color` then overwrites it. One
  `Set` validates both, so the error messages match.
- **`boolVar`** is a closure that registers a bool option under two names. `-m` and `--max-count` are
  registered the same way by hand.
- **Errors.** `flags.Parse` returns `flag.ErrHelp` for `-h`/`--help`, passed up unchanged. Other
  errors are wrapped in a `usageError`, so `report` prints the usage line after them.
- **`--version`** returns right after parsing, before the "missing PATTERN" check, so it works alone.
- `usageError` and `usageErrorf` moved here unchanged. `flag.ErrHelp` replaces the `errHelp`
  sentinel.

### cmd/seek/main.go

`main.go` gets shorter.

1. **Delete** `options`, `errHelp`, `usageError`, `usageErrorf`, and `parseArgs`. All but `errHelp`
   (gone) now live in `flags.go`.

2. Fix the imports: add `flag`; remove `path/filepath` and `strconv`, used only by the old parser:

```go
import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"

	"seek/internal/output"
	"seek/internal/search"
)
```

3. Remove `usage` from the first constant block (it's in `flags.go`):

```go
const (
	name    = "seek"
	version = "0.1.0"
)
```

4. In `run`, check for `flag.ErrHelp` instead of `errHelp`:

```go
	opts, err := parseArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(out, usage)
		return exitOK
	}
```

   and decide on color with a new helper, `useColor`:

```go
	if useColor(opts.color, stdout) {
		tp.Highlight = matcher
	}
```

5. `report` prints the short usage line plus a hint instead of the old `usage` constant. The whole
   function:

```go
// report prints err to standard error. Usage errors are followed by a
// reminder of how to call seek.
func report(stderr io.Writer, err error) {
	// *fs.PathError messages look like "open notes.txt: no such file or
	// directory". Users care about the path, not which call failed.
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		err = fmt.Errorf("%s: %w", pathErr.Path, pathErr.Err)
	}
	fmt.Fprintln(stderr, "seek:", err)

	var usageErr *usageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(stderr, usageLine)
		fmt.Fprintln(stderr, "Run 'seek --help' for more information.")
	}
}
```

6. Add `useColor` at the end of the file:

```go
// useColor reports whether to highlight matches written to w.
func useColor(mode colorMode, w io.Writer) bool {
	switch mode {
	case colorAlways:
		return true
	case colorNever:
		return false
	}
	// Auto: color only if w is a terminal. A type assertion asks whether
	// the interface value w holds an *os.File.
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
```

`useColor` uses the terminal check from [Chapter 11](11-io-and-files.md). `run` only knows `stdout` as
an `io.Writer`, so a type assertion ([Chapter 10](10-interfaces.md)) asks if it's really an `*os.File`.
In tests it's a `bytes.Buffer`, so auto means no color. `run` passes `stdout`, not `out`: a
`*bufio.Writer` is never a terminal.

### Checkpoint

```console
$ go run ./cmd/seek --help
Usage: seek [OPTIONS] PATTERN [PATH ...]

Search for PATTERN in each PATH. Directories are searched recursively.
With no PATH, search the current directory; a PATH of - means standard input.
Options must come before PATTERN.

Options:
  -i, --ignore-case         ignore case distinctions
  -v, --invert-match        select non-matching lines
  -n, --line-number         print line numbers
  -c, --count               print only a count of selected lines per file
  -l, --files-with-matches  print only the names of files with selected lines
  -m, --max-count N         stop after N selected lines per file
  -E                        treat PATTERN as a regular expression
      --hidden              search hidden files and directories
      --glob GLOB           only search files whose name matches GLOB, e.g. '*.go'
      --color WHEN          highlight matches: auto, always, or never
                            (default auto, or the value of $SEEK_COLOR)
      --version             print the version and exit
  -h, --help                show this help and exit

Exit status is 0 if a line was selected, 1 if none was, and 2 if an error occurred.
```

Long names, and both value syntaxes:

```console
$ go run ./cmd/seek --ignore-case --line-number fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek -i -n --max-count=1 fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/src/hello.go:4:	println("hello, fox")
```

Errors now come from the flag package. Short flags don't combine, and messages use one dash:

```console
$ go run ./cmd/seek -in fox testdata
seek: flag provided but not defined: -in
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
$ go run ./cmd/seek -m two fox testdata
seek: invalid value "two" for flag -m: parse error
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
$ go run ./cmd/seek -m
seek: flag needs an argument: -m
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
$ go run ./cmd/seek --color=blue fox testdata
seek: invalid value "blue" for flag -color: must be "auto", "always", or "never"
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
```

Color. `cat -v` shows the escape codes (`^[` is the escape character). Stdout is a pipe, so auto mode
would print no color; these runs force it:

```console
$ go run ./cmd/seek --color always -n fox testdata/poem.txt | cat -v
testdata/poem.txt:1:The quick brown ^[[1;31mfox^[[0m
testdata/poem.txt:3:The dog sleeps; the ^[[1;31mfox^[[0m runs.
$ SEEK_COLOR=always go run ./cmd/seek fox testdata/poem.txt | cat -v
testdata/poem.txt:The quick brown ^[[1;31mfox^[[0m
testdata/poem.txt:The dog sleeps; the ^[[1;31mfox^[[0m runs.
$ SEEK_COLOR=always go run ./cmd/seek --color=never fox testdata/poem.txt | cat -v
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
$ SEEK_COLOR=sometimes go run ./cmd/seek fox testdata
seek: invalid SEEK_COLOR "sometimes": must be "auto", "always", or "never"
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
```

Run `go run ./cmd/seek fox testdata` in a terminal without a pipe: matches are now red by default.

Options after PATTERN are PATHs, `--` ends the options, and `--version` wins early:

```console
$ go run ./cmd/seek fox testdata -n
seek: -n: no such file or directory
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
exit status 2
$ go run ./cmd/seek -- -n testdata
exit status 1
$ go run ./cmd/seek --version extra args are ignored
seek 0.1.0
```

`seek -- -n testdata` searches for the text `-n`. Nothing in `testdata` has it, so the exit status
is 1.

The complete project at this point is in `project/ch12/`.

## Summary

- `flag.String`/`Int`/`Bool`/`Duration` return pointers; `XxxVar` forms fill existing variables. Call
  `Parse`, then read positional arguments with `Args`.
- `-x` equals `--x`. `-x=v` always works; `-x v` only for non-bool flags. Bool flags need `-x=false`.
- Parsing stops at the first non-flag argument or at `--`.
- `flag.NewFlagSet(name, flag.ContinueOnError)` returns errors instead of exiting, so parsing stays
  testable. `-h` gives `flag.ErrHelp`.
- Register a variable twice for a short and a long name.
- Implement `flag.Value` (`String` + pointer-receiver `Set`) for custom types like enums and lists.
  `flag.Func` covers simple cases.
- `flag.Usage` and `PrintDefaults` produce help. Back-quoted words in usage strings name the argument.
- `os.LookupEnv` tells unset from empty. Use environment values as defaults that flags override.
- The package is deliberately minimal: no combined short flags, no flags after arguments.

## Exercises

1. **repeat.** Write a program with flags `-n` (default 2) and `-sep` (default a space) that prints its
   positional arguments joined by `-sep`, `-n` times.

<details><summary>Solution</summary>

```go
package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	n := flag.Int("n", 2, "print the line `N` times")
	sep := flag.String("sep", " ", "separator between words")
	flag.Parse()

	line := strings.Join(flag.Args(), *sep)
	for range *n {
		fmt.Println(line)
	}
}
```

```console
$ go run . -n 3 -sep - a b c
a-b-c
a-b-c
a-b-c
$ go run . hello world
hello world
hello world
```

</details>

2. **A list flag.** Write a `csvList` type implementing `flag.Value` that accepts a comma-separated
   list like `-langs go,zig,rust`. Reject empty items (`go,,rust`). The default list is `go`.

<details><summary>Solution</summary>

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

// csvList is a flag.Value holding a comma-separated list.
type csvList []string

func (c *csvList) String() string {
	if c == nil {
		return ""
	}
	return strings.Join(*c, ",")
}

func (c *csvList) Set(s string) error {
	items := strings.Split(s, ",")
	for _, item := range items {
		if item == "" {
			return errors.New("empty item in list")
		}
	}
	*c = items
	return nil
}

func main() {
	langs := csvList{"go"}
	flag.Var(&langs, "langs", "comma-separated `list` of languages")
	flag.Parse()
	fmt.Printf("%d: %q\n", len(langs), langs)
}
```

```console
$ go run .
1: ["go"]
$ go run . -langs go,zig,rust
3: ["go" "zig" "rust"]
$ go run . -langs go,,rust
invalid value "go,,rust" for flag -langs: empty item in list
Usage of /…/csvlist:
  -langs list
    	comma-separated list of languages (default go)
exit status 2
```

(The binary's path in the `Usage of` line varies.) Unlike `listFlag`, this `Set` replaces the list
instead of appending, so a second `-langs` overrides the first.

</details>

3. **Which flags were set?** `flag.Visit` calls a function for each flag *set on the command line*;
   `flag.VisitAll` visits every flag. Define `-port` (default 8080), `-host` (default `localhost`), and
   `-v`, and use both to print what was set.

<details><summary>Solution</summary>

```go
package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Int("port", 8080, "port")
	flag.String("host", "localhost", "host")
	flag.Bool("v", false, "verbose")
	flag.Parse()

	fmt.Println("set on the command line:")
	flag.Visit(func(f *flag.Flag) {
		fmt.Printf("  -%s=%v\n", f.Name, f.Value)
	})
	fmt.Println("all flags:")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("  -%s=%v (default %s)\n", f.Name, f.Value, f.DefValue)
	})
}
```

```console
$ go run . -v -port 9000
set on the command line:
  -port=9000
  -v=true
all flags:
  -host=localhost (default localhost)
  -port=9000 (default 8080)
  -v=true (default false)
```

`Visit` tells "the user passed the default value" apart from "the user passed nothing".

</details>

4. **seek: `-V` and `NO_COLOR`.** (a) Make `-V` a short alias for `--version`. (b) Many tools honor
   the [`NO_COLOR`](https://no-color.org) convention: a non-empty `NO_COLOR` turns off default color.
   Support it, but let `SEEK_COLOR` and `--color` still win.

<details><summary>Solution</summary>

In `parseArgs`, extend the environment check:

```go
	// The environment supplies the default; the command line overrides it.
	if v, ok := os.LookupEnv("SEEK_COLOR"); ok {
		if err := opts.color.Set(v); err != nil {
			return opts, usageErrorf("invalid SEEK_COLOR %q: %w", v, err)
		}
	} else if os.Getenv("NO_COLOR") != "" {
		opts.color = colorNever // https://no-color.org
	}
```

and register the alias:

```go
	boolVar(&opts.version, "V", "version")
```

Update the help text too (`-V, --version`). Check it:

```console
$ go run ./cmd/seek -V
seek 0.1.0
$ NO_COLOR=1 go run ./cmd/seek --color=always fox testdata/poem.txt | cat -v
testdata/poem.txt:The quick brown ^[[1;31mfox^[[0m
testdata/poem.txt:The dog sleeps; the ^[[1;31mfox^[[0m runs.
```

In a terminal, `NO_COLOR=1 go run ./cmd/seek fox testdata` prints no color, and
`NO_COLOR=1 SEEK_COLOR=always go run ./cmd/seek fox testdata` does. `Getenv` is enough: the convention
treats an empty `NO_COLOR` as unset. Undo these changes if you want your code to match the snapshots.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-12--command-line-flags-and-configuration)
