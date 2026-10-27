package search

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestMatchers is a table-driven test: each row is one case, and the
// loop runs them all as subtests named after the row.
func TestMatchers(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		ignoreCase bool
		regex      bool
		line       string
		want       [][]int
	}{
		{"literal", "fox", false, false, "a fox, a fox", [][]int{{2, 5}, {9, 12}}},
		{"literal no match", "Fox", false, false, "a fox", nil},
		{"literal empty pattern", "", false, false, "abc", [][]int{{0, 0}}},
		{"literal empty line", "", false, false, "", [][]int{{0, 0}}},
		{"fold", "FOX", true, false, "a Fox", [][]int{{2, 5}}},
		{"fold non-ASCII", "CAFÉ", true, false, "café", [][]int{{0, 5}}},
		{"fold kelvin sign", "k", true, false, "\u212A", [][]int{{0, 3}}},
		{"fold to kelvin sign", "\u212A", true, false, "ok", [][]int{{1, 2}}},
		{"fold empty line", "", true, false, "", [][]int{{0, 0}}},
		{"regex", "f.x", false, true, "fax fix", [][]int{{0, 3}, {4, 7}}},
		{"regex ignore case", "fox", true, true, "FOX", [][]int{{0, 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // subtests marked Parallel run at the same time
			m, err := NewMatcher(tt.pattern, tt.ignoreCase, tt.regex)
			if err != nil {
				t.Fatalf("NewMatcher: %v", err)
			}
			got := m.FindAllStringIndex(tt.line, -1)
			// slices.EqualFunc compares element by element; the
			// elements are slices too, so compare those with slices.Equal.
			if !slices.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("FindAllStringIndex(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestNewMatcherInvalidRegex(t *testing.T) {
	_, err := NewMatcher("a(b", false, true)
	if err == nil || !strings.HasPrefix(err.Error(), "invalid pattern:") {
		t.Errorf("got error %v, want an invalid pattern error", err)
	}
}

// FuzzLiteral checks literal against the standard library. Run it with
//
//	go test -fuzz=FuzzLiteral ./internal/search
//
// Without -fuzz, go test runs only the seed inputs added with f.Add and
// any failing inputs saved in testdata/fuzz.
func FuzzLiteral(f *testing.F) {
	f.Add("the quick brown fox", "fox")
	f.Add("aaaa", "aa")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, s, substr string) {
		got := literal(substr).FindAllStringIndex(s, -1)
		if substr == "" {
			if !slices.EqualFunc(got, [][]int{{0, 0}}, slices.Equal) {
				t.Fatalf("empty pattern: got %v, want [[0 0]]", got)
			}
			return
		}
		if want := strings.Count(s, substr); len(got) != want {
			t.Fatalf("got %d matches, strings.Count says %d", len(got), want)
		}
		prev := 0
		for _, span := range got {
			if span[0] < prev || s[span[0]:span[1]] != substr {
				t.Fatalf("bad span %v in %v", span, got)
			}
			prev = span[1]
		}
	})
}

// FuzzIndexFold checks two properties of indexFold that must hold for
// any valid UTF-8 input: what it finds equals substr ignoring case, and
// it finds an exact match at least as early as strings.Index does.
func FuzzIndexFold(f *testing.F) {
	f.Add("The Quick Fox", "qUiCk")
	f.Add("\u212Aelvin", "kel")
	f.Fuzz(func(t *testing.T, s, substr string) {
		// indexFold works rune by rune, so it can't find half a rune
		// ("\xd7") inside a whole one ("\u05c0"), but strings.Index can.
		if !utf8.ValidString(s) || !utf8.ValidString(substr) {
			t.Skip("not valid UTF-8")
		}
		start, end := indexFold(s, substr)
		if start < 0 {
			if strings.Contains(s, substr) {
				t.Fatalf("indexFold(%q, %q) found nothing, but s contains substr", s, substr)
			}
			return
		}
		if start > strings.Index(s, substr) && strings.Contains(s, substr) {
			t.Fatalf("indexFold(%q, %q) = %d, after the exact match", s, substr, start)
		}
		if !strings.EqualFold(s[start:end], substr) {
			t.Fatalf("indexFold(%q, %q) = %q, not equal ignoring case", s, substr, s[start:end])
		}
	})
}
