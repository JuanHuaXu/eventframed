//go:build research_candidate_only

package researchindex

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestCandidateOnlySameGraphAndPublicPayload(t *testing.T) {
	ctx := context.Background()
	db, e := libra.Open(libra.WithStoragePath(t.TempDir() + "/store"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	c, e := db.EnsureCollection(ctx, "base", 8, libra.WithHNSW(16, 200, 100), libra.WithMetric(libra.CosineDistance))
	if e != nil {
		t.Fatal(e)
	}
	rng := rand.New(rand.NewSource(104729))
	vectors := make([][]float32, 600)
	for i := range vectors {
		v := make([]float32, 8)
		for j := range v {
			v[j] = rng.Float32() - .5
		}
		vectors[i] = v
	}
	e = db.WithTx(ctx, func(tx libra.Tx) error {
		for i, v := range vectors {
			if e := tx.Insert(ctx, "base", fmt.Sprint(i), v, map[string]interface{}{"public": "fixture"}); e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 128; i++ {
		q := vectors[i]
		if i%2 == 1 {
			q = make([]float32, 8)
			for j := range q {
				q[j] = rng.Float32() - .5
			}
		}
		full, e := c.Search(ctx, q, 20)
		if e != nil {
			t.Fatal(e)
		}
		lean, e := c.ResearchImmutableCandidates(ctx, q, 20)
		if e != nil {
			t.Fatal(e)
		}
		if len(full.Results) != len(lean.Results) {
			t.Fatal("candidate count changed")
		}
		for j, a := range full.Results {
			b := lean.Results[j]
			if a.ID != b.ID || a.Score != b.Score {
				t.Fatal("candidate semantics changed", i, j, a, b)
			}
			if len(a.Vector) != 8 || a.Metadata["public"] != "fixture" || a.Version == 0 {
				t.Fatal("public payload changed", a)
			}
			if b.Vector != nil || b.Metadata != nil || b.Version != 0 {
				t.Fatal("candidate payload unexpectedly hydrated", b)
			}
		}
	}
	flat, e := db.EnsureCollection(ctx, "flat", 8, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	if e != nil {
		t.Fatal(e)
	}
	if _, e := flat.ResearchImmutableCandidates(ctx, vectors[0], 1); e == nil {
		t.Fatal("unsupported index accepted")
	}
	if _, e := c.ResearchImmutableCandidates(ctx, []float32{1}, 1); e == nil {
		t.Fatal("dimension error lost")
	}
	if _, e := c.ResearchImmutableCandidates(ctx, vectors[0], 0); e == nil {
		t.Fatal("invalid k accepted")
	}
	if e := db.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := c.ResearchImmutableCandidates(ctx, vectors[0], 1); e == nil {
		t.Fatal("closed collection accepted")
	}
}
