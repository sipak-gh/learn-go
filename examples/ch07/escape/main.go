// Escape analysis: the compiler decides between stack and heap.
package main

import "fmt"

type Point struct {
	X, Y int
}

// sum's Point never outlives the call, so it can live on the stack.
func sum() int {
	p := Point{1, 2}
	return p.X + p.Y
}

// newPoint returns a pointer to a local variable. That is safe in Go:
// the compiler sees that p outlives the call and puts it on the heap.
func newPoint() *Point {
	p := Point{3, 4}
	return &p
}

func main() {
	fmt.Println(sum(), *newPoint())
}
