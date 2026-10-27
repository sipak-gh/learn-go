// Package output formats search results for the terminal.
package output

import (
	"fmt"
	"strings"

	"seek/internal/search"
)

// A Printer writes search results to standard output.
type Printer struct {
	LineNumbers bool // show line numbers
	Count       bool // print only the number of matches per file
	FilesOnly   bool // print only the names of files that have matches

	// Find locates the text to highlight in a line.
	// If it is nil, nothing is highlighted.
	Find func(line string, n int) [][]int
}

// PrintFile prints the results for one file.
func (p *Printer) PrintFile(path string, matches []search.Match) {
	switch {
	case p.FilesOnly:
		if len(matches) > 0 {
			fmt.Println(path)
		}
	case p.Count:
		fmt.Printf("%s:%d\n", path, len(matches))
	default:
		for _, m := range matches {
			p.printMatch(m)
		}
	}
}

// printMatch prints one matching line as path:text or path:line:text.
func (p *Printer) printMatch(m search.Match) {
	text := m.Text
	if p.Find != nil {
		text = highlight(text, p.Find(text, -1))
	}
	if p.LineNumbers {
		fmt.Printf("%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		fmt.Printf("%s:%s\n", m.Path, text)
	}
}

// ANSI escape sequences understood by most terminals.
const (
	colorStart = "\x1b[1;31m" // bold red
	colorReset = "\x1b[0m"
)

// highlight wraps each span of line in color escape sequences.
func highlight(line string, spans [][]int) string {
	var b strings.Builder
	prev := 0
	for _, span := range spans {
		start, end := span[0], span[1]
		if start == end {
			continue // an empty match has nothing to color
		}
		b.WriteString(line[prev:start])
		b.WriteString(colorStart)
		b.WriteString(line[start:end])
		b.WriteString(colorReset)
		prev = end
	}
	b.WriteString(line[prev:])
	return b.String()
}
