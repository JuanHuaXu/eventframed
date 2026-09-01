package bayes

import (
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// AssessRevision combines the temporal changepoint signal with the
// shared-versus-split comparison. Only full-stream or independent-audit
// evidence may revoke a sharing certificate; selected evidence may update the
// posterior but cannot make the structural decision self-certifying.
func AssessRevision(posterior model.BayesianPosterior, memberEventIDs []string, triggerEventID string, changePoint, validationEligible bool, policy GroupPolicy) model.BayesianRevision {
	// A "shock" is the fast-revocation case: a changepoint and independently
	// validation-eligible group divergence split stale shared confidence and
	// reset the revealing member. A changepoint alone may reset an existing
	// posterior, but it cannot manufacture a split certificate.
	revision := model.BayesianRevision{Action: model.BayesianRevisionRetain, TriggerEventID: triggerEventID}
	if !strings.HasPrefix(posterior.PosteriorKey, "ap:") || len(memberEventIDs) < 2 {
		if changePoint {
			revision.Action = model.BayesianRevisionIndividualReset
		}
		return revision
	}
	revision.CertificateID = strings.TrimPrefix(posterior.PosteriorKey, "ap:")
	members := make([]model.BayesianGroupMember, 0, len(memberEventIDs))
	for _, eventID := range memberEventIDs {
		evidence := posterior.MemberEvidence[eventID]
		members = append(members, model.BayesianGroupMember{
			EventID: eventID, UsefulWeight: evidence.UsefulWeight, NotUsefulWeight: evidence.NotUsefulWeight,
			CurrentPosteriorKey: posterior.PosteriorKey,
		})
	}
	revision.Comparison = CompareGroup(members, policy)
	if !policy.DisableAutoRevision && validationEligible && revision.Comparison.Recommendation == GroupSplit {
		revision.Action = model.BayesianRevisionSplit
		if changePoint {
			revision.Action = model.BayesianRevisionSplitReset
		}
		revision.ValidationBasis = "full-stream-or-independent-audit divergence invalidated prior sharing"
		return revision
	}
	if changePoint {
		revision.Action = model.BayesianRevisionSharedReset
		revision.ValidationBasis = "bounded changepoint detector"
	}
	return revision
}

func RevisionSplits(action model.BayesianRevisionAction) bool {
	return action == model.BayesianRevisionSplit || action == model.BayesianRevisionSplitReset
}
