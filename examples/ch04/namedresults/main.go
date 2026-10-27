// Command namedresults shows named results and a naked return.
package main

import "fmt"

// split names its results. They start at their zero values and act
// like local variables. A bare "return" returns their current values.
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

// minMax names its results for documentation, but returns explicitly.
func minMax(a, b int) (lo, hi int) {
	if a < b {
		return a, b
	}
	return b, a
}

func main() {
	fmt.Println(split(17))
	fmt.Println(minMax(8, 3))
}
