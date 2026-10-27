# Chapter 5 — Arrays, Slices, and Maps

Go has three built-in collections: arrays (fixed size), slices (growable lists), and maps
(key/value lookups). You'll see how each works in memory, since that's where the gotchas live, plus
the `slices` and `maps` packages. `seek` searched one file in [Chapter 4](04-functions.md); now it
searches many and counts matches with `-c`. [Chapter 6](06-strings.md) looks closely at strings.

## Arrays

An array's fixed length is part of its type.

`examples/ch05/arrays/main.go`

```go
// Command arrays shows fixed-size arrays and their value semantics.
package main

import "fmt"

func main() {
	var a [3]int // three ints, all zero
	a[0] = 10
	fmt.Println(a, len(a))

	b := [3]string{"x", "y", "z"}
	c := [...]int{1, 2, 3, 4} // [...] counts the elements for you
	fmt.Println(b, len(c))

	// Arrays are values: assignment copies every element.
	d := a
	d[1] = 99
	fmt.Println(a, d)

	// Arrays of the same type can be compared with ==.
	fmt.Println(a == [3]int{10, 0, 0}, a == d)

	// The length is part of the type: [3]int and [4]int are different types.
	fmt.Printf("%T %T\n", a, c)

	var board [2][3]int // an array of arrays
	board[1][2] = 7
	fmt.Println(board)
}
```

```console
$ go run ./ch05/arrays
[10 0 0] 3
[x y z] 4
[10 0 0] [10 99 0]
true false
[3]int [4]int
[[0 0 0] [0 0 7]]
```

- Arrays are **values**: assigning or passing one copies every element.
- `==` compares element by element (if the element type is comparable).
- `[3]int` and `[4]int` are different types:
  `cannot use b (variable of type [4]int) as [3]int value in assignment`.
- Constant indexes are checked at compile time: `invalid argument: index 5 out of bounds [0:3]`.

You'll rarely use arrays directly; their fixed size makes them awkward parameters. They suit small
fixed-size data (a 16-byte hash, an RGB color `[3]uint8`) and map keys. Their real job is storage
behind slices.

## Slices

A slice is a view into an array: Go's everyday list type.

`examples/ch05/slices/main.go`

```go
// Command slices shows how to create, index, and slice slices.
package main

import "fmt"

func main() {
	primes := []int{2, 3, 5, 7, 11, 13} // a slice literal: no length
	fmt.Println(primes, len(primes), cap(primes))

	// s[low:high] includes low, excludes high.
	fmt.Println(primes[1:4]) // elements 1, 2, 3
	fmt.Println(primes[:2])  // from the start
	fmt.Println(primes[4:])  // to the end

	// make creates a slice with a length (and optionally a capacity).
	zeros := make([]int, 3)
	buf := make([]int, 0, 10)
	fmt.Println(zeros, len(buf), cap(buf))

	// append returns a new slice header; always assign the result.
	buf = append(buf, 1)
	buf = append(buf, 2, 3)
	buf = append(buf, primes[:2]...)
	fmt.Println(buf)

	// Index out of range is a run-time panic, so check lengths.
	i := 10
	if i < len(primes) {
		fmt.Println(primes[i])
	} else {
		fmt.Println("index", i, "out of range")
	}
}
```

```console
$ go run ./ch05/slices
[2 3 5 7 11 13] 6 6
[3 5 7]
[2 3]
[11 13]
[0 0 0] 0 10
[1 2 3 2 3]
index 10 out of range
```

### What a slice really is

A slice value is a small struct, the **slice header**, with three fields:

| Field | Meaning |
|-------|---------|
| pointer | the slice's first element in the underlying array |
| length (`len`) | number of elements in the slice |
| capacity (`cap`) | elements that fit from the slice's start to the array's end |

```text
underlying array:   [ 2 | 3 | 5 | 7 | 11 | 13 ]
primes      → ptr to index 0, len 6, cap 6
primes[1:4] → ptr to index 1, len 3, cap 5
```

Slicing (`s[low:high]`) makes a new header into the **same** array, copying no elements, so it's
cheap. Passing a slice copies only the header, so the function sees (and can change) the same
elements.

