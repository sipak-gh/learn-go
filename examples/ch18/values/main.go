package main

import (
	"context"
	"fmt"
)

// An unexported key type means no other package can use the same key,
// even by accident with the same name.
type ctxKey int

const requestIDKey ctxKey = 0

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string) // "" if missing
	return id
}

func handle(ctx context.Context) {
	fmt.Printf("[%s] handling\n", requestID(ctx))
}

func main() {
	ctx := withRequestID(context.Background(), "req-42")
	handle(ctx)
	handle(context.Background())
}
