package main

import (
	"fmt"
	"iter"
	"slices"
)

// Zip pairs up the values of two sequences until either one ends.
// A range loop can only drive one iterator, so Zip pulls from both.
func Zip[A, B any](as iter.Seq[A], bs iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		nextA, stopA := iter.Pull(as)
		defer stopA()
		nextB, stopB := iter.Pull(bs)
		defer stopB()
		for {
			a, ok1 := nextA()
			b, ok2 := nextB()
			if !ok1 || !ok2 || !yield(a, b) {
				return
			}
		}
	}
}

func main() {
	next, stop := iter.Pull(slices.Values([]string{"x", "y"}))
	defer stop() // always call stop to release the iterator
	for {
		v, ok := next()
		if !ok {
			break
		}
		fmt.Println("pulled", v)
	}

	names := slices.Values([]string{"ann", "bob", "cy"})
	ages := slices.Values([]int{31, 42})
	for name, age := range Zip(names, ages) {
		fmt.Println(name, age)
	}
}
