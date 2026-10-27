// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
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
}

// run is the whole program. It returns the exit status.
func run(args []string) int {
	opts, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(opts.paths) == 0 {
		return usageError("missing PATH")
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

	total := 0
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		total += len(matches)
		p.PrintFile(path, matches)
	}
	if total == 0 {
		return exitNoMatch
	}
	return exitOK
}

// parseArgs turns the command line into options. If the program should
// stop right away (after --help, --version, or a usage error), done is
// true and code is the exit status to use.
func parseArgs(args []string) (opts options, code int, done bool) {
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return opts, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return opts, exitOK, true
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
				return opts, usageError("-m needs a number"), true
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return opts, usageError("invalid number for -m: " + strconv.Quote(args[i])), true
			}
			opts.maxCount = n
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return opts, usageError("unknown option " + arg), true
			}
			break loop
		}
	}
	if i >= len(args) {
		return opts, usageError("missing PATTERN"), true
	}
	opts.pattern = args[i]
	opts.paths = args[i+1:]
	return opts, 0, false
}

// usageError reports a command-line mistake and returns the exit status.
func usageError(msg string) int {
	fmt.Fprintln(os.Stderr, "seek:", msg)
	fmt.Fprintln(os.Stderr, usage)
	return exitError
}
