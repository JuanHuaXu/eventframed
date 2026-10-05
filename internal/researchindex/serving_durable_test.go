package researchindex

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestServingRealPersistenceCompactionAndRecovery(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := root + "/authoritative.libravdb"
	db, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
	if err != nil {
		t.Fatal(err)
	}
	persist := func(c context.Context, revision uint64, changes []Mutation) error {
		return db.WithTx(c, func(tx libra.Tx) error {
			for _, m := range changes {
				if m.Delete {
					if err := tx.Delete(c, "records", m.ID); err != nil {
						return err
					}
				} else {
					if err := tx.Upsert(c, "records", m.ID, m.Vector, map[string]interface{}{"public": "Mars landing fixture"}); err != nil {
						return err
					}
				}
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"revision": strconv.FormatUint(revision, 10)})
		})
	}
	d, err := RestoreDurable(2, 4, 0, nil, persist)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewServingBounded(ctx, d, root+"/initial", 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	var old *ReadLease
	compactions := 0
	apply := func(mutations []Mutation) {
		t.Helper()
		err := s.Apply(ctx, mutations)
		if errors.Is(err, ErrCapacity) {
			compactions++
			if err := s.Compact(ctx, fmt.Sprintf("%s/base-%d", root, compactions)); err != nil {
				t.Fatal(err)
			}
			err = s.Apply(ctx, mutations)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 12; i++ {
		id := fmt.Sprint(i)
		vector := []float32{float32(i + 1), 1}
		apply([]Mutation{{ID: id, Vector: vector}})
		lease, err := s.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		got, err := lease.Search(ctx, vector, 1)
		if err != nil || len(got) != 1 || got[0].ID != id {
			t.Fatalf("self query%d: %v %v", i, got, err)
		}
		if i == 0 {
			old = lease
		} else {
			if err := lease.Release(); err != nil {
				t.Fatal(err)
			}
		}
	}
	apply([]Mutation{{ID: "0", Delete: true}, {ID: "1", Vector: []float32{-1, 0}}})
	lease, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := lease.Search(ctx, []float32{-1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "1" {
		t.Fatal(got, err)
	}
	if lease.Revision() != 13 {
		t.Fatal("incorrect acknowledged revision")
	}
	if err := lease.Release(); err != nil {
		t.Fatal(err)
	}
	got, err = old.Search(ctx, []float32{1, 1}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "0" {
		t.Fatal("old reader changed", got, err)
	}
	if err := old.Release(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	records, err := db.GetCollection("records")
	if err != nil {
		t.Fatal(err)
	}
	state, err := db.GetCollection("state")
	if err != nil {
		t.Fatal(err)
	}
	r, err := state.Get(ctx, "revision")
	if err != nil {
		t.Fatal(err)
	}
	revision, err := strconv.ParseUint(r.Metadata["revision"].(string), 10, 64)
	if err != nil || revision != 13 {
		t.Fatal(revision, err)
	}
	var restoredRecords []Mutation
	for i := 0; i < 12; i++ {
		id := fmt.Sprint(i)
		r, err := records.Get(ctx, id)
		if i == 0 {
			if !errors.Is(err, libra.ErrRecordNotFound) {
				t.Fatal("deleted record returned", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		restoredRecords = append(restoredRecords, Mutation{ID: id, Vector: r.Vector})
	}
	d, err = RestoreDurable(2, 4, revision, restoredRecords, persist)
	if err != nil {
		t.Fatal(err)
	}
	s, err = NewServingBounded(ctx, d, root+"/recovered", 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	lease, err = s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err = lease.Search(ctx, []float32{-1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "1" {
		t.Fatal("recovered ANN mismatch", got, err)
	}
	if lease.Revision() != 13 {
		t.Fatal("recovered revision mismatch")
	}
	if err := lease.Release(); err != nil {
		t.Fatal(err)
	}
	if compactions < 2 {
		t.Fatal("fixture did not exercise compaction")
	}
}
