package main

import (
	"fmt"
	"sync"
)

// Counter is safe for concurrent use. The mutex guards counts: every
// access to counts happens between Lock and Unlock.
type Counter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *Counter) Add(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[string]int)
	}
	c.counts[key]++
}

func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

func main() {
	var c Counter // the zero Mutex is unlocked and ready to use
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			c.Add("fox")
		})
	}
	wg.Wait()
	fmt.Println("fox:", c.Get("fox"))
}
