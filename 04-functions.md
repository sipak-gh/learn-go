# Chapter 4 — Functions, Closures, and Defer

Functions organize Go code. Learn to declare them, return several results, pass them as values,
capture variables in closures, and schedule cleanup with `defer`. In
[Chapter 3](03-control-flow.md) all of `seek` lived in `main`; now you'll split it into functions
and search a real file. [Chapter 5](05-collections.md) searches many files.

## Declaring functions

`examples/ch04/functions/main.go`

```go
// Command functions shows function declarations and multiple results.
package main

import "fmt"

// add takes two ints and returns their sum.
// Consecutive parameters of the same type can share it: a, b int.
func add(a, b int) int {
	return a + b
}

// divmod returns two results: the quotient and the remainder.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// greet has no results.
func greet(name string) {
	fmt.Println("hello,", name)
}

func main() {
	fmt.Println(add(2, 3))

	q, r := divmod(17, 5)
	fmt.Println(q, r)

	_, r = divmod(9, 4) // ignore a result with _
	fmt.Println(r)

	greet("gopher")
}
```

```console
$ go run ./ch04/functions
5
3 2
1
hello, gopher
```

The shape is `func name(parameters) results { body }`:

- Types come **after** names: `a int`, not `int a`. It reads left to right ("a is an int"), which
  keeps complex types readable.
- `a, b int` shares one type between consecutive parameters.
- One result is bare (`int`); several go in parentheses (`(int, int)`).
- Functions live at package level. You can't nest a named `func` in another (function literals,
  below, are fine).
- Order doesn't matter: `main` can call a function declared below it.

**Multiple results** replace output parameters and exceptions. The most common use is a value plus
an error: `data, err := os.ReadFile(path)`. You'll use it in the project below; errors in depth come
in [Chapter 9](09-errors.md).

Receive all results, or none:

```console
./main.go:10:7: assignment mismatch: 1 variable but divmod returns 2 values
```

Other errors you'll meet:

```console
./main.go:10:18: not enough arguments in call to add
	have (number)
	want (int, int)
./main.go:6:12: too many return values
	have (int, int)
	want (int)
./main.go:11:1: missing return
```

`missing return` means some path reaches the closing `}` without a `return`. The compiler doesn't
prove that an `if`/`else if` chain covers every case, so end with a final `return` (or a plain
`else`).

Go deliberately has no default parameter values, no overloading (two functions with one name), and
no named arguments. Each call means exactly one thing, readable without looking anything up.

Arguments are **passed by value**: the function gets a copy.
[Chapter 7](07-structs-methods-pointers.md) shows how pointers let it change the caller's
variables.

## Named results

`examples/ch04/namedresults/main.go`

```go
// Command namedresults shows named results and a naked return.
package main

import "fmt"

// split names its results. They start at their zero values and act
// like local variables. A bare "return" returns their current values.
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

// minMax names its results for documentation, but returns explicitly.
func minMax(a, b int) (lo, hi int) {
	if a < b {
		return a, b
	}
	return b, a
}

func main() {
	fmt.Println(split(17))
	fmt.Println(minMax(8, 3))
}
```

```console
$ go run ./ch04/namedresults
7 10
3 8
```

`fmt.Println(split(17))` passes both results straight into `Println`. That works when a call's
results match the parameters of the function receiving them.

Named results are good **documentation**: `(lo, hi int)` says more than `(int, int)`.

A **naked return** (bare `return` with named results) is legal, but avoid it beyond a few lines:
readers must hunt for the last assignment to each result. Prefer `return lo, hi`, as `minMax`
does.

## Variadic functions

A final parameter `...T` accepts any number of `T` values.

`examples/ch04/variadic/main.go`

```go
// Command variadic shows functions that take any number of arguments.
package main

import "fmt"

// sum accepts zero or more ints. Inside, nums is a []int.
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// join has a regular parameter before the variadic one, which must be last.
func join(sep string, parts ...string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func main() {
	fmt.Println(sum())
	fmt.Println(sum(1, 2, 3))

	nums := []int{4, 5, 6}
	fmt.Println(sum(nums...)) // spread a slice into the arguments

	fmt.Println(join("-", "a", "b", "c"))
}
```

