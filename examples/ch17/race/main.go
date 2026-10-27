package main

import "fmt"

func main() {
	count := 0
	done := make(chan bool)
	for range 4 {
		go func() {
			for range 100_000 {
				count++ // read, add, write: three steps, not one
			}
			done <- true
		}()
	}
	for range 4 {
		<-done
	}
	fmt.Println("count:", count, "(want 400000)")
}
