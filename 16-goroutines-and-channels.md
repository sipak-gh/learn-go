# Chapter 16 — Goroutines and Channels

Concurrency is where Go stands out. A *goroutine* is a function running at the same time as others.
A *channel* is a typed pipe that goroutines use to pass values and wait for each other. You'll learn
both, plus `select` and the common patterns built on them. Then `seek` searches many files at once,
with output still in order. Builds on [Chapter 14](14-iterators.md)'s iterators;
[Chapter 17](17-sync.md) adds the `sync` package and the race detector.

## Goroutines

`go f()` runs the call in a new goroutine. The `go` statement returns at once; it doesn't wait.

`examples/ch16/goroutines/main.go`

```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func say(word string) {
	for i := range 3 {
		fmt.Println(word, i)
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	fmt.Println("CPUs:", runtime.NumCPU(), "GOMAXPROCS:", runtime.GOMAXPROCS(0))

	go say("goroutine") // runs at the same time as main
	say("main")

	fmt.Println("goroutines now:", runtime.NumGoroutine())

	go fmt.Println("this may never print")
	// main returns here. The program ends at once, and any goroutines
	// still running are stopped mid-flight.
}
```

(`time.Sleep` pauses the goroutine; [Chapter 18](18-context-and-time.md) covers `time`.)

```console
$ go run ./ch16/goroutines
CPUs: 32 GOMAXPROCS: 32
main 0
goroutine 0
main 1
goroutine 1
goroutine 2
main 2
goroutines now: 2
```

Line order, CPU count, and the final count (1 or 2) vary by run and machine.

- **Goroutines are cheap.** Each starts with a few KiB of stack that grows as needed. You can run
  hundreds of thousands; threads cost far more.
- **M:N scheduling.** The runtime runs many goroutines (M) on a few OS threads (N). When one blocks
  on a channel, lock, or I/O, the runtime parks it and runs another. You write plain blocking code;
  the runtime makes it efficient.
- **`GOMAXPROCS`** = how many threads run Go code at once. Default: the number of available CPUs.
  `runtime.GOMAXPROCS(0)` reads it without changing it. You rarely set it.
- **When `main` returns, the program ends.** Other goroutines stop mid-flight, so the last line
  above usually never prints. Make `main` wait: with channels, or `sync.WaitGroup`
  ([Chapter 17](17-sync.md)).
- **A goroutine has no handle.** You can't kill it, join it, or fetch its result. It ends when its
  function returns. To get results or stop it, communicate with it.
- **A panic in any goroutine crashes the whole program**, unless that goroutine recovers.

### Concurrency is not parallelism

*Concurrency* = structuring a program as independent tasks. *Parallelism* = running tasks at the
same instant on several CPUs. Goroutines give you concurrency; the runtime decides on parallelism.
A concurrent design is still clearer and useful on one CPU, e.g. while one goroutine waits for a
disk.

## Channels

A channel carries values of one type between goroutines.

| Operation | Syntax | Notes |
|-----------|--------|-------|
| make | `ch := make(chan T)` | unbuffered |
| make buffered | `ch := make(chan T, n)` | holds up to `n` values |
| send | `ch <- v` | |
| receive | `v := <-ch` | or just `<-ch` to discard |
| receive with status | `v, ok := <-ch` | `ok` is false if closed and empty |
| close | `close(ch)` | no more sends |
| length / capacity | `len(ch)`, `cap(ch)` | values waiting / buffer size |

A channel's zero value is `nil`; always create channels with `make`.

### Unbuffered channels synchronize

An unbuffered channel has no storage. A send blocks until another goroutine receives, and vice
versa. Both sides meet, so the channel passes a value *and* synchronizes two goroutines.

`examples/ch16/channels/main.go`

```go
package main

import "fmt"

func main() {
	ch := make(chan string) // unbuffered: a send waits for a receiver

	go func() {
		ch <- "ping" // blocks until main receives
	}()
	msg := <-ch // blocks until the goroutine sends
	fmt.Println("main got", msg)

	// The other direction. done lets main wait until the worker has
	// printed; without it, main could return first and end the program.
	done := make(chan bool)
	go func() {
		v := <-ch
		fmt.Println("worker got", v)
		done <- true
	}()
	ch <- "pong"
	<-done // the value doesn't matter; receiving is the signal
	fmt.Println("main done")
}
```

```console
$ go run ./ch16/channels
main got ping
worker got pong
main done
```

