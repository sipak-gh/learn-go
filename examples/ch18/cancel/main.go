package main

import (
	"context"
	"fmt"
	"sync"
)

// count sends 1, 2, 3, ... on out until ctx is canceled.
func count(ctx context.Context, out chan<- int) {
	for i := 1; ; i++ {
		select {
		case out <- i:
		case <-ctx.Done():
			fmt.Println("count:", ctx.Err())
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // always release the context's resources

	nums := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { count(ctx, nums) })

	for range 3 {
		fmt.Println(<-nums)
	}
	cancel()  // tell count to stop
	wg.Wait() // and wait until it has
	fmt.Println("main:", ctx.Err())
}
