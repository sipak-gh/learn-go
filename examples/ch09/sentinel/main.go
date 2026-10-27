// Sentinel errors are package-level values that callers compare against.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrNotFound is returned by find when the key is missing.
var ErrNotFound = errors.New("not found")

func find(m map[string]int, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func main() {
	ages := map[string]int{"ana": 31}
	if _, err := find(ages, "bob"); err == ErrNotFound {
		fmt.Println("bob:", err)
	}

	// io.EOF is the standard library's best-known sentinel. It isn't a
	// failure: it means "no more data".
	r := strings.NewReader("hi")
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			fmt.Println("done")
			break
		}
		fmt.Printf("read %q\n", buf[:n])
	}

	// errors.New makes a new, distinct value each time.
	fmt.Println(errors.New("not found") == ErrNotFound)
}
