package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Duration is a time.Duration that is written as a string like "1m30s"
// in JSON, instead of a number of nanoseconds.
type Duration struct {
	time.Duration
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler. It needs a pointer
// receiver, because it changes d.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

// Level is written as text. TextMarshaler also works for map keys.
type Level int

func (l Level) MarshalText() ([]byte, error) {
	return []byte([]string{"debug", "info", "warn"}[l]), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "debug":
		*l = 0
	case "info":
		*l = 1
	case "warn":
		*l = 2
	default:
		return fmt.Errorf("unknown level %q", text)
	}
	return nil
}

type Config struct {
	Timeout Duration      `json:"timeout"`
	Level   Level         `json:"level"`
	Counts  map[Level]int `json:"counts"`
}

func main() {
	c := Config{Timeout: Duration{90 * time.Second}, Level: 1,
		Counts: map[Level]int{0: 5, 2: 1}}
	data, _ := json.Marshal(c)
	fmt.Println(string(data))

	var back Config
	err := json.Unmarshal([]byte(`{"timeout":"250ms","level":"WARN"}`), &back)
	fmt.Println(back.Timeout, back.Level, err)

	err = json.Unmarshal([]byte(`{"timeout":5}`), &back)
	fmt.Println("error:", err)
	err = json.Unmarshal([]byte(`{"level":"loud"}`), &back)
	fmt.Println("error:", err)
}
