// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
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
	pattern, paths, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(paths) != 1 {
		return usageError("expected exactly one PATH")
	}

	// match is a closure: it captures pattern from the enclosing function.
	match := func(line string) bool {
		return strings.Contains(line, pattern)
	}

	n, err := searchFile(paths[0], match)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seek:", err)
		return exitError
	}
	if n == 0 {
		return exitNoMatch
	}
	return exitOK
}

// parseArgs splits the command line into a pattern and paths. If the
// program should stop right away (after --help, --version, or a usage
// error), done is true and code is the exit status to use.
func parseArgs(args []string) (pattern string, paths []string, code int, done bool) {
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return "", nil, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return "", nil, exitOK, true
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return "", nil, usageError("unknown option " + arg), true
			}
			break loop
		}
	}
	if i >= len(args) {
		return "", nil, usageError("missing PATTERN"), true
	}
	return args[i], args[i+1:], 0, false
}

// usageError reports a command-line mistake and returns the exit status.
func usageError(msg string) int {
	fmt.Fprintln(os.Stderr, "seek:", msg)
	fmt.Fprintln(os.Stderr, usage)
	return exitError
}

// searchFile prints every line of the file at path for which match
// returns true, and returns how many lines it printed.
func searchFile(path string, match func(string) bool) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	count := 0
	for _, line := range strings.Split(text, "\n") {
		if match(line) {
			fmt.Println(line)
			count++
		}
	}
	return count, nil
}
