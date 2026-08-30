package claimsexperiment

import (
	"fmt"
	"math/rand"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	changepointTrajectories = 64
	changepointWindow       = 20
)

var frozenChangePolicy = bayes.ChangePolicy{
	Hazard: .05, Threshold: .30, MaxRun: 32,
	FastRate: .25, SlowRate: .025, DriftThreshold: .30, DriftPersistence: 12, MinSamples: 20,
	CUSUMSlack: .10, CUSUMThreshold: 8,
	CooldownSamples: 20,
}

type changeScenario struct {
	name        string
	probability []float64
	changes     []int
	window      int
}

func RunChangepointAdaptation() ChangepointReport {
	return RunChangepointAdaptationWithSeed(982451653)
}

func RunChangepointAdaptationWithSeed(seedBase int64) ChangepointReport {
	scenarios := []changeScenario{
		{name: "stable", probability: constantProbabilities(90, .8)},
		{name: "abrupt_noiseless", probability: joinedProbabilities(40, 1, 40, 0), changes: []int{40}},
		{name: "abrupt", probability: joinedProbabilities(40, .9, 40, .1), changes: []int{40}},
		{name: "gradual", probability: gradualScenarioProbabilities(), changes: []int{30}, window: 60},
		{name: "recurring_noiseless", probability: recurringProbabilities(1, 0), changes: []int{30, 60}},
		{name: "recurring", probability: recurringProbabilities(.9, .1), changes: []int{30, 60}},
	}
	report := ChangepointReport{
		Trajectories:    changepointTrajectories,
		SeedBase:        seedBase,
		DetectionWindow: changepointWindow,
		MatchingRule:    DetectionMatchingRule,
		MeanDelayBasis:  DetectedChangeDelayBasis,
		Policy: ChangepointPolicy{
			Hazard: frozenChangePolicy.Hazard, Threshold: frozenChangePolicy.Threshold, MaxRun: frozenChangePolicy.MaxRun,
			RecentWindow: frozenChangePolicy.RecentWindow, RecentThreshold: frozenChangePolicy.RecentThreshold,
			FastRate: frozenChangePolicy.FastRate, SlowRate: frozenChangePolicy.SlowRate, DriftThreshold: frozenChangePolicy.DriftThreshold,
			DriftPersistence: frozenChangePolicy.DriftPersistence, MinSamples: frozenChangePolicy.MinSamples,
			CUSUMSlack: frozenChangePolicy.CUSUMSlack, CUSUMThreshold: frozenChangePolicy.CUSUMThreshold,
			CooldownSamples: frozenChangePolicy.CooldownSamples,
		},
		Scenarios: make(map[string]ChangepointCase, len(scenarios)),
	}
	for scenarioIndex, scenario := range scenarios {
		report.Scenarios[scenario.name] = runChangeScenario(scenario, seedBase+int64(scenarioIndex*104729))
	}
	return report
}

func runChangeScenario(scenario changeScenario, seed int64) ChangepointCase {
	window := scenario.window
	if window == 0 {
		window = changepointWindow
	}
	result := ChangepointCase{DetectionWindow: window, ExpectedChanges: len(scenario.changes) * changepointTrajectories}
	delaySum := 0
	for trajectory := 0; trajectory < changepointTrajectories; trajectory++ {
		rng := rand.New(rand.NewSource(seed + int64(trajectory)*7919))
		var posterior model.BayesianPosterior
		detections := make([]int, 0, len(scenario.changes))
		for index, probability := range scenario.probability {
			var triggered bool
			posterior, triggered = bayes.ApplyOutcome(posterior, rng.Float64() < probability, 1, frozenChangePolicy)
			if triggered {
				detections = append(detections, index)
			}
		}
		matched := make([]bool, len(detections))
		for _, change := range scenario.changes {
			match := -1
			for index, detection := range detections {
				if !matched[index] && detection >= change && detection <= change+window {
					match = index
					break
				}
			}
			if match < 0 {
				continue
			}
			matched[match] = true
			delay := detections[match] - change
			result.DetectedChanges++
			delaySum += delay
			if delay > result.MaxDelay {
				result.MaxDelay = delay
			}
		}
		for _, used := range matched {
			if !used {
				result.FalseAlarms++
			}
		}
	}
	if result.DetectedChanges > 0 {
		result.MeanDelay = float64(delaySum) / float64(result.DetectedChanges)
	}
	result.DelaySampleCount = result.DetectedChanges
	result.MeanDelayBasis = DetectedChangeDelayBasis
	result.TotalTriggers = result.DetectedChanges + result.FalseAlarms
	result.UnmatchedAlarmsPerTrajectory = float64(result.FalseAlarms) / changepointTrajectories
	if result.TotalTriggers > 0 {
		result.UnmatchedTriggerRate = float64(result.FalseAlarms) / float64(result.TotalTriggers)
	}
	if result.ExpectedChanges > 0 {
		result.DetectionRate = float64(result.DetectedChanges) / float64(result.ExpectedChanges)
		result.MissRate = 1 - result.DetectionRate
		interval := wilsonInterval(result.DetectedChanges, result.ExpectedChanges)
		result.DetectionInterval = &interval
	}
	result.Acceptance = evaluateChangepointAcceptance(scenario.name, result)
	return result
}

func evaluateChangepointAcceptance(name string, result ChangepointCase) AcceptanceResult {
	criterion, ok := FrozenProtocol().ChangepointCriteria[name]
	if !ok {
		return AcceptanceResult{}
	}
	acceptance := AcceptanceResult{Evaluated: true, Passed: true}
	if criterion.MinimumDetectionRate != nil && result.DetectionRate < *criterion.MinimumDetectionRate {
		acceptance.Passed = false
		acceptance.Violations = append(acceptance.Violations, fmt.Sprintf("detection rate %.6f below %.6f", result.DetectionRate, *criterion.MinimumDetectionRate))
	}
	if result.UnmatchedAlarmsPerTrajectory > criterion.MaximumUnmatchedAlarmsPerTrajectory {
		acceptance.Passed = false
		acceptance.Violations = append(acceptance.Violations, fmt.Sprintf("unmatched alarms per trajectory %.6f above %.6f", result.UnmatchedAlarmsPerTrajectory, criterion.MaximumUnmatchedAlarmsPerTrajectory))
	}
	if criterion.MaximumMeanDelay != nil && result.MeanDelay > *criterion.MaximumMeanDelay {
		acceptance.Passed = false
		acceptance.Violations = append(acceptance.Violations, fmt.Sprintf("mean detected-change delay %.6f above %.6f", result.MeanDelay, *criterion.MaximumMeanDelay))
	}
	return acceptance
}

func constantProbabilities(length int, probability float64) []float64 {
	result := make([]float64, length)
	for index := range result {
		result[index] = probability
	}
	return result
}

func joinedProbabilities(firstLength int, firstProbability float64, secondLength int, secondProbability float64) []float64 {
	return append(constantProbabilities(firstLength, firstProbability), constantProbabilities(secondLength, secondProbability)...)
}

func gradualProbabilities(length int, start, end float64) []float64 {
	result := make([]float64, length)
	for index := range result {
		result[index] = start + (end-start)*float64(index)/float64(length-1)
	}
	return result
}

func gradualScenarioProbabilities() []float64 {
	probabilities := append(constantProbabilities(30, .9), gradualProbabilities(60, .9, .1)...)
	return append(probabilities, constantProbabilities(30, .1)...)
}

func recurringProbabilities(high, low float64) []float64 {
	return append(joinedProbabilities(30, high, 30, low), constantProbabilities(30, high)...)
}
