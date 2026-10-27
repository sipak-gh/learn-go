# Chapter 10 — Interfaces

An **interface** describes what a value can *do*, as a list of methods. Any type with those methods
satisfies it automatically; there's no `implements` keyword. This one idea gives Go polymorphism,
decoupled packages, and small, composable APIs like `io.Reader` and `io.Writer`. This chapter covers
interface types and values, the `nil`-interface trap, `any`, type assertions and switches,
embedding, method sets, and a short look at reflection. Then `seek` gets a `Matcher` interface with
three implementations (one a regular expression, for the new `-E` option) and a `Printer` interface
that writes to any `io.Writer`. [Chapter 9](09-errors.md) covered errors (an interface you've
already used); [Chapter 11](11-io-and-files.md) builds on `io.Reader` and `io.Writer`.

## Interfaces and implicit satisfaction

An interface type is a set of method signatures. A value of that type can hold any value whose type
has all of them:

`examples/ch10/shapes/main.go`

```go
// An interface lists methods. Any type with those methods satisfies it.
package main

import (
	"fmt"
	"math"
)

// A Shape is anything with an area and a perimeter.
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct {
	W, H float64
}

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Circle struct {
	R float64
}

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

// describe works with any Shape, including types written later.
func describe(s Shape) {
	fmt.Printf("%T %v: area %.2f, perimeter %.2f\n", s, s, s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}, Rect{W: 10, H: 0.5}}
	total := 0.0
	for _, s := range shapes {
		describe(s)
		total += s.Area()
	}
	fmt.Printf("total area %.2f\n", total)
}
```

```console
$ go run ./ch10/shapes
main.Rect {3 4}: area 12.00, perimeter 14.00
main.Circle {1}: area 3.14, perimeter 6.28
main.Rect {10 0.5}: area 5.00, perimeter 21.00
total area 20.14
```

`Rect` and `Circle` never mention `Shape`. They have the methods, so they are `Shape`s. This is
**implicit satisfaction**, sometimes called structural typing. Why?

- **Decoupling.** The package defining `Rect` needn't import the one defining `Shape`; neither
  knows the other.
- **Retroactive interfaces.** You can write an interface that *someone else's* types already
  satisfy, including standard library types. In the project, a `*regexp.Regexp` becomes a `seek`
  `Matcher` without a line of glue code.
- **Small interfaces.** Satisfying an interface is free, so interfaces tend to have one or two
  methods, and types satisfy many at once.

A type missing a method doesn't satisfy the interface, and the compiler names the method:

```console
./main.go:15:16: cannot use Square{…} (value of struct type Square) as Shape value in variable declaration: Square does not implement Shape (missing method Perimeter)
```

## Small interfaces in the standard library

Go's most important interfaces have one method each:

| Interface | Method | Meaning |
|-----------|--------|---------|
| `error` | `Error() string` | a failure with a message ([Chapter 9](09-errors.md)) |
| `fmt.Stringer` | `String() string` | how to print me |
| `io.Reader` | `Read(p []byte) (n int, err error)` | a source of bytes |
| `io.Writer` | `Write(p []byte) (n int, err error)` | a destination for bytes |

By convention, a one-method interface is named after its method plus `-er`: `Reader`, `Writer`,
`Stringer`, `Closer`.

### `fmt.Stringer`

The `fmt` functions check whether a value has a `String` method and, if so, print its result, as
[Chapter 2](02-values-and-types.md) promised for enum-like constants:

`examples/ch10/stringer/main.go`

```go
// fmt.Stringer controls how fmt prints your type.
package main

import "fmt"

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

// String makes Weekday a fmt.Stringer:
//
//	type Stringer interface { String() string }
func (d Weekday) String() string {
	switch d {
	case Sunday:
		return "Sunday"
	case Monday:
		return "Monday"
	case Tuesday:
		return "Tuesday"
	}
	return fmt.Sprintf("Weekday(%d)", int(d))
}

func main() {
	fmt.Println(Monday)
	fmt.Printf("%v %s %d\n", Tuesday, Tuesday, Tuesday) // %d still prints the number
	fmt.Println([]Weekday{Sunday, Tuesday, 9})
}
```

```console
$ go run ./ch10/stringer
Monday
Tuesday Tuesday 2
[Sunday Tuesday Weekday(9)]
```

The fallback in `String` converts `d` to `int` before formatting; `fmt.Sprintf("%v", d)` would call
`d.String()` again, forever.

### `io.Writer` and `io.Reader`

`fmt.Fprintf` can print to the terminal, a file, a buffer, or a network connection because it takes
an `io.Writer`, and all of those have a `Write` method. Your own types can join in:

`examples/ch10/io/main.go`

```go
// io.Reader and io.Writer: one method each, used everywhere.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// greet writes to any io.Writer: a file, the terminal, a buffer, a
// network connection, or a type you write yourself.
func greet(w io.Writer, name string) {
	fmt.Fprintf(w, "hello, %s\n", name)
}

// countingWriter counts the bytes written to it and passes them on.
type countingWriter struct {
	w io.Writer
	n int
}

// Write implements io.Writer.
func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
}

func main() {
	greet(os.Stdout, "terminal")

	var b strings.Builder // *strings.Builder has a Write method too
	greet(&b, "builder")
	fmt.Printf("%q\n", b.String())

	cw := &countingWriter{w: os.Stdout}
	greet(cw, "counter")
	fmt.Println(cw.n, "bytes")

	// strings.NewReader returns an io.Reader. io.Copy moves bytes from
	// any Reader to any Writer.
	r := strings.NewReader("copied from a reader\n")
	n, err := io.Copy(cw, r)
	fmt.Println(n, err, cw.n)
}
```

