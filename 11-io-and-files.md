# Chapter 11 — Readers, Writers, and the File System

Programs move bytes between files, pipes, the network, and the terminal. Go handles all of them with
two tiny interfaces, `io.Reader` and `io.Writer`, plus packages built on top: `bufio` (buffering, line
scanning), `os` (files), and `path/filepath` and `io/fs` (paths, directory trees). Then `seek` learns
to stream files line by line, read standard input, walk directories, skip binary files, and filter by
`--glob`.

Previous: [Chapter 10 — Interfaces](10-interfaces.md). Next: [Chapter 12 — Flags](12-flags.md).

## io.Reader: one method for every source of bytes

The whole `io.Reader` interface, from [Chapter 10](10-interfaces.md):

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

`Read` fills up to `len(p)` bytes of `p` and returns the count. At the end of the data it returns
`io.EOF`.

`examples/ch11/reader/main.go`

```go
package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	r := strings.NewReader("Hello, Reader!")
	buf := make([]byte, 8)
	for {
		// Read fills up to len(buf) bytes and reports how many it got.
		n, err := r.Read(buf)
		fmt.Printf("n=%d err=%v data=%q\n", n, err, buf[:n])
		if err == io.EOF {
			break
		}
	}
}
```

```console
$ go run ./ch11/reader
n=8 err=<nil> data="Hello, R"
n=6 err=<nil> data="eader!"
n=0 err=EOF data=""
```

Rules for callers:

- Use only `buf[:n]`; the bytes after `n` are leftovers.
- `Read` may return `n > 0` *and* an error (even `io.EOF`) in one call. Handle the bytes first, then
  the error.
- `io.EOF` is not a failure. It's a sentinel meaning "no more data". Compare with `==` or `errors.Is`.

Usually `io.Copy`, `io.ReadAll`, or `bufio.Scanner` runs this loop for you, correctly.

**Why so small?** One method is easy to implement, so *everything* does: `*os.File`, `strings.Reader`,
`bytes.Buffer`, network connections, HTTP bodies, gzip decompressors. Code using `io.Reader` works with
all of them, even future ones. `io.Writer` mirrors it: `Write(p []byte) (n int, err error)`.

## Composing readers and writers

Tiny interfaces nest like shell pipes. The `io` package ships these wrappers:

| Function | What it does |
|----------|--------------|
| `io.Copy(dst, src)` | reads `src` until EOF, writes everything to `dst` |
| `io.ReadAll(r)` | reads `r` until EOF, returns all bytes |
| `io.LimitReader(r, n)` | a reader that stops after `n` bytes |
| `io.MultiReader(r1, r2, ...)` | reads `r1`, then `r2`, ... |
| `io.TeeReader(r, w)` | a reader that also writes what it reads to `w` |
| `io.MultiWriter(w1, w2, ...)` | a writer that writes to all of them |
| `io.Discard` | a writer that throws everything away |

Your own wrapper is just a type with a `Read` method, like `upper`:

`examples/ch11/compose/main.go`

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// upper is an io.Reader that turns ASCII letters from r into upper case.
type upper struct {
	r io.Reader
}

func (u upper) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	for i := range p[:n] {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err
}

