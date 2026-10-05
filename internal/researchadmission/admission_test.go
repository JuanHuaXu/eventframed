package researchadmission

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func waitQueued(t *testing.T, g *Scheduler, want [3]int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if g.Snapshot().Queued == want {
			return
		}
		runtime.Gosched()
	}
	t.Fatal("queued timeout", g.Snapshot(), want)
}
func TestFixedRotationAndExclusion(t *testing.T) {
	g, _ := NewScheduler(2, 32)
	ctx := context.Background()
	hold, _ := g.Acquire(ctx, Ingest)
	type acquired struct {
		kind  Kind
		lease *Lease
	}
	ready := make(chan acquired, 16)
	var wg sync.WaitGroup
	for k := Outcome; k <= Ingest; k++ {
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(kind Kind) {
				defer wg.Done()
				l, e := g.Acquire(ctx, kind)
				if e != nil {
					t.Error(e)
					return
				}
				ready <- acquired{kind, l}
			}(k)
		}
	}
	waitQueued(t, g, [3]int{4, 4, 4})
	hold.Release()
	for _, kind := range []Kind{Outcome, Outcome, Recall, Recall, Ingest, Outcome, Outcome, Recall, Recall, Ingest, Ingest, Ingest} {
		select {
		case got := <-ready:
			if got.kind != kind {
				t.Fatal("rotation", got.kind, kind)
			}
			s := g.Snapshot()
			if s.ActiveReaders > 0 && s.ActiveWriter {
				t.Fatal("overlap", s)
			}
			got.lease.Release()
			got.lease.Release()
		case <-time.After(3 * time.Second):
			t.Fatal("grant timeout")
		}
	}
	wg.Wait()
	s := g.Snapshot()
	if s.ActiveReaders != 0 || s.ActiveWriter || s.MaxReaders != 2 {
		t.Fatal(s)
	}
}
func TestCancellationCapacityAndClose(t *testing.T) {
	g, _ := NewScheduler(2, 1)
	hold, _ := g.Acquire(context.Background(), Recall)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		l, e := g.Acquire(ctx, Outcome)
		if l != nil {
			l.Release()
		}
		done <- e
	}()
	waitQueued(t, g, [3]int{1, 0, 0})
	if _, e := g.Acquire(context.Background(), Ingest); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	cancel()
	if e := <-done; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	waitQueued(t, g, [3]int{})
	go func() { _, e := g.Acquire(context.Background(), Ingest); done <- e }()
	waitQueued(t, g, [3]int{0, 0, 1})
	g.Close()
	if e := <-done; !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
	hold.Release()
	g.Close()
	if _, e := g.Acquire(context.Background(), Recall); !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
	if _, e := g.Acquire(context.Background(), Kind(99)); !errors.Is(e, ErrKind) {
		t.Fatal(e)
	}
	if _, e := NewScheduler(0, 1); e == nil {
		t.Fatal("invalid bounds accepted")
	}
}
func TestConcurrentCancellationNoLeaks(t *testing.T) {
	g, _ := NewScheduler(8, 512)
	var wg sync.WaitGroup
	var check sync.Mutex
	readers, writers := 0, 0
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if i%5 == 0 {
				cancel()
			}
			kind := Kind(i % 3)
			l, e := g.Acquire(ctx, kind)
			if e != nil {
				if !errors.Is(e, context.Canceled) && !errors.Is(e, context.DeadlineExceeded) && !errors.Is(e, ErrCapacity) {
					t.Error(e)
				}
				return
			}
			check.Lock()
			if kind == Recall {
				readers++
			} else {
				writers++
			}
			if writers > 1 || writers > 0 && readers > 0 || readers > 8 {
				t.Error("overlapping owners", readers, writers)
			}
			check.Unlock()
			runtime.Gosched()
			check.Lock()
			if kind == Recall {
				readers--
			} else {
				writers--
			}
			check.Unlock()
			l.Release()
		}(i)
	}
	wg.Wait()
	s := g.Snapshot()
	if s.Queued != ([3]int{}) || s.ActiveReaders != 0 || s.ActiveWriter {
		t.Fatal("leaked lease", s)
	}
}
func BenchmarkLease(b *testing.B) {
	g, _ := NewScheduler(8, 512)
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l, e := g.Acquire(ctx, Recall)
		if e != nil {
			b.Fatal(e)
		}
		l.Release()
	}
}

func TestFIFOAndActiveCancellation(t *testing.T) {
	g, _ := NewScheduler(2, 32)
	ctx, cancel := context.WithCancel(context.Background())
	read, _ := g.Acquire(ctx, Recall)
	cancel()
	type result struct {
		index int
		lease *Lease
	}
	ready := make(chan result, 8)
	for i := 0; i < 8; i++ {
		go func(i int) {
			l, e := g.Acquire(context.Background(), Outcome)
			if e != nil {
				t.Error(e)
				return
			}
			ready <- result{i, l}
		}(i)
		waitQueued(t, g, [3]int{i + 1, 0, 0})
	}
	select {
	case <-ready:
		t.Fatal("active cancellation released caller lease")
	default:
	}
	read.Release()
	for i := 0; i < 8; i++ {
		select {
		case r := <-ready:
			if r.index != i {
				t.Fatal("FIFO", r.index, i)
			}
			r.lease.Release()
		case <-time.After(3 * time.Second):
			t.Fatal("grant timeout")
		}
	}
}