```console
$ go run ./ch10/io
hello, terminal
"hello, builder\n"
hello, counter
15 bytes
copied from a reader
21 <nil> 36
```

- `os.Stdout` is an `*os.File`, `&b` a `*strings.Builder`, `cw` a `*countingWriter`. `greet`
  doesn't care; it only calls `Write`.
- `countingWriter` holds another `io.Writer` and adds behavior around it. Wrapping one interface
  value in another is a common Go pattern.
- `io.Copy` connects any `io.Reader` to any `io.Writer`. [Chapter 11](11-io-and-files.md) covers
  readers and writers in depth.

## Interface values

An interface value is a pair: a **dynamic type** and a **dynamic value**. Assigning `Rect{3, 4}` to
a `Shape` stores (type `Rect`, value `{3 4}`). `s.Area()` looks up `Area` on the stored type and
calls it. `%T` prints the dynamic type.

An interface's zero value is `nil`: no type, no value. The trap: an interface holding a `nil`
pointer is **not** `nil`, because it has a type:

`examples/ch10/values/main.go`

```go
// An interface value holds a dynamic type and a value.
package main

import (
	"fmt"
	"strings"
)

type MyErr struct{}

func (*MyErr) Error() string { return "my error" }

// mayFail returns a nil *MyErr as an error. That error is not nil!
func mayFail() error {
	var p *MyErr // nil pointer
	return p     // interface value: (type *MyErr, value nil)
}

func main() {
	var s fmt.Stringer // nil interface: no type, no value
	fmt.Printf("%T %v %v\n", s, s, s == nil)

	var b *strings.Builder // nil pointer...
	s = b                  // ...stored in an interface
	fmt.Printf("%T %v\n", s, s == nil)

	err := mayFail()
	fmt.Printf("%T %v\n", err, err == nil)

	// Interface values compare equal if both type and value are equal.
	var x, y any = 1, 1
	var z any = int64(1)
	fmt.Println(x == y, x == z)
}
```

```console
$ go run ./ch10/values
<nil> <nil> true
*strings.Builder false
*main.MyErr false
true false
```

| Interface value | Dynamic type | Dynamic value | `== nil`? |
|-----------------|--------------|---------------|-----------|
| `var s fmt.Stringer` | none | none | `true` |
| `s = (*strings.Builder)(nil)` | `*strings.Builder` | `nil` | `false` |
| `mayFail()` | `*MyErr` | `nil` | `false` |

`mayFail` is the classic bug: the caller's `if err != nil` sees an error and reports a failure that
never happened. Rule: **when a function returns an interface, return a literal `nil` for
"nothing"**, never a typed `nil` pointer. That's why `parse` in [Chapter 9](09-errors.md) returned
`nil`, not a `*SyntaxError` variable.

Interface values are equal under `==` if both dynamic types and dynamic values are equal. `1` (an
`int`) and `int64(1)` differ in type, so `x == z` is `false`. Comparing interfaces that hold
uncomparable values, such as slices, panics.

## `any`: the empty interface

`interface{}` has no methods, so every type satisfies it. Go spells it `any`:

`examples/ch10/any/main.go`

```go
// any is the empty interface: every type satisfies it.
package main

import "fmt"

func main() {
	var v any // same as interface{}
	fmt.Printf("%v %T\n", v, v)

	v = 42
	fmt.Printf("%v %T\n", v, v)
	v = "hello"
	fmt.Printf("%v %T\n", v, v)
	v = []int{1, 2}
	fmt.Printf("%v %T\n", v, v)

	// fmt.Println takes ...any, which is why it accepts anything.
	things := []any{1, "two", 3.0, nil, struct{ X int }{4}}
	fmt.Println(things...)
}
```

```console
$ go run ./ch10/any
<nil> <nil>
42 int
hello string
[1 2] []int
1 two 3 <nil> {4}
```

`any` gives up static type checking: the compiler doesn't know what's inside, so you can't use the
value until you get it back out (next section). Use `any` when a function truly accepts every type,
like `fmt.Println`. For "any of several types" with type safety, use generics
([Chapter 13](13-generics.md)).

## Type assertions

A **type assertion** `x.(T)` gets the concrete value out of an interface value `x`:

`examples/ch10/assert/main.go`

```go
// A type assertion gets the concrete value out of an interface.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var v any = "gopher"

	s := v.(string) // v holds a string, so this works
	fmt.Println(s, len(s))

	n, ok := v.(int) // comma-ok: no panic, ok reports success
	fmt.Println(n, ok)

	// Asserting to an interface asks "does it have these methods?"
	var w io.Writer = os.Stdout
	if f, ok := w.(*os.File); ok {
		fmt.Println("a file:", f.Name())
	}
	w = &strings.Builder{}
	_, isStringer := w.(fmt.Stringer)
	fmt.Println("Builder is a Stringer:", isStringer)

	fmt.Println(v.(int)) // one-result form panics on failure
}
```

```console
$ go run ./ch10/assert
gopher 6
0 false
a file: /dev/stdout
Builder is a Stringer: true
panic: interface conversion: interface {} is string, not int

goroutine 1 [running]:
main.main()
	/path/to/examples/ch10/assert/main.go:29 +0x236
exit status 2
```

