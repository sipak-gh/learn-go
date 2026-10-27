// Package concat joins strings three ways, to compare their speed.
package concat

import "strings"

// Plus joins parts with +=, which copies the whole result every time.
func Plus(parts []string) string {
	s := ""
	for _, p := range parts {
		s += p
	}
	return s
}

// Builder joins parts with a strings.Builder, which grows a buffer.
func Builder(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

// Join uses strings.Join, which computes the final size first.
func Join(parts []string) string {
	return strings.Join(parts, "")
}
