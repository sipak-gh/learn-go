package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var (
	loadOnce sync.Once
	words    []string
)

func loadWords() {
	fmt.Println("loading words")
	words = strings.Fields("alpha beta gamma")
}

// OnceFunc returns a function that runs its argument only the first
// time it is called.
var greet = sync.OnceFunc(func() { fmt.Println("hello, once") })

// OnceValue caches a result: the function runs once, and every call
// returns the same value.
var answer = sync.OnceValue(func() int {
	fmt.Println("computing the answer")
	return 6 * 7
})

// OnceValues does the same for a (value, error) pair.
var port = sync.OnceValues(func() (int, error) {
	fmt.Println("parsing the port")
	return strconv.Atoi("8080")
})

func main() {
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			loadOnce.Do(loadWords) // every goroutine waits for the one call
			greet()
		})
	}
	wg.Wait()
	fmt.Println(words)

	fmt.Println(answer(), answer())
	p, err := port()
	fmt.Println(p, err)
	p, err = port()
	fmt.Println(p, err)
}
