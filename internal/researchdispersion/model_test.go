package researchdispersion

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchcalibration"
)

func bases(n int) []float64 {
	b := make([]float64, n)
	for i := range b {
		b[i] = .925 - .675*float64(i)/float64(n-1)
	}
	return b
}

// Batch beta integrals are independent of the sequential update recursion.
// Ordered binary trials do not carry the binomial-count combinatorial term.
func reference(m *Model, mode string) ([Hypotheses * Strengths]float64, []float64) {
	var logs, weights [Hypotheses * Strengths]float64
	maximum := math.Inf(-1)
	lbeta := func(a, b float64) float64 {
		x, _ := math.Lgamma(a)
		y, _ := math.Lgamma(b)
		z, _ := math.Lgamma(a + b)
		return x + y - z
	}
	for h := 0; h < Hypotheses; h++ {
		prior := .004
		if h == 0 {
			prior = .1
		} else if h == 1 {
			prior = .8
		}
		for k, c := range concentrations {
			z := h*Strengths + k
			logs[z] = math.Inf(-1)
			if mode != "adaptive" && !(mode == "fixed2" && k == 1) && !(mode == "shared" && k == 4) {
				continue
			}
			logs[z] = math.Log(prior)
			if mode == "adaptive" {
				logs[z] -= math.Log(Strengths)
			}
			for i, n := range m.n {
				p := m.p[i*Hypotheses+h]
				s, f := float64(m.success[i]), float64(n-m.success[i])
				if c == 0 {
					logs[z] += s*math.Log(p) + f*math.Log1p(-p)
				} else {
					logs[z] += lbeta(c*p+s, c*(1-p)+f) - lbeta(c*p, c*(1-p))
				}
			}
			maximum = math.Max(maximum, logs[z])
		}
	}
	total := 0.0
	for z, v := range logs {
		weights[z] = math.Exp(v - maximum)
		total += weights[z]
	}
	q := make([]float64, len(m.n))
	for z := range weights {
		weights[z] /= total
		h, k := z/Strengths, z%Strengths
		for i := range q {
			p, c := m.p[i*Hypotheses+h], concentrations[k]
			if c != 0 {
				p = (c*p + float64(m.success[i])) / (c + float64(m.n[i]))
			}
			q[i] += weights[z] * p
		}
	}
	return weights, q
}

func near(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if math.IsNaN(got) || math.Abs(got-want) > tolerance {
		t.Fatalf("got %.17g want %.17g tolerance %g", got, want, tolerance)
	}
}

func TestIntegratedJointAndPredictive(t *testing.T) {
	for _, size := range []int{2, 11, 150, 200} {
		for _, mode := range []string{"adaptive", "fixed2", "shared"} {
			m, err := New(bases(size), mode)
			if err != nil {
				t.Fatal(err)
			}
			rng := rand.New(rand.NewSource(int64(67 + size)))
			for round := 1; round <= MaxTrials; round++ {
				for _, i := range rng.Perm(size) {
					if err := m.Observe(i, round, rng.Float64() < .2+.6*float64(i)/float64(size-1)); err != nil {
						t.Fatal(err)
					}
				}
				if round != 1 && round != 2 && round != 16 && round != MaxTrials {
					continue
				}
				w, q := reference(m, mode)
				for z := range w {
					near(t, m.w[z], w[z], 2e-10)
				}
				for i := range q {
					got, err := m.Predict(i)
					if err != nil {
						t.Fatal(err)
					}
					near(t, got, q[i], 2e-10)
				}
			}
		}
	}
}

