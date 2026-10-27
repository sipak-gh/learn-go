package rle

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEncode(t *testing.T) {
	if got := Encode("aaabccdddd"); got != "3a1b2c4d" {
		t.Errorf("Encode = %q, want 3a1b2c4d", got)
	}
}

// FuzzRoundTrip checks that decoding an encoded string gives back the
// original, for any input Encode accepts.
func FuzzRoundTrip(f *testing.F) {
	f.Add("aaabccdddd") // seed corpus: typical inputs
	f.Add("")
	f.Add("héééllo")
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) || strings.ContainsAny(s, "0123456789") {
			t.Skip("Encode doesn't promise anything for this input")
		}
		got, err := Decode(Encode(s))
		if err != nil {
			t.Fatalf("Decode(Encode(%q)): %v", s, err)
		}
		if got != s {
			t.Errorf("Decode(Encode(%q)) = %q", s, got)
		}
	})
}
