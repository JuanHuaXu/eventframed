package observationpreserved

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

func TestSeedDomains(t *testing.T) {
	seen := map[int64]bool{FitSeed: true, 2026091401: true, 2026091301: true}
	for _, base := range []int64{2026091402, 2026091403, DesignSeed, ConfirmationSeed} {
		for j := range Scenarios {
			for i := 0; i < Streams; i++ {
				for role := 0; role < 4; role++ {
					s := Seed(base, j, i, role)
					if seen[s] {
						t.Fatalf("reused seed %d", s)
					}
					seen[s] = true
				}
			}
		}
	}
}
func TestSequentialEvidenceNullAndBothDirections(t *testing.T) {
	var same Evidence
	for i := 0; i < 512; i++ {
		if same.Observe(i%2 == 0, i%2 == 0) {
			t.Fatal("identical reliability rejected")
		}
	}
	for _, positive := range []bool{true, false} {
		var e Evidence
		rejected := false
		for i := 0; i < 512; i++ {
			ref, live := true, true
			if i >= 256 {
				ref, live = positive, !positive
			}
			rejected = rejected || e.Observe(ref, live)
		}
		if !rejected {
			t.Fatal("sustained divergence not detected")
		}
	}
}
func TestSharedObservationAndPreservedIncumbent(t *testing.T) {
	base, e := Base(false)
	if e != nil {
		t.Fatal(e)
	}
	s, _ := New(base, "mix_mmm_ap")
	reader := observationexperiment.Frames(511, "unit")
	before, _ := base.ForecastObserved(511, 511)
	p, e := s.Predict(reader, Models{}, 0, 7)
	if e != nil {
		t.Fatal(e)
	}
	if p.Cost > 6 {
		t.Fatal("foreground budget")
	}
	expected, e := base.ForecastObserved(p.Mask, p.Values)
	if e != nil || expected != p.Experts[0] {
		t.Fatal("experts not using recorded observations")
	}
	if _, e = s.Predict(reader, Models{}, 0, 7); e == nil {
		t.Fatal("duplicate prediction accepted")
	}
	if e = s.Observe(1, true, true); e == nil {
		t.Fatal("out of order feedback accepted")
	}
	want := (bayes.ForecastMix{}).Observe(p.Experts, true, 1).Weights
	if e = s.Observe(0, true, true); e != nil {
		t.Fatal(e)
	}
	if e = s.Observe(0, true, true); e == nil {
		t.Fatal("duplicate feedback accepted")
	}
	if !s.split || s.base != base {
		t.Fatal("split did not retain incumbent")
	}
	if math.Abs(s.mix.Weights[0]/s.mix.Weights[1]-want[0]/want[1]) > 1e-12 {
		t.Fatal("split reset unrelated influence")
	}
	after, _ := base.ForecastObserved(511, 511)
	if before != after || base.Support() != 4096 {
		t.Fatal("incumbent changed")
	}
	if _, e = base.ForecastObserved(1, 2); e == nil {
		t.Fatal("unobserved coordinate accepted")
	}
}
func TestCompleteStreamOrdering(t *testing.T) {
	base, e := Base(false)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunStream(base, "unit", 0, 0, 2026091805)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Ticks) != 512 || len(r.Arms) != 6 {
		t.Fatal("incomplete stream")
	}
	version := 0
	for i, tick := range r.Ticks {
		for _, arm := range r.Arms {
			p := arm.Predictions[i]
			if p.Version != version || p.Cost > 6 || p.Values&^p.Mask != 0 || p.P <= 0 || p.P >= 1 {
				t.Fatal("invalid prediction snapshot")
			}
		}
		if tick.Fit {
			if !tick.Audit {
				t.Fatal("fit without observed audit")
			}
			version++
		}
	}
}
func BenchmarkPreservedStep(b *testing.B) {
	base, e := Base(false)
	if e != nil {
		b.Fatal(e)
	}
	reader := observationexperiment.Frames(511, "bench")
	for _, arm := range []string{"frozen_mmm", "mix_breadth_ap", "mix_mmm_ap"} {
		b.Run(arm, func(b *testing.B) {
			s, _ := New(base, arm)
			models := Models{Short: base, Local: base, Pooled: base, Version: 1}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, e := s.Predict(reader, models, i, 0); e != nil {
					b.Fatal(e)
				}
				if e := s.Observe(i, true, false); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
func BenchmarkChallengerFit(b *testing.B) {
	samples := make([]observation.Sample, 256)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i), Outcome: i%2 == 0}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := observation.Fit(samples); e != nil {
			b.Fatal(e)
		}
	}
}
