// Package secret can be imported only by code inside
// examples/ch08/internal-demo, because it lives under internal/.
package secret

// Answer is exported, but only to packages that may import secret.
func Answer() int { return 42 }
