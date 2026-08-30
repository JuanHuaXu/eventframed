package ranking

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestContextualScoreReturnsMissingPosteriorWeightToBaseline(t *testing.T) {
	policy := DefaultPolicy()
	policy.ContextualEnabled = true
	without := Score(Features{Baseline: .8, Posterior: .1, Recency: .6}, policy)
	with := Score(Features{Baseline: .8, Posterior: .1, Recency: .6, HasPosterior: true}, policy)
	if without <= with {
		t.Fatalf("missing posterior reduced baseline support: without=%f with=%f", without, with)
	}
}

func TestHierarchicalMeanIsBoundedByWeakParentInfluence(t *testing.T) {
	policy := DefaultPolicy()
	policy.HierarchicalEnabled = true
	child := model.BayesianPosterior{Alpha: 9, Beta: 3}
	parent := model.BayesianPosterior{Alpha: 1, Beta: 101}
	got := HierarchicalMean(child, parent, policy)
	if got >= child.Mean() || child.Mean()-got > .10 {
		t.Fatalf("hierarchical mean=%f child=%f", got, child.Mean())
	}
}
