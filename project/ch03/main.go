// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
)

const (
	name    = "seek"
	version = "0.1.0"
	usage   = "Usage: seek [OPTIONS] PATTERN [PATH ...]"
)

// Exit codes. Like grep, seek reports whether anything matched.
// iota counts up from 0, so the order of these lines matters.
const (
	exitOK      = iota // 0: success (at least one line matched)
	exitNoMatch        // 1: no line matched
	exitError          // 2: bad usage or an error occurred
)

func main() {
	// os.Args[0] is the program name; the arguments follow it.
	args := os.Args[1:]

	// Options come first. The first argument that is not an option is
	// the PATTERN, and everything after it is a PATH.
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			os.Exit(exitOK)
		case "--version":
			fmt.Println(name, version)
			os.Exit(exitOK)
		case "--":
			// "--" ends the options, so a PATTERN may start with "-".
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "seek: unknown option %q\n", arg)
				os.Exit(exitError)
			}
			break loop // a plain break would only leave the switch
		}
	}

	if i >= len(args) {
		fmt.Fprintln(os.Stderr, "seek: missing PATTERN")
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(exitError)
	}
	pattern := args[i]
	paths := args[i+1:]

	fmt.Printf("pattern: %q\n", pattern)
	for n, path := range paths {
		fmt.Printf("path %d: %s\n", n+1, path)
	}
	if len(paths) == 0 {
		fmt.Println("no paths given")
	}
}
