package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Blob struct {
	Name string `json:"name"`
	Data []byte `json:"data"` // []byte is encoded as a base64 string
}

func main() {
	raw := []byte("hi?\x00\xff")
	s := base64.StdEncoding.EncodeToString(raw)
	fmt.Println(s)
	fmt.Println(base64.URLEncoding.EncodeToString(raw)) // safe in URLs
	back, err := base64.StdEncoding.DecodeString(s)
	fmt.Printf("%q %v\n", back, err)

	data, _ := json.Marshal(Blob{"tiny", raw})
	fmt.Println(string(data))
	var b Blob
	_ = json.Unmarshal(data, &b)
	fmt.Printf("%q\n", b.Data)
}
