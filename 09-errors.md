# Chapter 9 — Error Handling

Go has no exceptions. A function that can fail returns an **error value**, and the caller decides
what to do. This chapter covers the `error` type, the `if err != nil` idiom, sentinel errors, custom
error types, wrapping with `%w`, `errors.Is`, `errors.As`, Go 1.26's `errors.AsType`, joining
errors, and `panic`/`recover` for truly exceptional cases. Then `seek` stops quitting at the first
bad file: it reports each error, keeps searching, and exits with status 2.
[Chapter 8](08-packages-and-modules.md) split `seek` into packages;
[Chapter 10](10-interfaces.md) introduces interfaces.

## Errors are values

A function that can fail returns an `error` as its **last** result; `nil` means success. The caller
checks it right away:

`examples/ch09/values/main.go`

```go
// Errors are ordinary values, returned as the last result.
package main

import (
	"fmt"
	"strconv"
)

// half returns n/2, or an error if n is odd.
func half(n int) (int, error) {
	if n%2 != 0 {
		return 0, fmt.Errorf("%d is odd", n)
	}
	return n / 2, nil
}

func main() {
	for _, s := range []string{"42", "7", "seven"} {
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		h, err := half(n)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(s, "halved is", h)
	}

	_, err := strconv.Atoi("x")
	fmt.Printf("%T\n", err)  // the concrete type inside the error
	fmt.Println(err.Error()) // the message
}
```

```console
$ go run ./ch09/values
42 halved is 21
error: 7 is odd
error: strconv.Atoi: parsing "seven": invalid syntax
*strconv.NumError
strconv.Atoi: parsing "x": invalid syntax
```

- `fmt.Errorf` builds an error from a format string, like `fmt.Sprintf`.
- On failure, return zero values for the other results (`0`, `""`, `nil`). Callers must not use
  them when `err != nil`, unless the docs say otherwise.
- The second `h, err :=` works because `h` is new: `:=` needs at least one new variable, and it
  reuses `err`.

`error` is a built-in **interface** type:

```go
type error interface {
	Error() string
}
```

Any type with an `Error() string` method is an `error` ([Chapter 10](10-interfaces.md) covers
interfaces). An `error` value holds some concrete type (here `*strconv.NumError`), and `Error()`
returns its message. `fmt.Println` calls `Error()` for you.

### Why `if err != nil`?

The check is verbose on purpose:

- **Failure is visible.** Each call that can fail has its handling right below it. No exception
  jumps out of the middle of a function.
- **Errors are ordinary values.** Store, compare, wrap, collect, or pass them, with no special
  syntax.
- **You decide at each call:** return the error, retry, use a default, or report and continue.

You can't silently drop a result; the compiler makes you take both values:

```go
	n := strconv.Atoi("42")
```

```console
./main.go:9:7: assignment mismatch: 1 variable but strconv.Atoi returns 2 values
```

You *can* discard an error explicitly: `n, _ := strconv.Atoi(s)`. The `_` shows readers the choice.
Do it only when failure truly can't matter.

## Custom error types

A string message is often enough. When callers need details, define a type with an `Error` method:

`examples/ch09/custom-type/main.go`

```go
// A custom error type carries structured details.
package main

import "fmt"

// A SyntaxError reports a problem at a line of a config file.
type SyntaxError struct {
	Line int
	Msg  string
}

// Error makes *SyntaxError satisfy the error interface.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

func parse(lines []string) error {
	for i, line := range lines {
		if line == "" {
			return &SyntaxError{Line: i + 1, Msg: "empty line"}
		}
	}
	return nil // not a nil *SyntaxError: see Chapter 10
}

func main() {
	err := parse([]string{"a=1", "", "c=3"})
	fmt.Println(err)

	var e error = &SyntaxError{Line: 9, Msg: "missing '='"}
	fmt.Println(e.Error())

	fmt.Println(parse([]string{"a=1"}) == nil)
}
```

```console
$ go run ./ch09/custom-type
line 2: empty line
line 9: missing '='
true
```

- The method has a pointer receiver, so `*SyntaxError` is the error type. Return
  `&SyntaxError{...}`.
