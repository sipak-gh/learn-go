// Command conversions shows explicit type conversions.
package main

import "fmt"

func main() {
	count := 3
	price := 1.25

	// Go never converts between numeric types for you.
	total := float64(count) * price
	fmt.Println(total)

	// float to int truncates toward zero.
	a, b := 2.99, -2.99
	fmt.Println(int(a), int(b))

	// Converting to a smaller type keeps only the low bits.
	big := 300
	fmt.Println(uint8(big)) // 300 - 256

	// Converting an integer to a string gives the character with that
	// code point, not the digits. Use a rune to make that clear.
	fmt.Println(string(rune(65)))
	// To get the digits, format the number.
	fmt.Println(fmt.Sprint(65) + "!")
}
