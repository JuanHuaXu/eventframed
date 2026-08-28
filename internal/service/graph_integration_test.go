package service_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestPredictiveSnapPublishesInvalidatesRejectsStaleAndRollsBack(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	outcome := model.BayesianOutcomeRequest{IdempotencyKey: "outcome-1", TenantID: "tenant-a", Useful: true, AvailableAt: now}
	observation := model.ResidualObservation{
		ActionKey: "action", GeneralKey: "general", HorizonKey: model.RetrievalUsefulnessHorizon,
		BaseProbability: .5, CommittedProbability: .5, Useful: true, ValidationEligible: true,
		EventID: "event-a", JournalID: "journal", PosteriorKey: "posterior-a", AvailableAt: now,
	}
	if _, err := memory.ApplyBayesianOutcome(ctx, outcome, "posterior-a", "digest", 1, bayes.ChangePolicy{Hazard: .05, Threshold: .3, MaxRun: 64}, observation, residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}); err != nil {
		t.Fatal(err)
	}

	base := memory.Snapshot(ctx)
	request := validSnapRequest(base, now)
	nonFinite := request
	nonFinite.BucketCertificates = append([]model.SnapBucketCertificate(nil), request.BucketCertificates...)
	nonFinite.BucketCertificates[0].EffectiveSupport = math.NaN()
	if _, err := runtime.PublishPredictiveSnap(ctx, nonFinite); err == nil {
		t.Fatal("expected non-finite bucket support rejection")
	}
	published, err := runtime.PublishPredictiveSnap(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if !published.Accepted || published.Graph.SourceSnapID != request.ID || published.Snapshot.GraphVersion != base.GraphVersion+1 {
		t.Fatalf("published snap = %+v", published)
	}
	if published.Snapshot.EvidenceEpoch != base.EvidenceEpoch {
		t.Fatalf("predictive snap changed evidence epoch: before=%d after=%d", base.EvidenceEpoch, published.Snapshot.EvidenceEpoch)
	}
	posterior, err := memory.GetBayesianPosterior(ctx, "tenant-a", "posterior-a")
	if err != nil || posterior.Certified {
		t.Fatalf("affected posterior = %+v, %v", posterior, err)
	}
	residuals, err := memory.GetResidualCandidates(ctx, "tenant-a", "action", "general")
	if err != nil || residuals.Exact == nil || residuals.General == nil || residuals.Exact.Active || residuals.General.Active {
		t.Fatalf("affected residuals = %+v, %v", residuals, err)
	}
	otherTenant, err := memory.GetPredictiveGraph(ctx, "tenant-b")
	if err != nil || otherTenant.Version != published.Snapshot.GraphVersion {
		t.Fatalf("other tenant graph epoch = %+v, %v", otherTenant, err)
	}
	if _, err := runtime.PublishPredictiveSnap(ctx, request); !errors.Is(err, store.ErrStaleSnapshot) {
		t.Fatalf("stale publish error = %v", err)
	}

	rejectedRequest := validSnapRequest(published.Snapshot, now)
	rejectedRequest.ID = "snap-rejected"
	rejectedRequest.Candidate.Edges[0].Weight = 2
	rejectedRequest.PriorityGainLCB = .01
	rejected, err := runtime.PublishPredictiveSnap(ctx, rejectedRequest)
	if err != nil || rejected.Accepted || rejected.Snapshot != published.Snapshot {
		t.Fatalf("statistical rejection = %+v, %v", rejected, err)
	}

	rolledBack, err := runtime.RollbackPredictiveSnap(ctx, model.RollbackSnapRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SnapID: request.ID, Reason: "confirmation regression"})
	if err != nil {
		t.Fatal(err)
	}
	if !rolledBack.Accepted || len(rolledBack.Graph.Nodes) != 0 || rolledBack.Graph.SourceSnapID != "rollback:"+request.ID || rolledBack.Snapshot.GraphVersion != published.Snapshot.GraphVersion+1 {
		t.Fatalf("rollback = %+v", rolledBack)
	}
	if _, err := runtime.RollbackPredictiveSnap(ctx, model.RollbackSnapRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SnapID: request.ID, Reason: "duplicate"}); !errors.Is(err, store.ErrSnapConflict) {
		t.Fatalf("duplicate rollback error = %v", err)
	}
}

func validSnapRequest(base model.Snapshot, now time.Time) model.PredictiveSnapRequest {
	return model.PredictiveSnapRequest{
		ProtocolVersion: model.ProtocolVersion, ID: "snap-1", TenantID: "tenant-a", BaseSnapshot: base,
		Candidate: model.PredictiveGraph{TenantID: "tenant-a", Nodes: []model.CompatibilityNode{
			{ID: "bucket-a", Kind: "bucket", MemberEventIDs: []string{"event-a"}, PosteriorKeys: []string{"posterior-a"}, LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "predictor-a", Kind: "predictor", LawSpace: model.RetrievalUsefulnessHorizon},
		}, Edges: []model.CompatibilityEdge{{ID: "edge-a", From: "bucket-a", To: "predictor-a", ComparisonMap: "identity_bernoulli", Weight: 1}}},
		Obligations:        []model.ComparisonObligation{{From: "bucket-a", To: "predictor-a", Weight: 1}},
		BucketCertificates: []model.SnapBucketCertificate{{NodeID: "bucket-a", FutureDiameterUCB: .03, DiameterLimit: .05, EffectiveSupport: 50}},
		EdgeCertificates:   []model.SnapEdgeCertificate{{EdgeID: "edge-a", DefectUCB: .01, DefectLimit: .05}},
		DesignStart:        now.Add(-4 * time.Hour), DesignEnd: now.Add(-3 * time.Hour), ConfirmationStart: now.Add(-2 * time.Hour), ConfirmationEnd: now.Add(-time.Hour),
		CandidateFamilySize: 2, UnchangedCandidateIncluded: true, PriorityGainLCB: .1, ResourceCostUCB: .01, ProperRiskIncreaseUCB: .001,
		SimultaneousCoverage: .95, Procedure: "held-out bounded split-merge confirmation v1", Issuer: "external-auditor", ExternalAudit: true,
	}
}
