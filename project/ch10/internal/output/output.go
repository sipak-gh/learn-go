// Package output formats search results.
package output

import (
	"fmt"
	"io"
	"strings"

	"seek/internal/search"
)

// A Printer writes the results of a search.
type Printer interface {
	// PrintFile prints the matches found in the file at path.
	PrintFile(path string, matches []search.Match) error
}

// TextPrinter prints results as plain text, one line per match:
// path:text, or path:line:text if LineNumbers is set.
type TextPrinter struct {
	W           io.Writer // where to print
	LineNumbers bool      // show line numbers
	Count       bool      // print only the number of matches per file
	FilesOnly   bool      // print only the names of files that have matches

	// Highlight, if not nil, finds the text to color in each line.
	Highlight search.Matcher
}

var _ Printer = (*TextPrinter)(nil)

// PrintFile implements Printer.
func (p *TextPrinter) PrintFile(path string, matches []search.Match) error {
	switch {
	case p.FilesOnly:
		if len(matches) == 0 {
			return nil
		}
		_, err := fmt.Fprintln(p.W, path)
		return err
	case p.Count:
		_, err := fmt.Fprintf(p.W, "%s:%d\n", path, len(matches))
		return err
	}
	for _, m := range matches {
		if err := p.printMatch(m); err != nil {
			return err
		}
	}
	return nil
}

// printMatch prints one matching line.
func (p *TextPrinter) printMatch(m search.Match) error {
	text := m.Text
	if p.Highlight != nil {
		text = highlight(text, p.Highlight.FindAllStringIndex(text, -1))
	}
	var err error
	if p.LineNumbers {
		_, err = fmt.Fprintf(p.W, "%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		_, err = fmt.Fprintf(p.W, "%s:%s\n", m.Path, text)
	}
	return err
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