(The path, hex offset, and file name `/dev/stdout` may differ on your system.)

- `x.(T)` with a concrete `T` succeeds if the dynamic type is exactly `T`.
- `x.(I)` with an interface `I` succeeds if the dynamic type has `I`'s methods, and the result is
  an `I`. This is how code discovers optional abilities: "is this writer also a `Stringer`?".
- The two-result form `v, ok := x.(T)` never panics; on failure `v` is `T`'s zero value. Use it
  unless failure is a bug.

The compiler rejects assertions that can never succeed, and assertions on non-interface values:

```console
./main.go:12:14: impossible type assertion: w.(string)
	string does not implement io.Writer (missing method Write)
./main.go:12:26: invalid operation: n (variable of type int) is not an interface
```

## Type switches

A **type switch** is a sequence of type assertions in `switch` form:

`examples/ch10/type-switch/main.go`

```go
// A type switch picks a case by the dynamic type of an interface value.
package main

import (
	"errors"
	"fmt"
)

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("an integer: %v", x) // x is any here
	case string:
		return fmt.Sprintf("a string of %d bytes", len(x)) // x is a string
	case error:
		return "an error: " + x.Error()
	case fmt.Stringer:
		return "a Stringer: " + x.String()
	default:
		return fmt.Sprintf("something else: %T", x)
	}
}

type celsius float64

func (c celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func main() {
	for _, v := range []any{nil, 7, int64(8), "héllo", errors.New("boom"), celsius(21.5), 2.5} {
		fmt.Println(describe(v))
	}
}
```

```console
$ go run ./ch10/type-switch
nil
an integer: 7
an integer: 8
a string of 6 bytes
an error: boom
a Stringer: 21.5°C
something else: float64
```

- `v.(type)` is allowed only in a type switch.
- In a single-type case, `x` has that type: a `string` in `case string`, an `error` in
  `case error`. In a multi-type case and in `default`, `x` keeps the original type (`any`).
- Cases are tried in order. Interface cases match any type with the methods, so put more specific
  cases first.
- `case nil` matches a nil interface.

Type switches belong at the edges: decoding data, formatting values, handling errors. If you switch
on types all over your code, an interface method is usually better: each type brings its own
behavior, and adding a type doesn't mean editing every switch.

## Embedding interfaces

An interface can include other interfaces, getting all their methods:

`examples/ch10/embedding/main.go`

```go
// Interfaces can embed other interfaces.
package main

import "fmt"

type Getter interface {
	Get(key string) (string, bool)
}

type Setter interface {
	Set(key, value string)
}

// A Store has both methods: the union of Getter and Setter.
type Store interface {
	Getter
	Setter
}

type mapStore map[string]string

func (m mapStore) Get(k string) (string, bool) {
	v, ok := m[k]
	return v, ok
}

func (m mapStore) Set(k, v string) {
	m[k] = v
}

// lookup needs only Get, so it asks for only a Getter.
func lookup(g Getter, key string) string {
	if v, ok := g.Get(key); ok {
		return v
	}
	return "(none)"
}

func main() {
	var s Store = mapStore{}
	s.Set("lang", "go")
	fmt.Println(lookup(s, "lang"), lookup(s, "editor"))
}
```

```console
$ go run ./ch10/embedding
go (none)
```

- A `Store` can be passed where a `Getter` is expected, because it has `Get`.
- The standard library builds bigger interfaces the same way: `io.ReadWriter` is `Reader` plus
  `Writer`; `io.ReadWriteCloser` adds `Closer`.
- `mapStore` is a map type with methods. Any named type can have methods, not only structs.

## Method sets and pointer receivers

[Chapter 7](07-structs-methods-pointers.md) showed that a pointer-receiver method belongs to `*T`,
not `T`. Interfaces make this matter:

| Type | Methods it has (its method set) |
|------|---------------------------------|
| `T` | methods with receiver `T` |
| `*T` | methods with receiver `T` **and** `*T` |

`examples/ch10/method-sets/main.go`

```go
// Pointer-receiver methods belong to *T, not T.
package main

import "fmt"

type Incrementer interface {
	Inc()
}

type Counter struct {
	n int
}

func (c *Counter) Inc() { c.n++ }

// Compile-time check: the build fails if *Counter stops implementing
// Incrementer. It costs nothing at run time.
var _ Incrementer = (*Counter)(nil)

func bump(i Incrementer, times int) {
	for range times {
		i.Inc()
	}
}

func main() {
	c := Counter{}
	bump(&c, 3) // pass a pointer: *Counter has Inc
	fmt.Println(c.n)
}
```

```console
$ go run ./ch10/method-sets
3
```

Passing `c` instead of `&c`:

```console
./main.go:28:7: cannot use c (variable of struct type Counter) as Incrementer value in argument to bump: Counter does not implement Incrementer (method Inc has pointer receiver)
```

Why? An interface stores a *copy* of a `Counter`. If `Inc` ran on that copy, your `c` would never
change, a silent bug, so Go forbids it. Store a pointer, and `Inc` updates the original.

### Compile-time interface checks

```go
var _ Incrementer = (*Counter)(nil)
```

This converts `nil` to a `*Counter` and assigns it to a blank variable of type `Incrementer`.
Nothing is stored and nothing runs. But if `*Counter` stops satisfying the interface (say someone
renames `Inc` to `Increment`), the build fails right here, with a clear message:

