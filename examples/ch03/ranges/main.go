// Command ranges shows for-range over slices, strings, and maps.
package main

import "fmt"

func main() {
	// A slice is a list of values (Chapter 5).
	colors := []string{"red", "green", "blue"}
	for i, c := range colors {
		fmt.Println(i, c)
	}
	for _, c := range colors { // skip the index with _
		fmt.Print(c, " ")
	}
	fmt.Println()
	for i := range colors { // or leave out the value
		fmt.Print(i, " ")
	}
	fmt.Println()

	// Ranging over a string yields byte offsets and runes (characters).
	for i, r := range "héllo" {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	// A map holds key/value pairs (Chapter 5). Its order is random,
	// so this sums the values instead of printing them in order.
	stock := map[string]int{"apples": 5, "pears": 3, "plums": 7}
	total := 0
	for _, n := range stock {
		total += n
	}
	fmt.Println("total:", total)
}
