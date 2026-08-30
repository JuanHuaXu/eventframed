package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestObserveIsIdempotentAndRejectsConflicts(t *testing.T) {
	runtime := newMemoryService(t)
	event := testutil.Event("stable-id", "first", time.Now().UTC())
	request := model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}

	first, err := runtime.Observe(context.Background(), request)
	if err != nil || first.Duplicate {
		t.Fatalf("first observe = %+v, %v", first, err)
	}
	second, err := runtime.Observe(context.Background(), request)
	if err != nil || !second.Duplicate {
		t.Fatalf("duplicate observe = %+v, %v", second, err)
	}
	request.Event.Content = "changed"
	if _, err := runtime.Observe(context.Background(), request); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatalf("conflicting observe error = %v", err)
	}
}

func TestObserveRejectsSameIDWithDifferentExplicitVector(t *testing.T) {
	runtime := newMemoryService(t)
	event := testutil.Event("vector-id", "same", time.Now().UTC())
	event.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
	event.EmbeddingModel = "feature-hash-v1:d8"
	request := model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}
	if _, err := runtime.Observe(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	request.Event.Embedding = []float32{0, 1, 0, 0, 0, 0, 0, 0}
	if _, err := runtime.Observe(context.Background(), request); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatalf("error = %v", err)
	}
}

func TestRecallExcludesUnavailableEventsBeforeCandidateLimit(t *testing.T) {
	runtime := newMemoryService(t)
	now := time.Now().UTC()
	for index := 0; index < 30; index++ {
		event := testutil.Event(fmt.Sprintf("future-%02d", index), "same query", now.Add(time.Hour))
		observe(t, runtime, event)
	}
	past := testutil.Event("eligible", "same query", now.Add(-time.Minute))
	observe(t, runtime, past)

	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "same query", AsOf: now, RecallK: 3, PackK: 3, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != "eligible" {
		t.Fatalf("candidates = %+v", packet.Candidates)
	}
}

func TestRerankingRunsBeforeIndependentPackingCap(t *testing.T) {
	now := time.Now().UTC()
	results := make([]store.SearchResult, 20)
	for index := range results {
		event := testutil.Event(fmt.Sprintf("ordinary-%02d", index), "ordinary", now.Add(-time.Minute))
		event.Priority = 0
		results[index] = store.SearchResult{Event: event, Similarity: 0.5}
	}
	answer := testutil.Event("answer", "the needed answer", now.Add(-time.Minute))
	answer.Priority = 1
	results[15] = store.SearchResult{Event: answer, Similarity: 0.5}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: results}, embedder, service.Config{
		DefaultRecallK: 20, DefaultPackK: 3, DefaultTokenBudget: 100, OverfetchMultiplier: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "other-session",
		Query: "query", AsOf: now, RecallK: 20, PackK: 3, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if packet.Candidates[0].Event.ID != "answer" {
		t.Fatalf("answer was dropped before reranking: %+v", packet.Candidates)
	}
	if packet.BayesianShadow.JournalID == "" || !packet.BayesianShadow.JournalDurable {
		t.Fatalf("Bayesian decision was not durably journaled: %+v", packet.BayesianShadow)
	}
}

