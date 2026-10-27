package main

import (
	"fmt"
	"maps"
	"slices"
)

// Set is a generic type alias: Set[string] is just another name for
// map[string]struct{}. No new type is created.
type Set[T comparable] = map[T]struct{}

func Of[T comparable](items ...T) Set[T] {
	s := make(Set[T])
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

func main() {
	s := Of("go", "zig", "go")
	fmt.Println(len(s), slices.Sorted(maps.Keys(s)))

	// Because it's an alias, a plain map is interchangeable with it.
	var m map[string]struct{} = s
	_, ok := m["zig"]
	fmt.Println(ok)
	fmt.Printf("%T\n", s)
}