```console
$ go run ./ch04/variadic
0
6
15
a-b-c
```

- Inside, `nums` is a slice (`[]int`, a list type; [Chapter 5](05-collections.md)). With no
  arguments, it's empty.
- `nums...` passes an existing slice as the variadic arguments.
- `fmt.Println` is variadic: `func Println(a ...any) (n int, err error)`. `any` means "a value of
  any type" ([Chapter 10](10-interfaces.md)). Now you can read the signature from
  [Chapter 1](01-getting-started.md).

## Functions as values

Functions are values: store them in variables, pass them, return them.

`examples/ch04/funcvalues/main.go`

```go
// Command funcvalues shows functions used as values.
package main

import "fmt"

// op is a function type: any func taking two ints and returning an int.
type op func(int, int) int

func add(a, b int) int { return a + b }
func mul(a, b int) int { return a * b }

// apply calls f on a and b. Functions can be parameters.
func apply(f op, a, b int) int {
	return f(a, b)
}

// pick returns a function. Functions can be results.
func pick(name string) op {
	if name == "add" {
		return add
	}
	return mul
}

func main() {
	f := add // a variable holding a function
	fmt.Println(f(2, 3))
	fmt.Printf("%T\n", f)

	fmt.Println(apply(mul, 4, 5))
	fmt.Println(pick("add")(10, 1))

	// An anonymous function (function literal), called right away.
	square := func(n int) int { return n * n }
	fmt.Println(apply(func(a, b int) int { return a - b }, 9, 4), square(7))

	var g op // the zero value of a function type is nil
	fmt.Println(g == nil)
}
```

```console
$ go run ./ch04/funcvalues
5
func(int, int) int
20
11
5 49
true
```

- A function's type is its signature: `func(int, int) int`. Parameter names don't matter.
- `type op func(int, int) int` names a function type, making signatures shorter and clearer.
- A **function literal** (`func(n int) int { ... }`) has no name. Assign it, pass it, or call it
  right away.
- A function type's zero value is `nil`. Calling a `nil` function crashes with
  `panic: runtime error: invalid memory address or nil pointer dereference`.
- Functions compare only to `nil`, not to each other.

The standard library often takes behavior as a function: `strings.FieldsFunc(s, f)` splits `s`
wherever `f` returns true; `slices.SortFunc` sorts with your comparison function.

## Closures

A function literal can use (**capture**) variables from the function around it. That makes it a
**closure**.

`examples/ch04/closures/main.go`

```go
// Command closures shows function literals that capture variables.
package main

import "fmt"

// counter returns a function that counts its calls. Each call to counter
// creates a new n, which lives on as long as the returned function does.
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	next := counter()
	fmt.Println(next(), next(), next())

	other := counter() // a separate n
	fmt.Println(other())

	// Closures capture variables, not values: they see later changes.
	greeting := "hello"
	say := func() { fmt.Println(greeting) }
	greeting = "goodbye"
	say()

	// ...and they can change the variable.
	total := 0
	addTo := func(n int) { total += n }
	addTo(5)
	addTo(10)
	fmt.Println(total)
}
```

```console
$ go run ./ch04/closures
1 2 3
1
goodbye
15
```

- A closure captures the **variable itself** (by reference), not a copy. It sees later changes
  and can make its own.
- A captured variable lives as long as any closure using it, even after its declaring function
  returns: `n` in `counter` keeps counting. The compiler and garbage collector handle this.
- Closures give a function state without a new type. `seek` uses one below.

### Loop variables are per-iteration

Since Go 1.22, each `for` iteration has its **own** loop variable, so a closure made in the loop
captures that iteration's variable.

`examples/ch04/loopvars/main.go`

```go
// Command loopvars shows that each loop iteration has its own variable.
package main

import "fmt"

func main() {
	var funcs []func() // a slice of functions (Chapter 5)
	for i := range 3 {
		// Each iteration has a fresh i, so each closure captures a
		// different variable (since Go 1.22).
		funcs = append(funcs, func() { fmt.Print(i, " ") })
	}
	for _, f := range funcs {
		f()
	}
	fmt.Println()

	// To share one variable across iterations, declare it outside.
	shared := 0
	var later []func()
	for range 3 {
		shared++
		later = append(later, func() { fmt.Print(shared, " ") })
	}
	for _, f := range later {
		f()
	}
	fmt.Println()
}
```

