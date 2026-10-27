package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// upper is an io.Reader that turns ASCII letters from r into upper case.
type upper struct {
	r io.Reader
}

func (u upper) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	for i := range p[:n] {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err
}

func main() {
	// A pipeline of readers: source → first 15 bytes → upper case → stdout.
	src := strings.NewReader("hello, gophers! this part is cut off")
	n, err := io.Copy(os.Stdout, upper{io.LimitReader(src, 15)})
	fmt.Printf("\ncopied %d bytes, err=%v\n", n, err)

	// MultiReader reads its readers one after another.
	r := io.MultiReader(strings.NewReader("one "), strings.NewReader("two"))
	// TeeReader writes everything it reads to a second place.
	var seen strings.Builder
	data, err := io.ReadAll(io.TeeReader(r, &seen))
	fmt.Printf("ReadAll: %q err=%v\n", data, err)
	fmt.Printf("tee saw: %q\n", seen.String())

	// MultiWriter writes the same bytes to several writers.
	var a, b bytes.Buffer
	w := io.MultiWriter(&a, &b, os.Stdout)
	fmt.Fprintln(w, "to three places")
	fmt.Printf("a=%q b=%q\n", a.String(), b.String())
}
