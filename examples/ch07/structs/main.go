// Struct types, literals, zero values, and comparison.
package main

import "fmt"

// A Book groups related values under one name.
type Book struct {
	Title  string
	Author string
	Pages  int
	Tags   []string
}

// Fields of the same type can share a line.
type Point struct {
	X, Y int
}

func main() {
	// A literal with field names. Order doesn't matter; missing fields
	// get their zero value.
	b := Book{Title: "The Go Programming Language", Pages: 380}
	b.Author = "Donovan & Kernighan" // set a field
	fmt.Printf("%+v\n", b)

	// The zero value of a struct has every field at its zero value.
	var empty Book
	fmt.Printf("%q %q %d %v %t\n", empty.Title, empty.Author, empty.Pages, empty.Tags, empty.Tags == nil)

	// A positional literal lists every field in order. Fine for tiny
	// types like Point; fragile for anything bigger.
	p := Point{3, 4}

	// Structs are compared field by field with == if every field is comparable.
	fmt.Println(p == Point{X: 3, Y: 4}, p == Point{})

	// Comparable structs can be map keys.
	visits := map[Point]int{p: 1}
	visits[Point{3, 4}]++
	fmt.Println(visits)

	// Assignment copies the whole struct.
	q := p
	q.X = 100
	fmt.Println(p, q)
}
