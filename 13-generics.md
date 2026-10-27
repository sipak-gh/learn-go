# Chapter 13 — Generics

Some code is the same for every type: summing numbers, finding an element, counting things. Without
generics you write it once per type, or use `any` and lose type checking. With **type parameters** you
write it once and the compiler checks every use. You'll learn generic functions and types,
constraints, type inference, and the standard library's generic helpers. Then `seek` gets a generic
`Counter` and a `--stats` option.

Previous: [Chapter 12 — Command-Line Flags and Configuration](12-flags.md). Next: [Chapter 14 —
Iterators and Range-over-Func](14-iterators.md).

## Type parameters

A generic function lists **type parameters** in square brackets before its normal parameters. Each
has a **constraint**: the types it allows.

`examples/ch13/sum/main.go`

```go
package main

import "fmt"

// Without generics you write one function per type.
func SumInts(nums []int) int {
	var total int
	for _, n := range nums {
		total += n
	}
	return total
}

// Number is a constraint: the set of types Sum accepts.
// ~float64 means "float64 or any type whose underlying type is float64".
type Number interface {
	~int | ~int64 | ~float64
}

// Sum works for any type in Number's type set.
func Sum[T Number](nums []T) T {
	var total T // the zero value of T
	for _, n := range nums {
		total += n
	}
	return total
}

type Celsius float64

func main() {
	fmt.Println(SumInts([]int{1, 2, 3}))
	fmt.Println(Sum([]int{1, 2, 3}))         // T inferred as int
	fmt.Println(Sum([]float64{1.5, 2.25}))   // T inferred as float64
	fmt.Println(Sum([]Celsius{20.5, 21}))    // allowed thanks to ~float64
	fmt.Println(Sum[int64](nil))             // nothing to infer from: say it
	fmt.Printf("%T\n", Sum([]Celsius{1, 2})) // the result keeps the type
}
```

```console
$ go run ./ch13/sum
6
6
3.75
41.5
0
main.Celsius
```

- `[T Number]` declares type parameter `T` with constraint `Number`. Inside `Sum`, `T` works like any
  type name.
- `Sum([]int{...})` **instantiates** `Sum` with `T = int`. The compiler **infers** the type argument
  from the arguments, so you rarely write it. With nothing to infer from (`nil`), write it:
  `Sum[int64](nil)`.
- `var total T` is `T`'s zero value: `0` here, `""` for strings, `nil` for pointers. It's the usual
  way to get "nothing" of type `T`.
- The result has type `T`, so `Sum` of `[]Celsius` returns a `Celsius`, not a `float64`.

## Constraints are interfaces

A constraint is an interface. In [Chapter 10](10-interfaces.md) an interface was a set of methods;
here it can also list types. The types it allows form its **type set**.

- `int | int64 | float64` is a **union**: any one of these types.
- `~float64` means any type whose *underlying type* is `float64`, such as `Celsius`. Without `~`, only
  `float64` itself matches.

The body may use only operations that work for every type in the type set. `+` works in `Sum` because
`int`, `int64`, and `float64` all support it.

Wrong types are compile errors with clear messages. Calling `Sum` with strings:

```console
./main.go:18:17: string does not satisfy Number (string missing in ~int | ~int64 | ~float64)
```

Forgetting the `~` and calling `Sum([]Celsius{20.5, 21})`:

```console
./main.go:20:17: Celsius does not satisfy Number (possibly missing ~ for float64 in Number)
```

An interface with a type list works only as a constraint, not as a variable type:

```console
./main.go:8:8: cannot use type Number outside a type constraint: interface contains type constraints
```

### any, comparable, and cmp.Ordered

You rarely write your own constraints. These cover most cases:

| Constraint | Allows | Lets you use |
|---|---|---|
| `any` | every type | assignment, passing around |
| `comparable` | types that support `==` | `==`, `!=`, map keys |
| `cmp.Ordered` | integers, floats, strings (with `~`) | `<`, `<=`, `>`, `>=`, `==` |

`any` is another name for `interface{}`. With `any` you can store and return values, but not compare
or add them. Using `>` on an `any` type parameter:

```console
./main.go:4:5: invalid operation: a > b (type parameter T cannot use operator >)
```

`examples/ch13/constraints/main.go`

