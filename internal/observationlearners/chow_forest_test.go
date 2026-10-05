package observationlearners

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"sort"
	"testing"
)

type heldoutInputForest struct {
	weights  [512]float64
	scores   [10]float64 // uniform, independent fitted marginals, then1..8edges
	selected int
}

// Train and validation are disjoint chronological blocks. Validation chooses
// among a training-defined finite path; there is no refit on validation data.
// This is a smoothed discrete adaptation, not the paper's kernel estimator.
func fitHeldoutInputForest(samples []observation.Sample) (*heldoutInputForest, error) {
	if len(samples) < 4 || len(samples) > 8192 {
		return nil, fmt.Errorf("invalid forest sample count")
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, fmt.Errorf("invalid bits")
		}
	}
	cut := len(samples) / 2
	tree, err := fitChowInput(samples[:cut])
	if err != nil {
		return nil, err
	}
	type edge struct {
		i, j int
		mi   float64
	}
	var edges []edge
	for j := 1; j < 9; j++ {
		i := tree.parent[j]
		edges = append(edges, edge{i, j, tree.mi[i][j]})
	}
	sort.SliceStable(edges, func(i, j int) bool { return edges[i].mi > edges[j].mi })
	var singleton [9][2]float64
	var pairs [8][4]float64
	for x, p := range tree.weights {
		for i := 0; i < 9; i++ {
			singleton[i][(x>>i)&1] += p
		}
		for i, e := range edges {
			pairs[i][2*((x>>e.i)&1)+((x>>e.j)&1)] += p
		}
	}
	var candidates [10][512]float64
	for x := range candidates[0] {
		candidates[0][x] = 1. / 512
		p := 1.
		for i := 0; i < 9; i++ {
			p *= singleton[i][(x>>i)&1]
		}
		candidates[1][x] = p
		for i, e := range edges {
			a, b := (x>>e.i)&1, (x>>e.j)&1
			p *= pairs[i][2*a+b] / (singleton[e.i][a] * singleton[e.j][b])
			candidates[i+2][x] = p
		}
	}
	out := new(heldoutInputForest)
	for i, w := range candidates {
		mass := 0.
		for _, p := range w {
			if p <= 0 || math.IsNaN(p) || math.IsInf(p, 0) {
				return nil, fmt.Errorf("invalid forest mass")
			}
			mass += p
		}
		if math.Abs(mass-1) > 1e-10 {
			return nil, fmt.Errorf("unnormalized forest")
		}
		for _, s := range samples[cut:] {
			out.scores[i] += math.Log(w[s.Bits])
		}
		if out.scores[i] > out.scores[out.selected]+1e-12 {
			out.selected = i
		}
	}
	out.weights = candidates[out.selected]
	return out, nil
}

func TestHeldoutInputForest(t *testing.T) {
	for dep := 0; dep < 2; dep++ {
		var samples []observation.Sample
		for repeat := 0; repeat < 2; repeat++ {
			for x := uint16(0); x < 512; x++ {
				v := x
				if dep == 1 {
					v = (x &^ 1) | ((x >> 2) & 1)
				}
				samples = append(samples, observation.Sample{Bits: v})
			}
		}
		m, err := fitHeldoutInputForest(samples)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if dep == 1 {
			want = 2
		}
		if m.selected != want {
			t.Fatal("population fixture selection", dep, m.selected)
		}
		literal := 0.
		for _, s := range samples[len(samples)/2:] {
			literal += math.Log(m.weights[s.Bits])
		}
		if math.Abs(literal-m.scores[m.selected]) > 1e-10 {
			t.Fatal("selection score")
		}
		for _, score := range m.scores {
			if score > literal+1e-10 {
				t.Fatal("nonmaximal candidate")
			}
		}
		for i := range samples {
			samples[i].Outcome = !samples[i].Outcome
		}
		other, err := fitHeldoutInputForest(samples)
		if err != nil || *m != *other {
			t.Fatal("label leakage")
		}
		// Validation changes scores, not the fitted training law of a fixed topology.
		for i := len(samples) / 2; i < len(samples); i++ {
			samples[i].Bits = 1
		}
		changed, err := fitHeldoutInputForest(samples)
		if err != nil {
			t.Fatal(err)
		}
		if dep == 1 && changed.scores == m.scores {
			t.Fatal("validation unused")
		}
		if dep == 0 && changed.scores != m.scores {
			t.Fatal("uniform likelihood depends on input")
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 3), make([]observation.Sample, 8193), {{Bits: 0}, {Bits: 0}, {Bits: 512}, {Bits: 0}}} {
		if _, err := fitHeldoutInputForest(bad); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

var forestInputSink *heldoutInputForest

func BenchmarkHeldoutInputForest64(b *testing.B) {
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i].Bits = uint16(i * 7 % 512)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m, err := fitHeldoutInputForest(samples)
		if err != nil {
			b.Fatal(err)
		}
		forestInputSink = m
	}
}
