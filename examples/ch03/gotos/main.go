// Command gotos shows goto. Real code almost never needs it.
package main

import "fmt"

func main() {
	i := 0
again:
	if i < 3 {
		fmt.Println("i =", i)
		i++
		goto again
	}
}
