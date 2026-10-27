// Command builtins shows clear, min, and max.
package main

import "fmt"

func main() {
	// min and max take any number of arguments of the same ordered type.
	fmt.Println(min(3, 1, 2), max(2.5, 7, 1), min("b", "a"))

	// clear on a map deletes every entry.
	m := map[string]int{"a": 1, "b": 2}
	clear(m)
	fmt.Println(len(m))

	// clear on a slice sets every element to its zero value.
	// The length stays the same.
	s := []int{1, 2, 3}
	clear(s)
	fmt.Println(s, len(s))

	// To empty a slice but keep its capacity, reslice to length 0.
	s = s[:0]
	fmt.Println(len(s), cap(s))
}
