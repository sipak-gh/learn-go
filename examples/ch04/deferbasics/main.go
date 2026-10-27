// Command deferbasics shows when deferred calls run.
package main

import "fmt"

func main() {
	fmt.Println("start")

	// Deferred calls run when main returns, last-in first-out.
	defer fmt.Println("deferred 1")
	defer fmt.Println("deferred 2")

	// Arguments are evaluated NOW, when the defer statement runs.
	x := 10
	defer fmt.Println("deferred x =", x)

	// A deferred closure reads x when it finally runs.
	defer func() { fmt.Println("closure sees x =", x) }()

	x = 20
	fmt.Println("end of main, x =", x)
}
