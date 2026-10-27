// Wrapping adds context to an error and keeps the original inside.
package main

import (
	"errors"
	"fmt"
)

var ErrNoUser = errors.New("no such user")

func loadUser(id int) error {
	return ErrNoUser
}

func loadProfile(id int) error {
	if err := loadUser(id); err != nil {
		return fmt.Errorf("load user %d: %w", id, err) // %w wraps
	}
	return nil
}

func handle(id int) error {
	if err := loadProfile(id); err != nil {
		return fmt.Errorf("handle request: %w", err)
	}
	return nil
}

func main() {
	err := handle(7)
	fmt.Println(err)

	// Unwrap peels off one layer at a time.
	for e := err; e != nil; e = errors.Unwrap(e) {
		fmt.Printf("  %T: %v\n", e, e)
	}

	fmt.Println("err == ErrNoUser:", err == ErrNoUser)
	fmt.Println("errors.Is:", errors.Is(err, ErrNoUser))

	// %v formats the error but does not wrap it: the chain is cut.
	flat := fmt.Errorf("handle request: %v", ErrNoUser)
	fmt.Println(flat)
	fmt.Println("errors.Is on flat:", errors.Is(flat, ErrNoUser))
}
