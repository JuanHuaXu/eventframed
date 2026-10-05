package libravdbstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// applyOutcomeV17 is a research-only publication transition. Its caller must
// serialize event writes and read-to-journal Recalls around this operation.
func (g *incrementalSortGate) applyOutcomeV17(ctx context.Context, request model.BayesianOutcomeRequest,
	posteriorKey, parentPosteriorKey, digest string, weight float64,
	changePolicy bayes.ChangePolicy, groupPolicy bayes.GroupPolicy,
	observation model.ResidualObservation, residualPolicy residual.Policy,
	stopAt string,
) (store.BayesianOutcomeResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.count != g.wantRows || marker.lsn != g.verifiedLSN {
		return store.BayesianOutcomeResult{}, errors.New("outcome publication gate is not current")
	}
	beforeLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || beforeLSN != marker.lsn {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("outcome publication LSN moved: latest=%d marker=%d: %w", beforeLSN, marker.lsn, err)
	}
	beforeSnapshot := g.store.Snapshot(ctx)
	result, err := g.store.ApplyBayesianOutcome(ctx, request, posteriorKey, parentPosteriorKey, digest,
		weight, changePolicy, groupPolicy, observation, residualPolicy)
	if err != nil {
		// A failed write is safe to retry only if it left durable state unchanged.
		afterLSN, lsnErr := g.store.db.LatestCommitLSN(ctx)
		if lsnErr != nil || afterLSN != beforeLSN {
			g.verified = ""
		}
		return store.BayesianOutcomeResult{}, err
	}
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || result.Snapshot != g.store.Snapshot(ctx) {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("outcome snapshot or LSN readback failed: %w", err)
	}
	if result.Duplicate {
		if afterLSN != beforeLSN || result.Snapshot != beforeSnapshot {
			g.verified = ""
			return store.BayesianOutcomeResult{}, errors.New("duplicate outcome moved published state")
		}
		return result, nil
	}
	if afterLSN <= beforeLSN || result.Snapshot.RuntimeVersion <= beforeSnapshot.RuntimeVersion ||
		result.Snapshot.PosteriorVersion <= beforeSnapshot.PosteriorVersion ||
		result.Snapshot.ResidualVersion <= beforeSnapshot.ResidualVersion {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("new outcome moved LSN %d -> %d or did not advance runtime/posterior/residual versions: before=%+v after=%+v",
			beforeLSN, afterLSN, beforeSnapshot, result.Snapshot)
	}
	record, err := g.store.bayesian.Get(ctx, bayesianOutcomeRecordID(request.TenantID, request.IdempotencyKey))
	if err != nil {
		g.verified = ""
		return store.BayesianOutcomeResult{}, err
	}
	var saved model.BayesianOutcomeRequest
	encoded, _ := record.Metadata["outcome_json"].(string)
	if err := json.Unmarshal([]byte(encoded), &saved); err != nil || !reflect.DeepEqual(saved, request) ||
		record.Metadata["content_digest"] != digest || record.Metadata["posterior_key"] != posteriorKey {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("outcome readback mismatch: %v", err)
	}
	posterior, err := g.store.GetBayesianPosterior(ctx, request.TenantID, result.Posterior.PosteriorKey)
	if err != nil || !reflect.DeepEqual(posterior, result.Posterior) {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("posterior readback mismatch: %v", err)
	}
	if stopAt == "after_db" {
		g.verified = ""
		return store.BayesianOutcomeResult{}, errors.New("injected interruption after outcome DB commit")
	}
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return store.BayesianOutcomeResult{}, err
	}
	defer tx.Rollback()
	updated, err := tx.ExecContext(ctx,
		"UPDATE marker SET lsn=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		afterLSN, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return store.BayesianOutcomeResult{}, err
	}
	if n, err := updated.RowsAffected(); err != nil || n != 1 {
		g.verified = ""
		return store.BayesianOutcomeResult{}, fmt.Errorf("outcome marker update affected %d rows: %v", n, err)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return store.BayesianOutcomeResult{}, err
	}
	if stopAt == "after_sqlite" {
		g.verified = ""
		return store.BayesianOutcomeResult{}, errors.New("injected interruption after outcome marker commit")
	}
	g.verifiedLSN = afterLSN
	return result, nil
}

