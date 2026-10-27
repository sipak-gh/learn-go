// Command bools shows booleans, comparisons, and logical operators.
package main

import "fmt"

func main() {
	age := 20
	member := false

	adult := age >= 18 // comparisons produce a bool
	fmt.Println(adult, age == 20, age != 20)

	// && (and), || (or), ! (not). && and || stop early:
	// the right side runs only if it can change the result.
	fmt.Println(adult && member, adult || member, !member)

	// Bitwise operators work on integers.
	a, b := 0b1100, 0b1010
	fmt.Printf("%04b %04b %04b %04b\n", a&b, a|b, a^b, a&^b)
	fmt.Println(1<<4, 256>>2)
}
