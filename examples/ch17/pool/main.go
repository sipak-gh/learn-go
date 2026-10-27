package main

import (
	"bytes"
	"fmt"
	"sync"
)

// bufPool hands out reusable buffers. New makes one when the pool is
// empty.
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func render(name string, n int) string {
	buf := bufPool.Get().(*bytes.Buffer) // Get returns any
	defer bufPool.Put(buf)
	buf.Reset() // a reused buffer still holds old data
	fmt.Fprintf(buf, "%s=%d", name, n)
	return buf.String()
}

func main() {
	var wg sync.WaitGroup
	results := make([]string, 4)
	for i := range results {
		wg.Go(func() {
			results[i] = render("item", i)
		})
	}
	wg.Wait()
	fmt.Println(results)
}
