// A Must function panics instead of returning an error.
package main

import (
	"fmt"
	"strconv"
)

// mustAtoi is for values that are known to be valid, such as constants
// in the source code. A failure means the program itself is wrong.
func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

var defaultPort = mustAtoi("8080") // checked once, at startup

func main() {
	fmt.Println("port", defaultPort)
	fmt.Println(mustAtoi("80x"))
}
