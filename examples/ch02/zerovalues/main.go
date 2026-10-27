// Command zerovalues prints the zero value of several types.
package main

import "fmt"

func main() {
	var (
		i int
		f float64
		b bool
		s string
		r rune
	)
	fmt.Printf("int=%v float64=%v bool=%v string=%q rune=%v\n", i, f, b, s, r)

	// Zero values are useful right away: no setup needed.
	var total int
	total += 5
	var greeting string
	greeting += "hello"
	fmt.Println(total, greeting)
}
