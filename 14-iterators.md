# Chapter 14 — Iterators and Range-over-Func

`for range` works on slices, maps, strings, channels, and integers, and since Go 1.23 on
**functions**. A function of the right shape is an **iterator**: it produces values one at a time for
a normal `for` loop, where `break`, `continue`, and `return` work as usual. You'll learn how iterators
work, how to write them, and the standard library's iterator functions. Then `seek`'s callback-based
`Walk` becomes an iterator.

Previous: [Chapter 13 — Generics](13-generics.md). Next: [Chapter 15 — Encoding: JSON and Struct
Tags](15-json-and-encoding.md).

## Ranging over a function

You can range over a function whose only argument is a **yield** function. The iterator calls `yield`
once per value; `yield` returns `false` when the loop wants to stop.

`examples/ch14/rangefunc/main.go`

```go
package main

import "fmt"

// Countdown is a function that range can loop over. It calls yield once
// per value. yield returns false when the loop wants to stop.
func Countdown(yield func(int) bool) {
	for i := 3; i >= 1; i-- {
		if !yield(i) {
			return
		}
	}
}

// Three yield shapes: no values, one value, two values.
func Twice(yield func() bool) {
	_ = yield() && yield()
}

func Letters(yield func(int, string) bool) {
	for i, s := range []string{"a", "b", "c"} {
		if !yield(i, s) {
			return
		}
	}
}

func main() {
	for n := range Countdown {
		fmt.Print(n, " ")
	}
	fmt.Println("liftoff")

	for range Twice {
		fmt.Println("hello")
	}

	for i, s := range Letters {
		fmt.Println(i, s)
	}
}
```

```console
$ go run ./ch14/rangefunc
3 2 1 liftoff
hello
hello
0 a
1 b
2 c
```

There are exactly three shapes, for loops with zero, one, or two variables:

```go
func(yield func() bool)
func(yield func(V) bool)
func(yield func(K, V) bool)
```

`for n := range Countdown` uses the function `Countdown` itself; it doesn't call it.

## How the loop talks to the iterator

The compiler turns the loop body into a function and passes it as `yield`. Each `yield` call runs the
body once. When the body leaves the loop (`break`, `return`, or `goto`), `yield` returns `false` and
the iterator must stop.

`examples/ch14/earlystop/main.go`

```go
package main

import (
	"fmt"
	"iter"
)

// Numbers returns an iterator over 1..n that prints what happens.
func Numbers(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; i <= n; i++ {
			fmt.Println("  yield", i)
			if !yield(i) {
				fmt.Println("  loop stopped; cleaning up")
				return
			}
		}
		fmt.Println("  done")
	}
}

func main() {
	fmt.Println("full loop:")
	for n := range Numbers(2) {
		fmt.Println("body", n)
	}

	fmt.Println("break at 2:")
	for n := range Numbers(5) {
		fmt.Println("body", n)
		if n == 2 {
			break
		}
	}

	// A range loop is rewritten into a call like this one:
	fmt.Println("called by hand:")
	Numbers(5)(func(n int) bool {
		fmt.Println("body", n)
		return n < 2 // false means "stop"
	})
}
```

```console
$ go run ./ch14/earlystop
full loop:
  yield 1
body 1
  yield 2
body 2
  done
break at 2:
  yield 1
body 1
  yield 2
body 2
  loop stopped; cleaning up
called by hand:
  yield 1
body 1
  yield 2
body 2
  loop stopped; cleaning up
```

- Iterator and body take turns: "yield 1", "body 1", "yield 2", ... No slice of all values is built.
- After `break`, the iterator gets control back, so it can clean up (close a file, say) before
  returning.
- `Numbers(5)` *returns* an iterator. This is the usual style: a function or method takes arguments
  and returns the iterator.

### iter.Seq and iter.Seq2

The `iter` package names the two common shapes:

```go
type Seq[V any] func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)
```