For a pure signal, use `chan struct{}`: `struct{}` takes zero bytes and tells readers "no data".

### Buffered channels

A buffered channel holds up to `cap` values. A send blocks only when the buffer is full; a receive
only when it's empty. Values come out in FIFO order.

`examples/ch16/buffered/main.go`

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 3) // room for 3 values

	// No receiver yet, but these sends don't block: the buffer has room.
	ch <- 1
	ch <- 2
	fmt.Println("len", len(ch), "cap", cap(ch))

	fmt.Println(<-ch) // values come out in the order they went in
	ch <- 3
	ch <- 4
	fmt.Println("len", len(ch), "cap", cap(ch))
	fmt.Println(<-ch, <-ch, <-ch)
}
```

```console
$ go run ./ch16/buffered
len 2 cap 3
1
len 3 cap 3
2 3 4
```

Add a buffer only for a clear purpose: to let a sender move on, or to limit work in flight. A
buffer that "fixes" a deadlock usually hides a bug. Don't act on `len(ch)`: another goroutine can
change it right after you read it.

### Close, range, and comma-ok

`close(ch)` means "no more values". Receivers still drain the buffer; after that, every receive
returns the zero value at once. `for v := range ch` receives until the channel is closed and empty.

`examples/ch16/closing/main.go`

```go
package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i * i
		}
		close(ch) // no more values: ends the receiver's range loop
	}()

	// range receives until the channel is closed and empty.
	for v := range ch {
		fmt.Println("got", v)
	}

	// Receiving from a closed channel never blocks. It returns the zero
	// value; the second result, ok, is false to tell you why.
	v, ok := <-ch
	fmt.Println(v, ok)

	buf := make(chan string, 2)
	buf <- "a"
	close(buf) // values already in the buffer can still be received
	s, ok := <-buf
	fmt.Printf("%q %v\n", s, ok)
	s, ok = <-buf
	fmt.Printf("%q %v\n", s, ok)
}
```

```console
$ go run ./ch16/closing
got 1
got 4
got 9
0 false
"a" true
"" false
```

Rules for closing:

- **Only the sender closes**, and only when receivers need to know the stream ended (e.g. to end a
  `range`). Closing isn't needed to free a channel; the garbage collector does that.
- **Send on a closed channel panics**: `panic: send on closed channel`.
- **Close twice panics**: `panic: close of closed channel`. So does closing a nil channel.
- With several senders, none can safely close. A separate goroutine closes after all senders
  finish (see fan-in below).

### Directional channel types

`chan<- T` is send-only; `<-chan T` is receive-only. A plain `chan T` converts to either
automatically. The compiler enforces the direction, and the signature documents it.

`examples/ch16/directions/main.go`

```go
package main

import "fmt"

// produce can only send on out.
func produce(out chan<- string) {
	for _, w := range []string{"red", "green", "blue"} {
		out <- w
	}
	close(out) // closing is a send-side operation
}

// consume can only receive from in.
func consume(in <-chan string) {
	for w := range in {
		fmt.Println(w)
	}
}

func main() {
	ch := make(chan string) // bidirectional
	go produce(ch)          // converted to chan<- string automatically
	consume(ch)             // converted to <-chan string automatically
}
```

```console
$ go run ./ch16/directions
red
green
blue
```

Sending on a receive-only channel is a compile error:

```go
func consume(in <-chan string) {
	in <- "oops"
}
```

```text
./main.go:4:2: invalid operation: cannot send to receive-only channel <-chan string in (variable of type <-chan string)
```

Read the arrow as the data's direction: `chan<-` data goes *in*; `<-chan` data comes *out*. You
can't close a receive-only channel either.

## Deadlock

If every goroutine is blocked forever on a channel (or lock), the runtime detects it and crashes:

```go
package main

import "fmt"

