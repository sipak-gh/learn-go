# Chapter 20 — Shipping It: Tooling, Embedding, and Go's Design Ideals

`seek` is complete and tested ([Chapter 19](19-testing.md)). This last chapter is about shipping it
and the tools around the code: embedding files in the binary, choosing files per OS with build
constraints, stamping a version at build time, cross-compiling, installing, generating code,
logging with `log/slog`, and profiling. Then come `unsafe` and cgo and why to avoid them, Go's
design ideals, a recap of common mistakes, a tour of the finished code, and where to go next.

## Embedding files with `go:embed`

A Go program compiles to one file, which is a big part of why it's easy to ship: copy the binary
and run it. But programs often need data files too: help text, templates, web pages. The
`//go:embed` directive puts files *into* the binary at compile time.

`examples/ch20/embed/main.go`

```go
package main

import (
	"embed"
	"fmt"
	"io/fs"
)

// A string (or []byte) variable holds the contents of one file.
//
//go:embed hello.txt
var greeting string

// An embed.FS holds a tree of files. It implements fs.FS.
//
//go:embed static
var static embed.FS

func main() {
	fmt.Print(greeting)

	data, err := static.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}
	fmt.Printf("index.html: %q\n", data)

	fs.WalkDir(static, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path, d.IsDir())
		return nil
	})
}
```

The package directory holds these files:

```text
examples/ch20/embed/
  main.go
  hello.txt               Hello from a file!
  static/index.html       <h1>seek</h1>
  static/css/site.css     h1 { color: teal; }
```

```console
$ go run ./ch20/embed
Hello from a file!
index.html: "<h1>seek</h1>\n"
. true
static true
static/css true
static/css/site.css false
static/index.html false
```

The rules:

- The directive goes on the line right above a **package-level** `var`. It's a comment, with no
  space after `//`.
- The variable's type is `string` or `[]byte` for one file, or `embed.FS` for one or more files
  and directories. `embed.FS` is read-only and implements `fs.FS`, so everything from
  [Chapter 19](19-testing.md)'s `testing/fstest` section works on it.
- Patterns are relative to the package directory and can't contain `..` or start with `/`, so a
  package can embed only files from its own directory tree.
- When embedding a directory, files whose names start with `.` or `_` are left out. Write
  `//go:embed all:static` to include them.
- The file must import `embed`. With `embed.FS` you use the package anyway; with a `string` or
  `[]byte` you don't, so import it for its side effect: `import _ "embed"`.

Forget the import and the build fails:

```console
$ go build
# e20/a
a/main.go:5:3: go:embed requires import "embed" (or import _ "embed", if package is not used)
```

A missing file is a compile error too, not a run-time surprise:

```console
$ go build
b/main.go:8:12: pattern missing.txt: no matching files found
```

**Why a compiler feature?** The alternatives are worse. Reading files at run time breaks the
"one binary" promise: the program fails when run from a different directory. Generating Go source
with big string literals works but needs an extra build step. With `go:embed`, the files are
normal files while you edit and part of the binary when you ship.

## Build constraints and file-name suffixes

Sometimes code must differ by OS, CPU, or build mode. Go has no `#ifdef`; instead, whole *files*
are included in or left out of a build, in one of two ways:

1. A **file-name suffix**: `name_GOOS.go`, `name_GOARCH.go`, or `name_GOOS_GOARCH.go`, for
   example `term_windows.go`, `fast_amd64.go`, `poll_linux_arm64.go`. (`_test.go` is the same
   idea: only `go test` builds those files.)
2. A **`//go:build` line** at the top of the file, before `package` and followed by a blank line,
   holding a boolean expression of build tags with `&&`, `||`, `!`, and parentheses.

`examples/ch20/buildtags/main.go`

```go
package main

import "fmt"

func main() {
	fmt.Println("platform:", platform())
	fmt.Println("debug build:", debugBuild)
}
```

`examples/ch20/buildtags/platform_windows.go`

```go
package main

// The _windows suffix in the file name is an implicit build
// constraint: this file is only compiled when GOOS is windows.
func platform() string { return "Windows" }
```

`examples/ch20/buildtags/platform_other.go`

```go
//go:build !windows

package main

import "runtime"

func platform() string { return "not Windows (" + runtime.GOOS + ")" }
```

`examples/ch20/buildtags/debug_on.go`

```go
//go:build debug

package main

const debugBuild = true
```

`examples/ch20/buildtags/debug_off.go`

```go
//go:build !debug

package main

const debugBuild = false
```

Every build must see exactly one `platform` and one `debugBuild`, so the constraints come in
pairs that cover every case. `debug` isn't a built-in tag; turn it on with `-tags`:

```console
$ go run ./ch20/buildtags
platform: not Windows (linux)
debug build: false
$ go run -tags debug ./ch20/buildtags
platform: not Windows (linux)
debug build: true
```

`go list` shows which files a build would use. Set `GOOS` to ask about another OS:

```console
$ go list -f '{{.GoFiles}}' ./ch20/buildtags
[debug_off.go main.go platform_other.go]
$ GOOS=windows go list -f '{{.GoFiles}}' ./ch20/buildtags
[debug_off.go main.go platform_windows.go]
$ go list -tags debug -f '{{.GoFiles}} {{.IgnoredGoFiles}}' ./ch20/buildtags
[debug_on.go main.go platform_other.go] [debug_off.go platform_windows.go]
```

Tags that are set for you:

| Tag | Set when |
|-----|----------|
| `linux`, `windows`, `darwin`, … | `GOOS` has that value |
| `amd64`, `arm64`, … | `GOARCH` has that value |
| `unix` | `GOOS` is a Unix-like system (Linux, macOS, the BSDs, …) |
| `cgo` | cgo is enabled |
| `go1.21`, `go1.22`, … | the Go version is at least that |
| anything in `-tags a,b` | you passed it |

The special tag `ignore` is never set, so `//go:build ignore` keeps a file out of every build;
the next section uses that.

Gotchas:

