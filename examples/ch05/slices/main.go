// Command slices shows how to create, index, and slice slices.
package main

import "fmt"

func main() {
	primes := []int{2, 3, 5, 7, 11, 13} // a slice literal: no length
	fmt.Println(primes, len(primes), cap(primes))

	// s[low:high] includes low, excludes high.
	fmt.Println(primes[1:4]) // elements 1, 2, 3
	fmt.Println(primes[:2])  // from the start
	fmt.Println(primes[4:])  // to the end

	// make creates a slice with a length (and optionally a capacity).
	zeros := make([]int, 3)
	buf := make([]int, 0, 10)
	fmt.Println(zeros, len(buf), cap(buf))

	// append returns a new slice header; always assign the result.
	buf = append(buf, 1)
	buf = append(buf, 2, 3)
	buf = append(buf, primes[:2]...)
	fmt.Println(buf)

	// Index out of range is a run-time panic, so check lengths.
	i := 10
	if i < len(primes) {
		fmt.Println(primes[i])
	} else {
		fmt.Println("index", i, "out of range")
	}
}
