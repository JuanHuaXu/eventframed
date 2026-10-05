package observationgate

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

var monitorCreditBenchmarkSink float64

// Compares acquisition/forecast work only; no gate, fitting, network or database.
func BenchmarkMonitorCreditAcquisition(b *testing.B) {
	base, _, err := arrivalSwitchSetup(1, 3, 0)
	if err != nil {
		b.Fatal(err)
	}
	for _, useCredit := range []bool{false, true} {
		name := "original"
		if useCredit {
			name = "credit"
		}
		b.Run(name, func(b *testing.B) {
			credit := 0
			sum := 0.
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				readers := [2]observation.Reader{observationexperiment.Frames(uint16(i&511), "bench-live"), observationexperiment.Frames(uint16((37*i+23)&511), "bench-ref")}
				var cached [2]*monitorCreditReader
				if useCredit {
					for side, r := range readers {
						cached[side] = newMonitorCreditReader(r)
						readers[side] = cached[side]
					}
				}
				bounded := 0
				for _, r := range readers {
					p, e := observation.Run(base, r, 1, "mmm", 0)
					if e != nil {
						b.Fatal(e)
					}
					sum += p.Probability
					bounded += p.Cost
				}
				audit := i%4 == 0
				full := audit
				if useCredit {
					var e error
					full, _, credit, e = monitorCreditStep(credit, bounded, audit)
					if e != nil {
						b.Fatal(e)
					}
				}
				if full {
					for side, r := range readers {
						for scope := 0; scope < 3; scope++ {
							if _, _, e := r.Read(observation.View{Scope: scope, Depth: 2}); e != nil {
								b.Fatal(e)
							}
						}
						if useCredit {
							p, e := base.ForecastObserved(cached[side].mask, cached[side].values)
							if e != nil {
								b.Fatal(e)
							}
							sum += p
						}
					}
				}
			}
			monitorCreditBenchmarkSink = sum
		})
	}
}
