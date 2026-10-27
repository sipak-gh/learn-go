// any is the empty interface: every type satisfies it.
package main

import "fmt"

func main() {
	var v any // same as interface{}
	fmt.Printf("%v %T\n", v, v)

	v = 42
	fmt.Printf("%v %T\n", v, v)
	v = "hello"
	fmt.Printf("%v %T\n", v, v)
	v = []int{1, 2}
	fmt.Printf("%v %T\n", v, v)

	// fmt.Println takes ...any, which is why it accepts anything.
	things := []any{1, "two", 3.0, nil, struct{ X int }{4}}
	fmt.Println(things...)
}
