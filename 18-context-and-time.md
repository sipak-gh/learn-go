# Chapter 18 — Context, Cancellation, Time, and Signals

[Chapter 16](16-goroutines-and-channels.md) stopped goroutines by closing a `done` channel. That
works inside one function. Real programs must stop work spread across many functions and packages:
"the user pressed Ctrl-C", "this request took too long". The standard carrier for that message is
`context.Context`. This chapter covers contexts, the `time` package (durations, timers,
formatting), and `os/signal`. `seek` gets a `--timeout` flag and stops cleanly on Ctrl-C.
[Chapter 19](19-testing.md) then tests everything.

## Why context?

`seek` is searching a huge directory and the user presses Ctrl-C. Who needs to know?

- `run`, to print what it found so far.
- `SearchAll`, to stop its producer and workers.
- `Search`, deep in a worker, to stop reading a file halfway.

You could pass each of them a `done` channel, a deadline, and a reason. A `context.Context`
bundles exactly that into one value:

```go
type Context interface {
	Done() <-chan struct{}                   // closed when the work should stop
	Err() error                              // nil until Done is closed, then why
	Deadline() (deadline time.Time, ok bool) // when it will be canceled, if set
	Value(key any) any                       // request-scoped data (use sparingly)
}
```

You never implement this interface. Start from `context.Background()`, an empty context that is
never canceled, and derive new ones with the `With...` functions. Each derived context is a
*child*: canceling a parent cancels all its children, never the reverse. (`context.TODO()` is the
same as `Background`; it marks a spot where you haven't decided which context to pass yet.)

## `WithCancel`: stop on demand

`context.WithCancel(parent)` returns a child context and a `cancel` function. Calling `cancel`
closes the child's `Done` channel:

`examples/ch18/cancel/main.go`

```go
package main

import (
	"context"
	"fmt"
	"sync"
)

// count sends 1, 2, 3, ... on out until ctx is canceled.
func count(ctx context.Context, out chan<- int) {
	for i := 1; ; i++ {
		select {
		case out <- i:
		case <-ctx.Done():
			fmt.Println("count:", ctx.Err())
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // always release the context's resources

	nums := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { count(ctx, nums) })

	for range 3 {
		fmt.Println(<-nums)
	}
	cancel()  // tell count to stop
	wg.Wait() // and wait until it has
	fmt.Println("main:", ctx.Err())
}
```

```console
$ go run ./ch18/cancel
1
2
3
count: context canceled
main: context canceled
```

`<-ctx.Done()` in a `select` is Chapter 16's `<-done` pattern. Before `Done` is closed, `ctx.Err()`
returns `nil`; after, `context.Canceled` (printed `context canceled`). Calling `cancel` again does
nothing, so calling it from several places is safe.

### Rules for using contexts

The standard library follows these conventions; so should you.

- **Pass the context as the first parameter, named `ctx`:**
  `func Search(ctx context.Context, r io.Reader) error`. Don't store it in a struct; pass it to
  each call that needs it.
- **Never pass a `nil` context.** If you have nothing better, use `context.Background()`.
  `context.WithValue(nil, ...)` panics with `cannot create context from nil parent`.
- **Always call `cancel`**, usually `defer cancel()` right after creating the context. It frees the
  timer and the parent's reference to the child. Forgetting is a leak, and `go vet` catches it:

  ```console
  $ go vet .
  main.go:10:7: the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak
  ```

- **Cancellation is a request, not a command.** Canceling only closes a channel. The working code
  must check `ctx.Done()` or `ctx.Err()` and return. A function that never looks at its context
  can't be stopped.

## `WithTimeout` and `WithDeadline`: stop on time

`context.WithTimeout(parent, d)` cancels the child automatically after `d`. `ctx.Err()` then
returns `context.DeadlineExceeded`:

`examples/ch18/timeout/main.go`

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// slowSquare pretends to do work that takes d. It gives up early if ctx
// is canceled first.
func slowSquare(ctx context.Context, n int, d time.Duration) (int, error) {
	select {
	case <-time.After(d):
		return n * n, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	fmt.Println(slowSquare(ctx, 3, 10*time.Millisecond)) // in time

	n, err := slowSquare(ctx, 4, 100*time.Millisecond) // too slow
	fmt.Println(n, err)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}
```

```console
$ go run ./ch18/timeout
9 <nil>
0 context deadline exceeded
true
```

The 50 ms budget covers *everything* done with `ctx`: the first call used 10 ms, and the second ran
out after the remaining 40 ms. A deadline travels with the work, so a chain of calls can't take
longer than the caller allowed. (`time.After` is explained [below](#timers-and-tickers).)

`context.WithDeadline(parent, t)` takes a point in time instead of a duration.
`WithTimeout(ctx, d)` is just `WithDeadline(ctx, time.Now().Add(d))`.

`examples/ch18/deadline/main.go`

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	deadline := time.Now().Add(2 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	d, ok := ctx.Deadline()
	fmt.Println("has deadline:", ok, "same:", d.Equal(deadline))

	// A child can shorten its parent's deadline but never extend it.
	child, cancel2 := context.WithTimeout(ctx, time.Hour)
	defer cancel2()
	cd, _ := child.Deadline()
	fmt.Println("child deadline is the parent's:", cd.Equal(deadline))

	// Canceling the parent cancels every child.
	cancel()
	<-child.Done()
	fmt.Println("child:", child.Err())

	_, ok = context.Background().Deadline()
	fmt.Println("Background has a deadline:", ok)
}
```

```console
$ go run ./ch18/deadline
has deadline: true same: true
child deadline is the parent's: true
child: context canceled
Background has a deadline: false
```

Asking for an hour under a parent that ends in 2 seconds gives you 2 seconds. A callee never gets
more time than its caller granted.

## Causes: saying *why*

`ctx.Err()` only returns `context.Canceled` or `context.DeadlineExceeded`, often too vague for a
message. The `...Cause` variants attach your own error, which `context.Cause(ctx)` returns:

`examples/ch18/cause/main.go`

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errShutdown = errors.New("server shutting down")

func main() {
	// WithCancelCause: the cancel function takes the reason.
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errShutdown)
	fmt.Println("Err:  ", ctx.Err())          // always one of two generic errors
	fmt.Println("Cause:", context.Cause(ctx)) // the reason you gave

	// WithTimeoutCause: the cause is reported when time runs out.
	ctx2, cancel2 := context.WithTimeoutCause(context.Background(),
		10*time.Millisecond, errors.New("lookup took too long"))
	defer cancel2()
	<-ctx2.Done()
	fmt.Println("Err:  ", ctx2.Err())
	fmt.Println("Cause:", context.Cause(ctx2))

	// For a context that isn't canceled, Cause is nil.
	fmt.Println("Cause:", context.Cause(context.Background()))
}
```

```console
$ go run ./ch18/cause
Err:   context canceled
Cause: server shutting down
Err:   context deadline exceeded
Cause: lookup took too long
Cause: <nil>
```

- `ctx.Err()` stays generic, so `errors.Is(err, context.DeadlineExceeded)` checks keep working.
  `context.Cause` gives the human-readable reason.
- Canceled without a cause? `Cause` returns the same as `Err`.
- There's also `WithDeadlineCause`. `seek` will use `WithTimeoutCause` to print
  `timed out after 100ms`.

## `AfterFunc`: run something on cancellation

`context.AfterFunc(ctx, f)` runs `f` in its own goroutine once `ctx` is done. It returns a `stop`
function that unregisters `f` and reports whether it did so before `f` started:

`examples/ch18/afterfunc/main.go`

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	// AfterFunc runs f in its own goroutine once ctx is done.
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("cleaning up:", ctx.Err())
		close(done)
	})

	cancel()
	<-done
	// stop reports whether it stopped f from running. Too late here.
	fmt.Println("stopped:", stop())

	ctx2, cancel2 := context.WithCancel(context.Background())
	stop2 := context.AfterFunc(ctx2, func() { fmt.Println("never runs") })
	fmt.Println("stopped:", stop2()) // unregister f before ctx2 is done
	cancel2()
}
```

```console
$ go run ./ch18/afterfunc
cleaning up: context canceled
stopped: false
stopped: true
```

`AfterFunc` connects a context to something that doesn't understand contexts, e.g. closing a
network connection to unblock a `Read` that would otherwise wait forever. You won't need it often.

## `WithValue`, and why to avoid it

`context.WithValue(parent, key, value)` returns a child holding one key/value pair.
`ctx.Value(key)` looks it up, searching up through the parents:

`examples/ch18/values/main.go`

```go
package main

import (
	"context"
	"fmt"
)

// An unexported key type means no other package can use the same key,
// even by accident with the same name.
type ctxKey int

const requestIDKey ctxKey = 0

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string) // "" if missing
	return id
}

func handle(ctx context.Context) {
	fmt.Printf("[%s] handling\n", requestID(ctx))
}

func main() {
	ctx := withRequestID(context.Background(), "req-42")
	handle(ctx)
	handle(context.Background())
}
```

```console
$ go run ./ch18/values
[req-42] handling
[] handling
```

Caveats:

- **Values are untyped.** `Value` returns `any`, so you need a type assertion, and the compiler
  can't check the value is there. The dependency is invisible in function signatures.
- **Use a private key type**, as above. Keys are compared with `==`: two packages using the string
  `"id"` would overwrite each other. An unexported type can't collide.
- **Wrap access in functions** (`withRequestID`, `requestID`) so callers never touch the key.
- **Only for request-scoped data** that crosses API boundaries: a request ID, a trace ID, the
  authenticated user. Never for optional parameters or dependencies like a database or logger;
  pass those explicitly.

`seek` doesn't use `WithValue`. Most programs don't need it.

## Durations and measuring time

A `time.Duration` is an `int64` count of nanoseconds, with unit constants from `time.Nanosecond` to
`time.Hour`. Its `String` method prints a readable form; `time.ParseDuration` reads it back:

`examples/ch18/durations/main.go`

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	// A Duration is an int64 count of nanoseconds, with constants for
	// the common units.
	d := 1*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Println(d, int64(d))
	fmt.Println(d.Minutes(), d.Round(time.Hour), d.Truncate(time.Hour))
	fmt.Println(1500*time.Millisecond, 2*time.Microsecond, time.Duration(42))

	// ParseDuration accepts the same format String produces.
	p, err := time.ParseDuration("1m30s")
	fmt.Println(p, p.Seconds(), err)
	_, err = time.ParseDuration("soon")
	fmt.Println(err)

	// Multiplying by a variable needs a conversion: n is an int.
	n := 3
	fmt.Println(time.Duration(n) * time.Second)

	// Measuring elapsed time.
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	elapsed := time.Since(start) // same as time.Now().Sub(start)
	fmt.Println(elapsed >= 20*time.Millisecond, elapsed.Round(10*time.Millisecond))
}
```

```console
$ go run ./ch18/durations
1h30m15s 5415000000000
90.25 2h0m0s 1h0m0s
1.5s 2µs 42ns
1m30s 90 <nil>
time: invalid duration "soon"
3s
true 20ms
```

- `30*time.Minute` works because `30` is an untyped constant. An `int` variable needs a conversion:
  `time.Duration(n) * time.Second`. `n * time.Second` is a compile error (mismatched types `int`
  and `time.Duration`).
- `ParseDuration` units: `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. No `d` for days: a day isn't
  always 24 hours (daylight saving time).
