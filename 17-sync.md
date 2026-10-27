# Chapter 17 — Synchronization and the Memory Model

[Chapter 16](16-goroutines-and-channels.md) shared data over channels. Sometimes goroutines must
touch the *same* variable: a counter, a cache, a flag. You'll see what goes wrong (a *data race*),
how the race detector finds it, the `sync` and `sync/atomic` tools that make sharing safe, and the
rules behind them: Go's memory model. In the project, workers share a byte counter for `--stats`,
and everything runs under the race detector. [Chapter 18](18-context-and-time.md) adds
cancellation.

## Data races

A **data race**: two goroutines access the same variable at the same time, at least one writes,
and nothing orders them. This program has one:

`examples/ch17/race/main.go`

```go
package main

import "fmt"

func main() {
	count := 0
	done := make(chan bool)
	for range 4 {
		go func() {
			for range 100_000 {
				count++ // read, add, write: three steps, not one
			}
			done <- true
		}()
	}
	for range 4 {
		<-done
	}
	fmt.Println("count:", count, "(want 400000)")
}
```

```console
$ go run ./ch17/race
count: 285252 (want 400000)
$ go run ./ch17/race
count: 266349 (want 400000)
```

(The numbers change every run.) `count++` is three steps: read, add 1, write back. Two goroutines
can both read 41, both write 42, and one increment is lost.

Lost updates are the *mild* outcome. Go guarantees nothing for a racy program:

- The compiler may keep a value in a register and never see another goroutine's write.
- Strings, slices, and interfaces span several machine words. A racy read can see half an old
  value and half a new one, and crash.

Every data race is a bug, even one that "works".

### The race detector

Add `-race` to `go run`, `go build`, or `go test`. Every memory access gets checked, and any race
that actually happens is reported:

```console
$ go run -race ./ch17/race
==================
WARNING: DATA RACE
Read at 0x00c0001a6028 by goroutine 11:
  main.main.func1()
      /home/you/examples/ch17/race/main.go:11 +0x44

Previous write at 0x00c0001a6028 by goroutine 8:
  main.main.func1()
      /home/you/examples/ch17/race/main.go:11 +0x56

Goroutine 11 (running) created at:
  main.main()
      /home/you/examples/ch17/race/main.go:9 +0x66

Goroutine 8 (running) created at:
  main.main()
      /home/you/examples/ch17/race/main.go:9 +0x66
==================
==================
WARNING: DATA RACE
Write at 0x00c0001a6028 by goroutine 11:
  main.main.func1()
      /home/you/examples/ch17/race/main.go:11 +0x56

Previous write at 0x00c0001a6028 by goroutine 8:
  main.main.func1()
      /home/you/examples/ch17/race/main.go:11 +0x56

Goroutine 11 (running) created at:
  main.main()
      /home/you/examples/ch17/race/main.go:9 +0x66

Goroutine 8 (running) created at:
  main.main()
      /home/you/examples/ch17/race/main.go:9 +0x66
==================
count: 341390 (want 400000)
Found 2 data race(s)
exit status 66
```

Reading a report:

1. **The two accesses**: "Read at ... by goroutine 11" and "Previous write at ... by goroutine 8",
   each with a stack trace. Both are line 11, `count++`.
2. **Where each goroutine started**: "Goroutine 11 (running) created at" points to the `go`
   statement on line 9.
3. Exit status **66**. Under `go test -race`, the test fails.

About `-race`:

- It finds races that **happen during this run**, not every race that *could*. Run it often (e.g.
  in CI) on tests that exercise concurrent code.
- No false positives: every report is a real bug.
- Several times slower, more memory. For testing, not production builds.
- Works on the common 64-bit platforms (Linux, macOS, Windows, FreeBSD).

## `sync.WaitGroup`: wait for goroutines

Chapter 16 counted messages on a `done` channel. `sync.WaitGroup` does the counting for you:

- `wg.Add(n)` adds `n` to a counter. Call it **before** starting the goroutine.
- `wg.Done()` subtracts 1 when the goroutine finishes, usually via `defer`.
- `wg.Wait()` blocks until the counter is zero.
- `wg.Go(f)` (Go 1.25+) does Add(1), starts `f`, and calls Done when `f` returns. Prefer it.

