package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	// AfterFunc runs f in its own goroutine once ctx is done.
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("cleaning up:", ctx.Err())
		close(done)
	})

	cancel()
	<-done
	// stop reports whether it stopped f from running. Too late here.
	fmt.Println("stopped:", stop())

	ctx2, cancel2 := context.WithCancel(context.Background())
	stop2 := context.AfterFunc(ctx2, func() { fmt.Println("never runs") })
	fmt.Println("stopped:", stop2()) // unregister f before ctx2 is done
	cancel2()
}
