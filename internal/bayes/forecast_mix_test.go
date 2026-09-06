package bayes

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"testing"
)

func TestForecastMixture(t *testing.T) {
	var s ForecastMix
	experts := [4]float64{.2, .3, .8, .9}
	before := s
	if math.Abs(s.Forecast(experts)-.34) > 1e-14 {
		t.Fatal("wrong prior mixture")
	}
	if s != before {
		t.Fatal("read mutated state")
	}
	next := s.Observe(experts, true, 1)
	prior := mixPrior()
	for j, w := range next.Weights {
		if math.Abs(w-prior[j]*experts[j]/.34) > 1e-14 {
			t.Fatal("wrong posterior weights")
		}
	}
	if next != s.Observe(experts, true, 100) {
		t.Fatal("weight minted observations")
	}
	for _, w := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if s.Observe(experts, true, w) != s {
			t.Fatal("invalid weight learned")
		}
	}
	bad := experts
	bad[0] = math.NaN()
	if s.Observe(bad, true, 1) != s || s.Forecast(bad) != .5 {
		t.Fatal("invalid probabilities")
	}
	for i := 0; i < 10000; i++ {
		s = s.Observe(experts, true, 1)
	}
	for i := 0; i < 20; i++ {
		s = s.Observe(experts, false, 1)
	}
	if s.Forecast(experts) >= .5 {
		t.Fatal("failed reversal")
	}
	record := model.ExpertForecasts{Enabled: true, Probabilities: experts, PolicyVersion: 2, EvidenceEpoch: 3}
	snapshot := model.Snapshot{PolicyVersion: 2, EvidenceEpoch: 3}
	if UpdateForecastWeights(next.Weights, record, true, 1, true, true, snapshot) != ([4]float64{}) {
		t.Fatal("reset retained weights")
	}
	record.PolicyVersion++
	if UpdateForecastWeights(next.Weights, record, true, 1, false, true, snapshot) != next.Weights {
		t.Fatal("stale policy learned")
	}
	record.PolicyVersion--
	record.EvidenceEpoch++
	if UpdateForecastWeights(next.Weights, record, true, 1, false, true, snapshot) != next.Weights {
		t.Fatal("stale epoch learned")
	}
}

var forecastMixtureBenchmarkSink float64

func BenchmarkForecastMixture(b *testing.B) {
	var s ForecastMix
	p := [4]float64{.2, .3, .8, .9}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		forecastMixtureBenchmarkSink = s.Forecast(p)
		s = s.Observe(p, i%2 == 0, 1)
	}
}
