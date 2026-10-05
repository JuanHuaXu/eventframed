package observationgate

import (
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

// Force the observer branch for cost measurement only. These externally set
// weights are never used in the accuracy experiment or in ordinary prediction.
func BenchmarkSubsetActiveObserver(b *testing.B) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		b.Fatal(e)
	}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		x := uint16(i * 7 % 512)
		samples[i] = observation.Sample{Bits: x, Outcome: x&4 != 0}
	}
	trial := newSubsetTrial(base, 1)
	if e := trial.fit(samples); e != nil {
		b.Fatal(e)
	}
	count, e := observation.Fit(samples)
	if e != nil {
		b.Fatal(e)
	}
	models := observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}
	for _, enabled := range []bool{false, true} {
		b.Run(fmt.Sprintf("subset-%t", enabled), func(b *testing.B) {
			s := subsetState{base: base, enabled: enabled}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s.mix = bayes.ForecastMix{Weights: [4]float64{.01, .97, .01, .01}}
				s.inner = bayes.ForecastMix{Weights: [4]float64{.01, .33, .33, .33}}
				x := uint16(i * 13 % 512)
				rd := observationexperiment.Frames(x, "active-benchmark")
				if _, e := s.predict(rd, models, trial.model, i, 0); e != nil {
					b.Fatal(e)
				}
				if s.pending.subsetGuide != enabled {
					b.Fatal("observer path not exercised")
				}
				if e := s.observe(i, x&4 != 0, false); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
