// sort.Interface: the classic three-method interface.
package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// byLength sorts strings by length. It implements sort.Interface:
// Len, Less, and Swap.
type byLength []string

func (s byLength) Len() int           { return len(s) }
func (s byLength) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s byLength) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func main() {
	words := []string{"banana", "fig", "apple", "kiwi"}
	sort.Sort(byLength(words))
	fmt.Println(words)

	// sort.Reverse wraps any sort.Interface and flips Less.
	sort.Sort(sort.Reverse(byLength(words)))
	fmt.Println(words)

	// Newer code usually passes a comparison function instead.
	slices.SortFunc(words, func(a, b string) int { return strings.Compare(a, b) })
	fmt.Println(words)
}
