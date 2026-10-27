// Command seek searches files for lines that match a pattern.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"seek/internal/output"
	"seek/internal/search"
	"seek/internal/stats"
)

const name = "seek"

// version is set when building a release:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/seek
//
// -X can only set a package-level string variable, not a constant.
var version string

// versionString returns the version to print for --version.
func versionString() string {
	if version != "" {
		return version
	}
	// The go command records the module version in the binary: a tag
	// like "v1.2.3" for go install seek@v1.2.3, or "(devel)" for a
	// build from a local directory.
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "unknown"
}

// Exit codes. Like grep, seek reports whether anything matched.
// iota counts up from 0, so the order of these lines matters.
const (
	exitOK      = iota // 0: success (at least one line matched)
	exitNoMatch        // 1: no line matched
	exitError          // 2: bad usage or an error occurred
)

// main only translates run's result into an exit status. os.Exit ends the
// program immediately, without running deferred calls, so all real work
// happens in run, where defer behaves normally.
func main() {
	// Ctrl-C sends os.Interrupt. Instead of killing the program, it now
	// cancels ctx, so seek can stop cleanly and print what it found.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop() // os.Exit skips deferred calls, so call stop directly
	os.Exit(code)
}

// run is the whole program. It reads standard input from stdin, writes
// results to stdout and messages to stderr, and returns the exit status.
// Taking an io.Reader and io.Writers instead of using os.Stdin, os.Stdout,
// and os.Stderr directly lets tests supply input and capture output.
// Canceling ctx stops the search early.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	start := time.Now()

	// Writing many small pieces straight to a file costs one system call
	// each; a bufio.Writer collects them and writes in big chunks.
	out := bufio.NewWriter(stdout)
	defer out.Flush() // without this, the last chunk would be lost

	opts, err := parseArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(out, usage) // usage.txt ends with a newline already
		return exitOK
	}
	if err != nil {
		report(stderr, err)
		return exitError
	}
	if opts.version {
		fmt.Fprintln(out, name, versionString())
		return exitOK
	}

	logger := newLogger(stderr, opts.debug)
	logger.Debug("starting", "pattern", opts.pattern, "paths", opts.paths, "workers", opts.workers)

	if opts.timeout > 0 {
		// The cause is the error context.Cause reports once time is up.
		cause := fmt.Errorf("timed out after %v", opts.timeout)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, opts.timeout, cause)
		defer cancel() // release the timer if we finish in time
	}

	matcher, err := search.NewMatcher(opts.pattern, opts.ignoreCase, opts.regex)
	if err != nil {
		report(stderr, err)
		return exitError
	}
	s := &search.Searcher{
		Matcher:  matcher,
		Invert:   opts.invert,
		MaxCount: opts.maxCount,
		Stdin:    stdin,
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}

	// p's static type is the Printer interface; its dynamic type is
	// *output.JSONPrinter or *output.TextPrinter.
	var p output.Printer
	if opts.json {
		jp := output.NewJSONPrinter(out)
		jp.Count = opts.count
		jp.FilesOnly = opts.filesOnly
		p = jp
	} else {
		tp := &output.TextPrinter{
			W:           out,
			LineNumbers: opts.lineNumbers,
			Count:       opts.count,
			FilesOnly:   opts.filesOnly,
		}
		if useColor(opts.color, stdout) {
			tp.Highlight = matcher
		}
		p = tp
	}

	walkOpts := search.WalkOptions{Hidden: opts.hidden, Glob: opts.glob}

	// A bad file shouldn't stop the search: report it, remember that
	// something went wrong, and carry on with the next one.
	var (
		searched int                   // files searched
		matched  stats.Counter[string] // matches per file, for files with any
		hadError bool
	)
	files := search.Walk(opts.paths, walkOpts)
	for res := range s.SearchAll(ctx, files, opts.workers) {
		err := res.Err
		if errors.Is(err, search.ErrBinary) {
			logger.Debug("skipped binary file", "path", res.Path)
			continue // binary files are skipped silently
		}
		if err == nil {
			err = p.PrintFile(res.Path, res.Matches)
		}
		if err != nil {
			report(stderr, err)
			hadError = true
			continue
		}
		logger.Debug("searched", "path", res.Path, "matches", len(res.Matches))
		searched++
		if len(res.Matches) > 0 {
			matched.Add(res.Path, len(res.Matches))
		}
	}

	// context.Cause is nil unless ctx was canceled; then it says why,
	// e.g. "interrupt signal received" or "timed out after 2s".
	if err := context.Cause(ctx); err != nil {
		out.Flush() // print the results found so far first
		report(stderr, err)
		hadError = true
	}

	logger.Debug("done", "searched", searched, "matches", matched.Total())

	if opts.stats {
		out.Flush() // print the results before the statistics
		printStats(stderr, searched, s.BytesRead(), &matched, time.Since(start))
	}

	switch {
	case hadError:
		return exitError
	case matched.Total() == 0:
		return exitNoMatch
	default:
		return exitOK
	}
}

// newLogger returns a structured logger that writes to w if it is
// enabled, and discards everything otherwise.
func newLogger(w io.Writer, enabled bool) *slog.Logger {
	if !enabled {
		return slog.New(slog.DiscardHandler)
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug, // the default level, Info, would hide Debug
		// Drop the time from each record: it is noise for a short run
		// and would make the output differ every time.
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{} // an empty Attr is left out
			}
			return a
		},
	}))
}

// printStats prints a summary of the search to w.
func printStats(w io.Writer, searched int, bytes int64, matched *stats.Counter[string], elapsed time.Duration) {
	fmt.Fprintf(w, "files searched: %d\n", searched)
	fmt.Fprintf(w, "bytes read:     %d\n", bytes)
	fmt.Fprintf(w, "files matched:  %d\n", matched.Len())
	fmt.Fprintf(w, "total matches:  %d\n", matched.Total())
	fmt.Fprintf(w, "elapsed:        %v\n", elapsed.Round(time.Microsecond))
	top := stats.TopN(matched, 5) // K is inferred as string
	if len(top) > 0 {
		fmt.Fprintln(w, "top files:")
	}
	for _, e := range top {
		fmt.Fprintf(w, "%7d  %s\n", e.Count, e.Key)
	}
}

// report prints err to standard error. Usage errors are followed by a
// reminder of how to call seek.
func report(stderr io.Writer, err error) {
	// *fs.PathError messages look like "open notes.txt: no such file or
	// directory". Users care about the path, not which call failed.
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		err = fmt.Errorf("%s: %w", pathErr.Path, pathErr.Err)
	}
	fmt.Fprintln(stderr, "seek:", err)

	var usageErr *usageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(stderr, usageLine)
		fmt.Fprintln(stderr, "Run 'seek --help' for more information.")
	}
}

// useColor reports whether to highlight matches written to w.
func useColor(mode colorMode, w io.Writer) bool {
	switch mode {
	case colorAlways:
		return true
	case colorNever:
		return false
	}
	// Auto: color only if w is a terminal. A type assertion asks whether
	// the interface value w holds an *os.File.
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
