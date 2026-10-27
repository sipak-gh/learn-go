package main

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	s := []string{"c", "a", "b"}
	for i, v := range slices.All(s) {
		fmt.Printf("%d=%s ", i, v)
	}
	fmt.Println("(All)")
	for i, v := range slices.Backward(s) {
		fmt.Printf("%d=%s ", i, v)
	}
	fmt.Println("(Backward)")
	fmt.Println(slices.Sorted(slices.Values(s)))
	fmt.Println(slices.Collect(slices.Values(s)))

	m := map[string]int{"go": 2009, "c": 1972, "zig": 2016}
	fmt.Println(slices.Sorted(maps.Keys(m)))
	fmt.Println(slices.Sorted(maps.Values(m)))
	clone := maps.Collect(maps.All(m))
	fmt.Println(len(clone))

	text := "one\ntwo\nthree"
	fmt.Printf("%q\n", slices.Collect(strings.Lines(text))) // lines keep their "\n"
	fmt.Printf("%q\n", slices.Collect(strings.SplitSeq("a,b,,c", ",")))
	for f := range strings.FieldsSeq("  lots   of\tspace ") {
		fmt.Println(f)
	}
	n := 0
	for range bytes.Lines([]byte("x\ny\n")) {
		n++
	}
	fmt.Println(n, "lines")
}
