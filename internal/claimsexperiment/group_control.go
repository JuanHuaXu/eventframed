package claimsexperiment

import (
	"context"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func RunGroupAuthorityControl(ctx context.Context) (GroupAuthorityControl, error) {
	runner, err := newAntiPigeonRunner(ctx, "separate")
	if err != nil {
		return GroupAuthorityControl{}, err
	}
	if _, err := runAntiPigeonVariant(ctx, runner, 30, 20); err != nil {
		return GroupAuthorityControl{}, err
	}
	divergent, err := runner.runtime.CompareBayesianGroup(ctx, model.BayesianGroupComparisonRequest{
		ProtocolVersion: model.ProtocolVersion,
		TenantID:        "anti-pigeon-experiment",
		MemberEventIDs:  antiPigeonIDs,
	})
	if err != nil {
		return GroupAuthorityControl{}, err
	}
	compatible, err := runner.runtime.CompareBayesianGroup(ctx, model.BayesianGroupComparisonRequest{
		ProtocolVersion: model.ProtocolVersion,
		TenantID:        "anti-pigeon-experiment",
		MemberEventIDs:  antiPigeonIDs[:2],
	})
	if err != nil {
		return GroupAuthorityControl{}, err
	}
	authorityUnchanged := true
	for _, member := range compatible.Members {
		if member.CurrentPosteriorKey != member.EventID {
			authorityUnchanged = false
		}
	}
	control := GroupAuthorityControl{
		DivergentRecommendation:         divergent.Recommendation,
		CompatibleRecommendation:        compatible.Recommendation,
		RequiresAntiPigeonCertification: divergent.RequiresAntiPigeonCertification,
		SnapshotUnchanged:               compatible.Snapshot == divergent.Snapshot,
		PosteriorKeyAuthorityUnchanged:  authorityUnchanged,
	}
	control.Passed = control.DivergentRecommendation == bayes.GroupSplit &&
		control.CompatibleRecommendation == bayes.GroupShare &&
		control.RequiresAntiPigeonCertification && control.SnapshotUnchanged && control.PosteriorKeyAuthorityUnchanged
	return control, nil
}
