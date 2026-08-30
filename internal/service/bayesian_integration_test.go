package service_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestEventFrameDeltaReranksAfterRetrievalContract(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	ranker := &fixedBaseRanker{}
	deltaStore := &recordingRankDeltaStore{}
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 10, DefaultPackK: 2, DefaultTokenBudget: 1000, OverfetchMultiplier: 2,
		CandidateRanker: ranker, CandidateRankerRequired: true,
		RankDeltaStore: deltaStore, RankDeltaStoreRequired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()

	now := time.Now().UTC()
	for _, id := range []string{"one", "two"} {
		event := testutil.Event(id, "shared topic", now.Add(-time.Minute))
		event.Priority = 1
		observe(t, runtime, event)
	}
	snapshot := memory.Snapshot(ctx)
	selection := model.SelectionSupportCertificate{
		ID: "selection-post-rank", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		MinSelectionProbability: .2, SimultaneousCoverage: .95, Procedure: "held-out exhaustive frontier audit v1", Issuer: "test-auditor", ExternalAudit: true,
		ValidFrom: now.Add(-time.Minute), ValidUntil: now.Add(time.Hour),
	}
	if _, err := runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		t.Fatal(err)
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "omitted-post-rank", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		DivergenceUCB: .02, DivergenceLimit: .05, AuditProbability: .02, SimultaneousCoverage: .95,
		Procedure: "simultaneous update-all shadow audit v1", Issuer: "test-auditor", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	if _, err := runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted}); err != nil {
		t.Fatal(err)
	}

	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "shared topic", AsOf: now, RecallK: 10, PackK: 2, TokenBudget: 1000}
	packet, err := runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 20; index++ {
		feedbackAt := time.Now().UTC()
		if _, err := runtime.ObserveBayesianOutcome(ctx, model.BayesianOutcomeRequest{
			ProtocolVersion: model.ProtocolVersion, IdempotencyKey: fmt.Sprintf("post-rank-outcome-%d", index), TenantID: "tenant-a", JournalID: packet.BayesianShadow.JournalID,
			EventID: "one", Useful: true, ObservedAt: feedbackAt, AvailableAt: feedbackAt, Source: model.OutcomeFullStream, InclusionProbability: 1,
		}); err != nil {
			t.Fatal(err)
		}
	}
	request.AsOf = time.Now().UTC()
	packet, err = runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 2 || packet.Candidates[0].Event.ID != "one" {
		t.Fatalf("post-contract EventFrame delta did not rerank final packet: %+v", packet.Candidates)
	}
	winner := packet.Candidates[0]
	if winner.RetrievalScore != .50 || winner.RankDelta <= 0 || winner.Score <= .5001 {
		t.Fatalf("winner did not preserve base score plus positive delta: %+v", winner)
	}
	if deltaStore.putCalls == 0 || len(deltaStore.records) != 1 {
		t.Fatalf("rank deltas were not materialized through the store: %+v", deltaStore)
	}
	if ranker.lastInput["one"] == winner.Score {
		t.Fatalf("retrieval contract received EventFrame-adjusted score: input=%v winner=%+v", ranker.lastInput, winner)
	}
}

type fixedBaseRanker struct{ lastInput map[string]float64 }

func (r *fixedBaseRanker) ContractName() string { return "test/LibraVDBBaseRank" }

func (r *fixedBaseRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	r.lastInput = make(map[string]float64, len(request.Candidates))
	byID := make(map[string]retrieval.Candidate, len(request.Candidates))
	for _, candidate := range request.Candidates {
		r.lastInput[candidate.ID] = candidate.Score
		byID[candidate.ID] = candidate
	}
	one, two := byID["one"], byID["two"]
	one.Score, two.Score = .50, .5001
	return []retrieval.Candidate{two, one}, nil
}

type recordingRankDeltaStore struct {
	records  []rankdelta.Record
	putCalls int
}

func (s *recordingRankDeltaStore) PutBatch(_ context.Context, records []rankdelta.Record) error {
	s.putCalls++
	s.records = append([]rankdelta.Record(nil), records...)
	return nil
}

func (s *recordingRankDeltaStore) GetBatch(_ context.Context, _ string, keys []string, snapshot model.Snapshot, now time.Time) (map[string]rankdelta.Record, error) {
	wanted := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		wanted[key] = struct{}{}
	}
	result := make(map[string]rankdelta.Record, len(keys))
	for _, record := range s.records {
		if _, ok := wanted[record.Key]; ok && record.ValidFor(snapshot, now) {
			result[record.Key] = record
		}
	}
	return result, nil
}

func (*recordingRankDeltaStore) Close() error { return nil }

