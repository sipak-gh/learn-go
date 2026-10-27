// Command constants shows typed and untyped constants.
package main

import "fmt"

const greeting = "hello" // untyped string constant

const (
	big   = 1 << 100  // untyped: far larger than any integer type
	small = big >> 98 // 4: fine, because it fits
)

const limit int64 = 10 // typed constant

func main() {
	fmt.Println(greeting, small)

	// An untyped constant takes the type its context needs.
	var f float64 = small
	var i int8 = small
	fmt.Println(f, i)
	fmt.Println(big / 1e29) // exact arithmetic until here

	// A typed constant behaves like a value of its type.
	var n int64 = limit * 2
	fmt.Println(n)

	const third = 1.0 / 3 // computed exactly at compile time
	fmt.Println(third * 3)
}
