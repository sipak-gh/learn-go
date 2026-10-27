// Command numbers tours Go's numeric types.
package main

import (
	"fmt"
	"math"
)

func main() {
	// Sized integers: int8, int16, int32, int64 and unsigned uint8 ... uint64.
	fmt.Println("int8:  ", math.MinInt8, "to", math.MaxInt8)
	fmt.Println("uint8: ", 0, "to", math.MaxUint8)
	fmt.Println("int64: ", math.MinInt64, "to", math.MaxInt64)
	// int and uint are 64 bits on 64-bit systems, 32 bits on 32-bit ones.
	fmt.Println("int:   ", math.MinInt, "to", math.MaxInt)

	// Integer overflow wraps around silently at run time.
	var small int8 = 127
	small++
	fmt.Println("127 + 1 as int8 =", small)

	// Integer division truncates toward zero; % is the remainder.
	fmt.Println(7/2, -7/2, 7%3, -7%3)
	// Floating-point division. Floats are binary, so 0.1 is not exact.
	x, y := 0.1, 0.2
	fmt.Println(7.0/2, x+y)

	// byte is an alias for uint8, rune is an alias for int32.
	var b byte = 'A'
	var r rune = 'é'
	fmt.Println(b, r)
	fmt.Printf("%c %c\n", b, r)

	// float32 has about 7 significant digits, float64 about 16.
	var f32 float32 = 1.0 / 3
	var f64 float64 = 1.0 / 3
	fmt.Println(f32, f64)

	// Complex numbers are built in.
	c := complex(1, 2) // complex128
	fmt.Println(c*c, real(c), imag(c))
}
