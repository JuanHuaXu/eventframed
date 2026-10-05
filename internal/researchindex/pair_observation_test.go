//go:build research_pair_observation

package researchindex

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestBuildPairObserverEndsBeforeServing(t *testing.T) {
	var calls, finishes atomic.Uint64
	ctx := libra.ResearchObserveBuildPairs(context.Background(), func(n int) (func(uint32, uint32), func()) {
		var ended atomic.Bool
		return func(a, b uint32) {
			if ended.Load() {
				t.Error("observer invoked after build ended")
			}
			calls.Add(1)
		}, func() { ended.Store(true); finishes.Add(1) }
	})
	db, err := libra.Open(libra.WithStoragePath(t.TempDir() + "/db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	c, err := db.EnsureCollection(ctx, "base", 8, libra.WithHNSW(16, 200, 100), libra.WithMetric(libra.CosineDistance))
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(65537))
	err = db.WithTx(ctx, func(tx libra.Tx) error {
		for i := 0; i < 600; i++ {
			v := make([]float32, 8)
			for j := range v {
				v[j] = rng.Float32() - .5
			}
			if err := tx.Insert(ctx, "base", fmt.Sprint(i), v, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	before := calls.Load()
	if before == 0 || finishes.Load() == 0 {
		t.Fatal("build was not instrumented")
	}
	if _, err := c.Search(ctx, []float32{1, 0, 0, 0, 0, 0, 0, 0}, 10); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != before {
		t.Fatal("observer retained on serving graph")
	}
}
