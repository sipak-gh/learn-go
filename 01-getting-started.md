# Chapter 1 — Getting Started

Meet Go and its daily tools: the `go` command, modules, `gofmt`, and `go vet`. You'll run your
first programs, read compiler errors, and start the `seek` project.
[Chapter 2](02-values-and-types.md) covers variables, types, and constants.

## How this course works

- Read the chapters in order ([README](README.md) has the contents). Each uses only what came
  before.
- Every concept has a small program you can run.
- Each chapter adds a feature to `seek`, a `grep`-like search tool.
- Exercises end each chapter; solutions are in collapsible blocks.

You need **Go 1.26 or newer** (installing it isn't covered). Check:

```console
$ go version
go version go1.26.10 linux/amd64
```

Any `go1.26` or later works; your OS and architecture may differ.

### Conventions

- **Commands.** Type lines starting with `$ ` (without the `$`). Lines below are the output:

  ```console
  $ go run ./ch01/hello
  Hello, Go!
  ```

- **Example paths.** Each program's path (e.g. `examples/ch01/hello/main.go`) appears above it.
  From `examples/`, run it with `go run ./chNN/<name>`.
- **Project snapshots.** `project/chNN/` is the complete `seek` project at the end of chapter NN.
  Compare against it if you get stuck.
- **Shell.** Commands assume a Unix-like shell (Linux, macOS, WSL/Git Bash). On Windows, built
  programs end in `.exe`.

To type the examples yourself, create an `examples` module (see [Modules](#modules-and-gomod)):

```console
$ mkdir examples
$ cd examples
$ go mod init examples
go: creating new go.mod: module examples
```

Give each example its own folder, such as `ch01/hello/main.go`.

## What Go is

Go is compiled, statically typed, and garbage collected. Google started it in 2007, open sourced
it in 2009, and shipped Go 1.0 in 2012. It's built for large programs written by many people over
years. Its design goals shape this course:

| Goal | In practice |
|------|-------------|
| Simplicity | 25 keywords. Usually one obvious way to do things. |
| Readability | Code is read more than written. Clear beats clever. |
| Fast builds | Big programs compile in seconds. Unused imports are errors. |
| Concurrency | Goroutines and channels are built in ([Chapter 16](16-goroutines-and-channels.md)). |
| One format | `gofmt` formats all Go code the same way. No style debates. |
| Easy deployment | `go build` makes one self-contained binary. |
| Stability | Go 1 compatibility promise: code that compiles today keeps compiling. |

Go leaves out classes, inheritance, and exceptions on purpose. You'll learn what it uses instead,
and why.

## Your first program

`examples/ch01/hello/main.go`

```go
// Command hello prints a greeting.
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

```console
$ go run ./ch01/hello
Hello, Go!
```

Line by line:

- `// Command hello ...`: a comment right above `package` documents the program.
- `package main`: every file starts with a package clause. Package `main` builds into a runnable
  program.
- `import "fmt"`: makes package `fmt` (formatted I/O) available. It's in the **standard library**,
  the packages that ship with Go.
- `func main()`: the starting point. No arguments, no results. When it returns, the program ends.
- `fmt.Println(...)`: prints its arguments and a newline.

No semicolons at line ends: the compiler inserts them (why that matters:
[Reading compiler errors](#reading-compiler-errors)).

## Modules and go.mod

A **package** is a folder of `.go` files compiled together. A **module** is a set of packages with
a `go.mod` file at its root. Every program lives in a module. Create one with `go mod init`:

```console
$ mkdir hello
$ cd hello
$ go mod init hello
go: creating new go.mod: module hello
```

The new `go.mod`:

```text
module hello

go 1.26.10
```

- `module hello` is the **module path**, the name used to import its packages. Published modules
  use a URL-like path (`github.com/you/hello`); local programs can use a short name.
- `go 1.26.10` is the minimum Go version. `go mod init` writes yours, so it may differ. `go 1.26`
  accepts any Go 1.26 release.

Without a `go.mod`, most `go` commands refuse to work:

```console
$ go run .
go: go.mod file not found in current directory or any parent directory; see 'go help modules'
```

[Chapter 8](08-packages-and-modules.md) goes deeper into modules and dependencies.

## The go command

One tool does almost everything:

| Command | What it does |
|---------|--------------|
| `go run .` | Compile and run the package in this folder. Saves nothing. |
| `go build` | Compile and save the program (named after the module or folder). |
| `go fmt` | Format the files in place (runs `gofmt -l -w`). |
| `go vet` | Report code that compiles but is probably wrong. |
| `go doc fmt.Println` | Show docs for a package or symbol. |
| `go mod init NAME` | Create a module. |
| `go version` | Print the Go version. |
| `go help` / `go help build` | List commands / explain one. |

`.` means "the package in this folder"; `./ch01/hello` means "the package in `ch01/hello`".

`go run` is for trying things out. `go build` makes a program you can keep and share:

```console
$ go build
$ ./hello
Hello, Go!
```

The binary has no dependencies; it runs on any machine with the same OS and CPU architecture.

`go doc` prints documentation from source, offline:

```console
$ go doc fmt.Println
package fmt // import "fmt"

func Println(a ...any) (n int, err error)
    Println formats using the default formats for its operands and writes to
    standard output. Spaces are always added between operands and a newline
    is appended. It returns the number of bytes written and any write error
    encountered.

```

[Chapter 4](04-functions.md) explains signatures like `func Println(a ...any) (n int, err error)`.

## Printing with fmt

Three print functions for day one:

`examples/ch01/printing/main.go`

```go
// Command printing shows the three basic print functions in fmt.
package main

import "fmt"

func main() {
	// Println puts spaces between its arguments and ends with a newline.
	fmt.Println("Go", "has", 25, "keywords:", true)

	// Print adds no newline. It adds a space between two arguments
	// only when neither of them is a string.
	fmt.Print("one", "two", "\n")
	fmt.Print(1, 2, "\n")

	/* Printf fills in a format string. Each verb (%s, %d, %v)
	   is replaced by the next argument. */
	fmt.Printf("%s was announced in %d.\n", "Go", 2009)
	fmt.Printf("%v works for any value: %v, %v\n", "%v", 3.5, false)
}
```

```console
$ go run ./ch01/printing
Go has 25 keywords: true
onetwo
1 2
Go was announced in 2009.
%v works for any value: 3.5, false
```

- `Println` is simplest; use it to see values.
- `Printf` gives control but adds no newline; write `\n` yourself.
- Common verbs: `%v` (any value), `%s` (string), `%d` (integer), `%.2f` (2 decimals). More in
  [Chapter 2](02-values-and-types.md); `fmt` in depth in [Chapter 6](06-strings.md).

## Imports

Import several packages with a parenthesized list:

`examples/ch01/imports/main.go`

```go
// Command imports uses three packages from the standard library.
package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("gopher"))
	fmt.Println(math.Sqrt(2))
	fmt.Printf("pi is about %.2f\n", math.Pi)
}
```

```console
$ go run ./ch01/imports
GOPHER
1.4142135623730951
pi is about 3.14
```

- Use imported names as `package.Name`: `strings.ToUpper`, `math.Pi`.
- Only capitalized names are visible outside their package (**exported**): `fmt.Println` is,
  `fmt.println` wouldn't be. No `public`/`private` keywords needed.
- An unused import is a **compile error**, not a warning: it slows builds and confuses readers.

## Comments

```go
// A line comment runs to the end of the line.

/* A block comment can span
   several lines. */
```

Go code mostly uses `//`, even for long blocks. A comment directly above a declaration is its
**doc comment**, and starts with the name it describes:

```go
// Command hello prints a greeting.
package main
```

`go doc` shows them, so write them for anything others use. More in
[Chapter 8](08-packages-and-modules.md).

## gofmt: one format for everyone

`gofmt` applies Go's one official format. This messy code still compiles:

```go
package main
import "fmt"
func main()  {
fmt.Println( "Hello, Go!" )   // greet
	   }
```

`gofmt -l` lists unformatted files. `gofmt -d` shows what it would change:

```console
$ gofmt -l .
main.go
$ gofmt -d main.go
diff main.go.orig main.go
--- main.go.orig
+++ main.go
@@ -1,5 +1,7 @@
 package main
+
 import "fmt"
-func main()  {
-fmt.Println( "Hello, Go!" )   // greet
-	   }
+
+func main() {
+	fmt.Println("Hello, Go!") // greet
+}
```

`go fmt` rewrites the files and prints the ones it changed:

```console
$ go fmt
main.go
```

Treat `gofmt` as non-negotiable; have your editor run it on save. Since all Go code looks alike,
you can read any project without adjusting. It indents with tabs and aligns with spaces, so you
don't have to.

## go vet: compiles, but probably wrong

`go vet` finds mistakes the compiler allows. Here `%d` expects a number but gets a string:

```go
package main

import "fmt"

func main() {
	fmt.Printf("%d apples\n", "three")
}
```

```console
$ go vet
main.go:6:14: fmt.Printf format %d has arg "three" of wrong type string
$ go run .
%!d(string=three) apples
```

It runs but prints garbage: `%!d(...)` is how `fmt` reports a bad verb. Run `go vet` often;
`go test` runs some of its checks too ([Chapter 19](19-testing.md)). When `gofmt -l .` and
`go vet` print nothing, all is well.

## Reading compiler errors

Go's compiler errors are short, precise, and all shaped alike:

```console
$ go run .
# hello
./main.go:6:2: "os" imported and not used
```

| Part | Meaning |
|------|---------|
| `# hello` | The package that failed (module `hello`'s root package). |
| `./main.go` | The file. |
| `6:2` | Line 6, column 2. Editors jump straight there. |
| `"os" imported and not used` | What's wrong. |

Tips:

- **Fix the first error first.** One mistake can cause others. The compiler stops after 10.
- **Read the column.** It points at the exact unexpected token.
- **Look one line up** for syntax errors. A missing `)` or `,` is often reported at the next token.

Common beginner errors, each from a small change to the hello program:

**Forgot the import** (deleted the `import "fmt"` line):

```console
./main.go:6:2: undefined: fmt
```

**Wrong capitalization** (`fmt.println`):

```console
./main.go:7:6: undefined: fmt.println (but have Println)
```

**Unused import** (imported `"os"` without using it):

```console
./main.go:6:2: "os" imported and not used
```

**Opening brace on its own line:**

```go
func main()
{
	fmt.Println("Hello, Go!")
}
```

```console
./main.go:7:1: syntax error: unexpected semicolon or newline before {
```

Why? The compiler adds a semicolon after any line ending in a name, a literal, `)`, `]`, or `}`.
`func main()` becomes `func main();`, so the `{` below starts nothing. This rule lets Go drop
semicolons, and it's why the opening brace goes on the same line.

**Missing parenthesis** (`fmt.Println("one"`):

```console
./main.go:7:19: syntax error: unexpected newline in argument list; possibly missing comma or )
```

**Two functions with the same name:**

```console
./main.go:9:6: main redeclared in this block
	./main.go:6:6: other declaration of main
```

The indented line points to the related spot.

**Wrong package name** (`package hello` instead of `package main`):

```console
$ go run .
package hello is not a main package
```

Errors while the program *runs* are **panics** and look different. You'll see one in
[Chapter 5](05-collections.md); [Chapter 9](09-errors.md) covers error handling.

## Building seek: create the project

You'll build `seek` throughout the course. Its final usage:

```text
seek [OPTIONS] PATTERN [PATH ...]
```

It will find lines containing PATTERN, walk directories, search in parallel, and support options
like `-i` (ignore case), `-n` (line numbers), and `--json`. For now it only prints its usage line.

Create the module in a project folder outside `examples/`:

```console
$ mkdir seek
$ cd seek
$ go mod init seek
go: creating new go.mod: module seek
```

Your `go.mod` ends with your Go version (e.g. `go 1.26.10`); the `project/` snapshots say
`go 1.26`. Both work.

Create `main.go`:

```go
// Command seek searches files for lines that match a pattern.
package main

import "fmt"

func main() {
	fmt.Println("Usage: seek [OPTIONS] PATTERN [PATH ...]")
}
```

Usage-line convention: `[...]` is optional, `...` is "one or more", and capitalized words are
placeholders.

### Checkpoint

```console
$ go run .
Usage: seek [OPTIONS] PATTERN [PATH ...]
$ go build
$ ./seek
Usage: seek [OPTIONS] PATTERN [PATH ...]
```

`go build` names the binary after the module: `seek` (`seek.exe` on Windows). Then check your
code; silence means all is well:

```console
$ gofmt -l .
$ go vet
```

Delete the binary afterward (`rm seek`), or add `seek` to `.gitignore`.

The complete project at this point is in `project/ch01/`.

## Summary

- Go is small, compiled, and statically typed, built for readability, fast builds, and big teams.
- A program is `package main` with `func main()`. A module is a tree of packages with a `go.mod`
  at its root (`go mod init`).
- `go run` compiles and runs; `go build` makes a self-contained binary.
- `fmt.Println` prints values with spaces and a newline; `fmt.Printf` uses a format string.
- Capitalized names are exported. Unused imports are compile errors.
- `gofmt` is the one format. `go vet` catches mistakes the compiler allows.
- Errors read `file:line:column: message`. Fix the first one first.
- Semicolon insertion is why `{` goes on the same line.

## Exercises

1. Print this, using one `fmt.Printf` call for the two middle lines:

   ```text
   --------------------
   Hello,
   Gopher!
   --------------------
   ```

   `go doc strings.Repeat` shows a function that makes the dashes.

   <details><summary>Solution</summary>

   ```go
   package main

   import (
   	"fmt"
   	"strings"
   )

   func main() {
   	fmt.Println(strings.Repeat("-", 20))
   	fmt.Printf("%s\n%s\n", "Hello,", "Gopher!")
   	fmt.Println(strings.Repeat("-", 20))
   }
   ```

   </details>

2. In the hello program, change `Println` to `PrintLn`. Predict the error, then run it.

   <details><summary>Solution</summary>

   ```console
   $ go run .
   # hello
   ./main.go:7:6: undefined: fmt.PrintLn (but have Println)
   ```

   Names are case-sensitive. When one differs from a real name only in case, the compiler
   suggests the right one.

   </details>

3. `go build -o NAME` sets the output file name. Build `seek` as `mytool`, run it, and delete it.

   <details><summary>Solution</summary>

   ```console
   $ go build -o mytool
   $ ./mytool
   Usage: seek [OPTIONS] PATTERN [PATH ...]
   $ rm mytool
   ```

   </details>

4. Format this code in your head, then check with `gofmt`:

   ```go
   package main
   import ("fmt";"math")
   func main(){fmt.Println(math.Max(3,7))}
   ```

   <details><summary>Solution</summary>

   ```go
   package main

   import (
   	"fmt"
   	"math"
   )

   func main() { fmt.Println(math.Max(3, 7)) }
   ```

   `gofmt` fixes spacing and indentation but keeps your line breaks inside a function, so the
   one-line body stays. The program prints `7`.

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-1--getting-started)
