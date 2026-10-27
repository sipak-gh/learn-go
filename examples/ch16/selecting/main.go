package main

import "fmt"

func main() {
	ch := make(chan int, 1)

	// With a default case, select never blocks: if no other case is
	// ready, default runs.
	select {
	case v := <-ch:
		fmt.Println("received", v)
	default:
		fmt.Println("nothing to receive")
	}

	ch <- 1 // fills the buffer
	select {
	case ch <- 2:
		fmt.Println("sent 2")
	default:
		fmt.Println("buffer full, 2 dropped")
	}

	// When several cases are ready, select picks one at random.
	a := make(chan string, 1)
	b := make(chan string, 1)
	counts := map[string]int{}
	for range 1000 {
		a <- "a"
		b <- "b"
		select {
		case v := <-a:
			counts[v]++
			<-b // empty the other channel for the next round
		case v := <-b:
			counts[v]++
			<-a
		}
	}
	fmt.Println(counts)
}
