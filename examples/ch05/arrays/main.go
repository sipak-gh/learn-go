// Command arrays shows fixed-size arrays and their value semantics.
package main

import "fmt"

func main() {
	var a [3]int // three ints, all zero
	a[0] = 10
	fmt.Println(a, len(a))

	b := [3]string{"x", "y", "z"}
	c := [...]int{1, 2, 3, 4} // [...] counts the elements for you
	fmt.Println(b, len(c))

	// Arrays are values: assignment copies every element.
	d := a
	d[1] = 99
	fmt.Println(a, d)

	// Arrays of the same type can be compared with ==.
	fmt.Println(a == [3]int{10, 0, 0}, a == d)

	// The length is part of the type: [3]int and [4]int are different types.
	fmt.Printf("%T %T\n", a, c)

	var board [2][3]int // an array of arrays
	board[1][2] = 7
	fmt.Println(board)
}