func TestBayesianPromotionFeedbackAndEpochFallback(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	calibrator := calibration.Logit{Scale: 1, Bias: -0.1, Floor: 1e-6}
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000, OverfetchMultiplier: 2,
		BaselineCalibration: calibrator,
		BayesianPolicy:      bayes.Policy{VectorWeight: .8, IndependenceWeight: .2, Threshold: .7, CriticalThreshold: .5, AuditProbability: 1, MaxActive: 4, AuditSeed: "test"},
		ResidualPolicy:      residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .1, ConfidenceDelta: .5, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001},
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
	stalePacket, err := runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if hasBayesianCandidate(stalePacket.Candidates) {
		t.Fatalf("posterior updated after as_of leaked into stale recall: %+v", stalePacket.Candidates)
	}

	request.AsOf = time.Now().UTC()
	packet, err = runtime.Recall(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if packet.BayesianShadow.Mode != "production" || !hasBayesianCandidate(packet.Candidates) {
		t.Fatalf("post-feedback packet = %+v", packet)
	}
	for _, candidate := range packet.Candidates {
		law := candidate.Forecast.CorrectedLaw
		if candidate.Forecast.HorizonKey != model.RetrievalUsefulnessHorizon || math.Abs(law.Useful+law.NotUseful-1) > 1e-12 || candidate.Forecast.Template.EventID != candidate.Event.ID {
			t.Fatalf("misaligned forecast bundle: %+v", candidate.Forecast)
		}
		if candidate.BayesianApplied && math.Abs(candidate.Forecast.PreResidualLaw.Useful-calibrator.Apply(candidate.PredictiveScore)) > 1e-12 {
			t.Fatalf("belief-conditioned score was not calibrated after composition: %+v", candidate)
		}
	}
	for index := 2; index <= 20; index++ {
		outcomeID := fmt.Sprintf("outcome-%d", index)
		feedbackAt := time.Now().UTC()
		feedback := model.BayesianOutcomeRequest{
			ProtocolVersion: model.ProtocolVersion, IdempotencyKey: outcomeID, TenantID: "tenant-a", JournalID: packet.BayesianShadow.JournalID,
			EventID: "one", Useful: true, ObservedAt: feedbackAt, AvailableAt: feedbackAt, Source: model.OutcomeFullStream, InclusionProbability: 1,
		}
		if _, err := runtime.ObserveBayesianOutcome(ctx, feedback); err != nil {
			t.Fatal(err)
		}
		request.AsOf = time.Now().UTC()
		packet, err = runtime.Recall(ctx, request)
		if err != nil {
			t.Fatal(err)
		}
	}
	if packet.BayesianShadow.ResidualApplied == 0 || !packet.Candidates[0].Forecast.ResidualApplied {
		t.Fatalf("certified residual was not applied: %+v", packet)
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

func TestRuntimeEstimatesAndPublishesDeclaredOmittedInfluencePopulation(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	policy := bayes.Policy{VectorWeight: 1, AuditProbability: 1, MaxActive: 16, AuditSeed: "runtime-audit", CheapUpdateAll: true}
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000, BayesianPolicy: policy})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	event := testutil.Event("nominated", "runtime audit query", now.Add(-time.Hour))
	event.TenantID, event.SessionID = "tenant-a", "session-a"
	if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "runtime audit query", AsOf: now, RecallK: 10, PackK: 10, TokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	population := make([]string, 100)
	observations := make([]model.OmittedInfluenceAuditObservation, 100)
	for index := range population {
		population[index] = fmt.Sprintf("omitted-%03d", index)
		observations[index] = model.OmittedInfluenceAuditObservation{EventID: population[index], ExpandedLaw: model.BernoulliLaw{Useful: .5, NotUseful: .5}}
	}
	response, err := runtime.EstimateOmittedInfluence(ctx, model.EstimateOmittedInfluenceRequest{
		ProtocolVersion: model.ProtocolVersion, ID: "runtime-audit-1", TenantID: "tenant-a", JournalID: packet.BayesianShadow.JournalID,
		PopulationEventIDs: population, LocalLaw: model.BernoulliLaw{Useful: .5, NotUseful: .5}, Observations: observations,
		AuditSequence: 1, ConfidenceDelta: .05, DivergenceLimit: .2, ValidUntil: time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := memory.GetOmittedInfluenceCertificate(ctx, "tenant-a")
	if err != nil || response.CertificateID != certificate.ID || !certificate.RuntimeEstimated || certificate.ExternalAudit {
		t.Fatalf("certificate = %+v, response=%+v, err=%v", certificate, response, err)
	}
	sameQuery, err := runtime.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "runtime audit query", AsOf: now, RecallK: 10, PackK: 10, TokenBudget: 1000})
	if err != nil || !sameQuery.BayesianShadow.OmittedInfluenceCertified {
		t.Fatalf("query-bound certificate was not active: %+v, %v", sameQuery.BayesianShadow, err)
	}
	differentQuery, err := runtime.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "different query", AsOf: now, RecallK: 10, PackK: 10, TokenBudget: 1000})
	if err != nil || differentQuery.BayesianShadow.OmittedInfluenceCertified {
		t.Fatalf("query-bound certificate leaked across queries: %+v, %v", differentQuery.BayesianShadow, err)
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
