// Interfaces can embed other interfaces.
package main

import "fmt"

type Getter interface {
	Get(key string) (string, bool)
}

type Setter interface {
	Set(key, value string)
}

// A Store has both methods: the union of Getter and Setter.
type Store interface {
	Getter
	Setter
}

type mapStore map[string]string

func (m mapStore) Get(k string) (string, bool) {
	v, ok := m[k]
	return v, ok
}

func (m mapStore) Set(k, v string) {
	m[k] = v
}

// lookup needs only Get, so it asks for only a Getter.
func lookup(g Getter, key string) string {
	if v, ok := g.Get(key); ok {
		return v
	}
	return "(none)"
}

func main() {
	var s Store = mapStore{}
	s.Set("lang", "go")
	fmt.Println(lookup(s, "lang"), lookup(s, "editor"))
}