These are generic types ([Chapter 13](13-generics.md)). Returning `iter.Seq[int]` instead of the
spelled-out function type reads more easily and says "this is an iterator".

### The rule: stop when yield returns false

An iterator that ignores `false` and calls `yield` again panics:

```go
// Bad ignores yield's result.
func Bad(yield func(int) bool) {
	for i := 1; i <= 3; i++ {
		yield(i)
	}
}

func main() {
	for n := range Bad {
		fmt.Println(n)
		break
	}
}
```

```console
$ go run .
1
panic: runtime error: range function continued iteration after function for loop body returned false

goroutine 1 [running]:
...
exit status 2
```

Always write `if !yield(v) { return }`. `seek`'s iterator below must take care here.

## Iterators over your own types

A collection type usually has an `All` method returning an iterator. A tree is a good example: its
values are spread over many nodes, and the iterator hides how to reach them.

`examples/ch14/tree/main.go`

```go
package main

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
)

// Tree is a binary search tree of ordered values.
type Tree[T cmp.Ordered] struct {
	root *node[T]
}

type node[T cmp.Ordered] struct {
	val         T
	left, right *node[T]
}

func (t *Tree[T]) Insert(v T) {
	p := &t.root
	for *p != nil {
		if v < (*p).val {
			p = &(*p).left
		} else {
			p = &(*p).right
		}
	}
	*p = &node[T]{val: v}
}

// All returns an iterator over the values in sorted order.
func (t *Tree[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		t.root.walk(yield)
	}
}

// walk visits n's subtree in order. It returns false if yield did,
// so every level of the recursion stops at once.
func (n *node[T]) walk(yield func(T) bool) bool {
	if n == nil {
		return true
	}
	return n.left.walk(yield) && yield(n.val) && n.right.walk(yield)
}

func main() {
	var t Tree[string]
	for _, w := range []string{"m", "c", "x", "a", "e", "z"} {
		t.Insert(w)
	}
	fmt.Println(slices.Collect(t.All()))

	var small []string
	for v := range t.All() {
		if v > "d" {
			break // stops the walk deep inside the recursion
		}
		small = append(small, v)
	}
	fmt.Println(small)
}
```

```console
$ go run ./ch14/tree
[a c e m x z]
[a c]
```

- `Insert` uses a pointer to a pointer (`p := &t.root`) to find and fill the empty spot.
- The recursive `walk` returns `bool`. When `yield` returns `false`, every level sees it and returns
  at once (`&&` short-circuits). Otherwise the recursion would keep calling `yield` and panic.
- `slices.Collect` runs an iterator to the end and returns its values as a slice.

## Pull iterators

So far these are **push** iterators: the iterator drives the loop and pushes values into `yield`.
Sometimes you need to request values one at a time instead, for example to walk two sequences side by
side. `iter.Pull` turns a push iterator into `next` and `stop` functions:

`examples/ch14/pull/main.go`

```go
package main

import (
	"fmt"
	"iter"
	"slices"
)

// Zip pairs up the values of two sequences until either one ends.
// A range loop can only drive one iterator, so Zip pulls from both.
func Zip[A, B any](as iter.Seq[A], bs iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		nextA, stopA := iter.Pull(as)
		defer stopA()
		nextB, stopB := iter.Pull(bs)
		defer stopB()
		for {
			a, ok1 := nextA()
			b, ok2 := nextB()
			if !ok1 || !ok2 || !yield(a, b) {
				return
			}
		}
	}
}

func main() {
	next, stop := iter.Pull(slices.Values([]string{"x", "y"}))
	defer stop() // always call stop to release the iterator
	for {
		v, ok := next()
		if !ok {
			break
		}
		fmt.Println("pulled", v)
	}

	names := slices.Values([]string{"ann", "bob", "cy"})
	ages := slices.Values([]int{31, 42})
	for name, age := range Zip(names, ages) {
		fmt.Println(name, age)
	}
}
```

```console
$ go run ./ch14/pull
pulled x
pulled y
ann 31
bob 42
```

