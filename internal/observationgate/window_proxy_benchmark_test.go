package observationgate

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"testing"
)

var windowProxyBenchSink windowGuideAdvice

func BenchmarkWindowProxyPredict(b *testing.B) {
	frames := make([]innerArrivalFrame, 64)
	ids := make([]int, 64)
	for i := range frames {
		frames[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
		ids[i] = i
	}
	fitted, e := fitEventWindow(frames, ids)
	if e != nil {
		b.Fatal(e)
	}
	models := windowGuideModels{fitted.count, fitted.count, fitted.count, fitted.count, fitted.subset, fitted.subset}
	focus := bayes.ForecastMix{Weights: [4]float64{0, 1, 0, 0}}
	for _, enabled := range []bool{false, true} {
		b.Run(fmt.Sprint(enabled), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p, e := windowProxyPredict(models, bayes.ForecastMix{}, focus, focus, observationexperiment.Frames(uint16(i%512), "proxy-bench"), 0, enabled)
				if e != nil {
					b.Fatal(e)
				}
				windowProxyBenchSink = p
			}
		})
	}
}
