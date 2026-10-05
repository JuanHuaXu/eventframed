package researchindex

import (
	"context"
	"errors"
	"fmt"
	libra "github.com/xDarkicex/libravdb/libravdb"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunWriterRealCommitLostAckRecovery(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "authority")
	db, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err = db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly()); err != nil {
		t.Fatal(err)
	}
	persist := func(c context.Context, rev uint64, ms []Mutation) error {
		err := db.WithTx(c, func(tx libra.Tx) error {
			for _, m := range ms {
				if err := tx.Upsert(c, "records", m.ID, m.Vector, nil); err != nil {
					return err
				}
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(rev)})
		})
		if err != nil {
			return err
		}
		return errors.New("committed but acknowledgement lost")
	}
	w, err := NewDurableRunWriter(ctx, nil, 0, 2, 64, 2, persist)
	if err != nil {
		t.Fatal(err)
	}
	writes := []Mutation{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{0, 1}}}
	if err = w.Apply(ctx, writes); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if _, err = w.Snapshot(ctx); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal("uncertain state served", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	col, err := db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	if err != nil {
		t.Fatal(err)
	}
	state, err := db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
	if err != nil {
		t.Fatal(err)
	}
	revision, err := state.Get(ctx, "revision")
	if err != nil || revision.Metadata["value"] != "1" {
		t.Fatal(revision, err)
	}
	var recovered []Mutation
	for _, m := range writes {
		r, err := col.Get(ctx, m.ID)
		if err != nil || !reflect.DeepEqual(r.Vector, m.Vector) {
			t.Fatal(r, err)
		}
		recovered = append(recovered, Mutation{ID: m.ID, Vector: r.Vector})
	}
	run, err := BuildImmutableRun(ctx, recovered, 2, filepath.Join(root, "recovered-run"))
	if err != nil {
		t.Fatal(err)
	}
	defer run.Close()
	fresh, err := NewDurableRunWriter(ctx, []*ImmutableRun{run}, 1, 2, 64, 2, persist)
	if err != nil {
		t.Fatal(err)
	}
	s, err := fresh.Snapshot(ctx)
	if err != nil || s.Revision() != 1 {
		t.Fatal(s, err)
	}
	got, err := s.Search(ctx, []float32{1, 0})
	if err != nil || !reflect.DeepEqual(got, []Candidate{{ID: "a", Score: 1}, {ID: "b", Score: 0}}) {
		t.Fatal(got, err)
	}
}
