package bayes

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"testing"
)

func TestWorkingBeliefReversesAndMapsOutcome(t *testing.T) {
	p := DefaultWorkingPolicy()
	var state *model.WorkingBelief
	for i := 0; i < 10000; i++ {
		state = UpdateWorking(state, true, 20, false, p)
	}
	if state.LogOdds != p.MaxLogOdds {
		t.Fatal("not capped")
	}
	for i := 0; i < 5; i++ {
		state = UpdateWorking(state, false, 1, false, p)
	}
	if state.LogOdds >= 0 || state.PredictiveUseful >= .5 {
		t.Fatalf("did not reverse: %+v", state)
	}
	p.LowUseful = .1
	p.HighUseful = .6
	prior := PredictiveMean(model.BayesianPosterior{Alpha: 1e9, Beta: 1}, p)
	if math.Abs(prior-.35) > 1e-12 {
		t.Fatalf("confused hypothesis mass with outcome predictive: %g", prior)
	}
	reset := UpdateWorking(state, true, 1, true, p)
	fresh := UpdateWorking(nil, true, 1, false, p)
	if *reset != *fresh {
		t.Fatal("reset retains stale belief")
	}
}

func TestWorkingPolicyAndWeights(t *testing.T) {
	p := DefaultWorkingPolicy()
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 31} {
		q := p
		q.MaxLogOdds = v
		if q.Valid() {
			t.Fatalf("accepted %g", v)
		}
	}
	a := UpdateWorking(nil, true, 1, false, p)
	b := UpdateWorking(nil, true, 20, false, p)
	if *a != *b {
		t.Fatal("importance weight minted independent evidence")
	}
	p.Retention = math.NaN()
	if UpdateWorking(a, true, 1, false, p) != nil {
		t.Fatal("invalid policy")
	}
}

func BenchmarkWorkingBelief(b *testing.B) {
	p := DefaultWorkingPolicy()
	var state *model.WorkingBelief
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		state = UpdateWorking(state, i%2 == 0, 1, false, p)
	}
}
