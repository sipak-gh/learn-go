// Command methods previews methods: functions attached to a type.
package main

import "fmt"

type Celsius float64

// Fahrenheit is a method on Celsius. c is the receiver.
func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

func main() {
	t := Celsius(100)
	fmt.Println(t.Fahrenheit())
}
