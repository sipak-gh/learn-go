// Command printing shows the three basic print functions in fmt.
package main

import "fmt"

func main() {
	// Println puts spaces between its arguments and ends with a newline.
	fmt.Println("Go", "has", 25, "keywords:", true)

	// Print adds no newline. It adds a space between two arguments
	// only when neither of them is a string.
	fmt.Print("one", "two", "\n")
	fmt.Print(1, 2, "\n")

	/* Printf fills in a format string. Each verb (%s, %d, %v)
	   is replaced by the next argument. */
	fmt.Printf("%s was announced in %d.\n", "Go", 2009)
	fmt.Printf("%v works for any value: %v, %v\n", "%v", 3.5, false)
}