```console
$ go run ./ch04/loopvars
0 1 2 
3 3 3 
```

`append` adds an element to a slice ([Chapter 5](05-collections.md)). Before Go 1.22, all
iterations shared one `i`, so the first loop printed `3 3 3`: one of Go's most common bugs. The fix
applies to modules whose `go.mod` says `go 1.22` or later. Old code's `i := i` lines made the
per-iteration copy by hand; they're no longer needed.

## Recursion

`examples/ch04/recursion/main.go`

```go
// Command recursion shows functions that call themselves.
package main

import "fmt"

func factorial(n int) int {
	if n <= 1 {
		return 1 // the base case stops the recursion
	}
	return n * factorial(n-1)
}

func main() {
	fmt.Println(factorial(5), factorial(20))

	// A function literal can't refer to itself by a name declared with :=,
	// so declare the variable first, then assign.
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}
	fmt.Println(fib(10))
}
```

```console
$ go run ./ch04/recursion
120 2432902008176640000
55
```

With `fib := func(...) {... fib(n-1) ...}`, `fib` doesn't exist yet inside the literal:

```console
./main.go:10:10: undefined: fib
```

Go's stacks start small and grow, so deep recursion is fine. Infinite recursion still crashes at
the stack limit (1 GB on 64-bit systems). Go doesn't optimize tail calls, so prefer a loop when one
fits naturally.

## defer

`defer` schedules a call to run when the surrounding function returns.

`examples/ch04/deferbasics/main.go`

```go
// Command deferbasics shows when deferred calls run.
package main

import "fmt"

func main() {
	fmt.Println("start")

	// Deferred calls run when main returns, last-in first-out.
	defer fmt.Println("deferred 1")
	defer fmt.Println("deferred 2")

	// Arguments are evaluated NOW, when the defer statement runs.
	x := 10
	defer fmt.Println("deferred x =", x)

	// A deferred closure reads x when it finally runs.
	defer func() { fmt.Println("closure sees x =", x) }()

	x = 20
	fmt.Println("end of main, x =", x)
}
```

```console
$ go run ./ch04/deferbasics
start
end of main, x = 20
closure sees x = 20
deferred x = 10
deferred 2
deferred 1
```

Three rules:

1. Deferred calls run when the function returns: by `return`, by reaching the end, or by a panic
   ([Chapter 9](09-errors.md)).
2. They run **LIFO** (last deferred, first run), like a stack of plates.
3. The function value and **arguments are evaluated immediately**, at the `defer` statement; only
   the call waits. Hence `deferred x = 10`. A deferred closure reads variables when it runs, so it
   sees `20`.

Note the `()` in `defer func() { ... }()`: you defer a *call*, so you must call the literal.

### Why defer exists, and how it's used

`defer` puts cleanup next to setup, so no return path forgets it:

```go
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close() // runs however this function returns
```

You'll write this in [Chapter 11](11-io-and-files.md). Other uses: unlocking a mutex
(`mu.Lock(); defer mu.Unlock()`, [Chapter 17](17-sync.md)), flushing buffered output, timing a
function.

`examples/ch04/deferuses/main.go`

```go
// Command deferuses shows common ways to use defer.
package main

import "fmt"

// double sets its named result in a deferred call, after the return
// statement has set result but before the caller receives it.
func double(n int) (result int) {
	defer func() { result *= 2 }()
	return n
}

// process shows the usual shape: acquire, defer the release, use.
func process(name string, size int) {
	fmt.Println("open", name)
	defer fmt.Println("close", name) // runs however process returns
	if size == 0 {
		fmt.Println(name, "is empty")
		return
	}
	fmt.Println("read", size, "bytes from", name)
}

func main() {
	fmt.Println(double(21))
	process("a.txt", 120)
	process("b.txt", 0)

	// Defers in a loop pile up until the FUNCTION returns,
	// not until the iteration ends.
	for i := range 3 {
		defer fmt.Println("loop defer", i)
	}
	fmt.Println("loop done")
}
```

