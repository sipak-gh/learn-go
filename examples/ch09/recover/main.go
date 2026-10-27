// recover stops a panic, inside a deferred function.
package main

import "fmt"

// safeDiv turns a panic into an ordinary error. The named result err
// lets the deferred function change what safeDiv returns.
func safeDiv(a, b int) (q int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("safeDiv(%d, %d): %v", a, b, r)
		}
	}()
	return a / b, nil // integer division by zero panics
}

// notDeferred calls recover directly. Outside a deferred call, recover
// does nothing and returns nil.
func notDeferred() {
	fmt.Println("recover() =", recover())
}

func main() {
	fmt.Println(safeDiv(7, 2))
	fmt.Println(safeDiv(7, 0))
	notDeferred()
	fmt.Println("main keeps going")
}
