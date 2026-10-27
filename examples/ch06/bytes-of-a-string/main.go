// A string is a read-only sequence of bytes.
package main

import "fmt"

func main() {
	s := "hello"
	fmt.Println(len(s))      // number of bytes
	fmt.Println(s[0])        // indexing gives a byte (a uint8)
	fmt.Printf("%c\n", s[0]) // %c prints a byte or rune as a character
	fmt.Println(s[1:4])      // slicing gives a string (shares the bytes)

	t := s         // copies the string header, not the bytes
	s += ", world" // builds a NEW string; t still sees the old one
	fmt.Println(s)
	fmt.Println(t)

	fmt.Println("go" < "gopher", "a" == "a") // strings compare byte by byte
}
