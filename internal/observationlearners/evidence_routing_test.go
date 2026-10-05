package observationlearners

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestEvidenceRoutingReference(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		var s evidenceRouting
		s.gate.issued = 1
		p := [4]float64{.1, .3, .6, .9}
		w := [4]float64{.7, .1, .1, .1}
		var direct [5]float64
		for i := 0; i < 4; i++ {
			s.gate.tests[i].Rejected = mask&(1<<i) != 0
			for j := 0; j < 5; j++ {
				s.credits[i][j] = math.Log(float64(1 + i + j))
			}
			if mask&(1<<i) == 0 {
				direct[i+1] += w[i]
				continue
			}
			sum := 0.
			for j := 0; j < 5; j++ {
				if j == 0 || mask&(1<<(j-1)) == 0 {
					sum += float64(1 + i + j)
				}
			}
			for j := 0; j < 5; j++ {
				if j == 0 || mask&(1<<(j-1)) == 0 {
					direct[j] += w[i] * float64(1+i+j) / sum
				}
			}
		}
		f, err := s.predict(1, p, w)
		if err != nil {
			t.Fatal(err)
		}
		sum, want := 0., .5*direct[0]
		for j, v := range f.Weights {
			if math.Abs(v-direct[j]) > 1e-12 {
				t.Fatal("routing", mask, j, v, direct[j])
			}
			sum += v
			if j > 0 {
				want += direct[j] * p[j-1]
				if mask&(1<<(j-1)) != 0 && v != 0 {
					t.Fatal("rejected recipient")
				}
			}
		}
		if math.Abs(sum-1) > 1e-12 || math.Abs(f.P-want) > 1e-12 {
			t.Fatal("convexity")
		}
		if f.Neutral != (mask == 15) {
			t.Fatal("fallback")
		}
	}
}

func TestEvidenceRoutingLifecycle(t *testing.T) {
	var s evidenceRouting
	var control comparativeFalsification
	rng := rand.New(rand.NewSource(2074110000))
	w := [4]float64{.7, .1, .1, .1}
	for step := uint64(0); step < 256; step++ {
		p := [4]float64{.99, .98, .05 + .9*rng.Float64(), .01}
		f, err := s.predict(step, p, w)
		if err != nil {
			t.Fatal(err)
		}
		c, err := control.predict(step, p, w)
		if err != nil || c.Accepted != f.Accepted {
			t.Fatal("changed mask")
		}
		before := s
		if _, err = s.predict(step+1, p, w); err == nil || !reflect.DeepEqual(before, s) {
			t.Fatal("double issuance")
		}
		if err = s.observe(step+1, false); err == nil || !reflect.DeepEqual(before, s) {
			t.Fatal("wrong feedback")
		}
		if step%32 == 0 && f.Weights[0] != 0 {
			t.Fatal("credits survived publication")
		}
		y := rng.Float64() < .2
		if err = s.observe(step, y); err != nil {
			t.Fatal(err)
		}
		if err = control.observe(step, y); err != nil {
			t.Fatal(err)
		}
		if s.gate != control {
			t.Fatal("changed test or raw pending state")
		}
		for i, m := range s.gate.tests {
			if before.gate.tests[i].Rejected && s.credits[i] != before.credits[i] {
				t.Fatal("credits not frozen")
			}
			if m.Rejected && !before.gate.tests[i].Rejected {
				if !math.IsInf(s.credits[i][i+1], -1) {
					t.Fatal("self credit")
				}
				alt := 0
				total := 0.
				var direct [5]float64
				for j := 0; j < 5; j++ {
					if j == i+1 {
						continue
					}
					mass := 1. / 6
					if j == 0 {
						mass = .5
					}
					direct[j] = mass * (math.Exp(m.LogActive[alt]) + float64(32-m.Count)) / 32
					total += direct[j]
					alt++
				}
				for j, v := range direct {
					if math.Abs(math.Exp(s.credits[i][j])-v/total) > 1e-11 {
						t.Fatal("credit mapping")
					}
				}
			}
		}
		before = s
		if err = s.observe(step, y); err == nil || !reflect.DeepEqual(before, s) {
			t.Fatal("duplicate feedback")
		}
	}
	before := s
	if _, err := s.predict(256, [4]float64{.5, .5, .5, .5}, w); err == nil || !reflect.DeepEqual(before, s) {
		t.Fatal("horizon")
	}
}

func TestEvidenceRoutingInvalid(t *testing.T) {
	var s evidenceRouting
	s.gate.issued = 32
	s.credits[0][0] = 7
	before := s
	if _, err := s.predict(32, [4]float64{0, .5, .5, .5}, [4]float64{.25, .25, .25, .25}); err == nil || !reflect.DeepEqual(before, s) {
		t.Fatal("invalid reset")
	}
	var nilState *evidenceRouting
	if _, err := nilState.predict(0, [4]float64{}, [4]float64{}); err == nil {
		t.Fatal("nil predict")
	}
	if err := nilState.observe(0, false); err == nil {
		t.Fatal("nil observe")
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		s = evidenceRouting{}
		s.gate.issued = 1
		s.gate.tests[0].Rejected = true
		for j := range s.credits[0] {
			s.credits[0][j] = bad
		}
		f, err := s.predict(1, [4]float64{.1, .2, .3, .4}, [4]float64{1, 0, 0, 0})
		if err != nil || f.P != .5 || !f.Neutral {
			t.Fatal("unsafe normalization")
		}
	}
}

func BenchmarkEvidenceRouting(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s evidenceRouting
		for step := uint64(0); step < 256; step++ {
			if _, err := s.predict(step, [4]float64{.2, .8, .3, .7}, [4]float64{.25, .25, .25, .25}); err != nil {
				b.Fatal(err)
			}
			if err := s.observe(step, step%3 == 0); err != nil {
				b.Fatal(err)
			}
		}
	}
}