- `parse` returns type `error`, not `*SyntaxError`, and returns a literal `nil` on success.
  Returning a `nil` `*SyntaxError` through an `error` result gives an error that is **not** `nil`, a
  classic trap ([Chapter 10](10-interfaces.md) explains).
- Name error types `XxxError`, like the standard library's `*os.PathError`, `*strconv.NumError`,
  and `*json.SyntaxError`.

## Sentinel errors

A **sentinel** is an error value declared once at package level, so callers can test for it. Name it
`ErrXxx`:

`examples/ch09/sentinel/main.go`

```go
// Sentinel errors are package-level values that callers compare against.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrNotFound is returned by find when the key is missing.
var ErrNotFound = errors.New("not found")

func find(m map[string]int, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func main() {
	ages := map[string]int{"ana": 31}
	if _, err := find(ages, "bob"); err == ErrNotFound {
		fmt.Println("bob:", err)
	}

	// io.EOF is the standard library's best-known sentinel. It isn't a
	// failure: it means "no more data".
	r := strings.NewReader("hi")
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			fmt.Println("done")
			break
		}
		fmt.Printf("read %q\n", buf[:n])
	}

	// errors.New makes a new, distinct value each time.
	fmt.Println(errors.New("not found") == ErrNotFound)
}
```

```console
$ go run ./ch09/sentinel
bob: not found
read "h"
read "i"
done
false
```

- `errors.New` returns a new error with a fixed message. Two calls give two different values, even
  with the same text; that makes a sentinel unique.
- Compare against the variable, never the message. Messages are for people and may change.
- Some sentinels aren't failures: `io.EOF` just means "the end" ([Chapter 11](11-io-and-files.md)).
- A sentinel is part of your API. Export one only if callers need to act on it.

`err == ErrNotFound` works only because the error comes back unchanged. Once errors are wrapped
(next section), use `errors.Is`.

## Wrapping: adding context

A bare `no such user` deep in a program doesn't say what was happening. Each caller can add context
with `fmt.Errorf` and `%w`, which formats like `%v` and also **wraps** the error, keeping the old
one inside.

`examples/ch09/wrapping/main.go`

```go
// Wrapping adds context to an error and keeps the original inside.
package main

import (
	"errors"
	"fmt"
)

var ErrNoUser = errors.New("no such user")

func loadUser(id int) error {
	return ErrNoUser
}

func loadProfile(id int) error {
	if err := loadUser(id); err != nil {
		return fmt.Errorf("load user %d: %w", id, err) // %w wraps
	}
	return nil
}

func handle(id int) error {
	if err := loadProfile(id); err != nil {
		return fmt.Errorf("handle request: %w", err)
	}
	return nil
}

func main() {
	err := handle(7)
	fmt.Println(err)

	// Unwrap peels off one layer at a time.
	for e := err; e != nil; e = errors.Unwrap(e) {
		fmt.Printf("  %T: %v\n", e, e)
	}

	fmt.Println("err == ErrNoUser:", err == ErrNoUser)
	fmt.Println("errors.Is:", errors.Is(err, ErrNoUser))

	// %v formats the error but does not wrap it: the chain is cut.
	flat := fmt.Errorf("handle request: %v", ErrNoUser)
	fmt.Println(flat)
	fmt.Println("errors.Is on flat:", errors.Is(flat, ErrNoUser))
}
```

```console
$ go run ./ch09/wrapping
handle request: load user 7: no such user
  *fmt.wrapError: handle request: load user 7: no such user
  *fmt.wrapError: load user 7: no such user
  *errors.errorString: no such user
err == ErrNoUser: false
errors.Is: true
handle request: no such user
errors.Is on flat: false
```

Wrapped errors form a **chain**. Each link's `Unwrap() error` method returns the next;
`errors.Unwrap(err)` calls it, or returns `nil` at the end.

- **`errors.Is(err, target)`** walks the chain and reports whether any link equals `target`. Use it
  instead of `==` for sentinels.
- The message reads from outermost context to root cause: `handle request: load user 7: no such
  user`. It's a stack trace in words. Go errors don't record stack traces, so make this context
  good.
- `%w` vs `%v`: with `%w`, the inner error becomes part of your API, because callers can test for
  it. Use `%v` to hide it, such as an internal detail you may change.

### Error message style

Messages are joined with `: `, so each piece must fit mid-sentence:

| Do | Don't | Why |
|----|-------|-----|
| `open config: ...` | `Open config: ...` | lowercase: it's rarely first in the final message |
| `invalid port 99999` | `invalid port 99999.` | no final punctuation, no newline |
| `read config: ...` | `failed to read config: ...` | every error is a failure; skip the filler |
| `parse "x": ...` | `error parsing: ...` | say what you did, and with what |

Add only context the caller lacks. `os.Open` already puts the path in its error; don't add it again.

## `errors.As` and `errors.AsType`: getting at the details

`errors.Is` asks "is this error *that value*?". To ask "is there an error of *this type* in the
chain, and what are its fields?", use `errors.As` or the newer `errors.AsType`:

`examples/ch09/errors-as/main.go`

```go
// errors.As and errors.AsType find an error of a given type in a chain.
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// A RangeError reports a value outside the allowed range.
type RangeError struct {
	Value, Min, Max int
}

func (e *RangeError) Error() string {
	return fmt.Sprintf("%d is not between %d and %d", e.Value, e.Min, e.Max)
}

func parsePercent(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse percent: %w", err)
	}
	if n < 0 || n > 100 {
		return 0, fmt.Errorf("parse percent: %w", &RangeError{n, 0, 100})
	}
	return n, nil
}

func main() {
	for _, s := range []string{"150", "abc"} {
		_, err := parsePercent(s)
		fmt.Println(err)

		// errors.As: declare a variable of the target type, pass a pointer.
		var re *RangeError
		if errors.As(err, &re) {
			fmt.Println("  too far by", re.Value-re.Max)
		}

		// errors.AsType (Go 1.26): the type goes in brackets, the result
		// comes back. It reads like a comma-ok type check.
		if ne, ok := errors.AsType[*strconv.NumError](err); ok {
			fmt.Printf("  bad number %q from %s\n", ne.Num, ne.Func)
		}
	}
}
```

```console
$ go run ./ch09/errors-as
parse percent: 150 is not between 0 and 100
  too far by 50
parse percent: strconv.Atoi: parsing "abc": invalid syntax
  bad number "abc" from Atoi
```

- `errors.As(err, &target)` walks the chain. If a link has `target`'s type, it stores it in
  `target` and returns `true`. Declare the variable first and pass its address.
- `errors.AsType[T](err)` (new in Go 1.26) returns the match and an `ok` bool, so it fits in an `if`
  with no separate variable. `[*strconv.NumError]` is a **type argument**: the function is generic
  ([Chapter 13](13-generics.md)). For now, put the type you want in the brackets.
- Prefer `errors.AsType` in new code; you'll still see `errors.As` everywhere.

`go vet` catches common `errors.As` mistakes: passing a non-pointer, or a pointer to a type that
isn't an error:

```go
	var pe fs.PathError // should be *fs.PathError
	if errors.As(err, &pe) {
	...
	var pp *fs.PathError
	if errors.As(err, pp) { // should be &pp
```

```console
$ go vet .
main.go:13:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type
main.go:17:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type
```

## File errors: `*fs.PathError` and `fs.ErrNotExist`

`os` reports file problems as `*fs.PathError` values holding the operation, the path, and the
underlying cause. `io/fs` also defines sentinels for common causes:

`examples/ch09/path-error/main.go`

```go
// The os package returns *fs.PathError values.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	_, err := os.Open("missing.txt")
	fmt.Println(err)

	// Ask what kind of failure it was, not what the message says.
	fmt.Println("not exist:", errors.Is(err, fs.ErrNotExist))
	fmt.Println("permission:", errors.Is(err, fs.ErrPermission))

	// Pull out the details.
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		fmt.Printf("op=%q path=%q err=%q\n", pe.Op, pe.Path, pe.Err)
	}

	_, err = os.ReadFile(".") // a directory, not a file
	fmt.Println(err)
}
```

```console
$ go run ./ch09/path-error
open missing.txt: no such file or directory
not exist: true
permission: false
op="open" path="missing.txt" err="no such file or directory"
read .: is a directory
```

- `errors.Is(err, fs.ErrNotExist)` works on every OS. The message doesn't: on Windows it's
  `The system cannot find the file specified.` Never test error strings.