```console
./main.go:18:21: cannot use (*Counter)(nil) (value of type *Counter) as Incrementer value in variable declaration: *Counter does not implement Incrementer (missing method Inc)
```

Without the check, the error would show up only where a `*Counter` is used as an `Incrementer`,
maybe in another package. Add the check when a type exists to satisfy an interface. For a value
type, write `var _ I = T{}`.

## Designing with interfaces

Two guidelines shape most Go APIs:

**Accept interfaces, return concrete types.** A parameter should be the smallest interface the
function needs, like `greet(w io.Writer)` or `lookup(g Getter)`, so callers can pass anything that
fits. A result should usually be concrete, such as `*strings.Builder` or `*os.File`, so callers get
all its methods. Return an interface when the concrete type can vary (as with errors) or hiding it
is the point.

**Define interfaces where they're used.** In Java, the package providing a type declares the
interfaces it implements. In Go, the package that *consumes* a behavior declares the interface,
listing only the methods it calls: `lookup` needs only `Get`, so it asks for a `Getter`. Don't
create an interface "just in case" next to its only implementation. Write the concrete type, and add
the interface when a second implementation or a test needs it.

## `sort.Interface`: a classic example

`sort.Sort` sorts anything that implements three methods:

`examples/ch10/sort/main.go`

```go
// sort.Interface: the classic three-method interface.
package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// byLength sorts strings by length. It implements sort.Interface:
// Len, Less, and Swap.
type byLength []string

func (s byLength) Len() int           { return len(s) }
func (s byLength) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s byLength) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func main() {
	words := []string{"banana", "fig", "apple", "kiwi"}
	sort.Sort(byLength(words))
	fmt.Println(words)

	// sort.Reverse wraps any sort.Interface and flips Less.
	sort.Sort(sort.Reverse(byLength(words)))
	fmt.Println(words)

	// Newer code usually passes a comparison function instead.
	slices.SortFunc(words, func(a, b string) int { return strings.Compare(a, b) })
	fmt.Println(words)
}
```

```console
$ go run ./ch10/sort
[fig kiwi apple banana]
[banana apple kiwi fig]
[apple banana fig kiwi]
```

- `byLength(words)` is a conversion, not a copy. The new slice header shares the backing array, so
  sorting it sorts `words`.
- `sort.Reverse` returns another `sort.Interface` whose `Less` calls the original with arguments
  swapped. It wraps an interface value, like `countingWriter`.
- `sort.Interface` predates generics. For slices, `slices.Sort` and `slices.SortFunc`
  ([Chapter 5](05-collections.md)) are simpler and faster; the interface still works for any
  indexed collection.

## Reflection, briefly

The `reflect` package inspects an interface value's dynamic type and value at run time:
`reflect.TypeOf` returns the type, `reflect.ValueOf` the value.

`examples/ch10/reflect/main.go`

```go
// reflect inspects types and values at run time.
package main

import (
	"fmt"
	"reflect"
)

type Config struct {
	Name    string `env:"APP_NAME"`
	Port    int    `env:"APP_PORT"`
	Verbose bool
}

func main() {
	for _, v := range []any{42, "hi", 3.5, []int{1}, Config{}} {
		t := reflect.TypeOf(v)
		fmt.Printf("%-12v kind %v\n", t, t.Kind())
	}

	// Walk a struct's fields: names, types, values, and tags.
	c := Config{Name: "seek", Port: 8080}
	t := reflect.TypeOf(c)
	val := reflect.ValueOf(c)
	for i := range t.NumField() {
		f := t.Field(i)
		fmt.Printf("%s %v = %v  env=%q\n", f.Name, f.Type, val.Field(i), f.Tag.Get("env"))
	}
}
```

```console
$ go run ./ch10/reflect
int          kind int
string       kind string
float64      kind float64
[]int        kind slice
main.Config  kind struct
Name string = seek  env="APP_NAME"
Port int = 8080  env="APP_PORT"
Verbose bool = false  env=""
```

- A **type** can be named (`main.Config`); its **kind** is the underlying category (`struct`,
  `slice`, `int`).
- This is how `fmt` prints any value and how `encoding/json` reads struct tags
  ([Chapter 15](15-json-and-encoding.md)).

**When not to use it:** almost always. Reflection is slow and verbose, and moves type errors from
compile time to run time, often as panics. Use it in general-purpose libraries (encoders, printers,
ORMs). In application code, interfaces, type switches, and generics do the job with type checking.

## The `regexp` package, briefly

`seek` gets a `-E` option for regular expressions in this chapter, so here is the core of `regexp`:

`examples/ch10/regexp/main.go`

