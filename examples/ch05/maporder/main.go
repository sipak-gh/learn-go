// Command maporder shows that map iteration order is not fixed.
package main

import "fmt"

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	for range 3 {
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
}
