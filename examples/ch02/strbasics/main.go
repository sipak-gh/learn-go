// Command strbasics shows the basics of strings.
package main

import "fmt"

func main() {
	s := "hello, "
	s += "world" // concatenation builds a new string

	fmt.Println(s, len(s))
	fmt.Println(s[0], s[0:5]) // a byte, and a substring

	// Escape sequences work in double-quoted strings.
	fmt.Println("tab:\tquote:\" newline:\\n")

	// Raw strings use backquotes. No escapes; may span lines.
	raw := `C:\temp\new
second line`
	fmt.Println(raw)

	// len counts bytes, not characters.
	fmt.Println(len("café"))
}
