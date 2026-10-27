// Command slicespkg tours the slices package.
package main

import (
	"fmt"
	"slices"
	"strings"
)

func main() {
	words := []string{"pear", "apple", "fig", "banana"}

	fmt.Println(slices.Contains(words, "fig"), slices.Index(words, "fig"))
	fmt.Println(slices.Index(words, "kiwi")) // -1 means not found

	slices.Sort(words) // sorts in place
	fmt.Println(words)
	i, found := slices.BinarySearch(words, "fig") // needs sorted input
	fmt.Println(i, found)

	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
	fmt.Println(slices.Max(nums), slices.Min(nums))
	slices.Reverse(nums)
	fmt.Println(nums)

	// Sort with your own comparison: by length, then alphabetically.
	slices.SortFunc(words, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b) // -1, 0, or +1
	})
	fmt.Println(words)

	fmt.Println(slices.Equal([]int{1, 2}, []int{1, 2}))
	fmt.Println(slices.Insert([]int{1, 4}, 1, 2, 3))
	fmt.Println(slices.Compact([]int{1, 1, 2, 2, 2, 3, 1}))
}
