// Command recursion shows functions that call themselves.
package main

import "fmt"

func factorial(n int) int {
	if n <= 1 {
		return 1 // the base case stops the recursion
	}
	return n * factorial(n-1)
}

func main() {
	fmt.Println(factorial(5), factorial(20))

	// A function literal can't refer to itself by a name declared with :=,
	// so declare the variable first, then assign.
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}
	fmt.Println(fib(10))
}
