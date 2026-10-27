// Package registry keeps a list of named text transforms.
package registry

import (
	"fmt"
	"slices"
)

var transforms = map[string]func(string) string{}

// Package-level variables are initialized before any init function runs.
var ready = announce("registry: variables initialized")

func announce(msg string) bool {
	fmt.Println(msg)
	return true
}

func init() {
	fmt.Println("registry: init, ready =", ready)
}

// Register adds a transform under name.
func Register(name string, f func(string) string) {
	transforms[name] = f
}

// Names returns the registered names in sorted order.
func Names() []string {
	var names []string
	for name := range transforms {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Apply runs the transform called name on s.
func Apply(name, s string) string {
	return transforms[name](s)
}
