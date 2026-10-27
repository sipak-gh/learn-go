package output

import (
	"encoding/json"
	"io"

	"seek/internal/search"
)

// JSONPrinter prints results as JSON Lines: one JSON object per line.
//
//	{"path":"notes.txt","line":3,"text":"a fox"}  for each match
//	{"path":"notes.txt","count":2}                 with Count
//	{"path":"notes.txt"}                           with FilesOnly
//
// Create one with NewJSONPrinter.
type JSONPrinter struct {
	Count     bool // print only the number of matches per file
	FilesOnly bool // print only the names of files that have matches

	enc *json.Encoder
}

var _ Printer = (*JSONPrinter)(nil)

// NewJSONPrinter returns a JSONPrinter that writes to w.
func NewJSONPrinter(w io.Writer) *JSONPrinter {
	enc := json.NewEncoder(w)
	// By default <, >, and & are escaped (for embedding in HTML). Our
	// output is for programs, so keep the text as it is.
	enc.SetEscapeHTML(false)
	return &JSONPrinter{enc: enc}
}

// fileCount is the JSON form of a per-file count.
type fileCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// PrintFile implements Printer.
func (p *JSONPrinter) PrintFile(path string, matches []search.Match) error {
	path = displayPath(path)
	switch {
	case p.FilesOnly:
		if len(matches) == 0 {
			return nil
		}
		// An anonymous struct is handy for a one-off shape.
		return p.enc.Encode(struct {
			Path string `json:"path"`
		}{path})
	case p.Count:
		return p.enc.Encode(fileCount{Path: path, Count: len(matches)})
	}
	for _, m := range matches {
		m.Path = path // m is a copy, so this doesn't change matches
		if err := p.enc.Encode(m); err != nil {
			return err
		}
	}
	return nil
}
