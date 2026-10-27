# Chapter 7 — Structs, Methods, and Pointers

Go has no classes. **Structs** group data, **methods** attach behavior to types, and **pointers**
share a value instead of copying it. You'll also meet embedding (Go's composition) and how Go
manages memory. Then `seek` replaces its parallel slices and option map with an `options` struct,
a `Match` type, and a `Searcher` with methods, and gains `-l` and `-m N`.
[Chapter 6](06-strings.md) covered strings; [Chapter 8](08-packages-and-modules.md) splits `seek`
into packages.

## Structs

A struct type is a list of named fields. Literals build values; a dot reads or writes a field.

`examples/ch07/structs/main.go`

```go
// Struct types, literals, zero values, and comparison.
package main

import "fmt"

// A Book groups related values under one name.
type Book struct {
	Title  string
	Author string
	Pages  int
	Tags   []string
}

// Fields of the same type can share a line.
type Point struct {
	X, Y int
}

func main() {
	// A literal with field names. Order doesn't matter; missing fields
	// get their zero value.
	b := Book{Title: "The Go Programming Language", Pages: 380}
	b.Author = "Donovan & Kernighan" // set a field
	fmt.Printf("%+v\n", b)

	// The zero value of a struct has every field at its zero value.
	var empty Book
	fmt.Printf("%q %q %d %v %t\n", empty.Title, empty.Author, empty.Pages, empty.Tags, empty.Tags == nil)

	// A positional literal lists every field in order. Fine for tiny
	// types like Point; fragile for anything bigger.
	p := Point{3, 4}

	// Structs are compared field by field with == if every field is comparable.
	fmt.Println(p == Point{X: 3, Y: 4}, p == Point{})

	// Comparable structs can be map keys.
	visits := map[Point]int{p: 1}
	visits[Point{3, 4}]++
	fmt.Println(visits)

	// Assignment copies the whole struct.
	q := p
	q.X = 100
	fmt.Println(p, q)
}
```

```console
$ go run ./ch07/structs
{Title:The Go Programming Language Author:Donovan & Kernighan Pages:380 Tags:[]}
"" "" 0 [] true
true false
map[{3 4}:2]
{3 4} {100 4}
```

- **Use field names in literals.** `Book{Title: "..."}` still compiles after someone adds a field.
  A positional literal must list every field, in order:

  ```console
  ./main.go:8:17: too few values in struct literal of type Point
  ```

- **The zero value is usable.** `var empty Book` is a valid book with every field zeroed. Design
  types so that means something sensible (an empty list, a counter at 0).
- **`==` needs every field comparable.** Slices, maps, and functions aren't, so a struct holding
  one can't use `==` or be a map key:

  ```console
  ./main.go:12:14: invalid operation: a == b (struct containing []string cannot be compared)
  ```

- **Assignment copies every field.** Changing `q` leaves `p` alone. (A slice or map field copies as
  a reference, so both copies share its elements.)

### Anonymous structs

An inline struct type with no name. Handy for one-off values and test-case tables.

`examples/ch07/anonymous-structs/main.go`

```go
// Anonymous structs: struct types without a name.
package main

import (
	"fmt"
	"strings"
)

func main() {
	// A one-off value that groups a few settings.
	server := struct {
		Host string
		Port int
	}{"localhost", 8080}
	fmt.Printf("%s:%d\n", server.Host, server.Port)

	// A table of cases: a very common use, especially in tests.
	cases := []struct {
		in, want string
	}{
		{"go", "GO"},
		{"café", "CAFÉ"},
		{"", ""},
	}
	for _, c := range cases {
		got := strings.ToUpper(c.in)
		fmt.Printf("%q -> %q ok=%t\n", c.in, got, got == c.want)
	}

	// The empty struct takes no memory. map[string]struct{} is a set.
	seen := map[string]struct{}{"a": {}}
	_, ok := seen["a"]
	fmt.Println(ok)
}
```

```console
$ go run ./ch07/anonymous-structs
localhost:8080
"go" -> "GO" ok=true
"café" -> "CAFÉ" ok=true
"" -> "" ok=true
true
```

