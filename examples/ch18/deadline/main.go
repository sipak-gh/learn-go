package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	deadline := time.Now().Add(2 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	d, ok := ctx.Deadline()
	fmt.Println("has deadline:", ok, "same:", d.Equal(deadline))

	// A child can shorten its parent's deadline but never extend it.
	child, cancel2 := context.WithTimeout(ctx, time.Hour)
	defer cancel2()
	cd, _ := child.Deadline()
	fmt.Println("child deadline is the parent's:", cd.Equal(deadline))

	// Canceling the parent cancels every child.
	cancel()
	<-child.Done()
	fmt.Println("child:", child.Err())

	_, ok = context.Background().Deadline()
	fmt.Println("Background has a deadline:", ok)
}