- `len(s)`, `cap(s)`: length and capacity.
- `s[i]`: element `i`; needs `0 <= i < len(s)` or the program panics.
- `s[low:high]`: elements `low` to `high-1`. Either bound may be omitted; `high` may reach
  `cap(s)`, not just `len(s)`.
- `make([]T, len)` or `make([]T, len, cap)`: a new slice of zero values.
- `append(s, v...)`: adds elements, returns the updated slice.

An out-of-range index is a **panic**, a run-time crash. This program skips the length check:

```go
package main

import "fmt"

func main() {
	primes := []int{2, 3, 5, 7, 11, 13}
	i := 10
	fmt.Println(primes[i])
}
```

```console
$ go run .
panic: runtime error: index out of range [10] with length 6

goroutine 1 [running]:
main.main()
	/path/to/main.go:8 +0x9
exit status 2
```

(The path is shortened; yours is full.) The message says what happened. The **stack trace** says
where: function `main.main`, file `main.go`, line 8. Ignore `+0x9` (a machine-code offset). Go
checks every index because silently reading past an array's end is a classic C security hole.

### How append grows a slice

With spare capacity, `append` writes into the existing array. Otherwise it allocates a bigger
array, copies the elements, and returns a slice of the new one.

`examples/ch05/growth/main.go`

```go
// Command growth shows how append grows a slice's capacity.
package main

import "fmt"

func main() {
	var s []int
	prevCap := cap(s)
	for i := range 2000 {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("len=%-4d cap %4d -> %4d\n", len(s), prevCap, cap(s))
			prevCap = cap(s)
		}
	}
}
```

```console
$ go run ./ch05/growth
len=1    cap    0 ->    4
len=5    cap    4 ->    8
len=9    cap    8 ->   16
len=17   cap   16 ->   32
len=33   cap   32 ->   64
len=65   cap   64 ->  128
len=129  cap  128 ->  256
len=257  cap  256 ->  512
len=513  cap  512 ->  848
len=849  cap  848 -> 1280
len=1281 cap 1280 -> 1792
len=1793 cap 1792 -> 2560
```

Capacity roughly doubles for small slices and grows about 25% for large ones, rounded to sizes the
allocator likes. Exact numbers are an implementation detail and may change between Go versions.
What matters:

- Appending is fast on average; most appends don't allocate.
- **Always assign the result**: `s = append(s, x)`; it may point to a new array. Dropping it is a
  compile error: `append(s, 2) (value of type []int) is not used`.
- Know the final size? Preallocate with `make([]T, 0, n)`, and `append` never reallocates.

### Gotcha: slices share memory

`examples/ch05/aliasing/main.go`

```go
// Command aliasing shows that slices share their underlying array.
package main

import "fmt"

func main() {
	nums := []int{1, 2, 3, 4, 5}

	// A slice of a slice shares the same array.
	part := nums[1:3] // [2 3], len 2, cap 4
	part[0] = 20
	fmt.Println(nums, part, len(part), cap(part))

	// append writes into the shared array when there is spare capacity,
	// silently overwriting nums[3].
	part = append(part, 40)
	fmt.Println(nums, part)

	// A full slice expression a[low:high:max] limits the capacity,
	// so append must copy to a new array instead.
	nums = []int{1, 2, 3, 4, 5}
	safe := nums[1:3:3] // len 2, cap 2
	safe = append(safe, 40)
	safe[0] = 99
	fmt.Println(nums, safe)
}
```

```console
$ go run ./ch05/aliasing
[1 20 3 4 5] [20 3] 2 4
[1 20 3 40 5] [20 3 40]
[1 2 3 4 5] [99 3 40]
```

- Changing `part[0]` changed `nums[1]`: same array.
- `append(part, 40)` had spare capacity, so it overwrote `nums[3]`. This is the most surprising
  slice bug: appending to a slice cut from another.
- The **full slice expression** `nums[1:3:3]` sets capacity to `max - low` = 2. The next `append`
  has no room, so it copies to a new array, and `nums` is safe.

Rule of thumb: when a function returns a sub-slice of its data, or keeps a slice it was given, ask
who else sees that array. When in doubt, copy.

### Copying

