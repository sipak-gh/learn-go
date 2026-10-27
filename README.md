# Learn Go by Building `seek`

This course teaches Go from first principles. Along the way you build one real program:
**`seek`**, a small, fast, concurrent command-line search tool in the spirit of `grep`.

It borrows from two classics. As in *Go by Example*, every idea comes with a small runnable program
and its exact output. As in *The Rust Programming Language*, the chapters tell a story that
explains why the language works the way it does, and each one adds a feature to the project.

Read the chapters in order and type the code yourself. By the end you'll know Go's syntax, type
system, essential standard library, tooling, concurrency model, and design ideals well enough to
write and maintain idiomatic Go on your own.

## Requirements

- **Go 1.26 or newer.** Every example and project snapshot was verified with Go 1.26.
- A terminal and a text editor. Any editor works; one with Go support (via `gopls`) helps.
- Some experience with *any* programming language. No Go knowledge is needed.

Installation isn't covered. If `go version` prints `go1.26` or later, you're ready.

## How to use this course

- **Read in order.** Each chapter uses only what earlier chapters taught.
- **Type the code.** You learn more by typing the examples and project steps than by copying them.
- **Run everything.** Commands appear as `$ command` lines, followed by the output you should see.
- **Use the snapshots when stuck.** `project/chNN/` holds the complete `seek` project at the end of
  chapter NN. Compare your code with it (for example `diff -r myseek project/ch09`).
- **Run the examples.** `examples/chNN/<name>/` holds every example program from chapter NN. From
  the `examples/` directory, run one with `go run ./chNN/<name>`.
- **Do the exercises.** Each chapter ends with exercises whose solutions are hidden in collapsible
  blocks. Try each one before opening its solution.

## Table of contents

| #  | Chapter | You'll learn | `seek` milestone |
|----|---------|--------------|------------------|
| 1  | [Getting Started](01-getting-started.md) | The `go` command, modules, packages, `main`, gofmt | Create the module, print a usage line |
| 2  | [Values and Types](02-values-and-types.md) | Variables, zero values, basic types, constants, `iota` | Version and exit codes |
| 3  | [Control Flow](03-control-flow.md) | `if`, `for`, `switch`, labels | Read the pattern and options from the command line |
| 4  | [Functions](04-functions.md) | Multiple returns, variadics, closures, `defer` | First real search of one file |
| 5  | [Collections](05-collections.md) | Arrays, slices, maps, `slices`/`maps` packages | Multiple paths, `-c` |
| 6  | [Strings](06-strings.md) | UTF-8, bytes vs runes, `strings`, `strconv`, `fmt` | `-i`, `-n`, `-v`, highlighting |
| 7  | [Structs, Methods, and Pointers](07-structs-methods-pointers.md) | Structs, receivers, embedding, pointers | `options`, `Match`, and `Searcher` types |
| 8  | [Packages and Modules](08-packages-and-modules.md) | Visibility, layout, `internal/`, dependencies | Split into packages (no behavior change) |
| 9  | [Errors](09-errors.md) | Errors as values, wrapping, `Is`/`As`, panic/recover | Report errors and keep going |
| 10 | [Interfaces](10-interfaces.md) | Implicit interfaces, type switches, small interfaces | `Matcher`, `Printer`, `-E` regex |
| 11 | [I/O and Files](11-io-and-files.md) | `io.Reader`/`Writer`, `bufio`, walking directories | Streaming, recursive search, binary skipping, `--glob`, stdin |
| 12 | [Flags](12-flags.md) | The `flag` package, custom flag values | `FlagSet`, long options, `--help`, `--color auto` |
| 13 | [Generics](13-generics.md) | Type parameters, constraints, type sets | Generic `Counter`, `--stats` |
| 14 | [Iterators](14-iterators.md) | Range-over-func, `iter.Seq`, adapters | `Walk` as an iterator |
| 15 | [JSON and Encoding](15-json-and-encoding.md) | `encoding/json`, struct tags, custom marshaling | `--json` output |
| 16 | [Goroutines and Channels](16-goroutines-and-channels.md) | Goroutines, channels, `select`, pipelines | Concurrent worker pool |
| 17 | [Synchronization](17-sync.md) | Races, `sync`, atomics, the memory model | Race-free `bytes read` statistic, first `-race` test |
| 18 | [Context and Time](18-context-and-time.md) | `context`, cancellation, timers, signals | `--timeout`, Ctrl-C |
| 19 | [Testing](19-testing.md) | Table tests, golden files, benchmarks, fuzzing | A full test suite (finds a bug) |
| 20 | [Tooling and Design](20-tooling-and-design.md) | `embed`, build tags, `slog`, Go's design ideals | Embedded help, version stamping, `--debug` logging, cross-compiling |
| A  | [Cheat Sheet](appendix-a-cheatsheet.md) | Syntax quick reference | |
| B  | [Project Reference](appendix-b-project-reference.md) | `seek` usage, architecture, concept map | |

References: [REFERENCES.md](REFERENCES.md) lists the sources behind each chapter, with archived
(Wayback Machine) copies of every link.

## What you'll build

```console
$ seek -n -i --stats fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/src/hello.go:4:	println("hello, fox")
files searched: 3
bytes read:     223
files matched:  3
total matches:  5
elapsed:        407µs
top files:
      3  testdata/poem.txt
      1  testdata/notes.md
      1  testdata/src/hello.go
```

(The elapsed time varies from run to run.) By Chapter 20, `seek` searches directory trees
concurrently, with case-insensitive, inverted, and regular-expression matching, file-name globs,
JSON output, colored highlighting, statistics, timeouts, and graceful Ctrl-C handling. It has a
complete test suite and uses only the Go standard library.
