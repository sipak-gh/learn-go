package main

import (
	"fmt"
	"time"
)

func main() {
	// A Duration is an int64 count of nanoseconds, with constants for
	// the common units.
	d := 1*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Println(d, int64(d))
	fmt.Println(d.Minutes(), d.Round(time.Hour), d.Truncate(time.Hour))
	fmt.Println(1500*time.Millisecond, 2*time.Microsecond, time.Duration(42))

	// ParseDuration accepts the same format String produces.
	p, err := time.ParseDuration("1m30s")
	fmt.Println(p, p.Seconds(), err)
	_, err = time.ParseDuration("soon")
	fmt.Println(err)

	// Multiplying by a variable needs a conversion: n is an int.
	n := 3
	fmt.Println(time.Duration(n) * time.Second)

	// Measuring elapsed time.
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	elapsed := time.Since(start) // same as time.Now().Sub(start)
	fmt.Println(elapsed >= 20*time.Millisecond, elapsed.Round(10*time.Millisecond))
}