func main() {
	// A pipeline of readers: source → first 15 bytes → upper case → stdout.
	src := strings.NewReader("hello, gophers! this part is cut off")
	n, err := io.Copy(os.Stdout, upper{io.LimitReader(src, 15)})
	fmt.Printf("\ncopied %d bytes, err=%v\n", n, err)

	// MultiReader reads its readers one after another.
	r := io.MultiReader(strings.NewReader("one "), strings.NewReader("two"))
	// TeeReader writes everything it reads to a second place.
	var seen strings.Builder
	data, err := io.ReadAll(io.TeeReader(r, &seen))
	fmt.Printf("ReadAll: %q err=%v\n", data, err)
	fmt.Printf("tee saw: %q\n", seen.String())

	// MultiWriter writes the same bytes to several writers.
	var a, b bytes.Buffer
	w := io.MultiWriter(&a, &b, os.Stdout)
	fmt.Fprintln(w, "to three places")
	fmt.Printf("a=%q b=%q\n", a.String(), b.String())
}
```

```console
$ go run ./ch11/compose
HELLO, GOPHERS!
copied 15 bytes, err=<nil>
ReadAll: "one two" err=<nil>
tee saw: "one two"
to three places
a="to three places\n" b="to three places\n"
```

Notes:

- At EOF, `io.Copy` and `io.ReadAll` return a `nil` error: EOF is the normal finish.
- `&seen`, `&a`, and `&b` are pointers because `strings.Builder` and `bytes.Buffer` have
  pointer-receiver `Write` methods ([Chapter 7](07-structs-methods-pointers.md) covers method sets).
- `fmt.Fprintln` takes any `io.Writer`: terminal, file, buffer, or all three.
- `io.ReadAll` holds everything in memory. Stream big inputs with `io.Copy` or `bufio.Scanner`.

## Files: os.Open, os.Create, and Close

An open file is an `*os.File`: both an `io.Reader` and an `io.Writer`.

| Call | Opens for | If the file exists | If it doesn't |
|------|-----------|--------------------|---------------|
| `os.Open(name)` | reading | opens it | error |
| `os.Create(name)` | reading and writing | empties it | creates it |
| `os.OpenFile(name, flags, perm)` | you choose: `os.O_RDONLY`, `O_WRONLY`, `O_RDWR`, `O_APPEND`, `O_CREATE`, `O_TRUNC`, `O_EXCL` | depends on the flags | depends on the flags |
| `os.ReadFile(name)` | one call: open, read all, close | | error |
| `os.WriteFile(name, data, perm)` | one call: create/truncate, write, close | | creates it |

Close every file you open; the OS limits how many file descriptors you can hold. Write
`defer f.Close()` right after the error check, so Close runs however the function returns.

Closing a file you only **read** can't lose data, so you may ignore its error. A file you **write** may
report lost data (full disk, network drive) only at Close, so check it. `writeNotes` does this with a
named result and a deferred closure ([Chapter 4](04-functions.md)).

`examples/ch11/files/main.go`

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// A fresh directory under the system's temp dir, removed at the end.
	dir, err := os.MkdirTemp("", "files-example")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "notes.txt")

	if err := writeNotes(path); err != nil {
		return err
	}

	// OpenFile takes flags and permissions. Here: append, write only.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if _, err := f.WriteString("appended line\n"); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Open is read only. We only read, so a deferred Close is enough.
	f, err = os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	fmt.Printf("%s", data)

	// One-call shortcuts for small files.
	other := filepath.Join(dir, "other.txt")
	if err := os.WriteFile(other, []byte("short and sweet\n"), 0o644); err != nil {
		return err
	}
	data, err = os.ReadFile(other)
	fmt.Printf("%q %v\n", data, err)

	_, err = os.Open(filepath.Join(dir, "missing.txt"))
	fmt.Println("missing file:", errors.Is(err, fs.ErrNotExist))
	return nil
}

// writeNotes creates path and writes two lines to it. The named result
// lets the deferred function report an error from Close.
func writeNotes(path string) (err error) {
	f, err := os.Create(path) // create, or truncate if it exists
	if err != nil {
		return err
	}
	defer func() {
		// Close flushes data to the OS; a failure here means lost data.
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	if _, err := fmt.Fprintln(f, "first line"); err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d + %d = %d\n", 2, 2, 4)
	return err
}
```

```console
$ go run ./ch11/files
first line
2 + 2 = 4
appended line
"short and sweet\n" <nil>
missing file: true
```

Details:

- `0o644` is octal for `rw-r--r--` (owner writes, everyone reads). Your *umask* then usually clears
  the group/other write bits. `os.Create` uses `0o666` before the umask.
- `os.MkdirTemp("", pattern)` makes a unique directory under `os.TempDir()`; `os.RemoveAll` deletes a
  tree. Also: `os.Mkdir`, `os.MkdirAll` (like `mkdir -p`), `os.Remove`, `os.Rename`.
- `os` errors are `*fs.PathError`s like `open missing.txt: no such file or directory`. Test the cause
  with `errors.Is(err, fs.ErrNotExist)` (or `fs.ErrPermission`, `fs.ErrExist`), as in
  [Chapter 9](09-errors.md).

## bufio.Scanner: reading line by line

`os.ReadFile` loads the whole file: a 2 GB log needs 2 GB of RAM. `bufio.Scanner` reads a bit at a
time and returns one *token* (a line, by default) per `Scan` call.

`examples/ch11/scanner/main.go`

```go
package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	// Lines: the newline (and a "\r" before it) is removed.
	input := "alpha\nbeta gamma\r\n\ndelta" // no newline at the end
	sc := bufio.NewScanner(strings.NewReader(input))
	for n := 1; sc.Scan(); n++ {
		fmt.Printf("%d: %q\n", n, sc.Text())
	}

	// Words: a different split function.
	sc = bufio.NewScanner(strings.NewReader("  the quick\tbrown\n fox "))
	sc.Split(bufio.ScanWords)
	var words []string
	for sc.Scan() {
		words = append(words, sc.Text())
	}
	fmt.Printf("%q\n", words)

	// A line longer than the default limit (64 KiB) stops the scanner.
	long := strings.Repeat("x", 100_000) + "\nshort\n"
	sc = bufio.NewScanner(strings.NewReader(long))
	for sc.Scan() {
		fmt.Println("line length", len(sc.Text()))
	}
	fmt.Println("error:", sc.Err())

	// Buffer raises the limit: start with 64 KiB, grow up to 1 MiB.
	sc = bufio.NewScanner(strings.NewReader(long))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		fmt.Println("line length", len(sc.Text()))
	}
	fmt.Println("error:", sc.Err())
}
```

```console
$ go run ./ch11/scanner
1: "alpha"
2: "beta gamma"
3: ""
4: "delta"
["the" "quick" "brown" "fox"]
error: bufio.Scanner: token too long
line length 100000
line length 5
error: <nil>
```

How to use a Scanner:

1. `sc := bufio.NewScanner(r)` for any `io.Reader`.
2. `for sc.Scan() { ... sc.Text() ... }`. `Scan` returns false at the end of input *or* on an error.
3. After the loop, **always** check `sc.Err()` (`nil` at a normal EOF). Otherwise an error just ends
   the loop early, silently: the classic Scanner bug.

Gotchas:

