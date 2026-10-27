// Command iota shows the iota enum pattern.
package main

import "fmt"

// Weekday is a named integer type for days of the week.
type Weekday int

// iota is 0 on the first line of a const block and goes up by one per line.
// Lines with no expression repeat the previous one.
const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
)

// Sizes skip the 0 line with _ and use iota in an expression.
const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func main() {
	fmt.Println(Sunday, Monday, Tuesday, Wednesday)
	fmt.Printf("%T\n", Tuesday)
	fmt.Println(KB, MB, GB)
}
