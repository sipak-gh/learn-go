package main

import "fmt"

// Without generics you write one function per type.
func SumInts(nums []int) int {
	var total int
	for _, n := range nums {
		total += n
	}
	return total
}

// Number is a constraint: the set of types Sum accepts.
// ~float64 means "float64 or any type whose underlying type is float64".
type Number interface {
	~int | ~int64 | ~float64
}

// Sum works for any type in Number's type set.
func Sum[T Number](nums []T) T {
	var total T // the zero value of T
	for _, n := range nums {
		total += n
	}
	return total
}

type Celsius float64

func main() {
	fmt.Println(SumInts([]int{1, 2, 3}))
	fmt.Println(Sum([]int{1, 2, 3}))         // T inferred as int
	fmt.Println(Sum([]float64{1.5, 2.25}))   // T inferred as float64
	fmt.Println(Sum([]Celsius{20.5, 21}))    // allowed thanks to ~float64
	fmt.Println(Sum[int64](nil))             // nothing to infer from: say it
	fmt.Printf("%T\n", Sum([]Celsius{1, 2})) // the result keeps the type
}
