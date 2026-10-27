// A custom error type carries structured details.
package main

import "fmt"

// A SyntaxError reports a problem at a line of a config file.
type SyntaxError struct {
	Line int
	Msg  string
}

// Error makes *SyntaxError satisfy the error interface.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

func parse(lines []string) error {
	for i, line := range lines {
		if line == "" {
			return &SyntaxError{Line: i + 1, Msg: "empty line"}
		}
	}
	return nil // not a nil *SyntaxError: see Chapter 10
}

func main() {
	err := parse([]string{"a=1", "", "c=3"})
	fmt.Println(err)

	var e error = &SyntaxError{Line: 9, Msg: "missing '='"}
	fmt.Println(e.Error())

	fmt.Println(parse([]string{"a=1"}) == nil)
}
