package claimsexperiment

import (
	"fmt"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/audit"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func RunOmittedAuditCoverage(seed int64) OmittedAuditReport {
	const trials, populationSize = 256, 1000
	const probability, delta = .2, .05
	local := model.BernoulliLaw{Useful: .5, NotUseful: .5}
	report := OmittedAuditReport{Trials: trials, PopulationSize: populationSize, AuditProbability: probability, ConfidenceLevel: 1 - delta}
	for trial := 0; trial < trials; trial++ {
		population := make([]string, populationSize)
		expanded := make(map[string]model.BernoulliLaw)
		trueMean := 0.0
		seedKey := fmt.Sprintf("audit-%d-%d", seed, trial)
		for index := 0; index < populationSize; index++ {
			eventID := fmt.Sprintf("event-%04d", index)
			population[index] = eventID
			useful := .5 + .12*math.Sin(float64(index+trial)/17)
			law := model.BernoulliLaw{Useful: useful, NotUseful: 1 - useful}
			trueMean += audit.JensenShannon(local, law)
			if audit.Selected(eventID, seedKey, 1, probability) {
				expanded[eventID] = law
			}
		}
		trueMean /= populationSize
		estimate, err := audit.EstimateInfluence(population, local, expanded, seedKey, 1, 1, probability, delta)
		if err != nil {
			report.Errors++
			continue
		}
		if estimate.UpperBound+1e-12 >= trueMean {
			report.CoveredTrials++
		}
		report.MeanUpperBound += estimate.UpperBound
		report.MeanTrueInfluence += trueMean
	}
	completed := trials - report.Errors
	if completed > 0 {
		report.CoverageRate = float64(report.CoveredTrials) / float64(completed)
		report.CoverageInterval = wilsonInterval(report.CoveredTrials, completed)
		report.MeanUpperBound /= float64(completed)
		report.MeanTrueInfluence /= float64(completed)
	}
	criterion := FrozenProtocol().OmittedAuditCriteria
	report.Acceptance = AcceptanceResult{Evaluated: true, Passed: report.Errors == 0 && report.CoverageInterval.Lower >= criterion.MinimumCoverageLower}
	if report.Errors > 0 {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "one or more audit trials failed")
	}
	if report.CoverageInterval.Lower < criterion.MinimumCoverageLower {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "coverage Wilson lower bound is below 95 percent")
	}
	return report
}
