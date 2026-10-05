package researchindex

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestRunMergeRejectsInvalidInputs(t *testing.T) {
	for _, runs := range [][][]RunEntry{nil, {{{ID: ""}}}, {{{ID: "a"}, {ID: "a", Deleted: true}}}} {
		if _, err := NewRunMergePlan(runs, 1, 10); err == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	p, err := NewRunMergePlan([][]RunEntry{{{ID: "a"}, {ID: "b"}, {ID: "d", Deleted: true}}}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, prefix := range [][]Candidate{
		{{ID: "a", Score: 1}, {ID: "a", Score: 0}},
		{{ID: "a", Score: 1}, {ID: "d", Score: 0}},
		{{ID: "a", Score: 1}, {ID: "z", Score: 0}},
		{{ID: "b", Score: 1}, {ID: "a", Score: 2}},
		{{ID: "b", Score: 1}, {ID: "a", Score: 1}},
		{{ID: "a", Score: math.NaN()}, {ID: "b", Score: 0}},
		{{ID: "a", Score: math.Inf(1)}, {ID: "b", Score: 0}},
	} {
		if _, err := p.Merge([][]Candidate{prefix}); err == nil {
			t.Fatal("invalid candidates accepted", prefix)
		}
	}
	if _, err := p.Merge(nil); err == nil {
		t.Fatal("missing run accepted")
	}
}

func TestRunMergeHiddenPrefixAndDemotion(t *testing.T) {
	runs := [][]RunEntry{{{ID: "a", Deleted: true}, {ID: "b"}}, {{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}}
	p, err := NewRunMergePlan(runs, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Limits(), []int{1, 3}) {
		t.Fatal(p.Limits())
	}
	prefixes := [][]Candidate{{{ID: "b", Score: -1}}, {{ID: "a", Score: 10}, {ID: "b", Score: 9}, {ID: "c", Score: 8}}}
	got, err := p.Merge(prefixes)
	if err != nil || len(got) != 1 || got[0].ID != "c" {
		t.Fatal(got, err)
	}
	prefixes[1] = prefixes[1][:1]
	if _, err = p.Merge(prefixes); err == nil {
		t.Fatal("hidden candidate silently dropped")
	}
	if _, err = NewRunMergePlan(runs, 1, 2); !errors.Is(err, ErrCapacity) {
		t.Fatal("unbounded prefix accepted", err)
	}
	// Caller mutation cannot alter a plan already captured by readers.
	runs[0][0].ID = "c"
	limits := p.Limits()
	limits[1] = 0
	if p.Limits()[1] != 3 || p.owner["a"] != 0 {
		t.Fatal("plan aliases input")
	}
}

func TestRunMergeRandomExactOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(2147483647))
	for trial := 0; trial < 300; trial++ {
		n := 1 + rng.Intn(8)
		k := 1 + rng.Intn(12)
		runs := make([][]RunEntry, n)
		prefixes := make([][]Candidate, n)
		latest := map[string]bool{}
		var oracle []Candidate
		for i := range runs {
			for id := 0; id < 50; id++ {
				if rng.Intn(3) == 0 {
					continue
				}
				s := fmt.Sprint(id)
				del := rng.Intn(5) == 0
				c := Candidate{ID: s, Score: float64(rng.Intn(10))}
				runs[i] = append(runs[i], RunEntry{ID: s, Deleted: del})
				if !del {
					prefixes[i] = append(prefixes[i], c)
				}
				if !latest[s] {
					latest[s] = true
					if !del {
						oracle = append(oracle, c)
					}
				}
			}
			sort.Slice(prefixes[i], func(a, b int) bool { return runCandidateLess(prefixes[i][a], prefixes[i][b]) })
		}
		p, err := NewRunMergePlan(runs, k, 100)
		if err != nil {
			t.Fatal(err)
		}
		for i, limit := range p.Limits() {
			prefixes[i] = prefixes[i][:limit]
		}
		got, err := p.Merge(prefixes)
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(oracle, func(a, b int) bool { return runCandidateLess(oracle[a], oracle[b]) })
		oracle = oracle[:min(k, len(oracle))]
		if !reflect.DeepEqual(got, oracle) {
			t.Fatal(trial, got, oracle)
		}
	}
}
