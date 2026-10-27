// Command scope shows block scope and shadowing.
package main

import "fmt"

var level = "package" // visible in the whole package

func main() {
	fmt.Println(level)

	level := "function" // a new variable that shadows the package one
	fmt.Println(level)

	{
		level := "block" // shadows again, only inside these braces
		fmt.Println(level)
	}
	fmt.Println(level) // the block's variable is gone

	x := 1
	{
		x := 2 // := declares a NEW x here; the outer x is unchanged
		x++
	}
	fmt.Println("x is still", x)

	// The blank identifier _ discards a value you don't need.
	_, y := 10, 20
	fmt.Println(y)
}
