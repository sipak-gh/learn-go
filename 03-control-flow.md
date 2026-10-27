# Chapter 3 — Control Flow

Go makes decisions and repeats work with just three statements: `if`, `for`, and `switch`. No
`while`, no `do`, no ternary `?:`. You'll also meet labels for nested loops and the rare `goto`.
`seek` got its constants in [Chapter 2](02-values-and-types.md); now it reads its command line.
[Chapter 4](04-functions.md) splits it into functions and runs the first real search.

## if and else

`examples/ch03/ifelse/main.go`

```go
// Command ifelse shows if, else if, else, and if with an init statement.
package main

import "fmt"

func main() {
	temp := 23

	// No parentheses around the condition; braces are required.
	if temp > 30 {
		fmt.Println("hot")
	} else if temp > 15 {
		fmt.Println("mild")
	} else {
		fmt.Println("cold")
	}

	// An init statement runs before the condition. Variables it
	// declares exist only inside the if and its else branches.
	word := "gopher"
	if n := len(word); n > 5 {
		fmt.Println(word, "is long:", n)
	} else {
		fmt.Println(word, "is short:", n)
	}
	// n is not defined here.
}
```

```console
$ go run ./ch03/ifelse
mild
gopher is long: 6
```

- The condition must be a `bool`. `if n {` with an `int` fails
  (`non-boolean condition in if statement`).
- Braces are always required, even for one statement. That kills a classic bug: adding a second
  line to a brace-less `if` that then always runs.
- `else` goes on the same line as the closing `}`. Otherwise a semicolon is inserted after `}` and
  the `if` ends there:

  ```console
  ./main.go:10:2: syntax error: unexpected keyword else, expected }
  ```

- The **init statement** (`n := len(word);`) keeps a variable's scope small: using `n` after the
  `if` fails with `undefined: n`. You'll use it constantly for error checks:
  `if err := doSomething(); err != nil { ... }` ([Chapter 9](09-errors.md)).

No ternary (`a > b ? a : b`): write an `if`. Go's authors found `?:` led to dense, hard-to-read
code. For a maximum, use the built-in `max(a, b)` ([Chapter 5](05-collections.md)).

### Idiom: return early

Keep the "happy path" on the left. Handle special cases first and leave early, rather than nesting
the main logic in `else`:

```go
if len(args) == 0 {
	fmt.Println("nothing to do")
	return
}
// main logic here, not indented
```

`return` leaves the current function. You'll see this shape everywhere.

## for: Go's only loop

`examples/ch03/forloops/main.go`

```go
// Command forloops shows the forms of Go's only loop, for.
package main

import "fmt"

func main() {
	// Three clauses: init; condition; post.
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// Condition only: Go's "while" loop.
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println("n =", n)

	// No condition: loops forever until break (or return).
	count := 0
	for {
		count++
		if count == 4 {
			break // leave the loop
		}
		if count%2 == 0 {
			continue // skip to the next iteration
		}
		fmt.Println("odd count", count)
	}

	// Range over an integer (Go 1.22+): 0, 1, ..., n-1.
	for i := range 3 {
		fmt.Print(i*10, " ")
	}
	fmt.Println()
}
```

```console
$ go run ./ch03/forloops
0 1 2 
n = 243
odd count 1
odd count 3
0 10 20 
```

| Form | Equivalent in other languages |
|------|------------------------------|
| `for init; cond; post { }` | C-style `for` |
| `for cond { }` | `while (cond)` |
| `for { }` | `while (true)` |
| `for i := range n { }` | `for i in range(n)` (Python) |
| `for i, v := range x { }` | `for-each` (next section) |

- `break` leaves the innermost loop. `continue` jumps to the next iteration (after running the
  post statement, in a 3-clause loop).
- Any clause can be empty: `for ; i < n; i++ { }` reuses an existing `i`.
- A variable declared in the init clause exists only inside the loop. To use it afterward, declare
  it before the loop, or you get `undefined: i`. `seek` relies on this below.
- Since Go 1.22, each iteration gets a **fresh copy** of the loop variable. This matters for
  closures ([Chapter 4](04-functions.md)).

One loop keyword with a few forms is simpler to learn and read than four loop statements.

## for range

`range` loops over a collection, giving you an index (or key) and a value.

`examples/ch03/ranges/main.go`