```go
package main

import (
	"cmp"
	"fmt"
	"strconv"
)

// Index works for any type that supports == (the comparable constraint).
func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		if v == x {
			return i
		}
	}
	return -1
}

// Max works for any ordered type: integers, floats, strings.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// A constraint can require methods as well as a type set.
type IntStringer interface {
	~int
	String() string
}

type Weekday int

func (d Weekday) String() string {
	return [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}[d]
}

// Label can use both + (from ~int) and String (from the method).
func Label[T IntStringer](v T) string {
	return v.String() + "#" + strconv.Itoa(int(v+1))
}

func main() {
	fmt.Println(Index([]string{"a", "b", "c"}, "c"))
	fmt.Println(Index([]float64{1.5, 2.5}, 9))
	fmt.Println(Max(3, 7), Max("apple", "banana"), Max(2.5, -1))
	fmt.Println(cmp.Compare(1, 2), cmp.Compare("b", "a"), cmp.Compare(3.0, 3.0))
	fmt.Println(Label(Weekday(2)))
}
```

```console
$ go run ./ch13/constraints
2
-1
7 banana 2.5
-1 1 0
Tue#3
```

- In `Index([]float64{1.5, 2.5}, 9)`, the untyped constant `9` becomes a `float64`, since `T` was
  inferred as `float64` from the slice.
- `cmp.Compare(a, b)` returns -1, 0, or +1: the standard building block for sort comparisons.
- A constraint can combine a type set and methods: `T` needs underlying type `int` **and** a `String`
  method.

Slices, maps, and functions aren't `comparable`, so `Index` rejects a `[][]int`:

```console
./main.go:15:19: []int does not satisfy comparable
```

## Generic functions with several type parameters

Type parameters can depend on each other. `Map` turns a `[]T` into a `[]U`:

`examples/ch13/funcs/main.go`

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Map applies f to every element. It has two type parameters.
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// Filter keeps the elements for which keep returns true.
func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce combines the elements into one value, starting from init.
func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6}

	// T and U are inferred from the arguments: int and string.
	strs := Map(nums, strconv.Itoa)
	fmt.Printf("%q\n", strs)

	even := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println(even)

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println(sum)

	// Explicit type arguments are allowed, and sometimes needed.
	lens := Map[string, int](strings.Fields("go is fun"), func(s string) int { return len(s) })
	fmt.Println(lens)
}
```

```console
$ go run ./ch13/funcs
["1" "2" "3" "4" "5" "6"]
[2 4 6]
21
[2 2 3]
```

Inference works from the **arguments**: `Map(nums, strconv.Itoa)` gets `T = int` from `nums` and
`U = string` from `strconv.Itoa`'s result type. It never uses the variable you assign the result to.
A type parameter that appears only in the result must be written:

```go
func Zero[T any]() T {
	var zero T
	return zero
}

fmt.Println(Zero()) // error
```

```console
./main.go:11:18: in call to Zero, cannot infer T (declared at ./main.go:5:11)
```

`Zero[int]()` compiles.

## Generic types

Types can have type parameters too. A `Stack[T]` holds elements of one type `T`:

`examples/ch13/stack/main.go`

```go
package main

import "fmt"

// Stack is a last-in, first-out collection of any element type.
// The zero value is an empty stack, ready to use.
type Stack[T any] struct {
	items []T
}

