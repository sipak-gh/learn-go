// Command internal-demo imports a package from its own internal/ directory.
package main

import (
	"fmt"

	"examples/ch08/internal-demo/internal/secret"
)

func main() {
	fmt.Println("the answer is", secret.Answer())
}
