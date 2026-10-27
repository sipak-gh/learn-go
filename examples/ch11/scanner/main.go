package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	// Lines: the newline (and a "\r" before it) is removed.
	input := "alpha\nbeta gamma\r\n\ndelta" // no newline at the end
	sc := bufio.NewScanner(strings.NewReader(input))
	for n := 1; sc.Scan(); n++ {
		fmt.Printf("%d: %q\n", n, sc.Text())
	}

	// Words: a different split function.
	sc = bufio.NewScanner(strings.NewReader("  the quick\tbrown\n fox "))
	sc.Split(bufio.ScanWords)
	var words []string
	for sc.Scan() {
		words = append(words, sc.Text())
	}
	fmt.Printf("%q\n", words)

	// A line longer than the default limit (64 KiB) stops the scanner.
	long := strings.Repeat("x", 100_000) + "\nshort\n"
	sc = bufio.NewScanner(strings.NewReader(long))
	for sc.Scan() {
		fmt.Println("line length", len(sc.Text()))
	}
	fmt.Println("error:", sc.Err())

	// Buffer raises the limit: start with 64 KiB, grow up to 1 MiB.
	sc = bufio.NewScanner(strings.NewReader(long))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		fmt.Println("line length", len(sc.Text()))
	}
	fmt.Println("error:", sc.Err())
}
