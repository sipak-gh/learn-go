package reverse

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"one rune", "a", "a"},
		{"ascii", "hello", "olleh"},
		{"palindrome", "racecar", "racecar"},
		{"multi-byte", "héllo, 世界", "界世 ,olléh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := String(tt.in); got != tt.want {
				t.Errorf("String(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
