package search

import (
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

// WalkOptions controls which files Walk visits.
type WalkOptions struct {
	Hidden bool   // include hidden files and directories (names starting with ".")
	Glob   string // if not empty, only visit files whose base name matches it
}

// Walk returns an iterator over the files to search under roots, in
// order. A root that is a directory is walked recursively, in lexical
// order. The root "-" (standard input) is passed through unchanged. If a
// file or directory can't be read, the iterator yields its path with the
// error and carries on.
//
// Hidden entries, entries that aren't regular files (such as symbolic
// links), and files not matching opts.Glob are skipped, but only when
// they are found inside a directory: a root is always used as given.
//
//	for path, err := range search.Walk(roots, opts) { ... }
func Walk(roots []string, opts WalkOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, root := range roots {
			if !walkRoot(root, opts, yield) {
				return // the loop body stopped early (break or return)
			}
		}
	}
}

// walkRoot yields the files under root. It reports whether to keep
// going: false means yield returned false and we must stop.
func walkRoot(root string, opts WalkOptions, yield func(string, error) bool) bool {
	if root == "-" {
		return yield(root, nil)
	}
	keepGoing := true
	// emit hands a result to yield and ends the walk if yield says stop.
	emit := func(path string, err error) error {
		if !yield(path, err) {
			keepGoing = false
			return filepath.SkipAll
		}
		return nil
	}
	// Our callback returns only nil, SkipDir, or SkipAll, so WalkDir
	// itself never returns an error.
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return emit(path, err)
		}
		if path == root {
			if !d.IsDir() {
				return emit(path, nil)
			}
			return nil
		}
		if !opts.Hidden && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir // don't descend into it
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil // a directory, symbolic link, device, ...
		}
		if opts.Glob != "" {
			// The pattern was checked by the caller, so Match can't fail.
			if ok, _ := filepath.Match(opts.Glob, d.Name()); !ok {
				return nil
			}
		}
		return emit(path, nil)
	})
	return keepGoing
}
