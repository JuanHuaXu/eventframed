package researchindex

import (
	"context"
	"errors"
	"testing"
)

func TestServingLeaseBoundAndReleasedOwnership(t *testing.T) {
	ctx := context.Background()
	d, _ := RestoreDurable(2, 8, 0, nil, func(context.Context, uint64, []Mutation) error { return nil })
	s, err := NewServingBounded(ctx, d, t.TempDir()+"/base", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Compact(ctx, t.TempDir()+"/next"); err != nil {
		t.Fatal(err)
	}
	b, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ctx); !errors.Is(err, ErrServingBusy) {
		t.Fatal("lease cap exceeded", err)
	}
	revision := a.Revision()
	if err := a.Release(); err != nil {
		t.Fatal(err)
	}
	if a.owner != nil || a.handle != nil || a.view.g != nil || a.Revision() != revision {
		t.Fatal("released lease retains resources or loses identity")
	}
	c, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Release(); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ctx); !errors.Is(err, ErrServingBusy) {
		t.Fatal("double release granted capacity", err)
	}
	if err := b.Release(); err != nil {
		t.Fatal(err)
	}
	if err := c.Release(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}