- **Line limit.** A token can't exceed the buffer: 64 KiB by default (`bufio.MaxScanTokenSize`).
  Longer lines (common in minified JavaScript and generated files) stop the scan with
  `bufio.ErrTooLong`. Raise the limit with `sc.Buffer` before the first `Scan`.
- **`Bytes` vs `Text`.** `sc.Bytes()` points into the internal buffer; the next `Scan` overwrites it.
  `sc.Text()` is a new string, safe to keep.
- **Split functions.** `bufio.ScanLines` (default), `ScanWords`, `ScanRunes`, `ScanBytes`, or your own
  `bufio.SplitFunc`.
- The last line needs no trailing newline; `\r\n` endings lose the `\r`.

## bufio.Writer: fewer, bigger writes

Each `Write` to an `*os.File` is a slow system call. A `bufio.Writer` collects writes in memory
(4 KiB by default) and passes them on in big chunks. You must `Flush` at the end, or the last chunk is
lost.

`examples/ch11/bufwriter/main.go`

```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush() // write whatever is left when main returns

	fmt.Fprintln(w, "1. written to the buffer first")
	fmt.Println("2. printed directly; buffered bytes:", w.Buffered())

	lost := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(lost, "3. never flushed, so never printed")
}
```

```console
$ go run ./ch11/bufwriter
2. printed directly; buffered bytes: 31
1. written to the buffer first
```

Line 1 was written first but printed last: it waited for the deferred `Flush`. Line 3 never appears.

- `defer w.Flush()` works, but defers don't run after `os.Exit` ([Chapter 4](04-functions.md)).
- Errors are *sticky*: after one failed write, every later `Write` and `Flush` returns that error. So
  checking `Flush`'s error catches any write failure.
- `bufio.NewReader` is the reading side. It adds `ReadString`, `ReadRune`, and `Peek` (used below) to
  any reader.

## Standard input, output, and error

`os.Stdin`, `os.Stdout`, and `os.Stderr` are ordinary, **unbuffered** `*os.File`s.

Results go to stdout; messages and errors go to stderr. Then `prog > out.txt` saves the results and
still shows the errors. The file name `-` conventionally means standard input.

`examples/ch11/cat/main.go`

```go
// cat prints files, or standard input for the name "-".
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	status := 0
	for _, name := range os.Args[1:] {
		if err := cat(name); err != nil {
			fmt.Fprintln(os.Stderr, "cat:", err) // errors go to stderr
			status = 1
		}
	}
	os.Exit(status)
}

func cat(name string) error {
	if name == "-" {
		_, err := io.Copy(os.Stdout, os.Stdin)
		return err
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(os.Stdout, f)
	return err
}
```

```console
$ printf 'from stdin\n' | go run ./ch11/cat go.mod - nope.txt
module examples

go 1.26
from stdin
cat: open nope.txt: no such file or directory
exit status 1
```

`io.Copy` moves data in chunks (32 KiB by default), so `cat` never holds a whole file. To end typed
input in a terminal, press Ctrl-D (Windows: Ctrl-Z, then Enter) to send EOF.

## Paths: path/filepath

`path/filepath` uses the OS separator (`/` on Unix, `\` on Windows). Never build paths with
`+ "/" +`.

`examples/ch11/paths/main.go`

```go
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	p := filepath.Join("docs", "guide", "..", "intro.md") // Join also cleans
	fmt.Println(p)
	fmt.Println(filepath.Base(p), filepath.Dir(p), filepath.Ext(p))
	dir, file := filepath.Split(p)
	fmt.Printf("%q %q\n", dir, file)
	fmt.Println(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
	fmt.Println(filepath.Clean("a//b/./c/.."))
	fmt.Println(filepath.IsAbs("/etc/hosts"), filepath.IsAbs("etc/hosts"))
	rel, err := filepath.Rel("/home/gopher", "/home/gopher/src/app")
	fmt.Println(rel, err)

	// Match tests a name against a shell pattern: * ? [abc] [a-z]
	for _, name := range []string{"main.go", "main_test.go", "README.md", "src/x.go"} {
		ok, _ := filepath.Match("*.go", name)
		fmt.Printf("%-14s %v\n", name, ok)
	}
	_, err = filepath.Match("[", "x")
	fmt.Println(err)
}
```

```console
$ go run ./ch11/paths
docs/intro.md
intro.md docs .md
"docs/" "intro.md"
intro
a/b
true false
src/app <nil>
main.go        true
main_test.go   true
README.md      false
src/x.go       false
syntax error in pattern
```

(On Windows the output uses `\`.)

- `filepath.Match` matches the **whole** name, and `*` never crosses a separator: `*.go` doesn't match
  `src/x.go`. For "any `.go` file", match the base name (`filepath.Base`).
- A malformed pattern returns `filepath.ErrBadPattern` (`syntax error in pattern`).
- The `path` package does the same for `/`-separated paths that aren't OS paths: URLs and paths inside
  an `fs.FS` (below).

## Walking a directory tree: filepath.WalkDir

`filepath.WalkDir(root, fn)` calls `fn` for `root` and every file and directory under it:

```go
func(path string, d fs.DirEntry, err error) error
```

- `path`: `root` joined with the entry's relative path.
- `d`: an `fs.DirEntry` with `Name()` (base name), `IsDir()`, `Type()` (the mode's type bits), and
  `Info()` (full details; may cost a system call).
- `err`: set if something couldn't be read; `d` may then be nil.

`fn`'s result steers the walk:

| Return | Effect |
|--------|--------|
| `nil` | keep going |
| `filepath.SkipDir` (same value as `fs.SkipDir`) | on a directory: don't enter it. On a file: skip the rest of its directory |
| `filepath.SkipAll` | stop the whole walk, no error |
| any other error | stop, and `WalkDir` returns that error |

`examples/ch11/walkdir/main.go`

```go
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := makeTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // stop the walk on the first error
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			fmt.Println("skip ", rel)
			return filepath.SkipDir // don't go inside
		}
		fmt.Printf("%-5s %s\n", kind(d), rel)
		return nil
	})
	fmt.Println("err:", err)
}

