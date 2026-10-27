// Command mapspkg shows how to visit a map in a stable order.
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	stock := map[string]int{"pears": 3, "apples": 5, "plums": 7, "figs": 2}

	// maps.Keys returns an iterator over the keys (Chapter 14).
	// slices.Sorted collects them into a new, sorted slice.
	for _, name := range slices.Sorted(maps.Keys(stock)) {
		fmt.Println(name, stock[name])
	}

	// slices.Collect gathers an iterator's values into a slice.
	values := slices.Collect(maps.Values(stock))
	slices.Sort(values)
	fmt.Println(values)

	// maps.Clone copies a map; maps.Equal compares two.
	backup := maps.Clone(stock)
	stock["figs"] = 0
	fmt.Println(maps.Equal(stock, backup), backup["figs"])
}
