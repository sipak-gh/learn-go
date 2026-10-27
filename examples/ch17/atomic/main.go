package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type config struct {
	name  string
	limit int
}

func main() {
	var hits atomic.Int64 // the zero value is 0 and ready to use
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 100 {
				hits.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("hits:", hits.Load())

	// CompareAndSwap changes the value only if it still holds what you
	// expect. Exactly one goroutine wins the race to become "leader".
	var leader atomic.Int32
	var winners atomic.Int32
	for id := int32(1); id <= 10; id++ {
		wg.Go(func() {
			if leader.CompareAndSwap(0, id) {
				winners.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("winners:", winners.Load(), "leader set:", leader.Load() != 0)

	// atomic.Bool and atomic.Pointer[T] work the same way.
	var ready atomic.Bool
	var cfg atomic.Pointer[config]
	cfg.Store(&config{"default", 10})
	ready.Store(true)
	old := cfg.Swap(&config{"tuned", 20}) // replace the whole value at once
	fmt.Println(ready.Load(), old.name, cfg.Load().name, cfg.Load().limit)
}
