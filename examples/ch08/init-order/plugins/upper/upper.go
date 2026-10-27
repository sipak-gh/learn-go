// Package upper registers an upper-casing transform when imported.
package upper

import (
	"fmt"
	"strings"

	"examples/ch08/init-order/registry"
)

func init() {
	fmt.Println("upper: init")
	registry.Register("upper", strings.ToUpper)
}
