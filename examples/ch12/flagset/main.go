package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type config struct {
	port    int
	verbose bool
	files   []string
}

// parse is an ordinary function: it takes the arguments, returns a
// result and an error, and never exits. That makes it easy to test.
func parse(args []string) (config, error) {
	var c config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we report errors ourselves
	fs.IntVar(&c.port, "port", 8080, "port to listen on")
	fs.BoolVar(&c.verbose, "v", false, "verbose output")
	fs.BoolVar(&c.verbose, "verbose", false, "same as -v") // an alias
	if err := fs.Parse(args); err != nil {
		return c, err
	}
	c.files = fs.Args() // whatever is left after the flags
	return c, nil
}

func main() {
	for _, args := range [][]string{
		{},
		{"-port", "9000", "--verbose", "a.txt", "b.txt"},
		{"-port=9000", "-v=false", "a.txt", "-v"},
		{"-v", "--", "-not-a-flag"},
		{"-port", "http"},
		{"-x"},
		{"-help"},
	} {
		c, err := parse(args)
		fmt.Printf("%q\n  %+v\n", args, c)
		if errors.Is(err, flag.ErrHelp) {
			fmt.Println("  help requested")
		} else if err != nil {
			fmt.Println("  error:", err)
		}
	}
}
