package memorystore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestDivergentGroupOutcomeRevokesSharingAndMaterializesMembers(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	now := time.Now().UTC()
	snapshot := memory.Snapshot(ctx)
	certificate := model.AntiPigeonCertificate{
		ID: "bucket-a", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"},
		GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
	}
	if _, err := memory.PublishAntiPigeonCertificate(ctx, certificate); err != nil {
		t.Fatal(err)
	}
	changePolicy := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	var result store.BayesianOutcomeResult
	var lastRequest model.BayesianOutcomeRequest
	var lastObservation model.ResidualObservation
	for index := 0; index < 20; index++ {
		eventID, useful := "a", true
		if index%2 == 1 {
			eventID, useful = "b", false
		}
		request := model.BayesianOutcomeRequest{
			IdempotencyKey: fmt.Sprintf("outcome-%d", index), TenantID: "tenant-a", EventID: eventID,
			Useful: useful, AvailableAt: now.Add(time.Duration(index) * time.Second), Source: model.OutcomeFullStream,
		}
		observation := model.ResidualObservation{ActionKey: fmt.Sprintf("action-%d", index), GeneralKey: "general", PosteriorKey: "ap:bucket-a", CommittedProbability: .5, Useful: useful, AvailableAt: request.AvailableAt}
		lastRequest, lastObservation = request, observation
		var err error
		result, err = memory.ApplyBayesianOutcome(ctx, request, "ap:bucket-a", "", request.IdempotencyKey, 1, changePolicy, groupPolicy, observation, residualPolicy)
		if err != nil {
			t.Fatal(err)
		}
		if bayes.RevisionSplits(result.Revision.Action) {
			break
		}
	}
	if result.Revision.Action != model.BayesianRevisionSplit {
		t.Fatalf("revision did not split: %+v", result.Revision)
	}
	if _, err := memory.GetAntiPigeonCertificate(ctx, "tenant-a", []string{"a"}); !errors.Is(err, store.ErrCertificateNotFound) {
		t.Fatalf("sharing certificate survived split: %v", err)
	}
	left, err := memory.GetBayesianPosterior(ctx, "tenant-a", "a")
	if err != nil {
		t.Fatal(err)
	}
	right, err := memory.GetBayesianPosterior(ctx, "tenant-a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if left.Mean() <= .5 || right.Mean() >= .5 || result.Posterior.PosteriorKey != "b" {
		t.Fatalf("materialized posteriors left=%+v right=%+v result=%+v", left, right, result.Posterior)
	}
	shared, err := memory.GetBayesianPosterior(ctx, "tenant-a", "ap:bucket-a")
	if err != nil || shared.Certified {
		t.Fatalf("shared posterior remained usable: %+v, %v", shared, err)
	}

	duplicate, err := memory.ApplyBayesianOutcome(ctx, lastRequest, "ap:bucket-a", "", lastRequest.IdempotencyKey, 1, changePolicy, groupPolicy, lastObservation, residualPolicy)
	if err != nil || !duplicate.Duplicate || duplicate.Revision.Action != model.BayesianRevisionSplit {
		t.Fatalf("duplicate revision = %+v, %v", duplicate, err)
	}
}

func TestCertifiedSharingDiscountsPooledConfidenceButKeepsDirectEvidence(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	now := time.Now().UTC()
	snapshot := memory.Snapshot(ctx)
	certificate := model.AntiPigeonCertificate{
		ID: "compatible", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"},
		GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
	}
	if _, err := memory.PublishAntiPigeonCertificate(ctx, certificate); err != nil {
		t.Fatal(err)
	}
	changePolicy := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 100, MaxMembers: 16, SharedEvidenceWeight: .25}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	for index, eventID := range []string{"a", "b", "a", "b"} {
		request := model.BayesianOutcomeRequest{IdempotencyKey: fmt.Sprintf("shared-%d", index), TenantID: "tenant-a", EventID: eventID, Useful: true, AvailableAt: now.Add(time.Duration(index) * time.Second), Source: model.OutcomeFullStream}
		observation := model.ResidualObservation{ActionKey: request.IdempotencyKey, GeneralKey: "shared", PosteriorKey: "ap:compatible", CommittedProbability: .5, Useful: true, AvailableAt: request.AvailableAt}
		if _, err := memory.ApplyBayesianOutcome(ctx, request, "ap:compatible", "", request.IdempotencyKey, 1, changePolicy, groupPolicy, observation, residualPolicy); err != nil {
			t.Fatal(err)
		}
	}
	posterior, err := memory.GetBayesianPosterior(ctx, "tenant-a", "ap:compatible")
	if err != nil {
		t.Fatal(err)
	}
	if posterior.EffectiveSupport != 1 || posterior.Mean() != 2.0/3.0 {
		t.Fatalf("discounted posterior = %+v", posterior)
	}
	if evidence := posterior.MemberEvidence["a"]; evidence.UsefulWeight != 2 {
		t.Fatalf("direct member evidence was discounted: %+v", evidence)
	}
}

func TestStaleCertificateCannotRevokeReplacementIndexes(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	snapshot := memory.Snapshot(ctx)
	old := model.AntiPigeonCertificate{ID: "old", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"}, GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch}
	if _, err := memory.PublishAntiPigeonCertificate(ctx, old); err != nil {
		t.Fatal(err)
	}
	replacement := old
	replacement.ID = "replacement"
	if _, err := memory.PublishAntiPigeonCertificate(ctx, replacement); err != nil {
		t.Fatal(err)
	}
	changePolicy := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 1, MaxMembers: 16}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	for index := 0; index < 8; index++ {
		eventID, useful := "a", true
		if index%2 == 1 {
			eventID, useful = "b", false
		}
		request := model.BayesianOutcomeRequest{IdempotencyKey: fmt.Sprintf("stale-%d", index), TenantID: "tenant-a", EventID: eventID, Useful: useful, AvailableAt: time.Now().UTC(), Source: model.OutcomeFullStream}
		observation := model.ResidualObservation{ActionKey: request.IdempotencyKey, GeneralKey: "stale", PosteriorKey: "ap:old", CommittedProbability: .5, Useful: useful, AvailableAt: request.AvailableAt}
		result, err := memory.ApplyBayesianOutcome(ctx, request, "ap:old", "", request.IdempotencyKey, 1, changePolicy, groupPolicy, observation, residualPolicy)
		if err != nil {
			t.Fatal(err)
		}
		if bayes.RevisionSplits(result.Revision.Action) {
			t.Fatalf("stale certificate gained revision authority: %+v", result.Revision)
		}
	}
	active, err := memory.GetAntiPigeonCertificate(ctx, "tenant-a", []string{"a", "b"})
	if err != nil || active.ID != "replacement" {
		t.Fatalf("replacement index was revoked: %+v, %v", active, err)
	}
}
