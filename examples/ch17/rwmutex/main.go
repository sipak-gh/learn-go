package main

import (
	"fmt"
	"strings"
	"sync"
)

// Cache remembers the upper-case form of strings. Many goroutines may
// read at once; a writer gets exclusive access.
type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func (c *Cache) Get(key string) string {
	c.mu.RLock() // shared: other readers may hold it too
	v, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return v
	}

	c.mu.Lock() // exclusive: waits for all readers to leave
	defer c.mu.Unlock()
	// Another goroutine may have filled it while we had no lock.
	if v, ok := c.items[key]; ok {
		return v
	}
	v = strings.ToUpper(key)
	c.items[key] = v
	return v
}

func main() {
	c := &Cache{items: make(map[string]string)}
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			c.Get("fox")
			c.Get("dog")
		})
	}
	wg.Wait()
	fmt.Println(len(c.items), c.Get("fox"), c.Get("dog"))
}