- `time.Now()` includes a *monotonic clock* reading, so `time.Since(start)` stays correct even if
  the system clock changes meanwhile. `seek` has used `time.Since` for `--stats` since
  [Chapter 13](13-generics.md).
- `time.Sleep(d)` pauses only the current goroutine.

## Timers and tickers

`time.After(d)` returns a channel that receives one value after `d`: the simplest timeout in a
`select`. For more control, use a `time.Timer` (one event; can be stopped or reset) or a
`time.Ticker` (an event every `d`):

`examples/ch18/timers/main.go`

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	// time.After returns a channel that receives once, after d. It is
	// the simplest timeout in a select.
	work := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		work <- "done"
	}()
	select {
	case r := <-work:
		fmt.Println(r)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}

	// A Timer is the same thing, but you can stop or reset it.
	t := time.NewTimer(time.Hour)
	fmt.Println("stopped before firing:", t.Stop())
	t.Reset(10 * time.Millisecond)
	<-t.C
	fmt.Println("timer fired")

	// A Ticker sends on its channel again and again.
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for i := range 3 {
		<-tick.C
		fmt.Println("tick", i)
	}

	// AfterFunc runs a function in its own goroutine after d.
	fired := make(chan bool)
	time.AfterFunc(5*time.Millisecond, func() { fired <- true })
	fmt.Println("AfterFunc:", <-fired)
}
```

```console
$ go run ./ch18/timers
timed out
stopped before firing: true
timer fired
tick 0
tick 1
tick 2
AfterFunc: true
```

The first part's work goroutine stays blocked forever on `work <- "done"`, since nobody receives:
Chapter 16's goroutine leak. Here `main` returns right after, so it doesn't matter. In a
long-running program, make `work` buffered (`make(chan string, 1)`) so the send always completes.

`Stop` the tickers you create, as the `defer` does. Since Go 1.23, unreferenced timers and tickers
are garbage collected even if not stopped, and `Reset` and `Stop` no longer leave a stale value in
the channel. Old advice to drain `t.C` before `Reset` no longer applies.

## Formatting and parsing times

A `time.Time` is an instant, with a location (time zone) for display. Go doesn't format with codes
like `%Y-%m-%d`. You write an example of the layout using one *reference time*:

```text
Mon Jan 2 15:04:05 MST 2006
```

Remember it as 1 2 3 4 5 6 7: month 1, day 2, hour 3 PM (15), minute 4, second 5, year (20)06,
zone -7 (MST). Write the reference time the way you want your output to look:

`examples/ch18/format/main.go`

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	// Layouts are written with the reference time:
	//   Mon Jan 2 15:04:05 MST 2006   (1 2 3 4 5 6 7: month day hour min sec year zone)
	t := time.Date(2026, time.March, 7, 9, 5, 3, 0, time.UTC)

	fmt.Println(t.Format("2006-01-02"))
	fmt.Println(t.Format("02 Jan 2006 15:04"))
	fmt.Println(t.Format("Monday, January 2 at 3:04 PM"))
	fmt.Println(t.Format(time.RFC3339))
	fmt.Println(t.Format(time.Kitchen), t.Format(time.DateTime))

	// Parse uses the same layouts.
	p, err := time.Parse("2006-01-02 15:04", "2026-12-31 23:59")
	fmt.Println(p, err)
	_, err = time.Parse("2006-01-02", "31/12/2026")
	fmt.Println(err)

	// Calendar arithmetic and comparison.
	next := t.AddDate(0, 1, 0) // one month later
	fmt.Println(next.Format(time.DateOnly), next.Sub(t), next.After(t))
	fmt.Println(t.Year(), t.Month(), t.Day(), t.Weekday(), t.YearDay())
	fmt.Println(t.Unix(), time.Unix(0, 0).UTC())

	// Zones: the same instant shown in another location.
	est := time.FixedZone("EST", -5*60*60)
	fmt.Println(t.In(est).Format(time.RFC1123), t.Equal(t.In(est)))
}
```

