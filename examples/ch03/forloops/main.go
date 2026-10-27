// Command forloops shows the forms of Go's only loop, for.
package main

import "fmt"

func main() {
	// Three clauses: init; condition; post.
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// Condition only: Go's "while" loop.
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println("n =", n)

	// No condition: loops forever until break (or return).
	count := 0
	for {
		count++
		if count == 4 {
			break // leave the loop
		}
		if count%2 == 0 {
			continue // skip to the next iteration
		}
		fmt.Println("odd count", count)
	}

	// Range over an integer (Go 1.22+): 0, 1, ..., n-1.
	for i := range 3 {
		fmt.Print(i*10, " ")
	}
	fmt.Println()
}
