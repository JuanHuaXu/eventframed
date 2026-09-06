package bayes

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestGridBayesTimingAndInvalidState(t *testing.T) {
	p := GridWorkingPolicy()
	s := UpdateWorking(nil, true, 1, false, p)
	sum := 0.0
	for j, w := range s.GridWeights {
		want := gridProbability(j) / 10.5
		if math.Abs(w-want) > 1e-14 {
			t.Fatal("wrong likelihood normalization", j, w, want)
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-14 {
		t.Fatal("unnormalized")
	}
	mean := PredictiveMean(model.BayesianPosterior{WorkingBelief: s}, p)
	if mean != s.PredictiveUseful {
		t.Fatal("next-observation law mismatch")
	}
	old := *s
	_ = PredictiveMean(model.BayesianPosterior{WorkingBelief: s}, p)
	if *s != old {
		t.Fatal("read advanced hazard")
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		broken := *s
		broken.GridWeights[0] = bad
		if math.Abs(PredictiveMean(model.BayesianPosterior{WorkingBelief: &broken}, p)-.5) > 1e-14 {
			t.Fatal("bad state reused")
		}
	}
	s.PolicyID = "stale"
	if *UpdateWorking(s, true, 1, false, p) != *UpdateWorking(nil, true, 1, false, p) {
		t.Fatal("stale policy reused")
	}
	for _, weight := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		s = UpdateWorking(nil, true, weight, false, p)
		if math.Abs(s.PredictiveUseful-.5) > 1e-14 {
			t.Fatal("invalid weight added evidence")
		}
	}
	if *UpdateWorking(nil, true, 20, false, p) != *UpdateWorking(nil, true, 1, false, p) {
		t.Fatal("weight not capped")
	}
	b, err := json.Marshal(UpdateWorking(nil, true, 1, false, p))
	if err != nil {
		t.Fatal(err)
	}
	var copy model.WorkingBelief
	if err = json.Unmarshal(b, &copy); err != nil {
		t.Fatal(err)
	}
	if copy != *UpdateWorking(nil, true, 1, false, p) {
		t.Fatal("JSON roundtrip")
	}
}

func TestGridRangeReversalReset(t *testing.T) {
	p := GridWorkingPolicy()
	var s *model.WorkingBelief
	for i := 0; i < 10000; i++ {
		s = UpdateWorking(s, true, 1, false, p)
	}
	if s.PredictiveUseful <= .9 {
		t.Fatal("range not expanded", s.PredictiveUseful)
	}
	for i := 0; i < 10; i++ {
		s = UpdateWorking(s, false, 1, false, p)
	}
	if s.PredictiveUseful >= .5 {
		t.Fatal("not revisable", s.PredictiveUseful)
	}
	if *UpdateWorking(s, true, 1, true, p) != *UpdateWorking(nil, true, 1, false, p) {
		t.Fatal("reset leaked old state")
	}
	p.Retention = .9
	if p.Valid() {
		t.Fatal("accepted undefined grid policy")
	}
}

func BenchmarkGridBelief(b *testing.B) {
	p := GridWorkingPolicy()
	var s *model.WorkingBelief
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s = UpdateWorking(s, i%2 == 0, 1, false, p)
	}
}
