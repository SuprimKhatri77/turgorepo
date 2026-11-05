// Package background runs work that shouldn't delay an HTTP response.
package background

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
)

const jobTimeout = 30 * time.Second

// Runner starts jobs on their own goroutines and tracks them, so the API can let
// them finish before it shuts down.
type Runner struct {
	wg sync.WaitGroup
}

// Go runs job in the background. A panic is logged rather than crashing the API.
func (r *Runner) Go(name string, job func(ctx context.Context)) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				slog.Error("background job panicked", "job", name, "err", err, "stack", string(debug.Stack()))
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
		defer cancel()
		job(ctx)
	}()
}

// Wait blocks until all jobs finish or ctx ends, and reports whether they finished.
func (r *Runner) Wait(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
