// Case conversion can change a string's length in bytes.
package main

import (
	"fmt"
	"strings"
)

func main() {
	// The Kelvin sign K (U+212A) lowercases to an ordinary k.
	kelvin := "\u212A"
	lower := strings.ToLower(kelvin)
	fmt.Println(kelvin, len(kelvin), "->", lower, len(lower))

	// So offsets found in a lowercased copy may not fit the original.
	line := "5\u212A run"
	i := strings.Index(strings.ToLower(line), "run")
	fmt.Printf("found at %d; original[%d:%d] = %q\n", i, i, i+3, line[i:i+3])

	// EqualFold compares without building lowered copies.
	fmt.Println(strings.EqualFold("\u212A", "k"), strings.EqualFold("Straße", "STRASSE"))

	// Some letters have no single-letter upper case.
	fmt.Println(strings.ToUpper("straße"))
}
