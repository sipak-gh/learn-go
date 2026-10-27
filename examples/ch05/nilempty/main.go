// Command nilempty compares nil slices and empty slices.
package main

import "fmt"

func main() {
	var a []int    // nil: no underlying array
	b := []int{}   // empty, but not nil
	c := a[0:0:0]  // slicing a nil slice gives nil
	var d []string // nil

	fmt.Println(len(a), cap(a), a == nil)
	fmt.Println(len(b), cap(b), b == nil)
	fmt.Println(c == nil)

	// Both behave the same for len, range, and append.
	for range a {
		fmt.Println("never runs")
	}
	a = append(a, 1)
	d = append(d, "ok")
	fmt.Println(a, d)

	// To check for "no elements", test the length, not nil.
	fmt.Println(len(b) == 0)
}
