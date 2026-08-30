package claimsexperiment

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
)

func RunSafeResidualReplacement(seed int64) SafeResidualReport {
	const trajectories, training, evaluation = 64, 40, 24
	policy := residual.Policy{
		Clip: .15, MinSupport: 8, MinConfidence: .1, ConfidenceDelta: .05,
		MotionLimit: .1, MaxAge: 24 * time.Hour, ImprovementDelta: .001,
		MinMeanGain: .001, MaxConsecutiveHarm: 1, ShrinkByConfidence: true,
	}
	snapshot := model.Snapshot{PolicyVersion: 1, EvidenceEpoch: 1, ResidualVersion: 1}
	rng := rand.New(rand.NewSource(seed))
	gains := make([]float64, 0, trajectories)
	criterion := FrozenProtocol().ResidualReplacementCriteria
	report := SafeResidualReport{Trajectories: trajectories, EvaluationCases: trajectories * evaluation, HarmBudgetPerTrajectory: criterion.MaximumWorstTrajectoryExcess}
	for trajectory := 0; trajectory < trajectories; trajectory++ {
		trainP := .2
		if trajectory%2 == 0 {
			trainP = .8
		}
		evalP := trainP
		if trajectory >= trajectories/2 {
			evalP = 1 - trainP
		}
		var record model.ResidualRecord
		trajectoryGain := 0.0
		for turn := 0; turn < training+evaluation; turn++ {
			probability := trainP
			if turn >= training {
				probability = evalP
			}
			useful := rng.Float64() < probability
			now := experimentAnchor.Add(time.Duration(turn) * time.Minute)
			base := model.BernoulliLaw{Useful: .5, NotUseful: .5}
			corrected := base
			applied := residual.Eligible(record, base.Useful, snapshot, now, policy)
			if applied {
				corrected = residual.Apply(base, record, policy)
			}
			if turn >= training {
				target := 0.0
				if useful {
					target = 1
				}
				baseLoss, correctedLoss := square(target-base.Useful), square(target-corrected.Useful)
				report.BaselineBrier += baseLoss
				report.SafeResidualBrier += correctedLoss
				trajectoryGain += baseLoss - correctedLoss
				if applied {
					report.AppliedCases++
					if correctedLoss > baseLoss {
						report.HarmfulAppliedCases++
					}
				} else {
					report.AbstainedCases++
				}
			}
			observation := model.ResidualObservation{HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: .5, Useful: useful, ValidationEligible: true, EventID: fmt.Sprintf("event-%d", trajectory), JournalID: fmt.Sprintf("journal-%d-%d", trajectory, turn), AvailableAt: now}
			record = residual.Update(record, observation, model.ResidualExact, fmt.Sprintf("key-%d", trajectory), "safe-residual", 1, snapshot, policy)
		}
		meanTrajectoryGain := trajectoryGain / float64(evaluation)
		gains = append(gains, meanTrajectoryGain)
		report.WorstTrajectoryExcess = math.Max(report.WorstTrajectoryExcess, math.Max(0, -meanTrajectoryGain))
	}
	report.BaselineBrier /= float64(report.EvaluationCases)
	report.SafeResidualBrier /= float64(report.EvaluationCases)
	report.MeanGain = report.BaselineBrier - report.SafeResidualBrier
	report.GainInterval = bootstrapMeanInterval(gains, seed+29, 4000)
	if report.AppliedCases > 0 {
		report.HarmfulAppliedRate = float64(report.HarmfulAppliedCases) / float64(report.AppliedCases)
	}
	report.Acceptance = AcceptanceResult{Evaluated: true, Passed: report.GainInterval.Lower > criterion.MinimumGainLower && report.WorstTrajectoryExcess <= report.HarmBudgetPerTrajectory}
	if report.GainInterval.Lower <= criterion.MinimumGainLower {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "trajectory bootstrap gain lower bound is not above zero")
	}
	if report.WorstTrajectoryExcess > report.HarmBudgetPerTrajectory {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "worst-trajectory excess loss exceeds the replacement harm budget")
	}
	return report
}
