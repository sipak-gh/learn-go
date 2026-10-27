// io.Reader and io.Writer: one method each, used everywhere.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// greet writes to any io.Writer: a file, the terminal, a buffer, a
// network connection, or a type you write yourself.
func greet(w io.Writer, name string) {
	fmt.Fprintf(w, "hello, %s\n", name)
}

// countingWriter counts the bytes written to it and passes them on.
type countingWriter struct {
	w io.Writer
	n int
}

// Write implements io.Writer.
func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
}

func main() {
	greet(os.Stdout, "terminal")

	var b strings.Builder // *strings.Builder has a Write method too
	greet(&b, "builder")
	fmt.Printf("%q\n", b.String())

	cw := &countingWriter{w: os.Stdout}
	greet(cw, "counter")
	fmt.Println(cw.n, "bytes")

	// strings.NewReader returns an io.Reader. io.Copy moves bytes from
	// any Reader to any Writer.
	r := strings.NewReader("copied from a reader\n")
	n, err := io.Copy(cw, r)
	fmt.Println(n, err, cw.n)
}
