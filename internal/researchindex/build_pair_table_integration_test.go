//go:build research_pair_table

package researchindex

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestBuildPairTableRealDistanceAndLifetime(t *testing.T) {
	var calls, hits atomic.Uint64
	ctx := libra.ResearchTableBuildPairs(context.Background(), func(n int) (interface {
		Lookup(uint32, uint32) (float32, bool)
		Store(uint32, uint32, float32)
	}, func()) {
		table, err := NewBuildPairTable(65536)
		if err != nil {
			t.Fatal(err)
		}
		observed := &checkedPairTable{BuildPairTable: table, t: t, calls: &calls}
		return observed, func() { observed.ended.Store(true); hits.Add(table.Stats().Hits) }
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

type checkedPairTable struct {
	*BuildPairTable
	t     *testing.T
	calls *atomic.Uint64
	ended atomic.Bool
}

func (c *checkedPairTable) Lookup(a, b uint32) (float32, bool) {
	if c.ended.Load() {
		c.t.Error("lookup after build")
	}
	c.calls.Add(1)
	return c.BuildPairTable.Lookup(a, b)
}
func (c *checkedPairTable) Store(a, b uint32, v float32) {
	if c.ended.Load() {
		c.t.Error("store after build")
	}
	c.BuildPairTable.Store(a, b, v)
}
