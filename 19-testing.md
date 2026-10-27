# Chapter 19 — Testing, Benchmarks, and Fuzzing

You've written two small tests already, in [Chapter 17](17-sync.md) and
[Chapter 18](18-context-and-time.md). This chapter covers testing properly: `go test`,
table-driven tests and subtests, helpers, temporary files, tested examples, in-memory file
systems, benchmarks, fuzzing, and coverage. Go needs no test framework, just the `testing` package
and a few conventions. `seek` gets a test suite of about 800 lines, which finds a real bug. [Chapter 20](20-tooling-and-design.md) then ships the finished program.

## `go test` and `_test.go` files

Tests live next to the code they test, in files ending in `_test.go`. `go build` ignores those
files; `go test` compiles them with the package and runs every function of this form:

```go
func TestXxx(t *testing.T)
```

`Xxx` must start with an upper-case letter (or a digit, or `_`). Here is a package with one
function:

`examples/ch19/first/abs.go`

```go
// Package first holds the function tested in the first example.
package first

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
```

and its test, in the same directory and package:

`examples/ch19/first/abs_test.go`

```go
package first

import "testing"

func TestAbs(t *testing.T) {
	if got := Abs(-3); got != 3 {
		t.Errorf("Abs(-3) = %d, want 3", got)
	}
	if got := Abs(5); got != 5 {
		t.Errorf("Abs(5) = %d, want 5", got)
	}
	if got := Abs(0); got != 0 {
		t.Errorf("Abs(0) = %d, want 0", got)
	}
}
```

```console
$ go test ./ch19/first
ok  	examples/ch19/first	0.002s
$ go test -v ./ch19/first
=== RUN   TestAbs
--- PASS: TestAbs (0.00s)
PASS
ok  	examples/ch19/first	0.002s
```

`-v` (verbose) lists each test. Run it again and you'll see `(cached)` instead of a time: `go test`
reruns only tests whose code or inputs changed. `-count=1` forces a real run.

To see a failure, break `Abs` by changing `return -x` to `return x`:

```console
$ go test ./ch19/first
--- FAIL: TestAbs (0.00s)
    abs_test.go:7: Abs(-3) = -3, want 3
FAIL
FAIL	examples/ch19/first	0.002s
FAIL
```

The message gives the file and line, the call, the result, and the expected value. Go has no
assertion library on purpose: a plain `if` and a clear message, usually in the form
`Function(input) = got, want expected`, are enough.

The `*testing.T` methods you'll use most:

| Method | Effect |
| --- | --- |
| `t.Error(args...)`, `t.Errorf(format, args...)` | Report a failure; the test keeps going. |
| `t.Fatal(args...)`, `t.Fatalf(format, args...)` | Report a failure and stop this test now. |
| `t.Log(args...)`, `t.Logf(format, args...)` | Print a note; shown with `-v` or when the test fails. |
| `t.Skip(args...)` | Stop the test and mark it skipped, not failed. |

Use `Error` when later checks still make sense, so one run shows every problem; use `Fatal` when
continuing is pointless, usually after an unexpected `err != nil`. Call `Fatal` only from the
test's own goroutine.

## Table-driven tests and subtests

`TestAbs` repeats itself. Go style puts the cases in a table (a slice of structs) and loops over
it, with `t.Run` running each case as a named *subtest*:

`examples/ch19/reverse/reverse.go`

```go
// Package reverse reverses strings.
package reverse

// String returns s with its runes in reverse order.
func String(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
```

`examples/ch19/reverse/reverse_test.go`

```go
package reverse

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"one rune", "a", "a"},
		{"ascii", "hello", "olleh"},
		{"palindrome", "racecar", "racecar"},
		{"multi-byte", "héllo, 世界", "界世 ,olléh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := String(tt.in); got != tt.want {
				t.Errorf("String(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

Adding a case is now one line. The struct type is anonymous because it's used only here; `tt` is
the usual name for the current row.

`t.Run(name, f)` runs `f` as a subtest with its own `t`. A `Fatal` stops only that subtest, and
each subtest is reported by name. `t.Parallel()` marks a test as safe to run alongside other
parallel tests. The subtests pause, then continue together once the loop ends:

```console
$ go test -v ./ch19/reverse
=== RUN   TestString
=== RUN   TestString/empty
=== PAUSE TestString/empty
=== RUN   TestString/one_rune
=== PAUSE TestString/one_rune
=== RUN   TestString/ascii
=== PAUSE TestString/ascii
=== RUN   TestString/palindrome
=== PAUSE TestString/palindrome
=== RUN   TestString/multi-byte
=== PAUSE TestString/multi-byte
=== CONT  TestString/empty
=== CONT  TestString/ascii
=== CONT  TestString/one_rune
=== CONT  TestString/multi-byte
=== CONT  TestString/palindrome
--- PASS: TestString (0.00s)
    --- PASS: TestString/empty (0.00s)
    --- PASS: TestString/ascii (0.00s)
    --- PASS: TestString/one_rune (0.00s)
    --- PASS: TestString/multi-byte (0.00s)
    --- PASS: TestString/palindrome (0.00s)
=== RUN   ExampleString
--- PASS: ExampleString (0.00s)
PASS
ok  	examples/ch19/reverse	0.002s
```

The `CONT` order changes from run to run. Spaces in subtest names become underscores. (Since Go
1.22, each loop iteration has its own `tt`, so the parallel closures don't share a variable.)

`-run` selects tests by regular expression, with `/` separating levels. This runs one subtest:

```console
$ go test -v -run 'TestString/multi' ./ch19/reverse
=== RUN   TestString
=== RUN   TestString/multi-byte
=== PAUSE TestString/multi-byte
=== CONT  TestString/multi-byte
--- PASS: TestString (0.00s)
    --- PASS: TestString/multi-byte (0.00s)
PASS
ok  	examples/ch19/reverse	0.002s
```

## Black-box tests and Examples

A test file can declare package `reverse` (as above) or `reverse_test`. The second is an
*external test package*: compiled separately, it must import `reverse` like any other user and
sees only the exported API. Same-package (white-box) tests can reach unexported names; external
(black-box) tests check that the public API is enough. Both can live in the same directory.

An external test package is also the natural home for **Examples**: functions named `ExampleXxx`
with an `// Output:` comment.

`examples/ch19/reverse/example_test.go`

```go
// An external test package: it can only use reverse's exported names,
// just like any other user of the package.
package reverse_test

import (
	"fmt"

	"examples/ch19/reverse"
)

func ExampleString() {
	fmt.Println(reverse.String("stressed"))
	// Output: desserts
}
```

`go test` runs the example and compares its output with the `Output` comment (it passed in the
`-v` run above), and pkg.go.dev shows it next to `String`'s documentation: documentation that
every test run checks. Naming: `Example` (the package), `ExampleF` (function
`F`), `ExampleT` (type `T`), `ExampleT_M` (method `M` of `T`), plus an optional `_suffix` for more
than one. An example without an `Output` comment is compiled but not run.

## Helpers, cleanup, temporary files, and `testdata`

Tests that touch files need a few more tools. This package counts the lines in a file:

`examples/ch19/files/files.go`

```go
// Package files counts lines in files.
package files

import (
	"bufio"
	"os"
)

// CountLines returns the number of lines in the file at path.
func CountLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}
```

`examples/ch19/files/files_test.go`