func kind(d fs.DirEntry) string {
	if d.IsDir() {
		return "dir"
	}
	return "file"
}

// makeTree creates a few files under a new temporary directory.
func makeTree() (string, error) {
	root, err := os.MkdirTemp("", "walk")
	if err != nil {
		return "", err
	}
	files := map[string]string{
		"notes.md":        "# notes\n",
		"src/main.go":     "package main\n",
		"src/util/str.go": "package util\n",
		".git/config":     "[core]\n",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", err
		}
	}
	return root, nil
}
```

```console
$ go run ./ch11/walkdir
dir   .
skip  .git
file  notes.md
dir   src
file  src/main.go
dir   src/util
file  src/util/str.go
err: <nil>
```

Things to know:

- Each directory is read in **lexical** order (byte by byte: `.` before letters, upper case before
  lower), so output is deterministic.
- The root comes first.
- Symbolic links aren't followed; a link's `d.Type()` has `fs.ModeSymlink` set.
- Prefer `WalkDir` to the older `filepath.Walk`, which calls `os.Lstat` on every entry. `WalkDir` uses
  the cheaper `DirEntry`.

## io/fs: file systems as an interface

`io/fs` defines `fs.FS`, a read-only file tree:

```go
type FS interface {
	Open(name string) (fs.File, error)
}
```

Paths in an `fs.FS` use `/` and are relative (no leading `/`, no `..`); the root is `"."`.
Implementations: `os.DirFS(dir)` (a real directory), `embed.FS` (files compiled into the binary,
[Chapter 20](20-tooling-and-design.md)), `*zip.Reader`, and `fstest.MapFS` (an in-memory test fake,
[Chapter 19](19-testing.md)). Code taking an `fs.FS` works with all of them, as do the helpers
`fs.WalkDir`, `fs.ReadFile`, `fs.ReadDir`, `fs.Glob`, `fs.Stat`, and `fs.Sub`.

`examples/ch11/dirfs/main.go`

```go
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	root, err := makeTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)

	// An fs.FS is a read-only file tree. Its paths always use "/" and
	// are relative to the tree's root.
	fsys := os.DirFS(root)
	printTree(fsys)

	data, err := fs.ReadFile(fsys, "src/main.go")
	fmt.Printf("%q %v\n", data, err)
	matches, err := fs.Glob(fsys, "src/*.go")
	fmt.Println(matches, err)
}

// printTree works with any fs.FS: a directory, a zip file, files
// embedded in the binary, or a fake tree in a test.
func printTree(fsys fs.FS) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Printf("%-12s dir\n", path)
			return nil
		}
		info, err := d.Info() // the full FileInfo (may cost a system call)
		if err != nil {
			return err
		}
		fmt.Printf("%-12s %v %d bytes\n", path, info.Mode(), info.Size())
		return nil
	})
}