Inside a slice literal you can omit the element type: `{"go", "GO"}`, not
`struct{ in, want string }{"go", "GO"}`. [Chapter 19](19-testing.md) uses this for table-driven
tests.

## Pointers

A pointer holds a variable's address.

| Syntax | Means |
|--------|-------|
| `*T` | the type "pointer to T" |
| `&x` | the address of variable `x` |
| `*p` | the variable `p` points to (read or write) |
| `new(T)` | a new zero `T`; returns `*T` |
| `&T{...}` | a new `T` from a literal; returns `*T` |
| `nil` | the zero value of every pointer type |

`examples/ch07/pointers/main.go`

```go
// Pointers: &, *, new, and nil.
package main

import "fmt"

type Point struct {
	X, Y int
}

func main() {
	x := 1
	p := &x         // p has type *int and points at x
	fmt.Println(*p) // *p reads the value p points to
	*p = 2          // and writes it
	fmt.Println(x)  // x changed

	// new(T) allocates a zero T and returns a pointer to it.
	n := new(int)
	*n += 10
	fmt.Println(*n)

	// &T{...} is the usual way to get a pointer to a new struct.
	pt := &Point{X: 1, Y: 2}
	pt.X = 10 // shorthand for (*pt).X: Go follows the pointer for you
	fmt.Println(pt.X, (*pt).Y, *pt)

	// You can point at a field or a slice element too.
	py := &pt.Y
	*py = 20
	nums := []int{1, 2, 3}
	second := &nums[1]
	*second = 200
	fmt.Println(*pt, nums)

	// The zero value of a pointer is nil: it points at nothing.
	var none *Point
	fmt.Println(none == nil, none)

	// Two pointers are equal if they point at the same variable.
	a, b := Point{}, Point{}
	fmt.Println(&a == &a, &a == &b, a == b)
}
```

```console
$ go run ./ch07/pointers
1
2
10
10 2 {10 2}
{10 20} [1 200 3]
true <nil>
true false true
```

Go pointers are deliberately limited, which keeps them safe:

- **No pointer arithmetic:**

  ```console
  ./main.go:8:2: invalid operation: p++ (non-numeric type *int)
  ```

- **Fields dereference automatically:** `pt.X` works on a `*Point`, so you rarely write `(*pt).X`.
- **Dereferencing `nil` panics** with a stack trace:

  ```console
  $ go run .
  panic: runtime error: invalid memory address or nil pointer dereference
  [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x49e136]

  goroutine 1 [running]:
  main.main()
  	/home/you/nilderef/main.go:9 +0x16
  exit status 2
  ```

  (Addresses and paths will differ.) Check for `nil` when a pointer may be absent. Panics are in
  [Chapter 9](09-errors.md).

