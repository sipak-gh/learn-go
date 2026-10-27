// Command switchbreak shows that break inside a switch leaves only the switch.
package main

import "fmt"

func main() {
	words := []string{"go", "stop", "after"}

	for _, w := range words {
		switch w {
		case "stop":
			break // leaves the switch, NOT the loop
		default:
			fmt.Println("plain break:", w)
		}
	}

loop:
	for _, w := range words {
		switch w {
		case "stop":
			break loop // leaves the loop
		default:
			fmt.Println("labeled break:", w)
		}
	}
}