func makeTree() (string, error) {
	root, err := os.MkdirTemp("", "dirfs")
	if err != nil {
		return "", err
	}
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("# notes\n"), 0o644); err != nil {
		return "", err
	}
	err = os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644)
	return root, err
}
```

```console
$ go run ./ch11/dirfs
.            dir
notes.md     -rw-r--r-- 8 bytes
src          dir
src/main.go  -rw-r--r-- 13 bytes
"package main\n" <nil>
[src/main.go] <nil>
```

(The permission bits depend on your umask.)

Which walk to use:

- `filepath.WalkDir` gives paths as the user typed them (`testdata/src/hello.go`), right for CLI
  output. `seek` uses it.
- `fs.WalkDir` on an `fs.FS` is easier to test and handles embedded or zipped files. Library code that
  only reads files should consider taking an `fs.FS`.

## os.Stat and file modes

`os.Stat(name)` returns an `fs.FileInfo`: `Name()`, `Size()`, `Mode()`, `ModTime()`, `IsDir()`. Its
`fs.FileMode` has permission bits (the low 9, `Perm()`) and type bits like `fs.ModeDir`,
`fs.ModeSymlink`, `fs.ModeNamedPipe`, and `fs.ModeCharDevice`. `os.Lstat` doesn't follow a symbolic
link.

A **terminal** is a *character device*; a pipe or regular file isn't. Tools check this to decide
whether to print colors.

`examples/ch11/stat/main.go`

```go
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	info, err := os.Stat("go.mod")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(info.Name(), info.Size(), info.IsDir())
	fmt.Println(info.Mode(), info.Mode().IsRegular())
	fmt.Printf("permissions: %v (octal %o)\n", info.Mode().Perm(), info.Mode().Perm())

	info, _ = os.Stat(".")
	fmt.Println(info.Mode(), info.IsDir(), info.Mode()&fs.ModeDir != 0)

	_, err = os.Stat("nope.txt")
	fmt.Println(err)
	fmt.Println(errors.Is(err, fs.ErrNotExist))

	// Is standard output a terminal? Terminals are character devices.
	info, err = os.Stdout.Stat()
	if err == nil {
		fmt.Println("stdout mode:", info.Mode())
		fmt.Println("terminal:", info.Mode()&os.ModeCharDevice != 0)
	}
}
```

Run it piped into `cat` (stdout is a pipe), then straight in a terminal:

```console
$ go run ./ch11/stat | cat
go.mod 25 false
-rw-r--r-- true
permissions: -rw-r--r-- (octal 644)
drwxr-xr-x true true
stat nope.txt: no such file or directory
true
stdout mode: prw-------
terminal: false
$ go run ./ch11/stat
go.mod 25 false
-rw-r--r-- true
permissions: -rw-r--r-- (octal 644)
drwxr-xr-x true true
stat nope.txt: no such file or directory
true
stdout mode: Dcrw--w----
terminal: true
```

Permission bits depend on your system and umask. Mode letters: `d` directory, `p` named pipe, `D`
device, `c` character device.

## Detecting binary files

Searching an image or executable prints garbage. Tools like `git` and `grep` guess: a NUL byte (`0x00`) means binary
(`git` checks the first 8000 bytes), since text almost never has one.

Check the start **without consuming it**. `bufio.Reader.Peek(n)` returns the next `n` bytes but leaves
them buffered for later reads. The buffer must hold `n` bytes; size it with `bufio.NewReaderSize`.

`examples/ch11/binary/main.go`

```go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// looksBinary reports whether the start of br contains a NUL byte.
// Peek looks ahead without consuming, so br can still be read in full.
func looksBinary(br *bufio.Reader) (bool, error) {
	head, err := br.Peek(512)
	if err != nil && err != io.EOF {
		return false, err // io.EOF only means the input is shorter than 512
	}
	return bytes.IndexByte(head, 0) >= 0, nil
}

func main() {
	inputs := []string{
		"plain text\nline two\n",
		"GIF89a\x00\x01\x02",
		"",
	}
	for _, in := range inputs {
		br := bufio.NewReader(strings.NewReader(in))
		binary, err := looksBinary(br)
		rest, _ := io.ReadAll(br)
		fmt.Printf("binary=%-5v err=%v  bytes left to read: %d\n", binary, err, len(rest))
	}
}
```

```console
$ go run ./ch11/binary
binary=false err=<nil>  bytes left to read: 20
binary=true  err=<nil>  bytes left to read: 9
binary=false err=<nil>  bytes left to read: 0
```

On input shorter than `n`, `Peek` returns what there is *plus* `io.EOF`. That's fine here, so
`looksBinary` ignores it.

## Building seek: streaming, directories, and globs

So far `seek` read whole files with `os.ReadFile` and needed every path spelled out. Now:

- **Streaming**: a `bufio.Scanner` (lines up to 1 MiB) keeps memory flat. `-l` and `-m` stop reading
  early.
- **Binary files** (NUL in the first 8 KiB) are skipped silently.
- **Directories** are searched recursively, skipping hidden (dot-prefixed) entries unless `--hidden`.
- **`--glob GLOB`** searches only files whose base name matches.
- **`-`** is standard input; **no PATH** means `.`.
- **Output** goes through a `bufio.Writer`.

### New test fixtures

From the project root, add these to `testdata/`:

```sh
mkdir -p testdata/src testdata/.cache
printf 'package main\n\nfunc main() {\n\tprintln("hello, fox")\n}\n' > testdata/src/hello.go
printf 'a hidden fox\n' > testdata/.hidden.txt
printf 'a cached fox\n' > testdata/.cache/fox.txt
printf 'GIF89a\000fox\000\001' > testdata/image.bin
```

(Or use an editor; end each text file with a newline. On Windows, run `printf` in Git Bash or WSL.)
`image.bin` contains "fox" plus NUL bytes, so `seek` must skip it. `go build` ignores `testdata`
directories, so `hello.go` isn't compiled into `seek`.

### internal/search/search.go

`Searcher` gets a `Stdin` field. `SearchFile` opens the file and calls the new `Search`, which takes
any `io.Reader`: a file, stdin, or (in tests) a string. Replace the whole file:

```go
// Package search finds lines of text that match a pattern.
package search

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
)

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
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
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.Matcher.FindAllStringIndex(line, 1) != nil
	return found != s.Invert // != on bools is "exclusive or"
}

