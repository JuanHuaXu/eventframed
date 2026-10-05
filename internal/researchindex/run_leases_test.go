package researchindex

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func leaseFixture(t *testing.T) (*ImmutableRun, RunSnapshot) {
	t.Helper()
	ctx := context.Background()
	r, err := BuildImmutableRun(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}, 2, filepath.Join(t.TempDir(), "run"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	w, err := NewDurableRunWriter(ctx, []*ImmutableRun{r}, 1, 2, 64, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	s, err := w.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return r, s
}
func TestRunLeasesRetirementAndDrain(t *testing.T) {
	r, s := leaseFixture(t)
	p, err := NewRunLeasePool(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	p.closeRun = func(r *ImmutableRun) error { close(entered); <-release; return r.Close() }
	l, err := p.Acquire(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Acquire(s); !errors.Is(err, ErrServingBusy) {
		t.Fatal(err)
	}
	if err = p.Retire([]*ImmutableRun{r}); err != nil {
		t.Fatal(err)
	}
	got, err := l.Search(context.Background(), []float32{1, 0})
	if err != nil || len(got) != 1 {
		t.Fatal(got, err)
	}
	if err = p.Close(); !errors.Is(err, ErrServingBusy) {
		t.Fatal("closed active lease", err)
	}
	if err = l.Release(); err != nil {
		t.Fatal(err)
	}
	<-entered
	if l.snapshot.search != nil || l.pool != nil {
		t.Fatal("released snapshot retained")
	}
	if _, err = p.Acquire(s); !errors.Is(err, ErrServingBusy) {
		t.Fatal("retired snapshot reacquired", err)
	}
	other, _ := leaseFixture(t)
	if err = p.Retire([]*ImmutableRun{other}); !errors.Is(err, ErrCapacity) {
		t.Fatal("closing handle uncharged", err)
	}
	done := make(chan error, 1)
	go func() { done <- p.Close() }()
	select {
	case err := <-done:
		t.Fatal("drain returned before close", err)
	default:
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if _, err = l.Search(context.Background(), []float32{1, 0}); !errors.Is(err, ErrServingClosed) {
		t.Fatal(err)
	}
	if err = l.Release(); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
}
func TestRunLeasesFailedCloseRemainsCharged(t *testing.T) {
	r, _ := leaseFixture(t)
	p, _ := NewRunLeasePool(2, 1)
	p.closeRun = func(r *ImmutableRun) error {
		if err := r.Close(); err != nil {
			return err
		}
		return errors.New("injected cleanup failure")
	}
	if err := p.Retire([]*ImmutableRun{r}); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err == nil {
		t.Fatal("failed close hidden")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.retired) != 1 {
		t.Fatal("failed resource forgotten")
	}
}
