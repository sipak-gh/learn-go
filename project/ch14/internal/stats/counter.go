// Package stats provides generic counting helpers.
package stats

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// A Counter counts occurrences of keys of any comparable type K.
// The zero value is an empty Counter ready to use.
type Counter[K comparable] struct {
	counts map[K]int
}

// Add adds n to the count for key.
func (c *Counter[K]) Add(key K, n int) {
	if c.counts == nil {
		c.counts = make(map[K]int) // writing to a nil map would panic
	}
	c.counts[key] += n
}

// Get returns the count for key, or 0 if key was never added.
func (c *Counter[K]) Get(key K) int {
	return c.counts[key] // reading a nil map is fine
}

// Len returns the number of distinct keys.
func (c *Counter[K]) Len() int {
	return len(c.counts)
}

// Total returns the sum of all counts.
func (c *Counter[K]) Total() int {
	total := 0
	for _, n := range c.All() {
		total += n
	}
	return total
}

// All returns an iterator over the keys and their counts, in no
// particular order.
func (c *Counter[K]) All() iter.Seq2[K, int] {
	return maps.All(c.counts) // ranging over a nil map is fine
}

// An Entry is a key with its count.
type Entry[K any] struct {
	Key   K
	Count int
}

// TopN returns up to n entries with the highest counts, highest first.
// Equal counts are ordered by key, so the result is always the same.
//
// TopN is a function rather than a method because it needs to sort keys,
// so K must satisfy cmp.Ordered, which is stricter than Counter's
// comparable. Methods can't add constraints or type parameters of their own.
func TopN[K cmp.Ordered](c *Counter[K], n int) []Entry[K] {
	entries := make([]Entry[K], 0, c.Len())
	for key, count := range c.All() {
		entries = append(entries, Entry[K]{Key: key, Count: count})
	}
	slices.SortFunc(entries, func(a, b Entry[K]) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c // higher counts first
		}
		return cmp.Compare(a.Key, b.Key)
	})
	return entries[:min(n, len(entries))]
}