- **Returning a local variable's address is safe.** It lives as long as something points to it
  (see [memory](#memory-stack-heap-and-the-garbage-collector)).

## Everything is passed by value

Go copies every argument; there's no pass by reference. To let a function change your variable,
pass a pointer. The pointer is copied, but the copy points at the same variable.

`examples/ch07/pass-by-value/main.go`

```go
// Go passes everything by value: arguments are copies.
package main

import "fmt"

type Counter struct {
	Name string
	N    int
}

func incCopy(c Counter) { c.N++ } // changes a copy; the caller sees nothing

func incPtr(c *Counter) { c.N++ } // the pointer is a copy, but it points at the original

func setFirst(s []int) { s[0] = 99 } // a slice header holds a pointer to its array

func grow(s []int) { s = append(s, 4) } // changes only the local copy of the header

func main() {
	c := Counter{Name: "hits"}
	incCopy(c)
	fmt.Println(c.N)
	incPtr(&c)
	fmt.Println(c.N)

	nums := []int{1, 2, 3}
	setFirst(nums)
	grow(nums)
	fmt.Println(nums)

	m := map[string]int{}
	func(m map[string]int) { m["k"] = 1 }(m) // maps are references too
	fmt.Println(m)
}
```

```console
$ go run ./ch07/pass-by-value
0
1
[99 2 3]
map[k:1]
```

Slices, maps, channels, functions, and strings are small values that *contain* pointers. Copies are
cheap, and a copied slice or map reaches the same elements. But it's still a copy: `grow` appended
to its own slice header, so the caller never saw the new length. That's why `append` returns the
new slice ([Chapter 5](05-collections.md)).

Why? One rule is easier to reason about: `f(x)` can't change `x`; `f(&x)` might.

## Methods

A method is a function with a **receiver** before its name. You call it with a dot.

`examples/ch07/methods/main.go`

```go
// Methods: functions with a receiver.
package main

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

// Dist has a value receiver: it gets a copy of the Point.
func (p Point) Dist(q Point) float64 {
	return math.Hypot(q.X-p.X, q.Y-p.Y)
}

// Scale has a pointer receiver: it changes the Point it is called on.
func (p *Point) Scale(f float64) {
	p.X *= f
	p.Y *= f
}

// Methods work on any named type defined in your package, not just structs.
type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

func main() {
	p := Point{3, 4}
	fmt.Println(p.Dist(Point{}))

	p.Scale(2) // Go takes &p for you because p is a variable
	fmt.Println(p)

	pp := &Point{1, 1}
	fmt.Println(pp.Dist(Point{})) // and follows the pointer for value methods

	fmt.Println(Celsius(100).Fahrenheit())
}
```

```console
$ go run ./ch07/methods
5
{6 8}
1.4142135623730951
212
```

- The receiver is just a first parameter in a special place. There's no hidden `this`; you name it.
  Convention: short (often one letter) and the same in every method of the type.
- You can add methods to **any named type declared in your package**: structs, `type Celsius
  float64`, `type literal string`, named slices. Not to built-in types like `int` or types from
  other packages like `strings.Builder`.

## Value or pointer receiver?

`func (p Point)` gets a copy. `func (p *Point)` gets a pointer to the caller's value.

`examples/ch07/receivers/main.go`

```go
// Value receivers get a copy; pointer receivers share the original.
package main

import "fmt"

type Counter struct {
	n int
}

// IncBroken has a value receiver, so it increments a copy. A classic bug.
func (c Counter) IncBroken() { c.n++ }

// Inc has a pointer receiver, so it increments the caller's Counter.
func (c *Counter) Inc() { c.n++ }

// A List is a linked list. A nil *List is a valid empty list.
type List struct {
	Val  int
	Next *List
}

// Len works even when l is nil: a pointer method may be called on nil.
func (l *List) Len() int {
	if l == nil {
		return 0
	}
	return 1 + l.Next.Len()
}

func main() {
	var c Counter
	c.IncBroken()
	fmt.Println(c.n)
	c.Inc()
	c.Inc()
	fmt.Println(c.n)

	var empty *List
	l := &List{1, &List{2, &List{3, nil}}}
	fmt.Println(empty.Len(), l.Len())
}
```

```console
$ go run ./ch07/receivers
0
2
0 3
```

Use a **pointer receiver** when the method changes the receiver, when the struct is large (copying
each call is wasteful), or when it holds something that mustn't be copied (a `strings.Builder`, or a
`sync.Mutex` from [Chapter 17](17-sync.md)). Use a **value receiver** for small, unchanging values
like `Point`, `Celsius`, or `time.Time`.

**Be consistent.** If any method needs a pointer receiver, give them all pointer receivers. Mixing
confuses readers and causes surprises with interfaces.

### The calling rules and method sets

Go inserts `&` and `*` when it can:

- `v.PtrMethod()` on a **variable** `v` becomes `(&v).PtrMethod()`.
- `p.ValueMethod()` on a pointer `p` becomes `(*p).ValueMethod()`.

The first rule needs an addressable value (one with a fixed place in memory). Map elements and
function results aren't addressable:

```go
m := map[string]Point{"a": {1, 2}}
m["a"].Scale(2)
m["a"].X = 5
```

```console
./main.go:9:9: cannot call pointer method Scale on Point
./main.go:10:2: cannot assign to struct field m["a"].X in map
```

Store pointers in the map (`map[string]*Point`), or copy the value out, change it, and store it
back.

Formally, this is the **method set**:

| Type | Method set |
|------|-----------|
| `T` | methods with receiver `T` |
| `*T` | methods with receiver `T` **and** `*T` |

It matters most for interfaces: a `T` value lacks the `*T` methods, so it may not satisfy an
interface that `*T` does ([Chapter 10](10-interfaces.md)).

## Constructors and useful zero values

Go has no built-in constructors. Instead:

1. **Make the zero value useful** when you can, so `var s Stack` just works.
2. If a type needs setup, write a **`NewT`** function returning a ready `*T` (or `T`).

`examples/ch07/constructors/main.go`

```go
// Constructor functions and useful zero values.
package main

import (
	"fmt"
	"strings"
)

// A Stack of strings. The zero value is an empty stack, ready to use,
// so it needs no constructor.
type Stack struct {
	items []string
}

func (s *Stack) Push(v string) { s.items = append(s.items, v) }

func (s *Stack) Pop() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// A Ring keeps the last few strings added. It needs a size, so it has a
// constructor that sets it up. By convention it is named NewRing.
type Ring struct {
	buf  []string
	next int
}

func NewRing(size int) *Ring {
	if size < 1 {
		size = 1
	}
	return &Ring{buf: make([]string, size)}
}

func (r *Ring) Add(v string) {
	r.buf[r.next] = v
	r.next = (r.next + 1) % len(r.buf)
}

func (r *Ring) String() string { return strings.Join(r.buf, ",") }

func main() {
	var s Stack // no setup needed
	s.Push("a")
	s.Push("b")
	v, ok := s.Pop()
	fmt.Println(v, ok)

	r := NewRing(3)
	for _, w := range []string{"a", "b", "c", "d"} {
		r.Add(w)
	}
	fmt.Println(r.String())
}
```

```console
$ go run ./ch07/constructors
b true
d,b,c
```

`"d"` overwrote the oldest entry, `"a"`. A zero `Ring` has an empty buffer, so `Add` would panic on
`r.buf[0]`; that's why `Ring` gets a constructor and `Stack` doesn't. Constructors also validate
input and set defaults.

## Embedding: composition, not inheritance

A struct can include another type as a field **with no field name**. This is **embedding**. The
embedded type's fields and methods are **promoted**: usable as if they were the outer struct's.

`examples/ch07/embedding/main.go`

```go
// Struct embedding: composition with promoted fields and methods.
package main

import "fmt"

// Logger writes messages with a prefix.
type Logger struct {
	Prefix string
}

func (l *Logger) Log(msg string) { fmt.Println(l.Prefix + msg) }

// A Server embeds a Logger: the field has a type but no name.
type Server struct {
	Logger // embedded; the field's name is Logger
	Addr   string
}

// Start calls Log as if it were Server's own method.
func (s *Server) Start() { s.Log("listening on " + s.Addr) }

// A Cache also embeds Logger, but defines its own Log.
type Cache struct {
	Logger
}

// Log shadows the promoted Logger.Log.
func (c *Cache) Log(msg string) { c.Logger.Log("[cache] " + msg) }

func main() {
	s := &Server{Logger: Logger{Prefix: "srv: "}, Addr: ":8080"}
	s.Start()
	s.Log("promoted method")     // same as s.Logger.Log(...)
	fmt.Println(s.Prefix)        // promoted field: same as s.Logger.Prefix
	s.Logger.Prefix = "server: " // the embedded value is an ordinary field
	s.Log("changed prefix")

	c := &Cache{Logger{Prefix: "app: "}}
	c.Log("hit") // the outer Log wins
}
```

```console
$ go run ./ch07/embedding
srv: listening on :8080
srv: promoted method
srv: 
server: changed prefix
app: [cache] hit
```

How this differs from inheritance:

- **It's still a field**, named after its type (`s.Logger`). Promotion is a shortcut for
  `s.Logger.Log`.
- **No "is-a".** A `*Server` is not a `*Logger`; you can't pass one as the other.
- **No virtual dispatch.** Inside `Logger.Log`, the receiver is the `Logger`; it knows nothing of
  the `Server` or `Cache` around it.
- **The outer type wins.** An outer field or method with the same name shadows the promoted one
  (`Cache.Log`). If two embedded types at the same depth both have `Log`, `x.Log` is a compile
  error; say which one.
- You can embed a pointer (`*Logger`); the outer zero value then has a `nil` logger.

Why? Deep class hierarchies are hard to change. Embedding gives reuse without the coupling;
interfaces ([Chapter 10](10-interfaces.md)) give polymorphism.

## Method values and method expressions

- A **method value** `x.M` binds the method to receiver `x`, giving a function to call later, store,
  or pass along.
- A **method expression** `T.M` or `(*T).M` is an ordinary function taking the receiver as its
  first argument.

`examples/ch07/method-values/main.go`

```go
// Method values and method expressions.
package main

import "fmt"

type Counter struct {
	Name string
	n    int
}

func (c *Counter) Inc()          { c.n++ }
func (c Counter) Report() string { return fmt.Sprintf("%s=%d", c.Name, c.n) }

// repeat calls f n times. It knows nothing about Counter.
func repeat(n int, f func()) {
	for range n {
		f()
	}
}

func main() {
	hits := &Counter{Name: "hits"}

	// A method value binds a method to one receiver: a func() here.
	inc := hits.Inc
	repeat(3, inc)
	fmt.Println(hits.Report())

	// A value receiver is copied when the method value is made.
	report := hits.Report
	hits.Inc()
	fmt.Println(report(), hits.Report())

	// A method expression turns a method into an ordinary function
	// whose first parameter is the receiver.
	incFn := (*Counter).Inc // type func(*Counter)
	reportFn := Counter.Report
	misses := &Counter{Name: "misses"}
	incFn(misses)
	fmt.Println(reportFn(*misses))
	fmt.Printf("%T\n%T\n", inc, incFn)
}
```

```console
$ go run ./ch07/method-values
hits=3
hits=3 hits=4
misses=1
func()
func(*main.Counter)
```

Note line two: `report` copied the `Counter` when it was made (value receiver), so it still says
`hits=3`. A pointer-receiver method value copies only the pointer, so it sees later changes. `seek`
uses a method value below to give the printer "find matches with this searcher".

## Struct tags

A **tag** is a raw string after a field's type. The compiler ignores it; other packages read it at
run time. The most common use names JSON keys:

`examples/ch07/struct-tags/main.go`

```go
// Struct tags: metadata on fields, read by other packages.
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
	age      int    // unexported: invisible to other packages
}

func main() {
	u := User{Name: "Ada", Password: "secret", age: 36}
	// json.Marshal reads the tags to choose key names. Chapter 15
	// covers encoding/json; here we only peek at what tags do.
	data, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
```

```console
$ go run ./ch07/struct-tags
{"name":"Ada"}
```

- `Email` is omitted because it's empty (`omitempty`); `Password` because its tag is `"-"`.
- `age` is lowercase, so `encoding/json` can't see it ([Chapter 8](08-packages-and-modules.md)
  explains why). [Chapter 15](15-json-and-encoding.md) covers tags in depth.
