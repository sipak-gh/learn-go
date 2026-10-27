// Building strings efficiently with strings.Builder.
package main

import (
	"fmt"
	"strings"
)

func main() {
	var b strings.Builder // the zero value is ready to use
	for i := 3; i > 0; i-- {
		fmt.Fprintf(&b, "%d...", i) // Fprintf can write into a Builder
	}
	b.WriteString("liftoff")
	b.WriteByte('!')
	b.WriteRune('🚀')
	fmt.Println(b.String())
	fmt.Println(b.Len(), "bytes")

	// The same with +=. Each += copies everything built so far, so a
	// loop of n steps does O(n²) work. Fine for a few pieces, slow for many.
	s := ""
	for i := 3; i > 0; i-- {
		s += fmt.Sprint(i) + "..."
	}
	fmt.Println(s + "liftoff!🚀")
}
