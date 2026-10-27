// Command closures shows function literals that capture variables.
package main

import "fmt"

// counter returns a function that counts its calls. Each call to counter
// creates a new n, which lives on as long as the returned function does.
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	next := counter()
	fmt.Println(next(), next(), next())

	other := counter() // a separate n
	fmt.Println(other())

	// Closures capture variables, not values: they see later changes.
	greeting := "hello"
	say := func() { fmt.Println(greeting) }
	greeting = "goodbye"
	say()

	// ...and they can change the variable.
	total := 0
	addTo := func(n int) { total += n }
	addTo(5)
	addTo(10)
	fmt.Println(total)
}
