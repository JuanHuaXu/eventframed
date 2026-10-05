//go:build research_pair_cache

package researchindex

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestBuildPairCacheRealDistanceAndLifetime(t *testing.T) {
	var calls, hits atomic.Uint64
	ctx := libra.ResearchCacheBuildPairs(context.Background(), func(n int) (func(uint32, uint32, func() float32) float32, func()) {
		cache, err := NewBuildPairCache(65536)
		if err != nil {
			t.Fatal(err)
		}
		var ended atomic.Bool
		return func(a, b uint32, compute func() float32) float32 {
			if ended.Load() {
				t.Error("cache used after build")
			}
			calls.Add(1)
			want := compute()
			got := cache.Distance(a, b, compute)
			if math.Float32bits(want) != math.Float32bits(got) {
				t.Error("memo changed real distance", a, b, want, got)
			}
			return got
		}, func() { ended.Store(true); hits.Add(cache.Stats().Hits) }
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
	if before == 0 || hits.Load() == 0 {
		t.Fatal("real cache not exercised")
	}
	if _, err := c.Search(ctx, []float32{1, 0, 0, 0, 0, 0, 0, 0}, 10); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != before {
		t.Fatal("cache retained for serving")
	}
}
