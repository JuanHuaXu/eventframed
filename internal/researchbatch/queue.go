// Package researchbatch experiments with bounded, synchronous write batching.
// It is not wired into the daemon. Inputs must be immutable while submitted.
package researchbatch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrFull      = errors.New("research batch queue full")
	ErrClosed    = errors.New("research batch queue closed")
	ErrReconcile = errors.New("research batch requires reconciliation")
)

type reply[R any] struct {
	value R
	err   error
}
type job[T, R any] struct {
	ctx   context.Context
	value T
	reply chan reply[R]
}

type Queue[T, R any] struct {
	mu      sync.Mutex
	closed  bool
	in      chan job[T, R]
	done    chan struct{}
	max     int
	timeout time.Duration
	commit  func(context.Context, []T) ([]R, error)
}

func New[T, R any](capacity, max int, timeout time.Duration, commit func(context.Context, []T) ([]R, error)) (*Queue[T, R], error) {
	if capacity < 1 || max < 1 || max > 16 || timeout <= 0 || commit == nil {
		return nil, errors.New("invalid research queue configuration")
	}
	q := &Queue[T, R]{in: make(chan job[T, R], capacity), done: make(chan struct{}), max: max, timeout: timeout, commit: commit}
	go q.run()
	return q, nil
}

// Put waits for settlement after admission, even if its context is canceled.
// Canceling one caller cannot abandon an in-flight shared commit or make the
// caller believe it rolled back. Call latency therefore may exceed its deadline;
// experiments must measure that as a latency failure, not hide it as success.
func (q *Queue[T, R]) Put(ctx context.Context, value T) (R, error) {
	var zero R
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	j := job[T, R]{ctx: ctx, value: value, reply: make(chan reply[R], 1)}
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return zero, ErrClosed
	}
	select {
	case q.in <- j:
		q.mu.Unlock()
	default:
		q.mu.Unlock()
		return zero, ErrFull
	}
	r := <-j.reply
	return r.value, r.err
}

// Close drains accepted jobs and joins the worker. The commit callback must
// honor its context and settle before returning; it must not spawn writes that
// outlive its return. Close never closes the underlying store.
func (q *Queue[T, R]) Close() {
	q.mu.Lock()
	if !q.closed {
		q.closed = true
		close(q.in)
	}
	q.mu.Unlock()
	<-q.done
}

func (q *Queue[T, R]) run() {
	defer close(q.done)
	var failed error
	for first := range q.in {
		jobs := []job[T, R]{first}
	drain:
		for len(jobs) < q.max {
			select {
			case j, ok := <-q.in:
				if !ok {
					break drain
				}
				jobs = append(jobs, j)
			default:
				break drain
			}
		}
		active := make([]job[T, R], 0, len(jobs))
		values := make([]T, 0, len(jobs))
		deadline := time.Now().Add(q.timeout)
		for _, j := range jobs {
			err := failed
			if err == nil {
				err = j.ctx.Err()
			}
			if err != nil {
				j.reply <- reply[R]{err: err}
				continue
			}
			if d, ok := j.ctx.Deadline(); ok && d.Before(deadline) {
				deadline = d
			}
			active = append(active, j)
			values = append(values, j.value)
		}
		if len(active) == 0 {
			continue
		}
		// Only the earliest deadline, not one caller's manual cancellation, can
		// cancel a shared commit. This is an explicit settlement contract.
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		results, err := q.commit(ctx, values)
		cancel()
		if err == nil && len(results) != len(active) {
			err = errors.New("batch result cardinality mismatch")
		}
		if err != nil {
			// Conservatively halt later transactions after any commit error.
			// The store may have committed despite returning an error.
			failed = fmt.Errorf("%w: %v", ErrReconcile, err)
		}
		for i, j := range active {
			r := reply[R]{err: err}
			if err == nil {
				r.value = results[i]
			}
			j.reply <- r
		}
	}
}