`examples/ch17/waitgroup/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	words := []string{"alpha", "beta", "gamma"}
	lengths := make([]int, len(words))

	// The classic way: Add before starting, Done when finished.
	var wg sync.WaitGroup
	for i, w := range words {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lengths[i] = len(w) // each goroutine writes its own element
		}()
	}
	wg.Wait() // blocks until the counter is back to zero
	fmt.Println(lengths)

	// Since Go 1.25, wg.Go does the Add and Done for you.
	upper := make([]string, len(words))
	for i, w := range words {
		wg.Go(func() {
			upper[i] = fmt.Sprintf("%s!", w)
		})
	}
	wg.Wait()
	fmt.Println(upper)
}
```

```console
$ go run ./ch17/waitgroup
[5 4 5]
[alpha! beta! gamma!]
```

No race, even though goroutines write to one slice: each writes a **different element**, and
`main` reads only after `Wait`. Common mistakes:

- `wg.Add(1)` *inside* the goroutine: `Wait` may run first and return too early. `wg.Go` makes
  this impossible.
- Forgetting `Done`: `Wait` blocks forever.
- Copying a `WaitGroup` (passing it by value). Pass a pointer, or use a closure as above.

## `sync.Mutex`: one goroutine at a time

A *mutex* (mutual exclusion lock) protects some state. `mu.Lock()` waits until no one holds the
lock, then takes it; `mu.Unlock()` releases it. The code between is a *critical section*: only one
goroutine at a time can be inside.

`examples/ch17/mutex/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

// Counter is safe for concurrent use. The mutex guards counts: every
// access to counts happens between Lock and Unlock.
type Counter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *Counter) Add(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[string]int)
	}
	c.counts[key]++
}

func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

func main() {
	var c Counter // the zero Mutex is unlocked and ready to use
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			c.Add("fox")
		})
	}
	wg.Wait()
	fmt.Println("fox:", c.Get("fox"))
}
```

```console
$ go run ./ch17/mutex
fox: 100
```

Mutex habits:

- **Put the mutex next to what it guards** (same struct) and say so in a comment.
- **`defer mu.Unlock()`** right after `Lock`: released even on early return or panic.
- **Keep critical sections short.** No I/O or unknown code while locked.
- **Not reentrant.** `Lock` twice in one goroutine deadlocks.
- **Never copy a mutex** after first use; use pointer receivers. `go vet` catches copies:

```go
func (c Counter) Get() int { // value receiver: c is a copy, mutex included
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
```

```console
$ go vet .
main.go:13:9: Get passes lock by value: counter.Counter contains sync.Mutex
```

A plain `map` isn't safe for concurrent use if anyone writes; concurrent writes often crash with
`fatal error: concurrent map writes`. Guard it with a mutex, as above. (`sync.Map` suits a few
special cases, such as caches that only grow; a map plus a mutex is usually simpler.)

### `sync.RWMutex`: many readers, one writer

An `RWMutex` has two modes:

- `RLock`/`RUnlock`: a *shared* lock. Many readers may hold it at once.
- `Lock`/`Unlock`: an *exclusive* lock. It waits until all readers leave.

Use it when reads far outnumber writes.

`examples/ch17/rwmutex/main.go`

```go
package main

import (
	"fmt"
	"strings"
	"sync"
)

// Cache remembers the upper-case form of strings. Many goroutines may
// read at once; a writer gets exclusive access.
type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func (c *Cache) Get(key string) string {
	c.mu.RLock() // shared: other readers may hold it too
	v, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return v
	}

	c.mu.Lock() // exclusive: waits for all readers to leave
	defer c.mu.Unlock()
	// Another goroutine may have filled it while we had no lock.
	if v, ok := c.items[key]; ok {
		return v
	}
	v = strings.ToUpper(key)
	c.items[key] = v
	return v
}

func main() {
	c := &Cache{items: make(map[string]string)}
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			c.Get("fox")
			c.Get("dog")
		})
	}
	wg.Wait()
	fmt.Println(len(c.items), c.Get("fox"), c.Get("dog"))
}
```

```console
$ go run ./ch17/rwmutex
2 FOX DOG
```

You can't upgrade a read lock: calling `Lock` while holding `RLock` deadlocks. Release the read lock,
take the write lock, then check again, as `Get` does.

## `sync.Once` and friends: do it exactly once

`once.Do(f)` runs `f` on the first call and never again, even if many goroutines call it at once;
the others wait for the first call to finish. It's the safe way to do lazy setup.

Go 1.21 added three helpers that wrap a function instead:

| Helper | Returns | Use |
|--------|---------|-----|
| `sync.OnceFunc(f)` | `func()` | run `f` once |
| `sync.OnceValue(f)` | `func() T` | compute a value once, return it every time |
| `sync.OnceValues(f)` | `func() (T, error)` | same, for a value and an error |

`examples/ch17/once/main.go`

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var (
	loadOnce sync.Once
	words    []string
)

func loadWords() {
	fmt.Println("loading words")
	words = strings.Fields("alpha beta gamma")
}

// OnceFunc returns a function that runs its argument only the first
// time it is called.
var greet = sync.OnceFunc(func() { fmt.Println("hello, once") })

// OnceValue caches a result: the function runs once, and every call
// returns the same value.
var answer = sync.OnceValue(func() int {
	fmt.Println("computing the answer")
	return 6 * 7
})

// OnceValues does the same for a (value, error) pair.
var port = sync.OnceValues(func() (int, error) {
	fmt.Println("parsing the port")
	return strconv.Atoi("8080")
})

func main() {
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			loadOnce.Do(loadWords) // every goroutine waits for the one call
			greet()
		})
	}
	wg.Wait()
	fmt.Println(words)

	fmt.Println(answer(), answer())
	p, err := port()
	fmt.Println(p, err)
	p, err = port()
	fmt.Println(p, err)
}
```

```console
$ go run ./ch17/once
loading words
hello, once
[alpha beta gamma]
computing the answer
42 42
parsing the port
8080 <nil>
8080 <nil>
```

If the function panics, `Once.Do` counts it as done anyway. The `OnceFunc`/`OnceValue` helpers
re-panic with the same value on every call instead.

## `sync/atomic`: lock-free counters and flags

For a single number or pointer, a mutex is overkill. `sync/atomic` has **typed atomic values**
whose methods are indivisible steps, safe from any goroutine:

| Type | Main methods |
|------|--------------|
| `atomic.Int32`, `Int64`, `Uint32`, `Uint64`, `Uintptr` | `Load`, `Store`, `Add`, `Swap`, `CompareAndSwap`, `And`, `Or` |
| `atomic.Bool` | `Load`, `Store`, `Swap`, `CompareAndSwap` |
| `atomic.Pointer[T]` | `Load`, `Store`, `Swap`, `CompareAndSwap` |
| `atomic.Value` | the same, for `any` (older; prefer `Pointer[T]`) |

`examples/ch17/atomic/main.go`

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type config struct {
	name  string
	limit int
}

func main() {
	var hits atomic.Int64 // the zero value is 0 and ready to use
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 100 {
				hits.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("hits:", hits.Load())

	// CompareAndSwap changes the value only if it still holds what you
	// expect. Exactly one goroutine wins the race to become "leader".
	var leader atomic.Int32
	var winners atomic.Int32
	for id := int32(1); id <= 10; id++ {
		wg.Go(func() {
			if leader.CompareAndSwap(0, id) {
				winners.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("winners:", winners.Load(), "leader set:", leader.Load() != 0)

	// atomic.Bool and atomic.Pointer[T] work the same way.
	var ready atomic.Bool
	var cfg atomic.Pointer[config]
	cfg.Store(&config{"default", 10})
	ready.Store(true)
	old := cfg.Swap(&config{"tuned", 20}) // replace the whole value at once
	fmt.Println(ready.Load(), old.name, cfg.Load().name, cfg.Load().limit)
}
```

```console
$ go run ./ch17/atomic
hits: 5000
winners: 1 leader set: true
true default tuned 20
```

Rules:

- **Every access must be atomic.** The typed values enforce this: the number is hidden inside and
  reachable only through the methods.
- **Atomics protect one value, not a group.** If two fields must change together, use a mutex.
  `Load` and then act is also a race (the value may change in between); use `CompareAndSwap` or a
  mutex.
- **Don't copy them.** Like a mutex, an atomic value contains a `noCopy` marker that `go vet` checks.

Older functions such as `atomic.AddInt64(&x, 1)` work on plain integers, but nothing stops other
code from touching `x` directly. Prefer the types.

## `sync.Pool`: reuse temporary objects

A `sync.Pool` keeps spare objects so hot code can reuse them instead of allocating. `Get` returns a
pooled object (or calls `New`); `Put` gives one back.

`examples/ch17/pool/main.go`

