package main

import (
	"encoding/json"
	"fmt"
)

type Book struct {
	Title   string   `json:"title"`
	Author  string   `json:"author,omitempty"` // left out when ""
	Year    int      `json:"year"`
	Tags    []string `json:"tags"`
	ISBN    string   `json:"-"` // never encoded
	Pages   int      // no tag: the field name is used as is
	private string   // unexported: invisible to encoding/json
}

func main() {
	b := Book{Title: "The Go Programming Language", Year: 2015,
		Tags: []string{"go", "programming"}, ISBN: "978-0134190440", Pages: 380, private: "x"}
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))

	// A nil slice becomes null; an empty slice becomes [].
	var none []string
	n, _ := json.Marshal(none)
	e, _ := json.Marshal([]string{})
	fmt.Println(string(n), string(e))

	// Maps become objects, with the keys sorted.
	m, _ := json.Marshal(map[string]int{"b": 2, "a": 1, "c": 3})
	fmt.Println(string(m))

	// Indented output, for humans.
	pretty, _ := json.MarshalIndent(map[string]any{"ok": true, "n": []int{1, 2}}, "", "  ")
	fmt.Println(string(pretty))

	// Some values have no JSON form.
	_, err = json.Marshal(make(chan int))
	fmt.Println("error:", err)
}
