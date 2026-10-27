package main

import (
	"fmt"
	"unsafe"
)

type loose struct {
	a bool  // 1 byte, then 7 bytes of padding
	b int64 // 8 bytes, must start at a multiple of 8
	c bool  // 1 byte, then 7 bytes of padding
}

type tight struct {
	b int64
	a bool
	c bool // a and c share the padding at the end
}

func main() {
	fmt.Println("int:   ", unsafe.Sizeof(int(0)))
	fmt.Println("string:", unsafe.Sizeof(""))
	fmt.Println("slice: ", unsafe.Sizeof([]int{}))
	fmt.Println("loose: ", unsafe.Sizeof(loose{}), "offset of b:", unsafe.Offsetof(loose{}.b))
	fmt.Println("tight: ", unsafe.Sizeof(tight{}), "offset of a:", unsafe.Offsetof(tight{}.a))
}
