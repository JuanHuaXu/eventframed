package researchindex

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPrivateGraphBatch(t *testing.T) {
	ctx := context.Background()
	calls := 0
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 1, func(_ context.Context, rev uint64, ms []Mutation) error {
		calls++
		if rev != 4 || len(ms) != 4 {
			t.Fatal(rev, len(ms))
		}
		ms[0].Vector[0] = 99
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	old, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	rs := []PrivateInsertRequest{{"a", []float32{1, 0}, 0}, {"b", []float32{0, 1}, 0}, {"c", []float32{1, 1}, 1}, {"d", []float32{-1, 1}, 0}}
	if _, e = w.InsertBatch(ctx, rs); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal("not one commit")
	}
	if rev, _ := old.Revision(); rev != 0 {
		t.Fatal("old revision")
	}
	old.Release()
	got, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer got.Release()
	if rev, _ := got.Revision(); rev != 4 {
		t.Fatal(rev)
	}
	serial, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if e != nil {
		t.Fatal(e)
	}
	for _, r := range rs {
		if _, e = serial.Insert(ctx, r.ID, r.Vector, r.Level); e != nil {
			t.Fatal(e)
		}
	}
	control, e := serial.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer control.Release()
	for i := range rs {
		a, _, _ := got.Lookup(uint32(i))
		b, _, _ := control.Lookup(uint32(i))
		if !reflect.DeepEqual(a, b) {
			t.Fatal("batch differs from serial", i, a, b)
		}
	}
}

func TestPrivateGraphBatchFailure(t *testing.T) {
	for _, mode := range []string{"prepare", "duplicate", "error", "panic"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 1, func(context.Context, uint64, []Mutation) error {
				calls++
				if mode == "panic" {
					panic("lost ack")
				}
				return errors.New("lost ack")
			})
			if e != nil {
				t.Fatal(e)
			}
			rs := []PrivateInsertRequest{{"a", []float32{1, 0}, 0}, {"b", []float32{0, 1}, 0}}
			if mode == "prepare" {
				rs[1].Vector = []float32{0, 0}
			}
			if mode == "duplicate" {
				rs[1].ID = "a"
			}
			if _, e = w.InsertBatch(context.Background(), rs); e == nil {
				t.Fatal("failure ignored")
			}
			lease, e := w.Acquire(context.Background())
			if mode == "error" || mode == "panic" {
				if !errors.Is(e, ErrRecoveryRequired) || calls != 1 {
					t.Fatal(e, calls)
				}
			} else {
				if e != nil || calls != 0 || len(w.ids) != 0 {
					t.Fatal(e, calls)
				}
				if rev, _ := lease.Revision(); rev != 0 {
					t.Fatal(rev)
				}
				lease.Release()
			}
		})
	}
}