func main() {
	ch := make(chan int)
	ch <- 1 // nobody will ever receive
	fmt.Println(<-ch)
}
```

```console
$ go run .
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/home/you/deadlock/main.go:7 +0x36
exit status 2
```

The trace shows what each goroutine waits on (`chan send`, `chan receive`, `select`, ...). The check
fires only when *all* goroutines are stuck. If any other goroutine is alive (e.g. waiting on the
network), a stuck goroutine just stays stuck: a *leak*.

## `select`

`select` waits on several channel operations and runs the first that can proceed. It looks like a
`switch`, but each `case` is a send or receive.

- One case ready: it runs.
- Several ready: one is picked **at random**, so a busy channel can't starve the others.
- None ready: `select` blocks, unless there's a `default`, which then runs at once.

`examples/ch16/selecting/main.go`

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 1)

	// With a default case, select never blocks: if no other case is
	// ready, default runs.
	select {
	case v := <-ch:
		fmt.Println("received", v)
	default:
		fmt.Println("nothing to receive")
	}

	ch <- 1 // fills the buffer
	select {
	case ch <- 2:
		fmt.Println("sent 2")
	default:
		fmt.Println("buffer full, 2 dropped")
	}

	// When several cases are ready, select picks one at random.
	a := make(chan string, 1)
	b := make(chan string, 1)
	counts := map[string]int{}
	for range 1000 {
		a <- "a"
		b <- "b"
		select {
		case v := <-a:
			counts[v]++
			<-b // empty the other channel for the next round
		case v := <-b:
			counts[v]++
			<-a
		}
	}
	fmt.Println(counts)
}
```

```console
$ go run ./ch16/selecting
nothing to receive
buffer full, 2 dropped
map[a:498 b:502]
```

The last numbers vary but stay near 500 each.

An empty `select {}` blocks forever. A `select` in a `for` loop is the usual way to serve several
channels. [Chapter 18](18-context-and-time.md) adds timeouts (`time.After`) and cancellation
(`ctx.Done()`).

### Nil channels switch off a case

Sending or receiving on a nil channel blocks forever, so in a `select` that case is never chosen.
Set a channel variable to `nil` to switch its case off.

`examples/ch16/nilchannels/main.go`

```go
package main

import "fmt"

func send(values ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, v := range values {
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func main() {
	a, b := send(1, 2, 3), send(10, 20)
	sum := 0
	// Loop until both channels are closed. A receive from a nil channel
	// blocks forever, so setting a finished channel to nil switches its
	// case off. (A closed channel would be ready all the time instead.)
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			sum += v
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			sum += v
		}
	}
	fmt.Println("sum:", sum)
}
```

```console
$ go run ./ch16/nilchannels
sum: 36
```

## Goroutine leaks

A goroutine blocked forever is a *leak*. It keeps its stack and everything it references until the
program ends; the garbage collector can't free it because it might still wake up. The usual cause:
a sender whose receiver went away.

The fix: give every goroutine a way to stop, such as a *done channel* the owner closes. A closed
channel is always ready to receive from, so closing `done` wakes every goroutine selecting on it.

`examples/ch16/leak/main.go`

```go
package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var words = []string{"alpha", "beta", "gamma"}

// leakyFirst returns the first word in upper case. It leaks: nobody
// receives the second word, so the goroutine waits on its send forever.
func leakyFirst() string {
	ch := make(chan string)
	go func() {
		for _, w := range words {
			ch <- strings.ToUpper(w)
		}
		close(ch)
	}()
	return <-ch
}

// fixedFirst closes done when it returns. The goroutine selects on done
// as well as on its send, so it can always stop.
func fixedFirst() string {
	ch := make(chan string)
	done := make(chan struct{})
	defer close(done)
	go func() {
		defer close(ch)
		for _, w := range words {
			select {
			case ch <- strings.ToUpper(w):
			case <-done:
				return
			}
		}
	}()
	return <-ch
}

func main() {
	fmt.Println("goroutines at start:", runtime.NumGoroutine())
	for range 3 {
		leakyFirst()
	}
	time.Sleep(10 * time.Millisecond) // give goroutines time to finish
	fmt.Println("after leakyFirst x3:", runtime.NumGoroutine())
	for range 3 {
		fixedFirst()
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println("after fixedFirst x3:", runtime.NumGoroutine())
}
```

```console
$ go run ./ch16/leak
goroutines at start: 1
after leakyFirst x3: 4
after fixedFirst x3: 4
```

Three `leakyFirst` calls left three stuck goroutines; `fixedFirst` left none. Rule: **when you
start a goroutine, know how it will end.**

## Concurrency patterns

Go's motto:

> Don't communicate by sharing memory; share memory by communicating.

Instead of goroutines sharing a variable (and needing locks), pass the data over a channel; whoever
holds the value owns it. Every pattern below follows this idea.

### Generator

A function starts a goroutine that sends a sequence, and returns the channel.

`examples/ch16/generator/main.go`