```go
// A quick tour of the regexp package.
package main

import (
	"fmt"
	"regexp"
)

// MustCompile panics on a bad pattern. Use it for patterns fixed in the
// source; use Compile for patterns that come from users.
var date = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)

func main() {
	text := "released 2024-02-06, patched 2024-03-05"

	fmt.Println(date.MatchString(text))
	fmt.Println(date.FindString(text))
	fmt.Println(date.FindAllString(text, -1))
	fmt.Println(date.FindAllStringIndex(text, -1)) // byte offsets
	fmt.Println(date.FindStringSubmatch(text))     // whole match, then groups
	fmt.Println(date.ReplaceAllString(text, "$3/$2/$1"))

	fold := regexp.MustCompile(`(?i)go+`) // (?i) ignores case
	fmt.Println(fold.FindAllString("Go GOOO gopher", -1))

	_, err := regexp.Compile(`(unclosed`)
	fmt.Println(err)
}
```

```console
$ go run ./ch10/regexp
true
2024-02-06
[2024-02-06 2024-03-05]
[[9 19] [29 39]]
[2024-02-06 2024 02 06]
released 06/02/2024, patched 05/03/2024
[Go GOOO go]
error parsing regexp: missing closing ): `(unclosed`
```

- Write patterns as raw strings (backquotes), so `\d` needs no extra escaping.
- `regexp.Compile` returns `(*regexp.Regexp, error)`. `MustCompile` is the panicking `Must` form
  from [Chapter 9](09-errors.md).
- Method names follow a pattern: `Find`, then `All` (every match, up to `n`; `-1` means no limit),
  then `String` (string input, not `[]byte`), then `Submatch` (include groups) or `Index` (byte
  offsets instead of text).
- Go uses RE2 syntax: no backreferences or lookarounds, but matching always takes time linear in
  the input size, so a user-supplied pattern can't hang your program.
- A compiled `*regexp.Regexp` is safe to use from many goroutines at once.

Note `FindAllStringIndex(s string, n int) [][]int`. `seek`'s `Searcher.FindAll`
([Chapter 7](07-structs-methods-pointers.md)) has the same shape: byte offsets `{start, end}` of up
to `n` matches. That's no accident.

## Building seek: `Matcher`, `Printer`, and `-E`

`seek`'s search and output code is hard-wired:

- `Searcher` holds a pattern and an `IgnoreCase` flag, and `findAll` has an `if` for each kind of
  matching. Regular expressions would mean a third branch everywhere.
- `Printer` prints to standard output with `fmt.Printf`. Nothing else can see the output, so it
  can't be tested and no other format can be added.

This step introduces two interfaces:

| Interface | Package | Implementations |
|-----------|---------|-----------------|
| `search.Matcher` | `search` (where it's used) | `literal`, `foldLiteral`, `*regexp.Regexp` |
| `output.Printer` | `output` | `*TextPrinter` (Chapter 15 adds JSON) |

and one option: `-E` treats `PATTERN` as a regular expression.

Steps: add `matcher.go`, then update `search.go`, `output.go`, and `main.go`.

### Step 1: `internal/search/matcher.go`

The central trick: `Matcher`'s one method has **exactly the name and signature of
`(*regexp.Regexp).FindAllStringIndex`**, so a compiled regular expression already is a `Matcher`.
`seek` only writes the two plain-text matchers.

Create `internal/search/matcher.go`. `indexFold` and `prefixFold` move here from `search.go`
unchanged:

`internal/search/matcher.go`

```go
package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// A Matcher finds the occurrences of a pattern in a line of text.
//
// The method has the same name and signature as the one on
// *regexp.Regexp, so a compiled regular expression is a Matcher without
// any extra code: Go interfaces are satisfied implicitly.
type Matcher interface {
	// FindAllStringIndex returns the byte offsets {start, end} of up to n
	// successive non-overlapping matches in s (all of them if n < 0).
	// It returns nil if there is no match.
	FindAllStringIndex(s string, n int) [][]int
}

// Compile-time checks that each type implements Matcher.
var (
	_ Matcher = literal("")
	_ Matcher = foldLiteral("")
	_ Matcher = (*regexp.Regexp)(nil)
)

// NewMatcher returns a Matcher for pattern. If regex is true, pattern is
// a regular expression in the syntax of the regexp package; otherwise it
// is plain text.
func NewMatcher(pattern string, ignoreCase, regex bool) (Matcher, error) {
	switch {
	case regex:
		if ignoreCase {
			pattern = "(?i)" + pattern // the regexp flag for ignoring case
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		return re, nil
	case ignoreCase:
		return foldLiteral(pattern), nil
	default:
		return literal(pattern), nil
	}
}

// literal matches a fixed string exactly.
type literal string

func (l literal) FindAllStringIndex(s string, n int) [][]int {
	return findAll(s, n, func(s string) (int, int) {
		i := strings.Index(s, string(l))
		return i, i + len(l)
	})
}

// foldLiteral matches a fixed string, ignoring case.
type foldLiteral string

func (l foldLiteral) FindAllStringIndex(s string, n int) [][]int {
	return findAll(s, n, func(s string) (int, int) {
		return indexFold(s, string(l))
	})
}

// findAll calls index repeatedly to collect up to n non-overlapping
// matches in s (all of them if n < 0). index reports the first match in
// the string it is given, or a negative start if there is none.
func findAll(s string, n int, index func(string) (start, end int)) [][]int {
	var spans [][]int
	for pos := 0; n < 0 || len(spans) < n; {
		start, end := index(s[pos:])
		if start < 0 {
			break
		}
		spans = append(spans, []int{pos + start, pos + end})
		if start == end {
			break // an empty pattern: one match is enough
		}
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

Points to notice:

- **Methods on a string type.** `literal` and `foldLiteral` are named types with underlying type
  `string`. The value *is* the pattern; no struct needed. `literal("fox")` is a conversion that
  makes a `Matcher`.
- **One loop, three users.** The old `findAll` had an `if ignoreCase` inside its loop. Now the loop
  takes an `index` function, and each matcher passes its own. The empty-pattern special case became
  `if start == end { break }`; otherwise an empty match at `pos` would repeat forever.
- **`NewMatcher` returns an interface,** a valid exception to "return concrete types": the concrete
  type depends on the arguments. On failure it returns a literal `nil`, never a typed `nil` pointer.
- With `-i -E`, `NewMatcher` prefixes the pattern with the regexp flag `(?i)`.
- The compile-time checks list all three implementations, including `*regexp.Regexp`. If a future
  edit changed `Matcher`'s method, the build would point here.

### Step 2: `internal/search/search.go`

`Searcher` no longer knows about patterns; it holds a `Matcher` and asks it about each line.
Replace the `Searcher` type and `selects`, delete `FindAll`, `findAll`, `indexFold`, and
`prefixFold`, and drop the `unicode/utf8` import. The whole file is now:

`internal/search/search.go`

```go
// Package search finds lines of text that match a pattern.
package search

import (
	"os"
	"strings"
)

// A Match is one selected line of a file.
type Match struct {
	Path string // the file the line came from
	Line int    // line number, counting from 1
	Text string // the line itself, without its newline
}

// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Matcher  Matcher // decides which lines match; must not be nil
	Invert   bool    // select the lines that do NOT match
	MaxCount int     // stop after this many matches per file; 0 means no limit
}

// selects reports whether line belongs in the results.
func (s *Searcher) selects(line string) bool {
	found := s.Matcher.FindAllStringIndex(line, 1) != nil
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

`Matcher Matcher` is a field named after its type, which is normal in Go. The zero `Searcher` is no
longer useful (calling a method on a `nil` interface panics), so the doc comment says the field must
not be `nil`.

### Step 3: `internal/output/output.go`

`Printer` becomes an interface, and the old struct becomes its first implementation, `TextPrinter`.
Three more changes:

- `TextPrinter` writes to `W`, an `io.Writer`, instead of standard output.
- Writes can fail (a closed pipe, a full disk), so `PrintFile` and `printMatch` return
  `fmt.Fprintf`'s error.
- The `Find` function field becomes `Highlight`, a `search.Matcher`.

`internal/output/output.go`

```go
// Package output formats search results.
package output

import (
	"fmt"
	"io"
	"strings"

	"seek/internal/search"
)

// A Printer writes the results of a search.
type Printer interface {
	// PrintFile prints the matches found in the file at path.
	PrintFile(path string, matches []search.Match) error
}

// TextPrinter prints results as plain text, one line per match:
// path:text, or path:line:text if LineNumbers is set.
type TextPrinter struct {
	W           io.Writer // where to print
	LineNumbers bool      // show line numbers
	Count       bool      // print only the number of matches per file
	FilesOnly   bool      // print only the names of files that have matches

	// Highlight, if not nil, finds the text to color in each line.
	Highlight search.Matcher
}

var _ Printer = (*TextPrinter)(nil)

// PrintFile implements Printer.
func (p *TextPrinter) PrintFile(path string, matches []search.Match) error {
	switch {
	case p.FilesOnly:
		if len(matches) == 0 {
			return nil
		}
		_, err := fmt.Fprintln(p.W, path)
		return err
	case p.Count:
		_, err := fmt.Fprintf(p.W, "%s:%d\n", path, len(matches))
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
		_, err = fmt.Fprintf(p.W, "%s:%d:%s\n", m.Path, m.Line, text)
	} else {
		_, err = fmt.Fprintf(p.W, "%s:%s\n", m.Path, text)
	}
	return err
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

- `fmt.Fprintf` returns `(n int, err error)`; `_, err =` keeps only the error.
- `var _ Printer = (*TextPrinter)(nil)` checks the implementation at compile time.
- `highlight` didn't change.

`Highlight` is an interface field, and `nil` means "no color". `p.Highlight != nil` is safe only
because `main` either assigns a real matcher or leaves the field alone. Assigning a `nil`
`*regexp.Regexp` variable would store (type `*regexp.Regexp`, value `nil`): the check would pass and
the call would panic. Never store a typed `nil` in an interface that someone tests against `nil`.

### Step 4: `cmd/seek/main.go`

`main.go` gets `-E`, builds the matcher, and passes writers around instead of using `os.Stdout` and
`os.Stderr` directly. Add `"io"` to the imports:

```go
import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"

	"seek/internal/output"
	"seek/internal/search"
)
```

`main` passes the real standard output and standard error to `run`:

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

Add a `regex` field to `options`, before `version`:

```go
	color       bool // --color
	regex       bool // -E
	version     bool // --version
}
```

and a case for it in `parseArgs`, after `--color`:

```go
		case "--color":
			opts.color = true
		case "-E":
			opts.regex = true
```

`report` takes the writer to print to:

```go
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
		fmt.Fprintln(stderr, usage)
	}
}
```

`run` changes the most. Replace it:

```go
// run is the whole program. It writes results to stdout and messages to
// stderr, and returns the exit status. Taking io.Writers instead of using
// os.Stdout and os.Stderr directly lets tests capture the output.
func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseArgs(args)
	if errors.Is(err, errHelp) {
		fmt.Fprintln(stdout, usage)
		return exitOK
	}
	if err != nil {
		report(stderr, err)
		return exitError
	}
	if opts.version {
		fmt.Fprintln(stdout, name, version)
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
	}
	if opts.filesOnly {
		s.MaxCount = 1 // one match is enough to print the file name
	}

	// p's static type is the Printer interface; its dynamic type is
	// *output.TextPrinter. Chapter 15 adds a second implementation.
	var p output.Printer
	tp := &output.TextPrinter{
		W:           stdout,
		LineNumbers: opts.lineNumbers,
		Count:       opts.count,
		FilesOnly:   opts.filesOnly,
	}
	if opts.color {
		tp.Highlight = matcher
	}
	p = tp

	// A bad file shouldn't stop the search: report it, remember that
	// something went wrong, and carry on with the next one.
	total, hadError := 0, false
	for _, path := range opts.paths {
		matches, err := s.SearchFile(path)
		if err != nil {
			report(stderr, err)
			hadError = true
			continue
		}
		total += len(matches)
		if err := p.PrintFile(path, matches); err != nil {
			report(stderr, err)
			return exitError
		}
	}

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

