// Errors are ordinary values, returned as the last result.
package main

import (
	"fmt"
	"strconv"
)

// half returns n/2, or an error if n is odd.
func half(n int) (int, error) {
	if n%2 != 0 {
		return 0, fmt.Errorf("%d is odd", n)
	}
	return n / 2, nil
}

func main() {
	for _, s := range []string{"42", "7", "seven"} {
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		h, err := half(n)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(s, "halved is", h)
	}

	_, err := strconv.Atoi("x")
	fmt.Printf("%T\n", err)  // the concrete type inside the error
	fmt.Println(err.Error()) // the message
}
