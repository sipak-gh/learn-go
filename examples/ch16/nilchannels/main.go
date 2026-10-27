package main

import "fmt"

func send(values ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, v := range values {
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func main() {
	a, b := send(1, 2, 3), send(10, 20)
	sum := 0
	// Loop until both channels are closed. A receive from a nil channel
	// blocks forever, so setting a finished channel to nil switches its
	// case off. (A closed channel would be ready all the time instead.)
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			sum += v
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			sum += v
		}
	}
	fmt.Println("sum:", sum)
}
