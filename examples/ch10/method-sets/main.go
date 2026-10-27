// Pointer-receiver methods belong to *T, not T.
package main

import "fmt"

type Incrementer interface {
	Inc()
}

type Counter struct {
	n int
}

func (c *Counter) Inc() { c.n++ }

// Compile-time check: the build fails if *Counter stops implementing
// Incrementer. It costs nothing at run time.
var _ Incrementer = (*Counter)(nil)

func bump(i Incrementer, times int) {
	for range times {
		i.Inc()
	}
}

func main() {
	c := Counter{}
	bump(&c, 3) // pass a pointer: *Counter has Inc
	fmt.Println(c.n)
}
