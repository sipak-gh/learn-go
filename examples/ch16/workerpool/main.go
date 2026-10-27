package main

import (
	"fmt"
	"time"
)

type job struct {
	index int
	name  string
}

type result struct {
	index  int
	length int
}

func worker(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		time.Sleep(time.Duration(len(j.name)) * time.Millisecond) // pretend to work
		results <- result{j.index, len(j.name)}
	}
}

func main() {
	names := []string{"seek", "goroutine", "channel", "select", "go"}
	jobs := make(chan job)
	results := make(chan result)

	for range 3 { // a fixed pool of 3 workers
		go worker(jobs, results)
	}
	go func() {
		for i, n := range names {
			jobs <- job{i, n}
		}
		close(jobs)
	}()

	// Results arrive in any order. Each carries its index, so we can put
	// it in the right place.
	lengths := make([]int, len(names))
	for range names {
		r := <-results
		lengths[r.index] = r.length
	}
	for i, n := range names {
		fmt.Printf("%-10s %d\n", n, lengths[i])
	}
}
