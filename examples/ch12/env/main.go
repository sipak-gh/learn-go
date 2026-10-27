package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Getenv can't tell "unset" from "set to the empty string".
	fmt.Printf("Getenv:    %q\n", os.Getenv("GREETING"))
	v, ok := os.LookupEnv("GREETING")
	fmt.Printf("LookupEnv: %q, %v\n", v, ok)

	// Usual precedence: built-in default < environment < command line.
	greeting := "hello"
	if v, ok := os.LookupEnv("GREETING"); ok {
		greeting = v
	}
	flag.StringVar(&greeting, "greeting", greeting, "the greeting (default from $GREETING)")
	flag.Parse()
	fmt.Println("result:   ", greeting)
}
