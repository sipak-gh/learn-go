// Package search finds lines of text that match a pattern.
package search

import (
	"os"
	"strings"
)

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Matcher  Matcher // decides which lines match; must not be nil
	Invert   bool    // select the lines that do NOT match
	MaxCount int     // stop after this many matches per file; 0 means no limit
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.Matcher.FindAllStringIndex(line, 1) != nil
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
