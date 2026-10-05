package researchindex

import (
	"context"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

// Exact dot-product scorer isolates lifecycle/merge integrity from ANN quality.
func viewTop(v View, query []float32, k int) ([]Candidate, error) {
	score := func(vector []float32) float64 {
		var x float64
		for i, v := range vector {
			x += float64(v) * float64(query[i])
		}
		return x
	}
	base := map[string]Candidate{}
	for id, r := range v.g.base.records {
		base[id] = Candidate{id, score(r.vector)}
	}
	prefix := ordered(base)
	prefix = prefix[:min(len(prefix), k+len(v.g.delta))]
	var delta []Change
	for id, r := range v.g.delta {
		delta = append(delta, Change{Candidate{id, score(r.vector)}, r.deleted})
	}
	return Merge(prefix, delta, k)
}

func TestGenerationMergeMatchesCommittedHistory(t *testing.T) {
	m, _ := NewGenerations(2, 128)
	ctx := context.Background()
	rng := rand.New(rand.NewPCG(113, 557))
	truth := map[string][]float32{}
	var building *Compaction
	type saved struct {
		view View
		top  []Candidate
	}
	var savedViews []saved
	query := []float32{2, -1}
	for step := 0; step < 600; step++ {
		if building == nil && step%3 == 0 {
			var err error
			building, err = m.PrepareCompaction(ctx)
			if err != nil {
				t.Fatal(err)
			}
		}
		id := fmt.Sprint(rng.IntN(40))
		deleted := rng.IntN(4) == 0
		vector := []float32{float32(1 + rng.IntN(20)), float32(rng.IntN(20))}
		p, err := m.Prepare(ctx, []Mutation{{ID: id, Vector: vector, Delete: deleted}})
		if err != nil {
			t.Fatal(err)
		}
		if step%7 == 0 {
			if err = p.Abort(); err != nil {
				t.Fatal(err)
			}
		} else {
			if err = p.Commit(); err != nil {
				t.Fatal(err)
			}
			if deleted {
				delete(truth, id)
			} else {
				truth[id] = vector
			}
		}
		if building != nil && step%5 == 0 {
			if step%10 == 0 {
				err = building.Abort()
			} else {
				err = building.Publish(ctx)
			}
			if err != nil {
				t.Fatal(err)
			}
			building = nil
		}
		oracle := map[string]Candidate{}
		for id, v := range truth {
			oracle[id] = Candidate{id, 2*float64(v[0]) - float64(v[1])}
		}
		want := ordered(oracle)
		want = want[:min(10, len(want))]
		got, err := viewTop(m.View(), query, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) || len(got) > 0 && !reflect.DeepEqual(got, want) {
			t.Fatalf("step%d got%v want%v", step, got, want)
		}
		if step%50 == 0 {
			savedViews = append(savedViews, saved{m.View(), got})
		}
	}
	if building != nil {
		if err := building.Abort(); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range savedViews {
		got, err := viewTop(s.view, query, 10)
		if err != nil || !reflect.DeepEqual(got, s.top) {
			t.Fatal("old ranked view changed", err)
		}
	}
}
