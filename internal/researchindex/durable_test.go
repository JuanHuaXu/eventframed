package researchindex

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestDurableGenerationRealCommitAndAmbiguousErrors(t *testing.T) {
	for _, fault := range []string{"before", "staged", "after", "cancel-after", "none"} {
		t.Run(fault, func(t *testing.T) {
			ctx := context.Background()
			path := t.TempDir() + "/generation.libravdb"
			db, err := libra.Open(libra.WithStoragePath(path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.EnsureCollection(ctx, "vectors", 2, libra.WithFlat())
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
			if err != nil {
				t.Fatal(err)
			}
			activeFault := "none"
			lost := errors.New("injected lost acknowledgement")
			c, cancel := context.WithCancel(ctx)
			defer cancel()
			persist := func(ctx context.Context, revision uint64, changes []Mutation) error {
				if activeFault == "before" {
					return lost
				}
				err := db.WithTx(ctx, func(tx libra.Tx) error {
					for _, m := range changes {
						if m.Delete {
							if err := tx.Delete(ctx, "vectors", m.ID); err != nil {
								return err
							}
						} else {
							if err := tx.Upsert(ctx, "vectors", m.ID, m.Vector, map[string]interface{}{"public": "Mars landing fixture"}); err != nil {
								return err
							}
						}
					}
					if activeFault == "staged" {
						return lost
					}
					return tx.Upsert(ctx, "state", "revision", nil, map[string]interface{}{"revision": strconv.FormatUint(revision, 10)})
				})
				if err != nil {
					return err
				}
				if activeFault == "after" {
					return lost
				}
				if activeFault == "cancel-after" {
					cancel()
				}
				return nil
			}
			d, err := RestoreDurable(2, 8, 0, nil, persist)
			if err != nil {
				t.Fatal(err)
			}
			if err = d.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
				t.Fatal(err)
			}
			old, err := d.View(ctx)
			if err != nil {
				t.Fatal(err)
			}
			activeFault = fault
			err = d.Apply(c, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{0, 1}}})
			if fault == "before" || fault == "staged" || fault == "after" {
				if !errors.Is(err, ErrRecoveryRequired) || !errors.Is(err, lost) {
					t.Fatal(err)
				}
				if _, err := d.View(ctx); !errors.Is(err, ErrRecoveryRequired) {
					t.Fatal("read escaped quarantine", err)
				}
				if err := d.Apply(ctx, []Mutation{{ID: "c", Vector: []float32{1, 1}}}); !errors.Is(err, ErrRecoveryRequired) {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				v, err := d.View(ctx)
				if err != nil {
					t.Fatal(err)
				}
				expect(t, v, "a", nil)
				expect(t, v, "b", []float32{0, 1})
			}
			expect(t, old, "a", []float32{1, 0})
			expect(t, old, "b", nil)
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = libra.Open(libra.WithStoragePath(path))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			vectors, err := db.GetCollection("vectors")
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
			if err != nil {
				t.Fatal(err)
			}
			var recovered []Mutation
			for _, id := range []string{"a", "b", "c"} {
				r, err := vectors.Get(ctx, id)
				if errors.Is(err, libra.ErrRecordNotFound) {
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				recovered = append(recovered, Mutation{ID: id, Vector: append([]float32(nil), r.Vector...)})
			}
			activeFault = "none"
			restored, err := RestoreDurable(2, 8, revision, recovered, persist)
			if err != nil {
				t.Fatal(err)
			}
			v, err := restored.View(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if fault == "before" || fault == "staged" {
				if revision != 1 {
					t.Fatal(revision)
				}
				expect(t, v, "a", []float32{1, 0})
				expect(t, v, "b", nil)
			} else {
				if revision != 2 {
					t.Fatal(revision)
				}
				expect(t, v, "a", nil)
				expect(t, v, "b", []float32{0, 1})
			}
			if err := restored.Apply(ctx, []Mutation{{ID: "c", Vector: []float32{1, 1}}}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDurableViewWaitAndPanicQuarantine(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	d, _ := RestoreDurable(2, 8, 0, nil, func(context.Context, uint64, []Mutation) error { close(entered); <-release; return nil })
	done := make(chan error, 1)
	go func() { done <- d.Apply(context.Background(), []Mutation{{ID: "a", Vector: []float32{1, 0}}}) }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := d.View(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	d, _ = RestoreDurable(2, 8, 0, nil, func(context.Context, uint64, []Mutation) error { panic("injected") })
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()
		_ = d.Apply(context.Background(), []Mutation{{ID: "a", Vector: []float32{1, 0}}})
	}()
	if _, err := d.View(context.Background()); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
}
