package search

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// makeTree creates files (and their directories) under a new temporary
// directory, which the testing package deletes when the test ends.
func makeTree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fox\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// collect runs Walk and returns the visited paths relative to root,
// with forward slashes so the expectations work on every OS.
func collect(t *testing.T, root string, roots []string, opts WalkOptions) []string {
	t.Helper()
	var got []string
	for path, err := range Walk(roots, opts) {
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if rel, err := filepath.Rel(root, path); err == nil {
			path = filepath.ToSlash(rel)
		}
		got = append(got, path)
	}
	return got
}

func TestWalk(t *testing.T) {
	root := makeTree(t, "b.txt", "a.go", ".hidden.txt", ".git/config", "sub/c.txt")
	tests := []struct {
		name  string
		roots []string
		opts  WalkOptions
		want  []string
	}{
		{"default", []string{root}, WalkOptions{}, []string{"a.go", "b.txt", "sub/c.txt"}},
		{"hidden", []string{root}, WalkOptions{Hidden: true},
			[]string{".git/config", ".hidden.txt", "a.go", "b.txt", "sub/c.txt"}},
		{"glob", []string{root}, WalkOptions{Glob: "*.txt"}, []string{"b.txt", "sub/c.txt"}},
		{"hidden root is used", []string{filepath.Join(root, ".hidden.txt")}, WalkOptions{},
			[]string{".hidden.txt"}},
		{"stdin", []string{"-"}, WalkOptions{}, []string{"-"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collect(t, root, tt.roots, tt.opts)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWalkMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	var errs int
	for path, err := range Walk([]string{missing}, WalkOptions{}) {
		if path != missing || !os.IsNotExist(err) {
			t.Errorf("got %q, %v; want %q and a not-exist error", path, err, missing)
		}
		errs++
	}
	if errs != 1 {
		t.Errorf("got %d results, want 1", errs)
	}
}

func TestWalkBreak(t *testing.T) {
	root := makeTree(t, "a.txt", "b.txt", "c.txt")
	n := 0
	for range Walk([]string{root}, WalkOptions{}) {
		n++
		break // Walk must stop when the loop does, not panic
	}
	if n != 1 {
		t.Errorf("loop ran %d times, want 1", n)
	}
}
