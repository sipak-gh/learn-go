package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Map applies f to every element. It has two type parameters.
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// Filter keeps the elements for which keep returns true.
func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce combines the elements into one value, starting from init.
func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6}

	// T and U are inferred from the arguments: int and string.
	strs := Map(nums, strconv.Itoa)
	fmt.Printf("%q\n", strs)

	even := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println(even)

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println(sum)

	// Explicit type arguments are allowed, and sometimes needed.
	lens := Map[string, int](strings.Fields("go is fun"), func(s string) int { return len(s) })
	fmt.Println(lens)
}
