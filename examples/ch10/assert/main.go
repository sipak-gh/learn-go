// A type assertion gets the concrete value out of an interface.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var v any = "gopher"

	s := v.(string) // v holds a string, so this works
	fmt.Println(s, len(s))

	n, ok := v.(int) // comma-ok: no panic, ok reports success
	fmt.Println(n, ok)

	// Asserting to an interface asks "does it have these methods?"
	var w io.Writer = os.Stdout
	if f, ok := w.(*os.File); ok {
		fmt.Println("a file:", f.Name())
	}
	w = &strings.Builder{}
	_, isStringer := w.(fmt.Stringer)
	fmt.Println("Builder is a Stringer:", isStringer)

	fmt.Println(v.(int)) // one-result form panics on failure
}
