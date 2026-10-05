package observationlearners

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"sort"
	"testing"
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

func TestChowInputContracts(t *testing.T) {
	var uniform, copyData []observation.Sample
	for x := uint16(0); x < 512; x++ {
		uniform = append(uniform, observation.Sample{Bits: x})
		copyData = append(copyData, observation.Sample{Bits: (x &^ 1) | ((x >> 2) & 1)})
	}
	for _, samples := range [][]observation.Sample{uniform, copyData, {{Bits: 7, Outcome: true}}} {
		m, err := fitChowInput(samples)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0.
		for _, p := range m.weights {
			if p <= 0 || math.IsNaN(p) || math.IsInf(p, 0) {
				t.Fatal("support")
			}
			sum += p
		}
		if math.Abs(sum-1) > 1e-12 {
			t.Fatal("normalization", sum)
		}
		flipped := append([]observation.Sample(nil), samples...)
		for i := range flipped {
			flipped[i].Outcome = !flipped[i].Outcome
		}
		other, err := fitChowInput(flipped)
		if err != nil || *other != *m {
			t.Fatal("label leakage")
		}
		// Independent Kruskal check of Prim's total information weight.
		type edge struct {
			i, j int
			w    float64
		}
		var es []edge
		var roots [9]int
		for i := range roots {
			roots[i] = i
			for j := i + 1; j < 9; j++ {
				es = append(es, edge{i, j, m.mi[i][j]})
			}
		}
		sort.SliceStable(es, func(i, j int) bool { return es[i].w > es[j].w })
		find := func(i int) int {
			for roots[i] != i {
				i = roots[i]
			}
			return i
		}
		kruskal := 0.
		for _, e := range es {
			a, b := find(e.i), find(e.j)
			if a != b {
				roots[a] = b
				kruskal += e.w
			}
		}
		prim := 0.
		for j := 1; j < 9; j++ {
			prim += m.mi[j][m.parent[j]]
			visited := map[int]bool{}
			for v := j; v >= 0; v = m.parent[v] {
				if visited[v] {
					t.Fatal("cycle")
				}
				visited[v] = true
			}
		}
		if math.Abs(prim-kruskal) > 1e-12 {
			t.Fatal("spanning objective")
		}
		// Every selected edge must recover the declared smoothed pair marginal.
		for j := 1; j < 9; j++ {
			i := m.parent[j]
			want := [4]float64{.5, .5, .5, .5}
			var got [4]float64
			for _, s := range samples {
				want[2*((s.Bits>>i)&1)+((s.Bits>>j)&1)]++
			}
			for x, p := range m.weights {
				got[2*((x>>i)&1)+((x>>j)&1)] += p
			}
			for k := range got {
				if math.Abs(got[k]-want[k]/float64(len(samples)+2)) > 1e-12 {
					t.Fatal("edge marginal")
				}
			}
		}
	}
	u, _ := fitChowInput(uniform)
	for _, p := range u.weights {
		if p != 1./512 {
			t.Fatal("uniform negative control")
		}
	}
	c, _ := fitChowInput(copyData)
	mass := 0.
	for x, p := range c.weights {
		if (x & 1) == ((x >> 2) & 1) {
			mass += p
		}
	}
	if math.Abs(mass-513./514) > 1e-12 {
		t.Fatal("copied dependency", mass)
	}
	for _, bad := range [][]observation.Sample{nil, {{Bits: 512}}, make([]observation.Sample, 8193)} {
		if _, err := fitChowInput(bad); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

var chowSink *chowInput

func BenchmarkChowInput64(b *testing.B) {
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i].Bits = uint16(i * 7 % 512)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m, err := fitChowInput(samples)
		if err != nil {
			b.Fatal(err)
		}
		chowSink = m
	}
}