- `fs.ErrExist`, `fs.ErrPermission`, and `fs.ErrClosed` work the same way.
- `os.PathError` is another name for `fs.PathError`; both spellings work.

## Joining errors

To report several failures at once, `errors.Join` combines errors into one:

`examples/ch09/join/main.go`

```go
// errors.Join combines several errors into one.
package main

import (
	"errors"
	"fmt"
)

var (
	ErrNoName = errors.New("name is required")
	ErrNoAge  = errors.New("age is required")
)

type user struct {
	name string
	age  int
}

// validate reports every problem, not just the first.
func validate(u user) error {
	var errs []error
	if u.name == "" {
		errs = append(errs, ErrNoName)
	}
	if u.age == 0 {
		errs = append(errs, ErrNoAge)
	}
	return errors.Join(errs...) // nil if errs is empty
}

func main() {
	err := validate(user{})
	fmt.Println(err) // one line per error
	fmt.Println("no name:", errors.Is(err, ErrNoName))
	fmt.Println("no age:", errors.Is(err, ErrNoAge))

	fmt.Println(validate(user{"ana", 31}) == nil)

	// fmt.Errorf accepts several %w verbs too.
	both := fmt.Errorf("save failed: %w; cleanup failed: %w", ErrNoName, ErrNoAge)
	fmt.Println(both)
	fmt.Println("no age:", errors.Is(both, ErrNoAge))
}
```

```console
$ go run ./ch09/join
name is required
age is required
no name: true
no age: true
true
save failed: name is required; cleanup failed: age is required
no age: true
```

- `errors.Join` skips `nil` errors and returns `nil` if none remain, so `return errors.Join(errs...)`
  is correct even for an empty `errs`.
- The joined message puts each error on its own line.
- `errors.Is` and `errors.As` search every branch: a joined error is a tree, not a chain.
- `fmt.Errorf` with several `%w` verbs wraps them all.

## Handling vs. returning

At each `if err != nil`, pick **one**:

| Choice | When | Example |
|--------|------|---------|
| Return it, with context | You can't fix it here | `return fmt.Errorf("load config: %w", err)` |
| Handle it and continue | A fallback makes sense | use a default, skip the item, retry |
| Report it and stop | You're at the top (`main`, `run`) | print to stderr, exit with a status |

Handle each error **once**. Don't print an error *and* return it, or users see the same failure
twice. Low-level code returns errors; the top level decides how to show them.

## `panic`: for bugs, not for errors

A **panic** stops the current goroutine's normal flow. Deferred calls still run as it travels up the
stack. If nothing stops it, the program prints the panic value and a stack trace and exits with
status 2. You've seen run-time panics: index out of range ([Chapter 5](05-collections.md)) and `nil`
pointer dereference ([Chapter 7](07-structs-methods-pointers.md)). You can also call `panic`:

`examples/ch09/panic/main.go`

```go
// A panic stops normal execution, runs deferred calls, and crashes.
package main

import "fmt"

// day returns the name of weekday d. Callers must pass 0 through 6:
// anything else is a bug in the caller, so day panics.
func day(d int) string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if d < 0 || d >= len(names) {
		panic(fmt.Sprintf("day: invalid weekday %d", d))
	}
	return names[d]
}

func main() {
	defer fmt.Println("deferred calls still run")
	fmt.Println(day(2))
	fmt.Println(day(9))
	fmt.Println("never printed")
}
```

```console
$ go run ./ch09/panic
Tue
deferred calls still run
panic: day: invalid weekday 9

goroutine 1 [running]:
main.day(0x4d08f8?)
	/path/to/examples/ch09/panic/main.go:11 +0x134
main.main()
	/path/to/examples/ch09/panic/main.go:19 +0xaf
exit status 2
```

Paths and hex numbers in the trace will differ on your machine.

Panic rarely:

- **Programmer errors**: a broken invariant, an impossible `switch` case, a bad argument correct
  code never passes. Like an index out of range, it means the program is wrong.
- **Setup that can't fail** in a correct program. By convention, `MustXxx` functions panic instead
  of returning an error. Use them only on values fixed in the source, such as package-level
  variables:

`examples/ch09/must/main.go`

