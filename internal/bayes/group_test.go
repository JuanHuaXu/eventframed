package bayes_test

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestCompareGroupDistinguishesSharedAndSplitOutcomes(t *testing.T) {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	shared := bayes.CompareGroup([]model.BayesianGroupMember{
		{EventID: "a", UsefulWeight: 900, NotUsefulWeight: 100},
		{EventID: "b", UsefulWeight: 1800, NotUsefulWeight: 200},
	}, policy)
	if shared.Recommendation != bayes.GroupShare || !shared.RequiresAntiPigeonCertification {
		t.Fatalf("shared comparison = %+v", shared)
	}
	divergent := bayes.CompareGroup([]model.BayesianGroupMember{
		{EventID: "a", UsefulWeight: 20},
		{EventID: "b", NotUsefulWeight: 20},
	}, policy)
	if divergent.Recommendation != bayes.GroupSplit || divergent.SplitPosteriorProbability < policy.DecisionThreshold {
		t.Fatalf("split comparison = %+v", divergent)
	}
}

func TestCompareGroupRemainsUncertainWithoutPerMemberSupport(t *testing.T) {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	comparison := bayes.CompareGroup([]model.BayesianGroupMember{
		{EventID: "a", UsefulWeight: 30},
		{EventID: "b", NotUsefulWeight: 2},
	}, policy)
	if comparison.Recommendation != bayes.GroupUncertain || comparison.SufficientSupport {
		t.Fatalf("under-supported comparison = %+v", comparison)
	}
}

func TestCompareGroupUsesPracticalEquivalenceWithoutGrantingAuthority(t *testing.T) {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16, EquivalenceMargin: .15, EquivalenceThreshold: .95, MaxUncertainBorrowing: .1}
	compatible := bayes.CompareGroup([]model.BayesianGroupMember{
		{EventID: "a", UsefulWeight: 80, NotUsefulWeight: 20},
		{EventID: "b", UsefulWeight: 80, NotUsefulWeight: 20},
	}, policy)
	if compatible.Recommendation != bayes.GroupShare || compatible.EquivalenceProbability < policy.EquivalenceThreshold || !compatible.RequiresAntiPigeonCertification {
		t.Fatalf("compatible comparison = %+v", compatible)
	}
	divergent := bayes.CompareGroup([]model.BayesianGroupMember{
		{EventID: "a", UsefulWeight: 65, NotUsefulWeight: 35},
		{EventID: "b", UsefulWeight: 35, NotUsefulWeight: 65},
	}, policy)
	if divergent.Recommendation != bayes.GroupSplit || divergent.BorrowingWeight != 0 {
		t.Fatalf("divergent comparison = %+v", divergent)
	}
}
