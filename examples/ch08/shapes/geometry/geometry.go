// Package geometry computes the areas of simple shapes.
package geometry

import "math"

// Unit is the length unit used in descriptions.
const Unit = "cm"

// A Circle is a round shape centered on the origin.
type Circle struct {
	Radius float64 // exported: other packages can read and set it
	label  string  // unexported: only package geometry can see it
}

// NewCircle returns a Circle with the given radius and label.
func NewCircle(radius float64, label string) Circle {
	return Circle{Radius: radius, label: label}
}

// Area returns the area of c.
func (c Circle) Area() float64 {
	return math.Pi * square(c.Radius)
}

// Label returns the circle's label, or "circle" if it has none.
func (c Circle) Label() string {
	if c.label == "" {
		return "circle"
	}
	return c.label
}

// square is unexported: a helper for this package only.
func square(x float64) float64 {
	return x * x
}