```go
package files

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates a file holding text in a fresh temporary directory
// and returns its path.
func writeFile(t *testing.T, text string) string {
	t.Helper() // report failures at the caller's line
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCountLines(t *testing.T) {
	t.Cleanup(func() { t.Log("cleanup runs after the test and its subtests") })

	path := writeFile(t, "one\ntwo\nthree\n")
	n, err := CountLines(path)
	if err != nil {
		t.Fatal(err) // stop: n is meaningless if err != nil
	}
	if n != 3 {
		t.Errorf("CountLines = %d, want 3", n)
	}
	t.Log("checked", path)
}

func TestCountLinesTestdata(t *testing.T) {
	// go test runs in the package's directory, so testdata/ is right here.
	n, err := CountLines(filepath.Join("testdata", "poem.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("CountLines = %d, want 4", n)
	}
}

func TestCountLinesMissing(t *testing.T) {
	_, err := CountLines(filepath.Join(t.TempDir(), "missing.txt"))
	if !os.IsNotExist(err) {
		t.Errorf("got error %v, want a not-exist error", err)
	}
}
```

```console
$ go test -v ./ch19/files
=== RUN   TestCountLines
    files_test.go:31: checked /tmp/TestCountLines1981375858/001/file.txt
    files_test.go:21: cleanup runs after the test and its subtests
--- PASS: TestCountLines (0.00s)
=== RUN   TestCountLinesTestdata
--- PASS: TestCountLinesTestdata (0.00s)
=== RUN   TestCountLinesMissing
--- PASS: TestCountLinesMissing (0.00s)
PASS
ok  	examples/ch19/files	0.002s
```

The tools:

- **`t.TempDir()`** returns a new, empty directory (a different one per call) and deletes it when
  the test ends. Never write test files into the source tree.
- **`t.Cleanup(f)`** runs `f` when the test and all its subtests finish, last in, first out, like
  `defer`. Unlike `defer`, it works from inside a helper: the cleanup belongs to the test, not the
  helper. `t.TempDir` uses it to delete the directory.
- **`t.Helper()`** marks a function as a test helper, so when it calls `t.Fatal`, the report shows
  the *caller's* line. Make `writeFile` fail (say, by writing into a missing directory) and the
  report points at line 23, the call in `TestCountLines`:

  ```console
  $ go test ./ch19/files
  --- FAIL: TestCountLines (0.00s)
      files_test.go:23: open /tmp/TestCountLines729910340/001/no/such/dir/file.txt: no such file or directory
      files_test.go:21: cleanup runs after the test and its subtests
  FAIL
  FAIL	examples/ch19/files	0.002s
  FAIL
  ```

  Without `t.Helper()`, it would say `files_test.go:15`, the `t.Fatal` inside the helper: the
  same line for every caller, and much less useful.
- **The `testdata` directory.** `go test` runs each package's tests in that package's directory,
  so relative paths work. The Go tools ignore directories named `testdata`, which makes it the
  standard place for test input: `examples/ch19/files/testdata/poem.txt` here.

Three more methods:

- **`t.Context()`** (Go 1.24) returns a context canceled just before the test's cleanups run.
  Pass it to code under test instead of `context.Background()`.
- **`t.Setenv(key, value)`** sets an environment variable and restores it when the test ends.
- **`t.Chdir(dir)`** (Go 1.24) changes the working directory and changes back at the end.

The environment and working directory belong to the whole process, so a test that uses `Setenv`
or `Chdir` can't call `t.Parallel()`; the `testing` package panics if you try.

## Testing file-system code with `testing/fstest`

Code that walks real directories needs real files, made with `t.TempDir` and `os.WriteFile`. But
if the code takes an `fs.FS` (the read-only file-system interface from
[Chapter 11](11-io-and-files.md)) instead of a path, a test can pass an in-memory file system,
`fstest.MapFS`:

`examples/ch19/mapfs/mapfs.go`

```go
// Package mapfs finds files in any file system.
package mapfs

import (
	"io/fs"
	"path"
)

// FindExt returns the paths of the regular files in fsys whose names
// end in ext, such as ".go", in lexical order.
func FindExt(fsys fs.FS, ext string) ([]string, error) {
	var found []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() && path.Ext(p) == ext {
			found = append(found, p)
		}
		return nil
	})
	return found, err
}
```

`examples/ch19/mapfs/mapfs_test.go`

```go
package mapfs

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestFindExt(t *testing.T) {
	// A MapFS is an in-memory file system: a map from slash-separated
	// paths to files. Parent directories are implied.
	fsys := fstest.MapFS{
		"main.go":           {Data: []byte("package main\n")},
		"README.md":         {Data: []byte("# hi\n")},
		"internal/a/a.go":   {Data: []byte("package a\n")},
		"internal/a/a.txt":  {Data: []byte("notes\n")},
		"vendor/x/x.go":     {Data: []byte("package x\n")},
		"internal/b/doc.go": {Data: []byte("package b\n")},
	}
	got, err := FindExt(fsys, ".go")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/a/a.go", "internal/b/doc.go", "main.go", "vendor/x/x.go"}
	if !slices.Equal(got, want) {
		t.Errorf("FindExt = %q, want %q", got, want)
	}
}

// TestFS checks that a file system behaves correctly: fstest.TestFS is
// for people writing their own fs.FS implementations.
func TestFS(t *testing.T) {
	fsys := fstest.MapFS{"a/b.txt": {Data: []byte("b")}}
	if err := fstest.TestFS(fsys, "a/b.txt"); err != nil {
		t.Fatal(err)
	}
}
```

```console
$ go test ./ch19/mapfs
ok  	examples/ch19/mapfs	0.002s
```

