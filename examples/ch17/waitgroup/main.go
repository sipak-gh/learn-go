package main

import (
	"fmt"
	"sync"
)

func main() {
	words := []string{"alpha", "beta", "gamma"}
	lengths := make([]int, len(words))

	// The classic way: Add before starting, Done when finished.
	var wg sync.WaitGroup
	for i, w := range words {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lengths[i] = len(w) // each goroutine writes its own element
		}()
	}
	wg.Wait() // blocks until the counter is back to zero
	fmt.Println(lengths)

	// Since Go 1.25, wg.Go does the Add and Done for you.
	upper := make([]string, len(words))
	for i, w := range words {
		wg.Go(func() {
			upper[i] = fmt.Sprintf("%s!", w)
		})
	}
	wg.Wait()
	fmt.Println(upper)
}
