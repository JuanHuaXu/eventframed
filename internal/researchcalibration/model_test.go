package researchcalibration

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func fixtureBase(n int) []float64 {
	b := make([]float64, n)
	for i := range b {
		b[i] = .925 - .5*float64(i)/float64(n-1)
	}
	return b
}

func TestPriorIdentityAndOwnership(t *testing.T) {
	for _, n := range []int{2, 150, 200} {
		b := fixtureBase(n)
		m, err := New(b)
		if err != nil {
			t.Fatal(err)
		}
		for i, want := range b {
			q, _ := m.Predict(i)
			if math.Abs(q-want) > 1e-14 {
				t.Fatalf("n=%d i=%d: %g != %g", n, i, q, want)
			}
		}
		b[0] = .25
		q, _ := m.Predict(0)
		if math.Abs(q-.925) > 1e-14 {
			t.Fatal("input aliases state")
		}
	}
	for _, b := range [][]float64{nil, {.5}, make([]float64, 201), {.24, .5}, {.926, .5}, {math.NaN(), .5}, {math.Inf(1), .5}} {
		if _, err := New(b); err == nil {
			t.Fatal("accepted invalid frontier")
		}
	}
}

// Enumerate the joint theta/evidence distribution independently of Observe.
// For the observed event, E[phi^2]/E[phi]=(2*p+1)/3 after a positive label.
func TestEvidenceAndOutcomeAreOneJointModel(t *testing.T) {
	m, _ := New(fixtureBase(12))
	prior := m.weights
	if err := m.Observe(0, true); err != nil {
		t.Fatal(err)
	}
	if err := m.Observe(11, false); err != nil {
		t.Fatal(err)
	}
	z := 0.0
	var weights [Hypotheses]float64
	for h, w := range prior {
		weights[h] = w * m.p[h] * (1 - m.p[11*Hypotheses+h])
		z += weights[h]
	}
	for h := range weights {
		weights[h] /= z
		if math.Abs(weights[h]-m.weights[h]) > 1e-14 {
			t.Fatal("joint likelihood mismatch")
		}
	}
	for _, i := range []int{0, 5, 11} {
		want := 0.0
		for h, w := range weights {
			p := m.p[i*Hypotheses+h]
			if i == 0 {
				p = (2*p + 1) / 3
			}
			if i == 11 {
				p = 2 * p / 3
			}
			want += w * p
		}
		q, _ := m.Predict(i)
		if math.Abs(q-want) > 1e-14 {
			t.Fatalf("joint predictive mismatch: %g %g", q, want)
		}
	}
}

func TestRejectedEvidenceDoesNotMutate(t *testing.T) {
	m, _ := New(fixtureBase(8))
	_ = m.Observe(2, true)
	before := *m
	before.seen = append([]bool(nil), m.seen...)
	before.useful = append([]bool(nil), m.useful...)
	for _, i := range []int{-1, 8, 2} {
		if m.Observe(i, false) == nil {
			t.Fatal("accepted bad evidence")
		}
		if !reflect.DeepEqual(before, *m) {
			t.Fatal("rejection changed state")
		}
	}
	if _, err := m.Predict(8); err == nil {
		t.Fatal("accepted unknown prediction")
	}
	if _, _, err := m.Select("random", nil); err == nil {
		t.Fatal("missing RNG accepted")
	}
	if _, _, err := m.Select("other", nil); err == nil {
		t.Fatal("unknown policy accepted")
	}
}

func TestSelectorsAndInformation(t *testing.T) {
	for _, policy := range []string{"head", "random", "stratified", "uncertainty", "information"} {
		m, _ := New(fixtureBase(150))
		rng := rand.New(rand.NewSource(73091))
		seen := map[int]bool{}
		for k := 0; k < 150; k++ {
			i, score, err := m.Select(policy, rng)
			if err != nil || seen[i] {
				t.Fatalf("bad selection %s %d %v", policy, i, err)
			}
			if policy == "information" {
				q, _ := m.Predict(i)
				want := Entropy(q)
				for h, w := range m.weights {
					want -= w * Entropy(m.p[i*Hypotheses+h])
				}
				if score < 0 || math.Abs(score-math.Max(0, want)) > 1e-13 {
					t.Fatal("information not conditional entropy")
				}
			}
			seen[i] = true
			_ = m.Observe(i, k%3 == 0)
			if policy == "head" && i != k {
				t.Fatal("head order changed")
			}
			if policy == "stratified" && k == 1 && i != 128 {
				t.Fatal("bit reversal changed")
			}
		}
		if _, _, err := m.Select(policy, rng); err == nil {
			t.Fatal("exhaustion accepted")
		}
		for i := 0; i < 150; i++ {
			q, _ := m.Predict(i)
			if math.IsNaN(q) || q < 0 || q > 1 {
				t.Fatal("invalid final forecast")
			}
		}
	}
}

func BenchmarkChallenger(b *testing.B) {
	for _, policy := range []string{"head", "stratified", "uncertainty", "information"} {
		b.Run(policy, func(b *testing.B) {
			m, _ := New(fixtureBase(150))
			b.ReportAllocs()
			b.ResetTimer()
			for k := 0; k < b.N; k++ {
				_, _, _ = m.Select(policy, nil)
			}
		})
	}
	b.Run("construct", func(b *testing.B) {
		base := fixtureBase(150)
		b.ReportAllocs()
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			_, _ = New(base)
		}
	})
}
