// Command grid builds a two-dimensional slice.
package main

import "fmt"

func main() {
	rows, cols := 3, 4
	grid := make([][]rune, rows) // a slice of rows
	for r := range grid {
		grid[r] = make([]rune, cols) // each row is its own slice
		for c := range grid[r] {
			grid[r][c] = '.'
		}
	}
	grid[1][2] = '#'

	for _, row := range grid {
		fmt.Println(string(row))
	}

	// Rows may have different lengths ("jagged").
	tri := [][]int{{1}, {1, 1}, {1, 2, 1}}
	fmt.Println(tri, len(tri[2]))
}
