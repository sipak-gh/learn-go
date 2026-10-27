package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// ctx is canceled when the program receives os.Interrupt (Ctrl-C).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop() // restore the default behavior of Ctrl-C

	fmt.Println("working; press Ctrl-C to stop early")
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("finished normally")
	case <-ctx.Done():
		fmt.Println("stopping:", context.Cause(ctx))
	}
}
