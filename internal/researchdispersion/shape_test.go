package researchdispersion

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// Independent batch integrals and direct atom products avoid the cached
// conditional-mean and log-odds update used by the candidate.
func shapeReference(m *ShapeModel) ([Hypotheses * ShapeKernels]float64, []float64) {
	var logs, weights [Hypotheses * ShapeKernels]float64
	conditionals := make([]float64, len(m.n)*Hypotheses*ShapeKernels)
	lbeta := func(a, b float64) float64 {
		x, _ := math.Lgamma(a)
		y, _ := math.Lgamma(b)
		z, _ := math.Lgamma(a + b)
		return x + y - z
	}
	maximum := math.Inf(-1)
	for h := 0; h < Hypotheses; h++ {
		prior := .004
		if h == 0 {
			prior = .1
		} else if h == 1 {
			prior = .8
		}
		for k := 0; k < ShapeKernels; k++ {
			mass := .5 / 9
			if k < Strengths {
				mass = .5 / Strengths
			}
			z := h*ShapeKernels + k
			logs[z] = math.Log(prior * mass)
			for i, n := range m.n {
				p, s, f := m.p[i*Hypotheses+h], float64(m.success[i]), float64(n-m.success[i])
				logMass, q := 0., p
				if k < Strengths {
					c := concentrations[k]
					if c == 0 {
						logMass = s*math.Log(p) + f*math.Log1p(-p)
					} else {
						a, b := c*p, c*(1-p)
						logMass = lbeta(a+s, b+f) - lbeta(a, b)
						q = math.Exp(lbeta(a+s+1, b+f) - lbeta(a+s, b+f))
					}
				} else {
					d := float64(k-Strengths+1) / 20
					lo, hi := math.Max(.01, p-d), math.Min(.99, p+d)
					high, low := (p-lo)/(hi-lo), (hi-p)/(hi-lo)
					x := low * math.Pow(lo, s) * math.Pow(1-lo, f)
					y := high * math.Pow(hi, s) * math.Pow(1-hi, f)
					logMass = math.Log(x + y)
					q = (lo*x + hi*y) / (x + y)
				}
				logs[z] += logMass
				conditionals[i*Hypotheses*ShapeKernels+z] = q
			}
			maximum = math.Max(maximum, logs[z])
		}
	}
	sum := 0.
	for z := range weights {
		weights[z] = math.Exp(logs[z] - maximum)
		sum += weights[z]
	}
	q := make([]float64, len(m.n))
	for z := range weights {
		weights[z] /= sum
		for i := range q {
			q[i] += weights[z] * conditionals[i*Hypotheses*ShapeKernels+z]
		}
	}
	return weights, q
}

func cloneShape(m *ShapeModel) ShapeModel {
	v := *m
	v.p, v.cache = append([]float64(nil), m.p...), append([]float64(nil), m.cache...)
	v.n, v.success = append([]uint16(nil), m.n...), append([]uint16(nil), m.success...)
	return v
}

func TestShapeIntegratedJoint(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		m, err := NewShape(bases(n))
		if err != nil {
			t.Fatal(err)
		}
		rng := rand.New(rand.NewSource(int64(73 + n)))
		for round := 1; round <= MaxTrials; round++ {
			for _, i := range rng.Perm(n) {
				if err := m.Observe(i, round, rng.Float64() < .15+.7*float64(i)/float64(n-1)); err != nil {
					t.Fatal(err)
				}
			}
			if round != 1 && round != 2 && round != 8 && round != MaxTrials {
				continue
			}
			weights, forecasts := shapeReference(m)
			for z, w := range weights {
				near(t, m.w[z], w, 3e-10)
			}
			for i, q := range forecasts {
				got, err := m.Predict(i)
				if err != nil {
					t.Fatal(err)
				}
				near(t, got, q, 3e-10)
			}
		}
	}
}

func TestShapeMeanAndOneTrial(t *testing.T) {
	m, _ := NewShape(bases(150))
	for _, p := range []float64{.02, .05, .5, .925, .978125} {
		for k := Strengths; k < ShapeKernels; k++ {
			lo, hi := atomBounds(p, k)
			high := (p - lo) / (hi - lo)
			near(t, lo*(1-high)+hi*high, p, 2e-15)
			if lo <= 0 || hi >= 1 || lo >= p || hi <= p || high <= 0 || high >= 1 {
				t.Fatal("invalid clipped mean-preserving prior")
			}
		}
	}
	for i := range m.n {
		if err := m.Observe(i, 1, i%4 == 0); err != nil {
			t.Fatal(err)
		}
		for k, w := range m.Shapes() {
			prior := .5 / 9
			if k < Strengths {
				prior = .5 / Strengths
			}
			near(t, w, prior, 5e-14)
		}
	}
	prior := m.Shapes()
	if err := m.Observe(0, 2, true); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(prior, m.Shapes()) {
		t.Fatal("repeated-trial likelihood did not learn shape")
	}
}