What changed in `run`:

- **It accepts interfaces.** `stdout` and `stderr` are `io.Writer`s. `main` passes the real streams;
  a test ([Chapter 19](19-testing.md)) can pass buffers and check what was written.
- **One matcher, two jobs.** `NewMatcher` builds the matcher once; the `Searcher` uses it to select
  lines and the printer to highlight them. A bad `-E` pattern is reported like any other error, with
  status 2.
- **`p` has an interface type.** `tp` is configured as a `*TextPrinter`, then stored in `p`. The
  loop only knows that `p` can `PrintFile`. A write error stops the search: if standard output is
  gone, there's no point continuing.

`parseArgs`, the error types, and `usageErrorf` don't change.

### Checkpoint

```console
$ go run ./cmd/seek -E -n 'fox|dog' testdata/poem.txt
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:2:jumps over the lazy dog.
testdata/poem.txt:3:The dog sleeps; the fox runs.
$ go run ./cmd/seek -E -i '^the' testdata/poem.txt testdata/notes.md
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
$ go run ./cmd/seek -E '[0-9]+' testdata/notes.md
testdata/notes.md:- café opens at 9
$ go run ./cmd/seek --color -E 'o[a-z]' testdata/poem.txt | cat -v
testdata/poem.txt:The quick br^[[1;31mow^[[0mn f^[[1;31mox^[[0m
testdata/poem.txt:jumps ^[[1;31mov^[[0mer the lazy d^[[1;31mog^[[0m.
testdata/poem.txt:The d^[[1;31mog^[[0m sleeps; the f^[[1;31mox^[[0m runs.
testdata/poem.txt:F^[[1;31mox^[[0mes are quick and clever.
$ go run ./cmd/seek --color -i fox testdata/notes.md | cat -v
testdata/notes.md:- feed the ^[[1;31mFOX^[[0m
$ go run ./cmd/seek -E '(fox' testdata/poem.txt
seek: invalid pattern: error parsing regexp: missing closing ): `(fox`
exit status 2
```

- `^the` with `-i` matches only at the start of a line. In `notes.md`, `the` appears only mid-line
  (`- feed the FOX`), so there's no match.
- With `--color`, the regexp highlights each match. `cat -v` shows the escape byte as `^[`.
- The error message comes from `regexp`, wrapped by `NewMatcher` (`invalid pattern: %w`) and printed
  by `report`.

All earlier checkpoints still give the same output. The complete project at this point is in
`project/ch10/`.

## Summary

- An interface is a set of methods. A type satisfies it by having them; there's no `implements`.
  This decouples packages and lets you write interfaces for existing types.
- Keep interfaces small. One-method interfaces like `io.Reader`, `io.Writer`, `fmt.Stringer`, and
  `error` are everywhere.
- An interface value is a (type, value) pair, `nil` only if both are unset. One holding a `nil`
  pointer is not `nil`, so return a literal `nil` for "no error".
- `any` holds any value, but you must get it back out with a type assertion or type switch.
- `x.(T)` asserts; `v, ok := x.(T)` doesn't panic. Type switches branch on the dynamic type.
- Interfaces embed interfaces. Any named type can have methods.
- `*T`'s method set includes pointer-receiver methods; `T`'s doesn't. Store `&v` in the interface.
- `var _ I = (*T)(nil)` checks satisfaction at compile time.
- Accept interfaces, return concrete types. Define interfaces where they're used.
- `sort.Interface` is the classic three-method interface. `reflect` exists; rarely use it.
- `regexp` uses RE2 syntax with linear-time matching. Its `FindAllStringIndex` method makes it a
  `seek` `Matcher`.

## Exercises

1. Add a `Square` type with a `Side` field to the shapes example. Make it a `Shape`, and also a
   `fmt.Stringer` that prints `square(2)`. Add compile-time checks for both interfaces.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	R float64
}

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