// Methods use the type parameter from the receiver: Stack[T].
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top element. If the stack is empty it
// returns the zero value of T and false.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// A generic struct with two type parameters.
type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func main() {
	var words Stack[string] // instantiate: T = string
	words.Push("first")
	words.Push("second")
	fmt.Println(words.Len())
	for words.Len() > 0 {
		w, _ := words.Pop()
		fmt.Println(w)
	}
	v, ok := words.Pop()
	fmt.Printf("%q %v\n", v, ok)

	nums := &Stack[int]{}
	nums.Push(42)
	n, ok := nums.Pop()
	fmt.Println(n, ok)

	p := Pair[string, int]{"answer", 42}
	fmt.Printf("%+v %T\n", p, p)
}
```

```console
$ go run ./ch13/stack
2
second
first
"" false
42 true
{Key:answer Val:42} main.Pair[string,int]
```

- `Stack` alone isn't a type; `Stack[string]` and `Stack[int]` are two different types. For types you
  always write the type argument; composite literals don't infer it.
- A method's receiver names the type parameter again: `func (s *Stack[T]) Push(v T)`. Any name works
  (`Stack[E]`), but `T` is conventional.
- `Pop` returns `(T, bool)`, like comma-ok map lookup. The `bool` tells "empty" from a stored zero
  value.

### Methods can't have their own type parameters

A method can use its receiver's type parameters but can't add new ones:

```go
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U] { ... }
```

```console
./main.go:7:23: syntax error: method must have no type parameters
```

The reason is interfaces. Methods satisfy interfaces, so the compiler must know each type's full
method list. A method with its own type parameter would be an unlimited family of methods, and
there's no good way to compile the one a run-time interface check asks for. The rule: **for another
type parameter or a stricter constraint, write a function.**

```go
func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U] { ... }
```

`seek`'s `TopN` uses this rule below.

## Generic type aliases

Go 1.24 added generic type aliases. An alias (with `=`) names an existing type; it creates no new
type:

`examples/ch13/alias/main.go`

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

// Set is a generic type alias: Set[string] is just another name for
// map[string]struct{}. No new type is created.
type Set[T comparable] = map[T]struct{}

func Of[T comparable](items ...T) Set[T] {
	s := make(Set[T])
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

func main() {
	s := Of("go", "zig", "go")
	fmt.Println(len(s), slices.Sorted(maps.Keys(s)))

	// Because it's an alias, a plain map is interchangeable with it.
	var m map[string]struct{} = s
	_, ok := m["zig"]
	fmt.Println(ok)
	fmt.Printf("%T\n", s)
}
```

```console
$ go run ./ch13/alias
2 [go zig]
true
map[string]struct {}
```

`%T` shows the real type: there is no type called `Set`. Aliases mostly help move types between
packages without breaking code. For methods, define a new type (`type Set[T comparable]
struct{...}`), as Exercise 3 does.

## Generics in the standard library

The `slices`, `maps`, and `cmp` packages are generic. You used `slices.Sorted(maps.Keys(m))` in
[Chapter 5](05-collections.md); there's much more:

`examples/ch13/stdgeneric/main.go`

```go
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	nums := []int{5, 2, 8, 2, 9}
	fmt.Println(slices.Max(nums), slices.Min(nums), slices.Contains(nums, 8), slices.Index(nums, 2))
	sorted := slices.Clone(nums)
	slices.Sort(sorted)
	i, found := slices.BinarySearch(sorted, 8)
	fmt.Println(sorted, i, found)
	fmt.Println(slices.Compact(sorted)) // remove adjacent duplicates

	people := []Person{{"Cy", 30}, {"Al", 25}, {"Bo", 30}}
	// Sort by age, oldest first, then by name. cmp.Or returns the first
	// non-zero comparison.
	slices.SortFunc(people, func(a, b Person) int {
		return cmp.Or(
			cmp.Compare(b.Age, a.Age),
			strings.Compare(a.Name, b.Name),
		)
	})
	fmt.Println(people)

	ages := map[string]int{"Al": 25, "Bo": 30}
	clone := maps.Clone(ages)
	clone["Cy"] = 30
	fmt.Println(len(ages), len(clone), maps.Equal(ages, clone))
	fmt.Println(min(3, 1, 2), max("a", "b")) // built-in, also generic
}
```

```console
$ go run ./ch13/stdgeneric
9 2 true 1
[2 2 5 8 9] 3 true
[2 5 8 9]
[{Bo 30} {Cy 30} {Al 25}]
2 3 false
1 b
```

- `slices.SortFunc` takes a comparison returning negative, zero, or positive. Swapping `a` and `b`
  (`cmp.Compare(b.Age, a.Age)`) sorts descending.
- `cmp.Or` returns its first non-zero argument: "compare by age; if equal, by name".
- `slices.Clone` and `maps.Clone` make shallow copies. `slices.Sort` sorts in place.
- The built-in `min` and `max` (Go 1.21) take any number of `cmp.Ordered` arguments.

## Generics or interfaces?

Both let code work with many types, but they solve different problems:

- **Interface**: the code calls *methods* and behaviour differs per type (`io.Writer`, `seek`'s
  `Printer`). The concrete type is chosen at run time.
- **Type parameter**: the code is the same for every type and you want to keep the type: containers
  (`Stack[T]`, `Counter[K]`) and functions over slices and maps (`Map`, `slices.Max`). The compiler
  ensures a `Stack[int]` holds only `int`s, with no type assertions.

The Go team's advice: write normal code first. Add a type parameter when you find yourself writing the
same code for a second or third type.

