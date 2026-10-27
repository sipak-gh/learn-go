// Pointers: &, *, new, and nil.
package main

import "fmt"

type Point struct {
	X, Y int
}

func main() {
	x := 1
	p := &x         // p has type *int and points at x
	fmt.Println(*p) // *p reads the value p points to
	*p = 2          // and writes it
	fmt.Println(x)  // x changed

	// new(T) allocates a zero T and returns a pointer to it.
	n := new(int)
	*n += 10
	fmt.Println(*n)

	// &T{...} is the usual way to get a pointer to a new struct.
	pt := &Point{X: 1, Y: 2}
	pt.X = 10 // shorthand for (*pt).X: Go follows the pointer for you
	fmt.Println(pt.X, (*pt).Y, *pt)

	// You can point at a field or a slice element too.
	py := &pt.Y
	*py = 20
	nums := []int{1, 2, 3}
	second := &nums[1]
	*second = 200
	fmt.Println(*pt, nums)

	// The zero value of a pointer is nil: it points at nothing.
	var none *Point
	fmt.Println(none == nil, none)

	// Two pointers are equal if they point at the same variable.
	a, b := Point{}, Point{}
	fmt.Println(&a == &a, &a == &b, a == b)
}
