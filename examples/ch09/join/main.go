// errors.Join combines several errors into one.
package main

import (
	"errors"
	"fmt"
)

var (
	ErrNoName = errors.New("name is required")
	ErrNoAge  = errors.New("age is required")
)

type user struct {
	name string
	age  int
}

// validate reports every problem, not just the first.
func validate(u user) error {
	var errs []error
	if u.name == "" {
		errs = append(errs, ErrNoName)
	}
	if u.age == 0 {
		errs = append(errs, ErrNoAge)
	}
	return errors.Join(errs...) // nil if errs is empty
}

func main() {
	err := validate(user{})
	fmt.Println(err) // one line per error
	fmt.Println("no name:", errors.Is(err, ErrNoName))
	fmt.Println("no age:", errors.Is(err, ErrNoAge))

	fmt.Println(validate(user{"ana", 31}) == nil)

	// fmt.Errorf accepts several %w verbs too.
	both := fmt.Errorf("save failed: %w; cleanup failed: %w", ErrNoName, ErrNoAge)
	fmt.Println(both)
	fmt.Println("no age:", errors.Is(both, ErrNoAge))
}
