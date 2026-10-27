# Chapter 8 — Packages, Modules, and Project Layout

A Go program is a tree of **packages**, and packages live in **modules**. This chapter covers how a
directory becomes a package, how capitalization controls visibility, how imports and `init` work,
how `internal/` hides code, how to document a package for `go doc`, and how modules pin third-party
dependencies. Then `seek` moves from one big `main.go` into `cmd/seek`, `internal/search`, and
`internal/output`. [Chapter 7](07-structs-methods-pointers.md) gave `seek` its types;
[Chapter 9](09-errors.md) adds real error handling.

## Packages are directories

A package is one directory of `.go` files:

- Every file starts with the same `package` clause.
- All files share one scope: `a.go` can call a function in `b.go` with no import.
- `package main` with `func main()` builds a program; any other name builds a library.
- You import a package by its **import path**: the module path from `go.mod`, then the directory.

The `examples` module (`module examples` in its `go.mod`) is laid out like this for the next
example:

```text
examples/
├── go.mod                      module examples
└── ch08/
    └── shapes/
        ├── main.go             package main
        └── geometry/
            └── geometry.go     package geometry → import "examples/ch08/shapes/geometry"
```

`examples/ch08/shapes/geometry/geometry.go`

```go
// Package geometry computes the areas of simple shapes.
package geometry

import "math"

// Unit is the length unit used in descriptions.
const Unit = "cm"

// A Circle is a round shape centered on the origin.
type Circle struct {
	Radius float64 // exported: other packages can read and set it
	label  string  // unexported: only package geometry can see it
}

// NewCircle returns a Circle with the given radius and label.
func NewCircle(radius float64, label string) Circle {
	return Circle{Radius: radius, label: label}
}

// Area returns the area of c.
func (c Circle) Area() float64 {
	return math.Pi * square(c.Radius)
}

// Label returns the circle's label, or "circle" if it has none.
func (c Circle) Label() string {
	if c.label == "" {
		return "circle"
	}
	return c.label
}

// square is unexported: a helper for this package only.
func square(x float64) float64 {
	return x * x
}
```

`examples/ch08/shapes/main.go`

```go
// Command shapes uses the geometry package from the same module.
package main

import (
	"fmt"

	"examples/ch08/shapes/geometry" // module path + directory
)

func main() {
	c := geometry.NewCircle(2, "coin")
	fmt.Printf("%s: %.2f %s²\n", c.Label(), c.Area(), geometry.Unit)

	var d geometry.Circle // zero value; label is empty
	d.Radius = 1
	fmt.Printf("%s: %.2f %s²\n", d.Label(), d.Area(), geometry.Unit)
}
```

```console
$ go run ./ch08/shapes
coin: 12.57 cm²
circle: 3.14 cm²
```

- Outside the package, use the package name as a prefix: `geometry.NewCircle`, `geometry.Unit`.
  Inside `geometry`, it's just `square`.
- The package name (`geometry`) is the last element of the import path. That's a convention, not a
  rule; break it only for good reason.
- No header files, no forward declarations. The compiler reads the whole directory.

`go list` prints the packages a pattern matches. `./...` means "this directory and everything
below":

```console
$ go list ./ch08/...
examples/ch08/imports
examples/ch08/init-order
examples/ch08/init-order/plugins/reverse
examples/ch08/init-order/plugins/upper
examples/ch08/init-order/registry
examples/ch08/internal-demo
examples/ch08/internal-demo/internal/secret
examples/ch08/shapes
examples/ch08/shapes/geometry
```

Two package names in one directory is an error. With `a.go` saying `package a` and `b.go` saying
`package b` (the path is your directory):

```console
$ go build .
found packages a (a.go) and b (b.go) in /path/to/dir
```

## Exported and unexported names

A name starting with an upper-case letter is **exported** (visible to other packages). Otherwise
it's **unexported** (private to its package). This applies to package-level declarations, struct
fields, and methods:

| Name | Kind | Visible outside `geometry`? |
|------|------|-----------------------------|
| `Circle`, `NewCircle`, `Unit` | type, function, constant | yes |
| `Radius`, `Area`, `Label` | field, method, method | yes |
| `label`, `square` | field, function | no |

Why case instead of `public`/`private` keywords? Every *use* tells you, without finding the
declaration: `geometry.Unit` is public, `square(x)` isn't.

Using unexported names from another package won't compile. In `examples/bad/main.go`:

```go
	c := geometry.Circle{Radius: 1, label: "x"}
	fmt.Println(geometry.square(2), c.label)
```

```console
$ go build .
# examples/bad
./main.go:10:34: cannot refer to unexported field label in struct literal of type geometry.Circle
./main.go:11:23: name square not exported by package geometry
./main.go:11:36: c.label undefined (type geometry.Circle has no field or method label, but does have method Label)
```

That's Go's encapsulation, and the unit of privacy is the **package**, not the type. Any code in
`geometry` can read `c.label`; other packages must use `NewCircle` and `Label`. Other packages can
still declare a zero `Circle` (as `main` did with `d`), so `Label` handles an empty `label`: design
for a working zero value.

