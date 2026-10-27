// Command deferuses shows common ways to use defer.
package main

import "fmt"

// double sets its named result in a deferred call, after the return
// statement has set result but before the caller receives it.
func double(n int) (result int) {
	defer func() { result *= 2 }()
	return n
}

// process shows the usual shape: acquire, defer the release, use.
func process(name string, size int) {
	fmt.Println("open", name)
	defer fmt.Println("close", name) // runs however process returns
	if size == 0 {
		fmt.Println(name, "is empty")
		return
	}
	fmt.Println("read", size, "bytes from", name)
}

func main() {
	fmt.Println(double(21))
	process("a.txt", 120)
	process("b.txt", 0)

	// Defers in a loop pile up until the FUNCTION returns,
	// not until the iteration ends.
	for i := range 3 {
		defer fmt.Println("loop defer", i)
	}
	fmt.Println("loop done")
}
