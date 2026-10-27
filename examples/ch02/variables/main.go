// Command variables shows the ways to declare variables.
package main

import "fmt"

func main() {
	var a int = 10 // name, type, and value
	var b = 20     // type inferred from the value: int
	var c int      // no value: c starts at 0
	d := 30        // short form, only inside functions

	var x, y = 1, "two" // several at once; types may differ
	i, j := 3, 4

	var (
		name  = "gopher"
		ready bool
	)

	fmt.Println(a, b, c, d)
	fmt.Println(x, y, i, j)
	fmt.Println(name, ready)

	// Assignment uses =. The right side is evaluated first,
	// so swapping needs no temporary variable.
	i, j = j, i
	a = a + 5
	b += 5
	d++
	fmt.Println(i, j, a, b, d)

	// := needs at least one new variable on the left.
	// Here k is new, and i is simply assigned.
	i, k := 100, 200
	fmt.Println(i, k)
}