No disk, no cleanup, and the whole tree is visible in the test. The real program would pass
`os.DirFS(dir)`, a directory on disk behind the same interface. (The callback's
parameter is `p`, not `path`, so it doesn't hide the `path` package used inside it.)

What about `seek`? Its `Walk` uses `filepath.WalkDir` on real paths. It *could* take an `fs.FS`:
add an `fsys fs.FS` parameter, and use `fs.WalkDir(fsys, root, ...)`, `fs.SkipDir`/`fs.SkipAll`,
and `path.Match` instead of the `filepath` versions. A test could then do:

```go
	fsys := fstest.MapFS{"b.txt": {}, "a.go": {}, ".git/config": {}, "sub/c.txt": {}}
	for path, err := range Walk(fsys, []string{"."}, WalkOptions{}) {
```

and get `a.go`, `b.txt`, `sub/c.txt`. (This was tried, and it works.) But `fs.FS` paths are always
relative, slash-separated, and can't contain `..`, while a user who types `seek fox ../src` or
`seek fox /etc` needs real OS paths. Rather than add a translation layer, `seek` keeps
`filepath.WalkDir` and tests `Walk` with `t.TempDir()`. The general rule: use `fs.FS` when your
code reads a tree it controls (embedded files, archives, test fixtures), and OS paths when the user
names the files.

## Benchmarks

A benchmark measures how long code takes. It's a function `BenchmarkXxx(b *testing.B)` in a
`_test.go` file, with the code to measure inside a `for b.Loop()` loop:

`examples/ch19/concat/concat.go`

```go
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
```

`examples/ch19/concat/concat_test.go`

```go
package concat

import (
	"strings"
	"testing"
)

var parts = strings.Fields(strings.Repeat("the quick brown fox ", 250)) // 1000 words

func BenchmarkPlus(b *testing.B) {
	for b.Loop() {
		Plus(parts)
	}
}

func BenchmarkBuilder(b *testing.B) {
	for b.Loop() {
		Builder(parts)
	}
}

func BenchmarkJoin(b *testing.B) {
	for b.Loop() {
		Join(parts)
	}
}

// Benchmarks don't check results, so test that all three agree.
func TestSame(t *testing.T) {
	want := Plus(parts)
	if Builder(parts) != want || Join(parts) != want {
		t.Error("results differ")
	}
}
```

`go test` runs benchmarks only when asked with `-bench`, which takes a regexp (`.` means all).
`-run '^$'` skips the ordinary tests (no name matches the empty string), and `-benchmem` reports
memory allocations:

```console
$ go test -run '^$' -bench . -benchmem ./ch19/concat
goos: linux
goarch: amd64
pkg: examples/ch19/concat
cpu: 13th Gen Intel(R) Core(TM) i9-13950HX
BenchmarkPlus-32       	    2257	    524931 ns/op	 2137661 B/op	     999 allocs/op
BenchmarkBuilder-32    	  240385	      5496 ns/op	   12536 B/op	      12 allocs/op
BenchmarkJoin-32       	  182502	      6616 ns/op	    4096 B/op	       1 allocs/op
PASS
ok  	examples/ch19/concat	3.719s
```

Reading a line: `BenchmarkPlus-32` ran with `GOMAXPROCS` = 32; the loop ran 2257 times; each
iteration took 524931 ns (about half a millisecond) and allocated about 2 MB in 999 allocations.
Your numbers will differ; what matters is the shape. `+=` makes a new string each time and copies
everything so far, so its cost grows with the square of the length: about 100 times slower here.
`Builder` and `Join` are close (they may swap places on another run), but `Join` allocates exactly
once because it computes the final length first.

`b.Loop` (Go 1.24) runs the body as many times as needed for a stable measurement, doesn't time
setup code *before* the loop, and keeps the compiler from optimizing away calls whose results are
unused. Older code uses `for i := 0; i < b.N; i++`, which has neither protection. Other
`*testing.B` methods: `b.ReportAllocs()` (like `-benchmem`, for one benchmark), `b.SetBytes(n)`
(report MB/s), `b.ResetTimer()`, and `b.Run` for sub-benchmarks.

To compare before and after a change, run the benchmark several times (`-count=10`) and feed
both outputs to `benchstat` (`golang.org/x/perf/cmd/benchstat`), which tells you whether a
difference is real or noise.

## Fuzzing

A table test checks the inputs you thought of; a **fuzz test** generates ones you didn't. You
state a *property* that must hold for every input, and the fuzzer mutates inputs looking for one
that breaks it.

Here is a run-length encoder: `"aaab"` becomes `"3a1b"`.

`examples/ch19/rle/rle.go`

```go
// Package rle implements run-length encoding: "aaab" becomes "3a1b".
package rle

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Encode replaces each run of a repeated rune with its length and the
// rune. s must be valid UTF-8 and contain no digits 0-9; otherwise the
// result can't be decoded.
func Encode(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); {
		j := i
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		fmt.Fprintf(&b, "%d%c", j-i, runes[i])
		i = j
	}
	return b.String()
}

// Decode reverses Encode.
func Decode(s string) (string, error) {
	var b strings.Builder
	for s != "" {
		// Only ASCII digits: unicode.IsDigit would also accept digits
		// from other scripts, such as '߇' (NKo seven), which Encode
		// treats as ordinary runes.
		i := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
		if i <= 0 {
			return "", fmt.Errorf("rle: expected a count then a rune at %q", s)
		}
		n, err := strconv.Atoi(s[:i])
		if err != nil {
			return "", err
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(strings.Repeat(string(r), n))
		s = s[i+size:]
	}
	return b.String(), nil
}
```

The obvious property: decoding an encoded string gives back the original.

`examples/ch19/rle/rle_test.go`

```go
package rle

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEncode(t *testing.T) {
	if got := Encode("aaabccdddd"); got != "3a1b2c4d" {
		t.Errorf("Encode = %q, want 3a1b2c4d", got)
	}
}

// FuzzRoundTrip checks that decoding an encoded string gives back the
// original, for any input Encode accepts.
func FuzzRoundTrip(f *testing.F) {
	f.Add("aaabccdddd") // seed corpus: typical inputs
	f.Add("")
	f.Add("héééllo")
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) || strings.ContainsAny(s, "0123456789") {
			t.Skip("Encode doesn't promise anything for this input")
		}
		got, err := Decode(Encode(s))
		if err != nil {
			t.Fatalf("Decode(Encode(%q)): %v", s, err)
		}
		if got != s {
			t.Errorf("Decode(Encode(%q)) = %q", s, got)
		}
	})
}
```

A fuzz test is a function `FuzzXxx(f *testing.F)`. `f.Add` adds *seed* inputs. `f.Fuzz` takes the
property as a function whose parameters after `t` are the fuzzed values (strings, `[]byte`,
numbers, and `bool` are supported).

This code is the result of two rounds of fuzzing. The first version of the test skipped only
invalid UTF-8, and `Decode` used `unicode.IsDigit`. Fuzzing runs with `-fuzz` and the target name:

```console
$ go test -run '^$' -fuzz FuzzRoundTrip -fuzztime 30s ./ch19/rle
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 32 workers
fuzz: elapsed: 0s, execs: 14 (125/sec), new interesting: 0 (total: 3)
--- FAIL: FuzzRoundTrip (0.11s)
    --- FAIL: FuzzRoundTrip (0.00s)
        rle_test.go:26: Decode(Encode("00")): rle: expected a count then a rune at "20"

    Failing input written to testdata/fuzz/FuzzRoundTrip/abff33899e7cd4f9
    To re-run:
    go test -run=FuzzRoundTrip/abff33899e7cd4f9
FAIL
exit status 1
FAIL	examples/ch19/rle	0.127s
```

In a fraction of a second, the fuzzer found that `"00"` encodes to `"20"`, which reads as "twenty
of... nothing": the format can't tell a count from a digit in the text. A failed property has two
fixes: fix the code, or narrow the promise. Here the format itself is the limit, so `Encode`'s
documentation now says "no digits 0-9", and the test skips such inputs.

The second run found something subtler:

```console
$ go test -run '^$' -fuzz FuzzRoundTrip -fuzztime 5s ./ch19/rle
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 32 workers
fuzz: minimizing 31-byte failing input file
fuzz: elapsed: 0s, minimizing
--- FAIL: FuzzRoundTrip (0.11s)
    --- FAIL: FuzzRoundTrip (0.00s)
        rle_test.go:27: Decode(Encode("߇")): rle: expected a count then a rune at "1߇"

    Failing input written to testdata/fuzz/FuzzRoundTrip/ebb0fad247a39841
    To re-run:
    go test -run=FuzzRoundTrip/ebb0fad247a39841
FAIL
exit status 1
FAIL	examples/ch19/rle	0.111s
```

`'߇'` is the digit seven in the NKo script. `unicode.IsDigit` says it's a digit, so `Decode` read
`1߇` as one long number with no rune after it. This time the code was wrong: the count is always
ASCII, so `Decode` now accepts only `'0'`–`'9'`. No one would have written that table row by hand.

When a fuzz test fails, Go *minimizes* the input (shrinks it to the smallest failing case) and
saves it under `testdata/fuzz/FuzzXxx/`:

```text
go test fuzz v1
string("߇")
```

Commit that file. From then on, a plain `go test` (no `-fuzz`) runs the fuzz target on the seeds
and every saved input, so the bug can't come back. `examples/ch19/rle/testdata/fuzz/` contains
it. After the fix:

```console
$ go test -run '^$' -fuzz FuzzRoundTrip -fuzztime 10s ./ch19/rle
fuzz: elapsed: 0s, gathering baseline coverage: 0/40 completed
fuzz: elapsed: 0s, gathering baseline coverage: 40/40 completed, now fuzzing with 32 workers
fuzz: elapsed: 3s, execs: 388437 (129451/sec), new interesting: 41 (total: 81)
fuzz: elapsed: 6s, execs: 395151 (2237/sec), new interesting: 41 (total: 81)
fuzz: elapsed: 9s, execs: 841885 (148987/sec), new interesting: 42 (total: 82)
fuzz: elapsed: 11s, execs: 1166352 (154963/sec), new interesting: 42 (total: 82)
PASS
ok  	examples/ch19/rle	11.100s
```

"Interesting" inputs are ones that reached new code paths. The fuzzer caches them (under
`go env GOCACHE`), which is why the baseline here is 40 inputs, not 3, and why yours will differ.
Without `-fuzztime`, fuzzing runs until it finds a failure or you press Ctrl-C. You can fuzz only
one target in one package at a time.

## Coverage

`-cover` reports what fraction of statements the tests ran:

```console
$ go test -cover ./ch19/...
ok  	examples/ch19/concat	0.003s	coverage: 100.0% of statements
ok  	examples/ch19/files	0.003s	coverage: 100.0% of statements
ok  	examples/ch19/first	0.002s	coverage: 100.0% of statements
ok  	examples/ch19/mapfs	0.002s	coverage: 87.5% of statements
ok  	examples/ch19/reverse	0.003s	coverage: 100.0% of statements
ok  	examples/ch19/rle	0.002s	coverage: 90.9% of statements
```

For details, write a profile with `-coverprofile=cover.out` and inspect it with `go tool cover`:
`-func=cover.out` lists each function, and `-html=cover.out` opens your browser with the source
colored green (run) and red (not run). You'll do this on `seek` below.

Coverage shows what is *not* tested. A high number doesn't mean good tests: a test with no checks
still "covers" every line it runs. Write tests that fail when the code is wrong, and use the
report to find gaps.

## A `go test` flag reference

| Flag | Meaning |
| --- | --- |
| `-v` | Verbose: list each test, show `t.Log` output. |
| `-run REGEXP` | Run only matching tests, examples, and fuzz seeds; `/` separates subtest levels. |
| `-count=N` | Run each test N times; `-count=1` also bypasses the cache. |
| `-race` | Enable the race detector ([Chapter 17](17-sync.md)). |
| `-short` | Tell tests to be quick; a test checks `testing.Short()` and skips slow parts. |
| `-failfast` | Stop after the first failing test. |
| `-timeout D` | Panic if the tests run longer than D (default 10m). |
| `-bench REGEXP`, `-benchmem` | Run benchmarks; report allocations. |
| `-fuzz REGEXP`, `-fuzztime D` | Fuzz one target, for D (or forever). |
| `-cover`, `-coverprofile=FILE` | Measure coverage. |

`go test ./...` tests every package in the module. `go test` also runs a subset of `go vet`'s
checks (ones that catch real bugs) on the packages it tests.

## Building seek: a test suite

`seek` is easy to test: most of its logic is in small, pure functions, and because
[Chapter 11](11-io-and-files.md) put the whole program in `run(args, stdin, stdout, stderr)` (with
`ctx` since Chapter 18), even `main` can be tested without starting a process.

The suite is about 800 lines in eight files. You'll type the three most instructive parts: the
matcher tests (which find a bug), the end-to-end test with golden files, and the benchmark. Copy
the rest from `project/ch19/`; it uses the same techniques and is described at the end.

### Table tests for the matchers: `internal/search/matcher_test.go`

The matchers are the heart of `seek`, and they're unexported (`literal`, `foldLiteral`, and the
regexp). This test is in `package search`, so it can reach them. Create
`internal/search/matcher_test.go`:

