package search

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestSearchAllOrder checks that results come back in input order even
// though many workers search files at the same time. Run it with -race.
func TestSearchAllOrder(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for i := range 50 {
		path := filepath.Join(dir, fmt.Sprintf("file%02d.txt", i))
		text := fmt.Sprintf("fox %d\n", i)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}

	s := &Searcher{Matcher: literal("fox")}
	i := 0
	for res := range s.SearchAll(Walk(paths, WalkOptions{}), 8) {
		if res.Err != nil {
			t.Fatal(res.Err)
		}
		if res.Path != paths[i] {
			t.Fatalf("result %d: got %s, want %s", i, res.Path, paths[i])
		}
		i++
	}
	if i != len(paths) {
		t.Fatalf("got %d results, want %d", i, len(paths))
	}
	// Each file holds "fox N\n": 6 bytes for N < 10, 7 bytes otherwise.
	if got, want := s.BytesRead(), int64(10*6+40*7); got != want {
		t.Errorf("BytesRead() = %d, want %d", got, want)
	}
}
