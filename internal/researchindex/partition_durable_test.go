package researchindex

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func partitionID(p int) string {
	for i := 0; ; i++ {
		id := fmt.Sprintf("partition-test-%d", i)
		got, _ := ResearchPartition(id, 2)
		if got == p {
			return id
		}
	}
}
func mustPartitionView(t *testing.T, d *PartitionedDurable) PartitionView {
	t.Helper()
	v, e := d.View(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func requireVector(t *testing.T, v PartitionView, id string, want float32) {
	t.Helper()
	got, ok := v.Lookup(id)
	if !ok || len(got) != 2 || got[0] != want {
		t.Fatal(id, got, ok, want)
	}
}

func TestPartitionDurableAtomicAndHistorical(t *testing.T) {
	ctx := context.Background()
	a, b := partitionID(0), partitionID(1)
	entered, release := make(chan struct{}), make(chan struct{})
	d, e := RestorePartitioned(2, 64, 2, 1, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { close(entered); <-release; return nil })
	if e != nil {
		t.Fatal(e)
	}
	old := mustPartitionView(t, d)
	done := make(chan error, 1)
	go func() {
		done <- d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}, {ID: b, Vector: []float32{2, 0}}})
	}()
	<-entered
	timeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	_, e = d.View(timeout)
	cancel()
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("partial revision exposed", e)
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	now := mustPartitionView(t, d)
	if now.Revision() != 2 || now.DeltaCount() != 2 {
		t.Fatal(now.Revision(), now.DeltaCount())
	}
	for _, id := range []string{a, b} {
		requireVector(t, old, id, 1)
		requireVector(t, now, id, 2)
	}
}

func TestPartitionDurableGlobalCapAndPreflight(t *testing.T) {
	ctx := context.Background()
	calls := 0
	d, e := RestorePartitioned(2, 2, 2, 0, nil, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if e != nil {
		t.Fatal(e)
	}
	a, b := partitionID(0), partitionID(1)
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{1, 0}}}); e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(ctx, []Mutation{{ID: "third", Vector: []float32{1, 0}}}); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	if calls != 1 || mustPartitionView(t, d).DeltaCount() != 2 {
		t.Fatal("budget multiplied by partitions")
	}
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}, {ID: b, Vector: []float32{0, 0}}}); e == nil {
		t.Fatal("invalid vector accepted")
	}
	if calls != 1 {
		t.Fatal("invalid transaction persisted")
	}
	requireVector(t, mustPartitionView(t, d), a, 1)
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}}); e != nil {
		t.Fatal("prepare leaked exclusion", e)
	}
}

func TestPartitionCompactionRebasesAndRetainsOtherShard(t *testing.T) {
	ctx := context.Background()
	a, b := partitionID(0), partitionID(1)
	d, e := RestorePartitioned(2, 8, 2, 1, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}}); e != nil {
		t.Fatal(e)
	}
	old := mustPartitionView(t, d)
	c, e := d.PrepareCompaction(ctx, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := d.PrepareCompaction(ctx, 1); !errors.Is(e, ErrCompactionBusy) {
		t.Fatal("global build bound lost", e)
	}
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{3, 0}}, {ID: b, Delete: true}}); e != nil {
		t.Fatal(e)
	}
	before := mustPartitionView(t, d)
	if e := c.Publish(ctx); e != nil {
		t.Fatal(e)
	}
	now := mustPartitionView(t, d)
	if now.Revision() != 3 || now.views[1].g != before.views[1].g {
		t.Fatal("compaction changed other shard or semantic revision")
	}
	requireVector(t, old, a, 2)
	requireVector(t, now, a, 3)
	if _, ok := now.Lookup(b); ok {
		t.Fatal("tombstone lost")
	}
	if e := c.Publish(ctx); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
}

func TestPartitionUnknownCommitQuarantinesAllAndCompaction(t *testing.T) {
	ctx := context.Background()
	a := partitionID(0)
	d, e := RestorePartitioned(2, 8, 2, 1, []Mutation{{ID: a, Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return errors.New("lost acknowledgement") })
	if e != nil {
		t.Fatal(e)
	}
	old := mustPartitionView(t, d)
	c, e := d.PrepareCompaction(ctx, 0)
	if e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}, {ID: partitionID(1), Vector: []float32{2, 0}}}); !errors.Is(e, ErrRecoveryRequired) {
		t.Fatal(e)
	}
	if _, e := d.View(ctx); !errors.Is(e, ErrRecoveryRequired) {
		t.Fatal(e)
	}
	if e := c.Publish(ctx); !errors.Is(e, ErrRecoveryRequired) {
		t.Fatal(e)
	}
	requireVector(t, old, a, 1)
}

