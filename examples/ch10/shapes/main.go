// An interface lists methods. Any type with those methods satisfies it.
package main

import (
	"fmt"
	"math"
)

// A Shape is anything with an area and a perimeter.
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct {
	W, H float64
}

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Circle struct {
	R float64
}

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

// describe works with any Shape, including types written later.
func describe(s Shape) {
	fmt.Printf("%T %v: area %.2f, perimeter %.2f\n", s, s, s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}, Rect{W: 10, H: 0.5}}
	total := 0.0
	for _, s := range shapes {
		describe(s)
		total += s.Area()
	}
	fmt.Printf("total area %.2f\n", total)
}
