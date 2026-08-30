package audit

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestEstimateInfluenceRequiresThePreselectedSample(t *testing.T) {
	population := []string{"a", "b", "c"}
	local := model.BernoulliLaw{Useful: .5, NotUseful: .5}
	expanded := map[string]model.BernoulliLaw{
		"a": {Useful: .6, NotUseful: .4},
		"b": {Useful: .5, NotUseful: .5},
		"c": {Useful: .4, NotUseful: .6},
	}
	estimate, err := EstimateInfluence(population, local, expanded, "seed", 1, 1, 1, .05)
	if err != nil {
		t.Fatal(err)
	}
	if len(estimate.SelectedIDs) != len(population) || estimate.UpperBound < estimate.Mean || estimate.UpperBound > 1 {
		t.Fatalf("estimate = %+v", estimate)
	}
	delete(expanded, "b")
	if _, err := EstimateInfluence(population, local, expanded, "seed", 1, 1, 1, .05); err == nil {
		t.Fatal("expected missing selected observation rejection")
	}
}

func TestJensenShannonIsNormalizedAndSymmetric(t *testing.T) {
	left := model.BernoulliLaw{Useful: 1, NotUseful: 0}
	right := model.BernoulliLaw{Useful: 0, NotUseful: 1}
	if JensenShannon(left, right) != 1 || JensenShannon(right, left) != 1 || JensenShannon(left, left) != 0 {
		t.Fatal("unexpected normalized Jensen-Shannon divergence")
	}
}