// SearchFile searches the file at path. The path "-" means s.Stdin.
func (s *Searcher) SearchFile(path string) ([]Match, error) {
	if path == "-" {
		return s.Search(s.Stdin, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // we only read, so Close can't lose data
	return s.Search(f, path)
}

// Search reads r line by line and returns the selected lines. path is
// recorded in each Match. It returns ErrBinary, and no matches, if the
// first few KiB of r contain a NUL byte.
func (s *Searcher) Search(r io.Reader, path string) ([]Match, error) {
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
```

Points to notice:

- The scanner reads from `br` (which we peeked through), not `r`. The peeked bytes are still in
  `br`'s buffer, so the scanner starts at the first byte.
- `ErrBinary` is a sentinel error ([Chapter 9](09-errors.md)). The caller decides what to do; `seek`
  skips the file silently.
- `break` stops reading at `MaxCount` matches. `-l` sets `MaxCount = 1`, so a huge file is read only up
  to its first match.
- `sc.Err()` is `nil` after a normal EOF, `bufio.ErrTooLong` for a line over 1 MiB.
- `strings` is no longer imported.

### internal/search/walk.go (new)

`Walk` turns the PATHs into files to search, calling `fn` once per file. Create it:

```go
package search

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// WalkOptions controls which files Walk visits.
type WalkOptions struct {
	Hidden bool   // include hidden files and directories (names starting with ".")
	Glob   string // if not empty, only visit files whose base name matches it
}

// Walk calls fn for every file to search under roots, in order. A root
// that is a directory is walked recursively, in lexical order. The root
// "-" (standard input) is passed through unchanged. If a file or
// directory can't be read, fn is called with its path and the error.
//
// Hidden entries, entries that aren't regular files (such as symbolic
// links), and files not matching opts.Glob are skipped, but only when
// they are found inside a directory: a root is always used as given.
func Walk(roots []string, opts WalkOptions, fn func(path string, err error)) {
	for _, root := range roots {
		if root == "-" {
			fn(root, nil)
			continue
		}
		// Our callback never returns an error, so neither does WalkDir.
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fn(path, err)
				return nil // keep going
			}
			if path == root {
				if !d.IsDir() {
					fn(path, nil)
				}
				return nil
			}
			if !opts.Hidden && strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() {
					return filepath.SkipDir // don't descend into it
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil // a directory, symbolic link, device, ...
			}
			if opts.Glob != "" {
				// The pattern was checked by the caller, so Match can't fail.
				if ok, _ := filepath.Match(opts.Glob, d.Name()); !ok {
					return nil
				}
			}
			fn(path, nil)
			return nil
		})
	}
}
```

How it works:

- `filepath.WalkDir` on a file just visits it, so one code path handles files and directories.
- **Roots are used as given.** Naming `testdata/.hidden.txt` or `testdata/.cache` means you want it
  searched. Only entries *inside* a directory are filtered (the `path == root` check).
- A hidden directory returns `filepath.SkipDir`, so nothing inside is read.
- `d.Type().IsRegular()` is false for directories, symbolic links, devices, and pipes. Directories are
  still entered (returning `nil` means "go inside"), just not searched. Skipping symbolic links avoids
  loops and duplicates.
- `--glob` matches the base name, `d.Name()`, since `*` doesn't cross `/`.
- Errors (missing root, unreadable directory) go to `fn`, and returning `nil` keeps walking. So
  `WalkDir` never returns an error; `_ =` ignores it on purpose.

### internal/output/output.go

Show standard input as `(standard input)`, like grep, not `-`. Add `displayPath` after `printMatch`
and use it in the three places that print a path:

```go
// PrintFile implements Printer.
func (p *TextPrinter) PrintFile(path string, matches []search.Match) error {
	switch {
	case p.FilesOnly:
		if len(matches) == 0 {
			return nil
		}
		_, err := fmt.Fprintln(p.W, displayPath(path))
		return err
	case p.Count:
		_, err := fmt.Fprintf(p.W, "%s:%d\n", displayPath(path), len(matches))
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
		_, err = fmt.Fprintf(p.W, "%s:%d:%s\n", displayPath(m.Path), m.Line, text)
	} else {
		_, err = fmt.Fprintf(p.W, "%s:%s\n", displayPath(m.Path), text)
	}
	return err
}

