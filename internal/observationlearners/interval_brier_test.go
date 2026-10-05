package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

func TestIntervalBrierReference(t *testing.T) {
	// Enumerate every interval literally; unlike the implementation this holds
	// all intervals and uses direct probability updates rather than log weights.
	type ref struct {
		start, end   int
		rate, weight float64
		w            [2]float64
		p            float64
	}
	var refs []ref
	for length := 1; length <= 512; length *= 2 {
		for start := length; start <= 512; start += length {
			eta := math.Min(.5, 1/math.Sqrt(float64(length)))
			refs = append(refs, ref{start: start, end: start + length - 1, rate: eta, weight: eta, w: [2]float64{.95, .05}})
		}
	}
	s, _ := newIntervalBrier(512)
	rng := rand.New(rand.NewSource(2049119600))
	for t0 := 1; t0 <= 512; t0++ {
		p := [2]float64{rng.Float64(), rng.Float64()}
		if t0%7 == 0 {
			p = [2]float64{0, 1}
		}
		actual, e := s.predict(uint64(t0-1), p)
		if e != nil {
			t.Fatal(e)
		}
		total, want := 0., 0.
		count := 0
		for i := range refs {
			r := &refs[i]
			if r.start <= t0 && t0 <= r.end {
				r.p = r.w[0]*p[0] + r.w[1]*p[1]
				total += r.weight
				want += r.weight * r.p
				count++
			}
		}
		want /= total
		if math.Abs(want-actual) > 1e-12 || count != s.count {
			t.Fatal("reference prediction", t0, want, actual, count, s.count)
		}
		y := rng.Float64() < .5
		target := 0.
		if y {
			target = 1
		}
		surrogate := 0.
		for _, r := range refs {
			if r.start <= t0 && t0 <= r.end {
				surrogate += r.weight * (r.p - target) * (r.p - target) / total
			}
		}
		if (actual-target)*(actual-target) > surrogate+1e-12 {
			t.Fatal("Jensen")
		}
		for i := range refs {
			r := &refs[i]
			if r.start <= t0 && t0 <= r.end {
				r.weight *= 1 + r.rate*(surrogate-(r.p-target)*(r.p-target))
				z := 0.
				for k := range r.w {
					r.w[k] *= math.Exp(-.5 * (p[k] - target) * (p[k] - target))
					z += r.w[k]
				}
				for k := range r.w {
					r.w[k] /= z
				}
			}
		}
		if e = s.observe(uint64(t0-1), y); e != nil {
			t.Fatal(e)
		}
		for i := 0; i < s.count; i++ {
			if s.slots[i].weight <= 0 || math.IsInf(s.slots[i].weight, 0) || math.IsNaN(s.slots[i].weight) {
				t.Fatal("weight")
			}
		}
	}
	before := *s
	if _, e := s.predict(512, [2]float64{}); e == nil || before != *s {
		t.Fatal("horizon")
	}
}

func TestIntervalBrierLifecycle(t *testing.T) {
	for _, n := range []uint64{0, 513, math.MaxUint64} {
		if _, e := newIntervalBrier(n); e == nil {
			t.Fatal("horizon accepted")
		}
	}
	var zero intervalBrier
	if _, e := zero.predict(0, [2]float64{}); e == nil {
		t.Fatal("zero state")
	}
	s, _ := newIntervalBrier(8)
	before := *s
	for _, p := range [][2]float64{{math.NaN(), 0}, {0, math.Inf(1)}, {-1, 0}, {0, 2}} {
		if _, e := s.predict(0, p); e == nil || before != *s {
			t.Fatal("invalid mutation")
		}
	}
	if e := s.observe(0, true); e == nil || before != *s {
		t.Fatal("unissued")
	}
	if _, e := s.predict(1, [2]float64{}); e == nil || before != *s {
		t.Fatal("sequence")
	}
	if _, e := s.predict(0, [2]float64{.2, .8}); e != nil {
		t.Fatal(e)
	}
	before = *s
	if _, e := s.predict(1, [2]float64{}); e == nil || before != *s {
		t.Fatal("skipped feedback")
	}
	if e := s.observe(1, true); e == nil || before != *s {
		t.Fatal("wrong feedback")
	}
	if e := s.observe(0, true); e != nil {
		t.Fatal(e)
	}
	before = *s
	if e := s.observe(0, true); e == nil || before != *s {
		t.Fatal("duplicate")
	}
}

func BenchmarkIntervalBrier(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s, _ := newIntervalBrier(256)
		for t := uint64(0); t < 256; t++ {
			if _, e := s.predict(t, [2]float64{.2, .8}); e != nil {
				b.Fatal(e)
			}
			if e := s.observe(t, t%3 == 0); e != nil {
				b.Fatal(e)
			}
		}
	}
}