func TestCalibrationChangesForecastLawWithoutChangingRetrievalScore(t *testing.T) {
	now := time.Now().UTC()
	event := testutil.Event("candidate", "candidate", now.Add(-time.Minute))
	results := []store.SearchResult{{Event: event, Similarity: 0.6}}
	embedder, _ := embed.NewHashEmbedder(8)
	calibrator := calibration.Logit{Scale: 2, Bias: -1, Floor: 1e-6}
	runtime, err := service.New(&fixedStore{results: results}, embedder, service.Config{
		DefaultRecallK: 1, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		BaselineCalibration: calibrator,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "query", AsOf: now, RecallK: 1, PackK: 1, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	candidate := packet.Candidates[0]
	if candidate.Score != candidate.BaselineScore {
		t.Fatalf("calibration changed retrieval score: score=%f baseline=%f", candidate.Score, candidate.BaselineScore)
	}
	wantProbability := calibrator.Apply(candidate.BaselineScore)
	if math.Abs(candidate.Forecast.BaseLaw.Useful-wantProbability) > 1e-12 {
		t.Fatalf("base probability=%f, want %f", candidate.Forecast.BaseLaw.Useful, wantProbability)
	}
	if candidate.Forecast.BaseLaw.Useful == candidate.BaselineScore {
		t.Fatal("nonidentity calibrator did not change forecast probability")
	}
}

func TestOptionalRankerFailurePreservesFallbackBaseScore(t *testing.T) {
	now := time.Now().UTC()
	event := testutil.Event("candidate", "candidate", now.Add(-time.Minute))
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: []store.SearchResult{{Event: event, Similarity: .6}}}, embedder, service.Config{
		DefaultRecallK: 1, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRanker: failingRanker{},
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "query", AsOf: now, RecallK: 1, PackK: 1, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 1 || packet.Candidates[0].RetrievalScore <= 0 || packet.Candidates[0].Score != packet.Candidates[0].RetrievalScore {
		t.Fatalf("optional ranker failure destroyed fallback score: %+v", packet.Candidates)
	}
}

func TestLibraVDBContractRerankingPromotesBackendChoiceWithoutChangingLaw(t *testing.T) {
	now := time.Now().UTC()
	ordinary := testutil.Event("ordinary", "general discussion", now.Add(-time.Minute))
	exact := testutil.Event("exact", "fix RecallK in service.go", now.Add(-time.Minute))
	results := []store.SearchResult{{Event: ordinary, Similarity: .6}, {Event: exact, Similarity: .6}}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: results}, embedder, service.Config{
		DefaultRecallK: 2, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRanker: exactFirstRanker{}, CandidateRankerRequired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "RecallK service.go", AsOf: now, RecallK: 2, PackK: 1, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != "exact" {
		t.Fatalf("contract rerank packet = %+v", packet.Candidates)
	}
	if packet.RetrievalContract != "test/RankCandidates" || packet.Candidates[0].RetrievalScore != .95 {
		t.Fatalf("retrieval contract metadata = %+v", packet)
	}
	ordinaryDecision := findDecision(t, packet.BayesianShadow, "ordinary")
	exactDecision := findDecision(t, packet.BayesianShadow, "exact")
	if ordinaryDecision.Forecast.PreResidualLaw != exactDecision.Forecast.PreResidualLaw {
		t.Fatalf("backend rank signal changed probability law: ordinary=%+v exact=%+v", ordinaryDecision.Forecast, exactDecision.Forecast)
	}
}

func TestPartialContractRankingPreservesNominatedFrontier(t *testing.T) {
	now := time.Now().UTC()
	first := testutil.Event("first", "first nominated event", now.Add(-time.Minute))
	second := testutil.Event("second", "second nominated event", now.Add(-2*time.Minute))
	third := testutil.Event("third", "third nominated event", now.Add(-3*time.Minute))
	metadata := []byte(`{"collection":"session:frontier"}`)
	retriever := &fixedRetriever{results: []retrieval.Candidate{
		{ID: first.ID, Score: .8, Metadata: metadata}, {ID: second.ID, Score: .6, Metadata: metadata}, {ID: third.ID, Score: .4, Metadata: metadata},
	}}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: []store.SearchResult{
		{Event: first, Similarity: .8}, {Event: second, Similarity: .6}, {Event: third, Similarity: .4},
	}}, embedder, service.Config{
		DefaultRecallK: 3, DefaultPackK: 3, DefaultTokenBudget: 1_000, OverfetchMultiplier: 1,
		CandidateRetriever: retriever, CandidateRetrieverRequired: true,
		CandidateRanker: partialRanker{}, CandidateRankerRequired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "nominated event", AsOf: now, RecallK: 3, PackK: 3, TokenBudget: 1_000,
		RetrievalCollections: []string{"session:frontier"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 3 {
		t.Fatalf("partial contract response dropped nominated candidates: %+v", packet.Candidates)
	}
	byID := make(map[string]model.Candidate, len(packet.Candidates))
	for _, candidate := range packet.Candidates {
		byID[candidate.Event.ID] = candidate
	}
	if byID["first"].RetrievalScore != .95 || byID["second"].RetrievalScore != .6 || byID["third"].RetrievalScore != .4 {
		t.Fatalf("partial contract scores = %+v", byID)
	}
}

func TestPartialContractRankingStillRejectsMalformedResponses(t *testing.T) {
	now := time.Now().UTC()
	first := testutil.Event("first", "first nominated event", now.Add(-time.Minute))
	second := testutil.Event("second", "second nominated event", now.Add(-2*time.Minute))
	for _, mode := range []string{"unknown", "duplicate", "oversized"} {
		t.Run(mode, func(t *testing.T) {
			embedder, _ := embed.NewHashEmbedder(8)
			runtime, err := service.New(&fixedStore{results: []store.SearchResult{{Event: first, Similarity: .8}, {Event: second, Similarity: .6}}}, embedder, service.Config{
				DefaultRecallK: 2, DefaultPackK: 2, DefaultTokenBudget: 1_000, OverfetchMultiplier: 1,
				CandidateRanker: malformedRanker{mode: mode}, CandidateRankerRequired: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = runtime.Recall(context.Background(), model.RecallRequest{
				ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
				Query: "nominated event", AsOf: now, RecallK: 2, PackK: 2, TokenBudget: 1_000,
			})
			if err == nil {
				t.Fatalf("%s rank response was accepted", mode)
			}
		})
	}
}

func TestContractNominationUsesSearchTextCollectionsAndResolvesStoredEvents(t *testing.T) {
	now := time.Now().UTC()
	ordinary := testutil.Event("ordinary", "general discussion", now.Add(-time.Minute))
	exact := testutil.Event("exact", "fix RecallK in service.go", now.Add(-time.Minute))
	contractMetadata := []byte(`{"collection":"user:tenant-a"}`)
	retriever := &fixedRetriever{results: []retrieval.Candidate{{ID: exact.ID, Score: .9, Metadata: contractMetadata}, {ID: ordinary.ID, Score: .1, Metadata: contractMetadata}}}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: []store.SearchResult{{Event: ordinary}, {Event: exact}}}, embedder, service.Config{
		DefaultRecallK: 2, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRetriever: retriever, CandidateRetrieverRequired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "RecallK service.go", AsOf: now,
		RecallK: 2, PackK: 1, TokenBudget: 100, RetrievalCollections: []string{"user:tenant-a"},
		RetrievalExcludeByCollection: map[string][]string{"user:tenant-a": {"future"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if packet.NominationContract != "test/SearchTextCollections" || len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != exact.ID {
		t.Fatalf("contract nomination packet = %+v", packet)
	}
	if len(retriever.request.Collections) != 1 || retriever.request.Collections[0] != "user:tenant-a" || len(retriever.request.ExcludeByCollection["user:tenant-a"]) != 1 {
		t.Fatalf("contract nomination request = %+v", retriever.request)
	}
}

func TestProductionContractsOwnIndexNominationAndDeleteLifecycle(t *testing.T) {
	now := time.Now().UTC()
	backend := memorystore.New()
	retriever := &fixedRetriever{}
	index := &recordingIndex{}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(backend, embedder, service.Config{
		DefaultRecallK: 2, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRetriever: retriever, CandidateRetrieverRequired: true,
		CandidateIndex: index, CandidateCollectionPrefix: "eventframe-",
	})
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("contract-owned", "contract owned retrieval", now.Add(-time.Minute))
	observe(t, runtime, event)
	if index.ensured.ID != event.ID || !strings.HasPrefix(index.collection, "eventframe-") || len(index.collection) != len("eventframe-")+24 {
		t.Fatalf("indexed candidate = %#v collection=%q", index.ensured, index.collection)
	}
	var metadata map[string]any
	if err := json.Unmarshal(index.ensured.Metadata, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["collection"] != index.collection || metadata["eventframe_digest"] == "" || index.identityKey != "eventframe_digest" {
		t.Fatalf("index metadata=%#v identity=%s/%s", metadata, index.identityKey, index.identityValue)
	}
	retriever.results = []retrieval.Candidate{{ID: event.ID, Score: .9, Metadata: index.ensured.Metadata}}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, SessionID: event.SessionID,
		Query: event.Content, AsOf: now, RecallK: 2, PackK: 1, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retriever.request.Collections) != 1 || retriever.request.Collections[0] != index.collection || packet.NominationContract != retriever.RetrievalContractName() {
		t.Fatalf("nomination request=%#v packet=%#v", retriever.request, packet)
	}
	if _, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, SessionID: event.SessionID,
		Query: event.Content, AsOf: now, RecallK: 2, PackK: 1, TokenBudget: 100,
		RetrievalCollections: []string{"eventframe-another-tenant"},
	}); err == nil || !strings.Contains(err.Error(), "cannot override") {
		t.Fatalf("cross-tenant collection error = %v", err)
	}
	if _, err := runtime.Delete(context.Background(), model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, EventID: event.ID}); err != nil {
		t.Fatal(err)
	}
	if index.deletedCollection != index.collection || index.deletedID != event.ID {
		t.Fatalf("delete = %q/%q", index.deletedCollection, index.deletedID)
	}
}

func TestProductionRetentionMirrorsDeletedIDs(t *testing.T) {
	now := time.Now().UTC()
	backend := memorystore.New()
	retriever := &fixedRetriever{}
	index := &recordingIndex{}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(backend, embedder, service.Config{
		DefaultRecallK: 2, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRetriever: retriever, CandidateIndex: index, CandidateCollectionPrefix: "eventframe-",
	})
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("retained-contract", "expired contract memory", now.Add(-2*time.Hour))
	observe(t, runtime, event)
	if _, err := runtime.Retain(context.Background(), model.RetentionRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, Before: now.Add(-time.Hour), Limit: 10,
	}); err != nil {
		t.Fatal(err)
	}
	if len(index.batchIDs) != 1 || index.batchIDs[0] != event.ID || index.batchCollection != index.collection {
		t.Fatalf("retention batch = %q %#v", index.batchCollection, index.batchIDs)
	}
}

func TestProductionNominationRepairsOnlyActuallyStaleExternalIDs(t *testing.T) {
	now := time.Now().UTC()
	backend := memorystore.New()
	retriever := &fixedRetriever{}
	index := &recordingIndex{}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(backend, embedder, service.Config{
		DefaultRecallK: 3, DefaultPackK: 1, DefaultTokenBudget: 100, OverfetchMultiplier: 1,
		CandidateRetriever: retriever, CandidateRetrieverRequired: true,
		CandidateIndex: index, CandidateCollectionPrefix: "eventframe-",
	})
	if err != nil {
		t.Fatal(err)
	}
	future := testutil.Event("future-contract", "future contract memory", now.Add(time.Hour))
	observe(t, runtime, future)
	metadata := append([]byte(nil), index.ensured.Metadata...)
	index.batchIDs = nil
	retriever.results = []retrieval.Candidate{
		{ID: "stale-contract", Score: .9, Metadata: metadata},
		{ID: future.ID, Score: .8, Metadata: metadata},
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: future.TenantID, SessionID: future.SessionID,
		Query: "contract memory", AsOf: now, RecallK: 3, PackK: 1, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 0 {
		t.Fatalf("unavailable nomination entered packet: %#v", packet.Candidates)
	}
	if len(index.batchIDs) != 1 || index.batchIDs[0] != "stale-contract" {
		t.Fatalf("repair deleted the wrong IDs: %#v", index.batchIDs)
	}
}

func TestPackedOnlyOutcomeIsNotLearningEvidence(t *testing.T) {
	runtime := newMemoryService(t)
	now := time.Now().UTC()
	_, err := runtime.ObserveBayesianOutcome(context.Background(), model.BayesianOutcomeRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "packed-only", TenantID: "tenant-a", JournalID: "journal", EventID: "event",
		ObservedAt: now, AvailableAt: now, Source: model.OutcomeFullStream, InclusionProbability: 1,
		Signals: model.OutcomeSignals{Packed: true},
	})
	if err == nil || !strings.Contains(err.Error(), "packed-only") {
		t.Fatalf("packed-only outcome error = %v", err)
	}
}

func TestDefaultPolicyActivatesCompleteBoundedFrontier(t *testing.T) {
	runtime := newMemoryService(t)
	now := time.Now().UTC()
	for index := 0; index < 120; index++ {
		event := testutil.Event(fmt.Sprintf("frontier-%02d", index), fmt.Sprintf("bounded memory %02d", index), now.Add(-time.Minute))
		observe(t, runtime, event)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "bounded memory", AsOf: now, RecallK: 120, PackK: 100, TokenBudget: 2_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if packet.BayesianShadow.Nominated != 120 || packet.BayesianShadow.Activated != 120 {
		t.Fatalf("default policy did not update the complete bounded frontier: %+v", packet.BayesianShadow)
	}
}

func newMemoryService(t *testing.T) *service.Service {
	t.Helper()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

func observe(t *testing.T, runtime *service.Service, event model.Event) {
	t.Helper()
	_, err := runtime.Observe(context.Background(), model.ObserveRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event,
	})
	if err != nil {
		t.Fatal(err)
	}
}

type fixedStore struct{ results []store.SearchResult }

type exactFirstRanker struct{}

type partialRanker struct{}

type malformedRanker struct{ mode string }

type failingRanker struct{}

func (failingRanker) ContractName() string { return "test/failing-ranker" }

func (failingRanker) RankCandidates(context.Context, retrieval.RankRequest) ([]retrieval.Candidate, error) {
	return nil, errors.New("ranker unavailable")
}

type fixedRetriever struct {
	request retrieval.SearchRequest
	results []retrieval.Candidate
}

type recordingIndex struct {
	collection        string
	ensured           retrieval.Candidate
	identityKey       string
	identityValue     string
	deletedCollection string
	deletedID         string
	batchCollection   string
	batchIDs          []string
}

func (i *recordingIndex) EnsureText(_ context.Context, collection string, candidate retrieval.Candidate, identityKey, identityValue string) error {
	i.collection, i.ensured, i.identityKey, i.identityValue = collection, candidate, identityKey, identityValue
	return nil
}

func (i *recordingIndex) DeleteText(_ context.Context, collection, id string) error {
	i.deletedCollection, i.deletedID = collection, id
	return nil
}

func (i *recordingIndex) DeleteTextBatch(_ context.Context, collection string, ids []string) error {
	i.batchCollection, i.batchIDs = collection, append([]string(nil), ids...)
	return nil
}

func (r *fixedRetriever) RetrievalContractName() string { return "test/SearchTextCollections" }
func (r *fixedRetriever) SearchTextCollections(_ context.Context, request retrieval.SearchRequest) ([]retrieval.Candidate, error) {
	r.request = request
	return append([]retrieval.Candidate(nil), r.results...), nil
}

func (exactFirstRanker) ContractName() string { return "test/RankCandidates" }

func (exactFirstRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	if request.K1 != 2 || request.K2 != 2 || len(request.Candidates) != 2 {
		return nil, fmt.Errorf("unexpected rank request: %+v", request)
	}
	byID := map[string]retrieval.Candidate{}
	for _, candidate := range request.Candidates {
		byID[candidate.ID] = candidate
	}
	exact, ordinary := byID["exact"], byID["ordinary"]
	exact.Score, ordinary.Score = .95, .25
	return []retrieval.Candidate{exact, ordinary}, nil
}

func (partialRanker) ContractName() string { return "test/partial-RankCandidates" }

func (partialRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	if len(request.Candidates) != 3 || request.K2 != 3 {
		return nil, fmt.Errorf("unexpected rank request: %+v", request)
	}
	selected := request.Candidates[0]
	selected.Score = .95
	return []retrieval.Candidate{selected}, nil
}

func (r malformedRanker) ContractName() string { return "test/malformed-RankCandidates" }

func (r malformedRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	switch r.mode {
	case "unknown":
		return []retrieval.Candidate{{ID: "unknown", Score: .9}}, nil
	case "duplicate":
		return []retrieval.Candidate{request.Candidates[0], request.Candidates[0]}, nil
	case "oversized":
		return append(append([]retrieval.Candidate(nil), request.Candidates...), request.Candidates[0]), nil
	default:
		return nil, fmt.Errorf("unknown malformed-ranker mode %q", r.mode)
	}
}

func (s *fixedStore) BindBayesianPolicy(context.Context, string) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}

func (s *fixedStore) Put(context.Context, model.Event, []float32, string) (store.PutResult, error) {
	return store.PutResult{}, nil
}
func (s *fixedStore) Delete(context.Context, string, string) (store.DeleteResult, error) {
	return store.DeleteResult{}, nil
}
func (s *fixedStore) DeleteBefore(context.Context, string, time.Time, int) (store.RetentionResult, error) {
	return store.RetentionResult{}, nil
}
func (s *fixedStore) Backup(context.Context, string) error { return nil }
func (s *fixedStore) Compact(context.Context) error        { return nil }
func (s *fixedStore) Search(_ context.Context, _ string, _ []float32, _ time.Time, limit int) ([]store.SearchResult, error) {
	return s.results[:min(limit, len(s.results))], nil
}
func (s *fixedStore) GetEvents(_ context.Context, _ string, eventIDs []string, availableBy time.Time) ([]model.Event, error) {
	byID := make(map[string]model.Event, len(s.results))
	for _, result := range s.results {
		byID[result.Event.ID] = result.Event
	}
	events := make([]model.Event, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		event, ok := byID[eventID]
		if !ok || event.AvailableAt.After(availableBy) {
			return nil, store.ErrEventNotFound
		}
		events = append(events, event)
	}
	return events, nil
}
func (s *fixedStore) PutBayesianJournal(context.Context, model.BayesianJournalEntry) error {
	return nil
}
func (s *fixedStore) GetBayesianJournal(context.Context, string, string) (model.BayesianJournalEntry, error) {
	return model.BayesianJournalEntry{}, store.ErrJournalNotFound
}
func (s *fixedStore) PublishSelectionCertificate(context.Context, model.SelectionSupportCertificate) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}
func (s *fixedStore) GetSelectionCertificate(context.Context, string) (model.SelectionSupportCertificate, error) {
	return model.SelectionSupportCertificate{}, store.ErrCertificateNotFound
}
func (s *fixedStore) PublishAntiPigeonCertificate(context.Context, model.AntiPigeonCertificate) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}
func (s *fixedStore) GetAntiPigeonCertificate(context.Context, string, []string) (model.AntiPigeonCertificate, error) {
	return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
}
func (s *fixedStore) PublishOmittedInfluenceCertificate(context.Context, model.OmittedInfluenceCertificate) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}
func (s *fixedStore) GetOmittedInfluenceCertificate(context.Context, string) (model.OmittedInfluenceCertificate, error) {
	return model.OmittedInfluenceCertificate{}, store.ErrCertificateNotFound
}
func (s *fixedStore) ApplyBayesianOutcome(context.Context, model.BayesianOutcomeRequest, string, string, string, float64, bayes.ChangePolicy, bayes.GroupPolicy, model.ResidualObservation, residual.Policy) (store.BayesianOutcomeResult, error) {
	return store.BayesianOutcomeResult{}, nil
}
func (s *fixedStore) GetBayesianPosterior(context.Context, string, string) (model.BayesianPosterior, error) {
	return model.BayesianPosterior{}, store.ErrPosteriorNotFound
}
func (s *fixedStore) GetResidualCandidates(context.Context, string, string, string) (model.ResidualCandidates, error) {
	return model.ResidualCandidates{}, nil
}
func (s *fixedStore) GetPredictiveGraph(context.Context, string) (model.PredictiveGraph, error) {
	return model.PredictiveGraph{}, nil
}
func (s *fixedStore) PublishPredictiveSnap(context.Context, model.PredictiveSnapRecord) (model.PredictiveGraph, model.Snapshot, error) {
	return model.PredictiveGraph{}, model.Snapshot{}, nil
}
func (s *fixedStore) RollbackPredictiveSnap(context.Context, string, string, string) (model.PredictiveGraph, model.Snapshot, error) {
	return model.PredictiveGraph{}, model.Snapshot{}, nil
}
func (s *fixedStore) PutAgencyProposal(context.Context, model.AgencyProposalRecord, string, int, int, time.Time) (store.AgencyPutResult, error) {
	return store.AgencyPutResult{}, nil
}
func (s *fixedStore) ClaimAgencyProposals(context.Context, string, string, time.Time, int, time.Duration) ([]model.AgencyProposalRecord, model.Snapshot, error) {
	return nil, model.Snapshot{}, nil
}
func (s *fixedStore) ResolveAgencyProposal(context.Context, model.ResolveAgencyProposalRequest, time.Time) (store.AgencyResolveResult, error) {
	return store.AgencyResolveResult{}, nil
}
func (s *fixedStore) Stats(context.Context) (store.Stats, error) {
	return store.Stats{Backend: "fixed"}, nil
}
func (s *fixedStore) Snapshot(context.Context) model.Snapshot { return model.Snapshot{} }
func (s *fixedStore) Close() error                            { return nil }