```go
package main

import (
	"bytes"
	"fmt"
	"sync"
)

// bufPool hands out reusable buffers. New makes one when the pool is
// empty.
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func render(name string, n int) string {
	buf := bufPool.Get().(*bytes.Buffer) // Get returns any
	defer bufPool.Put(buf)
	buf.Reset() // a reused buffer still holds old data
	fmt.Fprintf(buf, "%s=%d", name, n)
	return buf.String()
}

func main() {
	var wg sync.WaitGroup
	results := make([]string, 4)
	for i := range results {
		wg.Go(func() {
			results[i] = render("item", i)
		})
	}
	wg.Wait()
	fmt.Println(results)
}
```

```console
$ go run ./ch17/pool
[item=0 item=1 item=2 item=3]
```

The garbage collector may empty a pool at any time: it's a cache, never storage. Use it only when a
profiler ([Chapter 20](20-tooling-and-design.md)) shows allocation is a real cost. `fmt` and
`encoding/json` use pools internally.

## Channels or mutexes?

Both are correct. Use whichever makes the code simplest.

| Use a channel to... | Use a mutex (or atomic) to... |
|---------------------|-------------------------------|
| hand data from one goroutine to another (transfer ownership) | guard the internal state of a struct (cache, counter, registry) |
| distribute work (worker pools, pipelines) | protect a few lines that read and update shared fields |
| deliver results asynchronously | count things (an atomic is enough) |
| signal events: "done", "stop", "ready" (close the channel) | |

Signs you picked wrong: a channel used as a lock (`chan struct{}` with capacity 1 around every
access), or a mutex plus a condition loop that is really waiting for a message.

### `errgroup`, briefly

"Run these tasks in goroutines, wait for all, return the first error" is a common need.
`errgroup`, in the module `golang.org/x/sync` (maintained by the Go team, outside the standard
library), does it:

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(4) // at most 4 at once
for _, p := range paths {
	g.Go(func() error { return process(ctx, p) })
}
if err := g.Wait(); err != nil { // the first error, after all have finished
	return err
}
```

`seek` doesn't use it: it sticks to the standard library and reports every file error instead of
stopping at the first.

## The memory model in plain words

Without synchronization, **nothing guarantees** one goroutine ever sees another's writes, or sees
them in order: CPUs and compilers reorder memory operations for speed. The *Go memory model* says
a write is guaranteed visible when it **happens before** the read.

Inside one goroutine, statements happen in program order. Between goroutines, only synchronization
creates happens-before edges:

| Event A | happens before event B |
|---------|------------------------|
| `go f()` statement | `f` starts running |
| a send on a channel | the matching receive completes |
| `close(ch)` | a receive that returns because `ch` is closed |
| a receive from an **unbuffered** channel | the matching send completes |
| `mu.Unlock()` | the next `mu.Lock()` returns |
| `wg.Done()` (or the end of a `wg.Go` function) | `wg.Wait()` returns |
| `f` returns inside `once.Do(f)` | any `once.Do` call returns |
| an atomic write | an atomic read that sees the value written |

Happens-before is transitive (A before B, B before C → A before C). So in the waitgroup example
`main` may read `lengths`: each write → that goroutine's `Done` → `Wait` returns → the read.

Rule: **if two goroutines touch one variable and one writes, connect them with one of these
edges.** Can't name the edge? It's a race. Don't get clever with "benign" races or busy-waiting on
a plain `bool`: the compiler may hoist the read out of the loop, which then spins forever.

## Building seek: race-free statistics

`--stats` gains a `bytes read:` line: the total size of all input. Workers read files at the same
time and add to one shared counter, so it must be atomic. You'll also start goroutines with
`wg.Go`, make `SearchAll` wait for them, and write the project's first test, run under `-race`.

### Count bytes in `internal/search/search.go`

Add `"sync/atomic"` to the imports:

```go
import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"sync/atomic"
)
```

Replace the `Searcher` type with this version, and add `BytesRead` and `countingReader` after it:

```go
// A Searcher finds the lines of a file that match a pattern.
type Searcher struct {
	Matcher  Matcher   // decides which lines match; must not be nil
	Invert   bool      // select the lines that do NOT match
	MaxCount int       // stop after this many matches per file; 0 means no limit
	Stdin    io.Reader // read for the path "-"

	// bytesRead is updated by many goroutines at once (see SearchAll),
	// so it must be atomic. A plain int64 would be a data race.
	bytesRead atomic.Int64
}

