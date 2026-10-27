// A type switch picks a case by the dynamic type of an interface value.
package main

import (
	"errors"
	"fmt"
)

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("an integer: %v", x) // x is any here
	case string:
		return fmt.Sprintf("a string of %d bytes", len(x)) // x is a string
	case error:
		return "an error: " + x.Error()
	case fmt.Stringer:
		return "a Stringer: " + x.String()
	default:
		return fmt.Sprintf("something else: %T", x)
	}
}

type celsius float64

func (c celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func main() {
	for _, v := range []any{nil, 7, int64(8), "héllo", errors.New("boom"), celsius(21.5), 2.5} {
		fmt.Println(describe(v))
	}
}
