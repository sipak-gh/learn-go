# Chapter 6 — Strings, Bytes, and Runes

`seek` is all about text. This chapter shows what a Go string really is (UTF-8 bytes), how bytes
differ from runes, and the text packages `strings`, `bytes`, `unicode/utf8`, `strconv`, and `fmt`.
Then `seek` gets `-i` (ignore case), `-n` (line numbers), `-v` (invert), and colored highlighting.
[Chapter 5](05-collections.md) covered slices and maps; [Chapter 7](07-structs-methods-pointers.md)
tidies the code with structs.

## Strings are read-only bytes

A Go string is an immutable sequence of **bytes**, not characters.

- `len(s)` counts bytes; `s[i]` is a `byte` (alias for `uint8`).
- `s[i:j]` is a new string sharing the same bytes (no copy).
- `+` and `+=` build a new string; the old one never changes.
- `==`, `<`, and friends compare byte by byte.

`examples/ch06/bytes-of-a-string/main.go`

```go
// A string is a read-only sequence of bytes.
package main

import "fmt"

func main() {
	s := "hello"
	fmt.Println(len(s))      // number of bytes
	fmt.Println(s[0])        // indexing gives a byte (a uint8)
	fmt.Printf("%c\n", s[0]) // %c prints a byte or rune as a character
	fmt.Println(s[1:4])      // slicing gives a string (shares the bytes)

	t := s         // copies the string header, not the bytes
	s += ", world" // builds a NEW string; t still sees the old one
	fmt.Println(s)
	fmt.Println(t)

	fmt.Println("go" < "gopher", "a" == "a") // strings compare byte by byte
}
```

```console
$ go run ./ch06/bytes-of-a-string
5
104
h
ell
hello, world
hello
true true
```

Why immutable? So strings can be shared freely. Assigning, slicing, or passing one copies only a
small header (pointer and length), never the text, and no one needs a defensive copy.

You can't change a byte in place:

```go
s := "hello"
s[0] = 'H'
```

```console
./main.go:7:2: cannot assign to s[0] (neither addressable nor a map index expression)
```

Instead, build a new string (`strings` functions or `strings.Builder`), or convert to `[]byte`,
change that, and convert back.

## UTF-8: bytes versus runes

Go source and string literals are UTF-8. UTF-8 stores each Unicode **code point** in 1 to 4 bytes:
ASCII letters 1, `é` 2, `世` 3, `😀` 4.

Go calls a code point a **rune** (alias for `int32`). Rune literals use single quotes: `'é'`.

| You write | You get |
|-----------|---------|
| `s[i]` | the byte at byte offset `i` |
| `for i, b := range []byte(s)` | each byte |
| `for i, r := range s` | each rune, decoded; `i` is its **byte** offset |
| `[]rune(s)` | a new slice of all the runes |
| `[]byte(s)` | a new slice with a copy of the bytes |
| `string(r)` / `string(runes)` / `string(bytes)` | a string built from them |

`examples/ch06/runes/main.go`

```go
// Bytes versus runes in a UTF-8 string.
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "héllo, 世界"
	fmt.Println("bytes:", len(s))
	fmt.Println("runes:", utf8.RuneCountInString(s))

	// Indexing reads single bytes. 'é' is two bytes: 0xc3 0xa9.
	fmt.Printf("s[1] = %#x, s[2] = %#x\n", s[1], s[2])

	// range decodes one rune at a time. i is the BYTE offset of the rune.
	for i, r := range s {
		fmt.Printf("%2d %c %U\n", i, r, r)
	}

	// Conversions.
	rs := []rune(s) // decode into a slice of code points
	bs := []byte(s) // copy the raw bytes
	fmt.Println(len(rs), len(bs))
	fmt.Println(string(rs[7:]))       // back to a string
	fmt.Println(string(rune(0x4e16))) // one code point to a string

	// Invalid UTF-8 decodes as U+FFFD, the replacement character.
	bad := "a\xffb"
	for _, r := range bad {
		fmt.Printf("%q ", r)
	}
	fmt.Println()
}
```

```console
$ go run ./ch06/runes
bytes: 14
runes: 9
s[1] = 0xc3, s[2] = 0xa9
 0 h U+0068
 1 é U+00E9
 3 l U+006C
 4 l U+006C
 5 o U+006F
 6 , U+002C
 7   U+0020
 8 世 U+4E16
11 界 U+754C
9 14
世界
世
'a' '�' 'b' 
```

