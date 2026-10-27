// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
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

	ignoreCase := opts["-i"]
	match := func(line string) bool {
		return findAll(line, pattern, ignoreCase, 1) != nil
	}
	if opts["-v"] {
		// Wrap the original closure in a new one that inverts its answer.
		contains := match
		match = func(line string) bool { return !contains(line) }
	}

	counts := make(map[string]int) // number of matching lines per path
	total := 0
	for _, path := range paths {
		nums, lines, err := searchFile(path, match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		counts[path] = len(lines)
		total += len(lines)
		if opts["-c"] {
			continue
		}
		for i, line := range lines {
			if opts["--color"] {
				line = highlight(line, findAll(line, pattern, ignoreCase, -1))
			}
			if opts["-n"] {
				fmt.Printf("%s:%d:%s\n", path, nums[i], line)
			} else {
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
var boolOptions = []string{"-c", "-i", "-n", "-v", "--color"}

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
// returns true, along with their line numbers (counting from 1).
func searchFile(path string, match func(string) bool) (nums []int, lines []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	for i, line := range strings.Split(text, "\n") {
		if match(line) {
			nums = append(nums, i+1)
			lines = append(lines, line)
		}
	}
	return nums, lines, nil
}

// findAll returns the byte offsets {start, end} of up to n
// non-overlapping occurrences of pattern in line, or all of them if n < 0.
// It returns nil if there are none.
func findAll(line, pattern string, ignoreCase bool, n int) [][]int {
	if pattern == "" {
		return [][]int{{0, 0}} // the empty pattern matches every line
	}
	var spans [][]int
	for pos := 0; n < 0 || len(spans) < n; {
		var start, end int
		if ignoreCase {
			start, end = indexFold(line[pos:], pattern)
		} else {
			start = strings.Index(line[pos:], pattern)
			end = start + len(pattern)
		}
		if start < 0 {
			break
		}
		spans = append(spans, []int{pos + start, pos + end})
		pos += end
	}
	return spans
}

// indexFold is like strings.Index, but ignores case the way
// strings.EqualFold does. A rune and its other-case form may have
// different UTF-8 lengths (the Kelvin sign "\u212A" is 3 bytes, "k" is 1),
// so it returns the end offset of the match as well as the start.
// It returns -1, -1 if substr is not in s.
func indexFold(s, substr string) (int, int) {
	// Ranging over a string visits the byte offset of each rune.
	for start := range s {
		if n := prefixFold(s[start:], substr); n >= 0 {
			return start, start + n
		}
	}
	return -1, -1
}

// prefixFold reports how many bytes at the start of s equal prefix,
// ignoring case, or -1 if s does not start with prefix.
func prefixFold(s, prefix string) int {
	i := 0
	for _, want := range prefix {
		if i >= len(s) {
			return -1
		}
		got, size := utf8.DecodeRuneInString(s[i:])
		if got != want && !strings.EqualFold(string(got), string(want)) {
			return -1
		}
		i += size
	}
	return i
}

// ANSI escape sequences understood by most terminals.
const (
	colorStart = "\x1b[1;31m" // bold red
	colorReset = "\x1b[0m"
)

// highlight wraps each span of line in color escape sequences.
func highlight(line string, spans [][]int) string {
	var b strings.Builder
	prev := 0
	for _, span := range spans {
		start, end := span[0], span[1]
		if start == end {
			continue // an empty match has nothing to color
		}
		b.WriteString(line[prev:start])
		b.WriteString(colorStart)
		b.WriteString(line[start:end])
		b.WriteString(colorReset)
		prev = end
	}
	b.WriteString(line[prev:])
	return b.String()
}