func TestShapeCacheOwnershipReplayAndAtomicity(t *testing.T) {
	input := bases(11)
	m, err := NewShape(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 0
	before := cloneShape(m)
	for i := 0; i < 11; i++ {
		if q, err := m.Predict(i); err != nil || q <= 0 || q >= 1 {
			t.Fatal("invalid owned prediction")
		}
	}
	if !reflect.DeepEqual(before, *m) {
		t.Fatal("prediction changed state")
	}
	for _, trial := range [][2]int{{-1, 1}, {11, 1}, {0, 0}, {0, 2}, {0, 65}} {
		if err := m.Observe(trial[0], trial[1], true); err == nil || !reflect.DeepEqual(before, *m) {
			t.Fatal("invalid trial changed state")
		}
	}
	if err := m.Observe(0, 1, true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.cache[Hypotheses*ShapeKernels:], m.cache[Hypotheses*ShapeKernels:]) {
		t.Fatal("unrelated member cache changed")
	}
	for round := 2; round <= MaxTrials; round++ {
		if err := m.Observe(0, round, round%2 == 0); err != nil {
			t.Fatal(err)
		}
	}
	before = cloneShape(m)
	for _, ordinal := range []int{1, 64, 65} {
		if err := m.Observe(0, ordinal, false); err == nil || !reflect.DeepEqual(before, *m) {
			t.Fatal("replayed/capped trial changed state")
		}
	}
	m.logs[0] = math.Inf(1)
	before = cloneShape(m)
	if err := m.Observe(1, 1, true); err == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("invalid normalization committed")
	}
	weights := m.Shapes()
	weights[0] = 10
	if m.Shapes()[0] == 10 {
		t.Fatal("exposed mutable weights")
	}
	for _, i := range []int{-1, 11} {
		if _, err := m.Predict(i); err == nil {
			t.Fatal("unknown member")
		}
	}
	for _, base := range [][]float64{nil, {.5}, make([]float64, 201), {math.NaN(), .5}, {.249, .5}, {.926, .5}} {
		if _, err := NewShape(base); err == nil {
			t.Fatal("unsupported baseline")
		}
	}
}

func TestShapeUnderflowRecovery(t *testing.T) {
	m, _ := NewShape([]float64{.25, .25, .25, .25, .25, .25, .25, .25})
	for z := range m.logs {
		m.logs[z], m.w[z] = math.Inf(-1), 0
	}
	low, high := 2*ShapeKernels+4, 24*ShapeKernels+4
	m.logs[low], m.w[low], m.logs[high] = 0, 1, -800
	for i := range m.n {
		for ordinal := 1; ordinal <= MaxTrials; ordinal++ {
			if err := m.Observe(i, ordinal, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	if m.w[high] < .999 || math.IsInf(m.logs[high], -1) {
		t.Fatal("underflow eliminated supported state")
	}
}

func BenchmarkShapePredict150(b *testing.B) {
	m, _ := NewShape(bases(150))
	b.ReportAllocs()
	b.ResetTimer()
	for z := 0; z < b.N; z++ {
		for i := 0; i < 150; i++ {
			if _, err := m.Predict(i); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkShapeUpdate16(b *testing.B) {
	m, _ := NewShape(bases(150))
	for round := 1; round <= 15; round++ {
		for i := range m.n {
			if err := m.Observe(i, round, (round+i)%3 != 0); err != nil {
				b.Fatal(err)
			}
		}
	}
	before := *m
	row := append([]float64(nil), m.cache[:Hypotheses*ShapeKernels]...)
	n, s := m.n[0], m.success[0]
	b.ReportAllocs()
	b.ResetTimer()
	for z := 0; z < b.N; z++ {
		m.logs, m.w, m.n[0], m.success[0] = before.logs, before.w, n, s
		copy(m.cache[:Hypotheses*ShapeKernels], row)
		if err := m.Observe(0, 16, z%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
}

var benchmarkShapeModel *ShapeModel

func BenchmarkShapeConstruct150(b *testing.B) {
	base := bases(150)
	b.ReportAllocs()
	b.ResetTimer()
	for z := 0; z < b.N; z++ {
		var err error
		benchmarkShapeModel, err = NewShape(base)
		if err != nil {
			b.Fatal(err)
		}
	}
}