```console
$ go run ./ch18/format
2026-03-07
07 Mar 2026 09:05
Saturday, March 7 at 9:05 AM
2026-03-07T09:05:03Z
9:05AM 2026-03-07 09:05:03
2026-12-31 23:59:00 +0000 UTC <nil>
parsing time "31/12/2026" as "2006-01-02": cannot parse "31/12/2026" as "2006"
2026-04-07 744h0m0s true
2026 March 7 Saturday 66
1772874303 1970-01-01 00:00:00 +0000 UTC
Sat, 07 Mar 2026 04:05:03 EST true
```

Notes:

- Constants for common layouts: `time.RFC3339` (use it for machine-readable timestamps),
  `time.DateTime`, `time.DateOnly`, `time.TimeOnly`, `time.Kitchen`, and more.
- A layout must use the reference values exactly. `"2006-01-02"` is right; `"2026-01-02"` prints
  the literal text `2026` in every date.
- `time.Parse` assumes UTC unless the input has a zone. `time.ParseInLocation` takes a default
  location; `time.LoadLocation("America/New_York")` loads one by name.
- **Compare times with `Equal`, not `==`.** The same instant in two locations is `Equal` but not
  `==`, because `==` also compares the location and monotonic clock reading.
- `t.Sub(u)` gives a `Duration`; `t.Add(d)` adds one; `AddDate(years, months, days)` does calendar
  arithmetic.

