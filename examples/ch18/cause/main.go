package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errShutdown = errors.New("server shutting down")

func main() {
	// WithCancelCause: the cancel function takes the reason.
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errShutdown)
	fmt.Println("Err:  ", ctx.Err())          // always one of two generic errors
	fmt.Println("Cause:", context.Cause(ctx)) // the reason you gave

	// WithTimeoutCause: the cause is reported when time runs out.
	ctx2, cancel2 := context.WithTimeoutCause(context.Background(),
		10*time.Millisecond, errors.New("lookup took too long"))
	defer cancel2()
	<-ctx2.Done()
	fmt.Println("Err:  ", ctx2.Err())
	fmt.Println("Cause:", context.Cause(ctx2))

	// For a context that isn't canceled, Cause is nil.
	fmt.Println("Cause:", context.Cause(context.Background()))
}
