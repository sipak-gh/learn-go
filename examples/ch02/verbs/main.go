// Command verbs shows the most common fmt verbs.
package main

import "fmt"

func main() {
	n := 255
	pi := 3.14159
	s := "go"

	fmt.Printf("%v %v %v\n", n, pi, s) // default format
	fmt.Printf("%d\n", n)              // decimal integer
	fmt.Printf("%x %X %b %o\n", n, n, n, n)
	fmt.Printf("%f %.2f %e\n", pi, pi, pi)
	fmt.Printf("%s %q\n", s, s) // %q adds quotes and escapes
	fmt.Printf("%q\n", "tab\there")
	fmt.Printf("%T %T %T %T\n", n, pi, s, 'x') // the type
	fmt.Printf("%x\n", "hi")                   // hex bytes of a string
	fmt.Printf("%t %c %U\n", true, 'G', 'G')
	fmt.Printf("%5d|%-5d|%05d\n", 42, 42, 42) // width and padding
	fmt.Printf("100%%\n")                     // a literal percent sign
}