// BytesRead returns the number of bytes read so far by all searches.
func (s *Searcher) BytesRead() int64 {
	return s.bytesRead.Load()
}

// countingReader is an io.Reader that adds the number of bytes read
// through it to a shared counter.
type countingReader struct {
	r io.Reader
	n *atomic.Int64
}

func (c countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n.Add(int64(n))
	return n, err
}
```

- `countingReader` wraps any reader, the same composition idea as `bufio.Reader`
  ([Chapter 11](11-io-and-files.md)). It holds a *pointer* to the counter: every file's reader must
  add to the one counter in the `Searcher`, not a copy.
- `bytesRead` is unexported. Outside code can read it through `BytesRead` but can't change it.

At the top of `Search`, wrap the reader. The rest of the function stays the same:

```go
func (s *Searcher) Search(r io.Reader, path string) ([]Match, error) {
	r = countingReader{r: r, n: &s.bytesRead}

	// Peek at the start of the input without consuming it, so the
	// scanner below still sees every byte.
	br := bufio.NewReaderSize(r, binaryCheckSize)
```

`Searcher` now contains an atomic value, so **never copy a `Searcher`**; always use `*Searcher`, as
`run` already does. Writing `copied := *s` in `run` would make `go vet` report:

```text
cmd/seek/main.go:76:12: assignment copies lock value to copied: seek/internal/search.Searcher contains sync/atomic.Int64 contains sync/atomic.noCopy
```

### `wg.Go` in `internal/search/parallel.go`

Add `"sync"` to the imports:

```go
import (
	"iter"
	"sync"
)
```

Update `SearchAll`: start the producer and workers with `wg.Go`, and wait for them after the
consumer loop ends normally. The full function:

```go
func (s *Searcher) SearchAll(files iter.Seq2[string, error], workers int) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		// Closing done tells every goroutine to stop. The deferred close
		// runs however we leave, including when the caller's loop breaks.
		done := make(chan struct{})
		defer close(done)

		jobs := make(chan job)
		// The buffer of pending limits how far the producer can get
		// ahead of the consumer, which bounds memory use.
		pending := make(chan chan Result, workers)

		// A WaitGroup counts running goroutines; wg.Go starts one and
		// counts it until it returns.
		var wg sync.WaitGroup

		// Producer.
		wg.Go(func() {
			defer close(pending) // no more results: ends the consumer's loop
			defer close(jobs)    // no more jobs: ends the workers' loops
			for path, err := range files {
				// Buffered, so a worker can always deliver its result and
				// move on, even if the consumer isn't ready for it yet.
				result := make(chan Result, 1)
				if err != nil {
					result <- Result{Path: path, Err: err}
				}
				select {
				case pending <- result:
				case <-done:
					return
				}
				if err != nil {
					continue
				}
				select {
				case jobs <- job{path: path, result: result}:
				case <-done:
					return
				}
			}
		})

		// Workers.
		for range workers {
			wg.Go(func() {
				for j := range jobs {
					matches, err := s.SearchFile(j.path)
					j.result <- Result{Path: j.path, Matches: matches, Err: err}
				}
			})
		}

		// Consumer: results in order.
		for result := range pending {
			if !yield(<-result) {
				// The deferred close(done) stops the other goroutines.
				// We don't wait for them: the caller wants to move on.
				return
			}
		}
		// Every result has been delivered, so the goroutines are all
		// finishing. Waiting guarantees that none outlives SearchAll.
		wg.Wait()
	}
}
```

Changes: `go func() {...}()` became `wg.Go(func() {...})` (twice), plus the two new comment blocks
and `wg.Wait()` at the end.

Why wait only on the normal path?

- Loop ended normally: every result was received, so the goroutines are already leaving. `Wait`
  costs nothing and guarantees none outlives `SearchAll`.
- Early `break`: the caller wants to move on now. The goroutines still stop (the done channel sees
  to that), just without the caller waiting.

`Wait` also creates a happens-before edge: everything the workers did, including their adds to
`bytesRead`, happens before `SearchAll` returns. (`BytesRead` is atomic, so it would be safe anyway;
the edge matters for plain data.)

### Print it in `cmd/seek/main.go`

`printStats` gets a `bytes` parameter and a new line:

```go
// printStats prints a summary of the search to w.
func printStats(w io.Writer, searched int, bytes int64, matched *stats.Counter[string], elapsed time.Duration) {
	fmt.Fprintf(w, "files searched: %d\n", searched)
	fmt.Fprintf(w, "bytes read:     %d\n", bytes)
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

In `run`, pass the count:

```go
	if opts.stats {
		out.Flush() // print the results before the statistics
		printStats(stderr, searched, s.BytesRead(), &matched, time.Since(start))
	}
```

### A first test: `internal/search/parallel_test.go`

The race detector only finds races in code that runs, so drive `SearchAll` hard in a test.
[Chapter 19](19-testing.md) covers testing; for now, four facts:

- A file ending in `_test.go` is compiled only by `go test`.
- In it, a function `func TestXxx(t *testing.T)` is a test.
- `t.Errorf` reports a failure and continues; `t.Fatal`/`t.Fatalf` report and stop the test.
- `t.TempDir()` returns a new empty directory, deleted when the test ends.

Create `internal/search/parallel_test.go`:

```go
package search

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestSearchAllOrder checks that results come back in input order even
// though many workers search files at the same time. Run it with -race.
func TestSearchAllOrder(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for i := range 50 {
		path := filepath.Join(dir, fmt.Sprintf("file%02d.txt", i))
		text := fmt.Sprintf("fox %d\n", i)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}

	s := &Searcher{Matcher: literal("fox")}
	i := 0
	for res := range s.SearchAll(Walk(paths, WalkOptions{}), 8) {
		if res.Err != nil {
			t.Fatal(res.Err)
		}
		if res.Path != paths[i] {
			t.Fatalf("result %d: got %s, want %s", i, res.Path, paths[i])
		}
		i++
	}
	if i != len(paths) {
		t.Fatalf("got %d results, want %d", i, len(paths))
	}
	// Each file holds "fox N\n": 6 bytes for N < 10, 7 bytes otherwise.
	if got, want := s.BytesRead(), int64(10*6+40*7); got != want {
		t.Errorf("BytesRead() = %d, want %d", got, want)
	}
}
```

The test is in `package search`, so it can use the unexported `literal` type. It searches 50 files
with 8 workers and checks the result order and the byte total (340).

### Try it: see the race the atomic prevents

Before the checkpoint, see what happens with a plain `int64` counter. Temporarily make these four
edits in `search.go` (and remove the `"sync/atomic"` import):

```go
	bytesRead int64           // was: atomic.Int64
	return s.bytesRead        // in BytesRead; was: s.bytesRead.Load()
	n *int64                  // in countingReader; was: *atomic.Int64
	*c.n += int64(n)          // in Read; was: c.n.Add(int64(n))
```

Run the tests without and with `-race`:

```console
$ go test -count=1 ./internal/search
ok  	seek/internal/search	0.005s
$ go test -race -count=1 ./internal/search
==================
WARNING: DATA RACE
Read at 0x00c00001aa70 by goroutine 14:
  seek/internal/search.countingReader.Read()
      /home/you/seek/internal/search/search.go:57 +0x8f
  seek/internal/search.(*countingReader).Read()
      <autogenerated>:1 +0x12
  bufio.(*Reader).fill()
      /usr/local/go/src/bufio/bufio.go:113 +0x2a3
  bufio.(*Reader).Peek()
      /usr/local/go/src/bufio/bufio.go:152 +0xaf
  seek/internal/search.(*Searcher).Search()
      /home/you/seek/internal/search/search.go:89 +0x1e4
  seek/internal/search.(*Searcher).SearchFile()
      /home/you/seek/internal/search/search.go:77 +0x1f7
  seek/internal/search.TestSearchAllOrder.(*Searcher).SearchAll.func2.2()
      /home/you/seek/internal/search/parallel.go:79 +0x86

Previous write at 0x00c00001aa70 by goroutine 11:
  seek/internal/search.countingReader.Read()
      /home/you/seek/internal/search/search.go:57 +0xa4
  seek/internal/search.(*countingReader).Read()
      <autogenerated>:1 +0x12
  bufio.(*Reader).fill()
      /usr/local/go/src/bufio/bufio.go:113 +0x2a3
  bufio.(*Reader).Peek()
      /usr/local/go/src/bufio/bufio.go:152 +0xaf
  seek/internal/search.(*Searcher).Search()
      /home/you/seek/internal/search/search.go:89 +0x1e4
  seek/internal/search.(*Searcher).SearchFile()
      /home/you/seek/internal/search/search.go:77 +0x1f7
  seek/internal/search.TestSearchAllOrder.(*Searcher).SearchAll.func2.2()
      /home/you/seek/internal/search/parallel.go:79 +0x86

Goroutine 14 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  seek/internal/search.TestSearchAllOrder()
      /home/you/seek/internal/search/parallel_test.go:26 +0x62e
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2036 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2101 +0x38

Goroutine 11 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  seek/internal/search.TestSearchAllOrder()
      /home/you/seek/internal/search/parallel_test.go:26 +0x62e
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2036 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2101 +0x38
==================
...
--- FAIL: TestSearchAllOrder (0.01s)
    parallel_test.go:40: BytesRead() = 292, want 340
    testing.go:1712: race detected during execution of test
FAIL
FAIL	seek/internal/search	0.014s
FAIL
```

(Abridged: 7 reports printed; `...` stands for 6. Paths shortened; addresses, goroutine numbers,
and the wrong count vary.)

Two lessons:

- **Without `-race`, the test passed.** The racy code *looked* correct. That's why the detector
  exists.
- **With `-race`, the count was wrong too** (292, not 340): lost updates, like `count++` earlier.

Read the report bottom-up: two workers, both started by `wg.Go` in `SearchAll`, both in
`countingReader.Read` line 57 (`*c.n += int64(n)`), same address. Now undo the four edits.

### Checkpoint

```console
$ go test -race ./...
?   	seek/cmd/seek	[no test files]
?   	seek/internal/output	[no test files]
ok  	seek/internal/search	1.013s
?   	seek/internal/stats	[no test files]
$ go run -race ./cmd/seek --stats -i fox testdata
testdata/notes.md:- feed the FOX
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/poem.txt:Foxes are quick and clever.
testdata/src/hello.go:	println("hello, fox")
files searched: 3
bytes read:     223
files matched:  3
total matches:  5
elapsed:        595µs
top files:
      3  testdata/poem.txt
      1  testdata/notes.md
      1  testdata/src/hello.go
$ go run ./cmd/seek --stats -c --hidden '' testdata
testdata/.cache/fox.txt:1
testdata/.hidden.txt:1
testdata/notes.md:4
testdata/poem.txt:4
testdata/src/hello.go:5
files searched: 5
bytes read:     249
files matched:  5
total matches:  15
elapsed:        270µs
top files:
      5  testdata/src/hello.go
      4  testdata/notes.md
      4  testdata/poem.txt
      1  testdata/.cache/fox.txt
      1  testdata/.hidden.txt
```

`elapsed:` and test durations will differ (a repeat run may show `(cached)`).

`bytes read` counts every byte read, including binary files that are then skipped:
223 = 103 (`poem.txt`) + 55 (`notes.md`) + 53 (`hello.go`) + 12 (`image.bin`). The second run adds
the two 13-byte hidden files: 249.

`go run -race` builds `seek` with the detector and runs it; any race would print to stderr. From
now on, test with `-race` by habit.

The complete project at this point is in `project/ch17/`.

## Summary

- Data race = unsynchronized accesses to one variable, at least one a write. Undefined behavior; a
  bug even if output looks right.
- `-race` (`go test`/`run`/`build`) reports races that happen during the run, with both stacks.
  Exit status 66.
- `sync.WaitGroup` waits for goroutines; `wg.Go(f)` (Go 1.25) replaces `Add`/`go`/`Done`.
- `sync.Mutex` guards state; `RWMutex` allows many readers. `defer Unlock`, keep sections short,
  never copy a lock (`go vet` checks).
- `sync.Once`, `OnceFunc`, `OnceValue`, `OnceValues`: run setup exactly once.
- `sync/atomic` types (`Int64`, `Bool`, `Pointer[T]`) protect single values without locks.
  `sync.Pool` reuses temporary objects.
- Channels pass ownership and signal; mutexes guard state. `errgroup` (`golang.org/x/sync`) returns
  the first error from a group of tasks.
- Memory model: a read sees a write only if the write *happens before* it; only synchronization
  (channels, locks, WaitGroup, Once, atomics) creates those edges.
- `seek` counts bytes in a shared `atomic.Int64`, starts workers with `wg.Go`, and tests under
  `-race`.

## Exercises

1. **Fix the race three ways.** Make `examples/ch17/race` print 400000 every time: with a
   `sync.Mutex`, with an `atomic.Int64`, and with no shared variable at all (each goroutine counts
   locally and sends its total on a channel). Check each with `go run -race`.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func withMutex() int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	count := 0
	for range 4 {
		wg.Go(func() {
			for range 100_000 {
				mu.Lock()
				count++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	return count
}

func withAtomic() int64 {
	var count atomic.Int64
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100_000 {
				count.Add(1)
			}
		})
	}
	wg.Wait()
	return count.Load()
}

func withChannel() int {
	totals := make(chan int)
	for range 4 {
		go func() {
			n := 0 // local to this goroutine: no sharing at all
			for range 100_000 {
				n++
			}
			totals <- n
		}()
	}
	count := 0
	for range 4 {
		count += <-totals
	}
	return count
}

func main() {
	fmt.Println(withMutex(), withAtomic(), withChannel())
}
```

```console
$ go run -race .
400000 400000 400000
```

In the channel version the goroutines share nothing until the end, so there's nothing to lock.

</details>

2. **A generic safe counter.** Using generics ([Chapter 13](13-generics.md)), write
   `SafeCounter[K comparable]` with `Add(key K, n int)` and `Get(key K) int`, safe for concurrent
   use, with a ready-to-use zero value.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"sync"
)

// SafeCounter counts keys of any comparable type. It is safe for
// concurrent use; the zero value is ready to use.
type SafeCounter[K comparable] struct {
	mu     sync.Mutex
	counts map[K]int
}

func (c *SafeCounter[K]) Add(key K, n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[K]int)
	}
	c.counts[key] += n
}

func (c *SafeCounter[K]) Get(key K) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

func main() {
	var c SafeCounter[int]
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() { c.Add(i%3, 1) })
	}
	wg.Wait()
	fmt.Println(c.Get(0), c.Get(1), c.Get(2)) // 34 33 33
}
```

</details>

3. **Lazy regexp.** A package-level `var re = regexp.MustCompile(...)` costs time at every program
   start, even if never used. Use `sync.OnceValue` so it compiles on first use, safely from many
   goroutines.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"regexp"
	"sync"
)

