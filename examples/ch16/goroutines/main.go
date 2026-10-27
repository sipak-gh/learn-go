package main

import (
	"fmt"
	"runtime"
	"time"
)

func say(word string) {
	for i := range 3 {
		fmt.Println(word, i)
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	fmt.Println("CPUs:", runtime.NumCPU(), "GOMAXPROCS:", runtime.GOMAXPROCS(0))

	go say("goroutine") // runs at the same time as main
	say("main")

	fmt.Println("goroutines now:", runtime.NumGoroutine())

	go fmt.Println("this may never print")
	// main returns here. The program ends at once, and any goroutines
	// still running are stopped mid-flight.
}
