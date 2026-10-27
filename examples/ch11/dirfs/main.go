package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	root, err := makeTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)

	// An fs.FS is a read-only file tree. Its paths always use "/" and
	// are relative to the tree's root.
	fsys := os.DirFS(root)
	printTree(fsys)

	data, err := fs.ReadFile(fsys, "src/main.go")
	fmt.Printf("%q %v\n", data, err)
	matches, err := fs.Glob(fsys, "src/*.go")
	fmt.Println(matches, err)
}

// printTree works with any fs.FS: a directory, a zip file, files
// embedded in the binary, or a fake tree in a test.
func printTree(fsys fs.FS) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Printf("%-12s dir\n", path)
			return nil
		}
		info, err := d.Info() // the full FileInfo (may cost a system call)
		if err != nil {
			return err
		}
		fmt.Printf("%-12s %v %d bytes\n", path, info.Mode(), info.Size())
		return nil
	})
}

func makeTree() (string, error) {
	root, err := os.MkdirTemp("", "dirfs")
	if err != nil {
		return "", err
	}
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("# notes\n"), 0o644); err != nil {
		return "", err
	}
	err = os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644)
	return root, err
}
