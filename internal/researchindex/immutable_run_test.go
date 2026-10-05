package researchindex

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestImmutableRunsRealGraphOracle(t *testing.T) {
	ctx := context.Background()
	rng := rand.New(rand.NewSource(32452843))
	root := t.TempDir()
	var runs []*ImmutableRun
	latest := map[string][]float32{}
	for stage := 0; stage < 3; stage++ {
		var mutations []Mutation
		for i := 0; i < 40; i++ {
			id := fmt.Sprint(i)
			if stage > 0 && i%3 != stage-1 {
				continue
			}
			m := Mutation{ID: id, Delete: stage > 0 && i%5 == 0}
			if !m.Delete {
				m.Vector = make([]float32, 8)
				for j := range m.Vector {
					m.Vector[j] = rng.Float32() - .5
				}
				latest[id] = append([]float32(nil), m.Vector...)
			} else {
				delete(latest, id)
			}
			mutations = append(mutations, m)
		}
		r, err := BuildImmutableRun(ctx, mutations, 8, filepath.Join(root, fmt.Sprint(stage)))
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		runs = append([]*ImmutableRun{r}, runs...)
		// Ingress vector mutation must not mutate the graph's owned rescoring data.
		for _, m := range mutations {
			for j := range m.Vector {
				m.Vector[j] = 99
			}
		}
		s, err := NewImmutableRunSearch(runs, 10)
		if err != nil {
			t.Fatal(err)
		}
		for probe := 0; probe < 40; probe++ {
			q := make([]float32, 8)
			for j := range q {
				q[j] = rng.Float32() - .5
			}
			var want []Candidate
			for id, v := range latest {
				want = append(want, Candidate{ID: id, Score: cosine(q, v)})
			}
			sort.Slice(want, func(a, b int) bool { return runCandidateLess(want[a], want[b]) })
			want = want[:min(10, len(want))]
			got, err := s.Search(ctx, q)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal(stage, probe, got, want, err)
			}
		}
	}
	old, err := NewImmutableRunSearch(runs[1:], 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = old.Search(ctx, []float32{1, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		t.Fatal("old runs invalidated", err)
	}
	if err = runs[1].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = old.Search(ctx, []float32{1, 0, 0, 0, 0, 0, 0, 0}); err == nil {
		t.Fatal("closed run ignored")
	}
}

func TestImmutableRunTombstoneOnlyAndCancellation(t *testing.T) {
	ctx := context.Background()
	r, err := BuildImmutableRun(ctx, []Mutation{{ID: "gone", Delete: true}}, 2, filepath.Join(t.TempDir(), "run"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	s, err := NewImmutableRunSearch([]*ImmutableRun{r}, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Search(ctx, []float32{1, 0})
	if err != nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	if _, err = s.Search(ctx, []float32{0, 0}); err == nil {
		t.Fatal("zero query accepted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.Search(cancelled, []float32{1, 0}); err == nil {
		t.Fatal("cancellation ignored")
	}
	r.Close()
	if _, err = s.Search(ctx, []float32{1, 0}); err == nil {
		t.Fatal("closed empty graph ignored")
	}
}