- Convention: space-separated `key:"value"` pairs, e.g. `` `json:"name" xml:"name"` ``.

## Memory: stack, heap, and the garbage collector

You never choose where a value lives, and you never free it.

- **Escape analysis:** the compiler decides whether each variable fits on the goroutine's **stack**
  (cheap; gone when the function returns) or must go on the **heap** (it outlives the function).
- The **garbage collector** frees heap memory once nothing points to it. It runs alongside your
  program and keeps pauses short.

So returning `&p` for a local `p` is safe: the compiler sees `p` escape and heap-allocates it.

`examples/ch07/escape/main.go`

```go
// Escape analysis: the compiler decides between stack and heap.
package main

import "fmt"

type Point struct {
	X, Y int
}

// sum's Point never outlives the call, so it can live on the stack.
func sum() int {
	p := Point{1, 2}
	return p.X + p.Y
}

// newPoint returns a pointer to a local variable. That is safe in Go:
// the compiler sees that p outlives the call and puts it on the heap.
func newPoint() *Point {
	p := Point{3, 4}
	return &p
}

func main() {
	fmt.Println(sum(), *newPoint())
}
```

```console
$ go run ./ch07/escape
3 {3 4}
```

`-gcflags=-m` shows the compiler's decisions (`-l` turns off inlining, for a clearer report):

