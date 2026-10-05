package researchindex

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func publish(t *testing.T, m *Generations, changes ...Mutation) {
	t.Helper()
	p, err := m.Prepare(context.Background(), changes)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Commit(); err != nil {
		t.Fatal(err)
	}
}
func expect(t *testing.T, v View, id string, want []float32) {
	t.Helper()
	got, ok := v.Lookup(id)
	if ok != (want != nil) || !reflect.DeepEqual(got, want) {
		t.Fatalf("%s got%v present%v want%v", id, got, ok, want)
	}
}

func TestGenerationInvisiblePrepareAbortAndOwnership(t *testing.T) {
	m, _ := NewGenerations(2, 4)
	ctx := context.Background()
	old := m.View()
	input := []float32{1, 2}
	p, err := m.Prepare(ctx, []Mutation{{ID: "a", Vector: input}})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 99
	expect(t, m.View(), "a", nil)
	if err := p.Abort(); err != nil {
		t.Fatal(err)
	}
	expect(t, m.View(), "a", nil)
	if err := p.Commit(); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
	input = []float32{1, 2}
	p, err = m.Prepare(ctx, []Mutation{{ID: "a", Vector: input}})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 77
	if err := p.Commit(); err != nil {
		t.Fatal(err)
	}
	expect(t, old, "a", nil)
	expect(t, m.View(), "a", []float32{1, 2})
	copy, _ := m.View().Lookup("a")
	copy[0] = 88
	expect(t, m.View(), "a", []float32{1, 2})
}

func TestGenerationCompactionRebasesInterveningChanges(t *testing.T) {
	m, _ := NewGenerations(2, 4)
	publish(t, m, Mutation{ID: "a", Vector: []float32{1, 0}}, Mutation{ID: "b", Vector: []float32{0, 1}})
	old := m.View()
	c, err := m.PrepareCompaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.PrepareCompaction(context.Background()); !errors.Is(err, ErrCompactionBusy) {
		t.Fatal(err)
	}
	publish(t, m, Mutation{ID: "a", Vector: []float32{2, 0}}, Mutation{ID: "b", Delete: true}, Mutation{ID: "c", Vector: []float32{1, 1}})
	rev := m.View().Revision()
	if err := c.Publish(context.Background()); err != nil {
		t.Fatal(err)
	}
	if m.View().Revision() != rev {
		t.Fatal("compaction changed semantic revision")
	}
	expect(t, m.View(), "a", []float32{2, 0})
	expect(t, m.View(), "b", nil)
	expect(t, m.View(), "c", []float32{1, 1})
	expect(t, old, "a", []float32{1, 0})
	expect(t, old, "b", []float32{0, 1})
	if len(m.current.Load().delta) != 3 {
		t.Fatal("lost newer delta")
	}
	c, err = m.PrepareCompaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Publish(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(m.current.Load().delta) != 0 {
		t.Fatal("compaction did not reclaim delta")
	}
	expect(t, m.View(), "b", nil)
}

func TestGenerationBoundsCancellationAndSettlement(t *testing.T) {
	m, _ := NewGenerations(2, 1)
	ctx, cancel := context.WithCancel(context.Background())
	p, err := m.Prepare(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}})
	if err != nil {
		t.Fatal(err)
	}
	wait, stop := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer stop()
	if _, err := m.Prepare(wait, []Mutation{{ID: "b", Vector: []float32{1, 0}}}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	cancel()
	if err := p.Commit(); err != nil {
		t.Fatal("post-durability publication canceled", err)
	}
	if _, err := m.Prepare(context.Background(), []Mutation{{ID: "b", Vector: []float32{1, 0}}}); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	expect(t, m.View(), "b", nil)
	c, err := m.PrepareCompaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Publish(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	c, err = m.PrepareCompaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Publish(context.Background()); err != nil {
		t.Fatal(err)
	}
	publish(t, m, Mutation{ID: "b", Vector: []float32{1, 0}})
}

func TestGenerationConcurrentSnapshotsAndCompaction(t *testing.T) {
	m, _ := NewGenerations(2, 128)
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 1; i <= 1000; i++ {
			p, err := m.Prepare(context.Background(), []Mutation{{ID: "a", Vector: []float32{float32(i), 1}}})
			if err != nil {
				errs <- err
				return
			}
			if err := p.Commit(); err != nil {
				errs <- err
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c, err := m.PrepareCompaction(context.Background())
			if err != nil {
				errs <- err
				return
			}
			if err := c.Publish(context.Background()); err != nil {
				errs <- err
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			v := m.View()
			a, ok := v.Lookup("a")
			b, again := v.Lookup("a")
			if ok != again || !reflect.DeepEqual(a, b) {
				errs <- fmt.Errorf("view mutated")
				return
			}
		}
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	expect(t, m.View(), "a", []float32{1000, 1})
	if m.View().Revision() != 1000 {
		t.Fatal("lost revision")
	}
}

func TestPreparedCommitHasNoAllocations(t *testing.T) {
	m, _ := NewGenerations(2, 1)
	p, err := m.Prepare(context.Background(), []Mutation{{ID: "a", Vector: []float32{1, 0}}})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Commit(); err != nil {
		t.Fatal(err)
	}
	// Reuse the already-materialized object to isolate publication, excluding
	// preparation and test scaffolding allocations. This is not a WAL benchmark.
	allocs := testing.AllocsPerRun(1000, func() {
		<-m.writer
		p.finished.Store(false)
		if p.Commit() != nil {
			panic("commit failed")
		}
	})
	if allocs != 0 {
		t.Fatalf("publication allocations=%v", allocs)
	}
}