Export as little as you can. Every exported name is a promise others may depend on, so you can't
change it freely.

## Naming packages

The package name is part of every use (`strings.Split`, `http.Get`), so choose it with care:

| Rule | Good | Avoid |
|------|------|-------|
| Short, lower case, one word | `search`, `output`, `strconv` | `searchUtils`, `search_utils` |
| Name it for what it provides | `geometry`, `json` | `util`, `common`, `helpers`, `misc` |
| Don't repeat the package name | `search.Searcher`, `bufio.Reader` | `search.SearchSearcher` |
| Match the directory name | `geometry/` → `package geometry` | `geometry/` → `package shapes` |

Names inside should read well with the prefix: `strings.Split`, not `strings.SplitString`.

## Imports

`import` takes one path or a parenthesized list, which `gofmt` sorts per group. Convention: the
standard library, a blank line, then everything else:

```go
import (
	"fmt"
	"os"

	"examples/ch08/shapes/geometry"
)
```

An unused import is a compile error, like an unused variable. This keeps builds fast and
dependencies honest:

```console
$ go build .
# examples/bad
./main.go:3:8: "os" imported and not used
```

Three special forms change the name a package gets in your file:

| Form | Meaning | Use it for |
|------|---------|------------|
| `name "path"` | alias: use `name.X` | two packages with the same name |
| `. "path"` | dot import: use `X` with no prefix | almost never |
| `_ "path"` | blank import: no name at all | running the package's `init` (next section) |

`examples/ch08/imports/main.go`

```go
// Import aliases and dot imports.
package main

import (
	crand "crypto/rand" // alias: two packages are both named "rand"
	"fmt"
	mrand "math/rand/v2" // the "/v2" is part of the path, not the name
	. "strings"          // dot import: use ToUpper without "strings." (avoid this)
)

func main() {
	b := make([]byte, 4)
	crand.Read(b) // cryptographically secure random bytes
	fmt.Println(len(b), "secure random bytes")

	n := mrand.IntN(10) // fast pseudo-random numbers
	fmt.Println(n >= 0 && n < 10)

	fmt.Println(ToUpper("dot imports hide where names come from"))
}
```

```console
$ go run ./ch08/imports
4 secure random bytes
true
DOT IMPORTS HIDE WHERE NAMES COME FROM
```

The random values change every run, so the program prints only facts about them.

