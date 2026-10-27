// Command seek searches files for lines that match a pattern.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"seek/internal/output"
	"seek/internal/search"
)

const (
	name    = "seek"
	version = "0.1.0"
	usage   = "Usage: seek [OPTIONS] PATTERN [PATH ...]"
)

// Exit codes. Like grep, seek reports whether anything matched.
// iota counts up from 0, so the order of these lines matters.
const (
	exitOK      = iota // 0: success (at least one line matched)
	exitNoMatch        // 1: no line matched
	exitError          // 2: bad usage or an error occurred
)

// main only translates run's result into an exit status. os.Exit ends the
// program immediately, without running deferred calls, so all real work
// happens in run, where defer behaves normally.
func main() {
	os.Exit(run(os.Args[1:]))
}

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