- `next()` returns the next value and `true`, or the zero value and `false` when done.
- `stop()` ends the iterator early. Always call it, usually with `defer`; calling it twice is safe.
- `iter.Pull2` does the same for `iter.Seq2`.

Pull iterators cost more, because the iterator runs apart from the pulling code (on its own goroutine,
[Chapter 16](16-goroutines-and-channels.md)). Use them only when a range loop can't do the job.

## Iterators in the standard library

Many packages return iterators. These are the most used:

`examples/ch14/stdlib/main.go`

```go
package main

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	s := []string{"c", "a", "b"}
	for i, v := range slices.All(s) {
		fmt.Printf("%d=%s ", i, v)
	}
	fmt.Println("(All)")
	for i, v := range slices.Backward(s) {
		fmt.Printf("%d=%s ", i, v)
	}
	fmt.Println("(Backward)")
	fmt.Println(slices.Sorted(slices.Values(s)))
	fmt.Println(slices.Collect(slices.Values(s)))

	m := map[string]int{"go": 2009, "c": 1972, "zig": 2016}
	fmt.Println(slices.Sorted(maps.Keys(m)))
	fmt.Println(slices.Sorted(maps.Values(m)))
	clone := maps.Collect(maps.All(m))
	fmt.Println(len(clone))

	text := "one\ntwo\nthree"
	fmt.Printf("%q\n", slices.Collect(strings.Lines(text))) // lines keep their "\n"
	fmt.Printf("%q\n", slices.Collect(strings.SplitSeq("a,b,,c", ",")))
	for f := range strings.FieldsSeq("  lots   of\tspace ") {
		fmt.Println(f)
	}
	n := 0
	for range bytes.Lines([]byte("x\ny\n")) {
		n++
	}
	fmt.Println(n, "lines")
}
```

```console
$ go run ./ch14/stdlib
0=c 1=a 2=b (All)
2=b 1=a 0=c (Backward)
[a b c]
[c a b]
[c go zig]
[1972 2009 2016]
3
["one\n" "two\n" "three"]
["a" "b" "" "c"]
lots
of
space
2 lines
```

| Function | Returns |
|---|---|
| `slices.All(s)` | `iter.Seq2[int, E]`: index and value |
| `slices.Values(s)` | `iter.Seq[E]`: values |
| `slices.Backward(s)` | `iter.Seq2[int, E]`, from the end |
| `slices.Collect(seq)` | a new slice with all values |
| `slices.Sorted(seq)` | a new sorted slice |
| `maps.Keys(m)`, `maps.Values(m)` | `iter.Seq` of keys or values, in random order |
| `maps.All(m)` | `iter.Seq2[K, V]` |
| `maps.Collect(seq2)` | a new map |
| `strings.Lines(s)`, `bytes.Lines(b)` | lines, each with its `"\n"` |
| `strings.SplitSeq`, `strings.FieldsSeq` | like `Split` and `Fields`, without building a slice |

`slices.Sorted(maps.Keys(m))` is the standard way to loop over a map in fixed order.

## Composing iterators

An iterator can transform another iterator. With generics, `Filter`, `Map`, and `Take` are written
once for every type. Values flow through the chain one at a time, so the source can even be infinite:

`examples/ch14/adapters/main.go`

```go
package main

import (
	"fmt"
	"iter"
	"slices"
)

// Naturals yields 1, 2, 3, ... forever. That's fine for an iterator:
// values are only produced when the loop asks for them.
func Naturals() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Filter yields the values of seq for which keep returns true.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Map yields f(v) for each value v of seq.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Take yields at most the first n values of seq.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for v := range seq {
			if !yield(v) {
				return
			}
			i++
			if i == n {
				return
			}
		}
	}
}

func main() {
	odd := Filter(Naturals(), func(n int) bool { return n%2 == 1 })
	squares := Map(odd, func(n int) int { return n * n })
	fmt.Println(slices.Collect(Take(squares, 5)))
}
```

