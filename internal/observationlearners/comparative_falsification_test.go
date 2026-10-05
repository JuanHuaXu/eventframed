package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

func TestComparativeFalsificationReference(t *testing.T) {
	for seed := int64(0); seed < 32; seed++ {
		rng := rand.New(rand.NewSource(2070110000 + seed))
		var s comparativeTest
		var direct [4][32]float64
		for j := range direct {
			for k := range direct[j] {
				direct[j][k] = 1
			}
		}
		for step := 0; step < 32; step++ {
			p := .05 + .9*rng.Float64()
			q := [4]float64{.5, .05 + .9*rng.Float64(), .05 + .9*rng.Float64(), .05 + .9*rng.Float64()}
			y := rng.Float64() < .5
			probability := func(v float64) float64 {
				if y {
					return v
				}
				return 1 - v
			}
			want := 0.
			for j := range q {
				for k := 0; k <= step; k++ {
					direct[j][k] *= probability(q[j]) / probability(p)
				}
				mass := 1. / 6
				if j == 0 {
					mass = .5
				}
				for _, v := range direct[j] {
					want += mass * v / 32
				}
			}
			got := s.update(p, q, y)
			if math.Abs(got-math.Log(want)) > 1e-11 {
				t.Fatal("literal", step, got, math.Log(want))
			}
		}
	}
	for _, p := range []float64{math.SmallestNonzeroFloat64, .01, .2, .5, .9, math.Nextafter(1, 0)} {
		var a, b comparativeTest
		q := [4]float64{.5, .1, .6, .9}
		trueLog := a.update(p, q, true)
		falseLog := b.update(p, q, false)
		expectation := math.Exp(math.Log(p)+trueLog) + math.Exp(math.Log1p(-p)+falseLog)
		if math.Abs(expectation-1) > 1e-12 {
			t.Fatal("null expectation", p, expectation)
		}
	}
	var neutral forecastTest
	var reduced, identical comparativeTest
	for step := 0; step < 32; step++ {
		p := .8
		y := step%3 == 0
		old := neutral.update(p, y)
		if math.Abs(reduced.update(p, [4]float64{.5, .5, .5, .5}, y)-old) > 1e-11 {
			t.Fatal("neutral reduction")
		}
		want := falsificationLogAdd(old-math.Ln2, -math.Ln2)
		if math.Abs(identical.update(p, [4]float64{.5, p, p, p}, y)-want) > 1e-11 {
			t.Fatal("identical alternatives")
		}
	}
}

func TestComparativeFalsificationLifecycle(t *testing.T) {
	var s comparativeFalsification
	p := [4]float64{.99, .99, .99, .99}
	w := [4]float64{.25, .25, .25, .25}
	before := s
	for _, bad := range []float64{0, 1, math.NaN(), math.Inf(1)} {
		q := p
		q[0] = bad
		if _, e := s.predict(0, q, w); e == nil || s != before {
			t.Fatal("invalid mutation")
		}
	}
	if e := s.observe(0, false); e == nil || s != before {
		t.Fatal("unissued")
	}
	for step := uint64(0); step < 33; step++ {
		f, e := s.predict(step, p, w)
		if e != nil {
			t.Fatal(e)
		}
		if step == 3 && f.Neutral {
			t.Fatal("future evidence")
		}
		if step >= 4 && step < 32 && !f.Neutral {
			t.Fatal("no rejection")
		}
		if step == 32 && f.Neutral {
			t.Fatal("no reentry")
		}
		before = s
		f.Raw[0] = .01
		if _, e = s.predict(step+1, p, w); e == nil || before != s {
			t.Fatal("missing outcome")
		}
		if e = s.observe(step+1, false); e == nil || before != s {
			t.Fatal("wrong id")
		}
		if e = s.observe(step, false); e != nil {
			t.Fatal(e)
		}
		before = s
		if e = s.observe(step, false); e == nil || before != s {
			t.Fatal("duplicate")
		}
	}
	s.issued = 256
	before = s
	if _, e := s.predict(256, p, w); e == nil || before != s {
		t.Fatal("budget")
	}
}

func BenchmarkComparativeFalsification(b *testing.B) {
	p := [4]float64{.2, .8, .3, .7}
	w := [4]float64{.25, .25, .25, .25}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s comparativeFalsification
		for step := uint64(0); step < 256; step++ {
			if _, e := s.predict(step, p, w); e != nil {
				b.Fatal(e)
			}
			if e := s.observe(step, step%3 == 0); e != nil {
				b.Fatal(e)
			}
		}
	}
}

// Explore both outcomes at every node, with forecasts chosen before the outcome.
// This checks conditional fairness with adaptive alternatives, not just fixed Q.
func TestComparativeFalsificationAdaptiveNull(t *testing.T) {
	var visit func(comparativeTest, int, uint64)
	visit = func(parent comparativeTest, depth int, history uint64) {
		if depth == 9 {
			return
		}
		p := .1 + .1*float64((history+uint64(depth))%9)
		q := [4]float64{.5, 1 - p, .1 + .1*float64(history%9), p}
		a, z := parent, parent
		ly := a.update(p, q, true)
		ln := z.update(p, q, false)
		got := falsificationLogAdd(math.Log(p)+ly, math.Log1p(-p)+ln)
		want := 0.
		if depth > 0 {
			want = math.Inf(-1)
			for j, active := range parent.LogActive {
				mass := -math.Log(6)
				if j == 0 {
					mass = -math.Ln2
				}
				want = falsificationLogAdd(want, mass+falsificationLogAdd(active, math.Log(float64(32-depth)))-math.Log(32))
			}
		}
		if math.Abs(got-want) > 1e-11 {
			t.Fatalf("conditional expectation at depth %d history %d: %g != %g", depth, history, got, want)
		}
		visit(a, depth+1, history<<1|1)
		visit(z, depth+1, history<<1)
	}
	visit(comparativeTest{}, 0, 0)
}

func TestComparativeFalsificationValidationAtomicity(t *testing.T) {
	p := [4]float64{.2, .4, .6, .8}
	w := [4]float64{.25, .25, .25, .25}
	var s comparativeFalsification
	for step := uint64(0); step < 32; step++ {
		if _, err := s.predict(step, p, w); err != nil {
			t.Fatal(err)
		}
		if err := s.observe(step, false); err != nil {
			t.Fatal(err)
		}
	}
	before := s
	for _, bad := range []float64{-1, 1.1, math.NaN(), math.Inf(1), 0} {
		weights := w
		weights[0] = bad
		if _, err := s.predict(32, p, weights); err == nil || s != before {
			t.Fatal("invalid weights reset publication state")
		}
	}
	if _, err := s.predict(33, p, w); err == nil || s != before {
		t.Fatal("out-of-order publication mutated state")
	}
	var absent *comparativeFalsification
	if _, err := absent.predict(0, p, w); err == nil {
		t.Fatal("nil issuance accepted")
	}
	if err := absent.observe(0, false); err == nil {
		t.Fatal("nil feedback accepted")
	}
}
