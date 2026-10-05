package researchindex

import (
	"context"
	"errors"
	"math"
	"testing"
)

func deleteWriterFixture(t *testing.T, persist PersistGeneration) *PrivateDeleteWriter {
	t.Helper()
	r := LayeredRecord{ID: "target", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	g, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, []LayeredEdit{{0, &r}}, 1, LayeredLimits{1, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	s, _ := (EntrySummary{}).With(0, 0)
	w, e := NewPrivateDeleteWriter(g, s, 1, 2, persist)
	if e != nil {
		t.Fatal(e)
	}
	return w
}
func TestPrivateDeleteWriterPublication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	w := deleteWriterFixture(t, func(_ context.Context, rev uint64, m []Mutation) error {
		if rev != 2 || len(m) != 1 || m[0].ID != "target" || !m[0].Delete {
			t.Error("durable payload")
		}
		close(entered)
		<-release
		m[0].ID = "mutated"
		cancel()
		return nil
	})
	old, _ := w.View(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Delete(ctx, 0) }()
	<-entered
	blocked, stop := context.WithCancel(context.Background())
	stop()
	if _, e := w.View(blocked); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, ok := old.Lookup(0); !ok {
		t.Fatal("old snapshot changed during persistence")
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	current, e := w.View(context.Background())
	if e != nil || current.Revision() != 2 {
		t.Fatal(e)
	}
	if _, ok := current.Lookup(0); ok {
		t.Fatal("delete unpublished")
	}
	if old.Revision() != 1 {
		t.Fatal("old revision mutated")
	}
	if _, ok := old.Lookup(0); !ok {
		t.Fatal("history lost")
	}
}
func TestPrivateDeleteWriterQuarantine(t *testing.T) {
	for _, panicMode := range []bool{false, true} {
		calls := 0
		w := deleteWriterFixture(t, func(context.Context, uint64, []Mutation) error {
			calls++
			if panicMode {
				panic("lost acknowledgment")
			}
			return errors.New("uncertain commit")
		})
		old, _ := w.View(context.Background())
		if e := w.Delete(context.Background(), 0); e == nil {
			t.Fatal("failed callback accepted")
		}
		if _, e := w.View(context.Background()); !errors.Is(e, ErrRecoveryRequired) {
			t.Fatal(e)
		}
		if e := w.Delete(context.Background(), 0); !errors.Is(e, ErrRecoveryRequired) {
			t.Fatal(e)
		}
		if calls != 1 {
			t.Fatal("retried uncertain transaction")
		}
		if _, ok := old.Lookup(0); !ok {
			t.Fatal("history changed")
		}
	}
}
func TestPrivateDeleteWriterPreflight(t *testing.T) {
	w := deleteWriterFixture(t, func(context.Context, uint64, []Mutation) error { t.Fatal("unexpected persistence"); return nil })
	if e := w.Delete(context.Background(), 1); e == nil {
		t.Fatal("missing target")
	}
	if _, e := w.View(context.Background()); e != nil {
		t.Fatal("preflight quarantined")
	}
	w.current.revision = math.MaxUint64
	if e := w.Delete(context.Background(), 0); e == nil {
		t.Fatal("overflow")
	}
}