// displayPath returns how path is shown to the user.
func displayPath(path string) string {
	if path == "-" {
		return "(standard input)"
	}
	return path
}
```

The rest of the file (`Printer`, `TextPrinter`, the color constants, `highlight`) is unchanged.

### cmd/seek/main.go

Five changes. The constants, `errHelp`, `usageError`, `usageErrorf`, and `report` stay as they are.

1. Import `bufio` and `path/filepath`:

```go
import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"seek/internal/output"
	"seek/internal/search"
)
```

2. `main` passes standard input to `run`:

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

3. `options` gets two fields. Run `gofmt -w .` (or format on save) to realign the comments, since
   `string` is longer than `bool`.

```go
// options holds everything the command line controls.
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool   // -i
	invert      bool   // -v
	lineNumbers bool   // -n
	count       bool   // -c
	filesOnly   bool   // -l
	maxCount    int    // -m N; 0 means no limit
	color       bool   // --color
	regex       bool   // -E
	hidden      bool   // --hidden
	glob        string // --glob PATTERN
	version     bool   // --version
}
```

4. `run` takes `stdin`, buffers stdout, and searches what `Walk` finds. Replace the whole function:

```go
// run is the whole program. It reads standard input from stdin, writes
// results to stdout and messages to stderr, and returns the exit status.
// Taking an io.Reader and io.Writers instead of using os.Stdin, os.Stdout,
// and os.Stderr directly lets tests supply input and capture output.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Writing many small pieces straight to a file costs one system call
	// each; a bufio.Writer collects them and writes in big chunks.
	out := bufio.NewWriter(stdout)
	defer out.Flush() // without this, the last chunk would be lost

	opts, err := parseArgs(args)
	if errors.Is(err, errHelp) {
		fmt.Fprintln(out, usage)
		return exitOK
	}
	if err != nil {
		report(stderr, err)
		return exitError
	}
	if opts.version {
		fmt.Fprintln(out, name, version)
		return exitOK
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
	// *output.TextPrinter. Chapter 15 adds a second implementation.
	var p output.Printer
	tp := &output.TextPrinter{
		W:           out,
		LineNumbers: opts.lineNumbers,
		Count:       opts.count,
		FilesOnly:   opts.filesOnly,
	}
	if opts.color {
		tp.Highlight = matcher
	}
	p = tp

	walkOpts := search.WalkOptions{Hidden: opts.hidden, Glob: opts.glob}

	// A bad file shouldn't stop the search: report it, remember that
	// something went wrong, and carry on with the next one.
	total, hadError := 0, false
	search.Walk(opts.paths, walkOpts, func(path string, err error) {
		var matches []search.Match
		if err == nil {
			matches, err = s.SearchFile(path)
		}
		if errors.Is(err, search.ErrBinary) {
			return // binary files are skipped silently
		}
		if err == nil {
			err = p.PrintFile(path, matches)
		}
		if err != nil {
			report(stderr, err)
			hadError = true
			return
		}
		total += len(matches)
	})

	switch {
	case hadError:
		return exitError
	case total == 0:
		return exitNoMatch
	default:
		return exitOK
	}
}
```

   The callback is a closure ([Chapter 4](04-functions.md)) that updates `run`'s `total` and
   `hadError`. Its `return` ends one call, like `continue`. The printer writes to the buffered `out`,
   not `stdout`.

5. `parseArgs` adds `--hidden`, `--glob`, a `value` helper for options with an argument, and the
   default PATH `.`. Replace the whole function:

```go
// parseArgs turns the command line into options.
func parseArgs(args []string) (options, error) {
	var opts options
	i := 0
	// value returns the argument that follows an option like -m N.
	value := func() (string, error) {
		i++
		if i >= len(args) {
			return "", usageErrorf("%s needs a value", args[i-1])
		}
		return args[i], nil
	}
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			return opts, errHelp
		case "--version":
			opts.version = true
			return opts, nil
		case "-i":
			opts.ignoreCase = true
		case "-v":
			opts.invert = true
		case "-n":
			opts.lineNumbers = true
		case "-c":
			opts.count = true
		case "-l":
			opts.filesOnly = true
		case "--color":
			opts.color = true
		case "-E":
			opts.regex = true
		case "--hidden":
			opts.hidden = true
		case "--glob":
			glob, err := value()
			if err != nil {
				return opts, err
			}
			// Match reports a malformed pattern no matter what the name is.
			if _, err := filepath.Match(glob, ""); err != nil {
				return opts, usageErrorf("invalid --glob pattern %q: %w", glob, err)
			}
			opts.glob = glob
		case "-m":
			v, err := value()
			if err != nil {
				return opts, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return opts, usageErrorf("invalid number for -m: %w", err)
			}
			if n < 0 {
				return opts, usageErrorf("invalid number for -m: %d is negative", n)
			}
			opts.maxCount = n
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return opts, usageErrorf("unknown option %s", arg)
			}
			break loop
		}
	}
	if i >= len(args) {
		return opts, usageErrorf("missing PATTERN")
	}
	opts.pattern = args[i]
	opts.paths = args[i+1:]
	if len(opts.paths) == 0 {
		opts.paths = []string{"."} // search the current directory
	}
	return opts, nil
}
```

   The `value` closure shares `i` with the loop, so calling it skips past the argument.
   `filepath.Match(glob, "")` returns `ErrBadPattern` for a broken pattern whatever the name, so the
   glob is checked once, up front. That's why `Walk` can ignore `Match`'s error.

### Checkpoint

Check the fixtures:

```console
$ find testdata | sort
testdata
testdata/.cache
testdata/.cache/fox.txt
testdata/.hidden.txt
testdata/image.bin
testdata/notes.md
testdata/poem.txt
testdata/src
testdata/src/hello.go
```

Search a directory. Hidden entries and the binary file are skipped:

```console
$ go run ./cmd/seek fox testdata
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
$ go run ./cmd/seek -n --hidden fox testdata
testdata/.cache/fox.txt:1:a cached fox
testdata/.hidden.txt:1:a hidden fox
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek -c fox testdata
testdata/notes.md:0
testdata/poem.txt:2
testdata/src/hello.go:1
```

Options must come before PATTERN. Here `--glob` and `*.md` follow it, so they become two more PATHs:

```console
$ go run ./cmd/seek -l -i fox --glob '*.md' testdata
seek: --glob: no such file or directory
seek: *.md: no such file or directory
testdata/notes.md
testdata/poem.txt
testdata/src/hello.go
exit status 2
$ go run ./cmd/seek --glob '*.md' -l -i fox testdata
testdata/notes.md
```

With no PATH, `seek` searches `.` (the project root). Of the `.go` files there, only `hello.go`
contains "fox":

```console
$ go run ./cmd/seek -n --glob '*.go' fox
testdata/src/hello.go:4:	println("hello, fox")
```

A root is used as given, even if hidden. A binary file is skipped even when named:

```console
$ go run ./cmd/seek fox testdata/.hidden.txt testdata/image.bin
testdata/.hidden.txt:a hidden fox
```

Standard input:

```console
$ printf 'one fox\ntwo dogs\nthree foxes\n' | go run ./cmd/seek -n fox -
(standard input):1:one fox
(standard input):3:three foxes
$ printf 'one fox\ntwo dogs\n' | go run ./cmd/seek -c fox - testdata/poem.txt
(standard input):1
testdata/poem.txt:2
```

Errors:

```console
$ go run ./cmd/seek --glob '[' fox testdata
seek: invalid --glob pattern "[": syntax error in pattern
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run ./cmd/seek fox testdata nope.txt
seek: nope.txt: no such file or directory
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
exit status 2
$ go run ./cmd/seek --glob
seek: --glob needs a value
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
```

`nope.txt` is searched *last*, but its error prints *first*: matches wait in the buffered `out` until
`run` returns, while errors go straight to unbuffered stderr. Each stream stays in order (a terminal
shows the same). For strict ordering across both, flush `out` before writing to stderr.

The complete project at this point is in `project/ch11/`.

## Summary

- `io.Reader` and `io.Writer` have one method each. Everything that moves bytes implements them, so
  code composes: wrap readers in readers, write to many writers at once.
- `io.Copy` and `io.ReadAll` run the read loop for you. `io.EOF` means "done", not "failed".
- `os.Open` reads, `os.Create` writes, `os.OpenFile` takes flags. `defer f.Close()` right after the
  error check; check `Close` errors for files you write.
- `bufio.Scanner` reads line by line. Check `sc.Err()` after the loop; raise the 64 KiB line limit
  with `sc.Buffer`.
- `bufio.Writer` batches writes. Always `Flush`; errors are sticky.
- Results to stdout, messages to stderr. `-` conventionally means stdin.
- `path/filepath` builds and splits OS paths. `filepath.Match` matches shell patterns; `*` doesn't
  cross `/`.
- `filepath.WalkDir` walks a tree in lexical order. Return `SkipDir` to skip a directory, `SkipAll` to
  stop.
- `fs.FS` abstracts a read-only tree; `os.DirFS`, `embed.FS`, and `fstest.MapFS` implement it.
- `os.Stat` returns size and mode. `Mode()&os.ModeCharDevice != 0` on stdout usually means a terminal.
- `bufio.Reader.Peek` looks ahead without consuming. A NUL byte near the start means binary.

## Exercises

1. **head.** Print the first 3 lines of standard input, then stop reading.
   `printf 'a\nb\nc\nd\ne\n' | go run .` should print `a`, `b`, `c`.

<details><summary>Solution</summary>

```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	for n := 0; n < 3 && sc.Scan(); n++ {
		fmt.Println(sc.Text())
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "head:", err)
		os.Exit(1)
	}
}
```

```console
$ printf 'a\nb\nc\nd\ne\n' | go run .
a
b
c
```

`n < 3` is checked first, so `Scan` isn't called a fourth time.

</details>

2. **A counting writer.** Write `countingWriter`: it wraps an `io.Writer`, passes writes through, and
   counts the bytes. Use it with `io.Copy` and `fmt.Fprintln`.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// countingWriter passes writes on to w and counts the bytes.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

func main() {
	cw := &countingWriter{w: os.Stdout}
	io.Copy(cw, strings.NewReader("hello\nworld\n"))
	fmt.Fprintln(cw, "one more line")
	fmt.Println("bytes written:", cw.n)
}
```

