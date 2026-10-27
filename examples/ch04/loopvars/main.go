// Command loopvars shows that each loop iteration has its own variable.
package main

import "fmt"

func main() {
	var funcs []func() // a slice of functions (Chapter 5)
	for i := range 3 {
		// Each iteration has a fresh i, so each closure captures a
		// different variable (since Go 1.22).
		funcs = append(funcs, func() { fmt.Print(i, " ") })
	}
	for _, f := range funcs {
		f()
	}
	fmt.Println()

	// To share one variable across iterations, declare it outside.
	shared := 0
	var later []func()
	for range 3 {
		shared++
		later = append(later, func() { fmt.Print(shared, " ") })
	}
	for _, f := range later {
		f()
	}
	fmt.Println()
}
