package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var words = []string{"alpha", "beta", "gamma"}

// leakyFirst returns the first word in upper case. It leaks: nobody
// receives the second word, so the goroutine waits on its send forever.
func leakyFirst() string {
	ch := make(chan string)
	go func() {
		for _, w := range words {
			ch <- strings.ToUpper(w)
		}
		close(ch)
	}()
	return <-ch
}

// fixedFirst closes done when it returns. The goroutine selects on done
// as well as on its send, so it can always stop.
func fixedFirst() string {
	ch := make(chan string)
	done := make(chan struct{})
	defer close(done)
	go func() {
		defer close(ch)
		for _, w := range words {
			select {
			case ch <- strings.ToUpper(w):
			case <-done:
				return
			}
		}
	}()
	return <-ch
}

func main() {
	fmt.Println("goroutines at start:", runtime.NumGoroutine())
	for range 3 {
		leakyFirst()
	}
	time.Sleep(10 * time.Millisecond) // give goroutines time to finish
	fmt.Println("after leakyFirst x3:", runtime.NumGoroutine())
	for range 3 {
		fixedFirst()
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println("after fixedFirst x3:", runtime.NumGoroutine())
}
