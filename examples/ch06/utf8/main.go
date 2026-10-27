// The unicode/utf8 and unicode packages.
package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func main() {
	s := "€5"

	// Decode the first rune by hand: the rune and its size in bytes.
	r, size := utf8.DecodeRuneInString(s)
	fmt.Printf("%c is %d bytes\n", r, size)
	fmt.Println(utf8.RuneLen('a'), utf8.RuneLen('é'), utf8.RuneLen('€'), utf8.RuneLen('😀'))

	// Check that bytes are valid UTF-8.
	fmt.Println(utf8.ValidString(s), utf8.ValidString("\xff"))

	// Encode a rune by appending its bytes to a byte slice.
	buf := utf8.AppendRune(nil, '世')
	fmt.Printf("% x\n", buf)

	// The unicode package classifies and converts single runes.
	for _, r := range "aZ9 ß!" {
		fmt.Printf("%q letter=%t upper=%t digit=%t space=%t -> %q\n",
			r, unicode.IsLetter(r), unicode.IsUpper(r),
			unicode.IsDigit(r), unicode.IsSpace(r), unicode.ToUpper(r))
	}
}