```go
// A Must function panics instead of returning an error.
package main

import (
	"fmt"
	"strconv"
)

// mustAtoi is for values that are known to be valid, such as constants
// in the source code. A failure means the program itself is wrong.
func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

var defaultPort = mustAtoi("8080") // checked once, at startup

func main() {
	fmt.Println("port", defaultPort)
	fmt.Println(mustAtoi("80x"))
}
```

```console
$ go run ./ch09/must
port 8080
panic: strconv.Atoi: parsing "80x": invalid syntax

goroutine 1 [running]:
main.mustAtoi(...)
	/path/to/examples/ch09/must/main.go:14
main.main()
	/path/to/examples/ch09/must/main.go:23 +0xda
exit status 2
```

Never panic for things that go wrong in normal use (missing files, bad input, network failures).
Those are errors; return them.

## `recover`: stopping a panic

`recover()` stops a panic and returns the value passed to `panic`. It works only when called
directly in a **deferred** function; anywhere else it returns `nil` and does nothing.

`examples/ch09/recover/main.go`

```go
// recover stops a panic, inside a deferred function.
package main

import "fmt"

// safeDiv turns a panic into an ordinary error. The named result err
// lets the deferred function change what safeDiv returns.
func safeDiv(a, b int) (q int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("safeDiv(%d, %d): %v", a, b, r)
		}
	}()
	return a / b, nil // integer division by zero panics
}

// notDeferred calls recover directly. Outside a deferred call, recover
// does nothing and returns nil.
func notDeferred() {
	fmt.Println("recover() =", recover())
}

func main() {
	fmt.Println(safeDiv(7, 2))
	fmt.Println(safeDiv(7, 0))
	notDeferred()
	fmt.Println("main keeps going")
}
```

```console
$ go run ./ch09/recover
3 <nil>
0 safeDiv(7, 0): runtime error: integer divide by zero
recover() = <nil>
main keeps going
```

How it works:

1. `a / b` panics. `safeDiv` stops, and its deferred function runs.
2. `recover()` returns the panic value, ending the panic.
3. The deferred function sets the named result `err` ([Chapter 4](04-functions.md)). `safeDiv`
   returns normally, with `q` still `0`.

`recover` returns `any`, the type that holds any value ([Chapter 10](10-interfaces.md)); `%v` prints
it.

Use `recover` at boundaries, where one failure shouldn't take down everything. A web server
recovers per request so one buggy handler can't kill it. A parser may panic deep in recursion and
recover at its entry point, turning the panic into an error. Don't use `panic`/`recover` as general
`try`/`catch`: it can be handy inside a package, but across an API, return errors.

One goroutine can't recover another's panic, and an unrecovered panic in any goroutine crashes the
whole program ([Chapter 16](16-goroutines-and-channels.md)).

## Building seek: report errors and keep going

`seek` stops at the first unreadable file:

```console
$ go run ./cmd/seek fox testdata/missing.txt testdata/poem.txt
seek: open testdata/missing.txt: no such file or directory
exit status 2
```

`poem.txt` was never searched. `grep` reports the bad file and carries on; `seek` should too. This
step changes only `cmd/seek/main.go`:

- `run` reports each file error, continues, and returns status 2 at the end if anything failed.
- `parseArgs` returns `(options, error)` instead of `(opts, code, done)` and prints nothing; `run`
  decides what to print. One job per function.
- A `usageError` type marks command-line mistakes, so `run` knows when to print the usage line.
- An `errHelp` sentinel means "the user asked for help".
- A `report` function prints errors in one format and shortens file errors to `PATH: reason`.

The `internal` packages don't change; `SearchFile` already returns `os.ReadFile`'s error.

### Step 1: imports and `options`

Add `errors` and `io/fs` to the imports:

```go
import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"seek/internal/output"
	"seek/internal/search"
)
```

`parseArgs` no longer prints the version; it sets a flag. Add a `version` field at the end of
`options`:

```go
// options holds everything the command line controls.
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool // -i
	invert      bool // -v
	lineNumbers bool // -n
	count       bool // -c
	filesOnly   bool // -l
	maxCount    int  // -m N; 0 means no limit
	color       bool // --color
	version     bool // --version
}
```

### Step 2: the error types

Delete the old `usageError` function; its name now belongs to a type. Put these right after
`options`:

