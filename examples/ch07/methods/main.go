// Methods: functions with a receiver.
package main

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

// Dist has a value receiver: it gets a copy of the Point.
func (p Point) Dist(q Point) float64 {
	return math.Hypot(q.X-p.X, q.Y-p.Y)
}

// Scale has a pointer receiver: it changes the Point it is called on.
func (p *Point) Scale(f float64) {
	p.X *= f
	p.Y *= f
}

// Methods work on any named type defined in your package, not just structs.
type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

func main() {
	p := Point{3, 4}
	fmt.Println(p.Dist(Point{}))

	p.Scale(2) // Go takes &p for you because p is a variable
	fmt.Println(p)

	pp := &Point{1, 1}
	fmt.Println(pp.Dist(Point{})) // and follows the pointer for value methods

	fmt.Println(Celsius(100).Fahrenheit())
}
