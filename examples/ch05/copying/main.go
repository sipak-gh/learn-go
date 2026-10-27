// Command copying shows copy and how to make an independent slice.
package main

import (
	"fmt"
	"slices"
)

func main() {
	src := []int{1, 2, 3, 4}

	dst := make([]int, 2)
	n := copy(dst, src) // copies min(len(dst), len(src)) elements
	fmt.Println(n, dst)

	// A full, independent copy.
	dup := make([]int, len(src))
	copy(dup, src)
	dup[0] = 100
	fmt.Println(src, dup)

	// The same in one call.
	clone := slices.Clone(src)
	clone[1] = 200
	fmt.Println(src, clone)

	// copy handles overlapping ranges: shift left by one.
	copy(src, src[1:])
	fmt.Println(src)
}
