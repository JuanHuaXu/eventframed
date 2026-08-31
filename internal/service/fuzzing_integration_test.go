package service_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestFuzzSensitivityUsesAsOfContextWithoutStateMutation(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(64)
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, event := range []model.Event{
		testutil.Event("relevant", "Venus is hottest because its atmosphere traps heat", now.Add(-2*time.Minute)),
		testutil.Event("distractor", "The Arctic centers on an ocean", now.Add(-time.Minute)),
	} {
		if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := memory.Snapshot(ctx)
	request := model.FuzzSensitivityRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", Query: "Which planet is hottest?", AsOf: now,
		BaseSnapshot: snapshot, EventIDs: []string{"relevant", "distractor"}, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .9, MinTrials: 1,
		Perturbations: []model.FuzzPerturbation{{
			ID: "relocation", PropertyID: "context-envelope", EventID: "relevant", ValidityRuleID: "public-relocation-v1",
			ValidationKind: model.FuzzValidationDeclaredRelocation,
			Replacements: map[model.FuzzField]model.Field{
				model.FuzzWho: {Value: "another public evaluator", Source: model.SourceSynthetic, Confidence: 1, Evidence: "declared context relocation"},
			},
		}},
	}
	response, err := runtime.FuzzSensitivity(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Trials) != 1 || len(response.BaselineLaw) != 2 || response.Snapshot != snapshot || memory.Snapshot(ctx) != snapshot || response.CausalClaim {
		t.Fatalf("fuzz response/state = %+v; current=%+v", response, memory.Snapshot(ctx))
	}
	request.BaseSnapshot.RuntimeVersion--
	if _, err := runtime.FuzzSensitivity(ctx, request); !errors.Is(err, store.ErrStaleSnapshot) {
		t.Fatalf("stale audit error = %v", err)
	}
}

func TestFuzzSensitivityRejectsConcurrentSnapshotMotion(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	base, _ := embed.NewHashEmbedder(32)
	gated := &oneShotGateEmbedder{base: base, entered: make(chan struct{}), release: make(chan struct{})}
	runtime, err := service.New(memory, gated, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, event := range []model.Event{
		testutil.Event("one", "first public fact", now.Add(-2*time.Minute)),
		testutil.Event("two", "second public fact", now.Add(-time.Minute)),
	} {
		if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	gated.armed.Store(true)
	request := model.FuzzSensitivityRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", Query: "first fact", AsOf: now,
		BaseSnapshot: memory.Snapshot(ctx), EventIDs: []string{"one", "two"}, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .9, MinTrials: 1,
		Perturbations: []model.FuzzPerturbation{{
			ID: "relocation", PropertyID: "context", EventID: "one", ValidityRuleID: "declared-relocation",
			ValidationKind: model.FuzzValidationDeclaredRelocation,
			Replacements: map[model.FuzzField]model.Field{
				model.FuzzWho: {Value: "other participant", Source: model.SourceSynthetic, Confidence: 1, Evidence: "declared relocation"},
			},
		}},
	}
	result := make(chan error, 1)
	go func() {
		_, err := runtime.FuzzSensitivity(ctx, request)
		result <- err
	}()
	<-gated.entered
	late := testutil.Event("late", "concurrent fact", now.Add(-30*time.Second))
	if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: late.ID, Event: late}); err != nil {
		t.Fatal(err)
	}
	close(gated.release)
	if err := <-result; !errors.Is(err, store.ErrStaleSnapshot) {
		t.Fatalf("concurrent fuzz error = %v", err)
	}
}

func TestAuditChainTranslationIsSnapshotBoundAndReadOnly(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(64)
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	values := map[string]string{
		"a0-0": "low input", "a0-1": "low outcome", "a1-0": "high input", "a1-1": "low outcome",
		"b0-0": "cold input", "b0-1": "cold outcome", "b1-0": "hot input", "b1-1": "cold outcome",
	}
	for id, value := range values {
		availableAt := now.Add(-2 * time.Minute)
		if strings.HasSuffix(id, "-1") {
			availableAt = availableAt.Add(time.Minute)
		}
		event := testutil.Event(id, value, availableAt)
		if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := memory.Snapshot(ctx)
	request := model.ChainTranslationRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", Query: "outcome", AsOf: now, BaseSnapshot: snapshot,
		DomainA:            model.ChainTrajectory{BaselineEventIDs: []string{"a0-0", "a0-1"}, RevealedEventIDs: []string{"a1-0", "a1-1"}},
		DomainB:            model.ChainTrajectory{BaselineEventIDs: []string{"b0-0", "b0-1"}, RevealedEventIDs: []string{"b1-0", "b1-1"}},
		InvariantThreshold: 1, TranslationThreshold: 1,
		StageMaps: []model.ChainStageMap{
			{Stage: 0, DomainAField: model.FuzzWhat, DomainBField: model.FuzzWhat, DomainABefore: "low input", DomainAAfter: "high input", DomainBBefore: "cold input", DomainBAfter: "hot input", CorrespondenceID: "input-map", ValidityEvidence: "test map"},
			{Stage: 1, DomainAField: model.FuzzWhat, DomainBField: model.FuzzWhat, DomainABefore: "low outcome", DomainAAfter: "low outcome", DomainBBefore: "cold outcome", DomainBAfter: "cold outcome", CorrespondenceID: "outcome-map", ValidityEvidence: "test map"},
		},
	}
	response, err := runtime.AuditChainTranslation(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainHigherOrderInvariant || !response.PredictionEvaluated || response.Snapshot != snapshot || memory.Snapshot(ctx) != snapshot || response.CausalClaim || response.PublishesGraphOrGrouping {
		t.Fatalf("translation response/state = %+v; current=%+v", response, memory.Snapshot(ctx))
	}
	request.BaseSnapshot.RuntimeVersion--
	if _, err := runtime.AuditChainTranslation(ctx, request); !errors.Is(err, store.ErrStaleSnapshot) {
		t.Fatalf("stale audit error = %v", err)
	}
}

func TestFuzzSensitivityRejectsFutureAsOf(t *testing.T) {
	runtime := newMemoryService(t)
	request := model.FuzzSensitivityRequest{ProtocolVersion: model.ProtocolVersion, AsOf: time.Now().UTC().Add(time.Hour)}
	if _, err := runtime.FuzzSensitivity(context.Background(), request); err == nil {
		t.Fatal("future as_of was accepted")
	}
}

type oneShotGateEmbedder struct {
	base    embed.Embedder
	armed   atomic.Bool
	entered chan struct{}
	release chan struct{}
}

func (embedder *oneShotGateEmbedder) Embed(text string) ([]float32, error) {
	if embedder.armed.CompareAndSwap(true, false) {
		close(embedder.entered)
		<-embedder.release
	}
	return embedder.base.Embed(text)
}

func (embedder *oneShotGateEmbedder) Dimension() int   { return embedder.base.Dimension() }
func (embedder *oneShotGateEmbedder) Name() string     { return embedder.base.Name() }
func (embedder *oneShotGateEmbedder) ModelKey() string { return embedder.base.ModelKey() }
