package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// looksBinary reports whether the start of br contains a NUL byte.
// Peek looks ahead without consuming, so br can still be read in full.
func looksBinary(br *bufio.Reader) (bool, error) {
	head, err := br.Peek(512)
	if err != nil && err != io.EOF {
		return false, err // io.EOF only means the input is shorter than 512
	}
	return bytes.IndexByte(head, 0) >= 0, nil
}

func main() {
	inputs := []string{
		"plain text\nline two\n",
		"GIF89a\x00\x01\x02",
		"",
	}
	for _, in := range inputs {
		br := bufio.NewReader(strings.NewReader(in))
		binary, err := looksBinary(br)
		rest, _ := io.ReadAll(br)
		fmt.Printf("binary=%-5v err=%v  bytes left to read: %d\n", binary, err, len(rest))
	}
}
