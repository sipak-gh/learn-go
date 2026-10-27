package main

import "fmt"

func main() {
	ch := make(chan string) // unbuffered: a send waits for a receiver

	go func() {
		ch <- "ping" // blocks until main receives
	}()
	msg := <-ch // blocks until the goroutine sends
	fmt.Println("main got", msg)

	// The other direction. done lets main wait until the worker has
	// printed; without it, main could return first and end the program.
	done := make(chan bool)
	go func() {
		v := <-ch
		fmt.Println("worker got", v)
		done <- true
	}()
	ch <- "pong"
	<-done // the value doesn't matter; receiving is the signal
	fmt.Println("main done")
}