```console
$ go run -gcflags='-m -l' ./ch07/escape
# examples/ch07/escape
ch07/escape/main.go:19:2: moved to heap: p
ch07/escape/main.go:24:13: ... argument does not escape
ch07/escape/main.go:24:17: sum() escapes to heap
ch07/escape/main.go:24:21: *newPoint() escapes to heap
3 {3 4}
```

Only `newPoint`'s `p` moved to the heap; `sum`'s `Point` stayed on the stack. The two "escapes to
heap" lines are `fmt.Println` copying its arguments into heap values ([Chapter 10](10-interfaces.md)
explains why).

Don't contort code to avoid the heap. Write clear code; check `-m` only if a profile
([Chapter 20](20-tooling-and-design.md)) shows allocation is a problem.

## Building seek: `Options`, `Match`, and `Searcher`

`seek` has outgrown loose variables:

- `parseArgs` returns five results; options live in a `map[string]bool` the compiler can't check.
- `searchFile` returns parallel slices (`nums`, `lines`) that must stay in step.
- `run` mixes searching and printing.

New types fix this:

| Type | Role |
|------|------|
| `options` | everything the command line sets |
| `Match` | one selected line: path, line number, text |
| `Searcher` | the search settings, with methods `FindAll`, `selects`, and `SearchFile` |
| `printer` | the output settings, with methods `printFile` and `printMatch` |

