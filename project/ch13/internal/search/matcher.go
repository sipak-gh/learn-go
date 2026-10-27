package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// A Matcher finds the occurrences of a pattern in a line of text.
//
// The method has the same name and signature as the one on
// *regexp.Regexp, so a compiled regular expression is a Matcher without
// any extra code: Go interfaces are satisfied implicitly.
type Matcher interface {
	// FindAllStringIndex returns the byte offsets {start, end} of up to n
	// successive non-overlapping matches in s (all of them if n < 0).
	// It returns nil if there is no match.
	FindAllStringIndex(s string, n int) [][]int
}

// Compile-time checks that each type implements Matcher.
var (
	_ Matcher = literal("")
	_ Matcher = foldLiteral("")
	_ Matcher = (*regexp.Regexp)(nil)
)

// NewMatcher returns a Matcher for pattern. If regex is true, pattern is
// a regular expression in the syntax of the regexp package; otherwise it
// is plain text.
func NewMatcher(pattern string, ignoreCase, regex bool) (Matcher, error) {
	switch {
	case regex:
		if ignoreCase {
			pattern = "(?i)" + pattern // the regexp flag for ignoring case
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		return re, nil
	case ignoreCase:
		return foldLiteral(pattern), nil
	default:
		return literal(pattern), nil
	}
}

// literal matches a fixed string exactly.
type literal string

func (l literal) FindAllStringIndex(s string, n int) [][]int {
	return findAll(s, n, func(s string) (int, int) {
		i := strings.Index(s, string(l))
		return i, i + len(l)
	})
}

// foldLiteral matches a fixed string, ignoring case.
type foldLiteral string

func (l foldLiteral) FindAllStringIndex(s string, n int) [][]int {
	return findAll(s, n, func(s string) (int, int) {
		return indexFold(s, string(l))
	})
}

// findAll calls index repeatedly to collect up to n non-overlapping
// matches in s (all of them if n < 0). index reports the first match in
// the string it is given, or a negative start if there is none.
func findAll(s string, n int, index func(string) (start, end int)) [][]int {
	var spans [][]int
	for pos := 0; n < 0 || len(spans) < n; {
		start, end := index(s[pos:])
		if start < 0 {
			break
		}
		spans = append(spans, []int{pos + start, pos + end})
		if start == end {
			break // an empty pattern: one match is enough
		}
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
