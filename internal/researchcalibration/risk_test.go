package researchcalibration

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func cloneRiskTest(m *Model) *Model {
	c := *m
	c.seen = append([]bool(nil), m.seen...)
	c.useful = append([]bool(nil), m.useful...)
	return &c
}

// Independent two-outcome lookahead includes the same-event Beta update.
func explicitRiskTest(m *Model, e int, priority []float64) float64 {
	q, _ := m.Predict(e)
	positive, negative := cloneRiskTest(m), cloneRiskTest(m)
	if positive.Observe(e, true) != nil || negative.Observe(e, false) != nil {
		panic("invalid lookahead")
	}
	mass, gain := 0.0, 0.0
	for j, v := range priority {
		before, _ := m.Predict(j)
		yes, _ := positive.Predict(j)
		no, _ := negative.Predict(j)
		gain += v * (before*(1-before) - q*yes*(1-yes) - (1-q)*no*(1-no))
		mass += v
	}
	return gain / mass
}

func TestRiskGramAgreesWithExactLookahead(t *testing.T) {
	for _, n := range []int{2, 11, 150} {
		m, _ := New(fixtureBase(n))
		rng := rand.New(rand.NewSource(881901))
		priority := make([]float64, n)
		for i := range priority {
			priority[i] = 1
			if i < 10 {
				priority[i] = 3
			}
		}
		for step := 0; step < 4 && step < n; step++ {
			before := cloneRiskTest(m)
			scores, err := m.RiskScores(priority)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, m) {
				t.Fatal("selection changed posterior")
			}
			for e, score := range scores {
				if m.seen[e] {
					if score != -1 {
						t.Fatal("seen event activated")
					}
					continue
				}
				want := explicitRiskTest(m, e, priority)
				if score < 0 || math.Abs(score-want) > 2e-14 {
					t.Fatalf("n=%d step=%d e=%d: %g != %g", n, step, e, score, want)
				}
			}
			e, _, err := m.SelectRisk(priority)
			if err != nil {
				t.Fatal(err)
			}
			_ = m.Observe(e, rng.Float64() < .5)
		}
	}
}

func TestRiskPriorityAndIndividualUncertainty(t *testing.T) {
	m, _ := New([]float64{.5, .5})
	p := []float64{1, 0}
	scores, err := m.RiskScores(p)
	if err != nil {
		t.Fatal(err)
	}
	if scores[0] <= scores[1] {
		t.Fatal("ignored queried-event uncertainty")
	}
	for _, bad := range [][]float64{nil, {0, 0}, {-1, 1}, {math.NaN(), 1}, {math.Inf(1), 1}, {math.MaxFloat64, math.MaxFloat64}} {
		if _, err := m.RiskScores(bad); err == nil {
			t.Fatal("accepted invalid priorities")
		}
	}
	if _, err := m.RiskScores([]float64{1, 1}); err != nil {
		t.Fatal(err)
	}
	_ = m.Observe(0, true)
	_ = m.Observe(1, false)
	if _, _, err := m.SelectRisk(p); err == nil {
		t.Fatal("exhausted selection accepted")
	}
}

func BenchmarkRiskSelection(b *testing.B) {
	m, _ := New(fixtureBase(150))
	p := make([]float64, 150)
	for i := range p {
		p[i] = 1
		if i < 10 {
			p[i] = 3
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		_, _, _ = m.SelectRisk(p)
	}
}
