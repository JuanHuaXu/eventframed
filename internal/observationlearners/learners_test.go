package observationlearners

import (
	"math/rand"
	"testing"
)

func TestWindowStationaryAndShift(t *testing.T) {
	var w Window
	for i := 0; i < 512; i++ {
		if w.Add(.05, i) {
			t.Fatal("constant loss cut")
		}
	}
	cut := false
	for i := 512; i < 768; i++ {
		if w.Add(.9, i) {
			cut = true
		}
	}
	if !cut || len(w.Points) > 256 || w.Cutoff < 256 {
		t.Fatal("missed change or unbounded state")
	}
}
func TestForestLearnsSimpleRuleAndIsBounded(t *testing.T) {
	f := NewForest(71)
	r := rand.New(rand.NewSource(72))
	for i := 0; i < 1024; i++ {
		x := uint16(r.Intn(512))
		f.Update(x, x&4 != 0)
	}
	correct := 0
	for x := uint16(0); x < 512; x++ {
		p := f.Predict(x)
		if p <= 0 || p >= 1 {
			t.Fatal("unbounded probability")
		}
		if (p >= .5) == (x&4 != 0) {
			correct++
		}
	}
	if correct < 460 || f.Nodes() > 155 {
		t.Fatalf("simple-rule control or size failed: correct=%d nodes=%d", correct, f.Nodes())
	}
}
func TestSeedDomains(t *testing.T) {
	seen := map[int64]bool{}
	for j := range Scenarios {
		for fit := 0; fit < 3; fit++ {
			s := FitBase*1000000 + int64(j)*1000 + int64(fit)
			if seen[s] {
				t.Fatal("fit reuse")
			}
			seen[s] = true
			for _, base := range []int64{DesignBase, ConfirmationBase} {
				for stream := 0; stream < 8; stream++ {
					for role := 0; role < 4; role++ {
						s := Seed(base, j, fit, stream, role)
						if seen[s] {
							t.Fatal("seed reuse")
						}
						seen[s] = true
					}
				}
			}
		}
	}
}
func TestDelayedMissingBoundary(t *testing.T) {
	s := Scenarios[7]
	base, e := Base(s, 7, 0)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunStream(base, "unit", 7, 0, 0, 2026092405)
	if e != nil {
		t.Fatal(e)
	}
	seen := map[int]bool{}
	audits := 0
	for t0, tick := range r.Ticks {
		for _, origin := range tick.Delivered {
			if seen[origin] || origin+s.Delay != t0 || r.Ticks[origin].Missing {
				t.Fatal("feedback availability or replay violation")
			}
			seen[origin] = true
			if r.Ticks[origin].Audit {
				audits++
			}
		}
		if tick.Audits != audits {
			t.Fatal("missing labels reached training")
		}
	}
	if r.Available != len(seen) || r.Audits != audits || r.Pending > 16 || r.Nodes > 155 {
		t.Fatal("accounting mismatch")
	}
}

var benchmarkPrediction float64

func BenchmarkForestPredict(b *testing.B) {
	f := NewForest(71)
	for i := 0; i < 1024; i++ {
		f.Update(uint16(i%512), i%4 == 0)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var p float64
	for i := 0; i < b.N; i++ {
		p = f.Predict(uint16(i % 512))
	}
	benchmarkPrediction = p
}
func BenchmarkForestUpdate(b *testing.B) {
	f := NewForest(71)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Update(uint16(i%512), i%4 == 0)
	}
}
func BenchmarkAdaptiveWindow(b *testing.B) {
	var w Window
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Add(float64(i%2), i)
	}
}