```go
// Command ranges shows for-range over slices, strings, and maps.
package main

import "fmt"

func main() {
	// A slice is a list of values (Chapter 5).
	colors := []string{"red", "green", "blue"}
	for i, c := range colors {
		fmt.Println(i, c)
	}
	for _, c := range colors { // skip the index with _
		fmt.Print(c, " ")
	}
	fmt.Println()
	for i := range colors { // or leave out the value
		fmt.Print(i, " ")
	}
	fmt.Println()

	// Ranging over a string yields byte offsets and runes (characters).
	for i, r := range "héllo" {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	// A map holds key/value pairs (Chapter 5). Its order is random,
	// so this sums the values instead of printing them in order.
	stock := map[string]int{"apples": 5, "pears": 3, "plums": 7}
	total := 0
	for _, n := range stock {
		total += n
	}
	fmt.Println("total:", total)
}
```

```console
$ go run ./ch03/ranges
0 red
1 green
2 blue
red green blue 
0 1 2 
0:h 1:é 3:l 4:l 5:o 
total: 15
```

| Range over | First value | Second value |
|------------|-------------|--------------|
| slice or array | index | element (a copy) |
| string | byte offset | rune (decoded UTF-8 character) |
| map | key | value; order is random |
| integer `n` | 0 … n-1 | — |
| channel | element received | — ([Chapter 16](16-goroutines-and-channels.md)) |
| function | whatever it yields ([Chapter 14](14-iterators.md)) | |

- Skip the index with `_`; skip the value by leaving out the second variable.
- Over a string, offsets jump (`1` → `3`) because `é` is 2 bytes ([Chapter 6](06-strings.md)).
- Map order is deliberately random so code can't depend on it. [Chapter 5](05-collections.md)
  shows how to visit keys in sorted order.
- The value is a **copy**; assigning to it doesn't change the collection. Use the index:
  `colors[i] = "x"`.

## Labels: break and continue in nested loops

`break` leaves the innermost `for`, `switch`, or `select`; `continue` affects the innermost `for`.
To target an outer loop, put a **label** before it and name it.

`examples/ch03/labels/main.go`

```go
// Command labels shows labeled break and continue in nested loops.
package main

import "fmt"

func main() {
	grid := [][]int{ // a slice of slices (Chapter 5)
		{1, 2, 3},
		{4, -5, 6},
		{7, 8, 9},
	}

	// Find the first negative number. A plain break would only
	// leave the inner loop.
search:
	for r, row := range grid {
		for c, v := range row {
			if v < 0 {
				fmt.Println("negative at", r, c)
				break search
			}
		}
	}

	// Sum only the rows that have no negative numbers.
rows:
	for r, row := range grid {
		sum := 0
		for _, v := range row {
			if v < 0 {
				continue rows // skip the rest of this row
			}
			sum += v
		}
		fmt.Println("row", r, "sum", sum)
	}
}
```

```console
$ go run ./ch03/labels
negative at 1 1
row 0 sum 6
row 2 sum 24
```

- A label is a name plus `:`. `gofmt` outdents it one level.
- `break label` leaves the labeled statement; `continue label` starts its next iteration.
- An unused label is an error: `label outer defined and not used`.

## switch

`switch` is a cleaner `if`/`else if` chain.

`examples/ch03/switches/main.go`

```go
// Command switches shows the forms of switch.
package main

import "fmt"

func main() {
	// Expression switch. Cases don't fall through: no break needed.
	day := "sat"
	switch day {
	case "sat", "sun": // several values in one case
		fmt.Println("weekend")
	case "fri":
		fmt.Println("almost weekend")
	default:
		fmt.Println("weekday")
	}

	// No expression: like a chain of if/else if. The first true case wins.
	score := 82
	switch {
	case score >= 90:
		fmt.Println("A")
	case score >= 80:
		fmt.Println("B")
	default:
		fmt.Println("C or lower")
	}

	// An init statement, like in if.
	switch n := len(day); {
	case n > 2:
		fmt.Println("long name")
	default:
		fmt.Println("short name")
	}

	// fallthrough continues into the next case's body without testing it.
	switch level := 2; level {
	case 2:
		fmt.Println("level 2")
		fallthrough
	case 1:
		fmt.Println("level 1")
	case 0:
		fmt.Println("level 0")
	}
}
```

```console
$ go run ./ch03/switches
weekend
B
long name
level 2
level 1
```

How it differs from C and Java:

- **No automatic fall-through.** Each case ends by itself, so there's no `break` to forget (a
  famous source of C bugs). Safe is the default.
- **`fallthrough` is explicit** and rare. It runs the next case's body without checking its value.
- **A case can list several values**: `case "sat", "sun":`.
- **Cases can be any expression**, not just constants. `switch {` (no expression) tests each case
  as a boolean, top to bottom: the idiomatic replacement for long `if`/`else if` chains.
- `default` runs when nothing matches. It can go anywhere, but usually goes last.
- Duplicate constant cases are a compile error:

  ```console
  ./main.go:9:7: duplicate case 1 (constant of type int) in expression switch
  	./main.go:7:7: previous case
  ```

### Gotcha: break inside switch

