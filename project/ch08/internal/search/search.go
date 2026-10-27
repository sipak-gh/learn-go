// Package search finds lines of text that match a pattern.
package search

import (
	"os"
	"strings"
	"unicode/utf8"
)

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
// Set its fields before calling its methods; the zero value matches
// every line (the empty pattern matches everywhere).
type Searcher struct {
	Pattern    string
	IgnoreCase bool
	Invert     bool // select the lines that do NOT match
	MaxCount   int  // stop after this many matches per file; 0 means no limit
}

// FindAll returns the positions of up to n occurrences of the pattern in
// line (all of them if n < 0). Each position is a pair of byte offsets
// {start, end}. It returns nil if the pattern does not occur in line.
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
