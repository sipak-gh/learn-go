//go:generate go run gen.go

package main

import "fmt"

func main() {
	fmt.Println(squares)
	fmt.Println(len(squares), "squares, the last is", squares[len(squares)-1])
}
