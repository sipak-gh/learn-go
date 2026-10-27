// An interface value holds a dynamic type and a value.
package main

import (
	"fmt"
	"strings"
)

type MyErr struct{}

func (*MyErr) Error() string { return "my error" }

// mayFail returns a nil *MyErr as an error. That error is not nil!
func mayFail() error {
	var p *MyErr // nil pointer
	return p     // interface value: (type *MyErr, value nil)
}

func main() {
	var s fmt.Stringer // nil interface: no type, no value
	fmt.Printf("%T %v %v\n", s, s, s == nil)

	var b *strings.Builder // nil pointer...
	s = b                  // ...stored in an interface
	fmt.Printf("%T %v\n", s, s == nil)

	err := mayFail()
	fmt.Printf("%T %v\n", err, err == nil)

	// Interface values compare equal if both type and value are equal.
	var x, y any = 1, 1
	var z any = int64(1)
	fmt.Println(x == y, x == z)
}