Offsets jump from 1 to 3 (after `é`) and from 8 to 11 (after `世`).

- **A string can hold any bytes**, valid UTF-8 or not. `range` turns each bad byte into `U+FFFD`
  (`�`), steps one byte, and never panics.
- **Slicing works on bytes.** `s[0:2]` of `"héllo"` is `"h\xc3"`, half an `é`. Slice only at
  offsets from `range`, `strings.Index`, and the like.
- **`[]rune(s)` costs 4 bytes per rune.** Use it for random access to runes (say, reversing);
  otherwise use `range`.
- **`string(65)` is `"A"`, not `"65"`.** An integer converts to a rune. `go vet` flags it:

```console
./main.go:7:14: conversion from int to string yields a string of one rune, not a string of digits
```

Use `strconv.Itoa(65)` for `"65"` (see [strconv](#strconv-strings-to-numbers-and-back)).

Why bytes? They're simple, fast, and match what files and networks carry, with no conversion step.
`range` and `unicode/utf8` decode runes when you need them.

## The `unicode/utf8` and `unicode` packages

`unicode/utf8` encodes and decodes UTF-8 by hand; `unicode` classifies single runes.

| Function | Does |
|----------|------|
| `utf8.RuneCountInString(s)` | number of runes in `s` |
| `utf8.DecodeRuneInString(s)` | the first rune of `s` and its size in bytes |
| `utf8.RuneLen(r)` | bytes needed to encode `r` |
| `utf8.ValidString(s)` | whether `s` is valid UTF-8 |
| `utf8.AppendRune(b, r)` | appends the encoding of `r` to `b` |
| `unicode.IsLetter`, `IsDigit`, `IsSpace`, `IsUpper`, `IsLower`, `IsPunct` | classify a rune |
| `unicode.ToUpper`, `ToLower` | convert one rune |

`examples/ch06/utf8/main.go`

```go
// The unicode/utf8 and unicode packages.
package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func main() {
	s := "€5"

	// Decode the first rune by hand: the rune and its size in bytes.
	r, size := utf8.DecodeRuneInString(s)
	fmt.Printf("%c is %d bytes\n", r, size)
	fmt.Println(utf8.RuneLen('a'), utf8.RuneLen('é'), utf8.RuneLen('€'), utf8.RuneLen('😀'))

	// Check that bytes are valid UTF-8.
	fmt.Println(utf8.ValidString(s), utf8.ValidString("\xff"))

	// Encode a rune by appending its bytes to a byte slice.
	buf := utf8.AppendRune(nil, '世')
	fmt.Printf("% x\n", buf)

	// The unicode package classifies and converts single runes.
	for _, r := range "aZ9 ß!" {
		fmt.Printf("%q letter=%t upper=%t digit=%t space=%t -> %q\n",
			r, unicode.IsLetter(r), unicode.IsUpper(r),
			unicode.IsDigit(r), unicode.IsSpace(r), unicode.ToUpper(r))
	}
}
```

```console
$ go run ./ch06/utf8
€ is 3 bytes
1 2 3 4
true false
e4 b8 96
'a' letter=true upper=false digit=false space=false -> 'A'
'Z' letter=true upper=true digit=false space=false -> 'Z'
'9' letter=false upper=false digit=true space=false -> '9'
' ' letter=false upper=false digit=false space=true -> ' '
'ß' letter=true upper=false digit=false space=false -> 'ß'
'!' letter=false upper=false digit=false space=false -> '!'
```

`DecodeRuneInString` walks a string rune by rune when `range` doesn't fit, such as stepping through
two strings at once. `seek` does this below.

## String literals: interpreted and raw

- **Interpreted** (`"..."`) understands escapes: `\n` newline, `\t` tab, `\"` quote, `\\`
  backslash, `\xNN` one byte, `\uNNNN` and `\UNNNNNNNN` a code point (as UTF-8).
- **Raw** (`` `...` ``): no escapes, may span lines, can't contain a backquote. Good for regular
  expressions, Windows paths, and multi-line text.

`examples/ch06/literals/main.go`

```go
// Interpreted and raw string literals.
package main

import "fmt"

func main() {
	// Interpreted literals use double quotes and understand escapes.
	fmt.Println("tab:\t|  quote:\"  backslash:\\")
	fmt.Println("caf\u00e9 \U0001F600 \x41") // \u: 4 hex digits, \U: 8, \x: one byte

	// Raw literals use backquotes. No escapes; newlines are kept as-is.
	path := `C:\Users\gopher` // handy for Windows paths and regexps
	fmt.Println(path)

	poem := `roses are red,
  \n stays \n here`
	fmt.Println(poem)

	// ANSI escape: ESC (0x1b) starts a terminal control sequence.
	fmt.Printf("%q\n", "\x1b[1mbold\x1b[0m")
}
```

```console
$ go run ./ch06/literals
tab:	|  quote:"  backslash:\
café 😀 A
C:\Users\gopher
roses are red,
  \n stays \n here
"\x1b[1mbold\x1b[0m"
```

The last line previews `seek`'s color. Most terminals read `ESC [ ... m` as a style command:
`\x1b[1;31m` is bold red; `\x1b[0m` resets.

## A tour of the `strings` package

`strings` functions never change their input; they return new strings or offsets. Offsets are
**byte** offsets; `-1` means "not found".

`examples/ch06/strings-tour/main.go`

```go
// A tour of the strings package.
package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "  The quick brown fox  "

	// Searching.
	fmt.Println(strings.Contains(s, "quick")) // true
	fmt.Println(strings.Index(s, "o"))        // first byte offset, or -1
	fmt.Println(strings.LastIndex(s, "o"))    // last byte offset, or -1
	fmt.Println(strings.Count(s, "o"))        // non-overlapping count
	fmt.Println(strings.HasPrefix("seek.go", "seek"), strings.HasSuffix("seek.go", ".go"))

	// Trimming.
	fmt.Printf("%q\n", strings.TrimSpace(s))
	fmt.Printf("%q\n", strings.Trim("--x--", "-"))      // any of the chars in "-"
	fmt.Printf("%q\n", strings.TrimPrefix("v1.2", "v")) // only if present

	// Splitting and joining.
	fmt.Printf("%q\n", strings.Fields(s))            // split on runs of spaces
	fmt.Printf("%q\n", strings.Split("a,b,,c", ",")) // keeps empty fields
	fmt.Printf("%q\n", strings.SplitN("a=b=c", "=", 2))
	fmt.Println(strings.Join([]string{"x", "y", "z"}, "/"))

	// Cut splits around the first separator.
	key, value, found := strings.Cut("color=red", "=")
	fmt.Println(key, value, found)

	// Changing (always returns a new string).
	fmt.Println(strings.ReplaceAll("oink oink", "k", "ky"))
	fmt.Println(strings.Replace("oink oink", "oink", "moo", 1)) // n = how many
	fmt.Println(strings.ToUpper("go"), strings.ToLower("GO"), strings.Repeat("ab", 3))

	// Comparing without regard to case.
	fmt.Println(strings.EqualFold("Gopher", "GOPHER"))
}
```

```console
$ go run ./ch06/strings-tour
true
14
19
2
true true
"The quick brown fox"
"x"
"1.2"
["The" "quick" "brown" "fox"]
["a" "b" "" "c"]
["a" "b=c"]
x/y/z
color red true
oinky oinky
moo oink
GO go ababab
true
```

- **`Fields` vs `Split`.** `Fields` splits on runs of white space and drops empty pieces. `Split`
  uses an exact separator and keeps them. `Split("", ",")` returns `[""]`, not an empty slice.
- **`Cut` replaces most `Index` + slicing code.** It returns the text before and after the first
  separator, and whether it was found.
- **`Trim` takes a set of characters, not a prefix.** `Trim("--x--", "-")` strips every leading and
  trailing `-`. For one exact prefix or suffix, use `TrimPrefix` or `TrimSuffix`.
- Variations: `ContainsRune`, `IndexRune`, `IndexByte`, `ContainsAny`, `IndexFunc`. `go doc strings`
  lists everything.

## Building strings: `strings.Builder`

Each `+=` copies everything so far, so a loop does quadratic work. A `strings.Builder` appends into
a growing buffer and makes the string once.

`examples/ch06/builder/main.go`

```go
// Building strings efficiently with strings.Builder.
package main

import (
	"fmt"
	"strings"
)

func main() {
	var b strings.Builder // the zero value is ready to use
	for i := 3; i > 0; i-- {
		fmt.Fprintf(&b, "%d...", i) // Fprintf can write into a Builder
	}
	b.WriteString("liftoff")
	b.WriteByte('!')
	b.WriteRune('🚀')
	fmt.Println(b.String())
	fmt.Println(b.Len(), "bytes")

	// The same with +=. Each += copies everything built so far, so a
	// loop of n steps does O(n²) work. Fine for a few pieces, slow for many.
	s := ""
	for i := 3; i > 0; i-- {
		s += fmt.Sprint(i) + "..."
	}
	fmt.Println(s + "liftoff!🚀")
}
```

```console
$ go run ./ch06/builder
3...2...1...liftoff!🚀
24 bytes
3...2...1...liftoff!🚀
```

- The zero value `var b strings.Builder` is ready to use; Go designs types that way.
- `fmt.Fprintf(&b, ...)` writes formatted text into it. The `&` makes a pointer
  ([Chapter 7](07-structs-methods-pointers.md)).
- Don't copy a `Builder` after writing to it; pass `&b`.
- To join an existing slice, `strings.Join` is simpler.

## The `bytes` package and `bytes.Buffer`

Files and networks give you `[]byte`. The `bytes` package mirrors `strings` for byte slices
(`Contains`, `Index`, `Split`, `Fields`, `TrimSpace`, `Cut`, `EqualFold`, ...), so you skip copying
into a string and back. `bytes.Buffer` is a growable buffer you can write into (like a `Builder`)
and read from.

`examples/ch06/bytes/main.go`

```go
// The bytes package mirrors strings for []byte.
package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	data := []byte("name: gopher\nrole: mascot\n")

	// Same functions as strings, but for byte slices.
	fmt.Println(bytes.Contains(data, []byte("gopher")))
	fmt.Println(bytes.Count(data, []byte("\n")))
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		k, v, _ := bytes.Cut(line, []byte(": "))
		fmt.Printf("%s=%s\n", k, v) // %s prints a []byte as text
	}
	fmt.Println(bytes.Equal([]byte("a"), []byte("a"))) // slices can't use ==

	// Unlike strings, byte slices are mutable.
	word := []byte("gopher")
	word[0] = 'G'
	fmt.Println(string(word))

	// bytes.Buffer is a growable buffer you can write to and read from.
	var buf bytes.Buffer
	buf.WriteString("hello ")
	fmt.Fprintf(&buf, "%d gophers\n", 3)
	fmt.Print(buf.String())
	buf.WriteTo(os.Stdout) // drains the buffer
	fmt.Println(buf.Len()) // 0: everything was read
}
```

```console
$ go run ./ch06/bytes
true
2
name=gopher
role=mascot
true
Gopher
hello 3 gophers
hello 3 gophers
0
```

Rules of thumb:

- Text you keep, compare, or use as a map key: `string`.
- Data you read, change, or pass to I/O: `[]byte`.
- Building a string: `strings.Builder`. Building bytes, or a buffer you also read: `bytes.Buffer`.
  Tests often capture output in one, as seek's do in [Chapter 19](19-testing.md).

## `strconv`: strings to numbers and back

`strconv` converts strings to and from numbers and booleans. Every `Parse` function returns an
error, since input can be wrong.

| Function | Does |
|----------|------|
| `strconv.Atoi(s)` | string → `int` ("ASCII to integer") |
| `strconv.Itoa(n)` | `int` → string |
| `strconv.ParseInt(s, base, bits)` / `ParseUint` | string → `int64` / `uint64`, any base |
| `strconv.ParseFloat(s, bits)` | string → `float64` |
| `strconv.ParseBool(s)` | `"1"`, `"t"`, `"true"`, ... → `true` |
| `strconv.FormatInt(n, base)`, `FormatFloat`, `FormatBool` | the other way |
| `strconv.Quote(s)` / `Unquote(s)` | add or remove Go quotes and escapes |

`examples/ch06/strconv/main.go`

```go
// Converting between strings and other types with strconv.
package main

import (
	"fmt"
	"strconv"
)

func main() {
	// Strings to numbers. Every Parse function also returns an error.
	n, err := strconv.Atoi("42")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("forty-two")
	fmt.Println(err)

	big, _ := strconv.ParseInt("-ff", 16, 64) // base 16, fits in 64 bits
	f, _ := strconv.ParseFloat("2.5e3", 64)
	b, _ := strconv.ParseBool("true") // also "1", "t", "T", "TRUE", "True"
	fmt.Println(big, f, b)

	_, err = strconv.ParseInt("300", 10, 8) // too big for 8 bits
	fmt.Println(err)

	// Numbers to strings.
	s := strconv.Itoa(99) + " bottles"
	fmt.Println(s)
	fmt.Println(strconv.FormatInt(255, 2), strconv.FormatFloat(3.14159, 'f', 2, 64))

	// string(n) does NOT format a number: it makes the rune with that code.
	fmt.Println(string(rune(65)), strconv.Itoa(65))

	// Quote adds quotes and escapes; Unquote reverses it.
	q := strconv.Quote("tab\there \"hi\"")
	fmt.Println(q)
	u, _ := strconv.Unquote(q)
	fmt.Println(u)
}
```

```console
$ go run ./ch06/strconv
43 <nil>
strconv.Atoi: parsing "forty-two": invalid syntax
-255 2500 true
strconv.ParseInt: parsing "300": value out of range
99 bottles
11111111 3.14
A 65
"tab\there \"hi\""
tab	here "hi"
```

- Errors name the function and quote the input. `seek` shows one in [Chapter 9](09-errors.md).
- `_, err =` and `big, _ :=` ignore a result with the blank identifier. Do that only when you know
  the input is good.
- `fmt.Sprint(n)` also formats a number, but `strconv` is faster and clearer for single values.

## `fmt` in depth

You've used `Println` and `Printf` since [Chapter 1](01-getting-started.md). The family has three
destinations:

| Prints to stdout | Returns a string | Writes to a destination you choose |
|------------------|------------------|-------------------------------------|
| `Print`, `Println`, `Printf` | `Sprint`, `Sprintln`, `Sprintf` | `Fprint`, `Fprintln`, `Fprintf` |

`Fprintf(os.Stderr, ...)` writes to standard error; `Fprintf(&b, ...)` to a builder or buffer. Any
type with the right `Write` method works ([Chapter 10](10-interfaces.md)). `fmt.Errorf` builds an
error ([Chapter 9](09-errors.md)).

A verb is `%`, optional flags, width, and precision, then a letter. `%-8.3s` is "left-align, width
8, at most 3 characters, as a string".

`examples/ch06/fmt-verbs/main.go`

```go
// fmt in depth: verbs, flags, width, and precision.
package main

import (
	"fmt"
	"os"
)

// A struct groups named fields; Chapter 7 covers structs fully.
type point struct {
	X, Y int
}

func main() {
	p := point{1, 2}
	fmt.Printf("%v | %+v | %#v | %T\n", p, p, p, p)

	// Width and alignment.
	fmt.Printf("[%5d] [%-5d] [%05d]\n", 42, 42, 42)
	fmt.Printf("[%8s] [%-8s]\n", "go", "go")

	// Precision: digits after the point, or max characters of a string.
	fmt.Printf("[%.2f] [%8.3f] [%.3s]\n", 3.14159, 3.14159, "gopher")
	fmt.Printf("[%*d]\n", 6, 7) // width taken from an argument

	// Other useful verbs.
	fmt.Printf("%x %X %o %b %e\n", 255, 255, 8, 5, 1234.5)
	fmt.Printf("%c %U %q\n", 'G', 'G', 'G')
	fmt.Printf("%q %x % x\n", "hi\n", "hi", "hi")
	fmt.Printf("%t %v\n", true, nil)
	fmt.Printf("100%%\n")

	// Sprintf returns the string; Fprintf writes to any destination.
	msg := fmt.Sprintf("%s has %d legs", "spider", 8)
	fmt.Println(msg)
	fmt.Fprintf(os.Stdout, "%-6s|%6s|\n", "left", "right")

	// Print adds spaces only between operands that are not strings;
	// Println always adds spaces and a newline.
	fmt.Print("a", "b", 1, 2, "\n")
	fmt.Println("a", "b", 1, 2)
}
```

```console
$ go run ./ch06/fmt-verbs
{1 2} | {X:1 Y:2} | main.point{X:1, Y:2} | main.point
[   42] [42   ] [00042]
[      go] [go      ]
[3.14] [   3.142] [gop]
[     7]
ff FF 10 101 1.234500e+03
G U+0047 'G'
"hi\n" 6869 68 69
true <nil>
100%
spider has 8 legs
left  | right|
ab1 2
a b 1 2
```

The verbs you'll use most:

| Verb | Prints |
|------|--------|
| `%v` | any value in its default format |
| `%+v` | like `%v`, plus struct field names |
| `%#v` | Go syntax for the value (great for debugging) |
| `%T` | the value's type |
| `%d` `%b` `%o` `%x` `%X` | integer in base 10, 2, 8, 16 |
| `%f` `%e` `%g` | float: decimal, exponent, shortest |
| `%s` `%q` | string (or `[]byte`), quoted string |
| `%c` `%U` | rune as a character, as `U+XXXX` |
| `%t` | bool |
| `%p` | pointer address |
| `%%` | a literal `%` |

| Flag | Means |
|------|-------|
| `-` | left-align within the width |
| `0` | pad numbers with zeros |
| `+` | always print a sign (`%+d`); field names (`%+v`) |
| `#` | alternate form: `0x` prefix (`%#x`), Go syntax (`%#v`) |
| space | `% x` puts spaces between bytes; `% d` leaves room for a sign |

Width and precision count **runes**, not bytes, so `%-8s` aligns `"café"` correctly.

A wrong verb doesn't crash; `fmt` prints a marker like `%!d(string=gopher)` or `%!s(MISSING)`.
`go vet` catches these before you run:

```go
name := "gopher"
fmt.Printf("hello %d\n", name)
fmt.Printf("%s and %s\n", name)
```

```console
$ go vet .
./main.go:7:20: fmt.Printf format %d has arg name of wrong type string
./main.go:8:21: fmt.Printf format %s reads arg #2, but call has 1 arg
```

## Unicode case: the gotchas

- **Changing case can change byte length.** The Kelvin sign `K` (U+212A, 3 bytes) lowercases
  to `k` (1 byte). So offsets found in a lowercased copy may not fit the original, and you can slice
  mid-rune.
- **Some letters have no single-letter upper case.** `strings.ToUpper("ß")` stays `ß`, and
  `strings.EqualFold("Straße", "STRASSE")` is false. Go uses *simple* case folding: one rune to one
  rune.
- **Use `strings.EqualFold` for "equal ignoring case".** It compares rune by rune and allocates
  nothing. `strings.ToLower(a) == strings.ToLower(b)` builds two new strings.

`examples/ch06/casefold/main.go`

```go
// Case conversion can change a string's length in bytes.
package main

import (
	"fmt"
	"strings"
)

func main() {
	// The Kelvin sign K (U+212A) lowercases to an ordinary k.
	kelvin := "\u212A"
	lower := strings.ToLower(kelvin)
	fmt.Println(kelvin, len(kelvin), "->", lower, len(lower))

	// So offsets found in a lowercased copy may not fit the original.
	line := "5\u212A run"
	i := strings.Index(strings.ToLower(line), "run")
	fmt.Printf("found at %d; original[%d:%d] = %q\n", i, i, i+3, line[i:i+3])

	// EqualFold compares without building lowered copies.
	fmt.Println(strings.EqualFold("\u212A", "k"), strings.EqualFold("Straße", "STRASSE"))

	// Some letters have no single-letter upper case.
	fmt.Println(strings.ToUpper("straße"))
}
```

```console
$ go run ./ch06/casefold
K 3 -> k 1
found at 3; original[3:6] = "\xaa r"
true false
STRAßE
```

`"run"` is at offset 5 in the original, but the lowercased copy says 3. `seek` avoids this bug next.
Full Unicode case mapping (`ß` → `SS`) lives in the `golang.org/x/text/cases` module, outside the
standard library; [Chapter 8](08-packages-and-modules.md) shows how to add a module.

## Building seek: `-i`, `-n`, `-v`, and highlighting

| Option | Effect |
|--------|--------|
| `-i` | ignore case (Unicode-correct) |
| `-n` | print `path:N:text`, where N is the line number |
| `-v` | invert: select lines that do **not** match |
| `--color` | highlight matches in bold red (always on, for now; [Chapter 12](12-flags.md) adds `auto`) |

The core is `findAll`, which returns the byte offsets `{start, end}` of each match in a line.
Matching uses it (any match?), and so does highlighting (wrap each match in color codes).

Why not lowercase line and pattern and call `strings.Index`? Lowercasing can change byte lengths, so
the offsets would be wrong in the original line you highlight. Instead, `indexFold` walks the
original line rune by rune using `strings.EqualFold`, so its offsets always fit.

### Step 1: register the options

In `main.go`, add `"unicode/utf8"` to the imports:

```go
import (
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
)
```

Then extend `boolOptions`:

```go
// boolOptions lists the on/off options seek understands.
var boolOptions = []string{"-c", "-i", "-n", "-v", "--color"}
```

`parseArgs` already puts every listed option in the `opts` set, so it needs no change.

### Step 2: find matches by byte offset

Add these functions at the end of `main.go`:

```go
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

- **`findAll`** searches `line[pos:]`, records the match, and moves `pos` past it. `n` caps the
  matches: `1` asks "any match?" (cheap); `-1` gets all (for highlighting). Each span is a
  two-element `[]int`, so the result is a slice of slices.
- **`indexFold`** tries each rune start in `s`. `for start := range s` yields rune byte offsets, so
  it never starts mid-rune.
- **`prefixFold`** steps through `prefix` with `range` and through `s` with
  `utf8.DecodeRuneInString` together; they may advance by different byte counts. It compares rune
  by rune (fast `==`, then `EqualFold`) and returns how many bytes of `s` matched.

### Step 3: highlight the matches

Add these at the end of `main.go` too:

```go
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

Text between matches is copied as is; each match is wrapped in `colorStart`...`colorReset`. The
`strings.Builder` avoids a new string per piece.

### Step 4: return line numbers from `searchFile`

`-n` needs line numbers. Replace `searchFile` with:

```go
// searchFile returns the lines of the file at path for which match
// returns true, along with their line numbers (counting from 1).
func searchFile(path string, match func(string) bool) (nums []int, lines []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	for i, line := range strings.Split(text, "\n") {
		if match(line) {
			nums = append(nums, i+1)
			lines = append(lines, line)
		}
	}
	return nums, lines, nil
}
```

It now returns two **parallel slices**: `nums[k]` is the line number of `lines[k]`. That's clumsy:
nothing keeps them in step, and each new field needs another slice.
[Chapter 7](07-structs-methods-pointers.md) replaces them with a `Match` struct.

### Step 5: use the new options in `run`

Replace `run`. What changed: the `match` closure, the `-v` wrapper, and the printing loop.

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

	ignoreCase := opts["-i"]
	match := func(line string) bool {
		return findAll(line, pattern, ignoreCase, 1) != nil
	}
	if opts["-v"] {
		// Wrap the original closure in a new one that inverts its answer.
		contains := match
		match = func(line string) bool { return !contains(line) }
	}

	counts := make(map[string]int) // number of matching lines per path
	total := 0
	for _, path := range paths {
		nums, lines, err := searchFile(path, match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "seek:", err)
			return exitError
		}
		counts[path] = len(lines)
		total += len(lines)
		if opts["-c"] {
			continue
		}
		for i, line := range lines {
			if opts["--color"] {
				line = highlight(line, findAll(line, pattern, ignoreCase, -1))
			}
			if opts["-n"] {
				fmt.Printf("%s:%d:%s\n", path, nums[i], line)
			} else {
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

- **`-v` wraps a closure in a closure.** `contains := match` saves the original; the new `match`
  negates it. The copy is required: `match = func(line string) bool { return !match(line) }` would
  call *itself* forever, because a closure sees the variable, not its old value.
- **With `-v`, the selected lines are the non-matching ones.** So `-c -v` counts them, and
  `-v --color` has nothing to highlight.

### Checkpoint

```console
$ go run . -i fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/poem.txt:Foxes are quick and clever.
testdata/notes.md:- feed the FOX
$ go run . -n -i fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/notes.md:2:- feed the FOX
$ go run . -v -n fox testdata/poem.txt
testdata/poem.txt:2:jumps over the lazy dog.
testdata/poem.txt:4:Foxes are quick and clever.
$ go run . -c -v fox testdata/poem.txt testdata/notes.md
testdata/poem.txt:2
testdata/notes.md:4
$ go run . -i CAFÉ testdata/notes.md
testdata/notes.md:- café opens at 9
$ go run . -i -n 'café' testdata/notes.md
testdata/notes.md:3:- café opens at 9
$ go run . --color -i fox testdata/poem.txt | cat -v
testdata/poem.txt:The quick brown ^[[1;31mfox^[[0m
testdata/poem.txt:The dog sleeps; the ^[[1;31mfox^[[0m runs.
testdata/poem.txt:^[[1;31mFox^[[0mes are quick and clever.
$ go run . -n -i -v e testdata/poem.txt testdata/notes.md
exit status 1
```

- `CAFÉ` finds `café`. (`É` and `é` are both 2 bytes, but `indexFold` doesn't rely on that.)
- `cat -v` shows the escape bytes (`^[` is ESC). Without it, your terminal shows `fox` in bold red.
- Every line in both files has an `e`, so `-v e` selects nothing and `seek` exits with status 1.

The complete project at this point is in `project/ch06/`.

## Summary

- A string is an immutable byte sequence, usually UTF-8. `len` counts bytes; `s[i]` is a byte.
- A rune is a code point (`int32`). `for i, r := range s` decodes runes; `i` is a byte offset.
- Slicing and `strings` offsets use bytes. Cut only at rune boundaries.
- `"..."` literals understand escapes; raw `` `...` `` literals don't.
- `strings` searches, splits, trims, replaces, and changes case; `bytes` mirrors it for `[]byte`.
- Build strings with `strings.Builder`; use `bytes.Buffer` for bytes you write and read back.
- `strconv` parses and formats numbers with clear errors. `string(65)` is `"A"`.
- `fmt` verbs take flags, width, and precision; `%v`, `%+v`, `%#v`, `%T` help debugging. `go vet`
  catches format mistakes.
- Changing case can change byte length. When offsets matter, use `strings.EqualFold` or compare rune
  by rune, not lowercased copies.

## Exercises

1. Write `reverse(s string) string` that reverses a string rune by rune, so
   `reverse("héllo, 世界")` is `"界世 ,olléh"`. Why would reversing the bytes be wrong?

<details><summary>Solution</summary>

```go
package main

import "fmt"

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func main() {
	fmt.Println(reverse("héllo, 世界"))
	fmt.Println(reverse(""))
}
```

```console
$ go run .
界世 ,olléh

```

Reversing bytes also reverses the bytes *inside* multi-byte runes, making invalid UTF-8. `[]rune`
decodes first, so each rune moves as a unit.

</details>

2. Write `title(s string) string` that upper-cases the first letter of each word and joins words
   with single spaces. It must handle non-ASCII first letters:
   `title("élan and   ünïcode go")` is `"Élan And Ünïcode Go"`.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// title upper-cases the first letter of each word.
func title(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r, size := utf8.DecodeRuneInString(w)
		words[i] = string(unicode.ToUpper(r)) + w[size:]
	}
	return strings.Join(words, " ")
}

func main() {
	fmt.Println(title("élan and   ünïcode go"))
}
```

```console
$ go run .
Élan And Ünïcode Go
```

`w[1:]` would cut `é` in half. `DecodeRuneInString` returns the first rune's size, so `w[size:]` is
a clean cut.

</details>

3. Parse this configuration text, held in a raw string literal. Skip blank lines and `#` lines.
   Split other lines at the first `=`, trim spaces, and parse the value with `strconv.Atoi`. Print
   `key = value` aligned with `fmt` width flags, or the line number and the error.

   ```text
   # limits
   workers = 8
   timeout=30

   retries = three
   ```

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

const config = `# limits
workers = 8
timeout=30

retries = three
`

func main() {
	for i, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			fmt.Printf("line %d: missing =\n", i+1)
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		n, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("line %d: %s: %v\n", i+1, key, err)
			continue
		}
		fmt.Printf("%-8s = %3d\n", key, n)
	}
}
```

```console
$ go run .
workers  =   8
timeout  =  30
line 5: retries: strconv.Atoi: parsing "three": invalid syntax
```

</details>

4. In the project, search `testdata/poem.txt` for the Kelvin sign with `-i` and `--color` (in bash,
   `$'\u212A'` types it). Predict the output first. Then explain why a `findAll` that lowercased the
   line and pattern would color the wrong bytes.

<details><summary>Solution</summary>

```console
$ go run . -n -i $'\u212A' testdata/poem.txt
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:4:Foxes are quick and clever.
$ go run . --color -i $'\u212A' testdata/poem.txt | cat -v
testdata/poem.txt:The quic^[[1;31mk^[[0m brown fox
testdata/poem.txt:Foxes are quic^[[1;31mk^[[0m and clever.
```

`EqualFold("K", "k")` is true, so the `k` in "quick" matches. Here the line is ASCII and the
pattern is the 3-byte Kelvin sign; lowercasing the *pattern* gives a 1-byte `k`, which happens to
work. But if the **line** had a Kelvin sign, lowercasing it would shrink the line by 2 bytes, and
every later offset would point 2 bytes too early in the original (as in the `casefold` example).
`indexFold` measures offsets in the original line, so they're always right.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-6--strings-bytes-and-runes)
