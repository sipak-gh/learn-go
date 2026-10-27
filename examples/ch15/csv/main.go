package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := `name,lang,note
seek,Go,"search, fast"
ripgrep,Rust,"says ""hi"""
`
	r := csv.NewReader(strings.NewReader(input))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, rec := range records[1:] { // skip the header
		fmt.Printf("%q\n", rec)
	}

	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{"path", "count"})
	_ = w.Write([]string{"a file, with comma.txt", "3"})
	w.Flush() // csv.Writer is buffered
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}
}
