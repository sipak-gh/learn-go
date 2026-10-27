// A tour of the strings package.
package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "  The quick brown fox  "

	// Searching.
	fmt.Println(strings.Contains(s, "quick")) // true
	fmt.Println(strings.Index(s, "o"))        // first byte offset, or -1
	fmt.Println(strings.LastIndex(s, "o"))    // last byte offset, or -1
	fmt.Println(strings.Count(s, "o"))        // non-overlapping count
	fmt.Println(strings.HasPrefix("seek.go", "seek"), strings.HasSuffix("seek.go", ".go"))

	// Trimming.
	fmt.Printf("%q\n", strings.TrimSpace(s))
	fmt.Printf("%q\n", strings.Trim("--x--", "-"))      // any of the chars in "-"
	fmt.Printf("%q\n", strings.TrimPrefix("v1.2", "v")) // only if present

	// Splitting and joining.
	fmt.Printf("%q\n", strings.Fields(s))            // split on runs of spaces
	fmt.Printf("%q\n", strings.Split("a,b,,c", ",")) // keeps empty fields
	fmt.Printf("%q\n", strings.SplitN("a=b=c", "=", 2))
	fmt.Println(strings.Join([]string{"x", "y", "z"}, "/"))

	// Cut splits around the first separator.
	key, value, found := strings.Cut("color=red", "=")
	fmt.Println(key, value, found)

	// Changing (always returns a new string).
	fmt.Println(strings.ReplaceAll("oink oink", "k", "ky"))
	fmt.Println(strings.Replace("oink oink", "oink", "moo", 1)) // n = how many
	fmt.Println(strings.ToUpper("go"), strings.ToLower("GO"), strings.Repeat("ab", 3))

	// Comparing without regard to case.
	fmt.Println(strings.EqualFold("Gopher", "GOPHER"))
}
