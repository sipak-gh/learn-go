package concat

import (
	"strings"
	"testing"
)

var parts = strings.Fields(strings.Repeat("the quick brown fox ", 250)) // 1000 words

func BenchmarkPlus(b *testing.B) {
	for b.Loop() {
		Plus(parts)
	}
}

func BenchmarkBuilder(b *testing.B) {
	for b.Loop() {
		Builder(parts)
	}
}

func BenchmarkJoin(b *testing.B) {
	for b.Loop() {
		Join(parts)
	}
}

// Benchmarks don't check results, so test that all three agree.
func TestSame(t *testing.T) {
	want := Plus(parts)
	if Builder(parts) != want || Join(parts) != want {
		t.Error("results differ")
	}
}
