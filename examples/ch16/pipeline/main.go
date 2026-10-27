package main

import (
	"fmt"
	"strings"
)

// Each stage receives from the previous stage's channel, does one job,
// and sends on its own channel. Closing flows down the pipeline.

func emit(lines ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, l := range lines {
			out <- l
		}
	}()
	return out
}

func grep(in <-chan string, word string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for l := range in {
			if strings.Contains(l, word) {
				out <- l
			}
		}
	}()
	return out
}

func upper(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for l := range in {
			out <- strings.ToUpper(l)
		}
	}()
	return out
}

func main() {
	lines := emit("the quick fox", "a lazy dog", "fox runs")
	for l := range upper(grep(lines, "fox")) {
		fmt.Println(l)
	}
}
