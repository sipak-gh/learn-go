// Command maps shows how to create, read, write, and delete map entries.
package main

import "fmt"

func main() {
	ages := map[string]int{ // a map literal
		"alice": 31,
		"bob":   27,
	}
	ages["carol"] = 45 // add
	ages["bob"]++      // update
	fmt.Println(ages["alice"], ages["bob"], len(ages))

	// A missing key gives the zero value...
	fmt.Println(ages["dave"])

	// ...so use the "comma ok" form to tell "missing" from "zero".
	if age, ok := ages["dave"]; ok {
		fmt.Println("dave is", age)
	} else {
		fmt.Println("no dave")
	}

	delete(ages, "alice") // deleting a missing key is fine too
	_, ok := ages["alice"]
	fmt.Println(ok, len(ages))

	// make creates an empty map. The zero value makes counting easy.
	counts := make(map[string]int)
	for _, w := range []string{"a", "b", "a", "c", "a"} {
		counts[w]++
	}
	fmt.Println(counts) // fmt prints maps sorted by key
}
