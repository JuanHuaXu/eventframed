package bayes_test

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestAssessRevisionSeparatesDivergenceFromSharedChange(t *testing.T) {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	divergent := model.BayesianPosterior{
		PosteriorKey: "ap:bucket-a",
		MemberEvidence: map[string]model.BayesianMemberEvidence{
			"a": {UsefulWeight: 20},
			"b": {NotUsefulWeight: 20},
		},
	}
	revision := bayes.AssessRevision(divergent, []string{"a", "b"}, "b", false, true, policy)
	if revision.Action != model.BayesianRevisionSplit || revision.CertificateID != "bucket-a" {
		t.Fatalf("divergent revision = %+v", revision)
	}
	revision = bayes.AssessRevision(divergent, []string{"a", "b"}, "b", true, true, policy)
	if revision.Action != model.BayesianRevisionSplitReset {
		t.Fatalf("divergent changepoint revision = %+v", revision)
	}

	coherent := model.BayesianPosterior{
		PosteriorKey: "ap:bucket-a",
		MemberEvidence: map[string]model.BayesianMemberEvidence{
			"a": {UsefulWeight: 20, NotUsefulWeight: 5},
			"b": {UsefulWeight: 20, NotUsefulWeight: 5},
		},
	}
	revision = bayes.AssessRevision(coherent, []string{"a", "b"}, "b", true, true, policy)
	if revision.Action != model.BayesianRevisionSharedReset {
		t.Fatalf("coherent changepoint revision = %+v", revision)
	}
}

func TestAssessRevisionDoesNotLetSelectedEvidenceRevokeSharing(t *testing.T) {
	policy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	posterior := model.BayesianPosterior{
		PosteriorKey: "ap:bucket-a",
		MemberEvidence: map[string]model.BayesianMemberEvidence{
			"a": {UsefulWeight: 20},
			"b": {NotUsefulWeight: 20},
		},
	}
	revision := bayes.AssessRevision(posterior, []string{"a", "b"}, "b", false, false, policy)
	if revision.Action != model.BayesianRevisionRetain {
		t.Fatalf("selected evidence gained structural authority: %+v", revision)
	}
}
