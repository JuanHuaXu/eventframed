package researchindex

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestDurableCompactionCannotPublishAfterUnknownCommit(t *testing.T) {
	ctx := context.Background()
	fail := false
	d, _ := RestoreDurable(2, 4, 0, nil, func(context.Context, uint64, []Mutation) error {
		if fail {
			return errors.New("unknown commit")
		}
		return nil
	})
	if err := d.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	old, _ := d.View(ctx)
	c, err := d.PrepareCompaction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fail = true
	if err := d.Apply(ctx, []Mutation{{ID: "b", Vector: []float32{0, 1}}}); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	before := d.core.current.Load()
	if err := c.Publish(ctx); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal("quarantined publication accepted", err)
	}
	if d.core.current.Load() != before || d.core.compacting.Load() {
		t.Fatal("candidate published or lease leaked")
	}
	if _, err := d.PrepareCompaction(ctx); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if _, err := d.View(ctx); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	expect(t, old, "a", []float32{1, 0})
}

func TestDurableCompactionCarriesNewerWrites(t *testing.T) {
	ctx := context.Background()
	calls := 0
	d, _ := RestoreDurable(2, 4, 0, nil, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if err := d.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	c, err := d.PrepareCompaction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{0, 1}}}); err != nil {
		t.Fatal(err)
	}
	if err := c.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	v, err := d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect(t, v, "a", nil)
	expect(t, v, "b", []float32{0, 1})
	if calls != 2 || v.Revision() != 2 {
		t.Fatal("compaction wrote semantic state")
	}
	if err := c.Abort(); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
}

func TestDurableCompactionCapacityAndCanceledPublication(t *testing.T) {
	ctx := context.Background()
	calls := 0
	d, _ := RestoreDurable(2, 2, 0, nil, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	for i := 0; i < 20; i++ {
		mutation := []Mutation{{ID: fmt.Sprint(i), Vector: []float32{1, float32(i)}}}
		err := d.Apply(ctx, mutation)
		if errors.Is(err, ErrCapacity) {
			before := calls
			c, e := d.PrepareCompaction(ctx)
			if e != nil {
				t.Fatal(e)
			}
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel()
			if e := c.Publish(cancelCtx); !errors.Is(e, context.Canceled) {
				t.Fatal(e)
			}
			c, e = d.PrepareCompaction(ctx)
			if e != nil {
				t.Fatal(e)
			}
			if e := c.Publish(ctx); e != nil {
				t.Fatal(e)
			}
			if calls != before {
				t.Fatal("compaction invoked persist")
			}
			err = d.Apply(ctx, mutation)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	v, err := d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 20 || v.Revision() != 20 {
		t.Fatal("rejected write reached persistence")
	}
	for i := 0; i < 20; i++ {
		expect(t, v, fmt.Sprint(i), []float32{1, float32(i)})
	}
}