```go
package main

import "fmt"

// count returns a channel that delivers from, from+1, ..., to and is
// then closed. The goroutine inside is the "generator".
func count(from, to int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := from; i <= to; i++ {
			ch <- i
		}
	}()
	return ch
}

func main() {
	for n := range count(1, 5) {
		fmt.Print(n, " ")
	}
	fmt.Println()
}
```

```console
$ go run ./ch16/generator
1 2 3 4 5 
```

For a plain sequence, prefer an iterator (`iter.Seq`, [Chapter 14](14-iterators.md)): no goroutine,
much cheaper per value, and no leak on early `break`. `count` above leaks if its caller stops
reading. Use channel generators only when producing should really run concurrently.

### Pipeline

A chain of stages. Each stage is a goroutine that receives from one channel, does one job, and
sends on the next. Closing flows down: when a stage's input closes, it closes its output.

`examples/ch16/pipeline/main.go`

```go
package main

import (
	"fmt"
	"strings"
)

// Each stage receives from the previous stage's channel, does one job,
// and sends on its own channel. Closing flows down the pipeline.

func emit(lines ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, l := range lines {
			out <- l
		}
	}()
	return out
}

func grep(in <-chan string, word string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for l := range in {
			if strings.Contains(l, word) {
				out <- l
			}
		}
	}()
	return out
}

func upper(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for l := range in {
			out <- strings.ToUpper(l)
		}
	}()
	return out
}

func main() {
	lines := emit("the quick fox", "a lazy dog", "fox runs")
	for l := range upper(grep(lines, "fox")) {
		fmt.Println(l)
	}
}
```

```console
$ go run ./ch16/pipeline
THE QUICK FOX
FOX RUNS
```

### Fan-out and fan-in

- *Fan-out*: several goroutines receive from one channel. Each value goes to exactly one, so they
  share the work.
- *Fan-in*: several goroutines send on one channel. Someone must close it after **all** senders
  finish.

`examples/ch16/fanin/main.go`

```go
package main

import (
	"fmt"
	"slices"
)

// square reads numbers from in and sends their squares on out. Several
// copies of it read the same in channel: that is fan-out.
func square(id int, in <-chan int, out chan<- string, done chan<- bool) {
	for n := range in {
		out <- fmt.Sprintf("%3d (worker %d)", n*n, id)
	}
	done <- true
}

func main() {
	in := make(chan int)
	out := make(chan string)
	done := make(chan bool)

	const workers = 3
	for id := range workers {
		go square(id, in, out, done)
	}
	go func() {
		for n := 1; n <= 6; n++ {
			in <- n
		}
		close(in)
	}()
	// Fan-in: all workers send on one channel. Close it after every
	// worker has said it's done. (Chapter 17's sync.WaitGroup does this
	// counting for you.)
	go func() {
		for range workers {
			<-done
		}
		close(out)
	}()

	var results []string
	for r := range out {
		results = append(results, r[:3]) // drop the worker id: it varies
	}
	slices.Sort(results)
	fmt.Println(results)
}
```

```console
$ go run ./ch16/fanin
[  1   4   9  16  25  36]
```

Results arrive in finishing order, so the example sorts them.

### Worker pool

Fan-out with a fixed number of long-lived workers. It caps how much runs at once (say, open files)
however many jobs there are. To keep input order, give each job an index and store each result in
its slot.

`examples/ch16/workerpool/main.go`

```go
package main

import (
	"fmt"
	"time"
)

type job struct {
	index int
	name  string
}

type result struct {
	index  int
	length int
}

func worker(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		time.Sleep(time.Duration(len(j.name)) * time.Millisecond) // pretend to work
		results <- result{j.index, len(j.name)}
	}
}

func main() {
	names := []string{"seek", "goroutine", "channel", "select", "go"}
	jobs := make(chan job)
	results := make(chan result)

	for range 3 { // a fixed pool of 3 workers
		go worker(jobs, results)
	}
	go func() {
		for i, n := range names {
			jobs <- job{i, n}
		}
		close(jobs)
	}()

	// Results arrive in any order. Each carries its index, so we can put
	// it in the right place.
	lengths := make([]int, len(names))
	for range names {
		r := <-results
		lengths[r.index] = r.length
	}
	for i, n := range names {
		fmt.Printf("%-10s %d\n", n, lengths[i])
	}
}
```

```console
$ go run ./ch16/workerpool
seek       4
goroutine  9
channel    7
select     6
go         2
```