## Building seek: a generic counter and --stats

`--stats` makes `seek` print, after the results and to standard error: files searched, files
matched, total matches, elapsed time, and the top five files by match count.

To count matches per file, you'll write a small generic package, `internal/stats`. Its `Counter[K]`
counts any comparable key; `seek` uses `Counter[string]` keyed by file path.

### internal/stats/counter.go (new)

```go
// Package stats provides generic counting helpers.
package stats

import (
	"cmp"
	"slices"
)

// A Counter counts occurrences of keys of any comparable type K.
// The zero value is an empty Counter ready to use.
type Counter[K comparable] struct {
	counts map[K]int
}

// Add adds n to the count for key.
func (c *Counter[K]) Add(key K, n int) {
	if c.counts == nil {
		c.counts = make(map[K]int) // writing to a nil map would panic
	}
	c.counts[key] += n
}

// Get returns the count for key, or 0 if key was never added.
func (c *Counter[K]) Get(key K) int {
	return c.counts[key] // reading a nil map is fine
}

// Len returns the number of distinct keys.
func (c *Counter[K]) Len() int {
	return len(c.counts)
}

// Total returns the sum of all counts.
func (c *Counter[K]) Total() int {
	total := 0
	for _, n := range c.counts {
		total += n
	}
	return total
}

// An Entry is a key with its count.
type Entry[K any] struct {
	Key   K
	Count int
}

// TopN returns up to n entries with the highest counts, highest first.
// Equal counts are ordered by key, so the result is always the same.
//
// TopN is a function rather than a method because it needs to sort keys,
// so K must satisfy cmp.Ordered, which is stricter than Counter's
// comparable. Methods can't add constraints or type parameters of their own.
func TopN[K cmp.Ordered](c *Counter[K], n int) []Entry[K] {
	entries := make([]Entry[K], 0, c.Len())
	for key, count := range c.counts {
		entries = append(entries, Entry[K]{Key: key, Count: count})
	}
	slices.SortFunc(entries, func(a, b Entry[K]) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c // higher counts first
		}
		return cmp.Compare(a.Key, b.Key)
	})
	return entries[:min(n, len(entries))]
}
```

- `K comparable` is enough: `K` is only a map key.
- The zero value is usable: `Add` creates the map on first use. `Get`, `Len`, and `Total` work on a nil
  map, since reading one is allowed.
- `TopN` breaks ties by sorting keys, so it needs `cmp.Ordered`. A method on `Counter[K]` only gets
  `comparable`, so `TopN` must be a function (see "Methods can't have their own type parameters").
- Map order is random. Without the key tie-break, files with equal counts would come out in a
  different order each run.
- `min(n, len(entries))` keeps the slice in bounds when there are fewer than `n` entries.

### cmd/seek/flags.go

Add the option to the help text, the `options` struct, and the parser:

```go
      --color WHEN          highlight matches: auto, always, or never
                            (default auto, or the value of $SEEK_COLOR)
      --stats               print statistics to standard error when done
      --version             print the version and exit
```

```go
	glob        string    // --glob PATTERN
	stats       bool      // --stats
	version     bool      // --version
```

```go
	boolVar(&opts.hidden, "hidden", "")
	boolVar(&opts.stats, "stats", "")
	boolVar(&opts.version, "version", "")
```

### cmd/seek/main.go

Import `time` and the new package:

```go
import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"seek/internal/output"
	"seek/internal/search"
	"seek/internal/stats"
)
```

Record the start time at the top of `run`. `time.Now` returns the current time; `time.Since(start)`
returns the `time.Duration` since then ([Chapter 18](18-context-and-time.md) covers `time`).

```go
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	start := time.Now()

	// Writing many small pieces straight to a file costs one system call
```

`total` becomes a `stats.Counter[string]`, and a `searched` count is added. Replace everything from
the `total, hadError := 0, false` line to the end of `run`:

```go
	// A bad file shouldn't stop the search: report it, remember that
	// something went wrong, and carry on with the next one.
	var (
		searched int                   // files searched
		matched  stats.Counter[string] // matches per file, for files with any
		hadError bool
	)
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
		searched++
		if len(matches) > 0 {
			matched.Add(path, len(matches))
		}
	})

	if opts.stats {
		out.Flush() // print the results before the statistics
		printStats(stderr, searched, &matched, time.Since(start))
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
```

And add `printStats` after `run`:

```go
// printStats prints a summary of the search to w.
func printStats(w io.Writer, searched int, matched *stats.Counter[string], elapsed time.Duration) {
	fmt.Fprintf(w, "files searched: %d\n", searched)
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
```

- `var matched stats.Counter[string]` needs no constructor; the zero value works.
- Only files with matches are added, so `matched.Len()` counts matching files.
- `stats.TopN(matched, 5)` infers `K = string` from the `*stats.Counter[string]` argument.
- Results go to `out` (buffered stdout); statistics go straight to stderr. Without `out.Flush()`
  first, the statistics could appear before the last results.
- `elapsed.Round(time.Microsecond)` drops the nanoseconds to keep the line short.

### Checkpoint

Your `elapsed` times will differ:

```console
$ go run ./cmd/seek --stats -i fox testdata
testdata/notes.md:- feed the FOX
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/poem.txt:Foxes are quick and clever.
testdata/src/hello.go:	println("hello, fox")
files searched: 3
files matched:  3
total matches:  5
elapsed:        131µs
top files:
      3  testdata/poem.txt
      1  testdata/notes.md
      1  testdata/src/hello.go
$ go run ./cmd/seek --stats -c --hidden o testdata
testdata/.cache/fox.txt:1
testdata/.hidden.txt:1
testdata/notes.md:3
testdata/poem.txt:4
testdata/src/hello.go:1
files searched: 5
files matched:  5
total matches:  10
elapsed:        152µs
top files:
      4  testdata/poem.txt
      3  testdata/notes.md
      1  testdata/.cache/fox.txt
      1  testdata/.hidden.txt
      1  testdata/src/hello.go
$ go run ./cmd/seek --stats -l zebra testdata
files searched: 3
files matched:  0
total matches:  0
elapsed:        115µs
exit status 1
```

- "files searched" excludes the binary file, which is skipped before `searched++`.
- The three files with one match each are sorted by path, thanks to `TopN`'s tie-break.

Statistics go to stderr, so you can separate them from the results:

```console
$ go run ./cmd/seek --stats -l fox testdata >/dev/null
files searched: 3
files matched:  2
total matches:  2
elapsed:        145µs
top files:
      1  testdata/poem.txt
      1  testdata/src/hello.go
$ go run ./cmd/seek --stats -l fox testdata 2>/dev/null
testdata/poem.txt
testdata/src/hello.go
```

With `-l`, each file stops at its first match, so every count is 1. An inverted search for any
character (`-v -E '.'`) finds only the empty line in `hello.go`:

```console
$ go run ./cmd/seek --stats -v -E '.' testdata
testdata/src/hello.go:
files searched: 3
files matched:  1
total matches:  1
elapsed:        149µs
top files:
      1  testdata/src/hello.go
```

The complete project at this point is in `project/ch13/`.

## Summary

- `[T Constraint]` makes a function or type generic; supplying type arguments instantiates it.
- Function type arguments are usually inferred from the arguments. One used only in the result must be
  written.
- Constraints are interfaces. They can list types (`int | float64`), use `~T` for all types with
  underlying type `T`, and require methods. Interfaces with type lists work only as constraints.
- `any` allows everything, `comparable` allows `==`, `cmp.Ordered` allows `<`. The body may use only
  operations every type in the type set supports.
- `var zero T` gives a type parameter's zero value.
- Methods use the receiver's type parameters but can't declare their own. Use a function for a new
  type parameter or a stricter constraint.
- Generic type aliases (`type Set[T comparable] = map[T]struct{}`) name an existing type.
- `slices`, `maps`, and `cmp` provide generic helpers: `Sort`, `SortFunc`, `Max`, `Contains`,
  `BinarySearch`, `Clone`, `cmp.Compare`, `cmp.Or`. `min` and `max` are built in.
- Interfaces for different behaviour; generics for the same code over different types.

## Exercises

1. **Reverse.** Write `Reverse[T any](s []T) []T`: return a new, reversed slice, leaving `s`
   unchanged.

<details><summary>Solution</summary>

```go
package main

import "fmt"

// Reverse returns a new slice with the elements of s in reverse order.
func Reverse[T any](s []T) []T {
	out := make([]T, len(s))
	for i, v := range s {
		out[len(s)-1-i] = v
	}
	return out
}

func main() {
	fmt.Println(Reverse([]int{1, 2, 3}))
	fmt.Printf("%q\n", Reverse([]string{"a", "b"}))
	fmt.Println(Reverse([]bool{}))
}
```

