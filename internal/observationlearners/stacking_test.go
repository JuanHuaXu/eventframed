package observationlearners

import (
	"math"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func explicitStackingLOO(samples []observation.Sample) ([][2]float64, error) {
	result := make([][2]float64, len(samples))
	for i, sample := range samples {
		left := append([]observation.Sample(nil), samples[:i]...)
		left = append(left, samples[i+1:]...)
		g, err := fitSubset(left)
		if err != nil {
			return nil, err
		}
		p, err := fitBooleanSpecialist(left)
		if err != nil {
			return nil, err
		}
		result[i] = [2]float64{g.predictions[sample.Bits], p.predictions[sample.Bits]}
	}
	return result, nil
}

func TestStackingExactLOO(t *testing.T) {
	constant := booleanFixture(8)
	for i := range constant {
		constant[i].Outcome = false
	}
	fixtures := [][]observation.Sample{booleanFixture(2), booleanFixture(16), constant, {{Bits: 0, Outcome: true}, {Bits: 0, Outcome: false}, {Bits: 0, Outcome: true}, {Bits: 511, Outcome: false}}}
	for _, s := range fixtures {
		m, err := fitStackingEvidence(s)
		if err != nil {
			t.Fatal(err)
		}
		ref, err := explicitStackingLOO(s)
		if err != nil {
			t.Fatal(err)
		}
		for i := range s {
			for family := 0; family < 2; family++ {
				if math.Abs(m.loo[i][family]-ref[i][family]) > 1e-12 {
					t.Fatal("LOO refit mismatch", len(s), i, family, m.loo[i], ref[i])
				}
			}
		}
		// Changing only the excluded label must not alter its LOO predictive law.
		for i := range s {
			changed := append([]observation.Sample(nil), s...)
			changed[i].Outcome = !changed[i].Outcome
			other, err := fitStackingEvidence(changed)
			if err != nil {
				t.Fatal(err)
			}
			for family := 0; family < 2; family++ {
				if math.Abs(m.loo[i][family]-other.loo[i][family]) > 1e-12 {
					t.Fatal("held-out label leaked", i)
				}
			}
		}
	}
}

func TestStackingObjectiveAndMarginals(t *testing.T) {
	s := booleanFixture(32)
	f, err := fitStackingEvidence(s)
	if err != nil {
		t.Fatal(err)
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	g, _ := NewSubsetConditional(s, weights)
	p, _ := NewBooleanConditional(s)
	for _, lambda := range []float64{0, 1} {
		m, a, err := compileStacking(f, lambda)
		if err != nil {
			t.Fatal(err)
		}
		objective := func(a float64) float64 {
			v := lambda * a * a
			for i, qs := range f.loo {
				y := 0.
				if s[i].Outcome {
					y = 1
				}
				d := (1-a)*qs[0] + a*qs[1] - y
				v += d * d
			}
			return v
		}
		for i := 0; i <= 1000; i++ {
			if objective(a) > objective(float64(i)/1000)+1e-11 {
				t.Fatal("quadratic optimum")
			}
		}
		for mask := uint16(0); mask < 512; mask++ {
			for values := mask; ; values = (values - 1) & mask {
				actual, e := m.Forecast(mask, values)
				gp, _ := g.Forecast(mask, values)
				pp, _ := p.Forecast(mask, values)
				if e != nil || math.Abs(actual-((1-a)*gp+a*pp)) > 1e-12 {
					t.Fatal("partial mixture")
				}
				if values == 0 {
					break
				}
			}
		}
	}
	zero := &stackingEvidence{loo: make([][2]float64, 2)}
	for i := range zero.generic {
		zero.generic[i] = .5
		zero.parity[i] = .5
	}
	_, a, err := compileStacking(zero, 0)
	if err != nil || a != 0 {
		t.Fatal("zero disagreement", err)
	}
	for _, lambda := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, err := NewStackedConditional(s, lambda); err == nil {
			t.Fatal("invalid penalty")
		}
	}
	for _, bad := range [][]observation.Sample{nil, {{Bits: 0}}, make([]observation.Sample, 257), {{Bits: 512}, {Bits: 0}}} {
		if _, err := NewStackedConditional(bad, 1); err == nil {
			t.Fatal("invalid evidence")
		}
	}
	if _, err := NewStackedConditional(booleanFixture(256), 1); err != nil {
		t.Fatal(err)
	}
}

func TestStackingSymmetryAndSnapshot(t *testing.T) {
	s := booleanFixture(16)
	f, err := fitStackingEvidence(s)
	if err != nil {
		t.Fatal(err)
	}
	m, a, _ := compileStacking(f, 1)
	perm := func(x uint16) uint16 { return ((x << 3) | (x >> 6)) & 511 }
	for mode := 0; mode < 3; mode++ {
		x := append([]observation.Sample(nil), s...)
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
		other, err := fitStackingEvidence(x)
		if err != nil {
			t.Fatal(err)
		}
		q, oa, _ := compileStacking(other, 1)
		if math.Abs(a-oa) > 1e-12 {
			t.Fatal("weight symmetry")
		}
		for x := uint16(0); x < 512; x++ {
			v := x
			if mode == 2 {
				v = perm(x)
			}
			p, _ := m.Forecast(511, x)
			r, _ := q.Forecast(511, v)
			if mode == 1 {
				r = 1 - r
			}
			if math.Abs(p-r) > 1e-12 {
				t.Fatal("predictive symmetry")
			}
		}
	}
	before := *m
	f.generic[0] = 1
	s[0].Outcome = !s[0].Outcome
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := uint16(0); x < 512; x++ {
				if _, err := m.Forecast(511, x); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if before != *m {
		t.Fatal("snapshot mutation")
	}
}

func TestStackingSeedIsolation(t *testing.T) {
	seen := map[int64]bool{}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300} {
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

var stackingLOOSink [][2]float64

func BenchmarkStackingLOO(b *testing.B) {
	for _, n := range []int{16, 64} {
		for _, mode := range []string{"shortcut", "explicit_refits"} {
			b.Run(map[int]string{16: "n16", 64: "n64"}[n]+"/"+mode, func(b *testing.B) {
				s := booleanFixture(n)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if mode == "shortcut" {
						f, e := fitStackingEvidence(s)
						if e != nil {
							b.Fatal(e)
						}
						stackingLOOSink = f.loo
					} else {
						v, e := explicitStackingLOO(s)
						if e != nil {
							b.Fatal(e)
						}
						stackingLOOSink = v
					}
				}
			})
		}
	}
}
func BenchmarkStackingBuild(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		for _, arm := range []string{"skeptical", "stacking"} {
			b.Run(map[int]string{16: "n16", 64: "n64", 256: "n256"}[n]+"/"+arm, func(b *testing.B) {
				s := booleanFixture(n)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var m *ConditionalForest
					var e error
					if arm == "skeptical" {
						m, e = NewPriorConditional(s, .1)
					} else {
						m, e = NewStackedConditional(s, 1)
					}
					if e != nil {
						b.Fatal(e)
					}
					booleanBenchmarkSink = m
				}
			})
		}
	}
}
func BenchmarkStackingForecast(b *testing.B) {
	m, e := NewStackedConditional(booleanFixture(64), 1)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, e := m.Forecast(63, uint16(i)&63)
		if e != nil {
			b.Fatal(e)
		}
		booleanForecastSink = p
	}
}
