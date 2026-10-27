# Chapter 2 — Variables, Constants, and Basic Types

Learn Go's variables, constants, and basic types; why it never converts types silently; and zero
values, one of its best ideas. [Chapter 1](01-getting-started.md) printed
fixed text; now `seek` gets a version number and exit codes. [Chapter 3](03-control-flow.md) adds
decisions and loops.

## Declaring variables

Go is **statically typed**: each variable's type is fixed at compile time. You rarely write the
type, because the compiler infers it from the value.

`examples/ch02/variables/main.go`

```go
// Command variables shows the ways to declare variables.
package main

import "fmt"

func main() {
	var a int = 10 // name, type, and value
	var b = 20     // type inferred from the value: int
	var c int      // no value: c starts at 0
	d := 30        // short form, only inside functions

	var x, y = 1, "two" // several at once; types may differ
	i, j := 3, 4

	var (
		name  = "gopher"
		ready bool
	)

	fmt.Println(a, b, c, d)
	fmt.Println(x, y, i, j)
	fmt.Println(name, ready)

	// Assignment uses =. The right side is evaluated first,
	// so swapping needs no temporary variable.
	i, j = j, i
	a = a + 5
	b += 5
	d++
	fmt.Println(i, j, a, b, d)

	// := needs at least one new variable on the left.
	// Here k is new, and i is simply assigned.
	i, k := 100, 200
	fmt.Println(i, k)
}
```

```console
$ go run ./ch02/variables
10 20 0 30
1 two 3 4
gopher false
4 3 15 25 31
100 200
```

| Form | Where | Use it when |
|------|-------|-------------|
| `x := value` | inside functions | Almost always. |
| `var x = value` | anywhere | At package level, where `:=` isn't allowed. |
| `var x T` | anywhere | You want the zero value (below). |
| `var x T = value` | anywhere | You want a type other than the value's default, e.g. `var f float64 = 1`. |

- `:=` declares *and* assigns. `=` only assigns to an existing variable.
- `x++` and `x--` are statements, not expressions: no `y := x++`.
- `-=`, `*=`, `/=`, `%=`, and friends work as expected.

Common errors:

```go
a := 1
a := 2 // no new variable on the left
```

```console
./main.go:7:4: no new variables on left side of :=
```

`:=` is a statement, so it can't appear at package level (outside a function):

```console
./main.go:3:1: syntax error: non-declaration statement outside function body
```

## Unused variables are errors

```go
package main

func main() {
	x := 1
}
```

```console
./main.go:4:2: declared and not used: x
```

Like unused imports, unused local variables are compile errors: they often mean you computed
something and forgot it. Unused package-level variables, constants, and parameters are fine.

## Zero values

A variable declared without a value holds its type's **zero value**, never garbage.

`examples/ch02/zerovalues/main.go`

```go
// Command zerovalues prints the zero value of several types.
package main

import "fmt"

func main() {
	var (
		i int
		f float64
		b bool
		s string
		r rune
	)
	fmt.Printf("int=%v float64=%v bool=%v string=%q rune=%v\n", i, f, b, s, r)

	// Zero values are useful right away: no setup needed.
	var total int
	total += 5
	var greeting string
	greeting += "hello"
	fmt.Println(total, greeting)
}
```

```console
$ go run ./ch02/zerovalues
int=0 float64=0 bool=false string="" rune=0
5 hello
```

| Type | Zero value |
|------|-----------|
| numbers | `0` |
| `bool` | `false` |
| `string` | `""` (empty string) |
| pointers, slices, maps, channels, functions, interfaces | `nil` |
| arrays and structs | every element or field set to its zero value |

Why it matters:

- No uninitialized variables: a whole class of bugs is gone.
- Good Go types work at their zero value: an empty `strings.Builder`, a zero `sync.Mutex`, a `nil`
  slice you can append to. The saying: *make the zero value useful*.

## Booleans and operators

`examples/ch02/bools/main.go`