```console
$ go run .
[3 2 1]
["b" "a"]
[]
```

The standard library's `slices.Reverse` reverses in place.

</details>

2. **GroupBy.** Write `GroupBy`: given a slice and a key function, return a map from each key to the
   elements with that key. Group words by first byte, and numbers by digit count.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"strings"
)

// GroupBy puts each element of s into a group named by key(element).
func GroupBy[T any, K comparable](s []T, key func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, v := range s {
		k := key(v)
		groups[k] = append(groups[k], v)
	}
	return groups
}

func main() {
	words := strings.Fields("go gopher zig rust ruby gc")
	byFirst := GroupBy(words, func(w string) byte { return w[0] })
	fmt.Printf("%q\n", byFirst['g'])
	fmt.Printf("%q\n", byFirst['r'])

	byLen := GroupBy([]int{1, 22, 333, 4, 55}, func(n int) int {
		return len(fmt.Sprint(n))
	})
	fmt.Println(byLen) // fmt prints maps with sorted keys
}
```

```console
$ go run .
["go" "gopher" "gc"]
["rust" "ruby"]
map[1:[1 4] 2:[22 55] 3:[333]]
```

`K` must be `comparable` (it's a map key); `T` can be `any`.

</details>

3. **A Set type.** Write a generic `Set[T comparable]` struct, usable as a zero value, with methods
   `Add`, `Has`, and `Len`. Then add a way to get the elements sorted. Can it be a method?

<details><summary>Solution</summary>

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

// A Set is an unordered collection of distinct values.
// The zero value is an empty set ready to use.
type Set[T comparable] struct {
	m map[T]struct{}
}

func (s *Set[T]) Add(v T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	s.m[v] = struct{}{}
}

func (s *Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set[T]) Len() int { return len(s.m) }

// Sorted needs ordered elements, so it's a function, not a method.
func Sorted[T cmp.Ordered](s *Set[T]) []T {
	out := make([]T, 0, s.Len())
	for v := range s.m {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}

func main() {
	var s Set[string]
	for _, w := range []string{"b", "a", "b", "c"} {
		s.Add(w)
	}
	fmt.Println(s.Len(), s.Has("a"), s.Has("z"))
	fmt.Println(Sorted(&s))
}
```

```console
$ go run .
3 true false
[a b c]
```

It can't be a method: sorting needs `cmp.Ordered`, stricter than `Set`'s `comparable`. Same as
`seek`'s `TopN`.

</details>

4. **Matches by extension.** In `seek`, use a second `stats.Counter[string]` to count matches per
   extension (`filepath.Ext(path)`) and print them under `by extension:` after the other statistics.
   Undo the change afterwards.

<details><summary>Solution</summary>

In `cmd/seek/main.go`, import `path/filepath` and add the counter next to `matched`:

```go
	var (
		searched int                   // files searched
		matched  stats.Counter[string] // matches per file, for files with any
		byExt    stats.Counter[string] // matches per file extension
		hadError bool
	)
```

Count in the callback:

```go
		if len(matches) > 0 {
			matched.Add(path, len(matches))
			byExt.Add(filepath.Ext(path), len(matches))
		}
```

And print after `printStats`:

```go
	if opts.stats {
		out.Flush() // print the results before the statistics
		printStats(stderr, searched, &matched, time.Since(start))
		if byExt.Len() > 0 {
			fmt.Fprintln(stderr, "by extension:")
		}
		for _, e := range stats.TopN(&byExt, byExt.Len()) {
			fmt.Fprintf(stderr, "%7d  %s\n", e.Count, e.Key)
		}
	}
```

```console
$ go run ./cmd/seek --stats -c --hidden o testdata
testdata/.cache/fox.txt:1
testdata/.hidden.txt:1
testdata/notes.md:3
testdata/poem.txt:4
testdata/src/hello.go:1
files searched: 5
files matched:  5
total matches:  10
elapsed:        185µs
top files:
      4  testdata/poem.txt
      3  testdata/notes.md
      1  testdata/.cache/fox.txt
      1  testdata/.hidden.txt
      1  testdata/src/hello.go
by extension:
      6  .txt
      3  .md
      1  .go
```

One `Counter` type serves both; it doesn't care what keys mean. Passing `byExt.Len()` as `n` asks
`TopN` for every entry.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-13--generics)
