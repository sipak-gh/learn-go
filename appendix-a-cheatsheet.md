# Appendix A — Cheat Sheet

A one-page reminder per area, each linking to the chapter that explains it. Snippets are
fragments: package-level declarations or lines from inside a function.

## Declarations

```go
package main // every file starts with its package; a program's is main

import (
	_ "embed" // import only for side effects
	"fmt"
	str "strings" // rename an import
)

const Pi = 3.14159  // untyped constant
const big = 1 << 62 // constants are exact at compile time
const (
	KB = 1 << (10 * (iota + 1)) // iota: 0, 1, 2, ... in each const block
	MB
	GB
)

var count int // zero value: 0
var name, ok = "go", true
var (
	x, y float64
	err  error // nil
)

func declarations() {
	n := 42          // short declaration, only inside functions
	a, b := 1, "two" // several at once
	a, c := 3, 4.5   // := needs at least one new name on the left
	n = 7            // plain assignment
	a, n = n, a      // swap
	_ = b            // the blank identifier discards a value
	fmt.Println(str.ToUpper(name), a, c, n, KB, MB, GB, x, y, err, count, ok, Pi, big)
}
```

- Names starting with an upper-case letter are **exported** (visible to other packages).
- Unused local variables and unused imports are compile errors.
- See [Chapter 1](01-getting-started.md), [Chapter 2](02-values-and-types.md),
  [Chapter 8](08-packages-and-modules.md).

## Types

| Type | Zero value | Literal / example |
|------|------------|-------------------|
| `bool` | `false` | `true` |
| `int`, `int8` … `int64`, `uint` … `uint64`, `uintptr` | `0` | `42`, `0x2A`, `0o52`, `0b101010`, `1_000_000` |
| `float32`, `float64` | `0` | `3.14`, `1e-9` |
| `complex64`, `complex128` | `0` | `1 + 2i` |
| `byte` (= `uint8`), `rune` (= `int32`) | `0` | `'a'`, `'\n'`, `'世'` |
| `string` | `""` | `"hi\n"`, `` `raw\n` `` |
| `[N]T` array | N zero `T`s | `[3]int{1, 2, 3}`, `[...]int{1, 2}` |
| `[]T` slice | `nil` | `[]int{1, 2}`, `make([]int, len, cap)` |
| `map[K]V` | `nil` | `map[string]int{"a": 1}`, `make(map[string]int)` |
| `*T` pointer | `nil` | `&v`, `new(T)` |
| `chan T` | `nil` | `make(chan T)`, `make(chan T, 10)` |
| `func(...) ...` | `nil` | `func(x int) int { return x * 2 }` |
| `struct{...}` | all fields zero | `Point{X: 1, Y: 2}` |
| `interface{...}`, `any` | `nil` | any value whose type has the methods |

```go
type Celsius float64 // a new named type
type ID = string     // an alias: the same type, another name

type Point struct {
	X, Y int
	Name string `json:"name,omitempty"` // a struct tag
}

func types() {
	f := float64(3) / 2  // conversions are always explicit: T(v)
	i := int(f)          // 1: truncates toward zero
	s := strconv.Itoa(i) // int to decimal string; string(i) would be a rune
	n, err := strconv.Atoi("42")

	nums := []int{1, 2, 3}
	nums = append(nums, 4, 5) // may reallocate: always use the result
	part := nums[1:3]         // [2 3], shares nums' array
	clone := slices.Clone(part)
	l, c := len(nums), cap(nums)

	m := map[string]int{"a": 1}
	m["b"] = 2
	v, found := m["z"] // 0, false
	delete(m, "a")
	for k, v := range m { // random order
		fmt.Println(k, v)
	}

	p := &Point{X: 1}
	p.Y = 2 // automatic dereference: (*p).Y

	for i, r := range "héllo" { // i is a byte offset, r a rune
		fmt.Println(i, string(r))
	}
	b := "héllo"[1]          // a byte, 0xC3
	runes := []rune("héllo") // 5 runes; len("héllo") is 6 bytes
	fmt.Println(s, n, err, clone, l, c, v, found, b, runes, Celsius(20), ID("x"))
}
```

See [Chapter 2](02-values-and-types.md), [Chapter 5](05-collections.md),
[Chapter 6](06-strings.md), [Chapter 7](07-structs-methods-pointers.md).