```go
// Command bools shows booleans, comparisons, and logical operators.
package main

import "fmt"

func main() {
	age := 20
	member := false

	adult := age >= 18 // comparisons produce a bool
	fmt.Println(adult, age == 20, age != 20)

	// && (and), || (or), ! (not). && and || stop early:
	// the right side runs only if it can change the result.
	fmt.Println(adult && member, adult || member, !member)

	// Bitwise operators work on integers.
	a, b := 0b1100, 0b1010
	fmt.Printf("%04b %04b %04b %04b\n", a&b, a|b, a^b, a&^b)
	fmt.Println(1<<4, 256>>2)
}
```

```console
$ go run ./ch02/bools
true true false
false true true
1000 1110 0110 0100
16 64
```

| Kind | Operators |
|------|-----------|
| Arithmetic | `+ - * / %` |
| Comparison | `== != < <= > >=` |
| Logical (bool only) | `&& \|\| !` |
| Bitwise (integers only) | `&` and, `\|` or, `^` xor, `&^` and-not (bit clear), `<<` `>>` shifts |

`bool` is its own type. Numbers aren't booleans, and booleans aren't numbers:

```console
./main.go:7:5: non-boolean condition in if statement
./main.go:10:16: cannot use 1 (untyped int constant) as bool value in variable declaration
```

Integer literals: `0b1100` (binary), `0o17` (octal), `0xff` (hex), `1_000_000` (underscores for
readability).

## Numbers

`examples/ch02/numbers/main.go`

```go
// Command numbers tours Go's numeric types.
package main

import (
	"fmt"
	"math"
)

func main() {
	// Sized integers: int8, int16, int32, int64 and unsigned uint8 ... uint64.
	fmt.Println("int8:  ", math.MinInt8, "to", math.MaxInt8)
	fmt.Println("uint8: ", 0, "to", math.MaxUint8)
	fmt.Println("int64: ", math.MinInt64, "to", math.MaxInt64)
	// int and uint are 64 bits on 64-bit systems, 32 bits on 32-bit ones.
	fmt.Println("int:   ", math.MinInt, "to", math.MaxInt)

	// Integer overflow wraps around silently at run time.
	var small int8 = 127
	small++
	fmt.Println("127 + 1 as int8 =", small)

	// Integer division truncates toward zero; % is the remainder.
	fmt.Println(7/2, -7/2, 7%3, -7%3)
	// Floating-point division. Floats are binary, so 0.1 is not exact.
	x, y := 0.1, 0.2
	fmt.Println(7.0/2, x+y)

	// byte is an alias for uint8, rune is an alias for int32.
	var b byte = 'A'
	var r rune = 'é'
	fmt.Println(b, r)
	fmt.Printf("%c %c\n", b, r)

	// float32 has about 7 significant digits, float64 about 16.
	var f32 float32 = 1.0 / 3
	var f64 float64 = 1.0 / 3
	fmt.Println(f32, f64)

	// Complex numbers are built in.
	c := complex(1, 2) // complex128
	fmt.Println(c*c, real(c), imag(c))
}
```

```console
$ go run ./ch02/numbers
int8:   -128 to 127
uint8:  0 to 255
int64:  -9223372036854775808 to 9223372036854775807
int:    -9223372036854775808 to 9223372036854775807
127 + 1 as int8 = -128
3 -3 1 -1
3.5 0.30000000000000004
65 233
A é
0.33333334 0.3333333333333333
(-3+4i) 1 2
```

The `int:` line shows 64-bit limits (a 64-bit machine).

| Type | Size | Use it for |
|------|------|-----------|
| `int` | 32 or 64 bits (platform) | Default integer: counts, indexes, lengths. |
| `int8` `int16` `int32` `int64` | as named | Exact sizes: file formats, protocols, saving memory. |
| `uint` `uint8` … `uint64` | as named | Bit manipulation, or values that can never be negative. |
| `uintptr` | pointer size | Low-level code only. |
| `byte` | alias for `uint8` | Raw data, the bytes of a string. |
| `rune` | alias for `int32` | One Unicode code point (a character). |
| `float64` | 64 bits | Default for decimals. |
| `float32` | 32 bits | Only when memory matters more than precision. |
| `complex128` `complex64` | 128 / 64 bits | Complex numbers (rare). |

