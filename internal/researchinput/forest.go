// Package researchinput is an isolated finite input-law prototype, not serving.
package researchinput

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"sort"
)

type chowInput struct {
	weights [512]float64
	parent  [9]int
	mi      [9][9]float64
}

// Research-only Chow-Liu input law. Pair cells receive .5 pseudo-count each;
// their singleton marginals therefore have one pseudo-count per outcome.
// Outcome labels never enter tree selection or parameter estimation.
func fitChowInput(samples []observation.Sample) (*chowInput, error) {
	if len(samples) == 0 || len(samples) > 8192 {
		return nil, fmt.Errorf("invalid input sample count")
	}
	var pairs [9][9][4]float64
	var single [9][2]float64
	for i := 0; i < 9; i++ {
		single[i] = [2]float64{1, 1}
		for j := i + 1; j < 9; j++ {
			pairs[i][j] = [4]float64{.5, .5, .5, .5}
		}
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, fmt.Errorf("invalid input bits")
		}
		for i := 0; i < 9; i++ {
			a := (s.Bits >> i) & 1
			single[i][a]++
			for j := i + 1; j < 9; j++ {
				b := (s.Bits >> j) & 1
				pairs[i][j][2*a+b]++
			}
		}
	}
	m := new(chowInput)
	total := float64(len(samples) + 2)
	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					p := pairs[i][j][2*a+b] / total
					m.mi[i][j] += p * math.Log(p*total*total/(single[i][a]*single[j][b]))
				}
			}
			m.mi[j][i] = m.mi[i][j]
		}
	}
	// Deterministic maximum spanning tree, rooted at field0. Ties use loop order.
	var seen [9]bool
	seen[0] = true
	m.parent[0] = -1
	for k := 1; k < 9; k++ {
		best := -math.MaxFloat64
		from, to := -1, -1
		for i := 0; i < 9; i++ {
			if !seen[i] {
				continue
			}
			for j := 0; j < 9; j++ {
				if !seen[j] && m.mi[i][j] > best {
					best = m.mi[i][j]
					from, to = i, j
				}
			}
		}
		m.parent[to] = from
		seen[to] = true
	}
	for x := uint16(0); x < 512; x++ {
		p := single[0][x&1] / total
		for j := 1; j < 9; j++ {
			i := m.parent[j]
			a, b := int((x>>i)&1), int((x>>j)&1)
			pair := 0.
			if i < j {
				pair = pairs[i][j][2*a+b]
			} else {
				pair = pairs[j][i][2*b+a]
			}
			p *= pair / single[i][a]
		}
		m.weights[x] = p
	}
	return m, nil
}

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

// Fit preserves the frozen component; returned weights have independent ownership.
func Fit(samples []observation.Sample) ([512]float64, error) {
	m, err := fitHeldoutInputForest(samples)
	if err != nil {
		return [512]float64{}, err
	}
	return m.weights, nil
}
