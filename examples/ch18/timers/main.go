package main

import (
	"fmt"
	"time"
)

func main() {
	// time.After returns a channel that receives once, after d. It is
	// the simplest timeout in a select.
	work := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		work <- "done"
	}()
	select {
	case r := <-work:
		fmt.Println(r)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}

	// A Timer is the same thing, but you can stop or reset it.
	t := time.NewTimer(time.Hour)
	fmt.Println("stopped before firing:", t.Stop())
	t.Reset(10 * time.Millisecond)
	<-t.C
	fmt.Println("timer fired")

	// A Ticker sends on its channel again and again.
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for i := range 3 {
		<-tick.C
		fmt.Println("tick", i)
	}

	// AfterFunc runs a function in its own goroutine after d.
	fired := make(chan bool)
	time.AfterFunc(5*time.Millisecond, func() { fired <- true })
	fmt.Println("AfterFunc:", <-fired)
}
