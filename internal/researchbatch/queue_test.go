package researchbatch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestQueuePendingBatchCancellationClose(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var batches [][]int
	q, err := New(8, 4, time.Second, func(ctx context.Context, values []int) ([]int, error) {
		batches = append(batches, append([]int(nil), values...))
		if len(batches) == 1 {
			close(entered)
			<-release
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return values, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	first := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _, e := q.Put(ctx, 0); first <- e }()
	<-entered
	cancel() // An in-flight commit must settle, not be abandoned.
	select {
	case <-first:
		t.Fatal("returned before settlement")
	default:
	}
	// Inject pending jobs deterministically; this avoids scheduler-dependent batch sizes.
	jobs := make([]job[int, int], 5)
	for i := range jobs {
		jobs[i] = job[int, int]{ctx: context.Background(), value: i + 1, reply: make(chan reply[int], 1)}
		q.in <- jobs[i]
	}
	close(release)
	q.Close()
	if err := <-first; err != nil {
		t.Fatal(err)
	}
	for _, j := range jobs {
		r := <-j.reply
		if r.err != nil || r.value != j.value {
			t.Fatal(r)
		}
	}
	if len(batches) != 3 || len(batches[0]) != 1 || len(batches[1]) != 4 || len(batches[2]) != 1 {
		t.Fatal(batches)
	}
	if _, err := q.Put(context.Background(), 9); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	q.Close()
}

func TestQueueRejectsCanceledAndStopsAfterCommitError(t *testing.T) {
	count := 0
	sentinel := errors.New("uncertain commit")
	q, _ := New(2, 2, time.Second, func(context.Context, []int) ([]int, error) { count++; return nil, sentinel })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := q.Put(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := q.Put(context.Background(), 2); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if _, err := q.Put(context.Background(), 3); !errors.Is(err, ErrReconcile) {
		t.Fatal(err)
	}
	q.Close()
	if count != 1 {
		t.Fatal(count)
	}
}

func TestQueueFullAndQueuedCancellation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	q, _ := New(1, 1, time.Second, func(_ context.Context, values []int) ([]int, error) { close(entered); <-release; return values, nil })
	var group sync.WaitGroup
	group.Add(1)
	go func() { defer group.Done(); _, _ = q.Put(context.Background(), 0) }()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	j := job[int, int]{ctx: ctx, value: 1, reply: make(chan reply[int], 1)}
	q.in <- j
	if _, err := q.Put(context.Background(), 2); !errors.Is(err, ErrFull) {
		t.Fatal(err)
	}
	cancel()
	close(release)
	q.Close()
	group.Wait()
	if r := <-j.reply; !errors.Is(r.err, context.Canceled) {
		t.Fatal(r)
	}
}

func TestQueueDeadlineAndCardinality(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		q, _ := New(1, 1, time.Second, func(ctx context.Context, values []int) ([]int, error) {
			if malformed {
				return nil, nil
			}
			d, ok := ctx.Deadline()
			if !ok || time.Until(d) > 100*time.Millisecond {
				return nil, errors.New("wrong deadline")
			}
			<-ctx.Done()
			return nil, ctx.Err()
		})
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		_, err := q.Put(ctx, 1)
		cancel()
		if err == nil {
			t.Fatal("failure accepted")
		}
		q.Close()
	}
}
