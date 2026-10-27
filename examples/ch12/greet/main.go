package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	// Each call defines a flag and returns a pointer to its value.
	name := flag.String("name", "world", "who to greet")
	times := flag.Int("n", 1, "how many `times` to greet")
	shout := flag.Bool("shout", false, "greet in upper case")
	pause := flag.Duration("pause", 0, "pause between greetings, e.g. 1.5s")

	// Usage is called for -h and for errors. The default prints the
	// program's path, which is a temporary file under go run.
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: greet [flags] [extra ...]")
		flag.PrintDefaults()
	}
	flag.Parse() // reads os.Args[1:]; on error prints usage and exits 2

	msg := "Hello, " + *name + "!"
	if *shout {
		msg = strings.ToUpper(msg)
	}
	for range *times {
		fmt.Println(msg)
	}
	fmt.Println("pause:", *pause)
	fmt.Printf("%d extra args: %q\n", flag.NArg(), flag.Args())
}
