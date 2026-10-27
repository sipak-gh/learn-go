package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// slowSquare pretends to do work that takes d. It gives up early if ctx
// is canceled first.
func slowSquare(ctx context.Context, n int, d time.Duration) (int, error) {
	select {
	case <-time.After(d):
		return n * n, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	fmt.Println(slowSquare(ctx, 3, 10*time.Millisecond)) // in time

	n, err := slowSquare(ctx, 4, 100*time.Millisecond) // too slow
	fmt.Println(n, err)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}
