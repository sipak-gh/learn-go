package main

import "fmt"

// Countdown is a function that range can loop over. It calls yield once
// per value. yield returns false when the loop wants to stop.
func Countdown(yield func(int) bool) {
	for i := 3; i >= 1; i-- {
		if !yield(i) {
			return
		}
	}
}

// Three yield shapes: no values, one value, two values.
func Twice(yield func() bool) {
	_ = yield() && yield()
}

func Letters(yield func(int, string) bool) {
	for i, s := range []string{"a", "b", "c"} {
		if !yield(i, s) {
			return
		}
	}
}

func main() {
	for n := range Countdown {
		fmt.Print(n, " ")
	}
	fmt.Println("liftoff")

	for range Twice {
		fmt.Println("hello")
	}

	for i, s := range Letters {
		fmt.Println(i, s)
	}
}
