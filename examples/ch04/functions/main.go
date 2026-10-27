// Command functions shows function declarations and multiple results.
package main

import "fmt"

// add takes two ints and returns their sum.
// Consecutive parameters of the same type can share it: a, b int.
func add(a, b int) int {
	return a + b
}

// divmod returns two results: the quotient and the remainder.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// greet has no results.
func greet(name string) {
	fmt.Println("hello,", name)
}

func main() {
	fmt.Println(add(2, 3))

	q, r := divmod(17, 5)
	fmt.Println(q, r)

	_, r = divmod(9, 4) // ignore a result with _
	fmt.Println(r)

	greet("gopher")
}