## Signals: handling Ctrl-C

Ctrl-C in a terminal sends the program the *interrupt* signal (`SIGINT` on Unix, `os.Interrupt` in
Go). By default a Go program dies at once: deferred calls don't run, and buffered output is lost.

`signal.NotifyContext` turns the signal into a context cancellation instead:

`examples/ch18/signals/main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// ctx is canceled when the program receives os.Interrupt (Ctrl-C).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop() // restore the default behavior of Ctrl-C

	fmt.Println("working; press Ctrl-C to stop early")
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("finished normally")
	case <-ctx.Done():
		fmt.Println("stopping:", context.Cause(ctx))
	}
}
```

Run it twice: once letting it finish, once pressing Ctrl-C within 2 seconds:

```console
$ go run ./ch18/signals
working; press Ctrl-C to stop early
finished normally
$ go run ./ch18/signals
working; press Ctrl-C to stop early
^Cstopping: interrupt signal received
```

- The cause, `interrupt signal received`, comes from the `signal` package (Go 1.26+);
  `ctx.Err()` would just say `context canceled`.
- `stop` unregisters the handler. After `stop`, a second Ctrl-C kills the program the default way,
  so if your cleanup hangs, the user can still get out.
- To catch more signals, list them: `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`
  (`SIGTERM` is what `kill` and container runtimes send by default).
- The lower-level `signal.Notify(ch, os.Interrupt)` delivers signals on a channel. Use it to react
  to each signal, not just cancel once.

## Building seek: timeouts and Ctrl-C

`seek` gains:

- `--timeout D`: give up after `D` (e.g. `500ms`, `2s`), print what was found, and exit 2 with
  `seek: timed out after D`;
- Ctrl-C handling: stop, print what was found, and exit 2 with
  `seek: interrupt signal received`.

Both become a canceled context, which flows from `main` to `run`, to `SearchAll`, to each worker's
`SearchFile` and `Search`, which checks it on every line.

`seek` can already stop early, since [Chapter 7](07-structs-methods-pointers.md): with `-m N`,
`Search` `break`s once it has enough matches and never reads the rest of the file. `-l` sets
`MaxCount` to 1 for the same effect. That's stopping because the *work* is done; the context is for
stopping because the *caller* says so.

### Check the context in `internal/search/search.go`

Add `"context"` to the imports:

```go
import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync/atomic"
)
```

`SearchFile` and `Search` take a context as their first parameter. `SearchFile` passes it on:

