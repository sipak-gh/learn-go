// Interpreted and raw string literals.
package main

import "fmt"

func main() {
	// Interpreted literals use double quotes and understand escapes.
	fmt.Println("tab:\t|  quote:\"  backslash:\\")
	fmt.Println("caf\u00e9 \U0001F600 \x41") // \u: 4 hex digits, \U: 8, \x: one byte

	// Raw literals use backquotes. No escapes; newlines are kept as-is.
	path := `C:\Users\gopher` // handy for Windows paths and regexps
	fmt.Println(path)

	poem := `roses are red,
  \n stays \n here`
	fmt.Println(poem)

	// ANSI escape: ESC (0x1b) starts a terminal control sequence.
	fmt.Printf("%q\n", "\x1b[1mbold\x1b[0m")
}
