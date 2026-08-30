package claimsexperiment

import (
	"fmt"
	"math/rand"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	groupComparisonTrajectories = 64
	groupComparisonSamples      = 100
)

type groupScenario struct {
	name, expected string
	left, right    float64
}

func RunGroupComparison(seedBase int64) GroupComparisonReport {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 64, EquivalenceMargin: .15, EquivalenceThreshold: .80, MaxUncertainBorrowing: .10}
	scenarios := []groupScenario{
		{name: "shared_0.8_0.8", expected: bayes.GroupShare, left: .8, right: .8},
		{name: "split_0.9_0.1", expected: bayes.GroupSplit, left: .9, right: .1},
		{name: "split_0.65_0.35", expected: bayes.GroupSplit, left: .65, right: .35},
	}
	report := GroupComparisonReport{Trajectories: groupComparisonTrajectories, SamplesPerMember: groupComparisonSamples, SeedBase: seedBase, Scenarios: make(map[string]GroupComparisonCase, len(scenarios))}
	for scenarioIndex, scenario := range scenarios {
		counts := map[string]int{bayes.GroupShare: 0, bayes.GroupSplit: 0, bayes.GroupUncertain: 0}
		for trajectory := 0; trajectory < groupComparisonTrajectories; trajectory++ {
			rng := rand.New(rand.NewSource(seedBase + int64(scenarioIndex*104729+trajectory*7919)))
			leftUseful, rightUseful := 0, 0
			for sample := 0; sample < groupComparisonSamples; sample++ {
				if rng.Float64() < scenario.left {
					leftUseful++
				}
				if rng.Float64() < scenario.right {
					rightUseful++
				}
			}
			comparison := bayes.CompareGroup([]model.BayesianGroupMember{
				{EventID: "left", UsefulWeight: float64(leftUseful), NotUsefulWeight: float64(groupComparisonSamples - leftUseful)},
				{EventID: "right", UsefulWeight: float64(rightUseful), NotUsefulWeight: float64(groupComparisonSamples - rightUseful)},
			}, policy)
			counts[comparison.Recommendation]++
		}
		denominator := float64(groupComparisonTrajectories)
		expected := counts[scenario.expected]
		wrong := counts[bayes.GroupSplit]
		if scenario.expected == bayes.GroupSplit {
			wrong = counts[bayes.GroupShare]
		}
		expectedRate := float64(expected) / denominator
		wrongRate := float64(wrong) / denominator
		expectedInterval := wilsonInterval(expected, groupComparisonTrajectories)
		wrongInterval := wilsonInterval(wrong, groupComparisonTrajectories)
		acceptance := evaluateGroupAcceptance(scenario.name, expectedRate, wrongInterval)
		report.Scenarios[scenario.name] = GroupComparisonCase{
			Expected: scenario.expected, ExpectedDecisionCount: expected, ExpectedDecisionRate: expectedRate,
			ExpectedDecisionInterval: expectedInterval,
			ShareRate:                float64(counts[bayes.GroupShare]) / denominator,
			SplitRate:                float64(counts[bayes.GroupSplit]) / denominator, UncertainRate: float64(counts[bayes.GroupUncertain]) / denominator,
			WrongCount: wrong, WrongRate: wrongRate, WrongRateInterval: wrongInterval, Acceptance: acceptance,
		}
	}
	return report
}

func evaluateGroupAcceptance(name string, expectedRate float64, wrongInterval ProportionInterval) AcceptanceResult {
	criterion, ok := FrozenProtocol().GroupCriteria[name]
	if !ok {
		return AcceptanceResult{}
	}
	result := AcceptanceResult{Evaluated: true, Passed: true}
	if expectedRate < criterion.MinimumExpectedDecisionRate {
		result.Passed = false
		result.Violations = append(result.Violations, fmt.Sprintf("expected decision rate %.6f below %.6f", expectedRate, criterion.MinimumExpectedDecisionRate))
	}
	if wrongInterval.Upper > criterion.MaximumWrongRateUpper {
		result.Passed = false
		result.Violations = append(result.Violations, fmt.Sprintf("wrong-rate Wilson upper %.6f above %.6f", wrongInterval.Upper, criterion.MaximumWrongRateUpper))
	}
	return result
}