```go
// SearchFile searches the file at path. The path "-" means s.Stdin.
func (s *Searcher) SearchFile(ctx context.Context, path string) ([]Match, error) {
	if path == "-" {
		return s.Search(ctx, s.Stdin, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // we only read, so Close can't lose data
	return s.Search(ctx, f, path)
}
```

`Search` checks it at the top of the scan loop:

```go
// Search reads r line by line and returns the selected lines. path is
// recorded in each Match. It returns ErrBinary, and no matches, if the
// first few KiB of r contain a NUL byte, and ctx.Err() if ctx is
// canceled before it reaches the end.
func (s *Searcher) Search(ctx context.Context, r io.Reader, path string) ([]Match, error) {
	r = countingReader{r: r, n: &s.bytesRead}

	// Peek at the start of the input without consuming it, so the
	// scanner below still sees every byte.
	br := bufio.NewReaderSize(r, binaryCheckSize)
	head, err := br.Peek(binaryCheckSize)
	if err != nil && err != io.EOF {
		return nil, err // io.EOF just means the input is short
	}
	if bytes.IndexByte(head, 0) >= 0 {
		return nil, ErrBinary
	}

	sc := bufio.NewScanner(br)
	sc.Buffer(make([]byte, 0, 64<<10), maxLineSize)
	var matches []Match
	for lineNum := 1; sc.Scan(); lineNum++ {
		// Checking once per line lets a huge file, or endless input,
		// be stopped quickly. ctx.Err is cheap: an atomic load.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := sc.Text()
		if !s.selects(line) {
			continue
		}
		matches = append(matches, Match{Path: path, Line: lineNum, Text: line})
		if len(matches) == s.MaxCount {
			break // stop reading: the rest of the file doesn't matter
		}
	}
	return matches, sc.Err()
}
```

Why `ctx.Err()` and not a `select` on `ctx.Done()`? The loop is busy working, not blocked; there's
nothing else to wait for. Poll `Err` in a busy loop; `select` on `Done` in a goroutine that waits.

### Replace `done` with a context in `internal/search/parallel.go`

Add `"context"` to the imports:

```go
import (
	"context"
	"iter"
	"sync"
)
```

`SearchAll` takes a context. Inside, its own child context from `WithCancel` replaces the `done`
channel. The full function:

```go
// SearchAll searches files using up to workers goroutines at once and
// returns an iterator over the results. Even though files finish in any
// order, results are yielded in the order files produced them. An error
// from files is passed through in the Result for its path.
//
// Canceling ctx stops the search: SearchAll yields nothing more, and
// the caller can check ctx.Err() to tell that apart from finishing.
//
// The design: a producer goroutine reads files and gives each one a
// buffered channel for its result. It hands the file to the workers
// through the jobs channel and also queues the result channel on
// pending, in order. The consumer (the loop below) receives from the
// queued channels one by one, so it waits for the oldest file even if
// newer ones are already done.
func (s *Searcher) SearchAll(ctx context.Context, files iter.Seq2[string, error], workers int) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		// Canceling ctx tells every goroutine to stop. Our own child
		// context lets us cancel it when we return early, without
		// touching the caller's ctx. The deferred cancel runs however we
		// leave, including when the caller's loop breaks.
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

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
				case <-ctx.Done():
					return
				}
				if err != nil {
					continue
				}
				select {
				case jobs <- job{path: path, result: result}:
				case <-ctx.Done():
					return
				}
			}
		})

		// Workers.
		for range workers {
			wg.Go(func() {
				for j := range jobs {
					matches, err := s.SearchFile(ctx, j.path)
					j.result <- Result{Path: j.path, Matches: matches, Err: err}
				}
			})
		}

		// Consumer: results in order.
		for result := range pending {
			var res Result
			select {
			case res = <-result:
			case <-ctx.Done():
				// Don't wait for the goroutines: a worker may be stuck
				// in a Read that ignores ctx, such as one from a
				// terminal. Returning lets the program exit anyway.
				return
			}
			// If both cases were ready, select picked one at random;
			// check again so that cancellation always wins.
			if ctx.Err() != nil || !yield(res) {
				return // the deferred cancel stops the other goroutines
			}
		}
		// Every result has been delivered, so the goroutines are all
		// finishing. Waiting guarantees that none outlives SearchAll.
		wg.Wait()
	}
}
```

What changed:

- `done := make(chan struct{})` / `defer close(done)` became `ctx, cancel :=
  context.WithCancel(ctx)` / `defer cancel()`, and `<-done` became `<-ctx.Done()`. The child is
  canceled by the caller (Ctrl-C, timeout) *or* by our `defer cancel()` (the caller's loop
  `break`s). Shadowing `ctx` is deliberate: from here on, nothing can accidentally use the parent.