## Control flow

```go
func control(x int, items []string, v any) {
	if x > 0 {
	} else if x < 0 {
	} else {
	}
	if n, err := strconv.Atoi("7"); err == nil { // n, err exist only in the if/else
		fmt.Println(n)
	}

	for i := 0; i < 3; i++ { // the only loop keyword is for
	}
	for x < 100 { // like while
		x *= 2
	}
	for { // forever; leave with break or return
		break
	}
	for i, item := range items { // index and copy of the element
		fmt.Println(i, item)
	}
	for i := range 3 { // 0, 1, 2
		fmt.Println(i)
	}

	switch x { // no fallthrough by default
	case 1, 2:
	case 3:
		fallthrough // explicit
	default:
	}
	switch { // switch true: replaces if-else chains
	case x < 0:
	case x > 100:
	}
	switch t := v.(type) { // type switch
	case int:
		fmt.Println(t + 1)
	case string, []byte:
	case nil:
	}

outer:
	for _, a := range items {
		for _, b := range items {
			if a == b {
				continue outer // labels work with break and continue
			}
		}
	}
	goto end // rare
end:
}
```

See [Chapter 3](03-control-flow.md).

## Functions

```go
func add(a, b int) int { return a + b }

func divmod(a, b int) (int, int) { return a / b, a % b } // multiple results

func split(sum int) (x, y int) { // named results start at their zero value
	x = sum * 4 / 9
	y = sum - x
	return // a "naked" return returns x, y
}

func sum(nums ...int) int { // variadic: nums is a []int
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func counter() func() int { // a closure captures variables, not values
	n := 0
	return func() int { n++; return n }
}

func functions() {
	q, r := divmod(7, 2)
	_, r = divmod(9, 4)
	nums := []int{1, 2, 3}
	fmt.Println(sum(), sum(1, 2), sum(nums...), q, r)

	f, err := os.Open("x.txt")
	if err != nil {
		return
	}
	defer f.Close() // runs when the function returns; LIFO; arguments evaluated now

	double := func(x int) int { return x * 2 } // function literal
	fmt.Println(double(add(1, 2)), counter()())
}
```

- Arguments are passed **by value** (copied). Slices, maps, channels, pointers, and functions
  are small values that refer to shared data.
- `main` calls `os.Exit(run())` so deferred calls in `run` still happen.
- See [Chapter 4](04-functions.md).

## Methods and interfaces

```go
type Rect struct{ W, H float64 }

func (r Rect) Area() float64    { return r.W * r.H }   // value receiver: gets a copy
func (r *Rect) Scale(k float64) { r.W *= k; r.H *= k } // pointer receiver: can modify

type Shape interface { // satisfied implicitly: no "implements"
	Area() float64
}

var _ Shape = Rect{} // compile-time check that Rect implements Shape

type Named struct {
	Rect // embedding: Named gets Area and Scale (promoted)
	Name string
}

func (r Rect) String() string { return fmt.Sprintf("%gx%g", r.W, r.H) } // fmt.Stringer

func methods() {
	r := Rect{W: 2, H: 3}
	r.Scale(2) // Go takes &r automatically
	var s Shape = r
	if rect, ok := s.(Rect); ok { // type assertion with ok
		fmt.Println(rect.W)
	}
	n := Named{Rect: r, Name: "box"}
	fmt.Println(s.Area(), n.Area(), n.Rect.H, r) // prints r with String
}
```

- Small interfaces: `io.Reader`, `io.Writer`, `fmt.Stringer`, `error`, `sort.Interface`.
- An interface holding a nil pointer is **not** nil.
- "Accept interfaces, return structs."
- See [Chapter 7](07-structs-methods-pointers.md), [Chapter 10](10-interfaces.md).

## Generics

```go
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

type Number interface {
	~int | ~int64 | ~float64 // a type set; ~ includes named types like Celsius
}

func Sum[N Number](nums []N) N {
	var total N
	for _, n := range nums {
		total += n
	}
	return total
}

type Stack[T any] struct{ items []T } // a generic type

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func generics() {
	strs := Map([]int{1, 2}, strconv.Itoa) // type arguments inferred
	var st Stack[string]                   // explicit type argument
	st.Push("a")
	fmt.Println(strs, Sum([]float64{1.5, 2}), max(3, 7), min("b", "a"))
	fmt.Println(slices.Index(strs, "2"), slices.Sorted(maps.Keys(map[string]int{"b": 1, "a": 2})))
}
```