```go
// errHelp is returned by parseArgs when the user asks for help.
// It is a sentinel error: callers compare against it with errors.Is.
var errHelp = errors.New("help requested")

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
```

- `usageError` wraps another error. Its `Error` returns the inner message unchanged, so users see
  `unknown option -q`, not a type name.
- Its `Unwrap` keeps the chain going, so `errors.Is` and `errors.As` see through it to the `strconv`
  error inside.
- `usageErrorf` is variadic like `fmt.Errorf` and passes its arguments on with `args...`
  ([Chapter 4](04-functions.md)). `...any` accepts values of any type.
- `errHelp` isn't a failure but a signal, like `io.EOF`. Making it an error keeps `parseArgs`'s
  simple `(options, error)` signature.
- The types are unexported; only `package main` uses them.

### Step 3: `parseArgs` returns an error

Replace `parseArgs`:

```go
// parseArgs turns the command line into options.
func parseArgs(args []string) (options, error) {
	var opts options
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			return opts, errHelp
		case "--version":
			opts.version = true
			return opts, nil
		case "-i":
			opts.ignoreCase = true
		case "-v":
			opts.invert = true
		case "-n":
			opts.lineNumbers = true
		case "-c":
			opts.count = true
		case "-l":
			opts.filesOnly = true
		case "--color":
			opts.color = true
		case "-m":
			// -m takes a value: the next argument.
			i++
			if i >= len(args) {
				return opts, usageErrorf("-m needs a number")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil {
				return opts, usageErrorf("invalid number for -m: %w", err)
			}
			if n < 0 {
				return opts, usageErrorf("invalid number for -m: %d is negative", n)
			}
			opts.maxCount = n
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return opts, usageErrorf("unknown option %s", arg)
			}
			break loop
		}
	}
	if i >= len(args) {
		return opts, usageErrorf("missing PATTERN")
	}
	opts.pattern = args[i]
	opts.paths = args[i+1:]
	if len(opts.paths) == 0 {
		return opts, usageErrorf("missing PATH")
	}
	return opts, nil
}
```

What changed:

- Results are `(options, error)`; `opts` is now a local variable.
- `-h` returns `errHelp`; `--version` sets `opts.version`. Neither prints.
- `-m` handles its two failures separately: a bad number **wraps** `strconv`'s error with `%w`; a
  negative number gets its own message.
- The `missing PATH` check moved here from `run`, so all command-line checks live in one place.

### Step 4: `report`

Add `report` after `run`. It's the one place that prints errors:

```go
// report prints err to standard error. Usage errors are followed by a
// reminder of how to call seek.
func report(err error) {
	// *fs.PathError messages look like "open notes.txt: no such file or
	// directory". Users care about the path, not which call failed.
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		err = fmt.Errorf("%s: %w", pathErr.Path, pathErr.Err)
	}
	fmt.Fprintln(os.Stderr, "seek:", err)

	var usageErr *usageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(os.Stderr, usage)
	}
}
```

- `errors.AsType` finds the `*fs.PathError` and rebuilds the message from `Path` and `Err`:
  `open testdata/missing.txt: no such file or directory` becomes
  `testdata/missing.txt: no such file or directory`. The new error still wraps the cause, so
  `errors.Is(err, fs.ErrNotExist)` keeps working.
- `errors.As` checks whether `err` is a usage error (even wrapped) and, if so, prints the usage
  line. `report` uses both forms so you see both; either would do.

### Step 5: `run` keeps going

Replace `run`:

```go
// run is the whole program. It returns the exit status.
func run(args []string) int {
	opts, err := parseArgs(args)
	if errors.Is(err, errHelp) {
		fmt.Println(usage)
		return exitOK
	}
	if err != nil {
		report(err)
		return exitError
	}
	if opts.version {
		fmt.Println(name, version)
		return exitOK
	}

	s := &search.Searcher{
		Pattern:    opts.pattern,
		IgnoreCase: opts.ignoreCase,
		Invert:     opts.invert,
		MaxCount:   opts.maxCount,
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}
	p := &output.Printer{
		LineNumbers: opts.lineNumbers,
		Count:       opts.count,
		FilesOnly:   opts.filesOnly,
	}
	if opts.color {
		p.Find = s.FindAll // a method value: FindAll bound to s
	}

	// A bad file shouldn't stop the search: report it, remember that
	// something went wrong, and carry on with the next one.
	total, hadError := 0, false
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			report(err)
			hadError = true
			continue
		}
		total += len(matches)
		p.PrintFile(path, matches)
	}

	switch {
	case hadError:
		return exitError
	case total == 0:
		return exitNoMatch
	default:
		return exitOK
	}
}
```

