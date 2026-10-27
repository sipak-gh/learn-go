# Appendix B — Project Reference

The finished `seek`, as built in [Chapter 20](20-tooling-and-design.md). The code is in
`project/ch20/`.

## Usage

```console
$ seek -h
Usage: seek [OPTIONS] PATTERN [PATH ...]

Search for PATTERN in each PATH. Directories are searched recursively.
With no PATH, search the current directory; a PATH of - means standard input.
Options must come before PATTERN.

Options:
  -i, --ignore-case         ignore case distinctions
  -v, --invert-match        select non-matching lines
  -n, --line-number         print line numbers
  -c, --count               print only a count of selected lines per file
  -l, --files-with-matches  print only the names of files with selected lines
  -m, --max-count N         stop after N selected lines per file
  -E                        treat PATTERN as a regular expression
      --hidden              search hidden files and directories
      --glob GLOB           only search files whose name matches GLOB, e.g. '*.go'
  -j N                      search N files at a time (default: number of CPUs)
      --timeout D           give up after duration D, e.g. 500ms or 2s
      --color WHEN          highlight matches: auto, always, or never
                            (default auto, or the value of $SEEK_COLOR)
      --json                print results as JSON Lines, one object per line
      --stats               print statistics to standard error when done
      --debug               log what seek is doing to standard error
      --version             print the version and exit
  -h, --help                show this help and exit

Exit status is 0 if a line was selected, 1 if none was, and 2 if an error occurred.
```

Behavior details:

- Output is `path:text`, or `path:line:text` with `-n`. Standard input shows as
  `(standard input)`.
- Files are searched concurrently, but results always print in walk order: the order of the PATH
  arguments, then lexical order inside each directory.
- Hidden files and directories (names starting with `.`) are skipped unless you pass `--hidden`,
  but a hidden file named on the command line is always searched.
- Binary files (a NUL byte in the first 8 KiB) are skipped silently.
- Lines can be up to 1 MiB long.
- A file that can't be read is reported on standard error; the search continues, and the exit
  status is 2.
- `-c` prints a count for every file searched, including 0. `-l` prints only files with a match.
- `--color auto` colors only when standard output is a terminal. `$SEEK_COLOR` sets the default.
- `--stats` and `--debug` write to standard error, so they never mix with results in a pipe.
- `--version` prints the version set with `-ldflags "-X main.version=..."` at build time, or else
  the version Go recorded for the module.

## Sample runs

These runs use the project directory's `testdata/`:

```text
testdata/
  .cache/fox.txt     a cached fox
  .hidden.txt        a hidden fox
  image.bin          (binary data)
  notes.md           # Notes / - feed the FOX / - café opens at 9 / - buy coffee
  poem.txt           The quick brown fox / jumps over the lazy dog. /
                     The dog sleeps; the fox runs. / Foxes are quick and clever.
  src/hello.go       a small Go program that prints "hello, fox"
```

The binary was built with `go build -ldflags "-X main.version=1.0.0" -o seek ./cmd/seek` and put
on the `PATH`.

```console
$ seek --version
seek 1.0.0
$ seek fox testdata
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
$ seek -n -i fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/src/hello.go:4:	println("hello, fox")
$ seek -c fox testdata
testdata/notes.md:0
testdata/poem.txt:2
testdata/src/hello.go:1
$ seek -l -i fox testdata
testdata/notes.md
testdata/poem.txt
testdata/src/hello.go
$ seek -n -v fox testdata/poem.txt
testdata/poem.txt:2:jumps over the lazy dog.
testdata/poem.txt:4:Foxes are quick and clever.
$ seek -n -E 'f[aeiou]x|coffee' testdata
testdata/notes.md:4:- buy coffee
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
$ seek --glob '*.go' fox testdata
testdata/src/hello.go:	println("hello, fox")
$ seek --hidden fox testdata
testdata/.cache/fox.txt:a cached fox
testdata/.hidden.txt:a hidden fox
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
$ seek -m 1 --json fox testdata
{"path":"testdata/poem.txt","line":1,"text":"The quick brown fox"}
{"path":"testdata/src/hello.go","line":4,"text":"\tprintln(\"hello, fox\")"}
$ seek -c --json fox testdata
{"path":"testdata/notes.md","count":0}
{"path":"testdata/poem.txt","count":2}
{"path":"testdata/src/hello.go","count":1}
$ echo "a fox on stdin" | seek -n fox -
(standard input):1:a fox on stdin
$ seek --color always fox testdata/poem.txt | cat -v
testdata/poem.txt:The quick brown ^[[1;31mfox^[[0m
testdata/poem.txt:The dog sleeps; the ^[[1;31mfox^[[0m runs.
```

