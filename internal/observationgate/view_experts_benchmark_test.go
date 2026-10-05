package observationgate

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
)

var viewExpertBenchmarkSink float64

// Retain both forecast and update results, so the compiler cannot discard the
// measured prediction arithmetic. This excludes the replay driver's I/O.
func BenchmarkViewExpertsForecastAndUpdate(b *testing.B) {
	var mix bayes.ForecastMix
	a := viewAdvice{.7, .8}.experts()
	sum := 0.
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sum += mix.Forecast(a)
		mix = mix.Observe(a, i%3 != 0, 1)
	}
	viewExpertBenchmarkSink = sum + mix.Weights[0]
}