This waits for *all* results before printing. `seek` should print the first file's results as soon
as they're ready; the project section shows how.

### Bounded parallelism with a semaphore

One goroutine per task, but at most N at once: a buffered channel of capacity N is a *semaphore*.
Send to take a slot, receive to give it back.

`examples/ch16/semaphore/main.go`

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	// A buffered channel with capacity 2 works as a semaphore: a send
	// takes a slot, a receive gives it back, and a third send blocks
	// until a slot is free.
	sem := make(chan struct{}, 2)
	inFlight := make(chan int)

	for i := range 6 {
		go func() {
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release
			inFlight <- len(sem)     // how many slots are taken right now
			time.Sleep(time.Duration(10+i) * time.Millisecond)
		}()
	}

	most := 0
	for range 6 {
		most = max(most, <-inFlight)
	}
	fmt.Println("most at once:", most)
}
```

```console
$ go run ./ch16/semaphore
most at once: 2
```

Pool or semaphore? A pool reuses N goroutines; a semaphore starts one per task and makes the extras
wait. Both bound the work. Pick whichever reads more simply.

## Building seek: a concurrent worker pool

`seek` searches one file at a time. Reading is mostly waiting for the disk, and matching uses one
CPU. Searching several files at once uses both better.

The hard part is output. Files finish in any order, but `seek` must print results grouped per file
and **in walk order**, exactly as before, for any number of workers. It should print each file as
soon as possible, without holding every result in memory.

The design:

```
Walk ──► producer ──jobs──► worker 1 ─┐
             │              worker 2 ─┼─► each file's own result channel
             │              worker N ─┘           │
             └──pending (result channels, in order)──► consumer ──► yield
```

1. A **producer** goroutine ranges over `Walk`'s files. For each file it makes a result channel
   (buffer 1), queues it on `pending`, then sends a job (path + result channel) on `jobs`.
2. **N workers** receive jobs, search the file, and send the `Result` on the job's channel.
3. The **consumer** (the iterator's loop) takes result channels off `pending` *in order* and
   receives from each. If file 3 finishes before file 1, its result waits in its buffer.

Why:

- `pending` is a **channel of channels** (`chan chan Result`). Its buffer (= workers) limits how far
  the producer runs ahead, so memory stays bounded.
- Result channels have a **buffer of 1**, so a worker can always hand off and move on, even while
  the consumer waits for an older file.
- A **done channel** stops the producer if the caller breaks early. `defer close(done)` runs however
  the iterator returns. Workers finish the queued jobs and exit when `jobs` closes. Nothing leaks.
- A walk error (missing path) goes straight into that file's result channel, so it appears in order
  too.
- Workers share one `Searcher`. That's safe because they only *read* its fields.
  ([Chapter 17](17-sync.md) adds a field workers write, and shows what that requires.)

### The new file: `internal/search/parallel.go`

Create `internal/search/parallel.go`:

```go
package search

import "iter"

// A Result is the outcome of searching one file.
type Result struct {
	Path    string
	Matches []Match
	Err     error // ErrBinary for a skipped binary file
}

// A job asks a worker to search one file and send the Result on result.
type job struct {
	path   string
	result chan<- Result // send-only: workers never receive from it
}

// SearchAll searches files using up to workers goroutines at once and
// returns an iterator over the results. Even though files finish in any
// order, results are yielded in the order files produced them. An error
// from files is passed through in the Result for its path.
//
// The design: a producer goroutine reads files and gives each one a
// buffered channel for its result. It hands the file to the workers
// through the jobs channel and also queues the result channel on
// pending, in order. The consumer (the loop below) receives from the
// queued channels one by one, so it waits for the oldest file even if
// newer ones are already done.
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

		// Producer.
		go func() {
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
		}()

		// Workers.
		for range workers {
			go func() {
				for j := range jobs {
					matches, err := s.SearchFile(j.path)
					j.result <- Result{Path: j.path, Matches: matches, Err: err}
				}
			}()
		}

		// Consumer: results in order.
		for result := range pending {
			if !yield(<-result) {
				return
			}
		}
	}
}
```

Notice:

- `SearchAll` returns an `iter.Seq[Result]`, so callers use a plain `for range`. The goroutines are
  hidden, and start when the loop starts, not when `SearchAll` is called.
- The producer's two `defer close(...)` run in reverse: `jobs` closes first, then `pending`.
- Both producer sends `select` on `<-done` too. Otherwise a producer blocked on a full `pending`
  would leak after an early `break`.
- `for result := range pending` ends when the producer closes `pending`. Each `<-result` waits for
  that one file.

### Using it in `cmd/seek/main.go`

In `run`, replace the old `for path, err := range search.Walk(...)` loop with this. Everything
above (`walkOpts`, the `var` block) and below (`if opts.stats`) stays:

```go
	files := search.Walk(opts.paths, walkOpts)
	for res := range s.SearchAll(files, opts.workers) {
		err := res.Err
		if errors.Is(err, search.ErrBinary) {
			continue // binary files are skipped silently
		}
		if err == nil {
			err = p.PrintFile(res.Path, res.Matches)
		}
		if err != nil {
			report(stderr, err)
			hadError = true
			continue
		}
		searched++
		if len(res.Matches) > 0 {
			matched.Add(res.Path, len(res.Matches))
		}
	}
