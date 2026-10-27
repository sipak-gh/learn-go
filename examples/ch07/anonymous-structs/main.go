// Anonymous structs: struct types without a name.
package main

import (
	"fmt"
	"strings"
)

func main() {
	// A one-off value that groups a few settings.
	server := struct {
		Host string
		Port int
	}{"localhost", 8080}
	fmt.Printf("%s:%d\n", server.Host, server.Port)

	// A table of cases: a very common use, especially in tests.
	cases := []struct {
		in, want string
	}{
		{"go", "GO"},
		{"café", "CAFÉ"},
		{"", ""},
	}
	for _, c := range cases {
		got := strings.ToUpper(c.in)
		fmt.Printf("%q -> %q ok=%t\n", c.in, got, got == c.want)
	}

	// The empty struct takes no memory. map[string]struct{} is a set.
	seen := map[string]struct{}{"a": {}}
	_, ok := seen["a"]
	fmt.Println(ok)
}
