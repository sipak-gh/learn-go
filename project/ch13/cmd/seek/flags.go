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
      --stats               print statistics to standard error when done
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
	stats       bool      // --stats
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
	boolVar(&opts.stats, "stats", "")
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
