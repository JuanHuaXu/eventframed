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

func TestPrivateGraphRealRecovery(t *testing.T) {
	for _, mode := range []string{"success", "rollback", "lost_ack", "panic_after_commit"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "authority")
			db, e := libra.Open(libra.WithStoragePath(path))
			if e != nil {
				t.Fatal(e)
			}
			defer func() { _ = db.Close() }()
			if _, e = db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance)); e != nil {
				t.Fatal(e)
			}
			if _, e = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly()); e != nil {
				t.Fatal(e)
			}
			if e = db.WithTx(ctx, func(tx libra.Tx) error {
				return tx.Upsert(ctx, "state", "revision", nil, map[string]interface{}{"value": "0"})
			}); e != nil {
				t.Fatal(e)
			}
			calls := 0
			persist := func(c context.Context, rev uint64, ms []Mutation) error {
				calls++
				err := db.WithTx(c, func(tx libra.Tx) error {
					for _, m := range ms {
						var e error
						if m.Delete {
							e = tx.Delete(c, "records", m.ID)
						} else {
							e = tx.Upsert(c, "records", m.ID, m.Vector, nil)
						}
						if e != nil {
							return e
						}
					}
					if mode == "rollback" {
						return errors.New("abort")
					}
					return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(rev)})
				})
				if err != nil {
					return err
				}
				if mode == "lost_ack" {
					return errors.New("lost ack")
				}
				if mode == "panic_after_commit" {
					panic("after commit")
				}
				return nil
			}
			w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, persist)
			if e != nil {
				t.Fatal(e)
			}
			old, e := w.Acquire(ctx)
			if e != nil {
				t.Fatal(e)
			}
			_, err := w.Insert(ctx, "a", []float32{1, 0}, 0)
			if (err == nil) != (mode == "success") {
				t.Fatal(err)
			}
			if mode != "success" {
				if _, e = w.Acquire(ctx); !errors.Is(e, ErrRecoveryRequired) {
					t.Fatal(e)
				}
			} else {
				if _, e = w.Insert(ctx, "b", []float32{0, 1}, 0); e != nil {
					t.Fatal(e)
				}
				if _, e = w.Delete(ctx, "a"); e != nil {
					t.Fatal(e)
				}
			}
			if rev, _ := old.Revision(); rev != 0 {
				t.Fatal("old root changed")
			}
			old.Release()
			wantCalls := 1
			if mode == "success" {
				wantCalls = 3
			}
			if calls != wantCalls {
				t.Fatal("unexpected calls", calls)
			}
			if e = db.Close(); e != nil {
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
			wantRev := uint64(1)
			wantCount := 1
			id := "a"
			v := []float32{1, 0}
			if mode == "rollback" {
				wantRev = 0
				wantCount = 0
			}
			if mode == "success" {
				wantRev = 3
				id = "b"
				v = []float32{0, 1}
			}
			count, e := col.Count(ctx)
			if e != nil || count != wantCount {
				t.Fatal(count, e)
			}
			rev, e := state.Get(ctx, "revision")
			if e != nil || rev.Metadata["value"] != fmt.Sprint(wantRev) {
				t.Fatal(rev, e)
			}
			var graph LayeredSnapshot
			if wantCount > 0 {
				r, e := col.Get(ctx, id)
				if e != nil || !reflect.DeepEqual(r.Vector, v) {
					t.Fatal(r, e)
				}
				graph, _, e = PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &LayeredRecord{ID: id, Vector: r.Vector, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}}}, 1, LayeredLimits{1, 2, 1, 8})
				if e != nil {
					t.Fatal(e)
				}
			}
			// This known-ID fixture tests constructor rebuilding, not a corpus enumerator.
			recovered, e := NewPrivateGraphWriter(graph, wantRev, 2, 4, 32, 4, 2, persist)
			if e != nil {
				t.Fatal(e)
			}
			lease, e := recovered.Acquire(ctx)
			if e != nil {
				t.Fatal(e)
			}
			defer lease.Release()
			if got, e := lease.Revision(); e != nil || got != wantRev {
				t.Fatal(got, e)
			}
		})
	}
}