(`cat -v` shows the escape character as `^[`; a terminal shows `fox` in bold red.)

Statistics and the debug log go to standard error:

```console
$ seek --stats -i fox testdata > /dev/null
files searched: 3
bytes read:     223
files matched:  3
total matches:  5
elapsed:        248µs
top files:
      3  testdata/poem.txt
      1  testdata/notes.md
      1  testdata/src/hello.go
$ seek --debug -n fox testdata
level=DEBUG msg=starting pattern=fox paths=[testdata] workers=32
level=DEBUG msg="skipped binary file" path=testdata/image.bin
level=DEBUG msg=searched path=testdata/notes.md matches=0
level=DEBUG msg=searched path=testdata/poem.txt matches=2
level=DEBUG msg=searched path=testdata/src/hello.go matches=1
level=DEBUG msg=done searched=3 matches=3
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
```

`elapsed` and `workers` (the number of CPUs) vary by machine.

Exit status and errors:

```console
$ seek wolf testdata; echo "exit status $?"
exit status 1
$ seek fox testdata/missing.txt testdata/poem.txt; echo "exit status $?"
seek: testdata/missing.txt: no such file or directory
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
exit status 2
$ seek -x fox; echo "exit status $?"
seek: flag provided but not defined: -x
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
$ seek -E 'f(' testdata; echo "exit status $?"
seek: invalid pattern: error parsing regexp: missing closing ): `f(`
exit status 2
$ seek --timeout 1ns fox testdata; echo "exit status $?"
seek: timed out after 1ns
exit status 2
```

In the missing-file run, the error comes first: it goes straight to standard error, while
results are buffered until the end.

## Architecture

```text
            command line, stdin, stdout, stderr, Ctrl-C
                               │
                      main ── signal.NotifyContext
                               │  run(ctx, args, stdin, stdout, stderr) → exit status
                               ▼
  ┌──────────────────────────── run (cmd/seek) ─────────────────────────────┐
  │ parseArgs ──► options        newLogger ──► *slog.Logger (--debug)       │
  │ search.NewMatcher ──► Matcher: literal | foldLiteral | *regexp.Regexp   │
  │ output.Printer: *TextPrinter | *JSONPrinter     (writes to bufio.Writer)│
  └───────────────┬─────────────────────────────────────────────────────────┘
                  │ search.Walk(paths, WalkOptions) ──► iter.Seq2[path, error]
                  ▼
  ┌───────────── Searcher.SearchAll(ctx, files, workers) ───────────────────┐
  │                                                                         │
  │   producer ──jobs──► worker 1..N ── SearchFile ── Search (per line:     │
  │      │                    │                           Matcher)          │
  │      │                    └── result chan (one per file, buffered)      │
  │      └──pending (result chans, in walk order)──► consumer               │
  │                                                                         │
  └───────────────┬─────────────────────────────────────────────────────────┘
                  │ iter.Seq[Result], in walk order
                  ▼
     run's loop: Printer.PrintFile, stats.Counter.Add, logger.Debug
                  │
                  ▼
     context.Cause (timeout / Ctrl-C) ─► --stats (stats.TopN) ─► exit 0, 1, or 2
```

- **`cmd/seek`** is the only package that knows about flags, the environment, exit codes, and
  standard streams. `run` takes them as parameters, so tests can call it directly.
- **`internal/search`** finds matches and knows nothing about output formats.
- **`internal/output`** formats matches and knows nothing about files or concurrency.
- **`internal/stats`** is a generic counter that depends on nothing else in `seek`.
- Under `internal/`, no other module can import these packages, so their APIs can change freely.

## File map

`seek` has 1,113 lines of Go plus 810 lines of tests, and uses only the standard library.

| File | Lines | Contents |
|------|------:|----------|
| `go.mod` | 3 | `module seek`, `go 1.26` |
| `cmd/seek/main.go` | 264 | `main`, `run`, exit codes, `version`/`versionString`, `newLogger`, `printStats`, `report`, `useColor` |
| `cmd/seek/flags.go` | 179 | embedded `usage`, `options`, `colorMode` (a `flag.Value`), `usageError`, `parseArgs` |
| `cmd/seek/usage.txt` | 27 | the `--help` text |
| `cmd/seek/main_test.go` | 109 | end-to-end tests of `run` with golden files and `-update` |
| `cmd/seek/flags_test.go` | 118 | `parseArgs` table tests, `TestUsage` |
| `cmd/seek/testdata/*.golden` | | expected output of the end-to-end tests (11 files) |
| `internal/search/matcher.go` | 121 | `Matcher`, `NewMatcher`, `literal`, `foldLiteral`, `findAll`, `indexFold`, `prefixFold` |
| `internal/search/search.go` | 119 | `Match`, `ErrBinary`, `Searcher`, `countingReader`, `SearchFile`, `Search` |
| `internal/search/walk.go` | 82 | `WalkOptions`, `Walk`, `walkRoot` |
| `internal/search/parallel.go` | 112 | `Result`, `job`, `SearchAll` |
| `internal/search/matcher_test.go` | 114 | matcher table tests, `FuzzLiteral`, `FuzzIndexFold` |
| `internal/search/search_test.go` | 129 | `Search` tests, `ExampleSearcher_Search`, `BenchmarkSearch` |
| `internal/search/walk_test.go` | 94 | `Walk` tests on a temporary directory tree |
| `internal/search/parallel_test.go` | 54 | order and cancellation tests for `SearchAll` |
| `internal/output/output.go` | 99 | `Printer`, `TextPrinter`, `displayPath`, `highlight` |
| `internal/output/json.go` | 63 | `JSONPrinter` |
| `internal/output/output_test.go` | 125 | printer tests with `bytes.Buffer` |
| `internal/stats/counter.go` | 74 | `Counter[K]`, `Entry[K]`, `TopN` |
| `internal/stats/counter_test.go` | 67 | counter tests, `ExampleTopN` |
| `testdata/` | | sample files used by the tests and the runs above |

## Concepts → where `seek` uses them

| Concept | Chapter | Where in `seek` |
|---------|---------|-----------------|
| Modules, packages, `main` | [1](01-getting-started.md), [8](08-packages-and-modules.md) | `go.mod`; `cmd/seek` is `package main` |
| Constants and `iota` | [2](02-values-and-types.md) | exit codes in `main.go`; `colorAuto`/`colorAlways`/`colorNever` in `flags.go` |
| Named types over basic types | [2](02-values-and-types.md), [7](07-structs-methods-pointers.md) | `literal` and `foldLiteral` (strings), `colorMode` (an int) |
| `for`, `switch` without a tag | [3](03-control-flow.md) | `NewMatcher`, `TextPrinter.PrintFile`, the end of `run` |
| Multiple results | [4](04-functions.md) | `indexFold` returns `(start, end)`; `parseArgs` returns `(options, error)` |
| Closures | [4](04-functions.md) | `boolVar` in `parseArgs`; `emit` in `walkRoot`; the index function in `literal` |
| Variadic functions | [4](04-functions.md) | `usageErrorf(format string, args ...any)` |
| `defer`; `os.Exit` only in `main` | [4](04-functions.md) | `defer out.Flush()` in `run`; `defer f.Close()` in `SearchFile`; `main` |
| Slices and `append` | [5](05-collections.md) | `matches` in `Search`; `entries` in `TopN` |
| Maps, nil maps | [5](05-collections.md) | `Counter.counts`, created on the first `Add` |
| UTF-8, runes, `strings` | [6](06-strings.md) | `prefixFold` (`utf8.DecodeRuneInString`, `strings.EqualFold`); `highlight` (`strings.Builder`) |
| Structs, methods, pointer receivers | [7](07-structs-methods-pointers.md) | `Searcher`, `options`, `TextPrinter`; `(*colorMode).Set` |
| `internal/` packages | [8](08-packages-and-modules.md) | `internal/search`, `internal/output`, `internal/stats` |
| Sentinel errors, `errors.Is` | [9](09-errors.md) | `ErrBinary`; `errors.Is(err, flag.ErrHelp)` |
| Custom error types, `Unwrap`, `errors.As`, `%w` | [9](09-errors.md) | `usageError`; `report` (`errors.As`, `errors.AsType[*fs.PathError]`) |
| Implicit interfaces, compile-time checks | [10](10-interfaces.md) | `Matcher` (also satisfied by `*regexp.Regexp`), `Printer`, `flag.Value`; `var _ Printer = (*TextPrinter)(nil)` |
| Type assertions | [10](10-interfaces.md) | `w.(*os.File)` in `useColor` |
| `io.Reader`/`io.Writer`, wrapping | [11](11-io-and-files.md) | `run`'s parameters; `countingReader`; `bufio.Writer` around stdout |
| `bufio.Reader.Peek`, `bufio.Scanner` | [11](11-io-and-files.md) | binary check and line reading in `Search` |
| `filepath.WalkDir`, `SkipDir`, `Match` | [11](11-io-and-files.md) | `walkRoot`; `--glob` |
| `flag.FlagSet`, custom flag values, env vars | [12](12-flags.md) | `parseArgs`; `colorMode`; `SEEK_COLOR` |
| Generic types and functions, constraints | [13](13-generics.md) | `Counter[K comparable]`, `Entry[K any]`, `TopN[K cmp.Ordered]` |
| Iterators | [14](14-iterators.md) | `Walk` (`iter.Seq2`), `SearchAll` (`iter.Seq`), `Counter.All` (`maps.All`) |
| `encoding/json`, struct tags | [15](15-json-and-encoding.md) | `Match` tags; `JSONPrinter` with `json.Encoder` and an anonymous struct |
| Goroutines, channels, `select` | [16](16-goroutines-and-channels.md) | `SearchAll`: producer, workers, a channel of result channels |
| `sync.WaitGroup.Go`, `atomic.Int64` | [17](17-sync.md) | goroutines in `SearchAll`; `Searcher.bytesRead` |
| `context`: cancel, timeout with cause, signals | [18](18-context-and-time.md) | `signal.NotifyContext` in `main`; `WithTimeoutCause` in `run`; `WithCancel` in `SearchAll`; `ctx.Err()` in `Search` |
| `time.Duration`, `time.Since` | [18](18-context-and-time.md) | `--timeout`; `elapsed` in `printStats` |
| Table tests, subtests, `t.Parallel` | [19](19-testing.md) | `matcher_test.go`, `flags_test.go` |
| `t.Helper`, `t.TempDir` | [19](19-testing.md) | `walk_test.go`, `parallel_test.go`, `main_test.go` |
| Golden files, `-update` | [19](19-testing.md) | `main_test.go`, `cmd/seek/testdata/` |
| Examples, benchmarks, fuzzing | [19](19-testing.md) | `ExampleTopN`, `ExampleSearcher_Search`, `BenchmarkSearch`, `FuzzLiteral`, `FuzzIndexFold` |
| `go:embed` | [20](20-tooling-and-design.md) | `usage` in `flags.go` |
| `-ldflags -X`, `debug.ReadBuildInfo` | [20](20-tooling-and-design.md) | `version`, `versionString` in `main.go` |
| `log/slog` | [20](20-tooling-and-design.md) | `newLogger` and the `logger.Debug` calls in `run` |
| Cross-compiling | [20](20-tooling-and-design.md) | `GOOS=... GOARCH=... go build ./cmd/seek` (pure Go, no cgo) |

## Building, testing, installing

```console
$ go vet ./...
$ go test ./...
$ go test -race ./...
$ go build -ldflags "-X main.version=1.0.0" -o seek ./cmd/seek
$ go install ./cmd/seek
$ GOOS=windows GOARCH=amd64 go build -o seek.exe ./cmd/seek
```

Sources for this appendix: [REFERENCES.md](REFERENCES.md#appendix-b--project-reference)