- `vet` and your editor check only the current OS's files. A typo in `platform_windows.go` goes
  unnoticed on Linux until you run `GOOS=windows go vet ./...`.
- Keep OS-specific files small: shared logic in normal files, only the few differing lines behind
  constraints.
- Older code may use the old syntax `// +build linux`. `gofmt` adds the matching `//go:build`
  line; new code needs only `//go:build`.

## Stamping a version: `-ldflags -X` and `debug.ReadBuildInfo`

A release binary should know its version, and typing it into the source for each release is easy
to forget. Instead, the *linker* can set a string variable at build time:

```text
go build -ldflags "-X importpath.name=value"
```

`-ldflags` passes flags to the linker, and `-X` sets the package-level string variable `name` in
package `importpath`. For package `main`, the import path is just `main`.

Go also records facts about every build in the binary: the Go version, the module path and
version, and build settings. `runtime/debug.ReadBuildInfo` reads them.

`examples/ch20/version/main.go`

```go
package main

import (
	"fmt"
	"runtime/debug"
)

// version is empty unless set with -ldflags "-X main.version=...".
var version string

func main() {
	fmt.Printf("version: %q\n", version)

	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("no build info")
		return
	}
	fmt.Println("go:", info.GoVersion)
	fmt.Println("path:", info.Path)
	fmt.Println("module:", info.Main.Path, info.Main.Version)
	for _, s := range info.Settings {
		if s.Key == "GOOS" || s.Key == "GOARCH" || s.Key == "-ldflags" {
			fmt.Printf("%s=%s\n", s.Key, s.Value)
		}
	}
}
```

```console
$ go run ./ch20/version
version: ""
go: go1.26.10
path: examples/ch20/version
module: examples (devel)
GOARCH=amd64
GOOS=linux
$ go run -ldflags "-X main.version=1.2.3" ./ch20/version
version: "1.2.3"
go: go1.26.10
path: examples/ch20/version
module: examples (devel)
-ldflags=-X main.version=1.2.3
GOARCH=amd64
GOOS=linux
```

Your Go patch version, OS, and CPU may differ.

What `info.Main.Version` holds:

- `(devel)` for a build of a local directory that isn't in a version control repository.
- Inside a git repository, a version from git: a tag such as `v1.2.0` if the current commit has
  one, or a "pseudo-version" such as `v0.0.0-20261027150405-abcdef123456`. A `+dirty` suffix
  means there were uncommitted changes.
- The module's real version, such as `v1.2.0`, for `go install example.com/tool@v1.2.0`.

`go version -m BINARY` shows the same information without running the program. You'll use it in
the checkpoint.

Two `-X` mistakes fail *silently*: the build succeeds and the value doesn't change.

- **The target is a constant.** `-X` can only set a variable, and Chapter 19's `seek` has
  `const version = "0.1.0"`:

  ```console
  $ go run -ldflags "-X main.version=1.0.0" ./cmd/seek --version
  seek 0.1.0
  ```

- **The name is wrong.** `-X main.Version=1.0.0` (capital V) sets nothing, because there is no
  `Version`. Only a malformed flag is an error:

  ```console
  $ go run -ldflags "-X version=1.0.0" ./cmd/seek --version
  # seek/cmd/seek
  /usr/local/go/pkg/tool/linux_amd64/link: -X flag requires argument of the form importpath.name=value
  ```

## Cross-compiling

The Go toolchain can build for any supported system from any other. Set two environment
variables:

- `GOOS`: the target operating system (`linux`, `windows`, `darwin` for macOS, `freebsd`, …).
- `GOARCH`: the target CPU (`amd64`, `arm64`, `386`, `riscv64`, …).

```console
$ GOOS=windows GOARCH=amd64 go build -o seek.exe ./cmd/seek
```

No extra compiler, SDK, or libraries are needed. `go tool dist list` prints every supported pair;
Go 1.26 has 47:

```console
$ go tool dist list | head -3
aix/ppc64
android/386
android/amd64
$ go tool dist list | wc -l
47
```

The catch is cgo (below): code that calls C needs a C cross-compiler for the target. cgo is off by
default when cross-compiling, so packages that need it won't build. `seek` is pure Go, so it
cross-compiles cleanly.

Useful flags for release builds:

| Flag | Effect |
|------|--------|
| `-trimpath` | Removes your local directory paths from the binary (and from panic stack traces) |
| `-ldflags "-s -w"` | Leaves out the symbol table and debug info. Smaller, but harder to debug |
| `-ldflags "-X main.version=1.0.0"` | Sets the version, as above |

On Linux/amd64, `-s -w` shrank `seek` from 3,926,386 to 2,670,754 bytes. (Sizes vary with the Go
version and platform.)

## Installing: `go install`

`go build` writes the binary to the current directory. `go install` builds it and puts it in a
directory meant for your programs:

```console
$ go install ./cmd/seek
$ seek --version
seek (devel)
```

- It goes in `$GOBIN` if set, else in `$(go env GOPATH)/bin`. `GOPATH` defaults to `go` in your
  home directory (`~/go/bin` on Linux and macOS).
- Add that directory to your `PATH` to run the program by name, as above.
- `go install example.com/some/tool@latest` installs a published tool without cloning it. It needs
  a version (`@latest`, `@v1.4.0`) and ignores your current module. This is the usual way to
  install Go tools.
- `go clean -i ./cmd/seek` removes the installed binary.

## Generating code: `go generate`

Some code is better written by a program: lookup tables, `String` methods for enums, code from a
specification. `go generate` runs commands listed in your source files in comments like:

```go
//go:generate command args...
```

It runs them in the package directory, and you commit the generated files like any other Go
files. The key rule: **`go build` never runs `go generate`.** You run it when the inputs change,
and people who build your code don't need the generator at all.

A common trick is to write the generator in Go, as a file in the same directory with
`//go:build ignore` so normal builds skip it.

`examples/ch20/generate/main.go`

