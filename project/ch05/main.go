// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
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

// run is the whole program. It returns the exit status.
func run(args []string) int {
	pattern, paths, opts, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(paths) == 0 {
		return usageError("missing PATH")
	}

	// match is a closure: it captures pattern from the enclosing function.
	match := func(line string) bool {
		return strings.Contains(line, pattern)
	}

	counts := make(map[string]int) // number of matching lines per path
	total := 0
	for _, path := range paths {
		lines, err := searchFile(path, match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		counts[path] = len(lines)
		total += len(lines)
		if !opts["-c"] {
			for _, line := range lines {
				fmt.Printf("%s:%s\n", path, line)
			}
		}
	}

	if opts["-c"] {
		// Ranging over the map would visit the paths in random order, so
		// we walk the paths slice and look each count up instead.
		for _, path := range paths {
			fmt.Printf("%s:%d\n", path, counts[path])
		}
	}
	if total == 0 {
		return exitNoMatch
	}
	return exitOK
}

// boolOptions lists the on/off options seek understands.
var boolOptions = []string{"-c"}

// parseArgs splits the command line into options, a pattern, and paths.
// The options that were given are returned as a set: opts["-c"] is true
// if -c was given. If the program should stop right away (after --help,
// --version, or a usage error), done is true and code is the exit status.
func parseArgs(args []string) (pattern string, paths []string, opts map[string]bool, code int, done bool) {
	opts = make(map[string]bool)
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return "", nil, nil, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return "", nil, nil, exitOK, true
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				if !slices.Contains(boolOptions, arg) {
					return "", nil, nil, usageError("unknown option " + arg), true
				}
				opts[arg] = true
				continue
			}
			break loop
		}
	}
	if i >= len(args) {
		return "", nil, nil, usageError("missing PATTERN"), true
	}
	return args[i], args[i+1:], opts, 0, false
}

// usageError reports a command-line mistake and returns the exit status.
func usageError(msg string) int {
	fmt.Fprintln(os.Stderr, "seek:", msg)
	fmt.Fprintln(os.Stderr, usage)
	return exitError
}

// searchFile returns the lines of the file at path for which match
// returns true.
func searchFile(path string, match func(string) bool) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	var matches []string // a nil slice is ready to append to
	for _, line := range strings.Split(text, "\n") {
		if match(line) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}