`examples/ch05/copying/main.go`

```go
// Command copying shows copy and how to make an independent slice.
package main

import (
	"fmt"
	"slices"
)

func main() {
	src := []int{1, 2, 3, 4}

	dst := make([]int, 2)
	n := copy(dst, src) // copies min(len(dst), len(src)) elements
	fmt.Println(n, dst)

	// A full, independent copy.
	dup := make([]int, len(src))
	copy(dup, src)
	dup[0] = 100
	fmt.Println(src, dup)

	// The same in one call.
	clone := slices.Clone(src)
	clone[1] = 200
	fmt.Println(src, clone)

	// copy handles overlapping ranges: shift left by one.
	copy(src, src[1:])
	fmt.Println(src)
}
```

```console
$ go run ./ch05/copying
2 [1 2]
[1 2 3 4] [100 2 3 4]
[1 2 3 4] [1 200 3 4]
[2 3 4 4]
```

- `copy(dst, src)` copies `min(len(dst), len(src))` elements and returns the count. It never grows
  `dst`, so copying into a `nil` or empty slice copies nothing.
- `slices.Clone(s)` is the easy independent copy.

### nil slices and empty slices

`examples/ch05/nilempty/main.go`

```go
// Command nilempty compares nil slices and empty slices.
package main

import "fmt"

func main() {
	var a []int    // nil: no underlying array
	b := []int{}   // empty, but not nil
	c := a[0:0:0]  // slicing a nil slice gives nil
	var d []string // nil

	fmt.Println(len(a), cap(a), a == nil)
	fmt.Println(len(b), cap(b), b == nil)
	fmt.Println(c == nil)

	// Both behave the same for len, range, and append.
	for range a {
		fmt.Println("never runs")
	}
	a = append(a, 1)
	d = append(d, "ok")
	fmt.Println(a, d)

	// To check for "no elements", test the length, not nil.
	fmt.Println(len(b) == 0)
}
```

```console
$ go run ./ch05/nilempty
0 0 true
0 0 false
true
[1] [ok]
true
```

A slice's zero value, `nil`, is ready to use: `len` is 0, `range` does nothing, `append` works.
So declare `var s []T` and append; no `make` needed. "Make the zero value useful" in action.

The nil/empty difference rarely matters. One exception: JSON encodes `nil` as `null` and empty as
`[]` ([Chapter 15](15-json-and-encoding.md)). Test emptiness with `len(s) == 0`, true for both.

Slices compare with `==` only to `nil`:

```console
./main.go:8:14: invalid operation: a == b (slice can only be compared to nil)
```

Use `slices.Equal` (below) instead.

### Two-dimensional slices

`examples/ch05/grid/main.go`

```go
// Command grid builds a two-dimensional slice.
package main

import "fmt"

func main() {
	rows, cols := 3, 4
	grid := make([][]rune, rows) // a slice of rows
	for r := range grid {
		grid[r] = make([]rune, cols) // each row is its own slice
		for c := range grid[r] {
			grid[r][c] = '.'
		}
	}
	grid[1][2] = '#'

	for _, row := range grid {
		fmt.Println(string(row))
	}

	// Rows may have different lengths ("jagged").
	tri := [][]int{{1}, {1, 1}, {1, 2, 1}}
	fmt.Println(tri, len(tri[2]))
}
```

```console
$ go run ./ch05/grid
....
..#.
....
[[1] [1 1] [1 2 1]] 3
```

There's no built-in 2-D type: `[][]T` is a slice of slices, and you allocate each row.
`string(row)` converts a `[]rune` to a string. Inside a composite literal you can drop the inner
type: `{1, 1}` instead of `[]int{1, 1}`.

## Maps

A map stores key/value pairs and finds a value by key in constant time on average.

`examples/ch05/maps/main.go`

