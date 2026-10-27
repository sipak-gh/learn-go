// Command growth shows how append grows a slice's capacity.
package main

import "fmt"

func main() {
	var s []int
	prevCap := cap(s)
	for i := range 2000 {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("len=%-4d cap %4d -> %4d\n", len(s), prevCap, cap(s))
			prevCap = cap(s)
		}
	}
}
