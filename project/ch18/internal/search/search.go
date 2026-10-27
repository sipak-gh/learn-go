// Package search finds lines of text that match a pattern.
package search

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync/atomic"
)

// A Match is one selected line of a file. The struct tags give the
// field names used when a Match is encoded as JSON.
type Match struct {
	Path string `json:"path"` // the file the line came from
	Line int    `json:"line"` // line number, counting from 1
	Text string `json:"text"` // the line itself, without its newline
}

// ErrBinary is returned for input that looks like binary data.
var ErrBinary = errors.New("binary file")

const (
	// binaryCheckSize is how much of a file is examined for NUL bytes.
	binaryCheckSize = 8 << 10 // 8 KiB
	// maxLineSize is the longest line Search can handle. bufio.Scanner's
	// default limit of 64 KiB is too small for some generated files.
	maxLineSize = 1 << 20 // 1 MiB
)

// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Matcher  Matcher   // decides which lines match; must not be nil
	Invert   bool      // select the lines that do NOT match
	MaxCount int       // stop after this many matches per file; 0 means no limit
	Stdin    io.Reader // read for the path "-"

	// bytesRead is updated by many goroutines at once (see SearchAll),
	// so it must be atomic. A plain int64 would be a data race.
	bytesRead atomic.Int64
}

// BytesRead returns the number of bytes read so far by all searches.
func (s *Searcher) BytesRead() int64 {
	return s.bytesRead.Load()
}

// countingReader is an io.Reader that adds the number of bytes read
// through it to a shared counter.
type countingReader struct {
	r io.Reader
	n *atomic.Int64
}

func (c countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n.Add(int64(n))
	return n, err
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.Matcher.FindAllStringIndex(line, 1) != nil
	return found != s.Invert // != on bools is "exclusive or"
}

// SearchFile searches the file at path. The path "-" means s.Stdin.
func (s *Searcher) SearchFile(ctx context.Context, path string) ([]Match, error) {
	if path == "-" {
		return s.Search(ctx, s.Stdin, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // we only read, so Close can't lose data
	return s.Search(ctx, f, path)
}

// Search reads r line by line and returns the selected lines. path is
// recorded in each Match. It returns ErrBinary, and no matches, if the
// first few KiB of r contain a NUL byte, and ctx.Err() if ctx is
// canceled before it reaches the end.
func (s *Searcher) Search(ctx context.Context, r io.Reader, path string) ([]Match, error) {
	r = countingReader{r: r, n: &s.bytesRead}

	// Peek at the start of the input without consuming it, so the
	// scanner below still sees every byte.
	br := bufio.NewReaderSize(r, binaryCheckSize)
	head, err := br.Peek(binaryCheckSize)
	if err != nil && err != io.EOF {
		return nil, err // io.EOF just means the input is short
	}
	if bytes.IndexByte(head, 0) >= 0 {
		return nil, ErrBinary
	}

	sc := bufio.NewScanner(br)
	sc.Buffer(make([]byte, 0, 64<<10), maxLineSize)
	var matches []Match
	for lineNum := 1; sc.Scan(); lineNum++ {
		// Checking once per line lets a huge file, or endless input,
		// be stopped quickly. ctx.Err is cheap: an atomic load.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := sc.Text()
		if !s.selects(line) {
			continue
		}
		matches = append(matches, Match{Path: path, Line: lineNum, Text: line})
		if len(matches) == s.MaxCount {
			break // stop reading: the rest of the file doesn't matter
		}
	}
	return matches, sc.Err()
}