```go
// Command maps shows how to create, read, write, and delete map entries.
package main

import "fmt"

func main() {
	ages := map[string]int{ // a map literal
		"alice": 31,
		"bob":   27,
	}
	ages["carol"] = 45 // add
	ages["bob"]++      // update
	fmt.Println(ages["alice"], ages["bob"], len(ages))

	// A missing key gives the zero value...
	fmt.Println(ages["dave"])

	// ...so use the "comma ok" form to tell "missing" from "zero".
	if age, ok := ages["dave"]; ok {
		fmt.Println("dave is", age)
	} else {
		fmt.Println("no dave")
	}

	delete(ages, "alice") // deleting a missing key is fine too
	_, ok := ages["alice"]
	fmt.Println(ok, len(ages))

	// make creates an empty map. The zero value makes counting easy.
	counts := make(map[string]int)
	for _, w := range []string{"a", "b", "a", "c", "a"} {
		counts[w]++
	}
	fmt.Println(counts) // fmt prints maps sorted by key
}
```

```console
$ go run ./ch05/maps
31 28 3
0
no dave
false 2
map[a:3 b:1 c:1]
```

- The type is `map[K]V`. Keys must be **comparable** with `==`: numbers, strings, booleans, arrays,
  and structs of those, for example. Slices, maps, and functions can't be keys.
- A missing key reads as `V`'s **zero value**. That's why `counts[w]++` works: a new word starts
  at 0.
- `v, ok := m[k]` is the **comma-ok** form: `ok` is `false` if the key is missing. Use it whenever
  zero is a valid value.
- `delete(m, k)` removes a key. `len(m)` counts entries.
- `fmt.Println` sorts map keys for stable output. Your own loops don't.

### Iteration order is random

`examples/ch05/maporder/main.go`

```go
// Command maporder shows that map iteration order is not fixed.
package main

import "fmt"

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	for range 3 {
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
}
```

```console
$ go run ./ch05/maporder
c d e a b 
a b c d e 
a b c d e 
```

Your output will differ, and change between runs. The runtime randomizes order on purpose: early
Go's order looked stable but differed across platforms, so code that relied on it passed on one
machine and failed on another. Randomizing exposes such bugs right away.