func researchOutcomeV17(asOf time.Time, id string) (model.BayesianOutcomeRequest, model.ResidualObservation) {
	available := asOf.Add(time.Nanosecond)
	request := model.BayesianOutcomeRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, TenantID: "tenant-a",
		JournalID: "outcome-journal-v17", EventID: "past100", Useful: true,
		ObservedAt: available, AvailableAt: available, Source: model.OutcomeFullStream,
		InclusionProbability: 1,
	}
	observation := model.ResidualObservation{
		ActionKey: "outcome-exact-v17", GeneralKey: "outcome-general-v17",
		HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: 0.5,
		CommittedProbability: 0.5, Useful: true, ValidationEligible: true,
		EventID: request.EventID, JournalID: request.JournalID,
		AvailableAt: available, PosteriorKey: request.EventID,
	}
	return request, observation
}

func openDenseOutcomeGateV17(root string) (*incrementalSortGate, error) {
	ctx := context.Background()
	s, err := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 256,
		EmbeddingModel: "research:d256", Quantization: "none", MemoryMapping: true})
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "sort-publication.sqlite"))
	if err != nil {
		s.Close()
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000"} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			s.Close()
			return nil, err
		}
	}
	g := &incrementalSortGate{&sortPublicationGate{store: s, sidecar: db,
		name: collectionName("tenant-a", "research:d256"), wantRows: 3}}
	marker, err := g.readMarker(ctx)
	if err != nil {
		g.close()
		return nil, err
	}
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		g.close()
		return nil, err
	}
	if marker.phase == "READY" && marker.lsn == lsn && marker.count >= 3 {
		rootHash, count, verifyErr := g.verifyJournal(ctx)
		if verifyErr == nil && count == marker.count && rootHash == marker.digest {
			g.wantRows, g.verified, g.verifiedLSN = count, rootHash, lsn
		}
	}
	return g, nil
}

func TestResearchPublishedOutcomeGateV17(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_OUTCOME_V17") != "1" {
		t.Skip("opt-in private outcome publication contract")
	}
	ctx := context.Background()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 120000000, time.UTC)
	change := bayes.ChangePolicy{Threshold: 2}
	policy := residual.Policy{Clip: 0.5}
	apply := func(g *incrementalSortGate, request model.BayesianOutcomeRequest, observation model.ResidualObservation, stopAt string) (store.BayesianOutcomeResult, error) {
		payload, err := json.Marshal(request)
		if err != nil {
			return store.BayesianOutcomeResult{}, err
		}
		contentHash := sha256.Sum256(payload)
		digest := hex.EncodeToString(contentHash[:])
		return g.applyOutcomeV17(ctx, request, request.EventID, "", digest, 1,
			change, bayes.GroupPolicy{}, observation, policy, stopAt)
	}
	t.Run("new_duplicate_conflict_reopen", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		defer func() {
			if g != nil {
				g.close()
			}
		}()
		request, observation := researchOutcomeV17(asOf, "outcome-v17-1")
		before, ready := g.capture(ctx)
		if !ready {
			t.Fatal("genesis gate not READY")
		}
		result, err := apply(g, request, observation, "")
		if err != nil || result.Duplicate || result.Posterior.Alpha != 2 {
			t.Fatal("new outcome", result, err)
		}
		after, ready := g.capture(ctx)
		if !ready || after <= before {
			t.Fatal("new outcome not published", before, after, ready)
		}
		duplicate, err := apply(g, request, observation, "")
		if err != nil || !duplicate.Duplicate {
			t.Fatal("exact duplicate", duplicate, err)
		}
		if same, ready := g.capture(ctx); !ready || same != after {
			t.Fatal("duplicate moved published LSN", after, same, ready)
		}
		conflict := request
		conflict.Useful = false
		_, err = apply(g, conflict, observation, "")
		if !errors.Is(err, store.ErrOutcomeConflict) {
			t.Fatal("conflicting outcome was not rejected", err)
		}
		if same, ready := g.capture(ctx); !ready || same != after {
			t.Fatal("conflict moved published LSN", after, same, ready)
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g = nil
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		if reopened, ready := g.capture(ctx); !ready || reopened != after {
			t.Fatal("reopen lost outcome publication", after, reopened, ready)
		}
	})
	for _, stopAt := range []string{"after_db", "after_sqlite"} {
		t.Run(stopAt, func(t *testing.T) {
			root := t.TempDir()
			g := createIncrementalSortGate(t, root)
			defer func() {
				if g != nil {
					g.close()
				}
			}()
			request, observation := researchOutcomeV17(asOf, "outcome-interrupt-"+stopAt)
			if _, err := apply(g, request, observation, stopAt); err == nil {
				t.Fatal("outcome interruption was not reached")
			}
			if _, ready := g.capture(ctx); ready {
				t.Fatal("interrupted outcome remained READY in memory")
			}
			if err := g.close(); err != nil {
				t.Fatal(err)
			}
			g = nil
			var err error
			g, err = openIncrementalSortGate(root)
			if err != nil {
				t.Fatal(err)
			}
			_, ready := g.capture(ctx)
			if ready != (stopAt == "after_sqlite") {
				t.Fatal("reopen had wrong READY state", stopAt, ready)
			}
		})
	}
	t.Run("direct_bypass_cannot_be_laundered", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		request, observation := researchOutcomeV17(asOf, "outcome-direct-v17")
		payload, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		contentHash := sha256.Sum256(payload)
		if _, err := g.store.ApplyBayesianOutcome(ctx, request, request.EventID, "", hex.EncodeToString(contentHash[:]), 1,
			change, bayes.GroupPolicy{}, observation, policy); err != nil {
			t.Fatal("direct bypass control", err)
		}
		if _, ready := g.capture(ctx); ready {
			t.Fatal("direct outcome bypass appeared READY")
		}
		another, nextObservation := researchOutcomeV17(asOf, "outcome-after-bypass-v17")
		if _, err := apply(g, another, nextObservation, ""); err == nil {
			t.Fatal("outcome transition laundered direct bypass")
		}
	})
}

