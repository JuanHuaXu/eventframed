package researchindex

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestPrivateDeleteRealRecovery(t *testing.T) {
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
			vectors := [][]float32{{1, 0}, {0, 1}}
			ids := []string{"target", "survivor"}
			if e = db.WithTx(ctx, func(tx libra.Tx) error {
				for i, id := range ids {
					if e := tx.Upsert(ctx, "records", id, vectors[i], nil); e != nil {
						return e
					}
				}
				return tx.Upsert(ctx, "state", "revision", nil, map[string]interface{}{"value": "1"})
			}); e != nil {
				t.Fatal(e)
			}
			var initial []LayeredEdit
			var summary EntrySummary
			for i, id := range ids {
				r := &LayeredRecord{ID: id, Vector: vectors[i], Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
				initial = append(initial, LayeredEdit{uint32(i), r})
				summary, _ = summary.With(uint32(i), 0)
			}
			graph, _, e := PrepareLayered(ctx, LayeredSnapshot{}, initial, 1, LayeredLimits{2, 2, 1, 8})
			if e != nil {
				t.Fatal(e)
			}
			calls := 0
			persist := func(c context.Context, rev uint64, ms []Mutation) error {
				calls++
				err := db.WithTx(c, func(tx libra.Tx) error {
					for _, m := range ms {
						if !m.Delete {
							return errors.New("unexpected upsert")
						}
						if e := tx.Delete(c, "records", m.ID); e != nil {
							return e
						}
					}
					if mode == "rollback" {
						return errors.New("injected transaction abort")
					}
					return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(rev)})
				})
				if err != nil {
					return err
				}
				if mode == "lost_ack" {
					return errors.New("committed acknowledgement lost")
				}
				if mode == "panic_after_commit" {
					panic("committed then panic")
				}
				return nil
			}
			w, e := NewPrivateDeleteWriter(graph, summary, 1, 2, persist)
			if e != nil {
				t.Fatal(e)
			}
			old, _ := w.View(ctx)
			err := w.Delete(ctx, 0)
			if (err == nil) != (mode == "success") {
				t.Fatal("unexpected outcome", err)
			}
			if mode != "success" {
				if _, e = w.View(ctx); !errors.Is(e, ErrRecoveryRequired) {
					t.Fatal("uncertain writer served", e)
				}
				if e = w.Delete(ctx, 0); !errors.Is(e, ErrRecoveryRequired) {
					t.Fatal(e)
				}
			} else {
				v, e := w.View(ctx)
				if e != nil || v.Revision() != 2 {
					t.Fatal(e)
				}
				if _, ok := v.Lookup(0); ok {
					t.Fatal("target visible")
				}
			}
			if calls != 1 {
				t.Fatal("uncertain retry")
			}
			if _, ok := old.Lookup(0); !ok {
				t.Fatal("historical target lost")
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
			wantCount, wantRev := 1, uint64(2)
			if mode == "rollback" {
				wantCount, wantRev = 2, 1
			}
			count, e := col.Count(ctx)
			if e != nil || count != wantCount {
				t.Fatal("reopened count", count, e)
			}
			rev, e := state.Get(ctx, "revision")
			if e != nil || rev.Metadata["value"] != fmt.Sprint(wantRev) {
				t.Fatal("reopened revision", rev, e)
			}
			var restored []LayeredEdit
			var restoredSummary EntrySummary
			for i, id := range ids {
				r, e := col.Get(ctx, id)
				if i == 0 && mode != "rollback" {
					if e == nil {
						t.Fatal("durable deletion absent")
					}
					continue
				}
				if e != nil || !reflect.DeepEqual(r.Vector, vectors[i]) {
					t.Fatal("survivor data", r, e)
				}
				restored = append(restored, LayeredEdit{uint32(i), &LayeredRecord{ID: id, Vector: r.Vector, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}})
				restoredSummary, _ = restoredSummary.With(uint32(i), 0)
			}
			entry, level, _ := restoredSummary.Best()
			rebuilt, _, e := PrepareLayered(ctx, LayeredSnapshot{}, restored, uint64(entry)<<32|uint64(level+1), LayeredLimits{2, 2, 1, 8})
			if e != nil {
				t.Fatal(e)
			}
			fresh, e := NewPrivateDeleteWriter(rebuilt, restoredSummary, wantRev, 2, persist)
			if e != nil {
				t.Fatal(e)
			}
			v, e := fresh.View(ctx)
			if e != nil || v.Revision() != wantRev {
				t.Fatal(e)
			}
			if _, ok := v.Lookup(1); !ok {
				t.Fatal("recovered survivor absent")
			}
		})
	}
}
