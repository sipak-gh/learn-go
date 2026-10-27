// An external test package: it can only use reverse's exported names,
// just like any other user of the package.
package reverse_test

import (
	"fmt"

	"examples/ch19/reverse"
)

func ExampleString() {
	fmt.Println(reverse.String("stressed"))
	// Output: desserts
}