- `errors.Is(err, errHelp)` comes first. Help goes to standard output with status 0: it's what the
  user asked for.
- Any other `parseArgs` error is a usage error: report it and return 2.
- In the loop, a file error is reported, recorded in `hadError`, and skipped with `continue`.
- The final `switch` picks the status. An error wins even if other files matched, as in `grep`:
  status 2 means "something went wrong; the results may be incomplete".

The searching and printing parts of `run` are unchanged.

### Checkpoint

`echo $?` prints the last command's exit status. `go run` prints `exit status 2` and then exits with
status 1 itself, so build a binary instead:

```console
$ go build -o seek ./cmd/seek
$ ./seek -n fox testdata/poem.txt testdata/missing.txt testdata/notes.md; echo $?
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
seek: testdata/missing.txt: no such file or directory
2
$ ./seek fox testdata testdata/poem.txt; echo $?
seek: testdata: is a directory
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
2
$ ./seek -m two fox testdata/poem.txt; echo $?
seek: invalid number for -m: strconv.Atoi: parsing "two": invalid syntax
Usage: seek [OPTIONS] PATTERN [PATH ...]
2
$ ./seek -m -3 fox testdata/poem.txt; echo $?
seek: invalid number for -m: -3 is negative
Usage: seek [OPTIONS] PATTERN [PATH ...]
2
$ ./seek -q fox testdata/poem.txt; echo $?
seek: unknown option -q
Usage: seek [OPTIONS] PATTERN [PATH ...]
2
$ ./seek fox; echo $?
seek: missing PATH
Usage: seek [OPTIONS] PATTERN [PATH ...]
2
$ ./seek --help; echo $?
Usage: seek [OPTIONS] PATTERN [PATH ...]
0
$ ./seek cat testdata/poem.txt; echo $?
1
```

- The search continues past `missing.txt`. `notes.md` is searched too but doesn't match: it only
  has `FOX`, and there's no `-i`. The status is 2 because one file failed.
- `testdata` is a directory, so reading it fails. [Chapter 11](11-io-and-files.md) teaches `seek`
  to walk directories.
- `-m two` shows the wrapped `strconv` message: wrapping keeps the root cause in the text.

Delete the binary when you're done (`rm seek`). The complete project at this point is in
`project/ch09/`.

## Summary

- Errors are values of the built-in `error` interface (anything with `Error() string`), returned
  last; `nil` means success.
- Check `if err != nil` right after the call. Return zero values with an error.
- `errors.New` makes sentinels (`ErrXxx`); custom types (`XxxError`) carry details.
- `fmt.Errorf` with `%w` wraps an error with context. Messages are lowercase, unpunctuated, and
  chain as `outer: inner: cause`.
- `errors.Is` finds a value in the chain; `errors.As` and `errors.AsType[T]` (Go 1.26) find a type.
  Don't compare error strings.
- `*fs.PathError` and `fs.ErrNotExist` describe file errors on every OS.
- `errors.Join` and multiple `%w` verbs combine errors; `Is` and `As` search them all.
- Handle each error once: return it with context, handle it, or report it at the top.
- `panic` is for bugs and impossible states. `recover`, in a deferred function, turns a panic back
  into normal flow. Use it at boundaries.

## Exercises

1. Write `parseKV(s string) (key, value string, err error)` that splits `"key=value"`. Return a
   wrapped `ErrNoEquals` sentinel if there's no `=`, and a wrapped `ErrEmptyKey` if the key is
   empty. In `main`, skip inputs without `=` and print other errors.

<details><summary>Solution</summary>

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoEquals = errors.New("missing '='")
	ErrEmptyKey = errors.New("empty key")
)

// parseKV splits "key=value" into its parts.
func parseKV(s string) (key, value string, err error) {
	key, value, ok := strings.Cut(s, "=")
	if !ok {
		return "", "", fmt.Errorf("parse %q: %w", s, ErrNoEquals)
	}
	if key == "" {
		return "", "", fmt.Errorf("parse %q: %w", s, ErrEmptyKey)
	}
	return key, value, nil
}

