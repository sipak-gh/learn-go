// Package reverse registers a string-reversing transform when imported.
package reverse

import (
	"fmt"
	"slices"

	"examples/ch08/init-order/registry"
)

func init() {
	fmt.Println("reverse: init")
	registry.Register("reverse", func(s string) string {
		r := []rune(s)
		slices.Reverse(r)
		return string(r)
	})
}
