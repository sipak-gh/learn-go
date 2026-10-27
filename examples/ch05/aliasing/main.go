// Command aliasing shows that slices share their underlying array.
package main

import "fmt"

func main() {
	nums := []int{1, 2, 3, 4, 5}

	// A slice of a slice shares the same array.
	part := nums[1:3] // [2 3], len 2, cap 4
	part[0] = 20
	fmt.Println(nums, part, len(part), cap(part))

	// append writes into the shared array when there is spare capacity,
	// silently overwriting nums[3].
	part = append(part, 40)
	fmt.Println(nums, part)

	// A full slice expression a[low:high:max] limits the capacity,
	// so append must copy to a new array instead.
	nums = []int{1, 2, 3, 4, 5}
	safe := nums[1:3:3] // len 2, cap 2
	safe = append(safe, 40)
	safe[0] = 99
	fmt.Println(nums, safe)
}
