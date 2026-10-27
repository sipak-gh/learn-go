// Command imports uses three packages from the standard library.
package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("gopher"))
	fmt.Println(math.Sqrt(2))
	fmt.Printf("pi is about %.2f\n", math.Pi)
}
