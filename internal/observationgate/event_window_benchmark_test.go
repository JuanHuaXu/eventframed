package observationgate

import (
	"fmt"
	"testing"
)

var eventWindowBenchSink eventWindowModels
var eventWindowForecastSink [2]float64

func BenchmarkEventWindowFit(b *testing.B) {
	for _, n := range []int{10, 64} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			frames := make([]innerArrivalFrame, n)
			ids := make([]int, n)
			for i := range frames {
				frames[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
				ids[i] = i
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m, e := fitEventWindow(frames, ids)
				if e != nil {
					b.Fatal(e)
				}
				eventWindowBenchSink = m
			}
		})
	}
}
func BenchmarkEventWindowForecast(b *testing.B) {
	frames := make([]innerArrivalFrame, 10)
	ids := make([]int, 10)
	for i := range frames {
		frames[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
		ids[i] = i
	}
	m, e := fitEventWindow(frames, ids)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, e := m.forecast(63, uint16(i%64))
		if e != nil {
			b.Fatal(e)
		}
		eventWindowForecastSink = p
	}
}