func main() {
	for _, s := range []string{"name=seek", "verbose", "=1"} {
		k, v, err := parseKV(s)
		switch {
		case errors.Is(err, ErrNoEquals):
			fmt.Println("skipping:", err)
		case err != nil:
			fmt.Println("error:", err)
		default:
			fmt.Printf("%s -> %s\n", k, v)
		}
	}
}
```

```console
$ go run .
name -> seek
skipping: parse "verbose": missing '='
error: parse "=1": empty key
```

`errors.Is` sees `ErrNoEquals` through the `%w` wrapping, and the message still says which input
failed.

</details>

2. Define a `NotFoundError` type with a `Name` field. Return it from `lookup(name)`, wrap it in
   `install(name)`, and wrap it again in `main`. Print the full message, then use `errors.AsType` to
   get the name back.

<details><summary>Solution</summary>

```go
package main

import (
	"errors"
	"fmt"
)

// A NotFoundError reports a missing item.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return e.Name + " not found"
}

func lookup(name string) error {
	return &NotFoundError{Name: name}
}

func install(name string) error {
	if err := lookup(name); err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}
	return nil
}

func main() {
	err := fmt.Errorf("setup: %w", install("gopls"))
	fmt.Println(err)
	if nf, ok := errors.AsType[*NotFoundError](err); ok {
		fmt.Println("missing package:", nf.Name)
	}
}
```

```console
$ go run .
setup: install gopls: gopls not found
missing package: gopls
```

</details>

3. Write `safeRun(f func()) (err error)` that calls `f` and turns any panic into an error. Try it
   with a function that works, one that writes to a `nil` map, and one that calls `panic` with a
   string.

<details><summary>Solution</summary>

```go
package main

import "fmt"

// safeRun calls f and turns a panic into an error.
func safeRun(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safeRun(func() { fmt.Println("working") }))
	fmt.Println(safeRun(func() {
		var m map[string]int
		m["x"] = 1
	}))
	fmt.Println(safeRun(func() {
		panic("something impossible happened")
	}))
}
```

```console
$ go run .
working
<nil>
recovered: assignment to entry in nil map
recovered: something impossible happened
```

`err` must be a named result: the deferred function runs after `return`, and a named result is the
only way it can change what the caller gets.

</details>

4. Write `checkFiles(paths ...string) error` that calls `os.Stat` on each path and returns **all**
   failures as one error. Check whether any file is missing with `errors.Is`.

<details><summary>Solution</summary>

```go
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// checkFiles reports every path that can't be read, not just the first.
func checkFiles(paths ...string) error {
	var errs []error
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func main() {
	err := checkFiles("main.go", "a.txt", "b.txt")
	fmt.Println(err)
	fmt.Println("some missing:", errors.Is(err, fs.ErrNotExist))
	fmt.Println(checkFiles("main.go") == nil)
}
```

```console
$ go run .
stat a.txt: no such file or directory
stat b.txt: no such file or directory
some missing: true
true
```

`errors.Join` of an empty slice is `nil`, so success needs no special code.

</details>

5. Change `seek` so that, when some files couldn't be read, it prints one more line at the end:
   `seek: 2 of 3 files could not be read`. Keep exit status 2.

<details><summary>Solution</summary>

In `run`, count failures instead of only remembering that one happened:

```go
	total, failed := 0, 0
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			report(err)
			failed++
			continue
		}
		total += len(matches)
		p.PrintFile(path, matches)
	}

	switch {
	case failed > 0:
		fmt.Fprintf(os.Stderr, "seek: %d of %d files could not be read\n", failed, len(opts.paths))
		return exitError
	case total == 0:
		return exitNoMatch
	default:
		return exitOK
	}
```

```console
$ go build -o seek ./cmd/seek
$ ./seek fox testdata/poem.txt testdata/missing.txt testdata; echo $?
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
seek: testdata/missing.txt: no such file or directory
seek: testdata: is a directory
seek: 2 of 3 files could not be read
2
```

A count holds more information than a `bool` at no extra cost. (The snapshots don't include this
change.)

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-9--error-handling)