Inside a `switch`, `break` leaves the **switch**, not the loop around it. Use a label to leave the
loop:

`examples/ch03/switchbreak/main.go`

```go
// Command switchbreak shows that break inside a switch leaves only the switch.
package main

import "fmt"

func main() {
	words := []string{"go", "stop", "after"}

	for _, w := range words {
		switch w {
		case "stop":
			break // leaves the switch, NOT the loop
		default:
			fmt.Println("plain break:", w)
		}
	}

loop:
	for _, w := range words {
		switch w {
		case "stop":
			break loop // leaves the loop
		default:
			fmt.Println("labeled break:", w)
		}
	}
}
```

```console
$ go run ./ch03/switchbreak
plain break: go
plain break: after
labeled break: go
```

`seek` uses this pattern below.

## goto

`goto` jumps to a label in the same function.

`examples/ch03/gotos/main.go`

```go
// Command gotos shows goto. Real code almost never needs it.
package main

import "fmt"

func main() {
	i := 0
again:
	if i < 3 {
		fmt.Println("i =", i)
		i++
		goto again
	}
}
```

```console
$ go run ./ch03/gotos
i = 0
i = 1
i = 2
```

You can't jump into a block or over a variable declaration:

```console
./main.go:6:7: goto done jumps over declaration of msg at ./main.go:7:6
```

You'll rarely see `goto` outside generated code and low-level corners of the standard library. A
loop or an early `return` is almost always clearer.

## Building seek: reading the command line

`seek` now reads its arguments:

```text
seek [OPTIONS] PATTERN [PATH ...]
```

Options come first. The first non-option argument is the PATTERN; the rest are PATHs. For now:

| Argument | Behavior |
|----------|----------|
| `-h`, `--help` | Print the usage line, exit 0. |
| `--version` | Print `seek 0.1.0`, exit 0. |
| `--` | End of options: the next argument is the PATTERN, even if it starts with `-`. |
| any other `-x` | Print `seek: unknown option "-x"` to stderr, exit 2. |
| (no PATTERN) | Print an error and the usage line to stderr, exit 2. |

Searching comes in [Chapter 4](04-functions.md); for now `seek` prints what it understood.

Three new tools from package `os`:

- `os.Args` is the command line: `os.Args[0]` is the program name, then the arguments. It's a
  **slice** of strings ([Chapter 5](05-collections.md)). For now: `len(args)` is the element
  count, `args[i]` is element `i` (from 0), and `args[i:]` is everything from element `i` on.
- `os.Exit(code)` ends the program immediately with an exit status.
- `os.Stderr` is the standard error stream. `fmt.Fprintln(os.Stderr, ...)` is `Println` aimed
  there (`fmt.Fprintf` matches `Printf`). Errors go on stderr so they don't mix with real output
  when stdout is piped to another program.

Replace `main.go` with:

```go
// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
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

func main() {
	// os.Args[0] is the program name; the arguments follow it.
	args := os.Args[1:]

	// Options come first. The first argument that is not an option is
	// the PATTERN, and everything after it is a PATH.
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			os.Exit(exitOK)
		case "--version":
			fmt.Println(name, version)
			os.Exit(exitOK)
		case "--":
			// "--" ends the options, so a PATTERN may start with "-".
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "seek: unknown option %q\n", arg)
				os.Exit(exitError)
			}
			break loop // a plain break would only leave the switch
		}
	}

	if i >= len(args) {
		fmt.Fprintln(os.Stderr, "seek: missing PATTERN")
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(exitError)
	}
	pattern := args[i]
	paths := args[i+1:]

	fmt.Printf("pattern: %q\n", pattern)
	for n, path := range paths {
		fmt.Printf("path %d: %s\n", n+1, path)
	}
	if len(paths) == 0 {
		fmt.Println("no paths given")
	}
}
```

The temporary `exit codes:` line from Chapter 2 is gone. The new `main`:

- **`i := 0` before the loop**: the code after the loop needs to know where options stopped, and a
  variable declared in the `for` clause would be gone. So the init clause is empty:
  `for ; i < len(args); i++`.
- **`switch arg := args[i]; arg`** names the current argument in an init statement, then switches
  on it.
- **`case "-h", "--help":`** handles two spellings at once.
- **`case "--":`** skips the `--` (`i++`) and stops parsing options, so `seek -- -fox` searches for
  `-fox`.
- **`default:`** handles the rest. `arg[0] == '-'` compares the first byte to the rune literal
  `'-'`. Checking `len(arg) > 1` first keeps `arg[0]` in range for an empty argument, and means a
  lone `-` isn't an option (it will mean "standard input" in [Chapter 11](11-io-and-files.md)).
  Anything else is the PATTERN, so the loop ends.
