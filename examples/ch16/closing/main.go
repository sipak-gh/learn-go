package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i * i
		}
		close(ch) // no more values: ends the receiver's range loop
	}()

	// range receives until the channel is closed and empty.
	for v := range ch {
		fmt.Println("got", v)
	}

	// Receiving from a closed channel never blocks. It returns the zero
	// value; the second result, ok, is false to tell you why.
	v, ok := <-ch
	fmt.Println(v, ok)

	buf := make(chan string, 2)
	buf <- "a"
	close(buf) // values already in the buffer can still be received
	s, ok := <-buf
	fmt.Printf("%q %v\n", s, ok)
	s, ok = <-buf
	fmt.Printf("%q %v\n", s, ok)
}
