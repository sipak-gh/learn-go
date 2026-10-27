// errors.As and errors.AsType find an error of a given type in a chain.
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// A RangeError reports a value outside the allowed range.
type RangeError struct {
	Value, Min, Max int
}

func (e *RangeError) Error() string {
	return fmt.Sprintf("%d is not between %d and %d", e.Value, e.Min, e.Max)
}

func parsePercent(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse percent: %w", err)
	}
	if n < 0 || n > 100 {
		return 0, fmt.Errorf("parse percent: %w", &RangeError{n, 0, 100})
	}
	return n, nil
}

func main() {
	for _, s := range []string{"150", "abc"} {
		_, err := parsePercent(s)
		fmt.Println(err)

		// errors.As: declare a variable of the target type, pass a pointer.
		var re *RangeError
		if errors.As(err, &re) {
			fmt.Println("  too far by", re.Value-re.Max)
		}

		// errors.AsType (Go 1.26): the type goes in brackets, the result
		// comes back. It reads like a comma-ok type check.
		if ne, ok := errors.AsType[*strconv.NumError](err); ok {
			fmt.Printf("  bad number %q from %s\n", ne.Num, ne.Func)
		}
	}
}
