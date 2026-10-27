// Value receivers get a copy; pointer receivers share the original.
package main

import "fmt"

type Counter struct {
	n int
}

// IncBroken has a value receiver, so it increments a copy. A classic bug.
func (c Counter) IncBroken() { c.n++ }

// Inc has a pointer receiver, so it increments the caller's Counter.
func (c *Counter) Inc() { c.n++ }

// A List is a linked list. A nil *List is a valid empty list.
type List struct {
	Val  int
	Next *List
}

// Len works even when l is nil: a pointer method may be called on nil.
func (l *List) Len() int {
	if l == nil {
		return 0
	}
	return 1 + l.Next.Len()
}

func main() {
	var c Counter
	c.IncBroken()
	fmt.Println(c.n)
	c.Inc()
	c.Inc()
	fmt.Println(c.n)

	var empty *List
	l := &List{1, &List{2, &List{3, nil}}}
	fmt.Println(empty.Len(), l.Len())
}