When order matters, sort the keys ([The maps package](#the-maps-package)) or keep a separate slice
that remembers the order, as `seek` does below.

### A nil map can't be written

`examples/ch05/nilmap/main.go`

```go
// Command nilmap shows what happens when you write to a nil map.
package main

import "fmt"

func main() {
	var m map[string]int // nil: reading works, writing panics
	fmt.Println(m["x"], len(m))
	m["x"] = 1
	fmt.Println("never printed")
}
```

```console
$ go run ./ch05/nilmap
0 0
panic: assignment to entry in nil map

goroutine 1 [running]:
main.main()
	/home/you/examples/ch05/nilmap/main.go:9 +0xa8
exit status 2
```

Your path and the `+0xa8` offset will differ. A map's zero value is `nil`. You can read, range
over, and `delete` from it, but writing panics. Create maps with `make(map[K]V)` or a literal
(`map[K]V{}`) before writing. This is where the zero value isn't fully ready: a map needs memory
for its table.

### Maps of slices and structs

`examples/ch05/mapvalues/main.go`

```go
// Command mapvalues shows maps whose values are slices or structs.
package main

import "fmt"

// point is a struct: a group of named fields (Chapter 7).
type point struct{ x, y int }

func main() {
	// Group words by their first letter.
	groups := make(map[byte][]string)
	for _, w := range []string{"apple", "bean", "avocado", "beet", "corn"} {
		groups[w[0]] = append(groups[w[0]], w) // nil slice works with append
	}
	fmt.Println(groups['a'], groups['b'], groups['c'], len(groups['z']))

	// Struct values in a map can't be changed in place.
	// Read the value, change the copy, and store it back.
	pos := map[string]point{"start": {0, 0}}
	p := pos["start"]
	p.x = 5
	pos["start"] = p
	fmt.Println(pos["start"])
}
```

```console
$ go run ./ch05/mapvalues
[apple avocado] [bean beet] [corn] 0
{5 0}
```

- `groups[k] = append(groups[k], w)` is the map-of-lists idiom. A missing key gives a `nil` slice,
  which `append` handles.
- You can't assign to a field of a struct in a map: map values aren't addressable (the map may move
  or delete them):

  ```console
  ./main.go:9:2: cannot assign to struct field pos["start"].x in map
  ```

  Copy the value out, change it, store it back. Or store pointers
  ([Chapter 7](07-structs-methods-pointers.md)).

### Sets

Go has no set type. Use a map with the elements as keys.

`examples/ch05/sets/main.go`

```go
// Command sets uses maps as sets.
package main

import "fmt"

func main() {
	// map[string]bool: simple and readable.
	seen := make(map[string]bool)
	for _, w := range []string{"go", "is", "fun", "go", "is"} {
		if seen[w] { // a missing key reads as false
			fmt.Println("duplicate:", w)
			continue
		}
		seen[w] = true
	}
	fmt.Println(len(seen), "distinct words")

	// map[string]struct{}: the empty struct takes no memory.
	allowed := map[string]struct{}{"-c": {}, "-n": {}}
	_, ok := allowed["-n"]
	_, bad := allowed["-x"]
	fmt.Println(ok, bad)
}
```

```console
$ go run ./ch05/sets
duplicate: go
duplicate: is
3 distinct words
true false
```

| Set type | Membership test | Pros |
|----------|-----------------|------|
| `map[T]bool` | `if s[x]` | Reads naturally. |
| `map[T]struct{}` | `if _, ok := s[x]; ok` | Values use no memory; can't hold a misleading `false`. |

Both are common; `map[T]bool` suits most code.

## The slices package

The `slices` package has generic functions for slices of any element type (how generics work:
[Chapter 13](13-generics.md)).

`examples/ch05/slicespkg/main.go`

```go
// Command slicespkg tours the slices package.
package main

import (
	"fmt"
	"slices"
	"strings"
)

func main() {
	words := []string{"pear", "apple", "fig", "banana"}

	fmt.Println(slices.Contains(words, "fig"), slices.Index(words, "fig"))
	fmt.Println(slices.Index(words, "kiwi")) // -1 means not found

	slices.Sort(words) // sorts in place
	fmt.Println(words)
	i, found := slices.BinarySearch(words, "fig") // needs sorted input
	fmt.Println(i, found)

	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
	fmt.Println(slices.Max(nums), slices.Min(nums))
	slices.Reverse(nums)
	fmt.Println(nums)

	// Sort with your own comparison: by length, then alphabetically.
	slices.SortFunc(words, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b) // -1, 0, or +1
	})
	fmt.Println(words)

	fmt.Println(slices.Equal([]int{1, 2}, []int{1, 2}))
	fmt.Println(slices.Insert([]int{1, 4}, 1, 2, 3))
	fmt.Println(slices.Compact([]int{1, 1, 2, 2, 2, 3, 1}))
}
```

```console
$ go run ./ch05/slicespkg
true 2
-1
[apple banana fig pear]
2 true
9 1
[6 2 9 5 1 4 1 3]
[fig pear apple banana]
true
[1 2 3 4]
[1 2 3 1]
```

| Function | Does |
|----------|------|
| `Contains(s, v)`, `Index(s, v)` | Linear search. `Index` returns -1 if not found. |
| `Sort(s)`, `SortFunc(s, cmp)` | Sort in place. `cmp(a, b)` returns negative, zero, or positive. |
| `BinarySearch(s, v)` | Fast search of a sorted slice; returns position and found. |
| `Min(s)`, `Max(s)` | Smallest / largest. Panics on an empty slice. |
| `Reverse(s)` | Reverse in place. |
| `Equal(a, b)` | Same length and elements. |
| `Clone(s)` | Independent copy. |
| `Insert(s, i, v...)`, `Delete(s, i, j)` | Return the modified slice; assign it, like `append`. |
| `Compact(s)` | Drop *consecutive* duplicates (sort first to drop all). |

## The maps package

`examples/ch05/mapspkg/main.go`

```go
// Command mapspkg shows how to visit a map in a stable order.
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	stock := map[string]int{"pears": 3, "apples": 5, "plums": 7, "figs": 2}

	// maps.Keys returns an iterator over the keys (Chapter 14).
	// slices.Sorted collects them into a new, sorted slice.
	for _, name := range slices.Sorted(maps.Keys(stock)) {
		fmt.Println(name, stock[name])
	}

	// slices.Collect gathers an iterator's values into a slice.
	values := slices.Collect(maps.Values(stock))
	slices.Sort(values)
	fmt.Println(values)

	// maps.Clone copies a map; maps.Equal compares two.
	backup := maps.Clone(stock)
	stock["figs"] = 0
	fmt.Println(maps.Equal(stock, backup), backup["figs"])
}
```

```console
$ go run ./ch05/mapspkg
apples 5
figs 2
pears 3
plums 7
[2 3 5 7]
false 2
```

`slices.Sorted(maps.Keys(m))` is the idiom for visiting a map in key order. `maps.Keys` and
`maps.Values` return **iterators**, not slices (write your own in [Chapter 14](14-iterators.md)).
For now, pass them to `slices.Sorted` or `slices.Collect`.

## Built-in functions

**Built-in** functions need no import:

| Built-in | Works on | Does |
|----------|----------|------|
| `len(x)` | string, array, slice, map, channel | number of elements (bytes for strings) |
| `cap(x)` | array, slice, channel | capacity |
| `make(T, ...)` | slice, map, channel | create and initialize |
| `append(s, v...)` | slice | add elements, return the new slice |
| `copy(dst, src)` | slice | copy elements, return the count |
| `delete(m, k)` | map | remove a key |
| `clear(x)` | map, slice | delete all entries / zero all elements |
| `min(...)`, `max(...)` | ordered values | smallest / largest argument |

`examples/ch05/builtins/main.go`

```go
// Command builtins shows clear, min, and max.
package main

import "fmt"

func main() {
	// min and max take any number of arguments of the same ordered type.
	fmt.Println(min(3, 1, 2), max(2.5, 7, 1), min("b", "a"))

	// clear on a map deletes every entry.
	m := map[string]int{"a": 1, "b": 2}
	clear(m)
	fmt.Println(len(m))

	// clear on a slice sets every element to its zero value.
	// The length stays the same.
	s := []int{1, 2, 3}
	clear(s)
	fmt.Println(s, len(s))

	// To empty a slice but keep its capacity, reslice to length 0.
	s = s[:0]
	fmt.Println(len(s), cap(s))
}
```

```console
$ go run ./ch05/builtins
1 7 a
0
[0 0 0] 3
0 3
```

`min` and `max` are built-ins (since Go 1.21); `slices.Min`/`Max` take a slice instead. Don't name
variables `min`, `max`, `len`, or `copy`: it's legal but hides the built-in for the rest of that
scope.

## Building seek: many files and -c

`seek` now takes any number of PATHs and a `-c` option:

- Matches print as `path:text`, so you can tell files apart. (`seek` always shows the path, even
  for one file: simpler than `grep`'s rule, and a stable format.)
- `-c` prints `path:count` for every file instead of the lines, including files with 0 matches.
- At least one PATH is required: `seek: missing PATH`.
- Options are checked against a list; unknown ones are still rejected.

You'll use a **slice** of known options (checked with `slices.Contains`), a **map as a set** of
given options (`opts["-c"]` is `true` if `-c` was given), a **map** from path to count, and
**append** to collect matching lines.

Each changed part of `main.go` is shown in full. The constants, `main`, and `usageError` stay the
same.

Add `"slices"` to the imports:

```go
import (
	"fmt"
	"os"
	"slices"
	"strings"
)
```

Replace `run`:

```go
// run is the whole program. It returns the exit status.
func run(args []string) int {
	pattern, paths, opts, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(paths) == 0 {
		return usageError("missing PATH")
	}

	// match is a closure: it captures pattern from the enclosing function.
	match := func(line string) bool {
		return strings.Contains(line, pattern)
	}

	counts := make(map[string]int) // number of matching lines per path
	total := 0
	for _, path := range paths {
		lines, err := searchFile(path, match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		counts[path] = len(lines)
		total += len(lines)
		if !opts["-c"] {
			for _, line := range lines {
				fmt.Printf("%s:%s\n", path, line)
			}
		}
	}

	if opts["-c"] {
		// Ranging over the map would visit the paths in random order, so
		// we walk the paths slice and look each count up instead.
		for _, path := range paths {
			fmt.Printf("%s:%d\n", path, counts[path])
		}
	}
	if total == 0 {
		return exitNoMatch
	}
	return exitOK
}
```

Replace `parseArgs`, and add the `boolOptions` variable above it:

```go
// boolOptions lists the on/off options seek understands.
var boolOptions = []string{"-c"}

// parseArgs splits the command line into options, a pattern, and paths.
// The options that were given are returned as a set: opts["-c"] is true
// if -c was given. If the program should stop right away (after --help,
// --version, or a usage error), done is true and code is the exit status.
func parseArgs(args []string) (pattern string, paths []string, opts map[string]bool, code int, done bool) {
	opts = make(map[string]bool)
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return "", nil, nil, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return "", nil, nil, exitOK, true
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				if !slices.Contains(boolOptions, arg) {
					return "", nil, nil, usageError("unknown option " + arg), true
				}
				opts[arg] = true
				continue
			}
			break loop
		}
	}
	if i >= len(args) {
		return "", nil, nil, usageError("missing PATTERN"), true
	}
	return args[i], args[i+1:], opts, 0, false
}
```

Replace `searchFile`:

```go
// searchFile returns the lines of the file at path for which match
// returns true.
func searchFile(path string, match func(string) bool) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	var matches []string // a nil slice is ready to append to
	for _, line := range strings.Split(text, "\n") {
		if match(line) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}
```

How it fits together:

- **`boolOptions`** is a package-level slice, so it uses `var` (`:=` only works in functions). It's
  a list, not a set: Chapter 6 adds options, and a short list reads easily. `slices.Contains`
  scans it.
- **`parseArgs`** gains a fifth result, `opts map[string]bool`. As a named result it starts as a
  `nil` map, and writing to that panics, so the first line calls `make`. A known option is recorded
  with `opts[arg] = true`, then `continue` moves to the next argument (inside a `switch`,
  `continue` still applies to the loop).
- **`opts["-c"]`** reads the set. Missing keys read as `false`, so no comma-ok is needed.
- **`searchFile`** no longer prints. It returns the matching lines and `run` decides how to show
  them. `var matches []string` starts `nil`; `append` grows it. Separating "find" from "print" makes
  `-c` easy.
- **`args[i+1:]`** is a sub-slice sharing memory with `os.Args`. That's fine; nothing modifies it.
- **`counts`** maps each path to its match count. For `-c`, `run` loops over the `paths` **slice**
  (the order the user typed) and looks each count up. Looping over the map would give random order.
- The first file error still stops the search with exit status 2. [Chapter 9](09-errors.md) makes
  `seek` report it and keep going.

### Checkpoint

```console
$ go run . fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
$ go run . -c fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:2
testdata/notes.md:0
$ go run . dog testdata/notes.md
exit status 1
$ go run . -x fox testdata/poem.txt
seek: unknown option -x
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run . fox
seek: missing PATH
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
```

`notes.md` has `FOX` in capitals, so it doesn't match yet. With the binary:

```console
$ go build
$ ./seek coffee testdata/poem.txt testdata/notes.md; echo $?
testdata/notes.md:- buy coffee
0
$ ./seek -c cat testdata/poem.txt testdata/notes.md; echo $?
testdata/poem.txt:0
testdata/notes.md:0
1
```

The complete project at this point is in `project/ch05/`.

## Summary

- Arrays: fixed length, part of the type. Values: assignment copies them.
- A slice is a header (pointer, length, capacity) over an array. Slicing shares the array.
- Always `s = append(s, ...)`. Preallocate with `make([]T, 0, n)` when you know the size.
- Appending to a sub-slice can overwrite the original. Limit capacity with `s[low:high:max]`, or
  copy (`copy`, `slices.Clone`).
- A `nil` slice is a valid empty slice. Test emptiness with `len(s) == 0`.
- Maps: `make` or a literal before writing (writing to `nil` panics); missing keys read as zero;
  `v, ok := m[k]` tells them apart.
- Map order is random. Use `slices.Sorted(maps.Keys(m))` or a separate slice for order.
- Sets: `map[T]bool` or `map[T]struct{}`.
- `slices` and `maps` handle common jobs: search, sort, compare, clone, collect keys.
- Built-ins: `len`, `cap`, `make`, `append`, `copy`, `delete`, `clear`, `min`, `max`.

## Exercises

1. Write `uniq(words []string) []string`: the words without duplicates, keeping each first
   occurrence, in order.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   // uniq returns words without duplicates, keeping the first occurrence.
   func uniq(words []string) []string {
   	seen := make(map[string]bool)
   	var out []string
   	for _, w := range words {
   		if !seen[w] {
   			seen[w] = true
   			out = append(out, w)
   		}
   	}
   	return out
   }

   func main() {
   	fmt.Println(uniq([]string{"b", "a", "b", "c", "a"}))
   	fmt.Println(uniq(nil) == nil)
   }
   ```

   ```console
   $ go run .
   [b a c]
   true
   ```

   The map remembers what you've seen; the slice remembers order.

   </details>

2. Count each word in `"the cat and the dog and the bird"` and print the counts alphabetically.
   `strings.Fields(s)` splits a string into words.

   <details><summary>Solution</summary>

   ```go
   package main

   import (
   	"fmt"
   	"maps"
   	"slices"
   	"strings"
   )

   func main() {
   	text := "the cat and the dog and the bird"
   	counts := make(map[string]int)
   	for _, w := range strings.Fields(text) {
   		counts[w]++
   	}
   	for _, w := range slices.Sorted(maps.Keys(counts)) {
   		fmt.Printf("%-4s %d\n", w, counts[w])
   	}
   }
   ```

   ```console
   $ go run .
   and  2
   bird 1
   cat  1
   dog  1
   the  3
   ```

   </details>

3. Predict the output, then run it:

   ```go
   package main

   import "fmt"

   func main() {
   	a := make([]int, 3, 4)
   	b := append(a, 1)
   	c := append(a, 2)
   	a[0] = 9
   	fmt.Println(a, b, c)
   }
   ```

   <details><summary>Solution</summary>

   ```console
   $ go run .
   [9 0 0] [9 0 0 2] [9 0 0 2]
   ```

   `a` has room for one more element, so both appends write the **same** slot (index 3); the second
   overwrites the first. `b` and `c` share `a`'s array, so `a[0] = 9` shows in all three.
   `b := append(a[:3:3], 1)` would give `b` its own array.

   </details>

4. Write `filter(nums []int, keep func(int) bool) []int`, returning a new slice of the elements
   where `keep` is true, without changing `nums`. Use it with two closures: one keeps even numbers,
   one keeps numbers up to a `limit` variable.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   // filter returns the elements of nums for which keep returns true.
   func filter(nums []int, keep func(int) bool) []int {
   	out := make([]int, 0, len(nums))
   	for _, n := range nums {
   		if keep(n) {
   			out = append(out, n)
   		}
   	}
   	return out
   }

   func main() {
   	nums := []int{1, 2, 3, 4, 5, 6}
   	even := filter(nums, func(n int) bool { return n%2 == 0 })
   	limit := 3
   	small := filter(nums, func(n int) bool { return n <= limit })
   	fmt.Println(even, small, nums)
   }
   ```

   ```console
   $ go run .
   [2 4 6] [1 2 3] [1 2 3 4 5 6]
   ```

   `make([]int, 0, len(nums))` preallocates, so `append` never reallocates. `slices.DeleteFunc` is
   the standard in-place version.

   </details>

5. In `seek`, make the `-c` loop range over the `counts` map instead of the `paths` slice. Build
   and run `./seek -c fox testdata/poem.txt testdata/notes.md main.go go.mod` several times. What
   happens? Undo the change afterward.

   <details><summary>Solution</summary>

   ```go
   		for path, n := range counts {
   			fmt.Printf("%s:%d\n", path, n)
   		}
   ```

   The order changes between runs. One run printed:

   ```text
   testdata/poem.txt:2
   testdata/notes.md:0
   main.go:0
   go.mod:0
   ```

   and another:

   ```text
   main.go:0
   go.mod:0
   testdata/poem.txt:2
   testdata/notes.md:0
   ```

   Your orders will differ. Randomly changing output is useless for scripts and tests, so `seek`
   loops over the `paths` slice.

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-5--arrays-slices-and-maps)
