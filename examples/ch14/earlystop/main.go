package main

import (
	"fmt"
	"iter"
)

// Numbers returns an iterator over 1..n that prints what happens.
func Numbers(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; i <= n; i++ {
			fmt.Println("  yield", i)
			if !yield(i) {
				fmt.Println("  loop stopped; cleaning up")
				return
			}
		}
		fmt.Println("  done")
	}
}

func main() {
	fmt.Println("full loop:")
	for n := range Numbers(2) {
		fmt.Println("body", n)
	}

	fmt.Println("break at 2:")
	for n := range Numbers(5) {
		fmt.Println("body", n)
		if n == 2 {
			break
		}
	}

	// A range loop is rewritten into a call like this one:
	fmt.Println("called by hand:")
	Numbers(5)(func(n int) bool {
		fmt.Println("body", n)
		return n < 2 // false means "stop"
	})
}
