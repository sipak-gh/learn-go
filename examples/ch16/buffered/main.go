package main

import "fmt"

func main() {
	ch := make(chan int, 3) // room for 3 values

	// No receiver yet, but these sends don't block: the buffer has room.
	ch <- 1
	ch <- 2
	fmt.Println("len", len(ch), "cap", cap(ch))

	fmt.Println(<-ch) // values come out in the order they went in
	ch <- 3
	ch <- 4
	fmt.Println("len", len(ch), "cap", cap(ch))
	fmt.Println(<-ch, <-ch, <-ch)
}
