package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

func TestBrierBankReference(t *testing.T) {
	for _, prior := range [][]float64{{1}, {.95, .05}, {.95, .05 / 3, .05 / 3, .05 / 3}} {
		for _, rho := range []float64{0, .001} {
			s, e := newBrierBank(prior, rho)
			if e != nil {
				t.Fatal(e)
			}
			reference := s.prior
			rng := rand.New(rand.NewSource(2053119700))
			var losses [2]float64
			old, _ := newBrierAggregator(rho)
			for step := uint64(0); step < 2048; step++ {
				var p [4]float64
				for i := range prior {
					p[i] = rng.Float64()
				}
				f, e := s.predict(step, p[:len(prior)])
				if e != nil {
					t.Fatal(e)
				}
				want := 0.
				for i := range prior {
					want += reference[i] * p[i]
					if math.Abs(reference[i]-f.Weights[i]) > 1e-12 {
						t.Fatal("reference weights")
					}
				}
				if math.Abs(f.P-want) > 1e-12 {
					t.Fatal("reference forecast")
				}
				if len(prior) == 2 {
					g, e := old.predict(step, [2]float64{p[0], p[1]})
					if e != nil || math.Abs(g.P-f.P) > 1e-12 {
						t.Fatal("predecessor", e)
					}
				}
				y := rng.Float64() < .5
				v := 0.
				if y {
					v = 1
				}
				losses[0] += (p[0] - v) * (p[0] - v)
				losses[1] += (f.P - v) * (f.P - v)
				if losses[1]-losses[0] > s.genericBound(step+1)+1e-8 {
					t.Fatal("regret")
				}
				sum := 0.
				for i := range prior {
					reference[i] *= math.Exp(-.5 * (p[i] - v) * (p[i] - v))
					sum += reference[i]
				}
				for i := range prior {
					reference[i] = (1-rho)*reference[i]/sum + rho*s.prior[i]
				}
				if e = s.observe(step, y); e != nil {
					t.Fatal(e)
				}
				if len(prior) == 2 {
					if e = old.observe(step, y); e != nil {
						t.Fatal(e)
					}
				}
			}
		}
	}
}

func TestBrierBankLifecycle(t *testing.T) {
	for _, prior := range [][]float64{nil, {0}, {.8, .3}, {math.NaN()}, {math.Inf(1)}, {.2, .2, .2, .2, .2}} {
		if _, e := newBrierBank(prior, 0); e == nil {
			t.Fatal("invalid prior")
		}
	}
	for _, rho := range []float64{-1, 1, math.NaN(), math.Inf(1)} {
		if _, e := newBrierBank([]float64{1}, rho); e == nil {
			t.Fatal("invalid share")
		}
	}
	var zero brierBank
	if _, e := zero.predict(0, []float64{1}); e == nil {
		t.Fatal("zero")
	}
	s, _ := newBrierBank([]float64{.95, .05}, .001)
	before := *s
	for _, p := range [][]float64{{1}, {0, 0, 0}, {-1, 0}, {0, 2}, {math.NaN(), 0}} {
		if _, e := s.predict(0, p); e == nil || before != *s {
			t.Fatal("invalid mutation")
		}
	}
	if e := s.observe(0, true); e == nil || before != *s {
		t.Fatal("unissued")
	}
	p := []float64{.2, .8}
	f, e := s.predict(0, p)
	if e != nil {
		t.Fatal(e)
	}
	before = *s
	p[0] = 1
	f.Experts[0] = 0
	if before != *s {
		t.Fatal("alias")
	}
	if _, e = s.predict(1, p); e == nil || before != *s {
		t.Fatal("skipped")
	}
	if e = s.observe(1, true); e == nil || before != *s {
		t.Fatal("wrong id")
	}
	if e = s.observe(0, true); e != nil {
		t.Fatal(e)
	}
	before = *s
	if e = s.observe(0, true); e == nil || before != *s {
		t.Fatal("duplicate")
	}
	s.issued = math.MaxUint64
	before = *s
	if _, e = s.predict(math.MaxUint64, p); e == nil || before != *s {
		t.Fatal("overflow")
	}
	equal, _ := newBrierBank([]float64{.25, .25, .25, .25}, .001)
	for i := uint64(0); i < 16; i++ {
		f, e := equal.predict(i, []float64{.3, .3, .3, .3})
		if e != nil || math.Abs(f.P-.3) > 1e-14 {
			t.Fatal("equal", e)
		}
		if e = equal.observe(i, true); e != nil {
			t.Fatal(e)
		}
	}
}

func TestBrierBankDuplicatedChallenger(t *testing.T) {
	bank, _ := newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	old, _ := newBrierAggregator(.001)
	rng := rand.New(rand.NewSource(2053119701))
	for step := uint64(0); step < 1024; step++ {
		g, p := rng.Float64(), rng.Float64()
		a, e := bank.predict(step, []float64{g, p, p, p})
		if e != nil {
			t.Fatal(e)
		}
		b, e := old.predict(step, [2]float64{g, p})
		if e != nil || math.Abs(a.P-b.P) > 1e-12 {
			t.Fatal("duplicate challenger normalization", e)
		}
		y := rng.Float64() < .5
		if e = bank.observe(step, y); e != nil {
			t.Fatal(e)
		}
		if e = old.observe(step, y); e != nil {
			t.Fatal(e)
		}
	}
}

func BenchmarkBrierBank(b *testing.B) {
	s, _ := newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := s.predict(uint64(i), []float64{.2, .8, .3, .7}); e != nil {
			b.Fatal(e)
		}
		if e := s.observe(uint64(i), i%3 == 0); e != nil {
			b.Fatal(e)
		}
	}
}
