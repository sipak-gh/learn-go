// The os package returns *fs.PathError values.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	_, err := os.Open("missing.txt")
	fmt.Println(err)

	// Ask what kind of failure it was, not what the message says.
	fmt.Println("not exist:", errors.Is(err, fs.ErrNotExist))
	fmt.Println("permission:", errors.Is(err, fs.ErrPermission))

	// Pull out the details.
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		fmt.Printf("op=%q path=%q err=%q\n", pe.Op, pe.Path, pe.Err)
	}

	_, err = os.ReadFile(".") // a directory, not a file
	fmt.Println(err)
}
