package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// usageLine is the first line of usage.txt, printed after usage errors.
const usageLine = "Usage: seek [OPTIONS] PATTERN [PATH ...]"

// usage is the text printed by --help. The go:embed directive makes the
// compiler store the contents of usage.txt in the variable, so the help
// text can be edited as a normal file but still ships inside the binary.
// The blank import of embed above is required for this.
//
//go:embed usage.txt
var usage string

// options holds everything the command line controls.
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool          // -i
	invert      bool          // -v
	lineNumbers bool          // -n
	count       bool          // -c
	filesOnly   bool          // -l
	maxCount    int           // -m N; 0 means no limit
	color       colorMode     // --color WHEN
	regex       bool          // -E
	hidden      bool          // --hidden
	glob        string        // --glob PATTERN
	workers     int           // -j N
	timeout     time.Duration // --timeout D; 0 means none
	json        bool          // --json
	stats       bool          // --stats
	debug       bool          // --debug
	version     bool          // --version
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
	boolVar(&opts.json, "json", "")
	boolVar(&opts.stats, "stats", "")
	boolVar(&opts.debug, "debug", "")
	boolVar(&opts.version, "version", "")
	flags.IntVar(&opts.maxCount, "m", 0, "")
	flags.IntVar(&opts.maxCount, "max-count", 0, "")
	flags.StringVar(&opts.glob, "glob", "", "")
	flags.IntVar(&opts.workers, "j", runtime.NumCPU(), "")
	flags.DurationVar(&opts.timeout, "timeout", 0, "") // parses "1m30s"
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
	if opts.workers < 1 {
		return opts, usageErrorf("invalid value for -j: must be at least 1")
	}
	if opts.timeout < 0 {
		return opts, usageErrorf("invalid value for --timeout: %v is negative", opts.timeout)
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
