// Command funcvalues shows functions used as values.
package main

import "fmt"

// op is a function type: any func taking two ints and returning an int.
type op func(int, int) int

func add(a, b int) int { return a + b }
func mul(a, b int) int { return a * b }

// apply calls f on a and b. Functions can be parameters.
func apply(f op, a, b int) int {
	return f(a, b)
}

// pick returns a function. Functions can be results.
func pick(name string) op {
	if name == "add" {
		return add
	}
	return mul
}

func main() {
	f := add // a variable holding a function
	fmt.Println(f(2, 3))
	fmt.Printf("%T\n", f)

	fmt.Println(apply(mul, 4, 5))
	fmt.Println(pick("add")(10, 1))

	// An anonymous function (function literal), called right away.
	square := func(n int) int { return n * n }
	fmt.Println(apply(func(a, b int) int { return a - b }, 9, 4), square(7))

	var g op // the zero value of a function type is nil
	fmt.Println(g == nil)
}