```console
$ go run ./ch14/adapters
[1 9 25 49 81]
```

- Each adapter ranges over its input and yields to its output. After five values `Take` returns,
  which stops the `for v := range seq` in `Map`, which stops `Filter`, which stops `Naturals`. A
  `return` inside a range-over-func loop passes "stop" up the chain.
- `Take` returns right after the n-th value rather than requesting one more it would never use.

Don't overdo it: a plain `for` with an `if` is often clearer than an adapter chain. Iterators shine at
hiding *how* values are found, as in the tree above and `seek`'s file walk.

## Building seek: Walk as an iterator

Since [Chapter 11](11-io-and-files.md), `search.Walk` has taken a callback:

```go
search.Walk(opts.paths, walkOpts, func(path string, err error) {
	...
	return // means "next file", not "leave run"
})
```

In the callback, `return` ends only one call, there's no `break` or `continue`, and the closure must
capture the counters. With an iterator it's a normal loop:

```go
for path, err := range search.Walk(opts.paths, walkOpts) {
	...
	continue
}
```

A caller can also stop early with `break`, which the callback couldn't do.

### internal/search/walk.go

Replace `Walk` with an iterator. Add `"iter"` to the imports:

```go
package search

import (
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

// WalkOptions controls which files Walk visits.
type WalkOptions struct {
	Hidden bool   // include hidden files and directories (names starting with ".")
	Glob   string // if not empty, only visit files whose base name matches it
}

// Walk returns an iterator over the files to search under roots, in
// order. A root that is a directory is walked recursively, in lexical
// order. The root "-" (standard input) is passed through unchanged. If a
// file or directory can't be read, the iterator yields its path with the
// error and carries on.
//
// Hidden entries, entries that aren't regular files (such as symbolic
// links), and files not matching opts.Glob are skipped, but only when
// they are found inside a directory: a root is always used as given.
//
//	for path, err := range search.Walk(roots, opts) { ... }
func Walk(roots []string, opts WalkOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, root := range roots {
			if !walkRoot(root, opts, yield) {
				return // the loop body stopped early (break or return)
			}
		}
	}
}

// walkRoot yields the files under root. It reports whether to keep
// going: false means yield returned false and we must stop.
func walkRoot(root string, opts WalkOptions, yield func(string, error) bool) bool {
	if root == "-" {
		return yield(root, nil)
	}
	keepGoing := true
	// emit hands a result to yield and ends the walk if yield says stop.
	emit := func(path string, err error) error {
		if !yield(path, err) {
			keepGoing = false
			return filepath.SkipAll
		}
		return nil
	}
	// Our callback returns only nil, SkipDir, or SkipAll, so WalkDir
	// itself never returns an error.
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return emit(path, err)
		}
		if path == root {
			if !d.IsDir() {
				return emit(path, nil)
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
		return emit(path, nil)
	})
	return keepGoing
}
```

The filtering rules are unchanged. What changed is how a result leaves the walk:

- `Walk` returns an `iter.Seq2[string, error]`, so a loop gets two values: path and error.
- `filepath.WalkDir` is itself callback-based; its callback can't `return` out of the whole walk. So
  `emit` calls `yield`, and if `yield` returns `false`, it returns `filepath.SkipAll` to stop `WalkDir`
  at once. Otherwise `WalkDir` would move to the next file and call `yield` again: the panic from "The
  rule: stop when yield returns false".
- `WalkDir` returns `nil` after `SkipAll`, so its result can't tell `walkRoot` the loop stopped. The
  `keepGoing` variable, set by `emit`, carries that back, and `Walk` skips the remaining roots.
- Before, a read error went to the callback and the walk went on (`return nil`). `emit` still returns
  `nil` then, unless the loop has stopped.

### internal/stats/counter.go

Give `Counter` an `All` iterator so outside code can loop over the counts without touching the map.
Add `"iter"` and `"maps"` to the imports:

