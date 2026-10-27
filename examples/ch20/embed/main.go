package main

import (
	"embed"
	"fmt"
	"io/fs"
)

// A string (or []byte) variable holds the contents of one file.
//
//go:embed hello.txt
var greeting string

// An embed.FS holds a tree of files. It implements fs.FS.
//
//go:embed static
var static embed.FS

func main() {
	fmt.Print(greeting)

	data, err := static.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}
	fmt.Printf("index.html: %q\n", data)

	fs.WalkDir(static, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path, d.IsDir())
		return nil
	})
}
