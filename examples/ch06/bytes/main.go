// The bytes package mirrors strings for []byte.
package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	data := []byte("name: gopher\nrole: mascot\n")

	// Same functions as strings, but for byte slices.
	fmt.Println(bytes.Contains(data, []byte("gopher")))
	fmt.Println(bytes.Count(data, []byte("\n")))
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		k, v, _ := bytes.Cut(line, []byte(": "))
		fmt.Printf("%s=%s\n", k, v) // %s prints a []byte as text
	}
	fmt.Println(bytes.Equal([]byte("a"), []byte("a"))) // slices can't use ==

	// Unlike strings, byte slices are mutable.
	word := []byte("gopher")
	word[0] = 'G'
	fmt.Println(string(word))

	// bytes.Buffer is a growable buffer you can write to and read from.
	var buf bytes.Buffer
	buf.WriteString("hello ")
	fmt.Fprintf(&buf, "%d gophers\n", 3)
	fmt.Print(buf.String())
	buf.WriteTo(os.Stdout) // drains the buffer
	fmt.Println(buf.Len()) // 0: everything was read
}