- Use `int` unless you have a reason not to. `len` and most APIs return `int`.
- Integer overflow **wraps around** silently at run time. No crash.
- Integer division truncates: `7/2` is `3`. Use `7.0/2` (or convert) for `3.5`.
- Floats are binary approximations: `0.1 + 0.2` isn't exactly `0.3`. Never use floats for money;
  count cents in an `int64`.
- `'A'` (single quotes) is a **rune literal**; its value is the code point (`65`). Print it as a
  character with `%c`. Runes in depth: [Chapter 6](06-strings.md).

## Explicit conversions

Go **never** converts between types implicitly, not even from `int` to `float64`:

```go
count := 3
price := 1.25
fmt.Println(count * price)
```

```console
./main.go:8:14: invalid operation: count * price (mismatched types int and float64)
```

Write the conversion yourself, as `T(value)`:

`examples/ch02/conversions/main.go`

```go
// Command conversions shows explicit type conversions.
package main

import "fmt"

func main() {
	count := 3
	price := 1.25

	// Go never converts between numeric types for you.
	total := float64(count) * price
	fmt.Println(total)

	// float to int truncates toward zero.
	a, b := 2.99, -2.99
	fmt.Println(int(a), int(b))

	// Converting to a smaller type keeps only the low bits.
	big := 300
	fmt.Println(uint8(big)) // 300 - 256

	// Converting an integer to a string gives the character with that
	// code point, not the digits. Use a rune to make that clear.
	fmt.Println(string(rune(65)))
	// To get the digits, format the number.
	fmt.Println(fmt.Sprint(65) + "!")
}
```

```console
$ go run ./ch02/conversions
3.75
2 -2
44
A
65!
```

Why so strict? Implicit conversions hide bugs (lost precision, overflow, sign changes). Go makes
every conversion visible.

- Converting a **constant** that doesn't fit is a compile error, not a truncation:
  `int(2.99)` fails with `cannot convert 2.99 (untyped float constant) to type int`.
- `string(n)` on an `int` variable compiles, but `go vet` flags it as almost always a mistake:

  ```console
  main.go:7:14: conversion from int to string yields a string of one rune, not a string of digits
  ```

  For the digits, use `fmt.Sprint(n)` (like `Print`, but returns the string) or `strconv.Itoa(n)`
  ([Chapter 6](06-strings.md)).

## Strings

A string is an immutable sequence of bytes, usually UTF-8 text.

`examples/ch02/strbasics/main.go`

```go
// Command strbasics shows the basics of strings.
package main

import "fmt"

func main() {
	s := "hello, "
	s += "world" // concatenation builds a new string

	fmt.Println(s, len(s))
	fmt.Println(s[0], s[0:5]) // a byte, and a substring

	// Escape sequences work in double-quoted strings.
	fmt.Println("tab:\tquote:\" newline:\\n")

	// Raw strings use backquotes. No escapes; may span lines.
	raw := `C:\temp\new
second line`
	fmt.Println(raw)

	// len counts bytes, not characters.
	fmt.Println(len("café"))
}
```

```console
$ go run ./ch02/strbasics
hello, world 12
104 hello
tab:	quote:" newline:\n
C:\temp\new
second line
5
```

- `len(s)` counts **bytes**. `é` is 2 bytes in UTF-8, so `len("café")` is 5.
- `s[i]` is a `byte` (`104` is `h`). `s[i:j]` is the substring from byte `i` up to (not including)
  byte `j`.
- Strings can't be changed in place:

  ```console
  ./main.go:7:2: cannot assign to s[0] (neither addressable nor a map index expression)
  ```

  Immutable strings are safe to share and cheap to copy: copies share the same bytes.

UTF-8, runes, and the `strings` package: [Chapter 6](06-strings.md).

## Constants

A constant is fixed at compile time. Use `const` for anything that never changes.

`examples/ch02/constants/main.go`