type Square struct {
	Side float64
}

func (s Square) Area() float64      { return s.Side * s.Side }
func (s Square) Perimeter() float64 { return 4 * s.Side }
func (s Square) String() string     { return fmt.Sprintf("square(%g)", s.Side) }

var (
	_ Shape        = Square{}
	_ fmt.Stringer = Square{}
)

func main() {
	for _, s := range []Shape{Circle{1}, Square{2}} {
		fmt.Printf("%v: area %.2f\n", s, s.Area())
	}
}
```

```console
$ go run .
{1}: area 3.14
square(2): area 4.00
```

`Square` has value receivers, so the checks use a `Square{}` value. `fmt` finds `String` through
the `Shape` interface value, because the dynamic type has it.

</details>

2. Write an `upperWriter` that wraps another `io.Writer` and upper-cases everything written
   through it. Use it with `fmt.Fprintf` and `io.WriteString`.

<details><summary>Solution</summary>

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// upperWriter upper-cases text before passing it on to w.
type upperWriter struct {
	w io.Writer
}

func (u upperWriter) Write(p []byte) (int, error) {
	return u.w.Write(bytes.ToUpper(p))
}

func main() {
	w := upperWriter{os.Stdout}
	fmt.Fprintf(w, "hello, %s! %d gophers\n", "world", 3)
	io.WriteString(w, "written by io.WriteString\n")
}
```

