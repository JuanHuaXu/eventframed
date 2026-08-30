package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestHigherOrderCompositionIsAuthorizedResolutionAwareAndReversible(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000,
		CandidateRanker: equalCompositionRanker{},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	memberIDs := []string{"stage-a", "stage-b", "stage-c"}
	for index, id := range memberIDs {
		event := testutil.Event(id, "mission stage", now.Add(-time.Duration(4-index)*time.Minute))
		event.Sequence = uint64(index + 1)
		event.What.Value = "mission stage"
		if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	beforeCertificate := memory.Snapshot(ctx)
	certificate := model.AntiPigeonCertificate{
		ID: "mission-compatible", TenantID: "tenant-a", MemberEventIDs: memberIDs,
		HorizonKey: model.RetrievalUsefulnessHorizon, GraphVersion: beforeCertificate.GraphVersion, EvidenceEpoch: beforeCertificate.EvidenceEpoch,
		TargetDiameterUCB: .01, DiameterLimit: .05, EffectiveSupport: 40, MinEffectiveSupport: 30,
		SimultaneousCoverage: .95, Procedure: "held-out mission-stage compatibility audit", Issuer: "test-auditor", ExternalAudit: true,
		ValidUntil: now.Add(time.Hour),
	}
	publishedCertificate, err := runtime.PublishAntiPigeonCertificate(ctx, model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: certificate})
	if err != nil {
		t.Fatal(err)
	}
	request := model.ComposeInvariantRequest{
		ProtocolVersion: model.ProtocolVersion, ID: "mission", TenantID: "tenant-a", SessionID: "session-a",
		MemberEventIDs: memberIDs, RepresentativeEventID: "stage-a", Label: "complete mission",
		RuleID: "ordered-mission-stages-v1", Resolution: "mission", Confidence: .9,
		AntiPigeonCertificateID: certificate.ID, PublishedAt: now, BaseSnapshot: publishedCertificate.Snapshot,
	}
	composed, err := runtime.ComposeInvariant(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if composed.Event.Composition == nil || composed.Event.Composition.RepresentativeEventID != "stage-a" || composed.Event.Provenance.Producer != "eventframed-invariant-seeker" {
		t.Fatalf("composition = %+v", composed.Event)
	}
	if composed.Snapshot.EvidenceEpoch != request.BaseSnapshot.EvidenceEpoch || composed.Snapshot.AbstractionVersion != request.BaseSnapshot.AbstractionVersion+1 {
		t.Fatalf("composition version motion = before=%+v after=%+v", request.BaseSnapshot, composed.Snapshot)
	}
	duplicate, err := runtime.ComposeInvariant(ctx, request)
	if err != nil || !duplicate.Duplicate || duplicate.Event.ID != composed.Event.ID {
		t.Fatalf("idempotent composition retry = %+v, %v", duplicate, err)
	}

	coarse := recallComposition(t, runtime, now.Add(time.Second), model.RecallResolutionCoarse)
	if coarse.Candidates[0].Event.ID != "mission" || coarse.Candidates[0].ResolutionRankDelta <= 0 {
		t.Fatalf("coarse recall = %+v", candidateIDs(coarse.Candidates))
	}
	fine := recallComposition(t, runtime, now.Add(time.Second), model.RecallResolutionFine)
	if fine.Candidates[0].Event.Composition != nil || fine.Candidates[len(fine.Candidates)-1].Event.ID != "mission" {
		t.Fatalf("fine recall = %+v", candidateIDs(fine.Candidates))
	}
	if !containsCandidate(fine.Candidates, "stage-a") {
		t.Fatal("representative member was not directly retrievable")
	}
	replacement := certificate
	replacement.ID = "mission-recertified-separate"
	replacementSnapshot := memory.Snapshot(ctx)
	replacement.GraphVersion = replacementSnapshot.GraphVersion
	replacement.EvidenceEpoch = replacementSnapshot.EvidenceEpoch
	if _, err := runtime.PublishAntiPigeonCertificate(ctx, model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: replacement}); err != nil {
		t.Fatal(err)
	}
	revoked := recallComposition(t, runtime, now.Add(time.Second), model.RecallResolutionCoarse)
	if containsCandidate(revoked.Candidates, "mission") {
		t.Fatal("higher-order event remained retrievable after its sharing authority was replaced")
	}

	forged := composed.Event
	forged.ID = "forged"
	if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: forged.ID, Event: forged}); err == nil {
		t.Fatal("ordinary observation accepted a forged higher-order event")
	}
	decomposed, err := runtime.DecomposeInvariant(ctx, model.DecomposeInvariantRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "mission", Reason: "invariant no longer holds"})
	if err != nil || !decomposed.Deleted || len(decomposed.RestoredMemberEventIDs) != len(memberIDs) {
		t.Fatalf("decomposition = %+v, %v", decomposed, err)
	}
	tombstone, err := memory.GetCompositionTombstone(ctx, "tenant-a", "mission")
	if err != nil || tombstone.Reason != "invariant no longer holds" {
		t.Fatalf("decomposition audit = %+v, %v", tombstone, err)
	}
	after := recallComposition(t, runtime, now.Add(time.Second), model.RecallResolutionCoarse)
	if containsCandidate(after.Candidates, "mission") {
		t.Fatal("decomposed higher-order event remained retrievable")
	}
	for _, id := range memberIDs {
		if !containsCandidate(after.Candidates, id) {
			t.Fatalf("constituent %q was lost after decomposition", id)
		}
	}
}

func TestHigherOrderCompositionRejectsMissingAuthorityAndStaleSnapshot(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	now := time.Now().UTC()
	for _, id := range []string{"a", "b"} {
		event := testutil.Event(id, "shared", now.Add(-time.Minute))
		_, _ = runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event})
	}
	request := model.ComposeInvariantRequest{ProtocolVersion: model.ProtocolVersion, ID: "ab", TenantID: "tenant-a", SessionID: "session-a", MemberEventIDs: []string{"a", "b"}, RepresentativeEventID: "a", Label: "shared invariant", RuleID: "rule", Resolution: "episode", Confidence: .8, AntiPigeonCertificateID: "missing", PublishedAt: now, BaseSnapshot: memory.Snapshot(ctx)}
	if _, err := runtime.ComposeInvariant(ctx, request); !errors.Is(err, store.ErrCompositionAuthority) {
		t.Fatalf("missing authority error = %v", err)
	}
	request.BaseSnapshot.RuntimeVersion--
	if _, err := runtime.ComposeInvariant(ctx, request); !errors.Is(err, store.ErrStaleSnapshot) {
		t.Fatalf("stale snapshot error = %v", err)
	}
}

type equalCompositionRanker struct{}

func (equalCompositionRanker) ContractName() string { return "test/equal-composition-ranker" }
func (equalCompositionRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	result := append([]retrieval.Candidate(nil), request.Candidates...)
	for index := range result {
		result[index].Score = .5
	}
	return result, nil
}

func recallComposition(t *testing.T, runtime *service.Service, asOf time.Time, resolution model.RecallResolution) model.ContextPacket {
	t.Helper()
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "mission", AsOf: asOf, RecallK: 10, PackK: 10, TokenBudget: 1000, Resolution: resolution})
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

func containsCandidate(candidates []model.Candidate, id string) bool {
	for _, candidate := range candidates {
		if candidate.Event.ID == id {
			return true
		}
	}
	return false
}

func candidateIDs(candidates []model.Candidate) []string {
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.Event.ID)
	}
	return ids
}