New options: `-l` prints only names of files with a match; `-m N` stops after `N` selected lines
per file (`0` = no limit).

Declaration order in a Go file doesn't matter; the steps follow `project/ch07/main.go`.

### Step 1: imports

Drop `slices`; add `strconv` to parse `-m`'s number:

```go
import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)
```

### Step 2: the `options` struct and a new `parseArgs`

Add `options` right after `main`:

```go
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
```

The names are lowercase because only this file uses them (Chapter 8 explains why case matters).

Delete `boolOptions`; each option now gets its own `case` that sets a typed field. Replace
`parseArgs` with:

```go
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
```

- The named result `opts` starts as the zero `options`: no pattern, flags off, no limit. Each
  `case` turns one thing on. A useful zero value at work.
- A typo like `opts.ignorCase` is now a compile error. With the map, `opts["-I"]` was silently
  `false`.
- `-m` takes the **next** argument as its value (`i++`). `strconv.Quote` quotes bad input in the
  error.

`usageError` is unchanged.

### Step 3: the `Match` and `Searcher` types

Delete `searchFile`. Add these after `usageError`:

```go
// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Pattern    string
	IgnoreCase bool
	Invert     bool // select the lines that do NOT match
	MaxCount   int  // stop after this many matches per file; 0 means no limit
}

// NewSearcher returns a Searcher configured from the command-line options.
func NewSearcher(opts options) *Searcher {
	s := &Searcher{
		Pattern:    opts.pattern,
		IgnoreCase: opts.ignoreCase,
		Invert:     opts.invert,
		MaxCount:   opts.maxCount,
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}
	return s
}

// FindAll returns the positions of up to n occurrences of the pattern in
// line (all of them if n < 0).
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
```

- **`[]Match` replaces the parallel slices.** Each match carries path, number, and text together.
- **`NewSearcher` is a constructor.** It also applies one rule: `-l` needs only one match per file,
  so `MaxCount = 1` stops the search early.
- **`selects` replaces the `-v` closure.** `found != s.Invert` is true when exactly one side is: a
  match without `-v`, or a non-match with `-v`.
- **`MaxCount` 0 means no limit**: after `append`, `len(matches)` is at least 1, so it never equals
  0.
- **All methods use pointer receivers.** `Searcher` is small, but `SearchFile` runs once per file,
  and one receiver style per type is the rule.

`findAll`, `indexFold`, `prefixFold`, and the color constants are unchanged.

### Step 4: the `printer` type

Add `printer` and its methods between the color constants and `highlight`:

```go
// A printer writes search results to standard output.
type printer struct {
	lineNumbers bool // -n: show line numbers
	count       bool // -c: print only a count per file
	filesOnly   bool // -l: print only the names of matching files
	// find locates the text to highlight. If it is nil, nothing is.
	find func(line string, n int) [][]int
}

// printFile prints the results for one file.
func (p *printer) printFile(path string, matches []Match) {
	switch {
	case p.filesOnly:
		if len(matches) > 0 {
			fmt.Println(path)
		}
	case p.count:
		fmt.Printf("%s:%d\n", path, len(matches))
	default:
		for _, m := range matches {
			p.printMatch(m)
		}
	}
}

// printMatch prints one matching line as path:text or path:line:text.
func (p *printer) printMatch(m Match) {
	text := m.Text
	if p.find != nil {
		text = highlight(text, p.find(text, -1))
	}
	if p.lineNumbers {
		fmt.Printf("%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		fmt.Printf("%s:%s\n", m.Path, text)
	}
}
```

`find` is a **function-typed field**. The printer doesn't know `Searcher`, only "a function that
finds spans in a line"; `nil` means no highlighting. The condition-less `switch` picks the first
true case, so `-l` beats `-c`.

`highlight` is unchanged.

### Step 5: a shorter `run`

`run` builds a `Searcher` and a `printer`, then loops over the files. Replace it with:

```go
// run is the whole program. It returns the exit status.
func run(args []string) int {
	opts, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(opts.paths) == 0 {
		return usageError("missing PATH")
	}

	s := NewSearcher(opts)
	p := printer{
		lineNumbers: opts.lineNumbers,
		count:       opts.count,
		filesOnly:   opts.filesOnly,
	}
	if opts.color {
		p.find = s.FindAll // a method value: FindAll bound to s
	}

	total := 0
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		total += len(matches)
		p.printFile(path, matches)
	}
	if total == 0 {
		return exitNoMatch
	}
	return exitOK
}
```

- **`p.find = s.FindAll` is a method value**: "call `FindAll` on this `s`" as a
  `func(string, int) [][]int`, exactly the field's type.
- **`p` is a value, but `printFile` is a pointer method.** `p` is a variable, so Go takes `&p`.
- **The count map is gone.** Files print in `opts.paths` order; `-c` prints each count as it goes.

### Checkpoint

```console
$ go run . -l -i fox testdata/poem.txt testdata/notes.md
testdata/poem.txt
testdata/notes.md
$ go run . -m 1 -n -i fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:1:The quick brown fox
testdata/notes.md:2:- feed the FOX
$ go run . -c -m 2 -i fox testdata/poem.txt
testdata/poem.txt:2
$ go run . -n -v -m 1 fox testdata/poem.txt
testdata/poem.txt:2:jumps over the lazy dog.
$ go run . -m testdata/poem.txt
seek: invalid number for -m: "testdata/poem.txt"
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run . -m two fox testdata/poem.txt
seek: invalid number for -m: "two"
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run . -m
seek: -m needs a number
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
```

- `-l -i` lists both files; `notes.md` has `FOX`.
- `-c -m 2` counts at most 2, though `poem.txt` has 3 matches with `-i`.
- With `-v`, `-m` counts selected (non-matching) lines.
- `-m testdata/poem.txt` takes the path as the number, so the error shows the path.

The [Chapter 6](06-strings.md) checkpoints still give the same output. The complete project at
this point is in `project/ch07/`.

## Summary

- Structs group named fields. Use field names in literals; make the zero value useful.
- `==` works if all fields are comparable. Assignment copies the struct.
- Pointers (`*T`, `&x`, `*p`, `new`) share a variable. No pointer arithmetic; `nil` dereference
  panics.
- Arguments are always copied. Pass a pointer to let a function change your variable.
- Use pointer receivers to modify, for large structs, or for values that mustn't be copied. One
  receiver style per type.
- Go takes `&v` for you on addressable values; map elements aren't addressable.
- `NewT` sets up types whose zero value isn't enough.
- Embedding promotes fields and methods: composition, not inheritance.
- Method values (`x.M`) bind a receiver; method expressions (`T.M`) don't.
- Struct tags are metadata for packages like `encoding/json`.
- Escape analysis picks stack or heap; the GC frees memory. Returning `&local` is safe.

## Exercises

