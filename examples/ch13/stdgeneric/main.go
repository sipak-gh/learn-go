package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	nums := []int{5, 2, 8, 2, 9}
	fmt.Println(slices.Max(nums), slices.Min(nums), slices.Contains(nums, 8), slices.Index(nums, 2))
	sorted := slices.Clone(nums)
	slices.Sort(sorted)
	i, found := slices.BinarySearch(sorted, 8)
	fmt.Println(sorted, i, found)
	fmt.Println(slices.Compact(sorted)) // remove adjacent duplicates

	people := []Person{{"Cy", 30}, {"Al", 25}, {"Bo", 30}}
	// Sort by age, oldest first, then by name. cmp.Or returns the first
	// non-zero comparison.
	slices.SortFunc(people, func(a, b Person) int {
		return cmp.Or(
			cmp.Compare(b.Age, a.Age),
			strings.Compare(a.Name, b.Name),
		)
	})
	fmt.Println(people)

	ages := map[string]int{"Al": 25, "Bo": 30}
	clone := maps.Clone(ages)
	clone["Cy"] = 30
	fmt.Println(len(ages), len(clone), maps.Equal(ages, clone))
	fmt.Println(min(3, 1, 2), max("a", "b")) // built-in, also generic
}