- Workers pass `ctx` to `SearchFile`, so a canceled search stops mid-file.
- The consumer no longer blocks on `<-result` alone. If the oldest file is stuck (e.g. reading a
  terminal where nobody types), a plain `<-result` would wait forever, Ctrl-C or not. The `select`
  lets cancellation win.
- After the `select`, `ctx.Err() != nil` is checked again. When both cases are ready, `select`
  picks one *at random*, so without the check a search could *sometimes* yield one more result
  after cancellation. "Sometimes" bugs are the worst kind.

### Test it: `internal/search/parallel_test.go`

Add `"context"` to the test's imports and pass `context.Background()` in `TestSearchAllOrder`:

```go
	for res := range s.SearchAll(context.Background(), Walk(paths, WalkOptions{}), 8) {
```

Then add a test that cancels before the search starts and expects nothing:

```go
// TestSearchAllCanceled checks that a canceled search yields nothing.
func TestSearchAllCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := &Searcher{Matcher: literal("fox")}
	for res := range s.SearchAll(ctx, Walk([]string{"."}, WalkOptions{}), 4) {
		t.Errorf("got result for %s after cancel", res.Path)
	}
}
```

This test catches the random-`select` bug: without the `ctx.Err() != nil` check, it fails some of
the time.

### Add `--timeout` in `cmd/seek/flags.go`

Add `"time"` to the imports, and a help line after `-j`:

```text
  -j N                      search N files at a time (default: number of CPUs)
      --timeout D           give up after duration D, e.g. 500ms or 2s
      --color WHEN          highlight matches: auto, always, or never
```

Add a `timeout` field to `options`, after `workers`. `gofmt` realigns the block's comments because
the new type name is longer:

```go
type options struct {
	pattern     string
	paths       []string
	ignoreCase  bool          // -i
	invert      bool          // -v
	lineNumbers bool          // -n
	count       bool          // -c
	filesOnly   bool          // -l
	maxCount    int           // -m N; 0 means no limit
	color       colorMode     // --color WHEN
	regex       bool          // -E
	hidden      bool          // --hidden
	glob        string        // --glob PATTERN
	workers     int           // -j N
	timeout     time.Duration // --timeout D; 0 means none
	json        bool          // --json
	stats       bool          // --stats
	version     bool          // --version
}
```

In `parseArgs`, define the flag after `-j`. `flag.DurationVar` parses the value with
`time.ParseDuration`:

```go
	flags.IntVar(&opts.workers, "j", runtime.NumCPU(), "")
	flags.DurationVar(&opts.timeout, "timeout", 0, "") // parses "1m30s"
	flags.Var(&opts.color, "color", "")
```

and reject negative durations next to the `-j` check:

```go
	if opts.workers < 1 {
		return opts, usageErrorf("invalid value for -j: must be at least 1")
	}
	if opts.timeout < 0 {
		return opts, usageErrorf("invalid value for --timeout: %v is negative", opts.timeout)
	}
```

### Wire it up in `cmd/seek/main.go`

Add `"context"` and `"os/signal"` to the imports. `main` creates the signal context and passes it
to `run`:

```go
func main() {
	// Ctrl-C sends os.Interrupt. Instead of killing the program, it now
	// cancels ctx, so seek can stop cleanly and print what it found.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop() // os.Exit skips deferred calls, so call stop directly
	os.Exit(code)
}
```

`defer stop()` would never run: `os.Exit` skips deferred calls ([Chapter 4](04-functions.md)). So
`main` calls `stop` itself.

`run` takes `ctx` as its first parameter:

```go
// run is the whole program. It reads standard input from stdin, writes
// results to stdout and messages to stderr, and returns the exit status.
// Taking an io.Reader and io.Writers instead of using os.Stdin, os.Stdout,
// and os.Stderr directly lets tests supply input and capture output.
// Canceling ctx stops the search early.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
```

After the `--version` check, add the timeout. It wraps the signal context, so either can cancel the
search:

```go
	if opts.timeout > 0 {
		// The cause is the error context.Cause reports once time is up.
		cause := fmt.Errorf("timed out after %v", opts.timeout)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, opts.timeout, cause)
		defer cancel() // release the timer if we finish in time
	}
```

`var cancel` and `=` (not `:=`) reassign `ctx` instead of shadowing it in the `if` block. With `:=`,
the new `ctx` would vanish at the closing brace.

Pass `ctx` to `SearchAll`:

```go
	for res := range s.SearchAll(ctx, files, opts.workers) {
```

and after the loop, before `--stats`, report a cancellation:

```go
	// context.Cause is nil unless ctx was canceled; then it says why,
	// e.g. "interrupt signal received" or "timed out after 2s".
	if err := context.Cause(ctx); err != nil {
		out.Flush() // print the results found so far first
		report(stderr, err)
		hadError = true
	}
```

`hadError` makes the exit status 2. Results printed before the cancellation stay printed.