// dateRE is compiled on first use, not at program start. OnceValue
// makes that safe even if many goroutines call it at once.
var dateRE = sync.OnceValue(func() *regexp.Regexp {
	fmt.Println("compiling")
	return regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
})

func main() {
	var wg sync.WaitGroup
	for _, s := range []string{"on 2026-10-27", "no date", "2025-01-01!"} {
		wg.Go(func() {
			_ = dateRE().MatchString(s)
		})
	}
	wg.Wait()
	fmt.Println(dateRE().FindString("due 2026-12-31"))
}
```

```console
$ go run -race .
compiling
2026-12-31
```

"compiling" prints once, though there are four calls, three at the same time.

</details>

4. **Why a pointer?** `countingReader` stores `n *atomic.Int64`. What goes wrong if it stored
   `n atomic.Int64` and `Search` built it with `countingReader{r: r, n: s.bytesRead}`? Try it and
   run `go vet ./...`.

<details><summary>Solution</summary>

Each `countingReader` gets its own *copy* of the counter. Reads add to the copy, which is thrown
away when `Search` returns, so `BytesRead()` stays 0. `go vet` catches both copies:

```text
internal/search/search.go:56:9: Read passes lock by value: seek/internal/search.countingReader contains sync/atomic.Int64 contains sync/atomic.noCopy
internal/search/search.go:85:30: literal copies lock value from s.bytesRead: sync/atomic.Int64 contains sync/atomic.noCopy
```

Sharing state means sharing one variable, through a pointer.

</details>

5. **Name the edge.** For each case, name the happens-before edge that makes it safe:
   (a) `main` reads `lengths` after `wg.Wait()` in the waitgroup example;
   (b) `seek`'s consumer reads `res.Matches`, which a worker built;
   (c) a worker reads `j.path`, which the producer wrote.

<details><summary>Solution</summary>

(a) Each goroutine's write happens before its `Done` (the end of its `wg.Go` function), which
happens before `Wait` returns.
(b) The worker builds the slice, then sends the `Result` on `j.result`. A send happens before the
matching receive (`<-result` in the consumer) completes.
(c) The producer creates the `job`, then sends it on `jobs`; the worker's receive in
`for j := range jobs` completes after that send.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-17--synchronization-and-the-memory-model)