```console
$ go run .
hello
world
one more line
bytes written: 26
```

`Write` needs a pointer receiver because it changes `c.n`, so `*countingWriter` (not
`countingWriter`) is the `io.Writer`.

</details>

3. **Files by extension.** Walk the directory given as the first argument and print how many regular
   files have each extension, sorted by extension. Run it on `seek`'s `testdata` directory.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

func main() {
	counts := make(map[string]int)
	err := filepath.WalkDir(os.Args[1], func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			counts[filepath.Ext(path)]++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, ext := range slices.Sorted(maps.Keys(counts)) {
		fmt.Printf("%-5s %d\n", ext, counts[ext])
	}
}
```

```console
$ go run . ~/seek/testdata
.bin  1
.go   1
.md   1
.txt  3
$ go run . nope
lstat nope: no such file or directory
exit status 1
```

(Use your own `testdata` path.) This walk doesn't skip hidden entries, so all three `.txt` files count.

</details>

4. **seek: skip `node_modules`.** Make `seek` never enter a `node_modules` directory found while
   walking, even with `--hidden`. A root named `node_modules` is still searched.

<details><summary>Solution</summary>

In `internal/search/walk.go`, add a check just before the hidden check:

```go
			if d.IsDir() && d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			if !opts.Hidden && strings.HasPrefix(d.Name(), ".") {
```

It comes after the `path == root` check, so a root is still used as given. Try it:

```console
$ mkdir -p testdata/node_modules
$ printf 'a fox in node_modules\n' > testdata/node_modules/x.txt
$ go run ./cmd/seek fox testdata
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
$ go run ./cmd/seek fox testdata/node_modules
testdata/node_modules/x.txt:a fox in node_modules
```

Afterwards, delete `testdata/node_modules` and undo the change; later chapters expect the original
fixtures and code.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-11--readers-writers-and-the-file-system)
