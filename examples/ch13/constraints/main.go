package main

import (
	"cmp"
	"fmt"
	"strconv"
)

// Index works for any type that supports == (the comparable constraint).
func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		if v == x {
			return i
		}
	}
	return -1
}

// Max works for any ordered type: integers, floats, strings.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// A constraint can require methods as well as a type set.
type IntStringer interface {
	~int
	String() string
}

type Weekday int

func (d Weekday) String() string {
	return [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}[d]
}

// Label can use both + (from ~int) and String (from the method).
func Label[T IntStringer](v T) string {
	return v.String() + "#" + strconv.Itoa(int(v+1))
}

func main() {
	fmt.Println(Index([]string{"a", "b", "c"}, "c"))
	fmt.Println(Index([]float64{1.5, 2.5}, 9))
	fmt.Println(Max(3, 7), Max("apple", "banana"), Max(2.5, -1))
	fmt.Println(cmp.Compare(1, 2), cmp.Compare("b", "a"), cmp.Compare(3.0, 3.0))
	fmt.Println(Label(Weekday(2)))
}