### Checkpoint

```console
$ go test -race ./...
?   	seek/cmd/seek	[no test files]
?   	seek/internal/output	[no test files]
ok  	seek/internal/search	1.014s
?   	seek/internal/stats	[no test files]
$ go build -o seek ./cmd/seek
$ ./seek --timeout 5s -n fox testdata; echo "exit $?"
testdata/poem.txt:1:The quick brown fox
testdata/poem.txt:3:The dog sleeps; the fox runs.
testdata/src/hello.go:4:	println("hello, fox")
exit 0
$ yes 'the quick brown fox' | ./seek -c --timeout 100ms fox -; echo "exit $?"
seek: timed out after 100ms
exit 2
$ yes 'the quick brown fox' | ./seek --stats -c --timeout 100ms fox - testdata; echo "exit $?"
seek: timed out after 100ms
files searched: 0
bytes read:     10666943
files matched:  0
total matches:  0
elapsed:        100.695ms
exit 2
$ ./seek --timeout soon fox testdata; echo "exit $?"
seek: invalid value "soon" for flag -timeout: parse error
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit 2
$ ./seek --timeout -1s fox testdata; echo "exit $?"
seek: invalid value for --timeout: -1s is negative
Usage: seek [OPTIONS] PATTERN [PATH ...]
Run 'seek --help' for more information.
exit 2
```

`yes` prints its argument forever, so searching `-` never ends on its own; the timeout stops it.
Your `bytes read` and `elapsed` will differ.

In the `--stats` run, `files searched: 0`, though `testdata` was given too. Other workers probably
searched its files, but results are delivered **in order**, and standard input comes first. Since
`-` never finished, nothing after it was printed. That's the price of ordered output: one slow file
holds up everything behind it. `grep` pays it too; `ripgrep`, which prints in any order by default,
doesn't.

Now try Ctrl-C. Run `./seek fox -`, type a few lines (lines with `fox` print once you press Ctrl-D,
as before), then press Ctrl-C instead:

```console
$ ./seek fox -
the quick brown fox
^Cseek: interrupt signal received
$ echo "exit $?"
exit 2
```

The worker searching standard input is still blocked in `Read`, waiting for the terminal, and a
blocked `Read` ignores contexts. `seek` exits at once anyway: the consumer's `select` sees
`ctx.Done()`, and `SearchAll` doesn't wait for the workers. (Your typed line isn't printed: the
search of `-` never finished.)

Without a terminal: a pipe that never sends anything (from `sleep`) plays the idle typist, and
`kill -INT` sends the same signal as Ctrl-C:

```console
$ sleep 3 | ./seek fox - & sleep 0.5; kill -INT $!; wait $!; echo "exit $?"
seek: interrupt signal received
exit 2
```

Ctrl-C on `yes ... | ./seek ...` gives mixed results: the terminal signals *both* programs, and if
`yes` dies first, `seek` sees end of input and finishes normally.

The complete project at this point is in `project/ch18/`.

## Summary

- A `context.Context` carries cancellation, a deadline, and request-scoped values across API
  boundaries. Start from `context.Background()`; derive children with `With...`. Canceling a parent
  cancels its children.
- `WithCancel` returns a `cancel` function; `WithTimeout`/`WithDeadline` cancel automatically.
  Always `defer cancel()`; `go vet` reports a discarded one.
- `ctx` is the first parameter. Waiting code selects on `ctx.Done()`; busy loops poll `ctx.Err()`
  (`nil`, `context.Canceled`, or `context.DeadlineExceeded`).
- `WithCancelCause`, `WithTimeoutCause`, and `context.Cause` say *why*. `context.AfterFunc` runs a
  function on cancellation.
- `WithValue` is for request-scoped data only, with an unexported key type and accessor functions.
- `time.Duration` is nanoseconds; `ParseDuration` reads `1m30s`; `time.Since` measures elapsed
  time. `time.After`, `Timer`, and `Ticker` deliver events on channels.
- Layouts use the reference time `Mon Jan 2 15:04:05 MST 2006`. Compare times with `Equal`.
- `signal.NotifyContext` turns Ctrl-C into cancellation; `stop` restores the default.
- `seek` has `--timeout` and stops cleanly on Ctrl-C, printing results found so far and exiting 2.

## Exercises

1. **A cancelable sleep.** `time.Sleep` can't be interrupted. Write `sleepCtx(ctx
   context.Context, d time.Duration) error` that returns `nil` after `d`, or `ctx.Err()` as soon
   as `ctx` is canceled. Use a `time.Timer` (not `time.After`) so the timer is stopped when the
   context wins. Test with a 50 ms timeout: a 10 ms sleep succeeds; a one-hour sleep fails after
   the remaining 40 ms.

<details><summary>Solution</summary>

```go
package main

import (
	"context"
	"fmt"
	"time"
)