- `crypto/rand` and `math/rand/v2` both declare `package rand`; without aliases they'd clash.
- `math/rand/v2` is the second major version of `math/rand`. `/v2` is in the path, but the package
  is still `rand` ([Versions](#versions-semver-and-minimal-version-selection) explains why).
- A dot import makes `ToUpper` look local, hiding where it comes from. Avoid it.

## Package initialization and `init`

Before `main` runs, Go initializes each package once:

1. A package's imports are initialized first. Among packages that are ready, Go goes in import path
   order.
2. Package-level variables get their values, in dependency order.
3. The package's `init` functions run, in source order.
4. When every package is ready, `main.main` runs.

`init` is special: no parameters, no results, and you can't call it. A file may have several.

Here a `registry` package and two plugins that register themselves in their `init` functions.
`main` imports the plugins with `_`, only for that side effect:

`examples/ch08/init-order/registry/registry.go`

```go
// Package registry keeps a list of named text transforms.
package registry

import (
	"fmt"
	"slices"
)

var transforms = map[string]func(string) string{}

// Package-level variables are initialized before any init function runs.
var ready = announce("registry: variables initialized")

func announce(msg string) bool {
	fmt.Println(msg)
	return true
}

func init() {
	fmt.Println("registry: init, ready =", ready)
}

// Register adds a transform under name.
func Register(name string, f func(string) string) {
	transforms[name] = f
}

// Names returns the registered names in sorted order.
func Names() []string {
	var names []string
	for name := range transforms {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Apply runs the transform called name on s.
func Apply(name, s string) string {
	return transforms[name](s)
}
```

`examples/ch08/init-order/plugins/upper/upper.go`

```go
// Package upper registers an upper-casing transform when imported.
package upper

import (
	"fmt"
	"strings"

	"examples/ch08/init-order/registry"
)

func init() {
	fmt.Println("upper: init")
	registry.Register("upper", strings.ToUpper)
}
```

`examples/ch08/init-order/plugins/reverse/reverse.go`

```go
// Package reverse registers a string-reversing transform when imported.
package reverse

import (
	"fmt"
	"slices"

	"examples/ch08/init-order/registry"
)

func init() {
	fmt.Println("reverse: init")
	registry.Register("reverse", func(s string) string {
		r := []rune(s)
		slices.Reverse(r)
		return string(r)
	})
}
```

`examples/ch08/init-order/main.go`

```go
// Command init-order shows when package initialization happens.
package main

import (
	"fmt"

	"examples/ch08/init-order/registry"

	// Blank imports: we never name these packages, we only want their
	// init functions to run and register their transforms.
	_ "examples/ch08/init-order/plugins/reverse"
	_ "examples/ch08/init-order/plugins/upper"
)

var greeting = makeGreeting()

func makeGreeting() string {
	fmt.Println("main: variables initialized")
	return "hello, gopher"
}

func init() {
	fmt.Println("main: first init")
}

func init() {
	fmt.Println("main: second init") // a file may have several init funcs
}

func main() {
	fmt.Println("main: main starts")
	for _, name := range registry.Names() {
		fmt.Printf("%s(%q) = %q\n", name, greeting, registry.Apply(name, greeting))
	}
}
```

```console
$ go run ./ch08/init-order
registry: variables initialized
registry: init, ready = true
reverse: init
upper: init
main: variables initialized
main: first init
main: second init
main: main starts
reverse("hello, gopher") = "rehpog ,olleh"
upper("hello, gopher") = "HELLO, GOPHER"
```

- `registry` goes first: both plugins import it.
- Within each package, variables come before `init`.
- `main` goes last, and `main()` runs after every `init`.

The standard library does this too: `import _ "image/png"` registers the PNG decoder with
`image.Decode`, and database drivers register with `database/sql`.

Use `init` sparingly. It runs hidden code at startup, can't be called from a test, and can't return
an error. Prefer explicit setup in `main` or a constructor.

## `internal` directories

A package whose path contains an `internal` element can be imported only by code rooted at the
parent of that `internal` directory. The compiler enforces this.

```text
examples/ch08/internal-demo/
├── main.go                    may import .../internal/secret
└── internal/
    └── secret/
        └── secret.go
```

`examples/ch08/internal-demo/internal/secret/secret.go`

```go
// Package secret can be imported only by code inside
// examples/ch08/internal-demo, because it lives under internal/.
package secret

// Answer is exported, but only to packages that may import secret.
func Answer() int { return 42 }
```

`examples/ch08/internal-demo/main.go`

```go
// Command internal-demo imports a package from its own internal/ directory.
package main

import (
	"fmt"

	"examples/ch08/internal-demo/internal/secret"
)

func main() {
	fmt.Println("the answer is", secret.Answer())
}
```

```console
$ go run ./ch08/internal-demo
the answer is 42
```

The same import from outside `internal-demo/` (here `examples/bad`) fails:

```console
$ go build ./bad
package examples/bad
	bad/main.go:6:2: use of internal package examples/ch08/internal-demo/internal/secret not allowed
```

Exported means "other packages may use this"; `internal/` adds "but only packages in this
project". No one outside can depend on code under `internal/`, so you can change it freely.

## Project layout: `cmd/` and `internal/`

Small programs can stay in one `main.go` at the module root. As a module grows, this layout is
common:

```text
mytool/
├── go.mod
├── cmd/
│   └── mytool/
│       └── main.go      package main: flags, wiring, exit codes
└── internal/
    ├── thing/           the real work, as library packages
    └── other/
```

- `cmd/NAME/` holds each program; `go build ./cmd/mytool` makes a binary named `mytool`. A module
  can have several programs.
- `internal/` holds packages only this module uses.
- Packages for other modules to import go at the root or in plain subdirectories.

Keep it simple. Don't make packages "just in case": start with one and split when a piece has a
clear job and a small API. Go needs no `src/` or `pkg/`.

## Doc comments and `go doc`

A **doc comment** sits directly above a `package` clause or exported declaration, with no blank line
between. Conventions:

- The package comment starts `Package name` (or `Command name` for a `main` package).
- A declaration's comment starts with its name: `// Area returns the area of c.`
- Write full sentences; tools show them as-is.

`go doc` prints them. It takes a package and an optional symbol:

```console
$ go doc ./ch08/shapes/geometry
package geometry // import "examples/ch08/shapes/geometry"

Package geometry computes the areas of simple shapes.

const Unit = "cm"
type Circle struct{ ... }
    func NewCircle(radius float64, label string) Circle
$ go doc ./ch08/shapes/geometry Circle
package geometry // import "examples/ch08/shapes/geometry"

type Circle struct {
	Radius float64 // exported: other packages can read and set it
	// Has unexported fields.
}
    A Circle is a round shape centered on the origin.

func NewCircle(radius float64, label string) Circle
func (c Circle) Area() float64
func (c Circle) Label() string
$ go doc ./ch08/shapes/geometry Circle.Area
package geometry // import "examples/ch08/shapes/geometry"

func (c Circle) Area() float64
    Area returns the area of c.

```

- `go doc` lists only exported names; `label` shows up as "Has unexported fields".
- `NewCircle` is listed under `Circle` because it returns one.
- It works on the standard library too, e.g. `go doc strings.Cut` or `go doc fmt`:

```console
$ go doc strings.Cut
package strings // import "strings"

func Cut(s, sep string) (before, after string, found bool)
    Cut slices s around the first instance of sep, returning the text before and
    after sep. The found result reports whether sep appears in s. If sep does
    not appear in s, cut returns s, "", false.

```

`go doc -all PKG` prints everything; `go doc -u` includes unexported names. Public modules get the
same docs as web pages on `pkg.go.dev`.

## Modules and dependencies

[Chapter 1](01-getting-started.md) made modules with `go mod init`. A module is a tree of packages
with a `go.mod` at its root, which sets:

- the **module path**, prefix of every import path in the module;
- the **go version**, the minimum Go the module needs;
- the **requirements**: other modules, each at a version.

`seek` uses only the standard library, so its `go.mod` has no requirements. To see dependencies at
work, make a throwaway module that uses `golang.org/x/text`, a Go-team module outside the standard
library with language-aware case mapping:

```console
$ mkdir hello && cd hello
$ go mod init example.com/hello
go: creating new go.mod: module example.com/hello
```

`hello/main.go`

```go
// Command hello title-cases its arguments with golang.org/x/text.
package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	text := strings.Join(os.Args[1:], " ")
	fmt.Println(cases.Title(language.English).String(text))
	fmt.Println(cases.Upper(language.German).String(text))
}
```

The module isn't required yet, so the build fails and says what to do:

```console
$ go run . hello straße
main.go:9:2: no required module provides package golang.org/x/text/cases; to add it:
	go get golang.org/x/text/cases
main.go:10:2: no required module provides package golang.org/x/text/language; to add it:
	go get golang.org/x/text/language
```

`go get` downloads a module and adds it to `go.mod`:

```console
$ go get golang.org/x/text
go: downloading golang.org/x/text v0.43.0
go: added golang.org/x/text v0.43.0
$ cat go.mod
module example.com/hello

go 1.26.10

require golang.org/x/text v0.43.0 // indirect
$ go run . hello straße
Hello Straße
HELLO STRASSE
```

Your version may be newer. `go get` records the requirement without checking your imports, so it
marks it `// indirect` ("not imported directly"). `go mod tidy` fixes that: it reads every import,
adds missing requirements, and removes unused ones:

```console
$ go mod tidy
$ cat go.mod
module example.com/hello

go 1.26.10

require golang.org/x/text v0.43.0
```

Run `go mod tidy` whenever you add or remove imports; it's the usual way to manage `go.mod`.

`go get` also wrote `go.sum`:

```console
$ cat go.sum
golang.org/x/text v0.43.0 h1:1QivrlhwAsnMOcqgOdl9my7s9OyZl8O5/UGUTJVXoWc=
golang.org/x/text v0.43.0/go.mod h1:hbSIYA/amXcRXTFSZqh8tWHEoqRvvbKPJ2uz/KU+a+s=
```

- Each line is a cryptographic hash: one for the module's files, one for its `go.mod`.
- The `go` command checks every download against `go.sum`. If a published version changes, the
  build fails instead of running different code.
- New modules are also checked against a public checksum database (`sum.golang.org`), so even your
  first download is verified.
- Commit both `go.mod` and `go.sum`.

Downloads come from the module proxy (`proxy.golang.org` by default) and are stored read-only in
the **module cache** (`go env GOMODCACHE`), shared by every project on your machine.

More `go mod` commands:

| Command | What it does |
|---------|--------------|
| `go list -m all` | list every module in the build |
| `go mod graph` | print the requirement graph |
| `go mod why PKG` | explain why a package is needed |
| `go mod verify` | check that cached modules still match `go.sum` |
| `go mod vendor` | copy all dependencies into a `vendor/` directory |

```console
$ go list -m all
example.com/hello
golang.org/x/mod v0.41.0
golang.org/x/sync v0.24.0
golang.org/x/text v0.43.0
golang.org/x/tools v0.51.0
```

`x/text`'s own `go.mod` requires the other three, so they're in the module graph. Go downloads only
the code your build actually imports.

**Vendoring.** `go mod vendor` copies dependency source into `vendor/`, and the `go` command builds
from there. Some teams use it to review dependency code in their repository or to build offline.
Most don't need it: the module cache plus `go.sum` already make builds reproducible.

## Versions: semver and minimal version selection

Module versions use **semantic versioning**: `vMAJOR.MINOR.PATCH`.

| Change | Bump | Example |
|--------|------|---------|
| Breaking API change | MAJOR | `v1.4.2` → `v2.0.0` |
| New feature, still compatible | MINOR | `v1.4.2` → `v1.5.0` |
| Bug fix only | PATCH | `v1.4.2` → `v1.4.3` |

- `v0.x.y` means "not stable yet": any release may break the API. `golang.org/x/text` is still `v0`.
- From `v2` on, the major version is part of the module path: `example.com/lib/v2`. Go treats `v1`
  and `v2` as different modules, and one program can use both. This is the **import compatibility
  rule**: same import path, no breaking change. The standard library does the same with
  `math/rand/v2`.

`go get` takes a version after `@`:

```console
$ go get golang.org/x/text@v0.30.0
go: downloading golang.org/x/text v0.30.0
go: downgraded golang.org/x/text v0.43.0 => v0.30.0
$ go get golang.org/x/text@latest
go: upgraded golang.org/x/text v0.30.0 => v0.43.0
```

`go get -u ./...` upgrades all dependencies to their latest minor or patch versions.

Which version wins when several modules need the same dependency? Say you require `A` and `B`. `A`
requires `C v1.2.0`, `B` requires `C v1.4.0`, and `C v1.9.0` exists:

```text
your module ──► A ──► C v1.2.0
            └─► B ──► C v1.4.0          chosen: C v1.4.0
```

Go uses **minimal version selection** (MVS): each requirement is a minimum, and Go picks the
smallest version that meets them all. Here that's `v1.4.0`, the highest minimum. It never picks
`v1.9.0` on its own, even though it's newer.

Why?

- Builds are reproducible from `go.mod` alone; a new release elsewhere can't change your build.
  That's why Go needs no lock file.
- You upgrade on purpose, with `go get`, then run your tests.

## Workspaces

When you change two modules at once, say an app and a library it uses, a **workspace** makes the
`go` command use your local copy of the library instead of a published version.

Two modules side by side, each with its own `go.mod` from `go mod init`: `greet`
(`module example.com/greet`) and `app` (`module example.com/app`), which imports it.

`greet/greet.go`

```go
// Package greet makes greetings.
package greet

// Hello returns a greeting.
func Hello() string { return "hello, workspace" }
```

`app/main.go`

```go
package main

import (
	"fmt"

	"example.com/greet"
)

func main() {
	fmt.Println(greet.Hello())
}
```

```console
$ cd app && go run .
main.go:6:2: no required module provides package example.com/greet; to add it:
	go get example.com/greet
$ cd ..
$ go work init ./app ./greet
$ cat go.work
go 1.26.10

use (
	./app
	./greet
)
$ go run ./app
hello, workspace
```

`go.work` lists the modules to use from disk. It's a local development tool; you usually don't
commit it. `go work use DIR` adds another module later.

## Building seek: packages

All of `seek` is one 325-line `main.go` doing three jobs: reading the command line, searching, and
printing. This step splits it into packages along those lines. **Behavior doesn't change**; almost
every line moves rather than changes.

The new layout:

```text
seek/
├── go.mod
├── cmd/
│   └── seek/
│       └── main.go          package main: options, run, parseArgs, usageError
├── internal/
│   ├── search/
│   │   └── search.go        package search: Match, Searcher, findAll, indexFold, prefixFold
│   └── output/
│       └── output.go        package output: Printer, printMatch, highlight
└── testdata/
    ├── notes.md
    └── poem.txt
```

Create the directories:

```console
$ mkdir -p cmd/seek internal/search internal/output
```

Then create the three files below and delete the old `main.go`. `go.mod` and `testdata/` don't
change.

What moves where:

| In Chapter 7 (`main.go`) | Now | Why |
|--------------------------|-----|-----|
| `Match`, `Searcher`, `FindAll`, `SearchFile` | `search.Match`, `search.Searcher`, ... | already exported |
| `selects`, `findAll`, `indexFold`, `prefixFold` | stay unexported in `search` | only `search` uses them |
| `NewSearcher(opts)` | removed; `run` builds `&search.Searcher{...}` | see below |
| `printer` | `output.Printer` | `main` needs the type |
| `lineNumbers`, `count`, `filesOnly`, `find` | `LineNumbers`, `Count`, `FilesOnly`, `Find` | `main` sets them |
| `printFile` | `PrintFile` | `main` calls it |
| `printMatch`, `highlight`, `colorStart`, `colorReset` | stay unexported in `output` | only `output` uses them |
| `options`, `run`, `parseArgs`, `usageError`, constants | stay in `package main` | command-line details |

The import graph has no cycles. `output` prints `search.Match` values, so it imports `search`;
`main` imports both:

```text
cmd/seek ──► internal/output ──► internal/search
    └──────────────────────────────────┘
```

### `internal/search/search.go`

The search code, unchanged except for the package clause, the package doc comment, and longer doc
comments on `Searcher` and `FindAll`, which now explain how to use the API.

`internal/search/search.go`

```go
// Package search finds lines of text that match a pattern.
package search

import (
	"os"
	"strings"
	"unicode/utf8"
)

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
// Set its fields before calling its methods; the zero value matches
// every line (the empty pattern matches everywhere).
type Searcher struct {
	Pattern    string
	IgnoreCase bool
	Invert     bool // select the lines that do NOT match
	MaxCount   int  // stop after this many matches per file; 0 means no limit
}

// FindAll returns the positions of up to n occurrences of the pattern in
// line (all of them if n < 0). Each position is a pair of byte offsets
// {start, end}. It returns nil if the pattern does not occur in line.
func (s *Searcher) FindAll(line string, n int) [][]int {
	return findAll(line, s.Pattern, s.IgnoreCase, n)
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.FindAll(line, 1) != nil
	return found != s.Invert // != on bools is "exclusive or"
}

// SearchFile returns the selected lines of the file at path.
func (s *Searcher) SearchFile(path string) ([]Match, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	var matches []Match
	for i, line := range strings.Split(text, "\n") {
		if !s.selects(line) {
			continue
		}
		matches = append(matches, Match{Path: path, Line: i + 1, Text: line})
		if len(matches) == s.MaxCount {
			break
		}
	}
	return matches, nil
}

// findAll returns the byte offsets {start, end} of up to n
// non-overlapping occurrences of pattern in line, or all of them if n < 0.
// It returns nil if there are none.
func findAll(line, pattern string, ignoreCase bool, n int) [][]int {
	if pattern == "" {
		return [][]int{{0, 0}} // the empty pattern matches every line
	}
	var spans [][]int
	for pos := 0; n < 0 || len(spans) < n; {
		var start, end int
		if ignoreCase {
			start, end = indexFold(line[pos:], pattern)
		} else {
			start = strings.Index(line[pos:], pattern)
			end = start + len(pattern)
		}
		if start < 0 {
			break
		}
		spans = append(spans, []int{pos + start, pos + end})
		pos += end
	}
	return spans
}

// indexFold is like strings.Index, but ignores case the way
// strings.EqualFold does. A rune and its other-case form may have
// different UTF-8 lengths (the Kelvin sign "\u212A" is 3 bytes, "k" is 1),
// so it returns the end offset of the match as well as the start.
// It returns -1, -1 if substr is not in s.
func indexFold(s, substr string) (int, int) {
	// Ranging over a string visits the byte offset of each rune.
	for start := range s {
		if n := prefixFold(s[start:], substr); n >= 0 {
			return start, start + n
		}
	}
	return -1, -1
}

// prefixFold reports how many bytes at the start of s equal prefix,
// ignoring case, or -1 if s does not start with prefix.
func prefixFold(s, prefix string) int {
	i := 0
	for _, want := range prefix {
		if i >= len(s) {
			return -1
		}
		got, size := utf8.DecodeRuneInString(s[i:])
		if got != want && !strings.EqualFold(string(got), string(want)) {
			return -1
		}
		i += size
	}
	return i
}
```

**Where did `NewSearcher` go?** In Chapter 7 it took `main`'s `options`:

```go
func NewSearcher(opts options) *Searcher
```

`options` belongs to `package main`, and `search` can't import `main`. Try adding `"seek/cmd/seek"`
as the first entry in `search.go`'s import list:

```console
$ go build ./...
internal/search/search.go:5:2: import "seek/cmd/seek" is a program, not an importable package
package seek/cmd/seek
	imports seek/internal/output from main.go
	imports seek/internal/search from output.go
	imports seek/cmd/seek from search.go: import cycle not allowed
```

Two problems:

- A `main` package is a program, not a library; nothing can import it.
- Even if it could, `main` → `output` → `search` → `main` is an **import cycle**. Go forbids
  cycles, so packages always have a clear build and initialization order.

The fix is to flip the dependency: `search` shouldn't know about the command line at all. Its fields
are exported and its zero value works, so `main` fills in a `Searcher` literal directly.
Lower-level packages know nothing about the packages that use them.

### `internal/output/output.go`

The printer, with exported names for the parts `main` uses. It now refers to `search.Match`, and its
doc comment says "standard output", since that's where `fmt.Printf` writes.

`internal/output/output.go`

```go
// Package output formats search results for the terminal.
package output

import (
	"fmt"
	"strings"

	"seek/internal/search"
)

// A Printer writes search results to standard output.
type Printer struct {
	LineNumbers bool // show line numbers
	Count       bool // print only the number of matches per file
	FilesOnly   bool // print only the names of files that have matches

	// Find locates the text to highlight in a line.
	// If it is nil, nothing is highlighted.
	Find func(line string, n int) [][]int
}

// PrintFile prints the results for one file.
func (p *Printer) PrintFile(path string, matches []search.Match) {
	switch {
	case p.FilesOnly:
		if len(matches) > 0 {
			fmt.Println(path)
		}
	case p.Count:
		fmt.Printf("%s:%d\n", path, len(matches))
	default:
		for _, m := range matches {
			p.printMatch(m)
		}
	}
}

// printMatch prints one matching line as path:text or path:line:text.
func (p *Printer) printMatch(m search.Match) {
	text := m.Text
	if p.Find != nil {
		text = highlight(text, p.Find(text, -1))
	}
	if p.LineNumbers {
		fmt.Printf("%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		fmt.Printf("%s:%s\n", m.Path, text)
	}
}

// ANSI escape sequences understood by most terminals.
const (
	colorStart = "\x1b[1;31m" // bold red
	colorReset = "\x1b[0m"
)

// highlight wraps each span of line in color escape sequences.
func highlight(line string, spans [][]int) string {
	var b strings.Builder
	prev := 0
	for _, span := range spans {
		start, end := span[0], span[1]
		if start == end {
			continue // an empty match has nothing to color
		}
		b.WriteString(line[prev:start])
		b.WriteString(colorStart)
		b.WriteString(line[start:end])
		b.WriteString(colorReset)
		prev = end
	}
	b.WriteString(line[prev:])
	return b.String()
}
```

`Find` is still a plain function field: `output` doesn't need to know a `Searcher` provides it.
`main` connects the two.

### `cmd/seek/main.go`

What's left of `main.go`: the constants, `options`, `run`, `parseArgs`, and `usageError`.
`parseArgs` and `usageError` don't change. `run` changes in three places:

- It builds the `Searcher` with a composite literal, including the `-l` rule from `NewSearcher`.
- It builds an `output.Printer` with the new field names.
- It calls `p.PrintFile`.

`cmd/seek/main.go`

```go
// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
	"strconv"

	"seek/internal/output"
	"seek/internal/search"
)

const (
	name    = "seek"
	version = "0.1.0"
	usage   = "Usage: seek [OPTIONS] PATTERN [PATH ...]"
)

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
	os.Exit(run(os.Args[1:]))
}

// options holds everything the command line controls.
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool // -i
	invert      bool // -v
	lineNumbers bool // -n
	count       bool // -c
	filesOnly   bool // -l
	maxCount    int  // -m N; 0 means no limit
	color       bool // --color
}

// run is the whole program. It returns the exit status.
func run(args []string) int {
	opts, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(opts.paths) == 0 {
		return usageError("missing PATH")
	}

	s := &search.Searcher{
		Pattern:    opts.pattern,
		IgnoreCase: opts.ignoreCase,
		Invert:     opts.invert,
		MaxCount:   opts.maxCount,
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}
	p := &output.Printer{
		LineNumbers: opts.lineNumbers,
		Count:       opts.count,
		FilesOnly:   opts.filesOnly,
	}
	if opts.color {
		p.Find = s.FindAll // a method value: FindAll bound to s
	}

	total := 0
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		total += len(matches)
		p.PrintFile(path, matches)
	}
	if total == 0 {
		return exitNoMatch
	}
	return exitOK
}

// parseArgs turns the command line into options. If the program should
// stop right away (after --help, --version, or a usage error), done is
// true and code is the exit status to use.
func parseArgs(args []string) (opts options, code int, done bool) {
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return opts, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return opts, exitOK, true
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
		case "-m":
			// -m takes a value: the next argument.
			i++
			if i >= len(args) {
				return opts, usageError("-m needs a number"), true
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return opts, usageError("invalid number for -m: " + strconv.Quote(args[i])), true
			}
			opts.maxCount = n
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return opts, usageError("unknown option " + arg), true
			}
			break loop
		}
	}
	if i >= len(args) {
		return opts, usageError("missing PATTERN"), true
	}
	opts.pattern = args[i]
	opts.paths = args[i+1:]
	return opts, 0, false
}

// usageError reports a command-line mistake and returns the exit status.
func usageError(msg string) int {
	fmt.Fprintln(os.Stderr, "seek:", msg)
	fmt.Fprintln(os.Stderr, usage)
	return exitError
}
```

`options` and its fields stay lowercase; only `package main` uses them.

### Running it now

With no Go file at the module root, `go run .` stops working (the path is your project directory):

```console
$ go run . fox testdata/poem.txt
no Go files in /path/to/seek
```

From now on, run `go run ./cmd/seek` from the project root. `go build ./cmd/seek` builds a binary
named `seek`.

### Checkpoint

```console
$ go list ./...
seek/cmd/seek
seek/internal/output
seek/internal/search
$ go run ./cmd/seek -n -i fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/notes.md:2:- feed the FOX
$ go run ./cmd/seek -l fox testdata/poem.txt testdata/notes.md
testdata/poem.txt
$ go doc ./internal/search
package search // import "seek/internal/search"

Package search finds lines of text that match a pattern.

type Match struct{ ... }
type Searcher struct{ ... }
$ go doc ./internal/search Searcher.FindAll
package search // import "seek/internal/search"

func (s *Searcher) FindAll(line string, n int) [][]int
    FindAll returns the positions of up to n occurrences of the pattern in line
    (all of them if n < 0). Each position is a pair of byte offsets {start,
    end}. It returns nil if the pattern does not occur in line.

$ go doc ./internal/output Printer
package output // import "seek/internal/output"

type Printer struct {
	LineNumbers bool // show line numbers
	Count       bool // print only the number of matches per file
	FilesOnly   bool // print only the names of files that have matches

	// Find locates the text to highlight in a line.
	// If it is nil, nothing is highlighted.
	Find func(line string, n int) [][]int
}
    A Printer writes search results to standard output.

func (p *Printer) PrintFile(path string, matches []search.Match)
```

Every earlier checkpoint gives the same output with `go run .` replaced by `go run ./cmd/seek`. The
complete project at this point is in `project/ch08/`.

## Summary

- A package is a directory whose files share one `package` clause and one scope. `package main`
  builds a program.
- Import path = module path + directory. The package name is usually the last element.
- Upper-case names are exported; lower-case are private to the package. The package, not the type,
  is the unit of privacy.
- Package names are short, lower case, and say what the package provides. Avoid `util`.
- Unused imports are errors. Imports can be aliased (`name "path"`), dotted (avoid), or blank (`_`,
  for side effects).
- Initialization: imported packages, then package variables, then `init` functions, then `main`.
  Keep `init` rare.
- `internal/` packages are importable only within the tree rooted at `internal`'s parent.
- A common layout is `cmd/NAME/main.go` plus `internal/...`.
- Doc comments start with the name they document; `go doc` reads them.
- `go.mod` holds the module path, Go version, and requirements; `go.sum` holds checksums. Commit
  both. `go get` adds or changes versions; `go mod tidy` cleans up.
- Semver: `v2+` puts the major version in the import path. MVS picks the lowest version satisfying
  every requirement, so builds are reproducible without a lock file.
- `go work` combines local modules for development; `go mod vendor` copies dependencies into the
  repository.
- Import cycles are compile errors. Lower-level packages must not import higher-level ones.

## Exercises

1. Add a `Rect` type with exported `Width` and `Height` fields and an `Area` method to
   `examples/ch08/shapes/geometry`. Give both doc comments, use `Rect` from `shapes/main.go`, and
   check `go doc` for the package and for `Rect`.

<details><summary>Solution</summary>

Append to `geometry.go`:

```go
// A Rect is an axis-aligned rectangle.
type Rect struct {
	Width, Height float64
}

// Area returns the area of r.
func (r Rect) Area() float64 {
	return r.Width * r.Height
}
```

Add to the end of `main` in `shapes/main.go`:

```go
	r := geometry.Rect{Width: 3, Height: 4}
	fmt.Printf("rect: %.2f %s²\n", r.Area(), geometry.Unit)
```

```console
$ go run ./ch08/shapes
coin: 12.57 cm²
circle: 3.14 cm²
rect: 12.00 cm²
$ go doc ./ch08/shapes/geometry
package geometry // import "examples/ch08/shapes/geometry"

Package geometry computes the areas of simple shapes.

const Unit = "cm"
type Circle struct{ ... }
    func NewCircle(radius float64, label string) Circle
type Rect struct{ ... }
$ go doc ./ch08/shapes/geometry Rect
package geometry // import "examples/ch08/shapes/geometry"

type Rect struct {
	Width, Height float64
}
    A Rect is an axis-aligned rectangle.

func (r Rect) Area() float64
```

`Circle` and `Rect` can both have `Area`: methods belong to their type, so the names don't clash.

</details>

2. Make a module `example.com/tally` with a package `counter`. Its `Counter` keeps its count in an
   unexported field, so other packages can only increase it (`Inc`) and read it (`Value`). The zero
   `Counter` should be ready to use. Use it from `main`, then try to set the field directly from
   `main` and read the error.

<details><summary>Solution</summary>

`counter/counter.go`

```go
// Package counter provides a counter that can only go up.
package counter

// A Counter counts events. The zero value is ready to use.
type Counter struct {
	n int // unexported: other packages can't change it directly
}

// Inc adds one to the count.
func (c *Counter) Inc() {
	c.n++
}

// Value returns the current count.
func (c *Counter) Value() int {
	return c.n
}
```

`main.go`

```go
package main

import (
	"fmt"

	"example.com/tally/counter"
)

func main() {
	var c counter.Counter
	c.Inc()
	c.Inc()
	fmt.Println(c.Value())
}
```

```console
$ go run .
2
```

Adding `c.n = -5` after the two `Inc` calls:

```console
$ go run .
# example.com/tally
./main.go:13:4: c.n undefined (cannot refer to unexported field n)
```

Only package `counter` can break the "only goes up" rule, so the rule is easy to check.

</details>

3. Predict the output, then run it. The module is `examples`; the files are `order/main.go`,
   `order/a/a.go`, and `order/b/b.go`.

```go
// order/b/b.go
package b

import "fmt"

var B = trace("b.B", 1)

func trace(name string, v int) int {
	fmt.Println("init var", name)
	return v
}

func init() { fmt.Println("init func b") }
```

```go
// order/a/a.go
package a

import (
	"fmt"

	"examples/order/b"
)

var A = trace("a.A", b.B+1)

func trace(name string, v int) int {
	fmt.Println("init var", name)
	return v
}

func init() { fmt.Println("init func a") }
```

```go
// order/main.go
package main

import (
	"fmt"

	"examples/order/a"
)

var M = a.A + 1

func init() { fmt.Println("init func main") }

func main() {
	fmt.Println("main:", M)
}
```

<details><summary>Solution</summary>

```console
$ go run ./order
init var b.B
init func b
init var a.A
init func a
init func main
main: 3
```

`main` imports `a`, which imports `b`, so `b` initializes completely (variables, then `init`) before
`a` starts, then `a`, then `main`. Each package finishes before the next begins. Both `a` and `b`
declare an unexported `trace` without clashing, because each package has its own scope.

</details>

4. In `seek`, `output` imports `search`. What if `search` also imports `output`? Add
   `"seek/internal/output"` to the imports in `internal/search/search.go`, plus a line that uses it,
   such as `var _ output.Printer`, and build. How would you share code both packages need?

<details><summary>Solution</summary>

```console
$ go build ./...
package seek/cmd/seek
	imports seek/internal/output from main.go
	imports seek/internal/search from output.go
	imports seek/internal/output from search.go: import cycle not allowed
```

Go rejects import cycles, so the package graph can have no loops. To share code between two
packages, move it into the lower one (here, `search`) or into a new third package both import. Undo
the change afterward.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-8--packages-modules-and-project-layout)