// Service Recalls hold admission through journal commit. This wrapper makes an
// outcome wait for that boundary before changing the published snapshot.
type publishedOutcomeStoreV17 struct {
	*publishedRecallLoadStoreV16
	admission sync.RWMutex
}

func (s *publishedOutcomeStoreV17) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest,
	posteriorKey, parentPosteriorKey, digest string, weight float64,
	changePolicy bayes.ChangePolicy, groupPolicy bayes.GroupPolicy,
	observation model.ResidualObservation, residualPolicy residual.Policy,
) (store.BayesianOutcomeResult, error) {
	s.admission.Lock()
	defer s.admission.Unlock()
	s.owner.Lock()
	defer s.owner.Unlock()
	result, err := s.gate.applyOutcomeV17(ctx, request, posteriorKey, parentPosteriorKey,
		digest, weight, changePolicy, groupPolicy, observation, residualPolicy, "")
	if err != nil {
		s.current.Store(nil)
		return store.BayesianOutcomeResult{}, err
	}
	if err := s.publishLocked(ctx); err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	return result, nil
}

func TestResearchPublishedOutcomeServiceV17(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_OUTCOME_V17") != "1" {
		t.Skip("opt-in private outcome service contract")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	root := t.TempDir()
	gate := createUnpublishedDenseGateV8(t, root, query)
	defer func() {
		if gate != nil {
			gate.close()
		}
	}()
	base := &publishedRecallLoadStoreV16{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV16, 8), workerEnd: make(chan struct{})}
	adapter := &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	go base.runJournalWorker()
	defer func() {
		if svc != nil {
			svc.Close()
		}
	}()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 120000000, time.UTC)
	snapshot := gate.store.Snapshot(ctx)
	now := time.Now().UTC()
	selection := model.SelectionSupportCertificate{
		ID: "selection-v17", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: .2,
		SimultaneousCoverage: .95, Procedure: "test-only exhaustive frontier audit",
		Issuer: "research-fixture", ExternalAudit: true,
		ValidFrom: asOf.Add(-time.Minute), ValidUntil: now.Add(time.Hour),
	}
	if _, err := gate.store.PublishSelectionCertificate(ctx, selection); err != nil {
		t.Fatal(err)
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "omitted-v17", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: .02, DivergenceLimit: .05,
		AuditProbability: 1, SimultaneousCoverage: .95,
		Procedure: "test-only update-all shadow audit", Issuer: "research-fixture",
		ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	if _, err := gate.store.PublishOmittedInfluenceCertificate(ctx, omitted); err != nil {
		t.Fatal(err)
	}
	if err := gate.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := base.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	recall := func(at time.Time) (model.ContextPacket, *publishedRecallViewV8, error) {
		pin := &publishedRecallPinV8{}
		callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
		request := model.RecallRequest{
			ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "lab-v17",
			Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
			AsOf: at, RecallK: 50, PackK: 10, TokenBudget: 10000,
		}
		adapter.admission.RLock()
		packet, err := svc.Recall(callCtx, request)
		adapter.admission.RUnlock()
		return packet, pin.view, err
	}
	first, firstView, err := recall(asOf)
	if err != nil || firstView == nil || first.BayesianShadow.JournalID == "" {
		t.Fatal("first Recall journal", firstView, err)
	}
	if !first.BayesianShadow.SelectionSupportCertified || !first.BayesianShadow.OmittedInfluenceCertified {
		t.Fatal("test certificates did not activate", first.BayesianShadow.Mode)
	}
	var nominated bool
	for _, decision := range first.BayesianShadow.Decisions {
		if decision.EventID == "past100" && decision.Activated {
			nominated = true
		}
	}
	if !nominated {
		t.Fatal("feedback event was not activated in committed journal")
	}
	feedbackAt := asOf.Add(time.Nanosecond)
	outcome := model.BayesianOutcomeRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "service-outcome-v17",
		TenantID: "tenant-a", JournalID: first.BayesianShadow.JournalID, EventID: "past100",
		Useful: true, ObservedAt: feedbackAt, AvailableAt: feedbackAt,
		Source: model.OutcomeFullStream, InclusionProbability: 1,
	}
	response, err := svc.ObserveBayesianOutcome(ctx, outcome)
	if err != nil || response.Duplicate || response.Posterior.Mean() <= .5 {
		t.Fatal("service outcome publication", response, err)
	}
	if view := base.current.Load(); view == nil || view.snapshot != response.Snapshot || view.lsn <= firstView.lsn {
		t.Fatal("outcome did not publish returned snapshot", view)
	}
	if _, ready := gate.capture(ctx); !ready {
		t.Fatal("outcome lost READY publication")
	}
	duplicate, err := svc.ObserveBayesianOutcome(ctx, outcome)
	if err != nil || !duplicate.Duplicate || duplicate.Snapshot != response.Snapshot {
		t.Fatal("service duplicate", duplicate, err)
	}
	old, oldView, err := recall(asOf)
	if err != nil || oldView == nil {
		t.Fatal("old-as-of Recall", err)
	}
	for _, candidate := range old.Candidates {
		if candidate.Event.ID == "past100" && candidate.BayesianApplied {
			t.Fatal("future feedback leaked into old-as-of forecast")
		}
	}
	current, currentView, err := recall(feedbackAt.Add(time.Nanosecond))
	if err != nil || currentView == nil || currentView.snapshot != response.Snapshot {
		t.Fatal("post-feedback Recall snapshot", currentView, err)
	}
	var applied bool
	for _, candidate := range current.Candidates {
		if candidate.Event.ID == "past100" && candidate.BayesianApplied && candidate.BayesianProbability > .5 {
			applied = true
		}
	}
	if !applied {
		t.Fatal("durable feedback did not reach scored forecast")
	}
	if err := svc.Close(); err != nil {
		t.Fatal("close first service", err)
	}
	svc = nil
	if err := gate.close(); err != nil {
		t.Fatal("close first gate", err)
	}
	gate = nil
	gate, err = openDenseOutcomeGateV17(root)
	if err != nil {
		t.Fatal("reopen dense gate", err)
	}
	if _, ready := gate.capture(ctx); !ready {
		t.Fatal("reopened outcome marker is not READY")
	}
	reopenedBase := &publishedRecallLoadStoreV16{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV16, 8), workerEnd: make(chan struct{})}
	reopenedAdapter := &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: reopenedBase}
	svc, err = service.New(reopenedAdapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
	})
	if err != nil {
		t.Fatal("reopen service", err)
	}
	go reopenedBase.runJournalWorker()
	if err := reopenedBase.publishLocked(ctx); err != nil {
		t.Fatal("reopen published view", err)
	}
	reopenPin := &publishedRecallPinV8{}
	reopenCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, reopenPin)
	reopenedAdapter.admission.RLock()
	reopened, err := svc.Recall(reopenCtx, model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "lab-v17-reopen",
		Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
		AsOf: feedbackAt.Add(time.Nanosecond), RecallK: 50, PackK: 10, TokenBudget: 10000,
	})
	reopenedAdapter.admission.RUnlock()
	if err != nil || reopenPin.view == nil || reopenPin.view.snapshot != response.Snapshot {
		t.Fatal("reopened Recall view", reopenPin.view, err)
	}
	applied = false
	for _, candidate := range reopened.Candidates {
		if candidate.Event.ID == "past100" && candidate.BayesianApplied && candidate.BayesianProbability > .5 {
			applied = true
		}
	}
	if !applied {
		t.Fatal("reopened feedback did not reach scored forecast")
	}
}
