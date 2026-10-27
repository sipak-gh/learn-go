package first

import "testing"

func TestAbs(t *testing.T) {
	if got := Abs(-3); got != 3 {
		t.Errorf("Abs(-3) = %d, want 3", got)
	}
	if got := Abs(5); got != 5 {
		t.Errorf("Abs(5) = %d, want 5", got)
	}
	if got := Abs(0); got != 0 {
		t.Errorf("Abs(0) = %d, want 0", got)
	}
}