```go
package search

import (
	"slices"
	"strings"
	"testing"
)

// TestMatchers is a table-driven test: each row is one case, and the
// loop runs them all as subtests named after the row.
func TestMatchers(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		ignoreCase bool
		regex      bool
		line       string
		want       [][]int
	}{
		{"literal", "fox", false, false, "a fox, a fox", [][]int{{2, 5}, {9, 12}}},
		{"literal no match", "Fox", false, false, "a fox", nil},
		{"literal empty pattern", "", false, false, "abc", [][]int{{0, 0}}},
		{"literal empty line", "", false, false, "", [][]int{{0, 0}}},
		{"fold", "FOX", true, false, "a Fox", [][]int{{2, 5}}},
		{"fold non-ASCII", "CAFÉ", true, false, "café", [][]int{{0, 5}}},
		{"fold kelvin sign", "k", true, false, "\u212A", [][]int{{0, 3}}},
		{"fold to kelvin sign", "\u212A", true, false, "ok", [][]int{{1, 2}}},
		{"fold empty line", "", true, false, "", [][]int{{0, 0}}},
		{"regex", "f.x", false, true, "fax fix", [][]int{{0, 3}, {4, 7}}},
		{"regex ignore case", "fox", true, true, "FOX", [][]int{{0, 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // subtests marked Parallel run at the same time
			m, err := NewMatcher(tt.pattern, tt.ignoreCase, tt.regex)
			if err != nil {
				t.Fatalf("NewMatcher: %v", err)
			}
			got := m.FindAllStringIndex(tt.line, -1)
			// slices.EqualFunc compares element by element; the
			// elements are slices too, so compare those with slices.Equal.
			if !slices.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("FindAllStringIndex(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestNewMatcherInvalidRegex(t *testing.T) {
	_, err := NewMatcher("a(b", false, true)
	if err == nil || !strings.HasPrefix(err.Error(), "invalid pattern:") {
		t.Errorf("got error %v, want an invalid pattern error", err)
	}
}
```

The table covers each matcher, the tricky Unicode cases from [Chapter 6](06-strings.md) (`é` is
two bytes; the Kelvin sign `K` folds to `k` but is three bytes), and the edge cases: empty
patterns and empty lines. `[][]int` can't be compared with `==`, so `slices.EqualFunc` compares the
outer slices, using `slices.Equal` on each inner pair.

Run it:

```console
$ go test ./internal/search
--- FAIL: TestMatchers (0.00s)
    --- FAIL: TestMatchers/fold_empty_line (0.00s)
        matcher_test.go:44: FindAllStringIndex("") = [], want [[0 0]]
FAIL
FAIL	seek/internal/search	0.008s
FAIL
```

A real bug. An empty pattern matches every line, including empty ones. `literal` gets that
right, but with `-i`, an empty line isn't matched. The program shows it too:

```console
$ printf 'a\n\nb\n' | ./seek -c '' -
(standard input):3
$ printf 'a\n\nb\n' | ./seek -i -c '' -
(standard input):2
```

Three lines, and `-i` shouldn't change which of them match an empty pattern. Here is `indexFold`
from `internal/search/matcher.go`:

```go
func indexFold(s, substr string) (int, int) {
	// Ranging over a string visits the byte offset of each rune.
	for start := range s {
		if n := prefixFold(s[start:], substr); n >= 0 {
			return start, start + n
		}
	}
	return -1, -1
}
```

When `s` is empty, the loop never runs, so it reports "not found", even though an empty `substr`
is found at offset 0 of any string (`strings.Index("", "")` is 0). `literal` uses `strings.Index`
and gets this right. Nobody tries `seek -i ''` by hand; the table found it because every matcher
got the same edge-case rows.

Fix it by handling the empty pattern first:

```go
func indexFold(s, substr string) (int, int) {
	if substr == "" {
		return 0, 0 // like strings.Index, even when s is empty too
	}
	// Ranging over a string visits the byte offset of each rune.
	for start := range s {
```

```console
$ go test ./internal/search
ok  	seek/internal/search	0.010s
$ go build -o seek ./cmd/seek
$ printf 'a\n\nb\n' | ./seek -i -c '' -
(standard input):3
```

That's the only change to `seek`'s own code in this chapter.

### Fuzzing the matchers

`literal` and `foldLiteral` each have a perfect reference: the `strings` package. Add
`"unicode/utf8"` to the imports of `matcher_test.go`, then add two fuzz targets:

```go
// FuzzLiteral checks literal against the standard library. Run it with
//
//	go test -fuzz=FuzzLiteral ./internal/search
//
// Without -fuzz, go test runs only the seed inputs added with f.Add and
// any failing inputs saved in testdata/fuzz.
func FuzzLiteral(f *testing.F) {
	f.Add("the quick brown fox", "fox")
	f.Add("aaaa", "aa")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, s, substr string) {
		got := literal(substr).FindAllStringIndex(s, -1)
		if substr == "" {
			if !slices.EqualFunc(got, [][]int{{0, 0}}, slices.Equal) {
				t.Fatalf("empty pattern: got %v, want [[0 0]]", got)
			}
			return
		}
		if want := strings.Count(s, substr); len(got) != want {
			t.Fatalf("got %d matches, strings.Count says %d", len(got), want)
		}
		prev := 0
		for _, span := range got {
			if span[0] < prev || s[span[0]:span[1]] != substr {
				t.Fatalf("bad span %v in %v", span, got)
			}
			prev = span[1]
		}
	})
}

// FuzzIndexFold checks two properties of indexFold that must hold for
// any valid UTF-8 input: what it finds equals substr ignoring case, and
// it finds an exact match at least as early as strings.Index does.
func FuzzIndexFold(f *testing.F) {
	f.Add("The Quick Fox", "qUiCk")
	f.Add("\u212Aelvin", "kel")
	f.Fuzz(func(t *testing.T, s, substr string) {
		// indexFold works rune by rune, so it can't find half a rune
		// ("\xd7") inside a whole one ("\u05c0"), but strings.Index can.
		if !utf8.ValidString(s) || !utf8.ValidString(substr) {
			t.Skip("not valid UTF-8")
		}
		start, end := indexFold(s, substr)
		if start < 0 {
			if strings.Contains(s, substr) {
				t.Fatalf("indexFold(%q, %q) found nothing, but s contains substr", s, substr)
			}
			return
		}
		if start > strings.Index(s, substr) && strings.Contains(s, substr) {
			t.Fatalf("indexFold(%q, %q) = %d, after the exact match", s, substr, start)
		}
		if !strings.EqualFold(s[start:end], substr) {
			t.Fatalf("indexFold(%q, %q) = %q, not equal ignoring case", s, substr, s[start:end])
		}
	})
}
```

`FuzzLiteral` checks three properties of every match list: the number of matches equals
`strings.Count` (both count non-overlapping matches), the spans move forward, and each span's text
is exactly the pattern.

There's no `strings.IndexFold` to compare against, so `FuzzIndexFold` checks properties instead:
whatever it finds equals the pattern ignoring case, and if the exact pattern occurs, `indexFold`
finds a match no later.

The `t.Skip` line has a story. The first version of `FuzzIndexFold` didn't have it, and the
fuzzer failed within a second:

```console
$ go test -run '^$' -fuzz FuzzIndexFold -fuzztime 120s ./internal/search
fuzz: elapsed: 0s, gathering baseline coverage: 0/143 completed
fuzz: elapsed: 0s, gathering baseline coverage: 143/143 completed, now fuzzing with 32 workers
fuzz: minimizing 56-byte failing input file
fuzz: elapsed: 1s, minimizing
--- FAIL: FuzzIndexFold (0.80s)
    --- FAIL: FuzzIndexFold (0.00s)
        matcher_test.go:97: indexFold("™", "\xe2") found nothing, but s contains substr

    Failing input written to testdata/fuzz/FuzzIndexFold/73a524828ff4d4c3
    To re-run:
    go test -run=FuzzIndexFold/73a524828ff4d4c3
FAIL
exit status 1
FAIL	seek/internal/search	0.800s
```

`"™"` is three bytes in UTF-8: `e2 84 a2`. The pattern `"\xe2"` is its first byte alone, which
isn't valid UTF-8. `strings.Contains` works on bytes, so it finds that byte; `indexFold` works on
whole runes, so it can't. (The comment in the test names a similar pair another run found: half
of the two-byte `"\u05c0"`.)

