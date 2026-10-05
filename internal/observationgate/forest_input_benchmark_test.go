package observationgate

import (
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"testing"
)

func BenchmarkForestIntegrationFit(b *testing.B) {
	samples := make([]observation.Sample, 64)
	for i := range samples {
		x := uint16(i * 7 % 512)
		x = (x &^ 1) | ((x >> 2) & 1)
		samples[i] = observation.Sample{Bits: x, Outcome: x&4 != 0}
	}
	for _, forest := range []bool{false, true} {
		name := "uniform"
		if forest {
			name = "forest"
		}
		b.Run(name, func(b *testing.B) {
			trial := new(subsetTrial)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if forest {
					err = forestInputFit(trial, samples)
				} else {
					err = trial.fit(samples)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
