package observationgate

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

var availableEvidenceBenchmarkSink float64

func BenchmarkAvailableEvidenceForecast(b *testing.B) {
	base, _, err := arrivalSwitchSetup(1, 3, 0)
	if err != nil {
		b.Fatal(err)
	}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 37 % 512), Outcome: i%3 == 0}
	}
	count, err := observation.Fit(samples)
	if err != nil {
		b.Fatal(err)
	}
	trial := newSubsetTrial(base, 1)
	if err := trial.fit(samples); err != nil {
		b.Fatal(err)
	}
	j := &subsetDelayJournal{state: subsetState{base: base, enabled: true}}
	if err := j.publish(observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}, trial.model); err != nil {
		b.Fatal(err)
	}
	p, err := j.predict(0, observationexperiment.Frames(137, "bench"), 0)
	if err != nil {
		b.Fatal(err)
	}
	for _, full := range []bool{false, true} {
		name := "requested"
		mask, values := p.Mask, p.Values
		if full {
			name = "available"
			mask, values = 511, 137
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var sum float64
			for i := 0; i < b.N; i++ {
				q, _, e := availableEvidenceForecast(j, 0, mask, values)
				if e != nil {
					b.Fatal(e)
				}
				sum += q
			}
			availableEvidenceBenchmarkSink = sum
		})
	}
}
