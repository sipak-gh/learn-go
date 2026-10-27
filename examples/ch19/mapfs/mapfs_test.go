package mapfs

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestFindExt(t *testing.T) {
	// A MapFS is an in-memory file system: a map from slash-separated
	// paths to files. Parent directories are implied.
	fsys := fstest.MapFS{
		"main.go":           {Data: []byte("package main\n")},
		"README.md":         {Data: []byte("# hi\n")},
		"internal/a/a.go":   {Data: []byte("package a\n")},
		"internal/a/a.txt":  {Data: []byte("notes\n")},
		"vendor/x/x.go":     {Data: []byte("package x\n")},
		"internal/b/doc.go": {Data: []byte("package b\n")},
	}
	got, err := FindExt(fsys, ".go")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/a/a.go", "internal/b/doc.go", "main.go", "vendor/x/x.go"}
	if !slices.Equal(got, want) {
		t.Errorf("FindExt = %q, want %q", got, want)
	}
}

// TestFS checks that a file system behaves correctly: fstest.TestFS is
// for people writing their own fs.FS implementations.
func TestFS(t *testing.T) {
	fsys := fstest.MapFS{"a/b.txt": {Data: []byte("b")}}
	if err := fstest.TestFS(fsys, "a/b.txt"); err != nil {
		t.Fatal(err)
	}
}
