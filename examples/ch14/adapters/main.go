package main

import (
	"fmt"
	"iter"
	"slices"
)

// Naturals yields 1, 2, 3, ... forever. That's fine for an iterator:
// values are only produced when the loop asks for them.
func Naturals() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Filter yields the values of seq for which keep returns true.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Map yields f(v) for each value v of seq.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Take yields at most the first n values of seq.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for v := range seq {
			if !yield(v) {
				return
			}
			i++
			if i == n {
				return
			}
		}
	}
}

func main() {
	odd := Filter(Naturals(), func(n int) bool { return n%2 == 1 })
	squares := Map(odd, func(n int) int { return n * n })
	fmt.Println(slices.Collect(Take(squares, 5)))
}
