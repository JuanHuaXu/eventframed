package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

func TestForecastFalsificationReference(t *testing.T) {
	for seed := int64(0); seed < 32; seed++ {
		var s forecastTest
		var direct [32]float64
		for i := range direct {
			direct[i] = 1
		}
		rng := rand.New(rand.NewSource(2058119900 + seed))
		for step := 0; step < 32; step++ {
			p := .05 + .9*rng.Float64()
			y := rng.Float64() < .5
			prob := 1 - p
			if y {
				prob = p
			}
			for start := 0; start <= step; start++ {
				direct[start] *= .5 / prob
			}
			total := 0.
			for _, v := range direct {
				total += v
			}
			want := math.Log(total / 32)
			got := s.update(p, y)
			if math.Abs(got-want) > 1e-11 {
				t.Fatal("mixture", step, got, want)
			}
		}
	}
	var neutral forecastTest
	for i := 0; i < 32; i++ {
		if math.Abs(neutral.update(.5, i%2 == 0)) > 1e-12 || neutral.Rejected {
			t.Fatal("neutral")
		}
	}
	var extreme forecastTest
	for i := 0; i < 32; i++ {
		v := extreme.update(math.SmallestNonzeroFloat64, true)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatal("log stability")
		}
	}
	if !extreme.Rejected {
		t.Fatal("extreme evidence")
	}
}

func TestForecastFalsificationLifecycle(t *testing.T) {
	var s forecastFalsification
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
	if _, e := s.predict(0, p, [4]float64{}); e == nil || s != before {
		t.Fatal("weights")
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
			t.Fatal("revealing outcome leaked into forecast")
		}
		if step >= 4 && step < 32 && !f.Neutral {
			t.Fatal("rejection missing")
		}
		if step == 32 && (f.Neutral || !f.Accepted[0]) {
			t.Fatal("new version not reopened")
		}
		before = s
		f.Raw[0] = .01
		if _, e = s.predict(step+1, p, w); e == nil || s != before {
			t.Fatal("missing feedback")
		}
		if e = s.observe(step+1, false); e == nil || s != before {
			t.Fatal("wrong id")
		}
		if e = s.observe(step, false); e != nil {
			t.Fatal(e)
		}
		before = s
		if e = s.observe(step, false); e == nil || s != before {
			t.Fatal("duplicate")
		}
	}
	s.issued = 256
	before = s
	if _, e := s.predict(256, p, w); e == nil || before != s {
		t.Fatal("budget reset")
	}
}

func BenchmarkForecastFalsification(b *testing.B) {
	p := [4]float64{.2, .8, .3, .7}
	w := [4]float64{.25, .25, .25, .25}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s forecastFalsification
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
