package claimsexperiment

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
)

func RunHeterogeneousResidual(seed int64) HeterogeneousResidualReport {
	const trajectories, training, evaluation = 64, 40, 24
	policy := residual.Policy{Clip: .15, MinSupport: 8, MinConfidence: .1, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: 24 * time.Hour, ImprovementDelta: .001}
	snapshot := model.Snapshot{PolicyVersion: 1, EvidenceEpoch: 1, ResidualVersion: 1}
	rng := rand.New(rand.NewSource(seed))
	gains := make([]float64, 0, trajectories)
	report := HeterogeneousResidualReport{Trajectories: trajectories, StableTrajectories: trajectories / 2, ReversedTrajectories: trajectories / 2, EvaluationCases: trajectories * evaluation}
	for trajectory := 0; trajectory < trajectories; trajectory++ {
		trajectoryGain := 0.0
		trainP := .2
		if trajectory%2 == 0 {
			trainP = .8
		}
		evalP := trainP
		if trajectory >= trajectories/2 {
			evalP = 1 - trainP
		}
		var record model.ResidualRecord
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
				baseLoss, residualLoss := square(target-base.Useful), square(target-corrected.Useful)
				report.BaselineBrier += baseLoss
				report.ResidualBrier += residualLoss
				trajectoryGain += baseLoss - residualLoss
				if applied {
					report.AppliedCases++
					if residualLoss > baseLoss {
						report.HarmfulReuseCases++
					}
				}
			}
			observation := model.ResidualObservation{HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: .5, Useful: useful, ValidationEligible: true, EventID: fmt.Sprintf("event-%d", trajectory), JournalID: fmt.Sprintf("journal-%d-%d", trajectory, turn), AvailableAt: now}
			record = residual.Update(record, observation, model.ResidualExact, fmt.Sprintf("key-%d", trajectory), "heterogeneous", 1, snapshot, policy)
			report.MaintenanceUpdates++
		}
		gains = append(gains, trajectoryGain/float64(evaluation))
	}
	report.BaselineBrier /= float64(report.EvaluationCases)
	report.ResidualBrier /= float64(report.EvaluationCases)
	report.MeanGain = report.BaselineBrier - report.ResidualBrier
	report.GainInterval = bootstrapMeanInterval(gains, seed+17, 4000)
	if report.AppliedCases > 0 {
		report.HarmfulReuseRate = float64(report.HarmfulReuseCases) / float64(report.AppliedCases)
	}
	report.HarmfulReuseInterval = wilsonInterval(report.HarmfulReuseCases, report.AppliedCases)
	report.Acceptance = AcceptanceResult{Evaluated: true, Passed: report.GainInterval.Lower > 0 && report.HarmfulReuseInterval.Upper <= .05}
	if report.GainInterval.Lower <= 0 {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "trajectory bootstrap gain lower bound is not above zero")
	}
	if report.HarmfulReuseInterval.Upper > .05 {
		report.Acceptance.Violations = append(report.Acceptance.Violations, "harmful false-reuse Wilson upper bound exceeds 5%")
	}
	return report
}
