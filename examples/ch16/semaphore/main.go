package main

import (
	"fmt"
	"time"
)

func main() {
	// A buffered channel with capacity 2 works as a semaphore: a send
	// takes a slot, a receive gives it back, and a third send blocks
	// until a slot is free.
	sem := make(chan struct{}, 2)
	inFlight := make(chan int)

	for i := range 6 {
		go func() {
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release
			inFlight <- len(sem)     // how many slots are taken right now
			time.Sleep(time.Duration(10+i) * time.Millisecond)
		}()
	}

	most := 0
	for range 6 {
		most = max(most, <-inFlight)
	}
	fmt.Println("most at once:", most)
}
