package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := makeTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // stop the walk on the first error
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			fmt.Println("skip ", rel)
			return filepath.SkipDir // don't go inside
		}
		fmt.Printf("%-5s %s\n", kind(d), rel)
		return nil
	})
	fmt.Println("err:", err)
}

func kind(d fs.DirEntry) string {
	if d.IsDir() {
		return "dir"
	}
	return "file"
}

// makeTree creates a few files under a new temporary directory.
func makeTree() (string, error) {
	root, err := os.MkdirTemp("", "walk")
	if err != nil {
		return "", err
	}
	files := map[string]string{
		"notes.md":        "# notes\n",
		"src/main.go":     "package main\n",
		"src/util/str.go": "package util\n",
		".git/config":     "[core]\n",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", err
		}
	}
	return root, nil
}
