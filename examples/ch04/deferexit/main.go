// Command deferexit shows that os.Exit skips deferred calls.
package main

import (
	"fmt"
	"os"
)

func main() {
	defer fmt.Println("cleanup") // never runs
	fmt.Println("exiting")
	os.Exit(3)
}