```go
import (
	"cmp"
	"iter"
	"maps"
	"slices"
)
```

Add `All` after `Total`, and use it in `Total` and `TopN`:

```go
// Total returns the sum of all counts.
func (c *Counter[K]) Total() int {
	total := 0
	for _, n := range c.All() {
		total += n
	}
	return total
}

// All returns an iterator over the keys and their counts, in no
// particular order.
func (c *Counter[K]) All() iter.Seq2[K, int] {
	return maps.All(c.counts) // ranging over a nil map is fine
}
```

```go
func TopN[K cmp.Ordered](c *Counter[K], n int) []Entry[K] {
	entries := make([]Entry[K], 0, c.Len())
	for key, count := range c.All() {
		entries = append(entries, Entry[K]{Key: key, Count: count})
	}
```

`maps.All` already returns an `iter.Seq2[K, V]`, so `All` is one line. `TopN` now keeps working if
`Counter` stores its counts differently.

### cmd/seek/main.go

The callback becomes a loop body; its two "next file" `return`s become `continue`:

```go
	for path, err := range search.Walk(opts.paths, walkOpts) {
		var matches []search.Match
		if err == nil {
			matches, err = s.SearchFile(path)
		}
		if errors.Is(err, search.ErrBinary) {
			continue // binary files are skipped silently
		}
		if err == nil {
			err = p.PrintFile(path, matches)
		}
		if err != nil {
			report(stderr, err)
			hadError = true
			continue
		}
		searched++
		if len(matches) > 0 {
			matched.Add(path, len(matches))
		}
	}
```

Nothing else changes. `searched`, `matched`, and `hadError` are plain local variables in plain loop
code again.

### Checkpoint

Behaviour matches Chapter 13. Your `elapsed` time will differ:

```console
$ go run ./cmd/seek -n fox testdata
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek --stats -i fox testdata
testdata/notes.md:- feed the FOX
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/poem.txt:Foxes are quick and clever.
testdata/src/hello.go:	println("hello, fox")
files searched: 3
files matched:  3
total matches:  5
elapsed:        137µs
top files:
      3  testdata/poem.txt
      1  testdata/notes.md
      1  testdata/src/hello.go
$ go run ./cmd/seek -l -i fox testdata/poem.txt - testdata/src < testdata/notes.md
testdata/poem.txt
(standard input)
testdata/src/hello.go
```

The last run mixes a file, standard input, and a directory, still searched in the given order.
`seek` never stops a walk early, so `SkipAll` isn't used yet; Exercise 4 tries it.

The complete project at this point is in `project/ch14/`.

## Summary

- `for x := range f` works when `f` is `func(yield func() bool)`, `func(yield func(V) bool)`, or
  `func(yield func(K, V) bool)`. Each `yield` call runs the loop body once.
- `yield` returns `false` when the loop ends early (`break`, `return`). The iterator must stop;
  calling `yield` again panics.
- `iter.Seq[V]` and `iter.Seq2[K, V]` name the common iterator types. Functions and methods usually
  *return* one, as `All()` does.
- Recursive iterators pass "stop" up through every level.
- `iter.Pull` turns a push iterator into `next`/`stop` functions. Always call `stop`.
- `slices.All/Values/Backward/Collect/Sorted`, `maps.Keys/Values/All/Collect`, `strings.Lines`,
  `strings.SplitSeq`, `strings.FieldsSeq`, and `bytes.Lines` work with iterators.
- Generic adapters like `Filter`, `Map`, and `Take` compose iterators lazily, even infinite ones.
- To wrap a callback API like `filepath.WalkDir`, turn `yield`'s `false` into the API's "stop" signal
  (`filepath.SkipAll`).

## Exercises

1. **Fibonacci.** Write `Fibonacci() iter.Seq[int]`, yielding 0, 1, 1, 2, 3, 5, ... forever. Print
   the numbers up to 100.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"iter"
)

