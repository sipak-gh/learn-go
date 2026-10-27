package main

import "fmt"

// produce can only send on out.
func produce(out chan<- string) {
	for _, w := range []string{"red", "green", "blue"} {
		out <- w
	}
	close(out) // closing is a send-side operation
}

// consume can only receive from in.
func consume(in <-chan string) {
	for w := range in {
		fmt.Println(w)
	}
}

func main() {
	ch := make(chan string) // bidirectional
	go produce(ch)          // converted to chan<- string automatically
	consume(ch)             // converted to <-chan string automatically
}
