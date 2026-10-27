// Command init-order shows when package initialization happens.
package main

import (
	"fmt"

	"examples/ch08/init-order/registry"

	// Blank imports: we never name these packages, we only want their
	// init functions to run and register their transforms.
	_ "examples/ch08/init-order/plugins/reverse"
	_ "examples/ch08/init-order/plugins/upper"
)

var greeting = makeGreeting()

func makeGreeting() string {
	fmt.Println("main: variables initialized")
	return "hello, gopher"
}

func init() {
	fmt.Println("main: first init")
}

func init() {
	fmt.Println("main: second init") // a file may have several init funcs
}

func main() {
	fmt.Println("main: main starts")
	for _, name := range registry.Names() {
		fmt.Printf("%s(%q) = %q\n", name, greeting, registry.Apply(name, greeting))
	}
}
