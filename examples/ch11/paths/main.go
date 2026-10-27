package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	p := filepath.Join("docs", "guide", "..", "intro.md") // Join also cleans
	fmt.Println(p)
	fmt.Println(filepath.Base(p), filepath.Dir(p), filepath.Ext(p))
	dir, file := filepath.Split(p)
	fmt.Printf("%q %q\n", dir, file)
	fmt.Println(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
	fmt.Println(filepath.Clean("a//b/./c/.."))
	fmt.Println(filepath.IsAbs("/etc/hosts"), filepath.IsAbs("etc/hosts"))
	rel, err := filepath.Rel("/home/gopher", "/home/gopher/src/app")
	fmt.Println(rel, err)

	// Match tests a name against a shell pattern: * ? [abc] [a-z]
	for _, name := range []string{"main.go", "main_test.go", "README.md", "src/x.go"} {
		ok, _ := filepath.Match("*.go", name)
		fmt.Printf("%-14s %v\n", name, ok)
	}
	_, err = filepath.Match("[", "x")
	fmt.Println(err)
}
