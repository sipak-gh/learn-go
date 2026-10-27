package search

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// WalkOptions controls which files Walk visits.
type WalkOptions struct {
	Hidden bool   // include hidden files and directories (names starting with ".")
	Glob   string // if not empty, only visit files whose base name matches it
}

// Walk calls fn for every file to search under roots, in order. A root
// that is a directory is walked recursively, in lexical order. The root
// "-" (standard input) is passed through unchanged. If a file or
// directory can't be read, fn is called with its path and the error.
//
// Hidden entries, entries that aren't regular files (such as symbolic
// links), and files not matching opts.Glob are skipped, but only when
// they are found inside a directory: a root is always used as given.
func Walk(roots []string, opts WalkOptions, fn func(path string, err error)) {
	for _, root := range roots {
		if root == "-" {
			fn(root, nil)
			continue
		}
		// Our callback never returns an error, so neither does WalkDir.
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fn(path, err)
				return nil // keep going
			}
			if path == root {
				if !d.IsDir() {
					fn(path, nil)
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
			fn(path, nil)
			return nil
		})
	}
}
