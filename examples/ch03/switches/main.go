// Command switches shows the forms of switch.
package main

import "fmt"

func main() {
	// Expression switch. Cases don't fall through: no break needed.
	day := "sat"
	switch day {
	case "sat", "sun": // several values in one case
		fmt.Println("weekend")
	case "fri":
		fmt.Println("almost weekend")
	default:
		fmt.Println("weekday")
	}

	// No expression: like a chain of if/else if. The first true case wins.
	score := 82
	switch {
	case score >= 90:
		fmt.Println("A")
	case score >= 80:
		fmt.Println("B")
	default:
		fmt.Println("C or lower")
	}

	// An init statement, like in if.
	switch n := len(day); {
	case n > 2:
		fmt.Println("long name")
	default:
		fmt.Println("short name")
	}

	// fallthrough continues into the next case's body without testing it.
	switch level := 2; level {
	case 2:
		fmt.Println("level 2")
		fallthrough
	case 1:
		fmt.Println("level 1")
	case 0:
		fmt.Println("level 0")
	}
}