```

The body barely changed: it reads `res.Path`, `res.Matches`, and `res.Err` instead of calling
`SearchFile`. Printing still happens only here, in one goroutine, so files' output can't interleave.

### The `-j` flag in `cmd/seek/flags.go`

Four small edits.

1. Add `"runtime"` to the imports:

```go
import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)
```

2. In the `usage` text, add a line after `--glob`:

```text
      --glob GLOB           only search files whose name matches GLOB, e.g. '*.go'
  -j N                      search N files at a time (default: number of CPUs)
      --color WHEN          highlight matches: auto, always, or never
```

3. Add a field to `options`, after `glob`:

```go
	glob        string    // --glob PATTERN
	workers     int       // -j N
	json        bool      // --json
```

4. In `parseArgs`, register the flag after `--glob` (default: CPU count), and validate it after the
   `-m` check:

```go
	flags.StringVar(&opts.glob, "glob", "", "")
	flags.IntVar(&opts.workers, "j", runtime.NumCPU(), "")
	flags.Var(&opts.color, "color", "")
```

```go
	if opts.maxCount < 0 {
		return opts, usageErrorf("invalid value for -m: %d is negative", opts.maxCount)
	}
	if opts.workers < 1 {
		return opts, usageErrorf("invalid value for -j: must be at least 1")
	}
```

With `-j 0` there would be no workers, and the consumer would wait forever. The check makes it a
usage error.

### Checkpoint

From the project root:

```console
$ go run ./cmd/seek -n -i fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek -j 1 -n -i fox testdata
testdata/notes.md:2:- feed the FOX
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/poem.txt:4:Foxes are quick and clever.
testdata/src/hello.go:4:	println("hello, fox")
$ go run ./cmd/seek -j 0 fox testdata
seek: invalid value for -j: must be at least 1
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit status 2
$ go run ./cmd/seek -c --hidden -j 8 o testdata
testdata/.cache/fox.txt:1
testdata/.hidden.txt:1
testdata/notes.md:3
testdata/poem.txt:4
testdata/src/hello.go:1
$ go run ./cmd/seek fox testdata nope.txt
seek: nope.txt: no such file or directory
testdata/poem.txt:The quick brown fox
testdata/poem.txt:The dog sleeps; the fox runs.
testdata/src/hello.go:	println("hello, fox")
exit status 2
```

Output matches [Chapter 15](15-json-and-encoding.md) for any `-j`. Rerun the `-j 8` command: the
order never changes. (In the last command the error comes first because stderr is unbuffered and
stdout buffered; see [Chapter 11](11-io-and-files.md).)

The tiny fixtures show no speed-up. Try a large tree, e.g. the Go source (`go env GOROOT`), with
`-j 1` and with the default, and compare times.

The complete project at this point is in `project/ch16/`.

## Summary

- `go f()` runs `f` in a new, cheap goroutine, scheduled onto `GOMAXPROCS` threads. When `main`
  returns, all goroutines stop.
- Unbuffered channels make sender and receiver meet; buffered ones let the sender run `cap` values
  ahead.
- Only the sender closes. `range ch` stops at close; `v, ok := <-ch` reports it. Sending on or
  closing a closed channel panics.
- `chan<- T` / `<-chan T` restrict direction, checked by the compiler.
- `select` waits on several operations, picking randomly among ready ones. `default` makes it
  non-blocking; a nil channel's case is never chosen.
- All goroutines blocked → `all goroutines are asleep - deadlock!`. One blocked goroutine is a
  silent leak; give every goroutine a way to stop (a done channel).
- Patterns: generator, pipeline, fan-out/fan-in, worker pool, semaphore. Share memory by
  communicating.
- `seek` searches with a worker pool (`-j N`); a queue of per-file result channels keeps walk order.

## Exercises

1. **Ping-pong.** Two goroutines pass a counter back and forth over two unbuffered channels, each
   adding 1, until it reaches 10. `main` prints the final value.

<details><summary>Solution</summary>

```go
package main

