// Converting between strings and other types with strconv.
package main

import (
	"fmt"
	"strconv"
)

func main() {
	// Strings to numbers. Every Parse function also returns an error.
	n, err := strconv.Atoi("42")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("forty-two")
	fmt.Println(err)

	big, _ := strconv.ParseInt("-ff", 16, 64) // base 16, fits in 64 bits
	f, _ := strconv.ParseFloat("2.5e3", 64)
	b, _ := strconv.ParseBool("true") // also "1", "t", "T", "TRUE", "True"
	fmt.Println(big, f, b)

	_, err = strconv.ParseInt("300", 10, 8) // too big for 8 bits
	fmt.Println(err)

	// Numbers to strings.
	s := strconv.Itoa(99) + " bottles"
	fmt.Println(s)
	fmt.Println(strconv.FormatInt(255, 2), strconv.FormatFloat(3.14159, 'f', 2, 64))

	// string(n) does NOT format a number: it makes the rune with that code.
	fmt.Println(string(rune(65)), strconv.Itoa(65))

	// Quote adds quotes and escapes; Unquote reverses it.
	q := strconv.Quote("tab\there \"hi\"")
	fmt.Println(q)
	u, _ := strconv.Unquote(q)
	fmt.Println(u)
}