func TestOneTrialCannotIdentifyDispersion(t *testing.T) {
	for _, size := range []int{2, 11, 150, 200} {
		m, _ := New(bases(size), "adaptive")
		fixed, _ := New(bases(size), "fixed2")
		old, _ := researchcalibration.New(bases(size))
		for i := 0; i < size; i++ {
			useful := i%3 == 0
			if err := m.Observe(i, 1, useful); err != nil {
				t.Fatal(err)
			}
			if err := fixed.Observe(i, 1, useful); err != nil {
				t.Fatal(err)
			}
			if err := old.Observe(i, useful); err != nil {
				t.Fatal(err)
			}
			for _, w := range m.Dispersion() {
				near(t, w, .2, 5e-14)
			}
			for j := 0; j < size; j++ {
				q, _ := fixed.Predict(j)
				wanted, _ := old.Predict(j)
				near(t, q, wanted, 5e-14)
			}
		}
		// A second trial has a concentration-dependent conditional likelihood.
		if err := m.Observe(0, 2, true); err != nil {
			t.Fatal(err)
		}
		moved := false
		for _, w := range m.Dispersion() {
			moved = moved || math.Abs(w-.2) > 1e-4
		}
		if !moved {
			t.Fatal("repeated trial failed to distinguish strength components")
		}
	}
}

func TestInvalidReplayOwnershipAndCap(t *testing.T) {
	for _, mode := range []string{"", "other"} {
		if _, err := New(bases(2), mode); err == nil {
			t.Fatal("invalid mode")
		}
	}
	for _, b := range [][]float64{nil, {.5}, make([]float64, 201), {math.NaN(), .5}, {math.Inf(1), .5}, {.249, .5}, {.926, .5}} {
		if _, err := New(b, "adaptive"); err == nil {
			t.Fatal("invalid baseline")
		}
	}
	input := bases(11)
	m, _ := New(input, "adaptive")
	input[0] = 0
	copyState := func() Model {
		v := *m
		v.p = append([]float64(nil), m.p...)
		v.n = append([]uint16(nil), m.n...)
		v.success = append([]uint16(nil), m.success...)
		return v
	}
	before := copyState()
	q, _ := m.Predict(0)
	if q <= 0 || !reflect.DeepEqual(before, *m) {
		t.Fatal("prediction changed or aliased state")
	}
	for _, trial := range [][2]int{{-1, 1}, {11, 1}, {0, 0}, {0, 2}, {0, -1}, {0, 65}} {
		if err := m.Observe(trial[0], trial[1], true); err == nil || !reflect.DeepEqual(before, *m) {
			t.Fatal("invalid evidence mutated state", trial)
		}
	}
	for round := 1; round <= MaxTrials; round++ {
		if err := m.Observe(0, round, round%2 == 0); err != nil {
			t.Fatal(err)
		}
	}
	before = copyState()
	for _, round := range []int{1, 64, 65} {
		if err := m.Observe(0, round, true); err == nil || !reflect.DeepEqual(before, *m) {
			t.Fatal("replayed/capped trial mutated state")
		}
	}
	w := m.Dispersion()
	w[0] = 8
	if m.Dispersion()[0] == 8 {
		t.Fatal("marginal aliases internal state")
	}
	for _, i := range []int{-1, 11} {
		if _, err := m.Predict(i); err == nil {
			t.Fatal("unknown member")
		}
	}
}

func TestRetainedLogsRecoverUnderflow(t *testing.T) {
	m, _ := New([]float64{.25, .25, .25, .25, .25, .25, .25, .25}, "shared")
	for z := range m.logs {
		m.logs[z], m.w[z] = math.Inf(-1), 0
	}
	low, high := 2*Strengths+4, 24*Strengths+4 // Constant means .1 and .9.
	m.logs[low], m.w[low], m.logs[high] = 0, 1, -800
	for i := 0; i < 8; i++ {
		for ordinal := 1; ordinal <= MaxTrials; ordinal++ {
			if err := m.Observe(i, ordinal, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	if m.w[high] < .999 || math.IsInf(m.logs[high], -1) {
		t.Fatal("numerical zero made supported hypothesis permanently absorbing")
	}
}

func BenchmarkPredict150(b *testing.B) {
	m, _ := New(bases(150), "adaptive")
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

func BenchmarkFreshMemberUpdate(b *testing.B) {
	m, _ := New(bases(150), "adaptive")
	original := *m
	b.ReportAllocs()
	b.ResetTimer()
	for z := 0; z < b.N; z++ {
		m.logs, m.w = original.logs, original.w
		m.n[0], m.success[0] = 0, 0
		if err := m.Observe(0, 1, z%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
}
