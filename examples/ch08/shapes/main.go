// Command shapes uses the geometry package from the same module.
package main

import (
	"fmt"

	"examples/ch08/shapes/geometry" // module path + directory
)

func main() {
	c := geometry.NewCircle(2, "coin")
	fmt.Printf("%s: %.2f %s²\n", c.Label(), c.Area(), geometry.Unit)

	var d geometry.Circle // zero value; label is empty
	d.Radius = 1
	fmt.Printf("%s: %.2f %s²\n", d.Label(), d.Area(), geometry.Unit)
}
