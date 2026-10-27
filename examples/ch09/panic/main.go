// A panic stops normal execution, runs deferred calls, and crashes.
package main

import "fmt"

// day returns the name of weekday d. Callers must pass 0 through 6:
// anything else is a bug in the caller, so day panics.
func day(d int) string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if d < 0 || d >= len(names) {
		panic(fmt.Sprintf("day: invalid weekday %d", d))
	}
	return names[d]
}

func main() {
	defer fmt.Println("deferred calls still run")
	fmt.Println(day(2))
	fmt.Println(day(9))
	fmt.Println("never printed")
}