```console
$ go run ./ch04/deferuses
42
open a.txt
read 120 bytes from a.txt
close a.txt
open b.txt
b.txt is empty
close b.txt
loop done
loop defer 2
loop defer 1
loop defer 0
```

Gotchas:

- **`defer` is per function, not per block.** In a loop opening many files, `defer f.Close()` keeps
  every file open until the function returns. Move the loop body into its own function so each
  `defer` runs per call.
- **A deferred closure can change named results** (see `double`). `return n` sets `result`, the
  deferred call runs, then the caller gets the value. Mainly used to adjust errors
  ([Chapter 9](09-errors.md)).

### os.Exit skips deferred calls

`examples/ch04/deferexit/main.go`

```go
// Command deferexit shows that os.Exit skips deferred calls.
package main

import (
	"fmt"
	"os"
)

func main() {
	defer fmt.Println("cleanup") // never runs
	fmt.Println("exiting")
	os.Exit(3)
}
```

```console
$ go run ./ch04/deferexit
exiting
exit status 3
```

`os.Exit` ends the process immediately, so `cleanup` never prints. Deferred work like flushing
output or removing a temp file is lost.

The standard fix is the **`run` pattern**:

```go
func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// All the real work, with defers that run normally.
	return 0
}
```

`run` does the work and returns an exit code; its defers run as it returns. Then `main` calls
`os.Exit` with nothing left to clean up. Taking arguments as a parameter and returning a result
also makes `run` easy to test ([Chapter 19](19-testing.md)).

## A first look at methods

A **method** is a function attached to a type. The extra parameter before its name is the
**receiver**.

`examples/ch04/methods/main.go`

```go
// Command methods previews methods: functions attached to a type.
package main

import "fmt"

type Celsius float64

// Fahrenheit is a method on Celsius. c is the receiver.
func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

func main() {
	t := Celsius(100)
	fmt.Println(t.Fahrenheit())
}
```

```console
$ go run ./ch04/methods
212
```

Call it with a dot: `t.Fahrenheit()`. You can declare methods on any type defined in the same
package. More in [Chapter 7](07-structs-methods-pointers.md).

## Building seek: the first real search

Now `seek` takes exactly one PATH, reads the file, and prints every line containing PATTERN. It
also adopts the `run` pattern and splits into functions.

First, create two test files. The `go` tool ignores folders named `testdata` when building, so
that's the standard place for test inputs. From the project root:

```console
$ mkdir -p testdata
$ printf 'The quick brown fox\njumps over the lazy dog.\nThe dog sleeps; the fox runs.\nFoxes are quick and clever.\n' > testdata/poem.txt
$ printf '# Notes\n- feed the FOX\n- café opens at 9\n- buy coffee\n' > testdata/notes.md
```

Or use your editor; end each file with a newline.

`testdata/poem.txt`:

```text
The quick brown fox
jumps over the lazy dog.
The dog sleeps; the fox runs.
Foxes are quick and clever.
```

`testdata/notes.md`:

```text
# Notes
- feed the FOX
- café opens at 9
- buy coffee
```

New standard library functions:

| Function | Does |
|----------|------|
| `os.ReadFile(path)` | The whole file as `[]byte` (a slice of bytes), plus an `error`. |
| `string(data)` | Converts bytes to a string ([Chapter 6](06-strings.md)). |
| `strings.TrimSuffix(s, suffix)` | `s` without a trailing `suffix`, if present. |
| `strings.Split(s, sep)` | Splits `s` at every `sep` into a `[]string`. |
| `strings.Contains(s, substr)` | Whether `substr` is in `s`. |

A non-`nil` `error` from `os.ReadFile` means something went wrong, such as a missing file. For
now, `seek` prints it and exits with status 2. [Chapter 9](09-errors.md) handles errors properly.

Replace `main.go` with:

```go
// Command seek searches files for lines that match a pattern.
package main

import (
	"fmt"
	"os"
	"strings"
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

// run is the whole program. It returns the exit status.
func run(args []string) int {
	pattern, paths, code, done := parseArgs(args)
	if done {
		return code
	}
	if len(paths) != 1 {
		return usageError("expected exactly one PATH")
	}

	// match is a closure: it captures pattern from the enclosing function.
	match := func(line string) bool {
		return strings.Contains(line, pattern)
	}

	n, err := searchFile(paths[0], match)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seek:", err)
		return exitError
	}
	if n == 0 {
		return exitNoMatch
	}
	return exitOK
}

// parseArgs splits the command line into a pattern and paths. If the
// program should stop right away (after --help, --version, or a usage
// error), done is true and code is the exit status to use.
func parseArgs(args []string) (pattern string, paths []string, code int, done bool) {
	i := 0
loop:
	for ; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-h", "--help":
			fmt.Println(usage)
			return "", nil, exitOK, true
		case "--version":
			fmt.Println(name, version)
			return "", nil, exitOK, true
		case "--":
			i++
			break loop
		default:
			if len(arg) > 1 && arg[0] == '-' {
				return "", nil, usageError("unknown option " + arg), true
			}
			break loop
		}
	}
	if i >= len(args) {
		return "", nil, usageError("missing PATTERN"), true
	}
	return args[i], args[i+1:], 0, false
}

// usageError reports a command-line mistake and returns the exit status.
func usageError(msg string) int {
	fmt.Fprintln(os.Stderr, "seek:", msg)
	fmt.Fprintln(os.Stderr, usage)
	return exitError
}

// searchFile prints every line of the file at path for which match
// returns true, and returns how many lines it printed.
func searchFile(path string, match func(string) bool) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	// Drop the final newline so Split doesn't produce an empty last line.
	text := strings.TrimSuffix(string(data), "\n")

	count := 0
	for _, line := range strings.Split(text, "\n") {
		if match(line) {
			fmt.Println(line)
			count++
		}
	}
	return count, nil
}
```

Function by function:

- **`main`** is one line: pass the arguments (minus the program name) to `run`, exit with its
  result. The comment explains why.
- **`run`** is the program. Each step continues or returns an exit code early, keeping the main
  path unindented.
- **`parseArgs`** is Chapter 3's option loop, moved into a function that returns instead of calling
  `os.Exit`. Its four **named results** document what each means; it still uses explicit, never
  naked, `return`s. When `done` is true, `run` returns `code` right away.
- **`usageError`** prints a message and the usage line to stderr and returns `exitError`, so
  callers write `return usageError("...")`. Unknown options go through it too, so `-x` now also
  gets the usage line, and the message loses its quotes: `seek: unknown option -x`.
- **`match`** is a closure capturing `pattern`. `searchFile` just calls `match(line)` and doesn't
  care how matching works, so later chapters can build smarter matchers (ignore case, invert)
  without touching it.
- **`searchFile`** returns the number of matching lines and an error. Its parameter
  `match func(string) bool` has a function type.
- **`strings.TrimSuffix`** matters: a file ending in `"\n"` would otherwise split into an extra
  empty last line, which would match an empty pattern.

No `defer` yet: `os.ReadFile` opens and closes the file for you. Your first real `defer f.Close()`
comes in [Chapter 11](11-io-and-files.md), when `seek` reads files line by line.

### Checkpoint

```console
$ go run . fox testdata/poem.txt
The quick brown fox
The dog sleeps; the fox runs.
$ go run . dog testdata/poem.txt
jumps over the lazy dog.
The dog sleeps; the fox runs.
$ go run . fox
seek: expected exactly one PATH
Usage: seek [OPTIONS] PATTERN [PATH ...]
exit status 2
$ go run . fox testdata/missing.txt
seek: open testdata/missing.txt: no such file or directory
exit status 2
```

The search is case-sensitive, so `Foxes` doesn't match `fox` (`-i` comes in
[Chapter 6](06-strings.md)). Check exit codes with the binary:

```console
$ go build
$ ./seek fox testdata/poem.txt; echo $?
The quick brown fox
The dog sleeps; the fox runs.
0
$ ./seek cat testdata/poem.txt; echo $?
1
$ ./seek fox testdata/missing.txt; echo $?
seek: open testdata/missing.txt: no such file or directory
2
```

