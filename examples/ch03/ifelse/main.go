// Command ifelse shows if, else if, else, and if with an init statement.
package main

import "fmt"

func main() {
	temp := 23

	// No parentheses around the condition; braces are required.
	if temp > 30 {
		fmt.Println("hot")
	} else if temp > 15 {
		fmt.Println("mild")
	} else {
		fmt.Println("cold")
	}

	// An init statement runs before the condition. Variables it
	// declares exist only inside the if and its else branches.
	word := "gopher"
	if n := len(word); n > 5 {
		fmt.Println(word, "is long:", n)
	} else {
		fmt.Println(word, "is short:", n)
	}
	// n is not defined here.
}