```go
//go:generate go run gen.go

package main

import "fmt"

func main() {
	fmt.Println(squares)
	fmt.Println(len(squares), "squares, the last is", squares[len(squares)-1])
}
```

`examples/ch20/generate/gen.go`

```go
//go:build ignore

// gen.go writes squares.go. Run it with go generate. The ignore
// constraint keeps it out of normal builds of this package.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
)

func main() {
	var b bytes.Buffer
	fmt.Fprintln(&b, "// Code generated by gen.go; DO NOT EDIT.")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "package main")
	fmt.Fprintln(&b)
	fmt.Fprint(&b, "var squares = [...]int{")
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&b, "%d, ", i*i)
	}
	fmt.Fprintln(&b, "}")

	// format.Source runs gofmt on the generated code.
	src, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("squares.go", src, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

`-x` prints each command as it runs:

```console
$ cd ch20/generate
$ go generate -x
go run gen.go
$ cat squares.go
// Code generated by gen.go; DO NOT EDIT.

package main

var squares = [...]int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
$ cd ../..
$ go run ./ch20/generate
[1 4 9 16 25 36 49 64 81 100]
10 squares, the last is 100
$ go list -f '{{.GoFiles}} {{.IgnoredGoFiles}}' ./ch20/generate
[main.go squares.go] [gen.go]
```

Notes:

- The `// Code generated ... DO NOT EDIT.` line is a convention tools check: editors warn before
  you edit such a file, and linters skip it.
- `go/format`, `go/ast`, and `go/parser` are in the standard library, so Go programs can easily
  read and write Go code. That's why so much Go tooling exists.
- `go generate ./...` runs the directives in every package.
- A well-known generator is `stringer` (`go install golang.org/x/tools/cmd/stringer@latest`). With
  `//go:generate stringer -type=colorMode`, it would write the `String` method you wrote by hand
  for `colorMode` in [Chapter 12](12-flags.md).

## Structured logging with `log/slog`

The `log` package writes lines of text: fine for people, hard for programs to search and filter.
`log/slog` writes **structured** records: a message plus key-value attributes, at a level.

The package-level functions use a default logger that writes through the `log` package:

```go
slog.Info("hello", "user", "ana")
```

```text
2026/10/27 11:22:36 INFO hello user=ana
```

(The date and time are when you run it.) Real programs build a `*slog.Logger` from a
**handler**, which decides the format and where the output goes.

`examples/ch20/slog/main.go`

```go
package main

import (
	"log/slog"
	"os"
	"time"
)

func main() {
	// Remove the time so the output is the same on every run.
	noTime := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		return a
	}

	text := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: noTime}))
	text.Info("file opened", "path", "notes.md", "size", 1024)
	text.Debug("not shown: below the default level, Info")
	text.Warn("slow file", "path", "big log.txt", "took", 1500*time.Millisecond)
	text.Error("open failed", "err", os.ErrNotExist)

	// With returns a logger that adds attributes to every record.
	reqLog := text.With("request", 42)
	reqLog.Info("started")

	// Groups nest attributes under a name.
	text.Info("done", slog.Group("stats", "files", 3, "matches", 7))

	// The JSON handler writes one JSON object per record.
	json := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: noTime,
	}))
	json.Debug("hello", "user", "ana", slog.Int("tries", 3))
	json.Info("done", slog.Group("stats", "files", 3, "matches", 7))
}
```

```console
$ go run ./ch20/slog
level=INFO msg="file opened" path=notes.md size=1024
level=WARN msg="slow file" path="big log.txt" took=1.5s
level=ERROR msg="open failed" err="file does not exist"
level=INFO msg=started request=42
level=INFO msg=done stats.files=3 stats.matches=7
{"level":"DEBUG","msg":"hello","user":"ana","tries":3}
{"level":"INFO","msg":"done","stats":{"files":3,"matches":7}}
```

The pieces:

- **Levels**: `Debug` < `Info` < `Warn` < `Error`. A handler drops records below its
  `HandlerOptions.Level`, which defaults to `Info`; that's why the `Debug` line is missing from
  the text output.
- **Attributes**: after the message come alternating keys and values (`"path", "notes.md"`), or
  typed attributes like `slog.Int("tries", 3)`, which avoid boxing values in `any`.
- **`TextHandler`** writes `key=value` pairs, quoting values with spaces. **`JSONHandler`** writes
  JSON Lines, ready for log-collection tools.
- **`ReplaceAttr`** can rewrite or drop any attribute; returning an empty `slog.Attr{}` drops it.
- **`With`** returns a child logger that adds attributes, like a request ID, to everything it
  logs. **`slog.Group`** nests attributes.
- **`slog.DiscardHandler`** throws everything away; a logger built on it costs almost nothing.
- **`slog.SetDefault(logger)`** makes your logger the one behind `slog.Info` and friends.

`go vet` catches a key with no value, like `logger.Info("x", "path")`: keys and values must come
in pairs.

**Why structured?** "searched notes.md, 0 matches" is easy to read but hard to filter.
`msg=searched path=notes.md matches=0` reads almost as well, and a tool can select `matches>0`
without fragile text parsing.

## Profiling with pprof

When code is slow, measure before changing anything. The profiler shows *where* the time goes,
and it's usually not where you'd guess.

`go test` can profile a benchmark. Here is the `-i` case of `seek`'s benchmark from
[Chapter 19](19-testing.md), run in the finished project:

```console
$ go test -run '^$' -bench /ignore-case -cpuprofile cpu.out ./internal/search
goos: linux
goarch: amd64
pkg: seek/internal/search
cpu: 13th Gen Intel(R) Core(TM) i9-13950HX
BenchmarkSearch/ignore-case-32         	    1219	    956289 ns/op	 107.71 MB/s
PASS
ok  	seek/internal/search	1.172s
$ ls
cmd  cpu.out  go.mod  internal  search.test  testdata
```

That leaves two files: `cpu.out`, the profile, and `search.test`, the test binary, which `pprof`
needs to turn addresses into function names. `go tool pprof` reads them; `-top` lists the
functions that used the most CPU:

