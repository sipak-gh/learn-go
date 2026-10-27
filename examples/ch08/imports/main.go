// Import aliases and dot imports.
package main

import (
	crand "crypto/rand" // alias: two packages are both named "rand"
	"fmt"
	mrand "math/rand/v2" // the "/v2" is part of the path, not the name
	. "strings"          // dot import: use ToUpper without "strings." (avoid this)
)

func main() {
	b := make([]byte, 4)
	crand.Read(b) // cryptographically secure random bytes
	fmt.Println(len(b), "secure random bytes")

	n := mrand.IntN(10) // fast pseudo-random numbers
	fmt.Println(n >= 0 && n < 10)

	fmt.Println(ToUpper("dot imports hide where names come from"))
}