// sleepCtx pauses for d, or until ctx is canceled, whichever is first.
// It returns nil after a full sleep and ctx.Err() otherwise.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	fmt.Println(sleepCtx(ctx, 10*time.Millisecond))
	fmt.Println(sleepCtx(ctx, time.Hour))
	fmt.Println(time.Since(start).Round(50 * time.Millisecond))
}
```

```console
$ go run .
<nil>
context deadline exceeded
50ms
```

The total is 50 ms, not an hour: the deadline covers both calls.

</details>

2. **A progress ticker.** Start a goroutine that simulates a slow job (`time.Sleep(350 *
   time.Millisecond)`, then sends `"result: 42"` on a channel). While waiting, print `still
   working (...)` with the elapsed time every 100 ms, then print the result.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan string)
	go func() {
		time.Sleep(350 * time.Millisecond) // the slow job
		done <- "result: 42"
	}()

	start := time.Now()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case r := <-done:
			fmt.Println(r)
			return
		case <-tick.C:
			fmt.Printf("still working (%v)\n", time.Since(start).Round(100*time.Millisecond))
		}
	}
}
```

```console
$ go run .
still working (100ms)
still working (200ms)
still working (300ms)
result: 42
```

</details>

3. **"5m ago".** Write `ago(t, now time.Time) string` that returns `just now` (under a minute),
   `5m ago` (under an hour), `3h ago` (under a day), `2d ago` (under a week), and otherwise `on
   Feb 5, 2026`. Taking `now` as a parameter, instead of calling `time.Now()` inside, makes it easy
   to test. Try times 10 s, 5 min, 3 h, 50 h, and 30 days before a fixed `now`.

<details><summary>Solution</summary>

```go
package main

import (
	"fmt"
	"time"
)

// ago describes how long before now t was.
func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return "on " + t.Format("Jan 2, 2006")
	}
}

func main() {
	now := time.Date(2026, time.March, 7, 12, 0, 0, 0, time.UTC)
	for _, d := range []time.Duration{
		10 * time.Second, 5 * time.Minute, 3 * time.Hour, 50 * time.Hour, 30 * 24 * time.Hour,
	} {
		fmt.Println(ago(now.Add(-d), now))
	}
}
```

```console
$ go run .
just now
5m ago
3h ago
2d ago
on Feb 5, 2026
```

</details>

4. **A separate exit status for timeouts.** Scripts may want to tell "timed out" from other errors.
   Make `seek` exit with status 3 when `--timeout` expires; Ctrl-C and other errors still exit 2.
   Hint: with `WithTimeoutCause`, `context.Cause` returns your message, but `ctx.Err()` is still
   `context.DeadlineExceeded`. (If you keep this change, update the help text's exit status list.)

<details><summary>Solution</summary>

Add a constant:

```go
const (
	exitOK      = iota // 0: success (at least one line matched)
	exitNoMatch        // 1: no line matched
	exitError          // 2: bad usage or an error occurred
	exitTimeout        // 3: --timeout expired
)
```

Remember the kind of cancellation after the loop:

```go
	timedOut := false
	if err := context.Cause(ctx); err != nil {
		out.Flush() // print the results found so far first
		report(stderr, err)
		hadError = true
		timedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
```

and check it first at the end:

```go
	switch {
	case timedOut:
		return exitTimeout
	case hadError:
		return exitError
	case matched.Total() == 0:
		return exitNoMatch
	default:
		return exitOK
	}
```

```console
$ yes 'the quick brown fox' | ./seek -c --timeout 100ms fox -; echo "exit $?"
seek: timed out after 100ms
exit 3
$ sleep 3 | ./seek fox - & sleep 0.5; kill -INT $!; wait $!; echo "exit $?"
seek: interrupt signal received
exit 2
```

Ctrl-C cancels the signal context, so `ctx.Err()` is `context.Canceled`, not `DeadlineExceeded`.

</details>

5. **Why check twice?** In `SearchAll`'s consumer, the `select` already has a `case
   <-ctx.Done()`. Explain why the next line still checks `ctx.Err() != nil` before calling
   `yield`, and describe a run of `TestSearchAllCanceled` that would fail without it.

<details><summary>Solution</summary>

When several `select` cases are ready, Go picks one at random. In `TestSearchAllCanceled` the
context is canceled before the search starts, but the producer and workers still race with the
cancellation:

1. A worker finishes a small file and sends its result before noticing the cancellation.
2. At the consumer's `select`, both `res = <-result` and `<-ctx.Done()` are ready; about half the
   time it picks the result.
3. Without the second check, it calls `yield`, and the test reports
   `got result for ... after cancel`.

It fails only sometimes, so the bug is hard to track down. The extra `ctx.Err()` check makes
cancellation always win.

</details>

Sources for this chapter: [REFERENCES.md](REFERENCES.md#chapter-18--context-cancellation-time-and-signals)