```console
$ go tool pprof -top -nodecount 8 search.test cpu.out
File: search.test
Build ID: b5098d46cdc70bd3681ddf6fd31ef64fe1f474f4
Type: cpu
Time: 2026-10-27 11:24:53 EDT
Duration: 1.17s, Total samples = 1300ms (111.22%)
Showing nodes accounting for 790ms, 60.77% of 1300ms total
Showing top 8 nodes out of 123
      flat  flat%   sum%        cum   cum%
     160ms 12.31% 12.31%      500ms 38.46%  seek/internal/search.prefixFold
     140ms 10.77% 23.08%      140ms 10.77%  runtime.encoderune
     130ms 10.00% 33.08%      630ms 48.46%  seek/internal/search.indexFold
     100ms  7.69% 40.77%      100ms  7.69%  runtime.futex
     100ms  7.69% 48.46%      100ms  7.69%  strings.EqualFold
      80ms  6.15% 54.62%      220ms 16.92%  runtime.intstring
      40ms  3.08% 57.69%       40ms  3.08%  indexbytebody
      40ms  3.08% 60.77%       40ms  3.08%  runtime.memmove
```

- **flat**: time spent in the function itself.
- **cum** (cumulative): time in the function plus everything it calls.

`prefixFold` and `indexFold` are expected, but `runtime.encoderune` and `runtime.intstring` are a
surprise: together they take over a quarter of the time. `-list` shows the time per source line
of a function:

```console
$ go tool pprof -list 'search.prefixFold' search.test cpu.out
Total: 1.30s
ROUTINE ======================== seek/internal/search.prefixFold in /home/you/seek/internal/search/matcher.go
     160ms      500ms (flat, cum) 38.46% of Total
      20ms       20ms    108:func prefixFold(s, prefix string) int {
         .          .    109:	i := 0
      30ms       30ms    110:	for _, want := range prefix {
         .          .    111:		if i >= len(s) {
         .          .    112:			return -1
         .          .    113:		}
      40ms       60ms    114:		got, size := utf8.DecodeRuneInString(s[i:])
      50ms      370ms    115:		if got != want && !strings.EqualFold(string(got), string(want)) {
      20ms       20ms    116:			return -1
         .          .    117:		}
         .          .    118:		i += size
         .          .    119:	}
         .          .    120:	return i
         .          .    121:}
```

Line 115 is the hot spot: `string(got)` and `string(want)` turn each rune into a string
(`intstring`, which calls `encoderune`) just to compare one character. Exercise 3 fixes that. Your
numbers will differ, but the profile's shape should be similar.

More pprof tools:

- `-memprofile mem.out` profiles memory allocations instead of CPU; `-benchmem` shows the totals.
- `go tool pprof -http=:8080 search.test cpu.out` opens a web UI with call graphs and flame graphs.
- `runtime/pprof` profiles a whole program: call `pprof.StartCPUProfile(f)` at the start and
  `pprof.StopCPUProfile()` at the end.
- `net/http/pprof` adds profiling endpoints to a long-running server, so you can profile it live.

Delete the profile and the test binary when you're done:

```console
$ rm cpu.out search.test
```

## `unsafe` and cgo

Go has two escape hatches. Know what they are, and avoid them.

**`unsafe`** steps around Go's type system: `unsafe.Pointer` converts any pointer to any other
pointer type and allows pointer arithmetic. Such code can corrupt memory, break with a new Go
version, and isn't portable. The standard library uses it in a few hot spots; you almost never
need it.

The harmless part of the package is three functions that report memory layout. They show
something useful: the order of struct fields affects the struct's size.

`examples/ch20/unsafe/main.go`

```go
package main

import (
	"fmt"
	"unsafe"
)

type loose struct {
	a bool  // 1 byte, then 7 bytes of padding
	b int64 // 8 bytes, must start at a multiple of 8
	c bool  // 1 byte, then 7 bytes of padding
}

type tight struct {
	b int64
	a bool
	c bool // a and c share the padding at the end
}

func main() {
	fmt.Println("int:   ", unsafe.Sizeof(int(0)))
	fmt.Println("string:", unsafe.Sizeof(""))
	fmt.Println("slice: ", unsafe.Sizeof([]int{}))
	fmt.Println("loose: ", unsafe.Sizeof(loose{}), "offset of b:", unsafe.Offsetof(loose{}.b))
	fmt.Println("tight: ", unsafe.Sizeof(tight{}), "offset of a:", unsafe.Offsetof(tight{}.a))
}
```

```console
$ go run ./ch20/unsafe
int:    8
string: 16
slice:  24
loose:  24 offset of b: 8
tight:  16 offset of a: 8
```

These are 64-bit sizes. A string is a pointer plus a length (16 bytes), and a slice is a pointer,
a length, and a capacity (24 bytes), however much data they refer to. That's why passing them
around is cheap.

**cgo** lets Go call C. You write C declarations in a comment right above `import "C"`:

```go
package main

// #include <stdlib.h>
import "C"

import "fmt"

func main() {
	// C.abs is the C library's abs function.
	fmt.Println(C.abs(-7))
}
```

```console
$ go run .
7
```