1. Define `Point{X, Y int}` and `Rect{Min, Max Point}`. Give `Rect` an `Area() int` method and a
   `Translate(dx, dy int)` method that moves it. Which receiver types should they have?

<details><summary>Solution</summary>

```go
package main

import "fmt"

type Point struct {
	X, Y int
}

type Rect struct {
	Min, Max Point
}

// Area only reads the Rect, but uses a pointer receiver to match Translate.
func (r *Rect) Area() int {
	return (r.Max.X - r.Min.X) * (r.Max.Y - r.Min.Y)
}

// Translate moves the Rect, so it needs a pointer receiver.
func (r *Rect) Translate(dx, dy int) {
	r.Min.X += dx
	r.Max.X += dx
	r.Min.Y += dy
	r.Max.Y += dy
}

func main() {
	r := Rect{Min: Point{0, 0}, Max: Point{3, 2}}
	r.Translate(10, 10)
	fmt.Printf("%+v area=%d\n", r, r.Area())
}
```

```console
$ go run .
{Min:{X:10 Y:10} Max:{X:13 Y:12}} area=6
```

`Translate` needs a pointer receiver. `Area` could take a value, but by convention one style per
type, so both use `*Rect`.

</details>

2. Write an `Animal` struct with a `Name` and pointer methods `Describe() string` and
   `Rename(string)`. Embed it in a `Dog` that adds a `Breed` and its own `Describe`, which calls the
   embedded one and appends the breed. Rename a dog through the promoted method, then print both
   `Describe` results.

<details><summary>Solution</summary>

```go
package main

import "fmt"

type Animal struct {
	Name string
}

func (a *Animal) Describe() string { return a.Name + " is an animal" }
func (a *Animal) Rename(n string)  { a.Name = n }

type Dog struct {
	Animal
	Breed string
}

func (d *Dog) Describe() string {
	return d.Animal.Describe() + ", a " + d.Breed
}

func main() {
	d := &Dog{Animal: Animal{Name: "Rex"}, Breed: "beagle"}
	d.Rename("Max") // promoted from Animal
	fmt.Println(d.Describe())
	fmt.Println(d.Animal.Describe())
}
```

```console
$ go run .
Max is an animal, a beagle
Max is an animal
```

</details>

3. Predict the output of this program, then run it to check.

   ```go
   package main

   import "fmt"

   type Temp struct {
   	C float64
   }

   func (t Temp) Set(c float64)     { t.C = c }
   func (t *Temp) SetPtr(c float64) { t.C = c }

   func main() {
   	t := Temp{20}
   	t.Set(30)
   	fmt.Println(t.C)
   	set := t.SetPtr
   	set(40)
   	fmt.Println(t.C)
   	u := t
   	u.SetPtr(50)
   	fmt.Println(t.C, u.C)
   }
   ```

<details><summary>Solution</summary>

```console
$ go run .
20
40
40 50
```

- `Set` has a value receiver, so it changes a copy: `t.C` stays 20.
- `t.SetPtr` is a method value bound to `&t`, so `set(40)` changes `t`.
- `u := t` copies the struct; `u.SetPtr(50)` changes only `u`.

</details>

4. In the project, add `--max-count` as a long name for `-m`. Error messages should name the option
   the user typed.

<details><summary>Solution</summary>

In `parseArgs`, list both names in the `case` and use `arg` (the option as typed) in the messages:

```go
		case "-m", "--max-count":
			// -m takes a value: the next argument.
			i++
			if i >= len(args) {
				return opts, usageError(arg + " needs a number"), true
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return opts, usageError("invalid number for " + arg + ": " + strconv.Quote(args[i])), true
			}
			opts.maxCount = n
```

```console
$ go run . --max-count 1 -n fox testdata/poem.txt
testdata/poem.txt:1:The quick brown fox
$ go run . --max-count x fox testdata/poem.txt
seek: invalid number for --max-count: "x"
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
```

`arg` comes from the `switch` init (`switch arg := args[i]; arg`), so every case can use it. (The
snapshots skip this change; [Chapter 12](12-flags.md) adds long names for all options.)

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-7--structs-methods-and-pointers)
