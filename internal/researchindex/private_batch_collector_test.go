package researchindex

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPrivateBatchCollectorGroups(t *testing.T) {
	calls := 0
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if e != nil {
		t.Fatal(e)
	}
	c, e := NewPrivateBatchCollector(w, 4, 50*time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make([]PrivateBatchResult, 4)
	errs := make([]error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = c.Submit(context.Background(), PrivateInsertRequest{fmt.Sprint(i), []float32{float32(i), 1}, 0})
		}(i)
	}
	wg.Wait()
	if e = c.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal("expected full batch", calls)
	}
	for i, r := range results {
		if errs[i] != nil || r.BatchSize != 4 || r.BatchID != 1 {
			t.Fatal(r, errs[i])
		}
	}
	if _, e = c.Submit(context.Background(), PrivateInsertRequest{"late", []float32{1, 0}, 0}); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if e = w.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
}

func TestPrivateBatchCollectorCapacityAndCommit(t *testing.T) {
	entered, proceed := make(chan struct{}), make(chan struct{})
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, func(context.Context, uint64, []Mutation) error { close(entered); <-proceed; return nil })
	if e != nil {
		t.Fatal(e)
	}
	c, e := NewPrivateBatchCollector(w, 1, 0)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, e := c.Submit(ctx, PrivateInsertRequest{"a", []float32{1, 0}, 0}); done <- e }()
	<-entered
	if _, e = c.Submit(context.Background(), PrivateInsertRequest{"b", []float32{0, 1}, 0}); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	cancel()
	select {
	case e := <-done:
		t.Fatal("returned before commit outcome", e)
	default:
	}
	expired, stop := context.WithCancel(context.Background())
	stop()
	if e = c.Close(expired); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	close(proceed)
	if e = <-done; e != nil {
		t.Fatal("durable success lost", e)
	}
	if e = c.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
	lease, e := w.Acquire(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if _, ok, e := lease.Lookup(0); e != nil || !ok {
		t.Fatal(e)
	}
	lease.Release()
}

func TestPrivateBatchCollectorExpiredQueued(t *testing.T) {
	calls := 0
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if e != nil {
		t.Fatal(e)
	}
	c, e := NewPrivateBatchCollector(w, 4, 50*time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, e = c.Submit(ctx, PrivateInsertRequest{"expired", []float32{1, 0}, 0}); e == nil {
		t.Fatal("expired committed")
	}
	if e = c.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
	if calls != 0 {
		t.Fatal("expired persisted", calls)
	}
}