(This one isn't in `examples/`, because it needs a C compiler.) The costs:

- Builds need a C compiler, and cross-compiling needs one for each target.
- Each call between Go and C is much slower than a Go function call.
- Memory rules get complicated: C can't keep Go pointers, and the garbage collector doesn't see C
  memory.
- With cgo turned off, files that use it are excluded, and the build fails:

  ```console
  $ CGO_ENABLED=0 go run .
  package e20/c: build constraints exclude all Go files in /home/you/e20/c
  ```

Use cgo when you must use an existing C library that has no Go alternative. A pure-Go program is
easier to build, ship, and debug.

## Go's design ideals

You've now seen the whole language, and its features follow from a small set of ideas. Rob Pike
collected many of them as the **Go Proverbs**. Here they are, with a few other common Go sayings,
and where `seek` follows them.

**Simplicity and clarity**

- **Clear is better than clever.** `run` reads top to bottom: parse, build the parts, loop over
  results, pick an exit code. No framework, no hidden control flow.
- **Orthogonality.** Go has few features, and they combine freely. There are no classes: structs,
  methods, interfaces, and embedding cover what classes do elsewhere.
- **gofmt's style is no one's favorite, yet gofmt is everyone's favorite.** No arguments about
  formatting; every Go file looks the same.
- **Documentation is for users.** Every exported type and function in `seek` has a doc comment
  that says what it does, not how.

**Types and interfaces**

- **The bigger the interface, the weaker the abstraction.** `Matcher` and `Printer` have one
  method each. Three matchers and two printers implement them, and tests can fake them.
- **Accept interfaces, return structs.** `run` takes an `io.Reader` and `io.Writer`s, so tests can
  pass buffers. `NewJSONPrinter` returns a concrete `*JSONPrinter`.
- **Composition over inheritance.** `countingReader` wraps any `io.Reader` and adds counting;
  `bufio.Writer` wraps `os.Stdout` and adds buffering.
- **`interface{}` says nothing.** (Today it's spelled `any`.) Generics let `stats.Counter[K]` keep
  its key type instead of falling back to `any`.
- **Make the zero value useful.** `var matched stats.Counter[string]` is ready to use, as are
  `sync.WaitGroup`, `bytes.Buffer`, and `strings.Builder`.

**Errors**

- **Errors are values.** `ErrBinary` is a variable you compare with `errors.Is`; `usageError` is a
  type you can match with `errors.As`. They're ordinary data, not a separate control-flow system.
- **Don't just check errors, handle them gracefully.** `seek` reports a bad file, remembers it for
  the exit code, and keeps searching. Wrapping with `%w` adds context without losing the cause.
- **Don't panic.** Panics are for bugs; `seek` never panics on bad input.

**Concurrency**

- **Concurrency is not parallelism.** `SearchAll` is a concurrent *design* (a producer, workers,
  and a consumer). Whether it runs in parallel depends on `-j` and the CPU; it's correct with
  `-j 1`.
- **Don't communicate by sharing memory; share memory by communicating.** Workers hand results
  back on channels instead of writing to a shared slice.
- **Channels orchestrate; mutexes serialize.** Channels structure the pipeline. The one shared
  counter, `bytesRead`, is an `atomic.Int64`.

**Dependencies and escape hatches**

- **A little copying is better than a little dependency.** `seek` uses only the standard library.
  `indexFold` is 30 lines, so writing it was cheaper than depending on a module.
- **Cgo is not Go.** **With the unsafe package there are no guarantees.** **Reflection is never
  clear.** All three are sometimes needed, rarely by you. (`encoding/json` uses reflection so your
  code doesn't have to.)

## Common mistakes: a recap

The mistakes that catch most Go programmers, and where this course covers them:

| Mistake | What happens | Fix | See |
|---------|--------------|-----|-----|
| `x := ...` inside a block when you meant `x = ...` | A new variable shadows the outer one, which keeps its old value | Use `=` to assign to the outer one | [Ch. 2](02-values-and-types.md) |
| Writing to a nil map | Panic: `assignment to entry in nil map` | `make(map[K]V)` or a literal | [Ch. 5](05-collections.md) |
| Keeping a slice after `append` to another slice sharing its array | Two slices change each other's elements | Copy with `slices.Clone`, or don't share | [Ch. 5](05-collections.md) |
| `s[i]` on a string, expecting a character | You get a byte, not a rune | `for i, r := range s`, or `[]rune(s)` | [Ch. 6](06-strings.md) |
| `defer` inside a long loop | Nothing is released until the function returns | Move the loop body into a function | [Ch. 4](04-functions.md) |
| A method with a value receiver that changes fields | It changes a copy | Use a pointer receiver | [Ch. 7](07-structs-methods-pointers.md) |
| Returning a nil `*MyError` as an `error` | The interface is not nil; `err != nil` is true | Return a literal `nil` | [Ch. 9](09-errors.md) |
| Comparing errors with `==` | Wrapped errors don't match | `errors.Is`, `errors.As` | [Ch. 9](09-errors.md) |
| Forgetting to `Flush` a `bufio.Writer` | The last chunk of output is lost | `defer w.Flush()` | [Ch. 11](11-io-and-files.md) |
| Starting a goroutine with no way to stop | A goroutine leak: it waits forever | Close channels; use `ctx.Done()` | [Ch. 16](16-goroutines-and-channels.md) |
| Sending on a closed channel, or closing twice | Panic | Only the sender closes, once | [Ch. 16](16-goroutines-and-channels.md) |
| Sharing a variable between goroutines without sync | A data race: wrong results, sometimes | Mutex, atomic, or channels; test with `-race` | [Ch. 17](17-sync.md) |
| Copying a struct that contains a `sync.Mutex` | Two locks, no protection | Pointers; `vet` reports it | [Ch. 17](17-sync.md) |
| Not calling a context's `cancel` | Its timer and resources leak until the parent ends | `defer cancel()`; `vet` reports it | [Ch. 18](18-context-and-time.md) |
| `-X` on a constant or a misspelled name | The version silently doesn't change | A package-level `var`; check `--version` | this chapter |

Two tools catch many of these. Run them often:

```console
$ go vet ./...
$ go test -race ./...
```

## Building seek: shipping it

`seek` gets its last four changes:

- The help text moves into `usage.txt`, embedded with `go:embed`.
- The version is set with `-ldflags -X`, with `debug.ReadBuildInfo` as a fallback.
- `--debug` turns on structured logging to standard error with `log/slog`.
- You cross-compile it for three other systems.

### The help text: `cmd/seek/usage.txt`

Help text in a Go string constant is awkward to edit: it sits inside backquotes, between code.
Move it to its own file and add the new `--debug` line. Create `cmd/seek/usage.txt` (27 lines,
ending with a newline):

```text
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

In `cmd/seek/flags.go`, add `_ "embed"` to the imports:

```go
import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
```

Then replace the `usage` constant (the whole backquoted string) with an embedded variable. Keep
`usageLine` as a constant, because error messages still print it:

```go
// usageLine is the first line of usage.txt, printed after usage errors.
const usageLine = "Usage: seek [OPTIONS] PATTERN [PATH ...]"

// usage is the text printed by --help. The go:embed directive makes the
// compiler store the contents of usage.txt in the variable, so the help
// text can be edited as a normal file but still ships inside the binary.
// The blank import of embed above is required for this.
//
//go:embed usage.txt
var usage string
```

Nothing makes sure the file and the constant agree, so add a test. In `cmd/seek/flags_test.go`,
add `"strings"` to the imports and this test at the end of the file:

```go
func TestUsage(t *testing.T) {
	// The embedded help text must start with the line printed after
	// usage errors, and mention every long option.
	if !strings.HasPrefix(usage, usageLine+"\n") {
		t.Errorf("usage.txt doesn't start with %q", usageLine)
	}
	for _, opt := range []string{"--ignore-case", "--timeout", "--json", "--debug"} {
		if !strings.Contains(usage, opt) {
			t.Errorf("usage.txt doesn't mention %s", opt)
		}
	}
}
```

The old constant had no final newline, so `run` printed it with `Fprintln`. The file has one, so
use `Fprint` in `run`:

```go
	opts, err := parseArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(out, usage) // usage.txt ends with a newline already
		return exitOK
	}
```

### The `--debug` flag: `cmd/seek/flags.go`

Add a field to `options`, between `stats` and `version`:

```go
	json        bool          // --json
	stats       bool          // --stats
	debug       bool          // --debug
	version     bool          // --version
}
```

And register it in `parseArgs`, after `stats`:

```go
	boolVar(&opts.stats, "stats", "")
	boolVar(&opts.debug, "debug", "")
	boolVar(&opts.version, "version", "")
```

### The version: `cmd/seek/main.go`

`version` was a constant, which `-X` can't change. Replace the `const` block that held `name` and
`version` with this:

```go
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
```

Add `"log/slog"` and `"runtime/debug"` to the imports. The full import block is now:

```go
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
```

### Debug logging: `cmd/seek/main.go`

Add `newLogger` after `run`, before `printStats`:

```go
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
```

When `--debug` is off, the logger discards everything, so the rest of `run` can log
unconditionally, with no `if opts.debug` checks scattered around.

Now use it in `run`: use `versionString()` for `--version`, and create the logger right after:

```go
	if opts.version {
		fmt.Fprintln(out, name, versionString())
		return exitOK
	}

	logger := newLogger(stderr, opts.debug)
	logger.Debug("starting", "pattern", opts.pattern, "paths", opts.paths, "workers", opts.workers)

	if opts.timeout > 0 {
```

Log each file in the results loop, which is now:

```go
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
```

And log a summary between the cancellation check and the statistics:

```go
	if err := context.Cause(ctx); err != nil {
		out.Flush() // print the results found so far first
		report(stderr, err)
		hadError = true
	}

	logger.Debug("done", "searched", searched, "matches", matched.Total())

	if opts.stats {
```

The logging happens in the consumer loop, not the workers. The consumer sees results in walk
order, so the log has the same order for any `-j`.

### Checkpoint

```console
$ go vet ./...
$ go test ./...
ok  	seek/cmd/seek	0.006s
ok  	seek/internal/output	0.002s
ok  	seek/internal/search	0.007s
ok  	seek/internal/stats	0.002s
$ go run ./cmd/seek --version
seek (devel)
$ go build -ldflags "-X main.version=1.0.0" -o seek ./cmd/seek && ./seek --version
seek 1.0.0
$ rm seek
```

If your project is in a git repository, `go run` prints a version from git instead of `(devel)`,
such as `v0.0.0-20261027150405-abcdef123456+dirty`.

Now the debug log:

```console
$ go run ./cmd/seek --debug -n fox testdata
level=DEBUG msg=starting pattern=fox paths=[testdata] workers=32
level=DEBUG msg="skipped binary file" path=testdata/image.bin
level=DEBUG msg=searched path=testdata/notes.md matches=0
level=DEBUG msg=searched path=testdata/poem.txt matches=2
level=DEBUG msg=searched path=testdata/src/hello.go matches=1
level=DEBUG msg=done searched=3 matches=3
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek --debug --hidden -c fox testdata/.hidden.txt testdata/image.bin
level=DEBUG msg=starting pattern=fox paths="[testdata/.hidden.txt testdata/image.bin]" workers=32
level=DEBUG msg=searched path=testdata/.hidden.txt matches=1
level=DEBUG msg="skipped binary file" path=testdata/image.bin
level=DEBUG msg=done searched=1 matches=1
testdata/.hidden.txt:1
```

- `workers=32` is the number of CPUs on the machine that ran this; yours may differ.
- The log lines come before the results: the log goes straight to standard error, but results
  wait in the `bufio.Writer` until it's flushed at the end.
- `slog` quotes values that contain spaces: `msg="skipped binary file"`, `paths="[a b]"`.

The help text now comes from `usage.txt`:

```console
$ go run ./cmd/seek --help
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

Finally, build for three other systems. `seek` is pure Go, so no extra tools are needed:

```console
$ GOOS=windows GOARCH=amd64 go build -o dist/seek-windows-amd64.exe ./cmd/seek
$ GOOS=darwin GOARCH=arm64 go build -o dist/seek-darwin-arm64 ./cmd/seek
$ GOOS=linux GOARCH=arm64 go build -o dist/seek-linux-arm64 ./cmd/seek
$ ls dist
seek-darwin-arm64
seek-linux-arm64
seek-windows-amd64.exe
$ go version -m dist/seek-darwin-arm64 | head -4
dist/seek-darwin-arm64: go1.26.10
	path	seek/cmd/seek
	mod	seek	(devel)	
	build	-buildmode=exe
$ rm -r dist
```

The complete project at this point is in `project/ch20/`.

### A tour of the finished code

`seek` is 1,113 lines of Go plus 810 lines of tests, and it uses only the standard library.

```text
seek/
  go.mod                       module seek
  cmd/seek/
    main.go         264        main, run, version, logging, statistics
    flags.go        179        options, parseArgs, --color, usage errors
    usage.txt        27        the embedded --help text
    main_test.go                end-to-end tests with golden files in testdata/
    flags_test.go
  internal/search/
    matcher.go      121        Matcher, NewMatcher, literal and case-folding matchers
    search.go       119        Searcher: search one file or reader, binary detection
    walk.go          82        Walk: an iterator over the files to search
    parallel.go     112        SearchAll: worker pool with results in walk order
  internal/output/
    output.go        99        Printer, TextPrinter, highlighting
    json.go          63        JSONPrinter
  internal/stats/
    counter.go       74        generic Counter[K] and TopN
```

A search flows through it like this:

1. `main` creates a context that Ctrl-C cancels, calls `run`, and exits with its result.
2. `run` parses the arguments into `options` and builds a `Matcher`, a `Searcher`, and a
   `Printer`.
3. `search.Walk` yields file paths one at a time, skipping hidden files and applying `--glob`.
4. `SearchAll` hands each path to a pool of workers. Each worker runs `SearchFile`, which reads
   the file line by line and asks the `Matcher` about each line.
5. Results come back in walk order. `run` prints each with the `Printer` and counts matches in a
   `stats.Counter`.
6. `run` reports a timeout or cancellation, prints `--stats`, and picks the exit status.

[Appendix B](appendix-b-project-reference.md) has a diagram and maps every concept in this course
to the place `seek` uses it. Read the code from `main` down; you know every line of it now.

## Where to go next

You've covered the language and much of the standard library. Next steps:

- **The language specification**, [go.dev/ref/spec](https://go.dev/ref/spec). It's short and
  readable; look up the exact rules whenever something surprises you.
- **Effective Go**, [go.dev/doc/effective_go](https://go.dev/doc/effective_go), and **Go Code
  Review Comments**, [go.dev/wiki/CodeReviewComments](https://go.dev/wiki/CodeReviewComments), for
  idiomatic style.
- **The memory model**, [go.dev/ref/mem](https://go.dev/ref/mem), for the exact rules behind
  [Chapter 17](17-sync.md).
- **The standard library source.** `go doc -src strings.Index` prints the code of any function.
  It's well written and a good model for your own code.
- **The Go blog**, [go.dev/blog](https://go.dev/blog), for in-depth articles on features, and
  the release notes at [go.dev/doc/devel/release](https://go.dev/doc/devel/release) for what each
  version adds.
- **Packages to learn next**: `net/http` (web servers and clients), `database/sql`,
  `text/template` and `html/template`, `os/exec`, and `golang.org/x/sync/errgroup`.
- **Tools**: `staticcheck` (more checks than `vet`), `golang.org/x/tools/cmd/stringer`, and
  `govulncheck`, which reports known vulnerabilities in your dependencies.
- **Build something.** Ideas that grow out of `seek`: a `--replace` mode that rewrites files, a
  file watcher that reruns a search when files change, or an HTTP server that searches a
  directory and returns JSON.

That's the end of the course. You started with `fmt.Println` and finished with a concurrent,
tested, cross-platform tool. Go rewards plain code: write the simple version first, measure, and
keep it readable.

## Summary

- `//go:embed` puts files into the binary as a `string`, `[]byte`, or `embed.FS`; missing files
  are compile errors. Import `embed`, with `_` if you don't use it by name.
- Build constraints choose whole files: suffixes like `_windows.go`, or `//go:build` expressions.
  `-tags` turns on custom tags; `ignore` keeps a file out of every build.
- `-ldflags "-X main.version=1.0.0"` sets a package-level string variable at link time, and
  silently does nothing for constants and wrong names. `debug.ReadBuildInfo` and
  `go version -m` report what Go recorded about the build.
- `GOOS` and `GOARCH` cross-compile pure-Go programs with no extra tools. `go install` puts
  binaries in `$GOBIN` or `$(go env GOPATH)/bin`.
- `go generate` runs `//go:generate` commands. `go build` never runs it, so commit the generated
  files.
- `log/slog` writes structured, leveled records through a handler: `TextHandler`, `JSONHandler`,
  or `DiscardHandler`. `With` and groups add context.
- Profile before you optimize: `go test -cpuprofile`, then `go tool pprof -top` and `-list`.
- `unsafe` and cgo exist for rare needs. Avoid them; pure Go is portable and easy to ship.
- Go's ideals: clear over clever, small interfaces, useful zero values, errors as values, share
  memory by communicating, a little copying over a little dependency.

## Exercises

1. Make `--version` also print the Go version that built `seek`, like `seek 1.0.0 (go1.26.10)`.

   <details><summary>Solution</summary>

   `runtime.Version()` returns the Go version. In `run`:

   ```go
   	if opts.version {
   		fmt.Fprintf(out, "%s %s (%s)\n", name, versionString(), runtime.Version())
   		return exitOK
   	}
   ```

   and add `"runtime"` to the imports. (`debug.ReadBuildInfo`'s `GoVersion` field works too.)

   ```console
   $ go run ./cmd/seek --version
   seek (devel) (go1.26.10)
   $ go build -ldflags "-X main.version=1.0.0" -o seek ./cmd/seek && ./seek --version
   seek 1.0.0 (go1.26.10)
   $ rm seek
   ```

   </details>

2. When both `--debug` and `--json` are given, write the debug log as JSON too, so one tool can
   read everything.

   <details><summary>Solution</summary>

   Pass `opts.json` to `newLogger` and choose the handler there; both handlers take the same
   options:

   ```go
   func newLogger(w io.Writer, enabled, asJSON bool) *slog.Logger {
   	if !enabled {
   		return slog.New(slog.DiscardHandler)
   	}
   	opts := &slog.HandlerOptions{
   		Level: slog.LevelDebug, // the default level, Info, would hide Debug
   		// Drop the time from each record: it is noise for a short run
   		// and would make the output differ every time.
   		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
   			if a.Key == slog.TimeKey && len(groups) == 0 {
   				return slog.Attr{} // an empty Attr is left out
   			}
   			return a
   		},
   	}
   	if asJSON {
   		return slog.New(slog.NewJSONHandler(w, opts))
   	}
   	return slog.New(slog.NewTextHandler(w, opts))
   }
   ```

   In `run`: `logger := newLogger(stderr, opts.debug, opts.json)`.

   ```console
   $ go run ./cmd/seek --debug --json fox testdata/poem.txt
   {"level":"DEBUG","msg":"starting","pattern":"fox","paths":["testdata/poem.txt"],"workers":32}
   {"level":"DEBUG","msg":"searched","path":"testdata/poem.txt","matches":2}
   {"level":"DEBUG","msg":"done","searched":1,"matches":2}
   {"path":"testdata/poem.txt","line":1,"text":"The quick brown fox"}
   {"path":"testdata/poem.txt","line":3,"text":"The dog sleeps; the fox runs."}
   ```

   `paths` is now a real JSON array.

   </details>

3. The profile above showed that `prefixFold` converts each rune to a string just to compare it.
   Write `equalFoldRune(a, b rune) bool`, which compares two runes ignoring case without
   converting them, and use it in `prefixFold`. Check the result with the fuzz test and the
   benchmark. Hint: `unicode.SimpleFold(r)` returns the next rune in `r`'s case-folding "orbit",
   for example `k` → `K` (Kelvin sign) → `K` → `k`. It's how `strings.EqualFold` compares
   runes.

   <details><summary>Solution</summary>

   In `internal/search/matcher.go`, add `"unicode"` to the imports, change the comparison in
   `prefixFold` to

   ```go
   		if !equalFoldRune(got, want) {
   ```

   and add:

   ```go
   // equalFoldRune reports whether a and b are equal under simple Unicode
   // case folding, like strings.EqualFold for one-rune strings, but
   // without converting the runes to strings.
   func equalFoldRune(a, b rune) bool {
   	if a == b {
   		return true
   	}
   	if a < utf8.RuneSelf && b < utf8.RuneSelf {
   		// Both are ASCII: only letters have another case.
   		if 'A' <= a && a <= 'Z' {
   			a += 'a' - 'A'
   		}
   		if 'A' <= b && b <= 'Z' {
   			b += 'a' - 'A'
   		}
   		return a == b
   	}
   	// SimpleFold steps through the runes that are equal to a when
   	// case is ignored, for example k -> K -> K (Kelvin) -> k.
   	for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
   		if r == b {
   			return true
   		}
   	}
   	return false
   }
   ```

   The fuzz test checks every match `indexFold` finds with `strings.EqualFold`, so it tests the
   new code against the standard library:

   ```console
   $ go test ./...
   ok  	seek/cmd/seek	0.006s
   ok  	seek/internal/output	0.002s
   ok  	seek/internal/search	0.007s
   ok  	seek/internal/stats	0.002s
   $ go test -run '^$' -fuzz FuzzIndexFold -fuzztime 15s ./internal/search
   ...
   PASS
   ok  	seek/internal/search	16.065s
   $ go test -run '^$' -bench /ignore-case -cpuprofile cpu.out ./internal/search | grep Bench
   BenchmarkSearch/ignore-case-32         	    1626	    728154 ns/op	 141.45 MB/s
   $ go tool pprof -top -nodecount 6 search.test cpu.out | tail -7
         flat  flat%   sum%        cum   cum%
        200ms 13.70% 13.70%      370ms 25.34%  seek/internal/search.prefixFold
        180ms 12.33% 26.03%      180ms 12.33%  runtime.futex
        160ms 10.96% 36.99%      160ms 10.96%  seek/internal/search.equalFoldRune
         90ms  6.16% 43.15%      460ms 31.51%  seek/internal/search.indexFold
         70ms  4.79% 47.95%       70ms  4.79%  runtime.memmove
         70ms  4.79% 52.74%       70ms  4.79%  runtime.nextFreeFast
   $ rm cpu.out search.test
   ```

   `intstring` and `encoderune` are gone, and the benchmark went from about 950 µs to about
   730 µs per run on this machine. Benchmarks are noisy on a busy computer, so run each version
   several times (`-count 5`) before trusting a difference. The profile now points at the real
   remaining cost: `indexFold` tries `prefixFold` at every rune offset of the line.

   </details>

4. Write a shell script `release.sh VERSION` that builds `seek` for `linux/amd64`, `linux/arm64`,
   `darwin/arm64`, and `windows/amd64` into `dist/`, with the version stamped in and local paths
   trimmed. Windows binaries need an `.exe` suffix.

   <details><summary>Solution</summary>

   ```sh
   #!/bin/sh
   # release.sh VERSION builds seek for several platforms into dist/.
   set -e
   version=$1
   for target in linux/amd64 linux/arm64 darwin/arm64 windows/amd64; do
   	os=${target%/*}
   	arch=${target#*/}
   	out=dist/seek-$version-$os-$arch
   	if [ "$os" = windows ]; then
   		out=$out.exe
   	fi
   	GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-X main.version=$version" -o "$out" ./cmd/seek
   	echo "built $out"
   done
   ```

   ```console
   $ sh release.sh 1.0.0
   built dist/seek-1.0.0-linux-amd64
   built dist/seek-1.0.0-linux-arm64
   built dist/seek-1.0.0-darwin-arm64
   built dist/seek-1.0.0-windows-amd64.exe
   $ ./dist/seek-1.0.0-linux-amd64 --version
   seek 1.0.0
   $ rm -r dist
   ```

   (Run the `linux-amd64` binary only on Linux/amd64; on another system, run the one built for
   it.)

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-20--shipping-it-tooling-embedding-and-gos-design-ideals)
