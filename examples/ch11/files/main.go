package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// A fresh directory under the system's temp dir, removed at the end.
	dir, err := os.MkdirTemp("", "files-example")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "notes.txt")

	if err := writeNotes(path); err != nil {
		return err
	}

	// OpenFile takes flags and permissions. Here: append, write only.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if _, err := f.WriteString("appended line\n"); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Open is read only. We only read, so a deferred Close is enough.
	f, err = os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	fmt.Printf("%s", data)

	// One-call shortcuts for small files.
	other := filepath.Join(dir, "other.txt")
	if err := os.WriteFile(other, []byte("short and sweet\n"), 0o644); err != nil {
		return err
	}
	data, err = os.ReadFile(other)
	fmt.Printf("%q %v\n", data, err)

	_, err = os.Open(filepath.Join(dir, "missing.txt"))
	fmt.Println("missing file:", errors.Is(err, fs.ErrNotExist))
	return nil
}

// writeNotes creates path and writes two lines to it. The named result
// lets the deferred function report an error from Close.
func writeNotes(path string) (err error) {
	f, err := os.Create(path) // create, or truncate if it exists
	if err != nil {
		return err
	}
	defer func() {
		// Close flushes data to the OS; a failure here means lost data.
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	if _, err := fmt.Fprintln(f, "first line"); err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d + %d = %d\n", 2, 2, 4)
	return err
}
