package observationgate

import (
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"testing"
)

var windowCompetitionBenchSink float64

// One view's two-level arithmetic. Model fitting, acquisition, journaling and
// database operations are deliberately not part of this kernel measurement.
func BenchmarkWindowCompetitionKernel(b *testing.B) {
	var inner, outer bayes.ForecastMix
	ia := [4]float64{.7, .6, .8, .8}
	sum := 0.
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		short := inner.Forecast(ia)
		oa := [4]float64{.7, short, .6, .5}
		sum += outer.Forecast(oa)
		inner = inner.Observe(ia, i%3 != 0, 1)
		outer = outer.Observe(oa, i%3 != 0, 1)
	}
	windowCompetitionBenchSink = sum + inner.Weights[0] + outer.Weights[0]
}
