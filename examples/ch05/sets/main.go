// Command sets uses maps as sets.
package main

import "fmt"

func main() {
	// map[string]bool: simple and readable.
	seen := make(map[string]bool)
	for _, w := range []string{"go", "is", "fun", "go", "is"} {
		if seen[w] { // a missing key reads as false
			fmt.Println("duplicate:", w)
			continue
		}
		seen[w] = true
	}
	fmt.Println(len(seen), "distinct words")

	// map[string]struct{}: the empty struct takes no memory.
	allowed := map[string]struct{}{"-c": {}, "-n": {}}
	_, ok := allowed["-n"]
	_, bad := allowed["-x"]
	fmt.Println(ok, bad)
}
