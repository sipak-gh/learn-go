package search

import (
	"context"
	"iter"
	"sync"
)

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
