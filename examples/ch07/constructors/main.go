// Constructor functions and useful zero values.
package main

import (
	"fmt"
	"strings"
)

// A Stack of strings. The zero value is an empty stack, ready to use,
// so it needs no constructor.
type Stack struct {
	items []string
}

func (s *Stack) Push(v string) { s.items = append(s.items, v) }

func (s *Stack) Pop() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// A Ring keeps the last few strings added. It needs a size, so it has a
// constructor that sets it up. By convention it is named NewRing.
type Ring struct {
	buf  []string
	next int
}

func NewRing(size int) *Ring {
	if size < 1 {
		size = 1
	}
	return &Ring{buf: make([]string, size)}
}

func (r *Ring) Add(v string) {
	r.buf[r.next] = v
	r.next = (r.next + 1) % len(r.buf)
}

func (r *Ring) String() string { return strings.Join(r.buf, ",") }

func main() {
	var s Stack // no setup needed
	s.Push("a")
	s.Push("b")
	v, ok := s.Pop()
	fmt.Println(v, ok)

	r := NewRing(3)
	for _, w := range []string{"a", "b", "c", "d"} {
		r.Add(w)
	}
	fmt.Println(r.String())
}