On Windows the missing-file text differs (`The system cannot find the file specified.`).

The complete project at this point is in `project/ch04/`.

## Summary

- `func name(a, b int) (int, error)`: types follow names; several results go in parentheses.
- Multiple results replace output parameters. Ignore one with `_`.
- Named results document; avoid naked `return` outside tiny functions.
- A variadic `...T` comes last and is a slice inside. Spread a slice with `s...`.
- Functions are values with types like `func(string) bool`; zero value `nil`.
- Closures capture variables by reference and keep them alive. Since Go 1.22, each loop iteration
  has its own variable.
- `defer` runs a call when the function returns, LIFO, with arguments evaluated at the `defer`.
  It keeps cleanup next to setup.
- `os.Exit` skips deferred calls. Use `func main() { os.Exit(run(...)) }`.

## Exercises

1. Write `stats(first int, rest ...int) (lo, hi int, mean float64)`, returning the smallest,
   largest, and average value. Why take `first` separately instead of just `nums ...int`?

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   // stats returns the smallest value, the largest value, and the mean.
   // It needs at least one number.
   func stats(first int, rest ...int) (lo, hi int, mean float64) {
   	lo, hi = first, first
   	sum := first
   	for _, n := range rest {
   		if n < lo {
   			lo = n
   		}
   		if n > hi {
   			hi = n
   		}
   		sum += n
   	}
   	return lo, hi, float64(sum) / float64(1+len(rest))
   }

   func main() {
   	lo, hi, mean := stats(3, 9, 4)
   	fmt.Printf("lo=%d hi=%d mean=%.2f\n", lo, hi, mean)
   	fmt.Println(stats(7))
   }
   ```

   ```console
   $ go run .
   lo=3 hi=9 mean=5.33
   7 7 7
   ```

   With only `nums ...int`, a caller could pass nothing, leaving no minimum or maximum. A required
   first parameter makes the compiler enforce "at least one number".

   </details>

2. Write `multiplier(n int) func(int) int`, returning a function that multiplies by `n`, and
   `compose(f, g func(int) int) func(int) int`, returning one that computes `g(f(x))`. Use them to
   build a multiply-by-6 function.

   <details><summary>Solution</summary>

   ```go
   package main

   import "fmt"

   func multiplier(n int) func(int) int {
   	return func(x int) int { return x * n }
   }

   func compose(f, g func(int) int) func(int) int {
   	return func(x int) int { return g(f(x)) }
   }

   func main() {
   	double := multiplier(2)
   	triple := multiplier(3)
   	fmt.Println(double(5), triple(5))

   	sixTimes := compose(double, triple)
   	fmt.Println(sixTimes(7))
   }
   ```

   ```console
   $ go run .
   10 15
   42
   ```

   </details>

3. Predict the output, then run it:

   ```go
   package main

   import "fmt"

   func main() {
   	for i := range 3 {
   		defer fmt.Println("loop", i)
   	}
   	n := 1
   	defer fmt.Println("n =", n)
   	defer func() { fmt.Println("closure n =", n) }()
   	n = 2
   	fmt.Println("body")
   }
   ```

   <details><summary>Solution</summary>

   ```console
   $ go run .
   body
   closure n = 2
   n = 1
   loop 2
   loop 1
   loop 0
   ```

   The body runs first, then the defers in reverse. `fmt.Println("n =", n)` got the value 1 at
   the `defer`; the closure reads `n` at the end, when it's 2.

   </details>

4. In `seek`, make the `match` closure accept only lines that *start with* the pattern (find the
   function with `go doc strings`). Run `go run . The testdata/poem.txt`, then undo the change.

   <details><summary>Solution</summary>

   ```go
   	match := func(line string) bool {
   		return strings.HasPrefix(line, pattern)
   	}
   ```

   ```console
   $ go run . The testdata/poem.txt
   The quick brown fox
   The dog sleeps; the fox runs.
   ```

   Only `match` changed; `searchFile` takes any `func(string) bool`.

   </details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-4--functions-closures-and-defer)
