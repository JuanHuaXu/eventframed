package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestBayesianPromotionFeedbackAndEpochFallback(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000, OverfetchMultiplier: 2,
		BayesianPolicy: bayes.Policy{VectorWeight: .8, IndependenceWeight: .2, Threshold: .7, CriticalThreshold: .5, AuditProbability: 1, MaxActive: 4, AuditSeed: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, id := range []string{"one", "two"} {
		event := testutil.Event(id, "shared topic", now.Add(-time.Minute))
		event.Priority = 1
		observe(t, runtime, event)
	}
	snapshot := memory.Snapshot(ctx)
	selection := model.SelectionSupportCertificate{
		ID: "selection-1", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		MinSelectionProbability: .2, SimultaneousCoverage: .95, Procedure: "held-out exhaustive frontier audit v1", Issuer: "test-auditor", ExternalAudit: true,
		ValidFrom: now.Add(-time.Minute), ValidUntil: now.Add(time.Hour),
	}
	if _, err := runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		t.Fatal(err)
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "omitted-1", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		DivergenceUCB: .02, DivergenceLimit: .05, AuditProbability: 1, SimultaneousCoverage: .95,
		Procedure: "simultaneous update-all shadow audit v1", Issuer: "test-auditor", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	if _, err := runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted}); err != nil {
		t.Fatal(err)
	}
	antiPigeon := model.AntiPigeonCertificate{
		ID: "bucket-1", TenantID: "tenant-a", MemberEventIDs: []string{"one", "two"}, HorizonKey: "retrieval-usefulness-v1",
		GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch, TargetDiameterUCB: .05, DiameterLimit: .1,
		EffectiveSupport: 50, MinEffectiveSupport: 30, SimultaneousCoverage: .95, Procedure: "external target-law diameter audit v1", Issuer: "test-auditor", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	if _, err := runtime.PublishAntiPigeonCertificate(ctx, model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: antiPigeon}); err != nil {
		t.Fatal(err)
	}
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "shared topic", AsOf: now, RecallK: 10, PackK: 10, TokenBudget: 1000}
	packet, err := runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if packet.BayesianShadow.Mode != "certified-shadow" || !packet.BayesianShadow.SelectionSupportCertified {
		t.Fatalf("pre-feedback report = %+v", packet.BayesianShadow)
	}
	decision := findDecision(t, packet.BayesianShadow, "one")
	if !decision.Activated || decision.PosteriorKey != "ap:bucket-1" || decision.TotalSelectionProbabilityLowerBound != .2 {
		t.Fatalf("decision = %+v", decision)
	}
	outcome := model.BayesianOutcomeRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "outcome-1", TenantID: "tenant-a", JournalID: packet.BayesianShadow.JournalID,
		EventID: "one", Useful: true, ObservedAt: now.Add(time.Second), AvailableAt: now.Add(time.Second), Source: model.OutcomeFullStream, InclusionProbability: 1,
	}
	// The service rejects evidence from the future even when its event time is plausible.
	if _, err := runtime.ObserveBayesianOutcome(ctx, outcome); err == nil {
		t.Fatal("expected future-availability rejection")
	}
	outcome.ObservedAt = time.Now().UTC()
	outcome.AvailableAt = outcome.ObservedAt
	first, err := runtime.ObserveBayesianOutcome(ctx, outcome)
	if err != nil {
		t.Fatal(err)
	}
	if first.Posterior.Mean() <= .5 || first.Posterior.PosteriorKey != "ap:bucket-1" {
		t.Fatalf("posterior = %+v", first.Posterior)
	}
	duplicate, err := runtime.ObserveBayesianOutcome(ctx, outcome)
	if err != nil || !duplicate.Duplicate {
		t.Fatalf("duplicate outcome = %+v, %v", duplicate, err)
	}
	conflict := outcome
	conflict.Useful = false
	if _, err := runtime.ObserveBayesianOutcome(ctx, conflict); !errors.Is(err, store.ErrOutcomeConflict) {
		t.Fatalf("conflicting outcome error = %v", err)
	}

	packet, err = runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if packet.BayesianShadow.Mode != "production" || !hasBayesianCandidate(packet.Candidates) {
		t.Fatalf("post-feedback packet = %+v", packet)
	}
	newEvent := testutil.Event("epoch-change", "shared topic", now.Add(-time.Second))
	observe(t, runtime, newEvent)
	packet, err = runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if packet.BayesianShadow.Mode != "shadow" || packet.BayesianShadow.SelectionSupportCertified || hasBayesianCandidate(packet.Candidates) {
		t.Fatalf("stale certificate or posterior survived epoch change: %+v", packet)
	}
}

func findDecision(t *testing.T, report model.BayesianShadowReport, eventID string) model.BayesianDecision {
	t.Helper()
	for _, decision := range report.Decisions {
		if decision.EventID == eventID {
			return decision
		}
	}
	t.Fatalf("decision %q not found", eventID)
	return model.BayesianDecision{}
}

func hasBayesianCandidate(candidates []model.Candidate) bool {
	for _, candidate := range candidates {
		if candidate.BayesianApplied {
			return true
		}
	}
	return false
}
