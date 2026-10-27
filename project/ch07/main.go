// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
	"strconv"
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

	s := NewSearcher(opts)
	p := printer{
		lineNumbers: opts.lineNumbers,
		count:       opts.count,
		filesOnly:   opts.filesOnly,
	}
	if opts.color {
		p.find = s.FindAll // a method value: FindAll bound to s
	}

	total := 0
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		total += len(matches)
		p.printFile(path, matches)
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

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Pattern    string
	IgnoreCase bool
	Invert     bool // select the lines that do NOT match
	MaxCount   int  // stop after this many matches per file; 0 means no limit
}

// NewSearcher returns a Searcher configured from the command-line options.
func NewSearcher(opts options) *Searcher {
	s := &Searcher{
		Pattern:    opts.pattern,
		IgnoreCase: opts.ignoreCase,
		Invert:     opts.invert,
		MaxCount:   opts.maxCount,
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}
	return s
}

// FindAll returns the positions of up to n occurrences of the pattern in
// line (all of them if n < 0).
func (s *Searcher) FindAll(line string, n int) [][]int {
	return findAll(line, s.Pattern, s.IgnoreCase, n)
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.FindAll(line, 1) != nil
	return found != s.Invert // != on bools is "exclusive or"
}

// SearchFile returns the selected lines of the file at path.
func (s *Searcher) SearchFile(path string) ([]Match, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	var matches []Match
	for i, line := range strings.Split(text, "\n") {
		if !s.selects(line) {
			continue
		}
		matches = append(matches, Match{Path: path, Line: i + 1, Text: line})
		if len(matches) == s.MaxCount {
			break
		}
	}
	return matches, nil
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

// A printer writes search results to standard output.
type printer struct {
	lineNumbers bool // -n: show line numbers
	count       bool // -c: print only a count per file
	filesOnly   bool // -l: print only the names of matching files
	// find locates the text to highlight. If it is nil, nothing is.
	find func(line string, n int) [][]int
}

// printFile prints the results for one file.
func (p *printer) printFile(path string, matches []Match) {
	switch {
	case p.filesOnly:
		if len(matches) > 0 {
			fmt.Println(path)
		}
	case p.count:
		fmt.Printf("%s:%d\n", path, len(matches))
	default:
		for _, m := range matches {
			p.printMatch(m)
		}
	}
}

// printMatch prints one matching line as path:text or path:line:text.
func (p *printer) printMatch(m Match) {
	text := m.Text
	if p.find != nil {
		text = highlight(text, p.find(text, -1))
	}
	if p.lineNumbers {
		fmt.Printf("%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		fmt.Printf("%s:%s\n", m.Path, text)
	}
}

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
