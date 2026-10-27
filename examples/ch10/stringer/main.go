// fmt.Stringer controls how fmt prints your type.
package main

import "fmt"

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

// String makes Weekday a fmt.Stringer:
//
//	type Stringer interface { String() string }
func (d Weekday) String() string {
	switch d {
	case Sunday:
		return "Sunday"
	case Monday:
		return "Monday"
	case Tuesday:
		return "Tuesday"
	}
	return fmt.Sprintf("Weekday(%d)", int(d))
}

func main() {
	fmt.Println(Monday)
	fmt.Printf("%v %s %d\n", Tuesday, Tuesday, Tuesday) // %d still prints the number
	fmt.Println([]Weekday{Sunday, Tuesday, 9})
}
