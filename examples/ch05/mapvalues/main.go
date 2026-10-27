// Command mapvalues shows maps whose values are slices or structs.
package main

import "fmt"

// point is a struct: a group of named fields (Chapter 7).
type point struct{ x, y int }

func main() {
	// Group words by their first letter.
	groups := make(map[byte][]string)
	for _, w := range []string{"apple", "bean", "avocado", "beet", "corn"} {
		groups[w[0]] = append(groups[w[0]], w) // nil slice works with append
	}
	fmt.Println(groups['a'], groups['b'], groups['c'], len(groups['z']))

	// Struct values in a map can't be changed in place.
	// Read the value, change the copy, and store it back.
	pos := map[string]point{"start": {0, 0}}
	p := pos["start"]
	p.x = 5
	pos["start"] = p
	fmt.Println(pos["start"])
}