- **`break loop`** is the labeled break from the gotcha above.
- **`if i >= len(args)`**: the arguments ran out before a PATTERN.
- **`os.Exit(exitError)`** passes an untyped constant to an `int` parameter with no conversion, as
  Chapter 2 promised.
- **`%q`** quotes the pattern, so an empty one or one with spaces is easy to see.

### Checkpoint

```console
$ go run . fox poem.txt notes.md
pattern: "fox"
path 1: poem.txt
path 2: notes.md
$ go run . --version
seek 0.1.0
$ go run . -h
Usage: seek [OPTIONS] PATTERN [PATH ...]
$ go run . -- -fox
pattern: "-fox"
no paths given
$ go run .
seek: missing PATTERN
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run . -x fox
seek: unknown option "-x"
exit status 2
```

`poem.txt` and `notes.md` don't exist yet; `seek` doesn't open them in this chapter.

The `exit status 2` lines come from `go run`, not `seek`. When a program exits non-zero, `go run`
reports it and then exits with status 1 itself. To see the real status, build the binary and run
it directly. `echo $?` prints the last command's exit status:

```console
$ go build
$ ./seek; echo $?
seek: missing PATTERN
Usage: seek [OPTIONS] PATTERN [PATH ...]
2
$ ./seek -x fox; echo $?
seek: unknown option "-x"
2
$ ./seek --version; echo $?
seek 0.1.0
0
```

On Windows, run `seek.exe` and check `$LASTEXITCODE` (PowerShell) or `%ERRORLEVEL%` (cmd). Delete
the binary afterward.

The complete project at this point is in `project/ch03/`.

## Summary

- `if`: no parentheses, always braces. An init statement (`if x := f(); x > 0`) limits scope.
- `for` is the only loop: 3-clause, condition-only, infinite, and `range`.
- `range` covers slices, arrays, strings (runes), maps (random order), integers, channels, and
  functions.
- `break`/`continue` affect the innermost loop. Labels reach outer loops, including from inside a
  `switch`.
- `switch` cases don't fall through; they can list several values or be any condition.
  `fallthrough` is explicit.
- `goto` exists but is rarely right.
- Write errors to `os.Stderr`; end with `os.Exit(code)`. `go run` reports a non-zero exit as
  `exit status N`.

## Exercises

1. Print 1 to 15, but `Fizz` for multiples of 3, `Buzz` for multiples of 5, and `FizzBuzz` for
   both. Use a `switch` with no expression.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   func main() {
   	for i := 1; i <= 15; i++ {
   		switch {
   		case i%15 == 0:
   			fmt.Println("FizzBuzz")
   		case i%3 == 0:
   			fmt.Println("Fizz")
   		case i%5 == 0:
   			fmt.Println("Buzz")
   		default:
   			fmt.Println(i)
   		}
   	}
   }
   ```

   `i%15` must come first: the first true case wins.

   </details>

2. Count vowels and other letters in `"Gophers love Go!"`, ignoring spaces and `!`. Use `range`
   over the string and a `switch` with multi-value cases.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   func main() {
   	vowels, others := 0, 0
   	for _, r := range "Gophers love Go!" {
   		switch r {
   		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
   			vowels++
   		case ' ', '!':
   			// not a letter: don't count it
   		default:
   			others++
   		}
   	}
   	fmt.Println("vowels:", vowels, "others:", others)
   }
   ```

   ```console
   $ go run .
   vowels: 5 others: 8
   ```

   An empty case is fine; it does nothing.

   </details>

3. Print all primes below 30. Use nested loops and a labeled `continue` to skip a number at its
   first divisor.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   func main() {
   next:
   	for n := 2; n < 30; n++ {
   		for d := 2; d*d <= n; d++ {
   			if n%d == 0 {
   				continue next
   			}
   		}
   		fmt.Print(n, " ")
   	}
   	fmt.Println()
   }
   ```

   ```console
   $ go run .
   2 3 5 7 11 13 17 19 23 29 
   ```

   </details>

4. Make `seek -V` work as a short form of `--version`.

   <details><summary>Solution</summary>

   Change one line in `main`:

   ```go
   		case "-V", "--version":
   ```

   ```console
   $ go run . -V
   seek 0.1.0
   ```

   (The finished `seek` has no `-V`; this is practice. Undo it before Chapter 4.)

   </details>

5. Why does `seek` declare `i := 0` before the loop instead of `for i := 0; ...`? What error
   would you get?

   <details><summary>Solution</summary>

   The code after the loop uses `i` to find the PATTERN (`args[i]`) and PATHs (`args[i+1:]`). A
   variable from the `for` init clause exists only inside the loop, so you'd get `undefined: i` at
   its first use after the loop.

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-3--control-flow)
