package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush() // write whatever is left when main returns

	fmt.Fprintln(w, "1. written to the buffer first")
	fmt.Println("2. printed directly; buffered bytes:", w.Buffered())

	lost := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(lost, "3. never flushed, so never printed")
}
