package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func main() {
	data := []byte(`{"name": "seek", "stars": 42, "tags": ["cli", "go"], "meta": {"ok": true}, "x": null}`)

	// Decoding into any gives maps, slices, and basic types.
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println("error:", err)
		return
	}
	obj := v.(map[string]any)
	for _, k := range []string{"name", "stars", "tags", "meta", "x"} {
		fmt.Printf("%-5s %-24T %v\n", k, obj[k], obj[k])
	}

	// Every JSON number becomes a float64. To use one as an int,
	// convert it, after a type assertion.
	stars := int(obj["stars"].(float64))
	fmt.Println(stars + 1)

	// Big integers lose precision as float64. A Decoder with UseNumber
	// keeps the original text as a json.Number instead.
	big := `{"id": 12345678901234567890}`
	var m1 map[string]any
	_ = json.Unmarshal([]byte(big), &m1)
	fmt.Println(m1["id"])

	dec := json.NewDecoder(strings.NewReader(big))
	dec.UseNumber()
	var m2 map[string]any
	_ = dec.Decode(&m2)
	fmt.Printf("%T %v\n", m2["id"], m2["id"])
}
