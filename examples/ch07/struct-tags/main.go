// Struct tags: metadata on fields, read by other packages.
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
	age      int    // unexported: invisible to other packages
}

func main() {
	u := User{Name: "Ada", Password: "secret", age: 36}
	// json.Marshal reads the tags to choose key names. Chapter 15
	// covers encoding/json; here we only peek at what tags do.
	data, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
