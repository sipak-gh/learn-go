// Package search_test is an external test package: it sees package
// search only from the outside, exactly as other packages do.
package search_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"seek/internal/search"
)

const poem = `The quick brown fox
jumps over the lazy dog.
The dog sleeps; the fox runs.
Foxes are quick and clever.
`

// newSearcher is a test helper. t.Helper makes failures point at the
// caller's line instead of a line in here.
func newSearcher(t *testing.T, pattern string, ignoreCase bool) *search.Searcher {
	t.Helper()
	m, err := search.NewMatcher(pattern, ignoreCase, false)
	if err != nil {
		t.Fatal(err)
	}
	return &search.Searcher{Matcher: m}
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		ignoreCase bool
		invert     bool
		maxCount   int
		wantLines  []int // line numbers of the expected matches
	}{
		{name: "plain", pattern: "fox", wantLines: []int{1, 3}},
		{name: "ignore case", pattern: "fox", ignoreCase: true, wantLines: []int{1, 3, 4}},
		{name: "invert", pattern: "fox", invert: true, wantLines: []int{2, 4}},
		{name: "max count", pattern: "o", maxCount: 2, wantLines: []int{1, 2}},
		{name: "no match", pattern: "zebra", wantLines: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSearcher(t, tt.pattern, tt.ignoreCase)
			s.Invert, s.MaxCount = tt.invert, tt.maxCount
			matches, err := s.Search(t.Context(), strings.NewReader(poem), "poem.txt")
			if err != nil {
				t.Fatal(err)
			}
			var lines []int
			for _, m := range matches {
				lines = append(lines, m.Line)
			}
			if !slices.Equal(lines, tt.wantLines) {
				t.Errorf("matched lines %v, want %v", lines, tt.wantLines)
			}
		})
	}
}

func TestSearchBinary(t *testing.T) {
	s := newSearcher(t, "fox", false)
	_, err := s.Search(t.Context(), strings.NewReader("GIF89a\x00fox"), "image.bin")
	if !errors.Is(err, search.ErrBinary) {
		t.Errorf("got error %v, want ErrBinary", err)
	}
}

func TestSearchLongLine(t *testing.T) {
	s := newSearcher(t, "fox", false)
	line := strings.Repeat("x", 100_000) + "fox" // longer than Scanner's default 64 KiB
	matches, err := s.Search(t.Context(), strings.NewReader(line), "long.txt")
	if err != nil || len(matches) != 1 {
		t.Errorf("got %d matches and error %v, want 1 match", len(matches), err)
	}
}

// An Example is documentation that go test runs: it fails if the
// printed output differs from the Output comment. go doc shows it too.
func ExampleSearcher_Search() {
	m, _ := search.NewMatcher("fox", true, false)
	s := &search.Searcher{Matcher: m}
	matches, _ := s.Search(context.Background(), strings.NewReader(poem), "poem.txt")
	for _, m := range matches {
		fmt.Printf("%d: %s\n", m.Line, m.Text)
	}
	// Output:
	// 1: The quick brown fox
	// 3: The dog sleeps; the fox runs.
	// 4: Foxes are quick and clever.
}

// BenchmarkSearch measures each kind of matcher on the same input. Run
// it with: go test -bench=. -benchmem ./internal/search
func BenchmarkSearch(b *testing.B) {
	text := strings.Repeat(poem, 1000) // about 100 KB
	cases := []struct {
		name              string
		pattern           string
		ignoreCase, regex bool
	}{
		{"literal", "fox", false, false},
		{"ignore-case", "fox", true, false},
		{"regex", "f[aeiou]x", false, true},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			m, err := search.NewMatcher(c.pattern, c.ignoreCase, c.regex)
			if err != nil {
				b.Fatal(err)
			}
			s := &search.Searcher{Matcher: m}
			b.SetBytes(int64(len(text))) // report throughput in MB/s
			// b.Loop runs the body as many times as needed for a
			// stable measurement; setup above it is not timed.
			for b.Loop() {
				if _, err := s.Search(b.Context(), strings.NewReader(text), "bench"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
