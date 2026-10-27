package main

import (
	"encoding/json"
	"fmt"
)

type Book struct {
	Title string   `json:"title"`
	Year  int      `json:"year"`
	Tags  []string `json:"tags"`
}

func main() {
	data := []byte(`{"TITLE": "Go", "year": 2015, "publisher": "AW"}`)
	var b Book
	if err := json.Unmarshal(data, &b); err != nil { // note the &
		fmt.Println("error:", err)
		return
	}
	// "TITLE" matched title (case-insensitive), "publisher" was ignored,
	// and Tags kept its zero value.
	fmt.Printf("%+v\n", b)

	// Unmarshal only sets the fields present in the input.
	_ = json.Unmarshal([]byte(`{"tags": ["go"]}`), &b)
	fmt.Printf("%+v\n", b)

	var errs = []string{
		`{"year": "2015"}`, // wrong type
		`{"year": 2015,}`,  // trailing comma
		`{"title": "Go"`,   // cut short
	}
	for _, in := range errs {
		var b Book
		err := json.Unmarshal([]byte(in), &b)
		fmt.Println("error:", err)
	}
}