// Fibonacci yields 0, 1, 1, 2, 3, 5, ... without end.
func Fibonacci() iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1
		for {
			if !yield(a) {
				return
			}
			a, b = b, a+b
		}
	}
}

func main() {
	for n := range Fibonacci() {
		if n > 100 {
			break
		}
		fmt.Print(n, ",")
	}
	fmt.Println("...")
}
```

```console
$ go run .
0,1,1,2,3,5,8,13,21,34,55,89,...
```

The infinite `for` is safe: the loop's `break` makes `yield` return `false`.

</details>

2. **Lines with errors.** Write `Lines(r io.Reader) iter.Seq2[string, error]`, yielding each line of
   `r` (without the line ending) with a `nil` error. If reading fails, yield `""` and the error once,
   then stop. Use `bufio.Scanner` ([Chapter 11](11-io-and-files.md)).

<details><summary>Solution</summary>

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strings"
)

// Lines yields the lines of r without their line endings. If reading
// fails, it yields ("", err) once and stops.
func Lines(r io.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			if !yield(sc.Text(), nil) {
				return
			}
		}
		if err := sc.Err(); err != nil {
			yield("", err)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("disk on fire")
}

func main() {
	r := strings.NewReader("alpha\nbeta\r\ngamma")
	for line, err := range Lines(r) {
		if err != nil {
			fmt.Println("error:", err)
			break
		}
		fmt.Printf("%q\n", line)
	}

	r2 := io.MultiReader(strings.NewReader("one\n"), failingReader{})
	for line, err := range Lines(r2) {
		if err != nil {
			fmt.Println("error:", err)
			break
		}
		fmt.Printf("%q\n", line)
	}
}
```

```console
$ go run .
"alpha"
"beta"
"gamma"
"one"
error: disk on fire
```

The final `yield("", err)` ignores the result: it's the last call, so there's nothing to stop.

</details>

3. **Enumerate.** Write `Enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T]`, pairing each value
   with its position as `slices.All` does for a slice. Use it to number a map's sorted keys.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

// Enumerate pairs each value of seq with its position, starting at 0.
func Enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}

func main() {
	m := map[string]bool{"go": true, "zig": true, "c": true}
	keys := slices.Sorted(maps.Keys(m))
	for i, k := range Enumerate(slices.Values(keys)) {
		fmt.Println(i, k)
	}
}
```

```console
$ go run .
0 c
1 go
2 zig
```

</details>

4. **Stopping the walk.** In `seek`, add a second command, `cmd/walkdemo/main.go`. It ranges over
   `search.Walk(os.Args[1:], ...)` with hidden files included, prints each path (or error), and
   `break`s after three files. Check that it stops cleanly, without a panic. Delete it afterwards.

<details><summary>Solution</summary>

`cmd/walkdemo` is inside the `seek` module, so it may import `seek/internal/search` ([Chapter
8](08-packages-and-modules.md)).

```go
// Command walkdemo prints the first few files seek would search.
package main

import (
	"fmt"
	"os"

	"seek/internal/search"
)

func main() {
	opts := search.WalkOptions{Hidden: true}
	n := 0
	for path, err := range search.Walk(os.Args[1:], opts) {
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(path)
		n++
		if n == 3 {
			break
		}
	}
	fmt.Println("files printed:", n)
}
```

```console
$ go run ./cmd/walkdemo testdata nope
testdata/.cache/fox.txt
testdata/.hidden.txt
testdata/image.bin
files printed: 3
$ go run ./cmd/walkdemo nope testdata/src
error: lstat nope: no such file or directory
testdata/src/hello.go
files printed: 1
```

In the first run, `break` happens inside `testdata`: `emit` returns `filepath.SkipAll`, `walkRoot`
returns `false`, and `Walk` never reaches the root `nope`. Replace `return filepath.SkipAll` with
`return nil` in `emit`, and the same command panics with "range function continued iteration after
function for loop body returned false".

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-14--iterators-and-range-over-func)
