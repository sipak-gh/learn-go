package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	r := strings.NewReader("Hello, Reader!")
	buf := make([]byte, 8)
	for {
		// Read fills up to len(buf) bytes and reports how many it got.
		n, err := r.Read(buf)
		fmt.Printf("n=%d err=%v data=%q\n", n, err, buf[:n])
		if err == io.EOF {
			break
		}
	}
}