import "fmt"

func player(in <-chan int, out chan<- int, done chan<- int) {
	for n := range in {
		if n >= 10 {
			done <- n
			return
		}
		out <- n + 1
	}
}

func main() {
	ping, pong := make(chan int), make(chan int)
	done := make(chan int)
	go player(ping, pong, done)
	go player(pong, ping, done)
	ping <- 0
	fmt.Println(<-done) // 10
}
```

The player that sees 10 reports it on `done` and returns. The other stays blocked on its receive
until `main` returns. In a long-running program, close a channel to stop it too.

</details>

2. **Ordered results without a big slice.** Rewrite the worker-pool example to print each name as
   soon as it and all earlier names are done, using `SearchAll`'s "queue of result channels" design.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	names := []string{"seek", "goroutine", "channel", "select", "go"}
	const workers = 3

	type job struct {
		name   string
		result chan<- int
	}
	jobs := make(chan job)
	pending := make(chan chan int, workers)

	go func() {
		defer close(pending)
		defer close(jobs)
		for _, n := range names {
			r := make(chan int, 1)
			pending <- r
			jobs <- job{n, r}
		}
	}()
	for range workers {
		go func() {
			for j := range jobs {
				time.Sleep(time.Duration(len(j.name)) * time.Millisecond)
				j.result <- len(j.name)
			}
		}()
	}

	i := 0
	for r := range pending {
		fmt.Printf("%-10s %d\n", names[i], <-r)
		i++
	}
}
```

Same output as the original. No `done` channel is needed: the consumer never stops early.

</details>

3. **Merge.** Write `func merge(chans ...<-chan int) <-chan int` that fans in any number of channels
   and closes its result after all inputs close. Use one goroutine per input and a counting
   goroutine (no `sync` package yet).

<details><summary>Solution</summary>

```go
package main

import "fmt"

func merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	finished := make(chan struct{})
	for _, c := range chans {
		go func() {
			for v := range c {
				out <- v
			}
			finished <- struct{}{}
		}()
	}
	go func() {
		for range chans {
			<-finished
		}
		close(out)
	}()
	return out
}

func gen(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func main() {
	sum := 0
	for v := range merge(gen(1, 2), gen(3), gen(4, 5, 6)) {
		sum += v
	}
	fmt.Println(sum) // 21
}
```

Each loop iteration has its own `c` (Go 1.22+), so each goroutine reads a different channel.

</details>

4. **Early break.** In a scratch copy of `seek`, write a small program (or temporary code in `run`)
   that ranges over `s.SearchAll(files, 4)` and `break`s after the first result. Print
   `runtime.NumGoroutine()` right after the loop and again after `time.Sleep(50 *
   time.Millisecond)`. What do you see, and why does the count drop?

<details><summary>Solution</summary>

Right after `break` the count can still be above 1. The deferred `close(done)` has run, but the
other goroutines need a moment:

- The producer returns from its `select` on `<-done` and closes `jobs`.
- Workers finish their current job (the result send never blocks: buffer of 1) and leave
  `range jobs`.

After the sleep the count is 1 (just `main`). Remove the producer's `<-done` cases and the count
stays high: the producer leaks, blocked on `pending` or `jobs`.

</details>

5. **Semaphore instead of a pool.** Sketch how `SearchAll` could use one goroutine per file plus a
   semaphore instead of N workers. What would you gain or lose?

<details><summary>Solution</summary>

Per file, the producer acquires a slot (`sem <- struct{}{}`, in a `select` with `<-done`), queues
the result channel on `pending`, then runs `go func() { defer func() { <-sem }(); ... }()` to
search.

- Ordering: unchanged; `pending` still holds result channels in order.
- Gain: slightly simpler code (no `jobs` channel, no worker loop).
- Cost: one goroutine start per file, cheap but not free.

Both cap concurrency at N. The pool is the more common shape when every task is the same.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-16--goroutines-and-channels)
