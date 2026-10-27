// Command labels shows labeled break and continue in nested loops.
package main

import "fmt"

func main() {
	grid := [][]int{ // a slice of slices (Chapter 5)
		{1, 2, 3},
		{4, -5, 6},
		{7, 8, 9},
	}

	// Find the first negative number. A plain break would only
	// leave the inner loop.
search:
	for r, row := range grid {
		for c, v := range row {
			if v < 0 {
				fmt.Println("negative at", r, c)
				break search
			}
		}
	}

	// Sum only the rows that have no negative numbers.
rows:
	for r, row := range grid {
		sum := 0
		for _, v := range row {
			if v < 0 {
				continue rows // skip the rest of this row
			}
			sum += v
		}
		fmt.Println("row", r, "sum", sum)
	}
}