```console
$ go run .
HELLO, WORLD! 3 GOPHERS
WRITTEN BY IO.WRITESTRING
```

`fmt.Fprintf` and `io.WriteString` know nothing about `upperWriter`; they call `Write`.

</details>

3. Write `sum(values ...any) (float64, error)` that adds `int`s, `float64`s, and numeric strings
   (use `strconv.ParseFloat`). Return an error for strings that aren't numbers and for any other
   type.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"strconv"
)

// sum adds up numbers given as ints, float64s, or numeric strings.
func sum(values ...any) (float64, error) {
	total := 0.0
	for i, v := range values {
		switch x := v.(type) {
		case int:
			total += float64(x)
		case float64:
			total += x
		case string:
			f, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return 0, fmt.Errorf("value %d: %w", i, err)
			}
			total += f
		default:
			return 0, fmt.Errorf("value %d: unsupported type %T", i, x)
		}
	}
	return total, nil
}

func main() {
	fmt.Println(sum(1, 2.5, "3.25"))
	fmt.Println(sum(1, "x"))
	fmt.Println(sum(1, true))
}
```

```console
$ go run .
6.75 <nil>
0 value 1: strconv.ParseFloat: parsing "x": invalid syntax
0 value 1: unsupported type bool
```

</details>

4. This program prints `failed: <nil>`. Why? Fix it.

```go
package main

import "fmt"

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string { return e.Field + " is invalid" }

func check(name string) error {
	var err *ValidationError
	if name == "" {
		err = &ValidationError{Field: "name"}
	}
	return err
}

func main() {
	if err := check("gopher"); err != nil {
		fmt.Println("failed:", err)
	} else {
		fmt.Println("ok")
	}
}
```

<details><summary>Solution</summary>

`check("gopher")` returns `err`, a `nil` `*ValidationError`. As an `error`, that's the pair
(type `*ValidationError`, value `nil`), which is not a `nil` interface, so `err != nil` is true.
(`fmt` prints `<nil>` because the pointer inside is `nil`.)

Return a literal `nil` on success:

```go
func check(name string) error {
	if name == "" {
		return &ValidationError{Field: "name"}
	}
	return nil
}

func main() {
	for _, name := range []string{"gopher", ""} {
		if err := check(name); err != nil {
			fmt.Println("failed:", err)
		} else {
			fmt.Println("ok")
		}
	}
}
```

```console
$ go run .
ok
failed: name is invalid
```

</details>

5. Add a `-x` option to `seek`: select only lines that the pattern matches **as a whole**. Write it
   as a `Matcher` that wraps another `Matcher`: `search.WholeLine(m Matcher) Matcher`.

<details><summary>Solution</summary>

Append to `internal/search/matcher.go`:

```go
// WholeLine returns a Matcher that matches a line only if m matches all
// of it. It wraps m: the new Matcher calls m and filters its results.
func WholeLine(m Matcher) Matcher {
	return wholeLine{m}
}

type wholeLine struct {
	m Matcher
}

func (w wholeLine) FindAllStringIndex(s string, n int) [][]int {
	if n == 0 {
		return nil
	}
	for _, span := range w.m.FindAllStringIndex(s, -1) {
		if span[0] == 0 && span[1] == len(s) {
			return [][]int{span}
		}
	}
	return nil
}
```

In `cmd/seek/main.go`, add `wholeLine bool // -x` to `options`, a `case "-x": opts.wholeLine = true`
to `parseArgs`, and wrap the matcher in `run` right after it's built:

```go
	matcher, err := search.NewMatcher(opts.pattern, opts.ignoreCase, opts.regex)
	if err != nil {
		report(stderr, err)
		return exitError
	}
	if opts.wholeLine {
		matcher = search.WholeLine(matcher)
	}
```

```console
$ go run ./cmd/seek -x fox testdata/poem.txt
exit status 1
$ go run ./cmd/seek -x -i -n 'foxes are quick and clever.' testdata/poem.txt
testdata/poem.txt:4:Foxes are quick and clever.
$ go run ./cmd/seek -x -E -i -n 'the .*' testdata/poem.txt
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
$ go run ./cmd/seek -x --color -i 'the quick brown fox' testdata/poem.txt | cat -v
testdata/poem.txt:^[[1;31mThe quick brown fox^[[0m
```

`Searcher` and `TextPrinter` didn't change: they only see a `Matcher`. Highlighting colors the whole
line, because the wrapped matcher does both jobs. With `-E`, the check uses the regexp's leftmost
matches, so for alternations like `fox|fox runs` it can miss a whole-line match; wrapping the
pattern as `^(?:...)$` would be exact. (The snapshots don't include this change.)

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-10--interfaces)
