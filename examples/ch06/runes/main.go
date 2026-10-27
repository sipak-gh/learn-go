// Bytes versus runes in a UTF-8 string.
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "héllo, 世界"
	fmt.Println("bytes:", len(s))
	fmt.Println("runes:", utf8.RuneCountInString(s))

	// Indexing reads single bytes. 'é' is two bytes: 0xc3 0xa9.
	fmt.Printf("s[1] = %#x, s[2] = %#x\n", s[1], s[2])

	// range decodes one rune at a time. i is the BYTE offset of the rune.
	for i, r := range s {
		fmt.Printf("%2d %c %U\n", i, r, r)
	}

	// Conversions.
	rs := []rune(s) // decode into a slice of code points
	bs := []byte(s) // copy the raw bytes
	fmt.Println(len(rs), len(bs))
	fmt.Println(string(rs[7:]))       // back to a string
	fmt.Println(string(rune(0x4e16))) // one code point to a string

	// Invalid UTF-8 decodes as U+FFFD, the replacement character.
	bad := "a\xffb"
	for _, r := range bad {
		fmt.Printf("%q ", r)
	}
	fmt.Println()
}
