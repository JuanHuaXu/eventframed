package residual

import (
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Policy struct {
	Clip             float64
	MinSupport       float64
	MinConfidence    float64
	ConfidenceDelta  float64
	MotionLimit      float64
	MaxAge           time.Duration
	ImprovementDelta float64
}

func (policy Policy) Valid() bool {
	return finite(policy.Clip) && policy.Clip > 0 && policy.Clip <= .5 && finite(policy.MinSupport) && policy.MinSupport > 0 && finite(policy.MinConfidence) && policy.MinConfidence >= 0 && policy.MinConfidence <= 1 && finite(policy.ConfidenceDelta) && policy.ConfidenceDelta > 0 && policy.ConfidenceDelta < 1 && finite(policy.MotionLimit) && policy.MotionLimit >= 0 && policy.MotionLimit <= 1 && policy.MaxAge > 0 && finite(policy.ImprovementDelta) && policy.ImprovementDelta >= 0
}

func Update(current model.ResidualRecord, observation model.ResidualObservation, scope model.ResidualScope, key, tenantID string, weight float64, snapshot model.Snapshot, policy Policy) model.ResidualRecord {
	target := 0.0
	if observation.Useful {
		target = 1
	}
	if current.ID == "" || current.PolicyVersion != snapshot.PolicyVersion || current.EvidenceEpoch != snapshot.EvidenceEpoch {
		current = model.ResidualRecord{
			ID: residualID(scope, key), TenantID: tenantID, Scope: scope, Key: key, HorizonKey: observation.HorizonKey,
			ImprovementAlpha: 1, ImprovementBeta: 1, ReferenceProbability: observation.BaseProbability,
			MotionLimit: policy.MotionLimit, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
			SourceJournalID: observation.JournalID, SourceEventID: observation.EventID, CreatedAt: observation.AvailableAt, Active: true,
		}
	}
	if current.EffectiveSupport > 0 && observation.ValidationEligible {
		baseLoss := square(target - observation.BaseProbability)
		corrected := clamp(observation.BaseProbability+current.Correction, 0, 1)
		correctedLoss := square(target - corrected)
		if correctedLoss+policy.ImprovementDelta < baseLoss {
			current.ImprovementAlpha++
		} else {
			current.ImprovementBeta++
		}
	}
	observedResidual := target - observation.BaseProbability
	current.Correction = clamp((current.Correction*current.EffectiveSupport+observedResidual*weight)/(current.EffectiveSupport+weight), -policy.Clip, policy.Clip)
	current.EffectiveSupport += weight
	current.UpdatedAt = observation.AvailableAt
	current.ResidualVersion = snapshot.ResidualVersion
	return current
}

func Eligible(record model.ResidualRecord, baseProbability float64, snapshot model.Snapshot, now time.Time, policy Policy) bool {
	if !record.Active || record.ID == "" || record.TenantID == "" || record.Key == "" || record.SourceJournalID == "" || record.SourceEventID == "" || record.HorizonKey != model.RetrievalUsefulnessHorizon || record.PolicyVersion != snapshot.PolicyVersion || record.EvidenceEpoch != snapshot.EvidenceEpoch || record.EffectiveSupport < policy.MinSupport || AnytimeImprovementLCB(record, policy.ConfidenceDelta) < policy.MinConfidence {
		return false
	}
	if record.UpdatedAt.IsZero() || now.Before(record.UpdatedAt) || now.Sub(record.UpdatedAt) > policy.MaxAge {
		return false
	}
	motionBound := math.Abs(baseProbability-record.ReferenceProbability) + record.ApproximationErrorBound
	return finite(record.Correction) && finite(motionBound) && motionBound <= record.MotionLimit
}

// AnytimeImprovementLCB uses a Hoeffding bound with delta/(n(n+1)) error
// spending. A union bound therefore covers every repeated cache check.
func AnytimeImprovementLCB(record model.ResidualRecord, delta float64) float64 {
	successes := record.ImprovementAlpha - 1
	failures := record.ImprovementBeta - 1
	n := successes + failures
	if n <= 0 || delta <= 0 || delta >= 1 {
		return 0
	}
	mean := successes / n
	radius := math.Sqrt(math.Log(2*n*(n+1)/delta) / (2 * n))
	return clamp(mean-radius, 0, 1)
}

func Apply(law model.BernoulliLaw, record model.ResidualRecord, policy Policy) model.BernoulliLaw {
	useful := clamp(law.Useful, 0, 1)
	useful = clamp(useful+clamp(record.Correction, -policy.Clip, policy.Clip), 0, 1)
	return model.BernoulliLaw{Useful: useful, NotUseful: 1 - useful}
}

func residualID(scope model.ResidualScope, key string) string { return string(scope) + ":" + key }
func square(value float64) float64                            { return value * value }
func finite(value float64) bool                               { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func clamp(value, low, high float64) float64                  { return math.Min(high, math.Max(low, value)) }
