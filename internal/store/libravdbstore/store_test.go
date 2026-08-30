package libravdbstore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	storecontract "github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func testConfig(path, modelKey string) libravdbstore.Config {
	return libravdbstore.Config{Path: path, Dimension: 4, Quantization: "none", MemoryMapping: true, EmbeddingModel: modelKey}
}

func TestLegacyMigrationRequiresBackupAndRecoversRecords(t *testing.T) {
	root := t.TempDir()
	path := root + "/legacy.libravdb"
	raw, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	collection, err := raw.CreateCollection(context.Background(), "events_legacy", libra.WithDimension(4))
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("legacy", "legacy event", time.Now().UTC())
	encoded, _ := json.Marshal(event)
	if err := collection.Insert(context.Background(), event.ID, []float32{1, 0, 0, 0}, map[string]interface{}{"event_json": string(encoded), "content_digest": "old"}); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	config := testConfig(path, "model-a:d4")
	if _, err := libravdbstore.Open(config); err == nil {
		t.Fatal("expected legacy migration requirement")
	}
	backup := root + "/legacy-backup.libravdb"
	if err := libravdbstore.MigrateLegacy(context.Background(), config, backup); err != nil {
		t.Fatal(err)
	}
	migrated, err := libravdbstore.Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	results, err := migrated.Search(context.Background(), "tenant-a", []float32{1, 0, 0, 0}, time.Now().Add(time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Event.EmbeddingModel != "model-a:d4" {
		t.Fatalf("migrated results = %+v", results)
	}
}

func TestStoreRoundTripAndAvailabilityGate(t *testing.T) {
	store, err := libravdbstore.Open(libravdbstore.Config{
		Path: t.TempDir() + "/events.libravdb", Dimension: 4, Quantization: "none", MemoryMapping: true, EmbeddingModel: "test:d4",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	now := time.Now().UTC()
	past := testutil.Event("past", "remember this", now.Add(-time.Minute))
	future := testutil.Event("future", "remember this", now.Add(time.Hour))
	vector := []float32{1, 0, 0, 0}
	if _, err := store.Put(context.Background(), past, vector, "past-digest"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Put(context.Background(), future, vector, "future-digest"); err != nil {
		t.Fatal(err)
	}
	results, err := store.Search(context.Background(), "tenant-a", vector, now, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Event.ID != "past" {
		t.Fatalf("results = %+v", results)
	}
	stats, err := store.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Tenants != 1 || stats.Events != 2 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestDurableStateSurvivesRestartAndPinsEmbeddingContract(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("one", "durable", time.Now().UTC())
	result, err := first.Put(context.Background(), event, []float32{1, 0, 0, 0}, "digest")
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.RuntimeVersion != 1 || result.Snapshot.EvidenceEpoch != 1 {
		t.Fatalf("snapshot = %+v", result.Snapshot)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := second.Snapshot(context.Background()); snapshot.RuntimeVersion != 1 || snapshot.EvidenceEpoch != 1 {
		t.Fatalf("reopened snapshot = %+v", snapshot)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := libravdbstore.Open(testConfig(path, "model-b:d4")); err == nil {
		t.Fatal("expected embedding contract mismatch")
	}
}

func TestBayesianJournalSurvivesRestartAndRejectsConflict(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	entry := model.BayesianJournalEntry{
		ID: "bj_stable", TenantID: "tenant-a", SessionID: "session-a",
		AsOf: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC), QueryDigest: "digest",
		Snapshot: first.Snapshot(context.Background()),
		Report:   model.BayesianShadowReport{Mode: "shadow", JournalID: "bj_stable", JournalDurable: true, Nominated: 1},
	}
	if err := first.PutBayesianJournal(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if err := first.PutBayesianJournal(context.Background(), entry); err != nil {
		t.Fatalf("idempotent journal write: %v", err)
	}
	conflict := entry
	conflict.QueryDigest = "different"
	if err := first.PutBayesianJournal(context.Background(), conflict); err == nil {
		t.Fatal("expected conflicting journal write to fail")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetBayesianJournal(context.Background(), entry.TenantID, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.QueryDigest != entry.QueryDigest || got.Report.Nominated != 1 {
		t.Fatalf("journal after restart = %+v", got)
	}
}

func TestBayesianPosteriorUpdateSurvivesRestart(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	request := model.BayesianOutcomeRequest{IdempotencyKey: "outcome-1", TenantID: "tenant-a", EventID: "event-1", Useful: true, AvailableAt: now}
	residualObservation := model.ResidualObservation{ActionKey: "action", GeneralKey: "general", HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: .5, Useful: true, ValidationEligible: true, EventID: "event-1", JournalID: "journal", AvailableAt: now}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	result, err := first.ApplyBayesianOutcome(context.Background(), request, "event-1", "global:retrieval-usefulness-v1", "digest", 2, bayes.ChangePolicy{Hazard: .05, Threshold: .3, MaxRun: 64}, bayes.GroupPolicy{}, residualObservation, residualPolicy)
	if err != nil {
		t.Fatal(err)
	}
	if result.Posterior.Mean() != .75 || result.Snapshot.PosteriorVersion != 2 {
		t.Fatalf("outcome result = %+v", result)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	posterior, err := second.GetBayesianPosterior(context.Background(), "tenant-a", "event-1")
	if err != nil {
		t.Fatal(err)
	}
	evidence := posterior.MemberEvidence["event-1"]
	if posterior.Mean() != .75 || evidence.UsefulWeight != 2 || evidence.NotUsefulWeight != 0 || second.Snapshot(context.Background()).PosteriorVersion != 2 {
		t.Fatalf("posterior after restart = %+v", posterior)
	}
	parent, err := second.GetBayesianPosterior(context.Background(), "tenant-a", "global:retrieval-usefulness-v1")
	if err != nil || parent.Mean() != .75 || parent.EffectiveSupport != 2 || len(parent.MemberEvidence) != 0 {
		t.Fatalf("parent posterior after restart = %+v, %v", parent, err)
	}
	candidates, err := second.GetResidualCandidates(context.Background(), "tenant-a", "action", "general")
	if err != nil || candidates.Exact == nil || candidates.General == nil {
		t.Fatalf("residuals after restart = %+v, %v", candidates, err)
	}
}

func TestAntiPigeonRevisionSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	snapshot := first.Snapshot(ctx)
	certificate := model.AntiPigeonCertificate{ID: "bucket-a", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"}, GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch}
	if _, err := first.PublishAntiPigeonCertificate(ctx, certificate); err != nil {
		t.Fatal(err)
	}
	changePolicy := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	var result storecontract.BayesianOutcomeResult
	for index := 0; index < 20; index++ {
		eventID, useful := "a", true
		if index%2 == 1 {
			eventID, useful = "b", false
		}
		request := model.BayesianOutcomeRequest{IdempotencyKey: fmt.Sprintf("revision-%d", index), TenantID: "tenant-a", EventID: eventID, Useful: useful, AvailableAt: now.Add(time.Duration(index) * time.Second), Source: model.OutcomeIndependentAudit}
		observation := model.ResidualObservation{ActionKey: request.IdempotencyKey, GeneralKey: "revision-general", PosteriorKey: "ap:bucket-a", CommittedProbability: .5, Useful: useful, AvailableAt: request.AvailableAt}
		result, err = first.ApplyBayesianOutcome(ctx, request, "ap:bucket-a", "", request.IdempotencyKey, 1, changePolicy, groupPolicy, observation, residualPolicy)
		if err != nil {
			t.Fatal(err)
		}
		if bayes.RevisionSplits(result.Revision.Action) {
			break
		}
	}
	if result.Revision.Action != model.BayesianRevisionSplit {
		t.Fatalf("revision = %+v", result.Revision)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if _, err := second.GetAntiPigeonCertificate(ctx, "tenant-a", []string{"a"}); !errors.Is(err, storecontract.ErrCertificateNotFound) {
		t.Fatalf("sharing certificate survived restart: %v", err)
	}
	left, leftErr := second.GetBayesianPosterior(ctx, "tenant-a", "a")
	right, rightErr := second.GetBayesianPosterior(ctx, "tenant-a", "b")
	shared, sharedErr := second.GetBayesianPosterior(ctx, "tenant-a", "ap:bucket-a")
	if leftErr != nil || rightErr != nil || sharedErr != nil || left.Mean() <= .5 || right.Mean() >= .5 || shared.Certified {
		t.Fatalf("durable revision left=%+v/%v right=%+v/%v shared=%+v/%v", left, leftErr, right, rightErr, shared, sharedErr)
	}
}

func TestSharedEvidenceDiscountSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	snapshot := first.Snapshot(ctx)
	certificate := model.AntiPigeonCertificate{ID: "compatible", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"}, GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch}
	if _, err := first.PublishAntiPigeonCertificate(ctx, certificate); err != nil {
		t.Fatal(err)
	}
	changePolicy := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 100, MaxMembers: 16, SharedEvidenceWeight: .25}
	residualPolicy := residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	for index, eventID := range []string{"a", "b", "a", "b"} {
		request := model.BayesianOutcomeRequest{IdempotencyKey: fmt.Sprintf("shared-%d", index), TenantID: "tenant-a", EventID: eventID, Useful: true, AvailableAt: now.Add(time.Duration(index) * time.Second), Source: model.OutcomeIndependentAudit}
		observation := model.ResidualObservation{ActionKey: request.IdempotencyKey, GeneralKey: "shared", PosteriorKey: "ap:compatible", CommittedProbability: .5, Useful: true, AvailableAt: request.AvailableAt}
		if _, err := first.ApplyBayesianOutcome(ctx, request, "ap:compatible", "", request.IdempotencyKey, 1, changePolicy, groupPolicy, observation, residualPolicy); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	posterior, err := second.GetBayesianPosterior(ctx, "tenant-a", "ap:compatible")
	if err != nil {
		t.Fatal(err)
	}
	if posterior.EffectiveSupport != 1 || posterior.Mean() != 2.0/3.0 || posterior.MemberEvidence["a"].UsefulWeight != 2 {
		t.Fatalf("durable discounted posterior = %+v", posterior)
	}
}

func TestPredictiveSnapInvalidationRollbackAndRestart(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
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
	if _, err := first.ApplyBayesianOutcome(ctx, outcome, "posterior-a", "", "digest", 1, bayes.ChangePolicy{Hazard: .05, Threshold: .3, MaxRun: 64}, bayes.GroupPolicy{}, observation, residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}); err != nil {
		t.Fatal(err)
	}
	previous, err := first.GetPredictiveGraph(ctx, "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	published := model.PredictiveGraph{
		TenantID: "tenant-a", SourceSnapID: "snap-1", PublishedAt: now,
		Nodes: []model.CompatibilityNode{{ID: "bucket-a", Kind: "bucket", MemberEventIDs: []string{"event-a"}, PosteriorKeys: []string{"posterior-a"}, LawSpace: model.RetrievalUsefulnessHorizon}},
	}
	closure := model.DependencyClosure{NodeIDs: []string{"bucket-a"}, EventIDs: []string{"event-a"}, PosteriorKeys: []string{"posterior-a"}}
	record := model.PredictiveSnapRecord{ID: "snap-1", TenantID: "tenant-a", PreviousGraph: previous, PublishedGraph: published, Closure: closure, SimultaneousCoverage: .95, Procedure: "test", Issuer: "external-auditor", PublishedAt: now}
	graph, snap, err := first.PublishPredictiveSnap(ctx, record)
	if err != nil {
		t.Fatal(err)
	}
	if graph.SourceSnapID != "snap-1" || snap.GraphVersion != previous.Version+1 {
		t.Fatalf("published graph = %+v, snapshot = %+v", graph, snap)
	}
	posterior, err := first.GetBayesianPosterior(ctx, "tenant-a", "posterior-a")
	if err != nil || posterior.Certified {
		t.Fatalf("invalidated posterior = %+v, %v", posterior, err)
	}
	residuals, err := first.GetResidualCandidates(ctx, "tenant-a", "action", "general")
	if err != nil || residuals.Exact == nil || residuals.Exact.Active || residuals.General == nil || residuals.General.Active {
		t.Fatalf("invalidated residuals = %+v, %v", residuals, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := second.GetPredictiveGraph(ctx, "tenant-a")
	if err != nil || reopened.SourceSnapID != "snap-1" || reopened.Version != second.Snapshot(ctx).GraphVersion {
		t.Fatalf("reopened graph = %+v, %v", reopened, err)
	}
	other, err := second.GetPredictiveGraph(ctx, "tenant-b")
	if err != nil || other.Version != reopened.Version {
		t.Fatalf("other tenant graph epoch = %+v, %v", other, err)
	}
	rolledBack, rollbackSnapshot, err := second.RollbackPredictiveSnap(ctx, "tenant-a", "snap-1", "confirmation regression")
	if err != nil {
		t.Fatal(err)
	}
	if len(rolledBack.Nodes) != 0 || rolledBack.SourceSnapID != "rollback:snap-1" || rollbackSnapshot.GraphVersion != reopened.Version+1 {
		t.Fatalf("rollback graph = %+v, snapshot = %+v", rolledBack, rollbackSnapshot)
	}
	reusedID := record
	reusedID.PreviousGraph = rolledBack
	if _, _, err := second.PublishPredictiveSnap(ctx, reusedID); err == nil {
		t.Fatal("expected reused snap id with different content to fail")
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}

	third, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer third.Close()
	finalGraph, err := third.GetPredictiveGraph(ctx, "tenant-a")
	if err != nil || finalGraph.SourceSnapID != "rollback:snap-1" || finalGraph.Version != third.Snapshot(ctx).GraphVersion {
		t.Fatalf("durable rollback graph = %+v, %v", finalGraph, err)
	}
}

func TestAgencyProposalLeaseAndResolutionSurviveRestart(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	proposal, err := agency.BuildProposal(model.AgencyProposalDraft{
		ID: "proposal-1", TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyWake,
		Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: .7,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: "proposal-1", CausalChainID: "chain-1",
	}, now, agency.DefaultPolicy(true))
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(proposal)
	if err != nil {
		t.Fatal(err)
	}
	record := model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	evidence := testutil.Event("event-a", "agency evidence", now.Add(-time.Second))
	if _, err := first.Put(ctx, evidence, []float32{1, 0, 0, 0}, "event-digest"); err != nil {
		t.Fatal(err)
	}
	put, err := first.PutAgencyProposal(ctx, record, "digest", 8, 1000, now)
	if err != nil || put.Snapshot.AgencyVersion != 2 {
		t.Fatalf("agency put = %+v, %v", put, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	claimed, claimSnapshot, err := second.ClaimAgencyProposals(ctx, "tenant-a", "authority-a", now.Add(time.Second), 10, time.Minute)
	if err != nil || len(claimed) != 1 || claimed[0].Status != model.AgencyClaimed || claimSnapshot.AgencyVersion != 3 {
		t.Fatalf("agency claim = %+v, %+v, %v", claimed, claimSnapshot, err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}

	third, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	resolve := model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: "proposal-1", ConsumerID: "authority-a", Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-1"}
	approved, err := third.ResolveAgencyProposal(ctx, resolve, now.Add(2*time.Second))
	if err != nil || approved.Record.Status != model.AgencyApproved || approved.Snapshot.AgencyVersion != 4 {
		t.Fatalf("agency resolution = %+v, %v", approved, err)
	}
	if err := third.Close(); err != nil {
		t.Fatal(err)
	}

	fourth, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := fourth.ResolveAgencyProposal(ctx, resolve, now.Add(3*time.Second))
	if err != nil || !duplicate.Duplicate || duplicate.Record.ExecutionRef != "job-1" || fourth.Snapshot(ctx).AgencyVersion != 4 {
		t.Fatalf("durable agency resolution = %+v, %v", duplicate, err)
	}

	secondProposal, err := agency.BuildProposal(model.AgencyProposalDraft{
		ID: "proposal-2", TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
		Reason: "A second follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .7, Priority: .6,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: "proposal-2", CausalChainID: "chain-2",
	}, now, agency.DefaultPolicy(true))
	if err != nil {
		t.Fatal(err)
	}
	secondSigned, err := signer.Sign(secondProposal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fourth.PutAgencyProposal(ctx, model.AgencyProposalRecord{Proposal: secondProposal, Signed: secondSigned, Status: model.AgencyPending}, "digest-2", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	beforeDelete := fourth.Snapshot(ctx)
	deleted, err := fourth.Delete(ctx, "tenant-a", "event-a")
	if err != nil || !deleted.Deleted || deleted.Snapshot.AgencyVersion != beforeDelete.AgencyVersion+1 {
		t.Fatalf("agency-aware durable delete = %+v, %v", deleted, err)
	}
	if err := fourth.Close(); err != nil {
		t.Fatal(err)
	}

	fifth, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer fifth.Close()
	remaining, _, err := fifth.ClaimAgencyProposals(ctx, "tenant-a", "authority-b", now.Add(4*time.Second), 10, time.Minute)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("deleted evidence left durable proposal claimable: %+v, %v", remaining, err)
	}
}

func TestExpiredAgencyProposalDoesNotBlockDurableQueueAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	runtime, err := libravdbstore.Open(testConfig(t.TempDir()+"/events.libravdb", "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "agency evidence", now), []float32{1, 0, 0, 0}, "event-digest"); err != nil {
		t.Fatal(err)
	}
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	put := func(id string, builtAt, expiresAt, availableBy time.Time) error {
		proposal, buildErr := agency.BuildProposal(model.AgencyProposalDraft{
			ID: id, TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
			Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: .7,
			NotBefore: builtAt, ExpiresAt: expiresAt, IdempotencyKey: id, CausalChainID: "chain-" + id,
		}, builtAt, agency.DefaultPolicy(true))
		if buildErr != nil {
			return buildErr
		}
		signed, signErr := signer.Sign(proposal)
		if signErr != nil {
			return signErr
		}
		_, putErr := runtime.PutAgencyProposal(ctx, model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}, "digest-"+id, 8, 1, availableBy)
		return putErr
	}
	if err := put("proposal-old", now, now.Add(time.Minute), now); err != nil {
		t.Fatal(err)
	}
	if err := put("proposal-new", now.Add(2*time.Minute), now.Add(time.Hour), now.Add(2*time.Minute)); err != nil {
		t.Fatalf("expired durable proposal blocked admission: %v", err)
	}
}

func TestAgencyClaimIgnoresCorruptTerminalArchive(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	runtime, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "agency evidence", now), []float32{1, 0, 0, 0}, "event-digest"); err != nil {
		t.Fatal(err)
	}
	first := durableAgencyRecord(t, now, "proposal-terminal", .9)
	second := durableAgencyRecord(t, now, "proposal-active", .8)
	if _, err := runtime.PutAgencyProposal(ctx, first, "digest-terminal", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.PutAgencyProposal(ctx, second, "digest-active", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	claimed, _, err := runtime.ClaimAgencyProposals(ctx, "tenant-a", "authority-a", now, 1, time.Minute)
	if err != nil || len(claimed) != 1 || claimed[0].Proposal.ID != first.Proposal.ID {
		t.Fatalf("first claim = %+v, %v", claimed, err)
	}
	if _, err := runtime.ResolveAgencyProposal(ctx, model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: first.Proposal.ID, ConsumerID: "authority-a", Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-1"}, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}

	raw, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	archive, err := raw.GetCollection("_eventframe_agency")
	if err != nil {
		t.Fatal(err)
	}
	archived, err := archive.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, stored := range archived {
		if stored.Metadata["proposal_id"] == first.Proposal.ID {
			metadata := cloneTestMetadata(stored.Metadata)
			metadata["agency_json"] = "{corrupt"
			if err := archive.Upsert(ctx, stored.ID, nil, metadata); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	remaining, _, err := reopened.ClaimAgencyProposals(ctx, "tenant-a", "authority-b", now.Add(2*time.Second), 1, time.Minute)
	if err != nil || len(remaining) != 1 || remaining[0].Proposal.ID != second.Proposal.ID {
		t.Fatalf("active claim consulted corrupt archive: %+v, %v", remaining, err)
	}
}

func TestContractSixRestartRebuildsActiveAgencyProjection(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/events.libravdb"
	now := time.Now().UTC()
	runtime, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "agency evidence", now), []float32{1, 0, 0, 0}, "event-digest"); err != nil {
		t.Fatal(err)
	}
	record := durableAgencyRecord(t, now, "proposal-migrate", .8)
	record.Proposal.ContractVersion = 6
	legacySigner, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	record.Signed, err = legacySigner.Sign(record.Proposal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.PutAgencyProposal(ctx, record, "digest-migrate", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}

	raw, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	names, err := raw.ListCollectionsWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if !strings.HasPrefix(name, "_eventframe_agency_active_") {
			continue
		}
		active, getErr := raw.GetCollection(name)
		if getErr != nil {
			t.Fatal(getErr)
		}
		records, listErr := active.ListAll(ctx)
		if listErr != nil {
			t.Fatal(listErr)
		}
		ids := make([]string, len(records))
		for index := range records {
			ids[index] = records[index].ID
		}
		if len(ids) > 0 {
			if err := active.DeleteBatch(ctx, ids); err != nil {
				t.Fatal(err)
			}
		}
	}
	system, err := raw.GetCollection("_eventframe_system")
	if err != nil {
		t.Fatal(err)
	}
	stateRecord, err := system.Get(ctx, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(stateRecord.Metadata["state_json"].(string)), &state); err != nil {
		t.Fatal(err)
	}
	state["snapshot"].(map[string]interface{})["contract_version"] = float64(6)
	encoded, _ := json.Marshal(state)
	metadata := cloneTestMetadata(stateRecord.Metadata)
	metadata["state_json"] = string(encoded)
	if err := system.Upsert(ctx, "runtime", nil, metadata); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	claimed, _, err := reopened.ClaimAgencyProposals(ctx, "tenant-a", "authority-a", now.Add(time.Second), 1, time.Minute)
	if err != nil || len(claimed) != 1 || claimed[0].Proposal.ID != record.Proposal.ID {
		t.Fatalf("contract-6 projection rebuild claim = %+v, %v", claimed, err)
	}
}

func durableAgencyRecord(t *testing.T, now time.Time, id string, priority float64) model.AgencyProposalRecord {
	t.Helper()
	proposal, err := agency.BuildProposal(model.AgencyProposalDraft{
		ID: id, TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
		Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: priority,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: id, CausalChainID: "chain-" + id,
	}, now, agency.DefaultPolicy(true))
	if err != nil {
		t.Fatal(err)
	}
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(proposal)
	if err != nil {
		t.Fatal(err)
	}
	return model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
}

func cloneTestMetadata(source map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}

func TestBayesianPolicyDigestSurvivesRestartAndAdvancesVersionOnChange(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	bound, err := first.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil {
		t.Fatal(err)
	}
	if bound.PolicyVersion != 2 {
		t.Fatalf("first policy version = %d", bound.PolicyVersion)
	}
	unchanged, err := first.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil || unchanged != bound {
		t.Fatalf("same policy bind = %+v, %v", unchanged, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	reopened, err := second.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil || reopened != bound {
		t.Fatalf("reopened policy bind = %+v, %v", reopened, err)
	}
	changed, err := second.BindBayesianPolicy(context.Background(), "policy-b")
	if err != nil {
		t.Fatal(err)
	}
	if changed.PolicyVersion != bound.PolicyVersion+1 || changed.RuntimeVersion != bound.RuntimeVersion+1 {
		t.Fatalf("changed policy snapshot = %+v", changed)
	}
}

func TestDeleteRetentionCompactAndBackup(t *testing.T) {
	root := t.TempDir()
	database, err := libravdbstore.Open(testConfig(root+"/events.libravdb", "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	items := []struct {
		value  model.Event
		digest string
	}{
		{testutil.Event("old", "old", now.Add(-2*time.Hour)), "old"},
		{testutil.Event("recent", "recent", now.Add(-time.Minute)), "recent"},
	}
	for _, item := range items {
		if _, err := database.Put(context.Background(), item.value, []float32{1, 0, 0, 0}, item.digest); err != nil {
			t.Fatal(err)
		}
	}
	backup := root + "/backup.libravdb"
	if err := database.Backup(context.Background(), backup); err != nil {
		t.Fatal(err)
	}
	deleted, err := database.Delete(context.Background(), "tenant-a", "recent")
	if err != nil || !deleted.Deleted {
		t.Fatalf("delete = %+v, %v", deleted, err)
	}
	if deleted.Snapshot.GraphVersion != 2 || deleted.Snapshot.PosteriorVersion != 2 {
		t.Fatalf("delete snapshot = %+v", deleted.Snapshot)
	}
	retained, err := database.DeleteBefore(context.Background(), "tenant-a", now.Add(-time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(retained.DeletedIDs) != 1 || retained.DeletedIDs[0] != "old" {
		t.Fatalf("retention = %+v", retained)
	}
	if err := database.Compact(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := libravdbstore.Open(testConfig(backup, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	results, err := recovered.Search(context.Background(), "tenant-a", []float32{1, 0, 0, 0}, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("backup recovered %d events", len(results))
	}
}
