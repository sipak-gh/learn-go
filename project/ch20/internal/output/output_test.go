package output

import (
	"bytes"
	"testing"

	"seek/internal/search"
)

var matches = []search.Match{
	{Path: "poem.txt", Line: 1, Text: "The quick brown fox"},
	{Path: "poem.txt", Line: 3, Text: "The dog sleeps; the fox runs."},
}

func TestPrinters(t *testing.T) {
	tests := []struct {
		name    string
		printer func(*bytes.Buffer) Printer // makes a Printer writing to the buffer
		path    string
		matches []search.Match
		want    string
	}{
		{
			name:    "text",
			printer: func(b *bytes.Buffer) Printer { return &TextPrinter{W: b} },
			path:    "poem.txt", matches: matches,
			want: "poem.txt:The quick brown fox\npoem.txt:The dog sleeps; the fox runs.\n",
		},
		{
			name:    "text line numbers",
			printer: func(b *bytes.Buffer) Printer { return &TextPrinter{W: b, LineNumbers: true} },
			path:    "poem.txt", matches: matches[:1],
			want: "poem.txt:1:The quick brown fox\n",
		},
		{
			name:    "text count zero",
			printer: func(b *bytes.Buffer) Printer { return &TextPrinter{W: b, Count: true} },
			path:    "empty.txt", matches: nil,
			want: "empty.txt:0\n",
		},
		{
			name:    "text files only",
			printer: func(b *bytes.Buffer) Printer { return &TextPrinter{W: b, FilesOnly: true} },
			path:    "poem.txt", matches: matches,
			want: "poem.txt\n",
		},
		{
			name: "text highlight",
			printer: func(b *bytes.Buffer) Printer {
				m, _ := search.NewMatcher("fox", false, false)
				return &TextPrinter{W: b, Highlight: m}
			},
			path: "poem.txt", matches: matches[:1],
			want: "poem.txt:The quick brown " + colorStart + "fox" + colorReset + "\n",
		},
		{
			name:    "text stdin",
			printer: func(b *bytes.Buffer) Printer { return &TextPrinter{W: b} },
			path:    "-", matches: []search.Match{{Path: "-", Line: 1, Text: "a fox"}},
			want: "(standard input):a fox\n",
		},
		{
			name:    "json",
			printer: func(b *bytes.Buffer) Printer { return NewJSONPrinter(b) },
			path:    "poem.txt", matches: matches[:1],
			want: `{"path":"poem.txt","line":1,"text":"The quick brown fox"}` + "\n",
		},
		{
			name: "json count",
			printer: func(b *bytes.Buffer) Printer {
				p := NewJSONPrinter(b)
				p.Count = true
				return p
			},
			path: "poem.txt", matches: matches,
			want: `{"path":"poem.txt","count":2}` + "\n",
		},
		{
			name: "json files only, no matches",
			printer: func(b *bytes.Buffer) Printer {
				p := NewJSONPrinter(b)
				p.FilesOnly = true
				return p
			},
			path: "poem.txt", matches: nil,
			want: "",
		},
		{
			name:    "json special characters",
			printer: func(b *bytes.Buffer) Printer { return NewJSONPrinter(b) },
			path:    "a.go", matches: []search.Match{{Path: "a.go", Line: 2, Text: "\tx := \"<fox>\""}},
			want: `{"path":"a.go","line":2,"text":"\tx := \"<fox>\""}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer // a bytes.Buffer is an io.Writer that keeps what it gets
			if err := tt.printer(&buf).PrintFile(tt.path, tt.matches); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestHighlight(t *testing.T) {
	const on, off = colorStart, colorReset
	tests := []struct {
		line  string
		spans [][]int
		want  string
	}{
		{"a fox", [][]int{{2, 5}}, "a " + on + "fox" + off},
		{"fox fox", [][]int{{0, 3}, {4, 7}}, on + "fox" + off + " " + on + "fox" + off},
		{"abc", [][]int{{0, 0}}, "abc"}, // empty match: nothing to color
		{"abc", nil, "abc"},
	}
	for _, tt := range tests {
		if got := highlight(tt.line, tt.spans); got != tt.want {
			t.Errorf("highlight(%q, %v) = %q, want %q", tt.line, tt.spans, got, tt.want)
		}
	}
}