Is that a bug? It depends on what `indexFold` promises. Case-insensitive matching only makes sense
for characters, and half a character isn't one. So the property holds for valid UTF-8 only, and
the test says so with `t.Skip` (skipped inputs aren't failures). The saved file in
`internal/search/testdata/fuzz/FuzzIndexFold/` was then deleted, since that input no longer
fails.

Fuzzing is as much about discovering what your code actually promises as about finding crashes.

### End-to-end tests with golden files: `cmd/seek/main_test.go`

The best test of a command-line program runs it as users do and checks everything it prints.
Since `run` takes its arguments, input, and outputs as parameters, a test can call it directly
with `bytes.Buffer`s for the outputs. Rather than bloat the test with expected strings, each
case's expected output lives in a **golden file**, `cmd/seek/testdata/NAME.golden`. When the output changes on
purpose, an `-update` flag rewrites the golden files, and you review the change with `git diff`.

Create `cmd/seek/main_test.go`:

```go
package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// go test ./cmd/seek -update rewrites the golden files with the current
// output. Review the changes with git diff before committing them!
var update = flag.Bool("update", false, "update golden files")

// TestRun runs the whole program, as the shell would, against the
// fixtures in the project's testdata directory. The expected output of
// each case is kept in a "golden file": cmd/seek/testdata/NAME.golden.
func TestRun(t *testing.T) {
	t.Setenv("SEEK_COLOR", "auto")
	golden, err := filepath.Abs("testdata") // before t.Chdir below
	if err != nil {
		t.Fatal(err)
	}
	// Run from the project root so the output shows paths like
	// testdata/poem.txt, just as in the book. t.Chdir changes back when
	// the test ends.
	t.Chdir("../..")

	tests := []struct {
		name     string
		args     []string
		stdin    string
		wantCode int
	}{
		{"fox", []string{"fox", "testdata"}, "", exitOK},
		{"ignore-case-numbers", []string{"-i", "-n", "fox", "testdata"}, "", exitOK},
		{"count-hidden", []string{"-c", "--hidden", "fox", "testdata"}, "", exitOK},
		{"files-with-matches", []string{"-l", "-i", "fox", "testdata"}, "", exitOK},
		{"regex-glob", []string{"-E", "--glob", "*.txt", "qu?ick|lazy", "testdata"}, "", exitOK},
		{"json", []string{"--json", "-i", "fox", "testdata/notes.md"}, "", exitOK},
		{"color", []string{"--color", "always", "-i", "fox", "testdata/poem.txt"}, "", exitOK},
		{"stdin-invert", []string{"-v", "fox", "-"}, "a fox\na dog\n", exitOK},
		{"no-match", []string{"zebra", "testdata"}, "", exitNoMatch},
		{"missing-file", []string{"fox", "testdata/poem.txt", "nope.txt"}, "", exitError},
		{"usage", []string{"-x", "fox"}, "", exitError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(t.Context(), tt.args, strings.NewReader(tt.stdin), &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code %d, want %d", code, tt.wantCode)
			}
			got := stdout.String()
			if stderr.Len() > 0 {
				got += "-- stderr --\n" + stderr.String()
			}
			checkGolden(t, filepath.Join(golden, tt.name+".golden"), got)
		})
	}
}

// checkGolden compares got with the contents of the file at path, or
// with -update, writes got to the file.
func checkGolden(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err) // run with -update to create it
	}
	if got != string(want) {
		t.Errorf("output differs from %s\n--- got ---\n%s--- want ---\n%s", path, got, want)
	}
}

// TestRunCanceled checks that a canceled search reports why and fails.
func TestRunCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr bytes.Buffer
	code := run(ctx, []string{"fox", "-"}, strings.NewReader("a fox\n"), &stdout, &stderr)
	if code != exitError || stderr.String() != "seek: context canceled\n" {
		t.Errorf("got code %d, stderr %q; want 2 and a cancellation message", code, stderr.String())
	}
}

// TestRunWorkers checks that the output doesn't depend on how many
// files are searched at once.
func TestRunWorkers(t *testing.T) {
	t.Chdir("../..")
	var want string
	for _, j := range []string{"1", "2", "8"} {
		var stdout, stderr bytes.Buffer
		run(t.Context(), []string{"-j", j, "-n", "--hidden", "o", "testdata"}, nil, &stdout, &stderr)
		if j == "1" {
			want = stdout.String()
		} else if stdout.String() != want {
			t.Errorf("-j %s output differs from -j 1:\n%s", j, stdout.String())
		}
	}
}
```

Points to notice:

- **The `-update` flag** is an ordinary package-level `flag.Bool` in the test file. `go test`
  builds a test binary and passes it any flags it doesn't know, and the `testing` package parses
  them.
- **Two `testdata` directories.** The golden files live in `cmd/seek/testdata/`, but the search
  input is the project's `testdata/`, so the output shows `testdata/poem.txt` as in every
  checkpoint of this book. `TestRun` makes the golden directory an absolute path first, then uses
  `t.Chdir("../..")` to run from the project root.
- **`SEEK_COLOR`** is set with `t.Setenv`, so a user's environment can't change the results. The
  `color` case still gets color, because `--color always` beats the variable.
- **stderr goes into the golden file too**, after a `-- stderr --` marker, so error messages are
  checked as well.
- **`TestRunCanceled`** passes an already-canceled context. `context.Cause` of a plain canceled
  context is `context.Canceled`, so the message is `seek: context canceled`.
- **`TestRunWorkers`** checks a property, not a fixed output: `-j 1`, `-j 2`, and `-j 8` must
  print exactly the same thing.

Create the golden files by running with `-update` once, then look at them:

```console
$ go test ./cmd/seek -update
ok  	seek/cmd/seek	0.006s
$ ls cmd/seek/testdata
color.golden
count-hidden.golden
files-with-matches.golden
fox.golden
ignore-case-numbers.golden
json.golden
missing-file.golden
no-match.golden
regex-glob.golden
stdin-invert.golden
usage.golden
$ cat cmd/seek/testdata/missing-file.golden
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
-- stderr --
seek: nope.txt: no such file or directory
$ cat cmd/seek/testdata/usage.golden
-- stderr --
seek: flag provided but not defined: -x
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
```

**Read every golden file the first time.** `-update` records whatever the program does, bugs
included.

From now on, plain `go test` compares against them, and a failure shows both versions. Here
`fox.golden` was edited by hand (`runs.` became `runs!`) to show one:

```console
$ go test ./cmd/seek
--- FAIL: TestRun (0.00s)
    --- FAIL: TestRun/fox (0.00s)
        main_test.go:60: output differs from /home/you/seek/cmd/seek/testdata/fox.golden
            --- got ---
            testdata/poem.txt:The quick brown fox
            testdata/poem.txt:The dog sleeps; the fox runs.
            testdata/src/hello.go:	println("hello, fox")
            --- want ---
            testdata/poem.txt:The quick brown fox
            testdata/poem.txt:The dog sleeps; the fox runs!
            testdata/src/hello.go:	println("hello, fox")
FAIL
FAIL	seek/cmd/seek	0.006s
FAIL
```

Line 60 is the `checkGolden` call in `TestRun`, not the `t.Errorf` inside `checkGolden`: that's
`t.Helper` at work. Undo the edit.

### A benchmark: `internal/search/search_test.go`

Copy `internal/search/search_test.go` from `project/ch19/`. It's an external test package,
`package search_test`, so it uses only what other packages can see:

```go
// Package search_test is an external test package: it sees package
// search only from the outside, exactly as other packages do.
package search_test
```

It has a helper (`newSearcher`, with `t.Helper()`), a table test of `Search` (plain, ignore case,
invert, max count, no match), a binary-file test, a test with a 100 KB line (longer than
`bufio.Scanner`'s default limit), and an example:

```go
// An Example is documentation that go test runs: it fails if the
// printed output differs from the Output comment. go doc shows it too.
func ExampleSearcher_Search() {
	m, _ := search.NewMatcher("fox", true, false)
	s := &search.Searcher{Matcher: m}
	matches, _ := s.Search(context.Background(), strings.NewReader(poem), "poem.txt")
	for _, m := range matches {
		fmt.Printf("%d: %s\n", m.Line, m.Text)
	}
	// Output:
	// 1: The quick brown fox
	// 3: The dog sleeps; the fox runs.
	// 4: Foxes are quick and clever.
}
```

`ExampleSearcher_Search` documents method `Search` of type `Searcher`. At the end comes the
benchmark, worth typing yourself to get a feel for `b.Run`, `b.SetBytes`, and `b.Loop`:

```go
// BenchmarkSearch measures each kind of matcher on the same input. Run
// it with: go test -bench=. -benchmem ./internal/search
func BenchmarkSearch(b *testing.B) {
	text := strings.Repeat(poem, 1000) // about 100 KB
	cases := []struct {
		name              string
		pattern           string
		ignoreCase, regex bool
	}{
		{"literal", "fox", false, false},
		{"ignore-case", "fox", true, false},
		{"regex", "f[aeiou]x", false, true},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			m, err := search.NewMatcher(c.pattern, c.ignoreCase, c.regex)
			if err != nil {
				b.Fatal(err)
			}
			s := &search.Searcher{Matcher: m}
			b.SetBytes(int64(len(text))) // report throughput in MB/s
			// b.Loop runs the body as many times as needed for a
			// stable measurement; setup above it is not timed.
			for b.Loop() {
				if _, err := s.Search(b.Context(), strings.NewReader(text), "bench"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
```

`poem` is a constant at the top of the file holding the four lines of `testdata/poem.txt`;
repeated 1000 times, it gives about 100 KB. The text and matcher are built before `b.Loop`, so
they aren't measured. `b.Context()` is the benchmark's version of `t.Context()`.

### The rest of the suite

Copy the remaining test files from `project/ch19/`. Each uses techniques you've now seen:

| File | Tests | Worth a look for |
| --- | --- | --- |
| `internal/search/walk_test.go` | `TestWalk` (default, hidden, glob, hidden root, stdin), `TestWalkMissingRoot`, `TestWalkBreak` | `makeTree(t, files...)` builds a tree in `t.TempDir()`; `collect` uses `filepath.Rel` and `filepath.ToSlash` so the expected paths work on Windows too. |
| `internal/search/parallel_test.go` | `TestSearchAllOrder`, `TestSearchAllCanceled` | Now uses `t.Context()` instead of `context.Background()`. |
| `internal/output/output_test.go` | `TestPrinters` (10 cases), `TestHighlight` | Each row holds a *function* that builds a printer writing into a `bytes.Buffer`, so text and JSON printers share one table. |
| `internal/stats/counter_test.go` | `TestCounterZeroValue`, `TestTopN`, `TestTopNInts`, `ExampleTopN` | The same generic code tested with `string` and `int` keys; a tie broken by key. |
| `cmd/seek/flags_test.go` | `TestParseArgs`, `TestParseArgsErrors`, `TestParseArgsHelp`, `TestParseArgsEnv` | `reflect.DeepEqual` compares whole `options` structs (`==` can't: they contain a slice); `errors.As` checks for `*usageError`; `t.Setenv`. |
| `cmd/seek/testdata/*.golden` | | Or generate them with `-update`, as above. |

In total: 796 lines of tests for about 1,100 lines of code, a normal ratio for Go.

### Checkpoint

```console
$ go test ./...
ok  	seek/cmd/seek	0.007s
ok  	seek/internal/output	(cached)
ok  	seek/internal/search	(cached)
ok  	seek/internal/stats	(cached)
$ go test -race ./...
ok  	seek/cmd/seek	1.018s
ok  	seek/internal/output	(cached)
ok  	seek/internal/search	(cached)
ok  	seek/internal/stats	(cached)
$ go test -cover ./...
ok  	seek/cmd/seek	0.006s	coverage: 79.5% of statements
ok  	seek/internal/output	(cached)	coverage: 91.8% of statements
ok  	seek/internal/search	(cached)	coverage: 92.5% of statements
ok  	seek/internal/stats	(cached)	coverage: 100.0% of statements
```

Every package has tests now: no more `[no test files]`. Times and `(cached)` vary.

Run one group of subtests:

```console
$ go test -v -run 'TestMatchers/fold' ./internal/search
=== RUN   TestMatchers
=== RUN   TestMatchers/fold
=== PAUSE TestMatchers/fold
=== RUN   TestMatchers/fold_non-ASCII
=== PAUSE TestMatchers/fold_non-ASCII
=== RUN   TestMatchers/fold_kelvin_sign
=== PAUSE TestMatchers/fold_kelvin_sign
=== RUN   TestMatchers/fold_to_kelvin_sign
=== PAUSE TestMatchers/fold_to_kelvin_sign
=== RUN   TestMatchers/fold_empty_line
=== PAUSE TestMatchers/fold_empty_line
=== CONT  TestMatchers/fold
=== CONT  TestMatchers/fold_empty_line
=== CONT  TestMatchers/fold_to_kelvin_sign
=== CONT  TestMatchers/fold_kelvin_sign
=== CONT  TestMatchers/fold_non-ASCII
--- PASS: TestMatchers (0.00s)
    --- PASS: TestMatchers/fold (0.00s)
    --- PASS: TestMatchers/fold_empty_line (0.00s)
    --- PASS: TestMatchers/fold_to_kelvin_sign (0.00s)
    --- PASS: TestMatchers/fold_kelvin_sign (0.00s)
    --- PASS: TestMatchers/fold_non-ASCII (0.00s)
PASS
ok  	seek/internal/search	0.002s
```

`-run` patterns are regular expressions matched *anywhere* in the name, which can surprise you:

```console
$ go test -run 'TestRun/json' -v ./cmd/seek
=== RUN   TestRun
=== RUN   TestRun/json
--- PASS: TestRun (0.00s)
    --- PASS: TestRun/json (0.00s)
=== RUN   TestRunCanceled
--- PASS: TestRunCanceled (0.00s)
=== RUN   TestRunWorkers
--- PASS: TestRunWorkers (0.00s)
PASS
ok  	seek/cmd/seek	0.003s
```

`TestRunCanceled` and `TestRunWorkers` also contain `TestRun`, and they have no subtests for the
`/json` part to filter. Anchor the pattern to make it exact:

```console
$ go test -run '^TestRun$/json' -v ./cmd/seek
=== RUN   TestRun
=== RUN   TestRun/json
--- PASS: TestRun (0.00s)
    --- PASS: TestRun/json (0.00s)
PASS
ok  	seek/cmd/seek	0.002s
```

Examples run like tests:

```console
$ go test -run ExampleTopN -v ./internal/stats
=== RUN   ExampleTopN
--- PASS: ExampleTopN (0.00s)
PASS
ok  	seek/internal/stats	0.002s
```

The benchmark (your numbers will differ):

```console
$ go test -count=1 -run '^$' -bench . -benchmem ./internal/search
goos: linux
goarch: amd64
pkg: seek/internal/search
cpu: 13th Gen Intel(R) Core(TM) i9-13950HX
BenchmarkSearch/literal-32         	    3790	    296064 ns/op	 347.90 MB/s	  511327 B/op	    8018 allocs/op
BenchmarkSearch/ignore-case-32     	    1302	    934726 ns/op	 110.19 MB/s	  674214 B/op	   10019 allocs/op
BenchmarkSearch/regex-32           	    2094	    573261 ns/op	 179.67 MB/s	  958477 B/op	    8021 allocs/op
PASS
ok  	seek/internal/search	3.546s
```

Ignore-case is about three times slower than literal, and slower even than the regexp:
`indexFold` tries `prefixFold` at every rune offset of every line, and `prefixFold` calls
`strings.EqualFold` rune by rune, while `strings.Index` uses fast, specialized byte search.
Allocations add up too: `sc.Text()` makes a string for each of the 4000 input lines, and every
matching line costs two small slices for its span (2000 lines match `fox`, 3000 ignoring case).
If `-i` ever became a bottleneck, this benchmark is how you'd measure a fix.

Fuzz for a few seconds:

```console
$ go test -run '^$' -fuzz FuzzLiteral -fuzztime 5s ./internal/search
fuzz: elapsed: 0s, gathering baseline coverage: 0/21 completed
fuzz: elapsed: 0s, gathering baseline coverage: 21/21 completed, now fuzzing with 32 workers
fuzz: elapsed: 3s, execs: 2031619 (677202/sec), new interesting: 0 (total: 21)
fuzz: elapsed: 5s, execs: 3410584 (655335/sec), new interesting: 0 (total: 21)
PASS
ok  	seek/internal/search	5.110s
```

Then find what isn't covered:

```console
$ go test -coverprofile=cover.out ./internal/search && go tool cover -func=cover.out | tail -4 && rm cover.out
ok  	seek/internal/search	0.010s	coverage: 92.5% of statements
seek/internal/search/search.go:86:	Search			90.0%
seek/internal/search/walk.go:27:	Walk			100.0%
seek/internal/search/walk.go:39:	walkRoot		100.0%
total:					(statements)		92.5%
```

(Leave out `| tail -4` to see every function, or use `go tool cover -html=cover.out` to see the
uncovered lines in red. Exercise 3 closes one of the gaps in `Search`.)

The complete project at this point is in `project/ch19/`.

## Summary

- Tests are `TestXxx(t *testing.T)` functions in `_test.go` files; `go test ./...` runs them and
  caches results (`-count=1` forces a rerun).
- `t.Error`/`t.Errorf` report and continue; `t.Fatal`/`t.Fatalf` report and stop. Messages read
  `F(input) = got, want expected`.
- Table-driven tests loop over a slice of structs; `t.Run` makes each case a named subtest that
  `-run 'Test/name'` can select; `t.Parallel` runs them at the same time.
- `package foo_test` tests from the outside. `ExampleXxx` functions with `// Output:` are tested
  documentation.
- `t.Helper`, `t.Cleanup`, `t.TempDir`, `t.Context`, `t.Setenv`, and `t.Chdir` handle helpers and
  setup/teardown; `testdata` holds fixtures.
- Golden files store expected output; an `-update` flag regenerates them. Review every change.
- `testing/fstest.MapFS` is an in-memory `fs.FS` for code that takes an `fs.FS`.
- Benchmarks are `BenchmarkXxx(b *testing.B)` with `for b.Loop()`; run with `-bench . -benchmem`.
- Fuzz tests are `FuzzXxx(f *testing.F)` checking a property; `-fuzz` saves failures in
  `testdata/fuzz` as permanent regression tests. Then fix the code, or narrow (and document) the
  promise.
- `-cover` and `-coverprofile` with `go tool cover` show what the tests miss.
- `seek`'s 796-line test suite found a bug: `-i ''` didn't match empty lines.

## Exercises

1. **Edge cases for `Abs`.** Rewrite `examples/ch19/first`'s `TestAbs` as a table test with
   subtests, and add rows for `math.MaxInt` and `math.MinInt`. What does `Abs(math.MinInt)`
   return, and why? Make the test document the behavior.

<details><summary>Solution</summary>

```go
package first

import (
	"math"
	"testing"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want int
	}{
		{"negative", -3, 3},
		{"positive", 5, 5},
		{"zero", 0, 0},
		{"max", math.MaxInt, math.MaxInt},
		{"min overflows", math.MinInt, math.MinInt}, // -MinInt doesn't fit in an int
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Abs(tt.x); got != tt.want {
				t.Errorf("Abs(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}
```

If you first write the last row as `{"min", math.MinInt, math.MaxInt}`, expecting `math.MaxInt`,
the test fails:

```console
$ go test ./ch19/first
--- FAIL: TestAbs (0.00s)
    --- FAIL: TestAbs/min (0.00s)
        abs_test.go:23: Abs(-9223372036854775808) = -9223372036854775808, want 9223372036854775807
FAIL
FAIL	examples/ch19/first	0.002s
FAIL
```

An `int64` holds -9223372036854775808 to 9223372036854775807, so the smallest value has no
negation. Integer overflow wraps around ([Chapter 2](02-values-and-types.md)), and `-x` gives `x`
back. (`math.Abs` for floats doesn't have this problem.) The row records the behavior, so anyone
who changes it will notice.

</details>

2. **A new golden case.** Add a `TestRun` case named `max-count` for `seek -m 1 -n fox testdata`.
   What happens if you run `go test ./cmd/seek` before creating the golden file? Create it, and
   check that its content is right.

<details><summary>Solution</summary>

Add a row to the table in `TestRun`:

```go
		{"max-count", []string{"-m", "1", "-n", "fox", "testdata"}, "", exitOK},
```

Without the golden file, `checkGolden`'s `os.ReadFile` fails:

```console
$ go test ./cmd/seek
--- FAIL: TestRun (0.00s)
    --- FAIL: TestRun/max-count (0.00s)
        main_test.go:61: open /home/you/seek/cmd/seek/testdata/max-count.golden: no such file or directory
FAIL
FAIL	seek/cmd/seek	0.007s
FAIL
$ go test ./cmd/seek -update
ok  	seek/cmd/seek	0.006s
$ cat cmd/seek/testdata/max-count.golden
testdata/poem.txt:1:The quick brown fox
testdata/src/hello.go:4:	println("hello, fox")
$ go test ./cmd/seek
ok  	seek/cmd/seek	0.020s
```

It's right: `-m 1` stops after the first match in each file, so `poem.txt` shows only line 1, not
line 3. (The failure is at line 61, not 60, because the new row moved the call down a line.)

</details>

3. **Close a coverage gap.** The coverage profile shows `Search` at 90%. One statement never run
   is the `return nil, err` after `ctx.Err()` in the scan loop. Write a test in
   `internal/search/search_test.go` that reaches it, and check the new coverage.

<details><summary>Solution</summary>

`search_test.go` already imports `context` and `errors` and has `newSearcher` and `poem`:

```go
func TestSearchCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	s := newSearcher(t, "fox", false)
	matches, err := s.Search(ctx, strings.NewReader(poem), "poem.txt")
	if !errors.Is(err, context.Canceled) || matches != nil {
		t.Errorf("got %v, %v; want no matches and context.Canceled", matches, err)
	}
}
```

```console
$ go test -cover ./internal/search
ok  	seek/internal/search	0.009s	coverage: 93.2% of statements
$ go test -coverprofile=cover.out ./internal/search >/dev/null && go tool cover -func=cover.out | grep 'Search\b\|total'; rm cover.out
seek/internal/search/search.go:86:	Search			95.0%
total:					(statements)		93.2%
```

`TestSearchAllCanceled` never reached that line: with the context canceled up front, no worker
got as far as scanning a file.

</details>

4. **Skip `vendor`.** Change `examples/ch19/mapfs`'s `FindExt` to skip directories named
   `vendor` entirely, and update the test. Notice how little the test changes.

<details><summary>Solution</summary>

In the `fs.WalkDir` callback, before the extension check:

```go
		if d.IsDir() && d.Name() == "vendor" {
			return fs.SkipDir
		}
```

Only the test's expected list changes; the `MapFS` already contains `vendor/x/x.go`:

```go
	want := []string{"internal/a/a.go", "internal/b/doc.go", "main.go"}
```

```console
$ go test ./ch19/mapfs
ok  	examples/ch19/mapfs	0.002s
```

Update the doc comment too: "It skips directories named vendor."

</details>

5. **Fuzz `reverse.String`.** Write `FuzzString` in `examples/ch19/reverse` with two properties:
   reversing twice gives back the input, and reversing doesn't change the number of runes. Run it.
   What does it find, and what should you do about it?

<details><summary>Solution</summary>

`examples/ch19/reverse/fuzz_test.go`:

```go
package reverse

import (
	"testing"
	"unicode/utf8"
)

func FuzzString(f *testing.F) {
	f.Add("hello")
	f.Add("héllo, 世界")
	f.Fuzz(func(t *testing.T, s string) {
		if got := String(String(s)); got != s {
			t.Errorf("String(String(%q)) = %q", s, got)
		}
		if utf8.RuneCountInString(String(s)) != utf8.RuneCountInString(s) {
			t.Errorf("String(%q) changed the rune count", s)
		}
	})
}
```

```console
$ go test -run '^$' -fuzz FuzzString -fuzztime 20s ./ch19/reverse
...
--- FAIL: FuzzString (0.09s)
    --- FAIL: FuzzString (0.00s)
        fuzz_test.go:13: String(String("\xb8")) = "�"

    Failing input written to testdata/fuzz/FuzzString/db7b6a8cd735c025
    To re-run:
    go test -run=FuzzString/db7b6a8cd735c025
FAIL
exit status 1
FAIL	examples/ch19/reverse	0.095s
```

`[]rune(s)` turns each invalid byte into U+FFFD, the replacement character `�`, so invalid UTF-8
can't survive a round trip. As with `indexFold`, decide what you promise: either document "s must
be valid UTF-8" and `t.Skip` invalid inputs in the fuzz target, or reverse invalid bytes as bytes
(with `utf8.DecodeRuneInString`, which reports a size of 1 for them). The first is simpler. If you
choose to skip, delete `testdata/fuzz/FuzzString`, since that input no longer fails.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-19--testing-benchmarks-and-fuzzing)
