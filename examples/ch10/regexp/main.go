// A quick tour of the regexp package.
package main

import (
	"fmt"
	"regexp"
)

// MustCompile panics on a bad pattern. Use it for patterns fixed in the
// source; use Compile for patterns that come from users.
var date = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)

func main() {
	text := "released 2024-02-06, patched 2024-03-05"

	fmt.Println(date.MatchString(text))
	fmt.Println(date.FindString(text))
	fmt.Println(date.FindAllString(text, -1))
	fmt.Println(date.FindAllStringIndex(text, -1)) // byte offsets
	fmt.Println(date.FindStringSubmatch(text))     // whole match, then groups
	fmt.Println(date.ReplaceAllString(text, "$3/$2/$1"))

	fold := regexp.MustCompile(`(?i)go+`) // (?i) ignores case
	fmt.Println(fold.FindAllString("Go GOOO gopher", -1))

	_, err := regexp.Compile(`(unclosed`)
	fmt.Println(err)
}
