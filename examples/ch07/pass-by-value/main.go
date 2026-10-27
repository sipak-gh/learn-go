// Go passes everything by value: arguments are copies.
package main

import "fmt"

type Counter struct {
	Name string
	N    int
}

func incCopy(c Counter) { c.N++ } // changes a copy; the caller sees nothing

func incPtr(c *Counter) { c.N++ } // the pointer is a copy, but it points at the original

func setFirst(s []int) { s[0] = 99 } // a slice header holds a pointer to its array

func grow(s []int) { s = append(s, 4) } // changes only the local copy of the header

func main() {
	c := Counter{Name: "hits"}
	incCopy(c)
	fmt.Println(c.N)
	incPtr(&c)
	fmt.Println(c.N)

	nums := []int{1, 2, 3}
	setFirst(nums)
	grow(nums)
	fmt.Println(nums)

	m := map[string]int{}
	func(m map[string]int) { m["k"] = 1 }(m) // maps are references too
	fmt.Println(m)
}