```go
// Command constants shows typed and untyped constants.
package main

import "fmt"

const greeting = "hello" // untyped string constant

const (
	big   = 1 << 100  // untyped: far larger than any integer type
	small = big >> 98 // 4: fine, because it fits
)

const limit int64 = 10 // typed constant

func main() {
	fmt.Println(greeting, small)

	// An untyped constant takes the type its context needs.
	var f float64 = small
	var i int8 = small
	fmt.Println(f, i)
	fmt.Println(big / 1e29) // exact arithmetic until here

	// A typed constant behaves like a value of its type.
	var n int64 = limit * 2
	fmt.Println(n)

	const third = 1.0 / 3 // computed exactly at compile time
	fmt.Println(third * 3)
}
```

```console
$ go run ./ch02/constants
hello 4
4 4
12.676506002282293
20
1
```

**Untyped constants** like `greeting`, `big`, and `small` have a *kind* (integer, float,
string…) but no fixed type yet:

- They take the type the context needs (`small` became a `float64` and an `int8`), so `x * 2`
  works for any numeric `x`.
- Their arithmetic is **exact** (at least 256 bits). `1 << 100` is fine if the final value fits
  where it's used. `third * 3` is exactly `1`, not `0.999…`.
- If the value doesn't fit, the compiler says so:

  ```console
  ./main.go:8:14: cannot use big (untyped int constant 1267650600228229401496703205376) as int value in argument to fmt.Println (overflows)
  ./main.go:6:15: cannot use 200 (untyped int constant) as int8 value in variable declaration (overflows)
  ```

A **typed constant** like `limit int64` works only where an `int64` does. Use one when the type is
part of the meaning.

Constants can't be changed (`cannot assign to limit (neither addressable nor a map index
expression)`). Only numbers, strings, booleans, and runes can be constants.

## iota and enumerations

Go has no `enum` keyword. Instead, define a type and a block of constants using `iota`.

`examples/ch02/iota/main.go`

```go
// Command iota shows the iota enum pattern.
package main

import "fmt"

// Weekday is a named integer type for days of the week.
type Weekday int

// iota is 0 on the first line of a const block and goes up by one per line.
// Lines with no expression repeat the previous one.
const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
)

// Sizes skip the 0 line with _ and use iota in an expression.
const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func main() {
	fmt.Println(Sunday, Monday, Tuesday, Wednesday)
	fmt.Printf("%T\n", Tuesday)
	fmt.Println(KB, MB, GB)
}
```

```console
$ go run ./ch02/iota
0 1 2 3
main.Weekday
1024 1048576 1073741824
```

- `iota` counts lines in one `const (...)` block, from 0. Each block starts over.
- A line with only a name repeats the previous line's type and expression with the next `iota`:
  `Monday` means `Monday Weekday = iota` with `iota` = 1.
- `_ = iota` skips a value. Starting at 1 (`iota + 1`) leaves 0 free to mean "not set".
- Their own type (`Weekday`) lets the compiler catch mix-ups with plain numbers.
  [Chapter 10](10-interfaces.md) shows how a `String` method makes them print `Tuesday`, not `2`.

Gotcha: values depend on line order. Inserting a line renumbers everything after it. For values
that are stored or sent elsewhere and must never change, write them out instead.

## Type definitions and aliases

`type Name Underlying` creates a **new, distinct type**. `type Name = Other` creates an
**alias**: another name for the same type.

`examples/ch02/typedefs/main.go`

```go
// Command typedefs shows type definitions and type aliases.
package main

import "fmt"

// Celsius and Fahrenheit are new, distinct types.
type Celsius float64
type Fahrenheit float64

// Number is an alias: just another name for float64.
type Number = float64

func main() {
	boiling := Celsius(100)
	f := Fahrenheit(boiling*9/5 + 32) // explicit conversion required
	fmt.Println(boiling, f)
	fmt.Printf("%T %T\n", boiling, f)

	var n Number = 1.5
	var x float64 = n // fine: Number and float64 are the same type
	fmt.Printf("%T %v\n", n, x)
}
```

```console
$ go run ./ch02/typedefs
100 212
main.Celsius main.Fahrenheit
float64 1.5
```

A defined type has its underlying type's operations, but the compiler keeps the types apart:

```go
c := Celsius(20)
f := Fahrenheit(68)
fmt.Println(c + f)
```

```console
./main.go:11:14: invalid operation: c + f (mismatched types Celsius and Fahrenheit)
```

This catches real bugs (mixed-up units, IDs, durations) at no runtime cost. `%T` shows an alias
as its real type (`float64`). Aliases mainly help move types between packages when refactoring;
use definitions in your own code.

## Scope and shadowing

A name is visible from its declaration to the end of its **block** (`{ ... }`). Package-level
names are visible in every file of the package.

`examples/ch02/scope/main.go`

```go
// Command scope shows block scope and shadowing.
package main

