// Command typedefs shows type definitions and type aliases.
package main

import "fmt"

// Celsius and Fahrenheit are new, distinct types.
type Celsius float64
type Fahrenheit float64

// Number is an alias: just another name for float64.
type Number = float64

func main() {
	boiling := Celsius(100)
	f := Fahrenheit(boiling*9/5 + 32) // explicit conversion required
	fmt.Println(boiling, f)
	fmt.Printf("%T %T\n", boiling, f)

	var n Number = 1.5
	var x float64 = n // fine: Number and float64 are the same type
	fmt.Printf("%T %v\n", n, x)
}