Built-in constraints: `any`, `comparable` (`==` works). From package `cmp`: `cmp.Ordered` (`<`
works). See [Chapter 13](13-generics.md).

## Iterators

```go
func Count(n int) iter.Seq[int] { // iter.Seq[V] is func(yield func(V) bool)
	return func(yield func(int) bool) {
		for i := range n {
			if !yield(i) { // false: the loop broke; stop
				return
			}
		}
	}
}

func iterators() {
	for i := range Count(3) {
		fmt.Println(i)
	}
	m := map[string]int{"a": 1}
	for k, v := range maps.All(m) { // iter.Seq2[K, V]
		fmt.Println(k, v)
	}
	keys := slices.Collect(maps.Keys(m))
	for i, line := range slices.All([]string{"x"}) {
		fmt.Println(i, line, keys)
	}
}
```

See [Chapter 14](14-iterators.md).

## Concurrency

```go
func concurrency(ctx context.Context) {
	ch := make(chan int)      // unbuffered: send waits for a receiver
	buf := make(chan int, 10) // buffered: send waits only when full
	go func() {               // start a goroutine
		defer close(ch) // only the sender closes
		for i := range 3 {
			ch <- i
		}
	}()
	for v := range ch { // until ch is closed and empty
		fmt.Println(v)
	}
	v, ok := <-ch // ok is false: closed

	select { // wait for whichever is ready first
	case buf <- 1:
	case <-ctx.Done():
		fmt.Println(context.Cause(ctx))
	case <-time.After(time.Second):
	default: // makes select non-blocking
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var hits atomic.Int64
	total := 0
	for range 4 {
		wg.Go(func() { // Go 1.25+: Add(1), go, Done
			mu.Lock()
			defer mu.Unlock()
			total++
			hits.Add(1)
		})
	}
	wg.Wait()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel() // always
	fmt.Println(v, ok, total, hits.Load(), ctx.Err())
}
```

- Directional channels: `chan<- T` (send only), `<-chan T` (receive only).
- `sync.Once`, `sync.OnceValue`, `sync.RWMutex`, `sync.Pool` (rare).
- Run tests and programs with `-race`.
- See [Chapter 16](16-goroutines-and-channels.md), [Chapter 17](17-sync.md),
  [Chapter 18](18-context-and-time.md).

## Errors

```go
var ErrNotFound = errors.New("not found") // a sentinel error

type ParseError struct { // a custom error type
	Line int
	Err  error
}

func (e *ParseError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }
func (e *ParseError) Unwrap() error { return e.Err }

func find(key string) error {
	return fmt.Errorf("find %q: %w", key, ErrNotFound) // %w wraps
}

func errorsDemo() {
	err := find("x")
	if errors.Is(err, ErrNotFound) { // anywhere in the chain
		fmt.Println("missing:", err)
	}
	var pe *ParseError
	if errors.As(err, &pe) { // finds a *ParseError in the chain
		fmt.Println(pe.Line)
	}
	err = errors.Join(err, io.EOF) // several errors in one

	defer func() {
		if r := recover(); r != nil { // only in deferred functions
			fmt.Println("recovered:", r)
		}
	}()
	panic("bug") // for programmer errors, not for bad input
}
```

See [Chapter 9](09-errors.md).

## Testing

```go
// file: abs_test.go, in package abs (or abs_test for black-box tests)

func TestAbs(t *testing.T) {
	tests := []struct {
		name    string
		in, out int
	}{
		{"positive", 3, 3},
		{"negative", -3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Abs(tt.in); got != tt.out {
				t.Errorf("Abs(%d) = %d, want %d", tt.in, got, tt.out) // keeps going
			}
		})
	}
}

func helper(t *testing.T) string {
	t.Helper()           // failures point at the caller
	dir := t.TempDir()   // deleted after the test
	t.Cleanup(func() {}) // runs after the test
	return dir
}

func BenchmarkAbs(b *testing.B) {
	for b.Loop() {
		Abs(-5)
	}
}

func FuzzAbs(f *testing.F) {
	f.Add(-5)
	f.Fuzz(func(t *testing.T, n int) {
		if n != math.MinInt && Abs(n) < 0 {
			t.Fatalf("Abs(%d) < 0", n) // stops this test
		}
	})
}

func ExampleAbs() {
	fmt.Println(Abs(-2))
	// Output: 2
}
```

