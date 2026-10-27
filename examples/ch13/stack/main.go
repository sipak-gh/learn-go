package main

import "fmt"

// Stack is a last-in, first-out collection of any element type.
// The zero value is an empty stack, ready to use.
type Stack[T any] struct {
	items []T
}

// Methods use the type parameter from the receiver: Stack[T].
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top element. If the stack is empty it
// returns the zero value of T and false.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// A generic struct with two type parameters.
type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func main() {
	var words Stack[string] // instantiate: T = string
	words.Push("first")
	words.Push("second")
	fmt.Println(words.Len())
	for words.Len() > 0 {
		w, _ := words.Pop()
		fmt.Println(w)
	}
	v, ok := words.Pop()
	fmt.Printf("%q %v\n", v, ok)

	nums := &Stack[int]{}
	nums.Push(42)
	n, ok := nums.Pop()
	fmt.Println(n, ok)

	p := Pair[string, int]{"answer", 42}
	fmt.Printf("%+v %T\n", p, p)
}
