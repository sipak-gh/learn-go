// fmt in depth: verbs, flags, width, and precision.
package main

import (
	"fmt"
	"os"
)

// A struct groups named fields; Chapter 7 covers structs fully.
type point struct {
	X, Y int
}

func main() {
	p := point{1, 2}
	fmt.Printf("%v | %+v | %#v | %T\n", p, p, p, p)

	// Width and alignment.
	fmt.Printf("[%5d] [%-5d] [%05d]\n", 42, 42, 42)
	fmt.Printf("[%8s] [%-8s]\n", "go", "go")

	// Precision: digits after the point, or max characters of a string.
	fmt.Printf("[%.2f] [%8.3f] [%.3s]\n", 3.14159, 3.14159, "gopher")
	fmt.Printf("[%*d]\n", 6, 7) // width taken from an argument

	// Other useful verbs.
	fmt.Printf("%x %X %o %b %e\n", 255, 255, 8, 5, 1234.5)
	fmt.Printf("%c %U %q\n", 'G', 'G', 'G')
	fmt.Printf("%q %x % x\n", "hi\n", "hi", "hi")
	fmt.Printf("%t %v\n", true, nil)
	fmt.Printf("100%%\n")

	// Sprintf returns the string; Fprintf writes to any destination.
	msg := fmt.Sprintf("%s has %d legs", "spider", 8)
	fmt.Println(msg)
	fmt.Fprintf(os.Stdout, "%-6s|%6s|\n", "left", "right")

	// Print adds spaces only between operands that are not strings;
	// Println always adds spaces and a newline.
	fmt.Print("a", "b", 1, 2, "\n")
	fmt.Println("a", "b", 1, 2)
}
