package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestLowCertaintyRecallEnqueuesAndExecutesBackgroundFuzz(t *testing.T) {
	runtime, now, modelKey := backgroundFuzzService(t, service.BackgroundFuzzPolicy{
		Enabled: true, AnswerCertaintyThreshold: .5, QueueCapacity: 4,
		WorkerInterval: 2 * time.Millisecond, JobTimeout: time.Second, Cooldown: time.Hour,
		MaxEvents: 3, MaxPerturbations: 3, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .90, MinTrials: 1,
	})

	packet, err := runtime.Recall(context.Background(), backgroundFuzzRecall(now, modelKey, 2))
	if err != nil {
		t.Fatal(err)
	}
	if packet.PacketAnswerCertainty != 0 {
		t.Fatalf("answer certainty = %v, want 0", packet.PacketAnswerCertainty)
	}
	status := waitForBackgroundFuzz(t, runtime, func(status model.BackgroundFuzzQueueStatus) bool {
		return status.CompletedTotal == 1
	})
	if status.EnqueuedTotal != 1 || status.FailedTotal != 0 || status.StaleTotal != 0 || status.LastResult == nil {
		t.Fatalf("background fuzz status = %+v", status)
	}
	if status.LastResult.Status != "completed" || status.LastResult.TrialCount == 0 || status.LastResult.EventCount != 3 || status.LastResult.TriggerReason != "low-packing-boundary-certainty" {
		t.Fatalf("background fuzz result = %+v", status.LastResult)
	}
	if status.LastResult.Error != "" {
		t.Fatalf("background fuzz leaked an error: %+v", status.LastResult)
	}
	encoded, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "fuzz-0") || strings.Contains(string(encoded), "public retrieval question") {
		t.Fatalf("background queue status leaked job inputs: %s", encoded)
	}

	if _, err := runtime.Recall(context.Background(), backgroundFuzzRecall(now, modelKey, 2)); err != nil {
		t.Fatal(err)
	}
	status = waitForBackgroundFuzz(t, runtime, func(status model.BackgroundFuzzQueueStatus) bool {
		return status.DeduplicatedTotal == 1
	})
	if status.EnqueuedTotal != 1 || status.CompletedTotal != 1 {
		t.Fatalf("duplicate recall enqueued another job: %+v", status)
	}
}

func TestCertainRecallDoesNotEnqueueBackgroundFuzz(t *testing.T) {
	runtime, now, modelKey := backgroundFuzzService(t, service.BackgroundFuzzPolicy{
		Enabled: true, AnswerCertaintyThreshold: .5, QueueCapacity: 4,
		WorkerInterval: time.Millisecond, JobTimeout: time.Second, Cooldown: time.Hour,
		MaxEvents: 3, MaxPerturbations: 3, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .90, MinTrials: 1,
	})
	if _, err := runtime.Recall(context.Background(), backgroundFuzzRecall(now, modelKey, 3)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	status := runtime.BackgroundFuzzStatus()
	if status.EnqueuedTotal != 0 || status.CompletedTotal != 0 || status.QueueDepth != 0 {
		t.Fatalf("certain recall nominated fuzz work: %+v", status)
	}
}

func TestBackgroundFuzzDropsStaleSnapshotWithoutRetry(t *testing.T) {
	runtime, now, modelKey := backgroundFuzzService(t, service.BackgroundFuzzPolicy{
		Enabled: true, AnswerCertaintyThreshold: .5, QueueCapacity: 4,
		WorkerInterval: 50 * time.Millisecond, JobTimeout: time.Second, Cooldown: time.Hour,
		MaxEvents: 3, MaxPerturbations: 3, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .90, MinTrials: 1,
	})
	if _, err := runtime.Recall(context.Background(), backgroundFuzzRecall(now, modelKey, 2)); err != nil {
		t.Fatal(err)
	}
	newEvidence := testutil.Event("later-evidence", "new evidence invalidates the captured audit snapshot", now)
	if _, err := runtime.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: newEvidence.ID, Event: newEvidence}); err != nil {
		t.Fatal(err)
	}
	status := waitForBackgroundFuzz(t, runtime, func(status model.BackgroundFuzzQueueStatus) bool {
		return status.StaleTotal == 1
	})
	if status.CompletedTotal != 0 || status.FailedTotal != 0 || status.LastResult == nil || status.LastResult.Status != "stale" {
		t.Fatalf("stale background fuzz was not rejected cleanly: %+v", status)
	}
	if status.LastResult.Error != "snapshot changed before audit" {
		t.Fatalf("stale background fuzz exposed an unbounded error: %+v", status.LastResult)
	}
}

func TestBackgroundFuzzQueueAppliesNonblockingBackpressure(t *testing.T) {
	runtime, now, modelKey := backgroundFuzzService(t, service.BackgroundFuzzPolicy{
		Enabled: true, AnswerCertaintyThreshold: .5, QueueCapacity: 1,
		WorkerInterval: time.Hour, JobTimeout: time.Second, Cooldown: time.Hour,
		MaxEvents: 3, MaxPerturbations: 3, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .90, MinTrials: 1,
	})
	first := backgroundFuzzRecall(now, modelKey, 2)
	if _, err := runtime.Recall(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	second := first
	second.Query = "a distinct unresolved public question"
	if _, err := runtime.Recall(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	status := runtime.BackgroundFuzzStatus()
	if status.EnqueuedTotal != 1 || status.QueueDepth != 1 || status.DroppedTotal != 1 {
		t.Fatalf("bounded queue status = %+v", status)
	}
}

func backgroundFuzzService(t *testing.T, policy service.BackgroundFuzzPolicy) (*service.Service, time.Time, string) {
	t.Helper()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000,
		BackgroundFuzz: policy,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	now := time.Now().UTC().Truncate(time.Millisecond)
	vector := []float32{1, 0, 0, 0, 0, 0, 0, 0}
	for index := range 3 {
		event := testutil.Event(fmt.Sprintf("fuzz-%d", index), "shared retrieval surface", now.Add(-time.Minute))
		event.What.Value = fmt.Sprintf("public fact %d", index)
		event.Why.Value = fmt.Sprintf("public reason %d", index)
		event.How.Value = fmt.Sprintf("public method %d", index)
		event.Embedding = append([]float32(nil), vector...)
		event.EmbeddingModel = embedder.ModelKey()
		if _, err := runtime.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	return runtime, now, embedder.ModelKey()
}

func backgroundFuzzRecall(now time.Time, modelKey string, packK int) model.RecallRequest {
	return model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "public retrieval question", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: modelKey,
		AsOf: now, RecallK: 3, PackK: packK, TokenBudget: 1000,
	}
}

func waitForBackgroundFuzz(t *testing.T, runtime *service.Service, done func(model.BackgroundFuzzQueueStatus) bool) model.BackgroundFuzzQueueStatus {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status := runtime.BackgroundFuzzStatus()
		if done(status) {
			return status
		}
		time.Sleep(time.Millisecond)
	}
	status := runtime.BackgroundFuzzStatus()
	t.Fatalf("timed out waiting for background fuzz: %+v", status)
	return model.BackgroundFuzzQueueStatus{}
}
