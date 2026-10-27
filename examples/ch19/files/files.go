// Package files counts lines in files.
package files

import (
	"bufio"
	"os"
)

// CountLines returns the number of lines in the file at path.
func CountLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}
