// Command nilmap shows what happens when you write to a nil map.
package main

import "fmt"

func main() {
	var m map[string]int // nil: reading works, writing panics
	fmt.Println(m["x"], len(m))
	m["x"] = 1
	fmt.Println("never printed")
}
