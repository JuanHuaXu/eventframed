package researchmemory

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCompletionWaitLifecycle(t *testing.T) {
	b, e := NewBackground(1, 42, 64)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = b.WaitProcessed(ctx, 0); e != nil {
		t.Fatal(e)
	}
	cancelled, stop := context.WithCancel(ctx)
	stop()
	if e = b.WaitProcessed(cancelled, 1); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	now := time.Now()
	// Many overlapping waiters race registration against actual worker completion.
	done := make(chan error, 64)
	for i := 1; i <= 64; i++ {
		go func(n uint64) { done <- b.WaitProcessed(ctx, n) }(uint64(i))
		p, e := b.Predict(uint16(i), .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		if e = b.Feedback(p.ID, true, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	for i := 0; i < 64; i++ {
		if e = <-done; e != nil {
			t.Fatal(e)
		}
	}
	n, f, _, _ := b.Counts()
	if n != 64 || f != 0 {
		t.Fatal(n, f)
	}
	if e = b.WaitProcessed(ctx, 64); e != nil {
		t.Fatal(e)
	}
	go func() { done <- b.WaitProcessed(ctx, 65) }()
	b.Close()
	if e = <-done; e == nil {
		t.Fatal("close satisfied unfinished target")
	}
}

func TestCompletionFailureAndCancellation(t *testing.T) {
	// Explicit counter fixture covers failed completion semantics independently
	// of the learner's rejection paths. Registration is observed under its lock.
	b := &Background{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- b.WaitProcessed(ctx, 1) }()
	for {
		b.mu.Lock()
		ready := b.changed != nil
		if ready {
			b.failed = 1
			b.notifyLocked()
		}
		b.mu.Unlock()
		if ready {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	short, stop := context.WithTimeout(ctx, time.Millisecond)
	defer stop()
	if e := b.WaitProcessed(short, 2); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
