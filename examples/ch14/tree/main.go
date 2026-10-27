package main

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
)

// Tree is a binary search tree of ordered values.
type Tree[T cmp.Ordered] struct {
	root *node[T]
}

type node[T cmp.Ordered] struct {
	val         T
	left, right *node[T]
}

func (t *Tree[T]) Insert(v T) {
	p := &t.root
	for *p != nil {
		if v < (*p).val {
			p = &(*p).left
		} else {
			p = &(*p).right
		}
	}
	*p = &node[T]{val: v}
}

// All returns an iterator over the values in sorted order.
func (t *Tree[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		t.root.walk(yield)
	}
}

// walk visits n's subtree in order. It returns false if yield did,
// so every level of the recursion stops at once.
func (n *node[T]) walk(yield func(T) bool) bool {
	if n == nil {
		return true
	}
	return n.left.walk(yield) && yield(n.val) && n.right.walk(yield)
}

func main() {
	var t Tree[string]
	for _, w := range []string{"m", "c", "x", "a", "e", "z"} {
		t.Insert(w)
	}
	fmt.Println(slices.Collect(t.All()))

	var small []string
	for v := range t.All() {
		if v > "d" {
			break // stops the walk deep inside the recursion
		}
		small = append(small, v)
	}
	fmt.Println(small)
}
