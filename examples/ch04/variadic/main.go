// Command variadic shows functions that take any number of arguments.
package main

import "fmt"

// sum accepts zero or more ints. Inside, nums is a []int.
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// join has a regular parameter before the variadic one, which must be last.
func join(sep string, parts ...string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func main() {
	fmt.Println(sum())
	fmt.Println(sum(1, 2, 3))

	nums := []int{4, 5, 6}
	fmt.Println(sum(nums...)) // spread a slice into the arguments

	fmt.Println(join("-", "a", "b", "c"))
}
