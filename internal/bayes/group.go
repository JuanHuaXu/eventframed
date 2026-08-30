package bayes

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	GroupUncertain = "uncertain"
	GroupShare     = "share"
	GroupSplit     = "split"
)

type GroupPolicy struct {
	PriorSplit            float64
	DecisionThreshold     float64
	MinMemberSupport      float64
	MaxMembers            int
	EquivalenceMargin     float64
	EquivalenceThreshold  float64
	MaxUncertainBorrowing float64
	// SharedEvidenceWeight discounts only the pooled posterior. Per-member
	// evidence remains full strength so Anti-Pigeon retains split authority.
	// Zero preserves the pre-discount behavior for explicit legacy policies.
	SharedEvidenceWeight float64
	// DisableAutoRevision exists for frozen control arms and deployments that
	// require manual certificate revocation. Production defaults to false.
	DisableAutoRevision bool
}

func (policy GroupPolicy) Valid() bool {
	return policy.PriorSplit > 0 && policy.PriorSplit < 1 && policy.DecisionThreshold > .5 && policy.DecisionThreshold < 1 && policy.MinMemberSupport > 0 && policy.MaxMembers >= 2 && policy.EquivalenceMargin >= 0 && policy.EquivalenceMargin < 1 && policy.EquivalenceThreshold >= 0 && policy.EquivalenceThreshold < 1 && policy.MaxUncertainBorrowing >= 0 && policy.MaxUncertainBorrowing <= 1 && policy.SharedEvidenceWeight >= 0 && policy.SharedEvidenceWeight <= 1
}

func SharedOutcomeWeight(policy GroupPolicy, certifiedShared bool, weight float64) float64 {
	if !certifiedShared || policy.SharedEvidenceWeight == 0 {
		return weight
	}
	return weight * policy.SharedEvidenceWeight
}

func CompareGroup(members []model.BayesianGroupMember, policy GroupPolicy) model.BayesianGroupComparison {
	comparison := model.BayesianGroupComparison{Members: members, Recommendation: GroupUncertain, RequiresAntiPigeonCertification: true}
	if !policy.Valid() || len(members) < 2 || len(members) > policy.MaxMembers {
		return comparison
	}
	sharedUseful, sharedNotUseful := 0.0, 0.0
	comparison.SufficientSupport = true
	for _, member := range members {
		if member.UsefulWeight < 0 || member.NotUsefulWeight < 0 || !finite(member.UsefulWeight) || !finite(member.NotUsefulWeight) {
			comparison.SufficientSupport = false
			return comparison
		}
		if member.UsefulWeight+member.NotUsefulWeight < policy.MinMemberSupport {
			comparison.SufficientSupport = false
		}
		sharedUseful += member.UsefulWeight
		sharedNotUseful += member.NotUsefulWeight
		comparison.SplitLogEvidence += betaBernoulliLogEvidence(member.UsefulWeight, member.NotUsefulWeight)
	}
	comparison.SharedLogEvidence = betaBernoulliLogEvidence(sharedUseful, sharedNotUseful)
	logPriorOdds := math.Log(policy.PriorSplit) - math.Log1p(-policy.PriorSplit)
	comparison.SplitPosteriorProbability = logistic(logPriorOdds + comparison.SplitLogEvidence - comparison.SharedLogEvidence)
	comparison.EquivalenceProbability = practicalEquivalenceProbability(members, policy.EquivalenceMargin)
	if !comparison.SufficientSupport {
		return comparison
	}
	if comparison.SplitPosteriorProbability >= policy.DecisionThreshold {
		if policy.EquivalenceMargin > 0 && comparison.EquivalenceProbability >= policy.EquivalenceThreshold {
			comparison.Recommendation = GroupUncertain
			comparison.BorrowingWeight = 0
		} else {
			comparison.Recommendation = GroupSplit
			comparison.BorrowingWeight = 0
		}
	} else if policy.EquivalenceMargin > 0 && comparison.EquivalenceProbability >= policy.EquivalenceThreshold {
		comparison.Recommendation = GroupShare
		comparison.BorrowingWeight = comparison.EquivalenceProbability
	} else if policy.EquivalenceMargin == 0 && comparison.SplitPosteriorProbability <= 1-policy.DecisionThreshold {
		comparison.Recommendation = GroupShare
		comparison.BorrowingWeight = 1 - comparison.SplitPosteriorProbability
	} else {
		comparison.BorrowingWeight = math.Min(policy.MaxUncertainBorrowing, comparison.EquivalenceProbability*policy.MaxUncertainBorrowing)
	}
	return comparison
}

// practicalEquivalenceProbability uses an independent Beta-posterior normal
// approximation for the largest pairwise difference. It proposes bounded
// borrowing; it is not an Anti-Pigeon certificate.
func practicalEquivalenceProbability(members []model.BayesianGroupMember, margin float64) float64 {
	if margin <= 0 || len(members) < 2 {
		return 0
	}
	minimum := 1.0
	for left := 0; left < len(members); left++ {
		for right := left + 1; right < len(members); right++ {
			leftMean, leftVariance := betaMeanVariance(members[left])
			rightMean, rightVariance := betaMeanVariance(members[right])
			standardDeviation := math.Sqrt(leftVariance + rightVariance)
			probability := 0.0
			if standardDeviation == 0 {
				if math.Abs(leftMean-rightMean) <= margin {
					probability = 1
				}
			} else {
				difference := leftMean - rightMean
				probability = normalCDF((margin-difference)/standardDeviation) - normalCDF((-margin-difference)/standardDeviation)
			}
			minimum = math.Min(minimum, math.Max(0, math.Min(1, probability)))
		}
	}
	return minimum
}

func betaMeanVariance(member model.BayesianGroupMember) (float64, float64) {
	alpha, beta := 1+member.UsefulWeight, 1+member.NotUsefulWeight
	total := alpha + beta
	return alpha / total, alpha * beta / (total * total * (total + 1))
}

func normalCDF(value float64) float64 {
	return .5 * (1 + math.Erf(value/math.Sqrt2))
}

func UpdateMemberEvidence(posterior *model.BayesianPosterior, eventID string, useful bool, weight float64) {
	if posterior.MemberEvidence == nil {
		posterior.MemberEvidence = make(map[string]model.BayesianMemberEvidence)
	}
	evidence := posterior.MemberEvidence[eventID]
	if useful {
		evidence.UsefulWeight += weight
	} else {
		evidence.NotUsefulWeight += weight
	}
	posterior.MemberEvidence[eventID] = evidence
}

func betaBernoulliLogEvidence(useful, notUseful float64) float64 {
	return logBeta(1+useful, 1+notUseful) - logBeta(1, 1)
}

func logBeta(alpha, beta float64) float64 {
	left, _ := math.Lgamma(alpha)
	right, _ := math.Lgamma(beta)
	total, _ := math.Lgamma(alpha + beta)
	return left + right - total
}

func logistic(value float64) float64 {
	if value >= 0 {
		return 1 / (1 + math.Exp(-value))
	}
	exponential := math.Exp(value)
	return exponential / (1 + exponential)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
