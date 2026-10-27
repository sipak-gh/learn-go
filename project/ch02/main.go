// Command seek searches files for lines that match a pattern.
package main

import "fmt"

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
	fmt.Printf("%s %s\n", name, version)
	fmt.Println(usage)
	fmt.Printf("exit codes: ok=%d no-match=%d error=%d\n", exitOK, exitNoMatch, exitError)
}