```text
go test ./...                         all packages
go test -v -run 'TestAbs/negative' .  verbose; regexp per subtest level
go test -count=1 -race ./...          no cache; race detector
go test -cover -coverprofile=c.out .  coverage; go tool cover -html=c.out
go test -run '^$' -bench . -benchmem  benchmarks only
go test -run '^$' -fuzz FuzzAbs -fuzztime 30s
```

Also: `testdata/` directories, golden files with an `-update` flag, `testing/fstest.MapFS`,
`t.Context()`. See [Chapter 19](19-testing.md).

## The `go` command

| Command | Does |
|---------|------|
| `go mod init example.com/m` | Start a module (writes `go.mod`) |
| `go run .` / `go run ./cmd/x` | Build and run |
| `go build -o bin/x ./cmd/x` | Build a binary |
| `go install ./cmd/x`, `go install pkg@latest` | Build into `$GOBIN` or `$(go env GOPATH)/bin` |
| `go test ./...` | Run tests (see above) |
| `go fmt ./...` / `gofmt -l .` | Format / list unformatted files |
| `go vet ./...` | Report likely bugs |
| `go doc strings.Cut`, `go doc -src strings.Cut` | Documentation / source |
| `go get example.com/pkg@v1.2.3` | Add or change a dependency |
| `go mod tidy` | Sync `go.mod`/`go.sum` with the imports |
| `go work init ./a ./b` | Work on several modules together |
| `go generate ./...` | Run `//go:generate` commands |
| `go list -f '{{.GoFiles}}' ./...` | Inspect packages |
| `go env GOPATH GOOS`, `go version`, `go version -m bin/x` | Environment / build info |
| `go tool pprof -top x.test cpu.out` | Read a profile |
| `go clean -cache -testcache` | Delete cached builds / test results |

Build flags:

```text
GOOS=windows GOARCH=amd64 go build ...   cross-compile (go tool dist list)
go build -tags debug                     set build tags for //go:build
go build -ldflags "-X main.version=1.0"  set a string variable
go build -trimpath -ldflags "-s -w"      no local paths; smaller binary
go run -race .                           race detector
CGO_ENABLED=0 go build                   pure Go, no C toolchain
```

Directives:

```text
//go:build linux && !arm64     (before package; then a blank line)
//go:embed usage.txt           (above a package-level string, []byte, or embed.FS var)
//go:generate go run gen.go
```

See [Chapter 1](01-getting-started.md), [Chapter 8](08-packages-and-modules.md),
[Chapter 20](20-tooling-and-design.md).

## `fmt` verbs

| Verb | Prints | Example output |
|------|--------|----------------|
| `%v`, `%+v`, `%#v` | default, with field names, Go syntax | `{1 2}`, `{X:1 Y:2}`, `main.Point{X:1, Y:2}` |
| `%T` | the type | `main.Point` |
| `%d`, `%x`, `%o`, `%b` | integer in base 10, 16, 8, 2 | `42`, `2a`, `52`, `101010` |
| `%f`, `%.2f`, `%e`, `%g` | float (here `math.Pi`) | `3.141593`, `3.14`, `3.141593e+00`, `3.141592653589793` |
| `%s`, `%q` | string, quoted string | `hi`, `"hi"` |
| `%c`, `%U` | rune, Unicode code point | `世`, `U+4E16` |
| `%t` | bool | `true` |
| `%p` | pointer | `0xc000012345` |
| `%5d`, `%-5s`, `%05d` | width, left-align, zero-pad | `   42`, `hi   `, `00042` |
| `%w` | (only in `fmt.Errorf`) wrap an error | |
| `%%` | a literal `%` | `%` |

See [Chapter 6](06-strings.md).

Sources for this appendix: [REFERENCES.md](REFERENCES.md#appendix-a--cheat-sheet)
