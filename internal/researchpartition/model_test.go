package researchpartition

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func inputs(n int) ([]float64, []float64) {
	b, r := make([]float64, n), make([]float64, n)
	for i := range b {
		r[i] = float64(i) / float64(n-1)
		b[i] = .9 - .6*r[i]
	}
	return b, r
}
func TestFamilyAndColdPrior(t *testing.T) {
	b, r := inputs(150)
	m, err := New(b, r)
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, p := range m.partitions {
		sum += p.prior
		for bin, node := range p.node {
			if node > 14 {
				t.Fatal("invalid leaf")
			}
			ancestor := bin + 7
			for ancestor != int(node) && ancestor > 0 {
				ancestor = (ancestor - 1) / 2
			}
			if ancestor != int(node) {
				t.Fatal("partition leaf is not an ancestor")
			}
		}
	}
	if math.Abs(sum-1) > 1e-14 {
		t.Fatal("prior not normalized")
	}
	for i := range b {
		q, _ := m.Predict(i)
		if math.Abs(q-(.99*b[i]+.005)) > 1e-14 {
			t.Fatal("cold prior mismatch")
		}
	}
	b[0] = .1
	r[0] = 1
	q, _ := m.Predict(0)
	if math.Abs(q-.896) > 1e-14 {
		t.Fatal("aliased input")
	}
}

func logBeta(a, b float64) float64 {
	x, _ := math.Lgamma(a)
	y, _ := math.Lgamma(b)
	z, _ := math.Lgamma(a + b)
	return x + y - z
}

// Reconstruct the posterior from integrated Beta leaf likelihoods, independent
// of the sequential normalizer and its cached node counts.
func TestIntegratedJointEvidenceAndPredictive(t *testing.T) {
	b, r := inputs(16)
	m, _ := New(b, r)
	seen := map[int]bool{}
	for _, observation := range []struct {
		i int
		y bool
	}{{0, true}, {1, false}, {8, true}, {15, false}, {4, true}} {
		if err := m.Observe(observation.i, observation.y); err != nil {
			t.Fatal(err)
		}
		seen[observation.i] = observation.y
		var weights [Trees + 1]float64
		weights[0] = .99
		for i, y := range seen {
			if y {
				weights[0] *= b[i]
			} else {
				weights[0] *= 1 - b[i]
			}
		}
		for k, p := range m.partitions {
			logL := math.Log(.01 * p.prior)
			counts := map[uint8][2]float64{}
			for i, y := range seen {
				node := p.node[m.bin[i]]
				v := counts[node]
				if y {
					v[0]++
				} else {
					v[1]++
				}
				counts[node] = v
			}
			for _, v := range counts {
				logL += logBeta(1+v[0], 1+v[1]) - logBeta(1, 1)
			}
			weights[k+1] = math.Exp(logL)
		}
		mass := 0.0
		for _, w := range weights {
			mass += w
		}
		for k := range weights {
			weights[k] /= mass
			if math.Abs(weights[k]-m.weights[k]) > 2e-14 {
				t.Fatal("integrated joint likelihood mismatch")
			}
		}
		for i := range b {
			want := 0.0
			for k, w := range weights {
				mean := b[i]
				if k > 0 {
					node := m.partitions[k-1].node[m.bin[i]]
					s, f := 0.0, 0.0
					for j, y := range seen {
						if m.partitions[k-1].node[m.bin[j]] == node {
							if y {
								s++
							} else {
								f++
							}
						}
					}
					mean = (1 + s) / (2 + s + f)
				}
				if y, ok := seen[i]; ok {
					v := 0.0
					if y {
						v = 1
					}
					mean = (2*mean + v) / 3
				}
				want += w * mean
			}
			q, _ := m.Predict(i)
			if math.Abs(q-want) > 2e-14 {
				t.Fatal("joint posterior predictive mismatch")
			}
		}
	}
}

func TestRejectionsAndSelectors(t *testing.T) {
	b, r := inputs(8)
	m, _ := New(b, r)
	_ = m.Observe(3, true)
	copy := *m
	copy.seen = append([]bool(nil), m.seen...)
	copy.useful = append([]bool(nil), m.useful...)
	for _, i := range []int{-1, 8, 3} {
		if m.Observe(i, false) == nil {
			t.Fatal("accepted invalid evidence")
		}
		if !reflect.DeepEqual(copy, *m) {
			t.Fatal("rejection mutated state")
		}
	}
	for _, policy := range []string{"head", "stratified", "random", "uncertainty"} {
		m, _ := New(b, r)
		rng := rand.New(rand.NewSource(78813))
		for k := 0; k < 8; k++ {
			i, err := m.Select(policy, rng)
			if err != nil || m.seen[i] {
				t.Fatal("bad selection")
			}
			_ = m.Observe(i, k%3 == 0)
		}
		if _, err := m.Select(policy, rng); err == nil {
			t.Fatal("exhaustion accepted")
		}
		for i := range b {
			q, _ := m.Predict(i)
			if math.IsNaN(q) || q < 0 || q > 1 {
				t.Fatal("invalid forecast")
			}
		}
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		bad := append([]float64(nil), r...)
		bad[0] = v
		if _, err := New(b, bad); err == nil {
			t.Fatal("invalid coordinate")
		}
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), 0, 1} {
		bad := append([]float64(nil), b...)
		bad[0] = v
		if _, err := New(bad, r); err == nil {
			t.Fatal("invalid baseline")
		}
	}
	if _, err := m.Select("random", nil); err == nil {
		t.Fatal("missing RNG accepted")
	}
	if _, err := m.Select("other", nil); err == nil {
		t.Fatal("unknown policy accepted")
	}
}

func BenchmarkPartition(b *testing.B) {
	base, coordinate := inputs(150)
	m, _ := New(base, coordinate)
	b.Run("predict150", func(b *testing.B) {
		b.ReportAllocs()
		for k := 0; k < b.N; k++ {
			for i := 0; i < 150; i++ {
				_, _ = m.Predict(i)
			}
		}
	})
	b.Run("construct", func(b *testing.B) {
		b.ReportAllocs()
		for k := 0; k < b.N; k++ {
			_, _ = New(base, coordinate)
		}
	})
}