import "fmt"

var level = "package" // visible in the whole package

func main() {
	fmt.Println(level)

	level := "function" // a new variable that shadows the package one
	fmt.Println(level)

	{
		level := "block" // shadows again, only inside these braces
		fmt.Println(level)
	}
	fmt.Println(level) // the block's variable is gone

	x := 1
	{
		x := 2 // := declares a NEW x here; the outer x is unchanged
		x++
	}
	fmt.Println("x is still", x)

	// The blank identifier _ discards a value you don't need.
	_, y := 10, 20
	fmt.Println(y)
}
```

```console
$ go run ./ch02/scope
package
function
block
function
x is still 1
20
```

**Shadowing** hides an outer name with a new one. It's legal but a common bug: you think you're
updating `x`, but `:=` made a new one. Watch `:=` inside `if`, `for`, and `{}` blocks; use `=` to
update.

The **blank identifier** `_` is a write-only placeholder: assigning to it discards the value. You'll use it with multi-result functions ([Chapter 4](04-functions.md)).

## Formatting verbs

The `fmt.Printf` verbs you'll use most:

`examples/ch02/verbs/main.go`

```go
// Command verbs shows the most common fmt verbs.
package main

import "fmt"

func main() {
	n := 255
	pi := 3.14159
	s := "go"

	fmt.Printf("%v %v %v\n", n, pi, s) // default format
	fmt.Printf("%d\n", n)              // decimal integer
	fmt.Printf("%x %X %b %o\n", n, n, n, n)
	fmt.Printf("%f %.2f %e\n", pi, pi, pi)
	fmt.Printf("%s %q\n", s, s) // %q adds quotes and escapes
	fmt.Printf("%q\n", "tab\there")
	fmt.Printf("%T %T %T %T\n", n, pi, s, 'x') // the type
	fmt.Printf("%x\n", "hi")                   // hex bytes of a string
	fmt.Printf("%t %c %U\n", true, 'G', 'G')
	fmt.Printf("%5d|%-5d|%05d\n", 42, 42, 42) // width and padding
	fmt.Printf("100%%\n")                     // a literal percent sign
}
```

```console
$ go run ./ch02/verbs
255 3.14159 go
255
ff FF 11111111 377
3.141590 3.14 3.141590e+00
go "go"
"tab\there"
int float64 string int32
6869
true G U+0047
   42|42   |00042
100%
```

| Verb | Prints |
|------|--------|
| `%v` | any value, default format |
| `%d` | decimal integer |
| `%x` `%X` `%b` `%o` | integer in hex, upper hex, binary, octal; `%x` on a string gives its bytes in hex |
| `%f` `%.2f` `%e` | float: 6 decimals, 2 decimals, scientific |
| `%s` | string |
| `%q` | quoted, escaped string (great for showing whitespace), or a quoted rune |
| `%T` | the value's type |
| `%t` | `true` / `false` |
| `%c` `%U` | a rune as a character, or as `U+XXXX` |
| `%%` | a literal `%` |

A number after `%` sets a width: `%5d` pads on the left, `%-5d` on the right, `%05d` with zeros.
The untyped rune constant `'x'` defaults to `int32` (that is, `rune`). `%+v`, `%#v`, and more:
[Chapter 6](06-strings.md).

## Building seek: version and exit codes

`seek` needs a version for `--version`, and exit codes so scripts know what happened. Like
`grep`: 0 = found something, 1 = found nothing, 2 = error. They never change, so they're
constants; the exit codes suit `iota`.

Replace `main.go` with:

```go
// Command seek searches files for lines that match a pattern.
package main

import "fmt"

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

func main() {
	fmt.Printf("%s %s\n", name, version)
	fmt.Println(usage)
	fmt.Printf("exit codes: ok=%d no-match=%d error=%d\n", exitOK, exitNoMatch, exitError)
}
```

