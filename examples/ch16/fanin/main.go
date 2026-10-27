package main

import (
	"fmt"
	"slices"
)

// square reads numbers from in and sends their squares on out. Several
// copies of it read the same in channel: that is fan-out.
func square(id int, in <-chan int, out chan<- string, done chan<- bool) {
	for n := range in {
		out <- fmt.Sprintf("%3d (worker %d)", n*n, id)
	}
	done <- true
}

func main() {
	in := make(chan int)
	out := make(chan string)
	done := make(chan bool)

	const workers = 3
	for id := range workers {
		go square(id, in, out, done)
	}
	go func() {
		for n := 1; n <= 6; n++ {
			in <- n
		}
		close(in)
	}()
	// Fan-in: all workers send on one channel. Close it after every
	// worker has said it's done. (Chapter 17's sync.WaitGroup does this
	// counting for you.)
	go func() {
		for range workers {
			<-done
		}
		close(out)
	}()

	var results []string
	for r := range out {
		results = append(results, r[:3]) // drop the worker id: it varies
	}
	slices.Sort(results)
	fmt.Println(results)
}
