package main

import (
	"encoding/json"
	"fmt"
)

// An envelope's Data depends on its Type, so it is kept as raw JSON
// and decoded later, once we know what it is.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Point struct{ X, Y int }

type Text struct {
	Body string `json:"body"`
}

func main() {
	input := []byte(`[
		{"type": "point", "data": {"X": 1, "Y": 2}},
		{"type": "text",  "data": {"body": "hi"}},
		{"type": "sound", "data": [1, 2, 3]}
	]`)
	var envs []Envelope
	if err := json.Unmarshal(input, &envs); err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, e := range envs {
		switch e.Type {
		case "point":
			var p Point
			_ = json.Unmarshal(e.Data, &p)
			fmt.Printf("point %+v\n", p)
		case "text":
			var t Text
			_ = json.Unmarshal(e.Data, &t)
			fmt.Printf("text %q\n", t.Body)
		default:
			fmt.Printf("unknown %s: %s\n", e.Type, e.Data)
		}
	}
}
