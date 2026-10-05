package researchindex

import (
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"sort"
	"testing"
)

func ordered(m map[string]Candidate) []Candidate {
	var out []Candidate
	for _, c := range m {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].ID < out[j].ID
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func TestMergeMatchesFullExactOracle(t *testing.T) {
	rng := rand.New(rand.NewPCG(173, 991))
	for trial := 0; trial < 2000; trial++ {
		n, k := 1+rng.IntN(400), 1+rng.IntN(200)
		base := map[string]Candidate{}
		truth := map[string]Candidate{}
		for i := 0; i < n; i++ {
			c := Candidate{fmt.Sprint(i), float64(rng.IntN(20))}
			base[c.ID] = c
			truth[c.ID] = c
		}
		var delta []Change
		ids := rng.Perm(n + 50)
		for _, id := range ids[:rng.IntN(min(128, len(ids))+1)] {
			c := Candidate{fmt.Sprint(id), float64(rng.IntN(20))}
			deleted := rng.IntN(3) == 0
			delta = append(delta, Change{c, deleted})
			if deleted {
				delete(truth, c.ID)
			} else {
				truth[c.ID] = c
			}
		}
		prefix := ordered(base)
		prefix = prefix[:min(len(prefix), k+len(delta))]
		got, err := Merge(prefix, delta, k)
		if err != nil {
			t.Fatal(err)
		}
		want := ordered(truth)
		want = want[:min(k, len(want))]
		if len(got) != len(want) || len(got) > 0 && !reflect.DeepEqual(got, want) {
			t.Fatalf("trial%d: got%v want%v", trial, got, want)
		}
	}
}

func TestMergeMustOverfetchAndShadowDemotions(t *testing.T) {
	base := []Candidate{{"a", 10}, {"b", 9}, {"c", 8}}
	delta := []Change{{Candidate{"a", 0}, true}, {Candidate{"b", -1}, false}}
	got, err := Merge(base, delta, 1)
	if err != nil || len(got) != 1 || got[0].ID != "c" {
		t.Fatal(got, err)
	}
	short, err := Merge(base[:1], delta, 1)
	if err != nil || len(short) != 1 || short[0].ID == "c" {
		t.Fatal("negative control lost its missing-prefix limitation")
	}
}

func TestMergeRejectsInvalidInputs(t *testing.T) {
	for _, base := range [][]Candidate{{{"a", math.NaN()}}, {{"a", 1}, {"a", 2}}} {
		if _, err := Merge(base, nil, 2); err == nil {
			t.Fatal("invalid base accepted")
		}
	}
	if _, err := Merge(nil, []Change{{Candidate{"a", 1}, false}, {Candidate{"a", 2}, false}}, 2); err == nil {
		t.Fatal("duplicate delta")
	}
	if _, err := Merge(nil, make([]Change, 129), 1); err == nil {
		t.Fatal("unbounded delta")
	}
}