- Lowercase constants are private to the package; nothing outside `seek` needs them.
- The exit codes are **untyped**, so in [Chapter 3](03-control-flow.md) they go straight into
  `os.Exit` (which takes an `int`) without a conversion.
- Scripts rely on exactly 0, 1, 2. `iota` is safe here: a comment warns about order, and trailing
  comments state the values.
- The last `Printf` is temporary scaffolding to show the values. Chapter 3 removes it.

### Checkpoint

```console
$ go run .
seek 0.1.0
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit codes: ok=0 no-match=1 error=2
```

The complete project at this point is in `project/ch02/`.

## Summary

- `x := v` in functions; `var x T` for a zero value; `var` at package level.
- Every type has a useful zero value (`0`, `false`, `""`, `nil`).
- Unused local variables and imports are compile errors.
- Default to `int` and `float64`. Integer overflow wraps; floats are approximate.
- `byte` is `uint8`; `rune` is `int32`, one Unicode code point.
- No implicit conversions: write `T(v)`. `int` to `string` gives a character, not digits.
- Strings are immutable bytes; `len` counts bytes.
- Untyped constants are exact and adapt to context. `iota` builds enums.
- `type T U` is a distinct type; `type T = U` is another name.
- Inner `:=` can shadow an outer variable. `_` discards values.

## Exercises

1. Declare an `int`, `float64`, `string`, and `bool` without values. Print each type and value on
   one line with one `Printf`, the string in quotes.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   func main() {
   	var (
   		count int
   		ratio float64
   		name  string
   		done  bool
   	)
   	fmt.Printf("%T=%v %T=%v %T=%q %T=%v\n", count, count, ratio, ratio, name, name, done, done)
   }
   ```

   ```console
   $ go run .
   int=0 float64=0 string="" bool=false
   ```

   </details>

2. With `type Celsius float64` and `type Fahrenheit float64`, convert 98.6°F to Celsius
   (`C = (F - 32) × 5 / 9`) and print both with one decimal.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   type Celsius float64
   type Fahrenheit float64

   func main() {
   	body := Fahrenheit(98.6)
   	c := Celsius((body - 32) * 5 / 9)
   	fmt.Printf("%.1f°F = %.1f°C\n", body, c)
   }
   ```

   ```console
   $ go run .
   98.6°F = 37.0°C
   ```

   The math happens in `Fahrenheit`; converting to `Celsius` comes last.

   </details>

3. Define type `Level` and constants `Debug`, `Info`, `Warn`, `Error` = 1, 2, 3, 4 using `iota`.
   Print them.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   type Level int

   const (
   	Debug Level = iota + 1
   	Info
   	Warn
   	Error
   )

   func main() {
   	fmt.Println(Debug, Info, Warn, Error)
   }
   ```

   ```console
   $ go run .
   1 2 3 4
   ```

   </details>

4. Predict the output, then run it:

   ```go
   package main

   import "fmt"

   func main() {
   	n := 1
   	{
   		n := n + 10
   		n++
   		fmt.Println("inner:", n)
   	}
   	n += 5
   	fmt.Println("outer:", n)
   }
   ```

   <details><summary>Solution</summary>

   ```console
   $ go run .
   inner: 12
   outer: 6
   ```

   `n := n + 10` declares a new inner `n`. On the right, `n` is still the outer one (the new
   variable's scope starts *after* its declaration), so inner `n` is 11, then 12. Outer `n` goes
   from 1 to 6.

   </details>

5. Why does the first program fail to compile while the second prints `0`?

   ```go
   fmt.Println(uint8(255) + 1)
   ```

   ```go
   var b uint8 = 255
   fmt.Println(b + 1)
   ```

   <details><summary>Solution</summary>

   `uint8(255) + 1` is a **constant** expression. The compiler computes it exactly (256), sees it
   doesn't fit, and fails:

   ```console
   ./main.go:6:14: uint8(255) + 1 (constant 256 of type uint8) overflows uint8
   ```

   `b + 1` runs at run time, where overflow wraps: 255 + 1 becomes 0.

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-2--variables-constants-and-basic-types)
