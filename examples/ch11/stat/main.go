package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	info, err := os.Stat("go.mod")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(info.Name(), info.Size(), info.IsDir())
	fmt.Println(info.Mode(), info.Mode().IsRegular())
	fmt.Printf("permissions: %v (octal %o)\n", info.Mode().Perm(), info.Mode().Perm())

	info, _ = os.Stat(".")
	fmt.Println(info.Mode(), info.IsDir(), info.Mode()&fs.ModeDir != 0)

	_, err = os.Stat("nope.txt")
	fmt.Println(err)
	fmt.Println(errors.Is(err, fs.ErrNotExist))

	// Is standard output a terminal? Terminals are character devices.
	info, err = os.Stdout.Stat()
	if err == nil {
		fmt.Println("stdout mode:", info.Mode())
		fmt.Println("terminal:", info.Mode()&os.ModeCharDevice != 0)
	}
}
