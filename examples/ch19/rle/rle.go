// Package rle implements run-length encoding: "aaab" becomes "3a1b".
package rle

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Encode replaces each run of a repeated rune with its length and the
// rune. s must be valid UTF-8 and contain no digits 0-9; otherwise the
// result can't be decoded.
func Encode(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); {
		j := i
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		fmt.Fprintf(&b, "%d%c", j-i, runes[i])
		i = j
	}
	return b.String()
}

// Decode reverses Encode.
func Decode(s string) (string, error) {
	var b strings.Builder
	for s != "" {
		// Only ASCII digits: unicode.IsDigit would also accept digits
		// from other scripts, such as '߇' (NKo seven), which Encode
		// treats as ordinary runes.
		i := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
		if i <= 0 {
			return "", fmt.Errorf("rle: expected a count then a rune at %q", s)
		}
		n, err := strconv.Atoi(s[:i])
		if err != nil {
			return "", err
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(strings.Repeat(string(r), n))
		s = s[i+size:]
	}
	return b.String(), nil
}
