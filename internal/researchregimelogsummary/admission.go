package researchregimelogsummary

import (
	"errors"
	"math"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/researchbounds"
)

// AdmissionPolicy is opt-in research admission, not a daemon default. A named
// external numerical bound is required; measured oracle parity is NOT one.
// The supplied bound must cover the resulting operation's law, including
// replay, likelihood normalization, truncation and predictive arithmetic.
type AdmissionPolicy struct {
	MaximumTV, MaximumBrierShift float64
	NumericalTVBound             *float64
	NumericalBoundID             string
}

type Admission struct {
	Token
	ModelTV, NumericalTV, TotalTV, BrierShift float64
	NumericalBoundID                          string
}

func admission(token Token, modelTV float64, policy AdmissionPolicy) (Admission, error) {
	if policy.NumericalTVBound == nil || strings.TrimSpace(policy.NumericalBoundID) == "" || len(policy.NumericalBoundID) > 256 || math.IsNaN(policy.MaximumTV) || math.IsNaN(policy.MaximumBrierShift) || policy.MaximumTV < 0 || policy.MaximumTV > 1 || policy.MaximumBrierShift < 0 || policy.MaximumBrierShift > 1 {
		return Admission{}, errors.New("missing or invalid approximation admission contract")
	}
	tv, err := researchbounds.TotalTV(modelTV, *policy.NumericalTVBound)
	if err != nil {
		return Admission{}, err
	}
	brier, err := researchbounds.BinaryBrierShift(tv)
	if err != nil {
		return Admission{}, err
	}
	if tv > policy.MaximumTV || brier > policy.MaximumBrierShift {
		return Admission{}, errors.New("approximation budget exceeded: use declared fallback or shadow")
	}
	return Admission{token, modelTV, *policy.NumericalTVBound, tv, brier, policy.NumericalBoundID}, nil
}

// IssueBounded checks the POST-selection law before publishing a receipt.
// Rejected trials leave history, version and support epoch untouched. A passed
// check remains conditional on the externally justified numerical bound.
func (m *Ledger) IssueBounded(token Token, member int, at int64, policy AdmissionPolicy) (Receipt, Admission, error) {
	if token != m.token {
		return Receipt{}, Admission{}, errors.New("stale approximation admission")
	}
	trial := *m
	trial.cache = append([]prefix(nil), m.cache...)
	receipt, err := trial.Issue(member, at)
	if err != nil {
		return Receipt{}, Admission{}, err
	}
	check, err := admission(trial.token, trial.state.Envelope, policy)
	if err != nil {
		return Receipt{}, Admission{}, err
	}
	*m = trial
	return receipt, check, nil
}

type BoundedCleanBranch struct {
	NextClean []float64
	Admission Admission
}

type BoundedCleanQuery struct {
	Token
	Branches [2]BoundedCleanBranch
}

// PendingCleanBounded cannot certify counterfactual branches from the current-state
// envelope alone: rare evidence can amplify truncation error. Each replay's
// conditional clean forecast is checked separately under one immutable token.
// Evidence-ratio branch weights concern historical latent paths and are NOT
// certified by this current-state bound; they are deliberately not returned.
func (m *Ledger) PendingCleanBounded(token Token, ordinal, which int, policy AdmissionPolicy) (BoundedCleanQuery, error) {
	q, err := m.Pending(token, ordinal, which)
	if err != nil {
		return BoundedCleanQuery{}, err
	}
	result := BoundedCleanQuery{Token: token}
	for value, branch := range q.Branches {
		check, err := admission(token, branch.TVEnvelope, policy)
		if err != nil {
			return BoundedCleanQuery{}, err
		}
		result.Branches[value] = BoundedCleanBranch{branch.NextClean, check}
	}
	return result, nil
}
