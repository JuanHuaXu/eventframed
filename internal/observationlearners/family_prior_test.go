package observationlearners

import (
	"math"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestPriorContracts(t *testing.T) {
	s := booleanFixture(32)
	equal, ew, err := fitFamilyConditional(s)
	if err != nil {
		t.Fatal(err)
	}
	var uniform [512]float64
	for i := range uniform {
		uniform[i] = 1
	}
	g, _ := NewSubsetConditional(s, uniform)
	p, _ := NewBooleanConditional(s)
	last := -1.
	for _, prior := range []float64{0, .1, .5, 1} {
		m, w, err := fitPriorConditional(s, prior)
		if err != nil {
			t.Fatal(err)
		}
		wantWeight := prior * ew[1] / ((1-prior)*ew[0] + prior*ew[1])
		if math.Abs(w[1]-wantWeight) > 1e-12 || math.Abs(w[0]+w[1]-1) > 1e-12 || w[1] < last {
			t.Fatal("posterior odds")
		}
		last = w[1]
		for mask := uint16(0); mask < 512; mask++ {
			for values := mask; ; values = (values - 1) & mask {
				actual, err := m.Forecast(mask, values)
				gp, _ := g.Forecast(mask, values)
				pp, _ := p.Forecast(mask, values)
				want := (1-wantWeight)*gp + wantWeight*pp
				if prior == .5 {
					want, _ = equal.Forecast(mask, values)
				}
				if err != nil || math.Abs(actual-want) > 1e-12 {
					t.Fatal("partial/endpoint equivalence", prior, mask, values)
				}
				if values == 0 {
					break
				}
			}
		}
	}
	for _, prior := range []float64{-1, 1.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := NewPriorConditional(s, prior); err == nil {
			t.Fatal("invalid prior accepted")
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}} {
		if _, err := NewPriorConditional(bad, .1); err == nil {
			t.Fatal("invalid evidence accepted")
		}
	}
	if _, err := NewPriorConditional(booleanFixture(256), .1); err != nil {
		t.Fatal("valid cap rejected", err)
	}
}

func TestPriorSymmetryAndSnapshot(t *testing.T) {
	s := booleanFixture(64)
	m, w, err := fitPriorConditional(s, .1)
	if err != nil {
		t.Fatal(err)
	}
	for mode := 0; mode < 3; mode++ {
		x := append([]observation.Sample(nil), s...)
		perm := func(v uint16) uint16 { return ((v << 3) | (v >> 6)) & 511 }
		for i := range x {
			if mode == 0 {
				x[i] = s[len(s)-1-i]
			}
			if mode == 1 {
				x[i].Outcome = !x[i].Outcome
			}
			if mode == 2 {
				x[i].Bits = perm(x[i].Bits)
			}
		}
		other, ow, e := fitPriorConditional(x, .1)
		if e != nil || math.Abs(w[1]-ow[1]) > 1e-12 {
			t.Fatal("weight symmetry", e)
		}
		for x := uint16(0); x < 512; x++ {
			values := x
			if mode == 2 {
				values = perm(x)
			}
			p, _ := m.Forecast(511, x)
			q, _ := other.Forecast(511, values)
			if mode == 1 {
				q = 1 - q
			}
			if math.Abs(p-q) > 1e-12 {
				t.Fatal("prediction symmetry")
			}
		}
	}
	before := *m
	for i := range s {
		s[i].Outcome = !s[i].Outcome
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := uint16(0); x < 512; x++ {
				if _, e := m.Forecast(511, x); e != nil {
					t.Error(e)
				}
			}
		}()
	}
	wg.Wait()
	if before != *m {
		t.Fatal("mutable snapshot")
	}
}

func TestPriorSeedIsolation(t *testing.T) {
	seen := map[int64]bool{}
	for _, base := range []int64{2026119000, 2030119100, 2034119200} {
		for phase := int64(0); phase < 2; phase++ {
			for scenario := int64(0); scenario < 10; scenario++ {
				for index := int64(0); index < 64; index++ {
					for role := int64(0); role < 3; role++ {
						seed := base + phase*1000000 + scenario*10000 + index*10 + role
						if seen[seed] {
							t.Fatal("reused seed", seed)
						}
						seen[seed] = true
					}
				}
			}
		}
	}
}

func BenchmarkPriorComparison(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		for _, arm := range []string{"equal", "skeptical"} {
			name := map[int]string{16: "n16", 64: "n64", 256: "n256"}[n] + "/" + arm
			b.Run(name, func(b *testing.B) {
				s := booleanFixture(n)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var m *ConditionalForest
					var err error
					if arm == "equal" {
						m, err = NewFamilyConditional(s)
					} else {
						m, err = NewPriorConditional(s, .1)
					}
					if err != nil {
						b.Fatal(err)
					}
					booleanBenchmarkSink = m
				}
			})
		}
	}
}

func BenchmarkPriorForecast(b *testing.B) {
	m, err := NewPriorConditional(booleanFixture(64), .1)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := m.Forecast(63, uint16(i)&63)
		if err != nil {
			b.Fatal(err)
		}
		booleanForecastSink = p
	}
}
