package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Post struct {
	Title     string    `json:"title"`
	Published time.Time `json:"published"`
	Edited    time.Time `json:"edited,omitzero"` // left out when zero
	Draft     time.Time `json:"draft,omitempty"` // omitempty can't tell
}

func main() {
	p := Post{
		Title:     "Hello",
		Published: time.Date(2025, 2, 11, 9, 30, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(p)
	fmt.Println(string(data))

	var back Post
	err := json.Unmarshal([]byte(`{"published":"2025-02-11T10:30:00+01:00"}`), &back)
	fmt.Println(back.Published, err)
	fmt.Println(back.Published.Equal(p.Published))

	err = json.Unmarshal([]byte(`{"published":"11/02/2025"}`), &back)
	fmt.Println("error:", err)
}