func TestPartitionRevisionAndCancellation(t *testing.T) {
	ctx := context.Background()
	calls := 0
	d, e := RestorePartitioned(2, 8, 2, math.MaxUint64, nil, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); e == nil || calls != 0 {
		t.Fatal("overflow reached persistence")
	}
	c, cancel := context.WithCancel(ctx)
	d, e = RestorePartitioned(2, 8, 2, 0, nil, func(context.Context, uint64, []Mutation) error { cancel(); return nil })
	if e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(c, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); e != nil {
		t.Fatal("durable success hidden by cancellation", e)
	}
	requireVector(t, mustPartitionView(t, d), "a", 1)
}

func TestPartitionRealTransactionLostAckRecovery(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/store.libravdb"
	db, e := libra.Open(libra.WithStoragePath(path))
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = db.Close() }()
	_, e = db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
	if e != nil {
		t.Fatal(e)
	}
	persist := func(c context.Context, rev uint64, rs []Mutation) error {
		e := db.WithTx(c, func(tx libra.Tx) error {
			for _, r := range rs {
				if e := tx.Upsert(c, "records", r.ID, r.Vector, nil); e != nil {
					return e
				}
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(rev)})
		})
		if e != nil {
			return e
		}
		return errors.New("commit succeeded, acknowledgement lost")
	}
	d, e := RestorePartitioned(2, 8, 2, 0, nil, persist)
	if e != nil {
		t.Fatal(e)
	}
	a, b := partitionID(0), partitionID(1)
	if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{2, 0}}, {ID: b, Vector: []float32{3, 0}}}); !errors.Is(e, ErrRecoveryRequired) {
		t.Fatal(e)
	}
	if e := db.Close(); e != nil {
		t.Fatal(e)
	}
	db, e = libra.Open(libra.WithStoragePath(path))
	if e != nil {
		t.Fatal(e)
	}
	col, e := db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	if e != nil {
		t.Fatal(e)
	}
	state, e := db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
	if e != nil {
		t.Fatal(e)
	}
	r, e := state.Get(ctx, "revision")
	if e != nil || r.Metadata["value"] != "1" {
		t.Fatal(r, e)
	}
	var restored []Mutation
	for _, id := range []string{a, b} {
		r, e := col.Get(ctx, id)
		if e != nil {
			t.Fatal(e)
		}
		restored = append(restored, Mutation{ID: id, Vector: r.Vector})
	}
	recovered, e := RestorePartitioned(2, 8, 2, 1, restored, persist)
	if e != nil {
		t.Fatal(e)
	}
	v := mustPartitionView(t, recovered)
	requireVector(t, v, a, 2)
	requireVector(t, v, b, 3)
}

func TestPartitionConcurrentGlobalSnapshots(t *testing.T) {
	ctx := context.Background()
	a, b := partitionID(0), partitionID(1)
	d, err := RestorePartitioned(2, 8, 2, 1, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 2)
	go func() {
		for i := 2; i <= 201; i++ {
			if e := d.Apply(ctx, []Mutation{{ID: a, Vector: []float32{float32(i), 0}}, {ID: b, Vector: []float32{float32(i), 0}}}); e != nil {
				done <- e
				return
			}
		}
		done <- nil
	}()
	go func() {
		for i := 0; i < 200; i++ {
			c, e := d.PrepareCompaction(ctx, i%2)
			if e == nil {
				e = c.Publish(ctx)
			}
			if e != nil {
				done <- e
				return
			}
		}
		done <- nil
	}()
	for i := 0; i < 1000; i++ {
		v := mustPartitionView(t, d)
		x, ok := v.Lookup(a)
		y, ok2 := v.Lookup(b)
		if !ok || !ok2 || x[0] != y[0] || uint64(x[0]) != v.Revision() {
			t.Error("mixed global revision", v.Revision(), x, y)
		}
	}
	for i := 0; i < 2; i++ {
		if e := <-done; e != nil {
			t.Error(e)
		}
	}
	if v := mustPartitionView(t, d); v.Revision() != 201 {
		t.Fatal(v.Revision())
	}
}

func TestPartitionCallbackPanicQuarantines(t *testing.T) {
	d, err := RestorePartitioned(2, 8, 2, 0, nil, func(context.Context, uint64, []Mutation) error { panic("uncertain callback") })
	if err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected callback panic")
			}
		}()
		_ = d.Apply(context.Background(), []Mutation{{ID: partitionID(0), Vector: []float32{1, 0}}, {ID: partitionID(1), Vector: []float32{1, 0}}})
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, e := d.View(ctx); !errors.Is(e, ErrRecoveryRequired) {
		t.Fatal("panic did not quarantine/release", e)
	}
}
