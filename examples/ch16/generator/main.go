package main

import "fmt"

// count returns a channel that delivers from, from+1, ..., to and is
// then closed. The goroutine inside is the "generator".
func count(from, to int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := from; i <= to; i++ {
			ch <- i
		}
	}()
	return ch
}

func main() {
	for n := range count(1, 5) {
		fmt.Print(n, " ")
	}
	fmt.Println()
}
