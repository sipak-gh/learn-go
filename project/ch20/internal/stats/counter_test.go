package stats

import (
	"fmt"
	"slices"
	"testing"
)

func TestCounterZeroValue(t *testing.T) {
	var c Counter[string] // usable without initialization
	if c.Len() != 0 || c.Total() != 0 || c.Get("x") != 0 {
		t.Errorf("zero Counter: Len %d, Total %d, Get %d; want all 0", c.Len(), c.Total(), c.Get("x"))
	}
	c.Add("x", 2)
	c.Add("x", 3)
	c.Add("y", 1)
	if c.Len() != 2 || c.Total() != 6 || c.Get("x") != 5 {
		t.Errorf("got Len %d, Total %d, Get(x) %d; want 2, 6, 5", c.Len(), c.Total(), c.Get("x"))
	}
}

func TestTopN(t *testing.T) {
	var c Counter[string]
	c.Add("b", 2)
	c.Add("a", 2) // same count as "b": sorted by key
	c.Add("c", 5)
	c.Add("d", 1)
	tests := []struct {
		n    int
		want []Entry[string]
	}{
		{0, []Entry[string]{}},
		{2, []Entry[string]{{"c", 5}, {"a", 2}}},
		{10, []Entry[string]{{"c", 5}, {"a", 2}, {"b", 2}, {"d", 1}}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint("n=", tt.n), func(t *testing.T) {
			// Entry[string] is comparable, so slices.Equal works on it.
			if got := TopN(&c, tt.n); !slices.Equal(got, tt.want) {
				t.Errorf("TopN(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestTopNInts(t *testing.T) {
	var c Counter[int] // the same code works for any ordered key type
	c.Add(404, 3)
	c.Add(200, 7)
	want := []Entry[int]{{200, 7}, {404, 3}}
	if got := TopN(&c, 5); !slices.Equal(got, want) {
		t.Errorf("TopN = %v, want %v", got, want)
	}
}

func ExampleTopN() {
	var words Counter[string]
	for _, w := range []string{"fox", "dog", "fox", "cat", "fox", "dog"} {
		words.Add(w, 1)
	}
	for _, e := range TopN(&words, 2) {
		fmt.Println(e.Key, e.Count)
	}
	// Output:
	// fox 3
	// dog 2
}
