package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Event struct {
	Kind string `json:"kind"`
	Msg  string `json:"msg"`
}

func main() {
	// An Encoder writes one JSON value per Encode call, plus a newline.
	// That's the JSON Lines format.
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(Event{"start", "a <b> & c"})
	enc.SetEscapeHTML(false)
	_ = enc.Encode(Event{"start", "a <b> & c"})
	enc.SetIndent("", "  ")
	_ = enc.Encode(Event{"stop", "bye"})

	// A Decoder reads values one after another, until io.EOF.
	input := `{"kind":"a","msg":"one"}
{"kind":"b","msg":"two"}
{"kind":"c","msg":"three","extra":1}`
	dec := json.NewDecoder(strings.NewReader(input))
	dec.DisallowUnknownFields()
	for {
		var e Event
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("error:", err)
			break
		}
		fmt.Printf("%+v\n", e)
	}
}
