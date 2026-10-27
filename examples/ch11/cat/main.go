// cat prints files, or standard input for the name "-".
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	status := 0
	for _, name := range os.Args[1:] {
		if err := cat(name); err != nil {
			fmt.Fprintln(os.Stderr, "cat:", err) // errors go to stderr
			status = 1
		}
	}
	os.Exit(status)
}

func cat(name string) error {
	if name == "-" {
		_, err := io.Copy(os.Stdout, os.Stdin)
		return err
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(os.Stdout, f)
	return err
}
