package benchmark_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

const (
	benchmarkDimension = 384
	benchmarkTenant    = "benchmark-tenant"
	benchmarkSession   = "benchmark-session"
)

var benchmarkPacket model.ContextPacket

func BenchmarkRecall(b *testing.B) {
	for _, backend := range []string{"memory", "libravdb-sq8"} {
		eventCounts := []int{100, 1_000}
		if backend == "memory" || os.Getenv("EVENTFRAME_BENCH_LARGE") == "1" {
			eventCounts = append(eventCounts, 10_000)
		}
		for _, events := range eventCounts {
			b.Run(fmt.Sprintf("backend=%s/events=%d", backend, events), func(b *testing.B) {
				runtime, closeRuntime := seededRuntime(b, backend, events, 0)
				defer closeRuntime()
				request := recallRequest(time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
				if _, err := runtime.Recall(context.Background(), request); err != nil {
					b.Fatalf("warm recall: %v", err)
				}

				b.ReportAllocs()
				b.ReportMetric(float64(events), "events")
				latency := newLatencyRecorder()
				b.ResetTimer()
				for b.Loop() {
					request.AsOf = request.AsOf.Add(time.Nanosecond)
					started := time.Now()
					packet, err := runtime.Recall(context.Background(), request)
					latency.Observe(time.Since(started))
					if err != nil {
						b.Fatal(err)
					}
					benchmarkPacket = packet
				}
				latency.Report(b)
			})
		}
	}
}

func BenchmarkConcurrentContractRecall(b *testing.B) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		b.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	ctx := context.Background()
	embedder, err := embed.NewHashEmbedder(benchmarkDimension)
	if err != nil {
		b.Fatal(err)
	}
	backend, err := libravdbstore.Open(libravdbstore.Config{
		Path: b.TempDir() + "/events.libravdb", Dimension: benchmarkDimension,
		Quantization: "sq8", MemoryMapping: true, EmbeddingModel: embedder.ModelKey(),
	})
	if err != nil {
		b.Fatal(err)
	}
	defer backend.Close()
	contracts, err := retrieval.OpenLibraVDBContracts(endpoint)
	if err != nil {
		b.Fatal(err)
	}
	defer contracts.Close()
	deltaStore, err := rankdelta.Open(b.TempDir()+"/rank-deltas.sqlite", 10_000)
	if err != nil {
		b.Fatal(err)
	}
	defer deltaStore.Close()
	tenant := fmt.Sprintf("contract-benchmark-%d", time.Now().UnixNano())
	runtime, err := service.New(backend, embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
		CandidateRanker: contracts, CandidateRankerRequired: true,
		CandidateRetriever: contracts, CandidateRetrieverRequired: true,
		CandidateIndex: contracts, CandidateCollectionPrefix: "eventframe-benchmark-",
		RankDeltaStore: deltaStore, RankDeltaStoreRequired: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer runtime.Close()
	asOf := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	for index := 0; index < 200; index++ {
		event := testutil.Event(fmt.Sprintf("contract-event-%06d", index), fmt.Sprintf("benchmark target memory item %06d", index), asOf.Add(-time.Duration(index+1)*time.Second))
		event.TenantID, event.SessionID = tenant, benchmarkSession
		if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
			b.Fatalf("seed event %d: %v", index, err)
		}
	}
	request := recallRequest(asOf)
	request.TenantID = tenant
	if _, err := runtime.Recall(ctx, request); err != nil {
		b.Fatalf("warm contract recall: %v", err)
	}

	const maxSamples = 100_000
	samples := make([]int64, maxSamples)
	var sequence atomic.Int64
	b.ReportAllocs()
	b.ReportMetric(200, "events")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			index := sequence.Add(1) - 1
			current := request
			current.AsOf = current.AsOf.Add(time.Duration(index+1) * time.Nanosecond)
			started := time.Now()
			packet, recallErr := runtime.Recall(ctx, current)
			if index < maxSamples {
				samples[index] = time.Since(started).Nanoseconds()
			}
			if recallErr != nil {
				b.Error(recallErr)
				return
			}
			_ = packet
		}
	})
	b.StopTimer()
	count := min(int(sequence.Load()), len(samples))
	if count > 0 {
		sort.Slice(samples[:count], func(i, j int) bool { return samples[i] < samples[j] })
		b.ReportMetric(float64(samples[percentileIndex(count, .50)]), "p50-ns/op")
		b.ReportMetric(float64(samples[percentileIndex(count, .95)]), "p95-ns/op")
		b.ReportMetric(float64(samples[percentileIndex(count, .99)]), "p99-ns/op")
	}
}

func BenchmarkAvailabilityExpansion(b *testing.B) {
	for _, futurePercent := range []int{0, 50, 95} {
		b.Run(fmt.Sprintf("future=%d%%", futurePercent), func(b *testing.B) {
			runtime, closeRuntime := seededRuntime(b, "libravdb-sq8", 1_000, futurePercent)
			defer closeRuntime()
			request := recallRequest(time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
			if _, err := runtime.Recall(context.Background(), request); err != nil {
				b.Fatalf("warm recall: %v", err)
			}

			b.ReportAllocs()
			b.ReportMetric(float64(futurePercent), "future_percent")
			latency := newLatencyRecorder()
			b.ResetTimer()
			for b.Loop() {
				request.AsOf = request.AsOf.Add(time.Nanosecond)
				started := time.Now()
				packet, err := runtime.Recall(context.Background(), request)
				latency.Observe(time.Since(started))
				if err != nil {
					b.Fatal(err)
				}
				benchmarkPacket = packet
			}
			latency.Report(b)
		})
	}
}

func BenchmarkAgencyClaim(b *testing.B) {
	activeCounts := []int{1, 100}
	if os.Getenv("EVENTFRAME_BENCH_LARGE") == "1" {
		activeCounts = append(activeCounts, 1_000)
	}
	for _, active := range activeCounts {
		b.Run(fmt.Sprintf("active=%d", active), func(b *testing.B) {
			ctx := context.Background()
			now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
			backend, err := libravdbstore.Open(libravdbstore.Config{
				Path: b.TempDir() + "/events.libravdb", Dimension: 4, Quantization: "none",
				MemoryMapping: true, EmbeddingModel: "benchmark:d4",
			})
			if err != nil {
				b.Fatal(err)
			}
			defer backend.Close()
			evidence := testutil.Event("agency-evidence", "agency benchmark evidence", now.Add(-time.Minute))
			evidence.TenantID = benchmarkTenant
			if _, err := backend.Put(ctx, evidence, []float32{1, 0, 0, 0}, "agency-evidence-digest"); err != nil {
				b.Fatal(err)
			}
			signer, err := agency.NewSignerForTest()
			if err != nil {
				b.Fatal(err)
			}
			for index := 0; index < active; index++ {
				id := fmt.Sprintf("proposal-%04d", index)
				proposal, buildErr := agency.BuildProposal(model.AgencyProposalDraft{
					ID: id, TenantID: benchmarkTenant, SessionID: "openclaw:" + benchmarkSession,
					Action: model.AgencyNotify, Reason: "Benchmark follow-up is ready.",
					EvidenceIDs: []string{evidence.ID}, ExpectedUtility: .8,
					Priority: float64(index%100) / 100, NotBefore: now, ExpiresAt: now.Add(24 * time.Hour),
					IdempotencyKey: id, CausalChainID: "chain-" + id,
				}, now, agency.DefaultPolicy(true))
				if buildErr != nil {
					b.Fatal(buildErr)
				}
				signed, signErr := signer.Sign(proposal)
				if signErr != nil {
					b.Fatal(signErr)
				}
				record := model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
				if _, putErr := backend.PutAgencyProposal(ctx, record, "digest-"+id, 8, 1_000, now); putErr != nil {
					b.Fatal(putErr)
				}
			}

			// A zero lease makes the same bounded active set eligible on every
			// iteration without rebuilding the fixture inside the timed region.
			if _, _, err := backend.ClaimAgencyProposals(ctx, benchmarkTenant, "benchmark-authority", now, 10, 0); err != nil {
				b.Fatalf("warm claim: %v", err)
			}
			b.ReportAllocs()
			b.ReportMetric(float64(active), "active_records")
			latency := newLatencyRecorder()
			b.ResetTimer()
			for b.Loop() {
				started := time.Now()
				if _, _, err := backend.ClaimAgencyProposals(ctx, benchmarkTenant, "benchmark-authority", now, 10, 0); err != nil {
					b.Fatal(err)
				}
				latency.Observe(time.Since(started))
			}
			latency.Report(b)
		})
	}
}

func seededRuntime(b *testing.B, backendName string, eventCount, futurePercent int) (*service.Service, func()) {
	b.Helper()
	ctx := context.Background()
	embedder, err := embed.NewHashEmbedder(benchmarkDimension)
	if err != nil {
		b.Fatal(err)
	}
	var backend store.EventStore
	switch backendName {
	case "memory":
		backend = memorystore.New()
	case "libravdb-sq8":
		backend, err = libravdbstore.Open(libravdbstore.Config{
			Path: b.TempDir() + "/events.libravdb", Dimension: benchmarkDimension,
			Quantization: "sq8", MemoryMapping: true, EmbeddingModel: embedder.ModelKey(),
		})
		if err != nil {
			b.Fatal(err)
		}
	default:
		b.Fatalf("unknown benchmark backend %q", backendName)
	}
	runtime, err := service.New(backend, embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	if err != nil {
		b.Fatal(err)
	}
	asOf := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	futureStart := eventCount - eventCount*futurePercent/100
	for index := 0; index < eventCount; index++ {
		availableAt := asOf.Add(-time.Duration(index+1) * time.Second)
		contentClass := "historical"
		if index >= futureStart {
			availableAt = asOf.Add(time.Duration(index-futureStart+1) * time.Second)
			contentClass = "benchmark target"
		}
		event := testutil.Event(fmt.Sprintf("event-%06d", index), fmt.Sprintf("%s memory item %06d", contentClass, index), availableAt)
		event.TenantID = benchmarkTenant
		event.SessionID = benchmarkSession
		if _, err := runtime.Observe(ctx, model.ObserveRequest{
			ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event,
		}); err != nil {
			b.Fatalf("seed event %d: %v", index, err)
		}
	}
	return runtime, func() { _ = backend.Close() }
}

func recallRequest(asOf time.Time) model.RecallRequest {
	return model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: benchmarkTenant, SessionID: benchmarkSession,
		Query: "benchmark target memory item", AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 2_000,
	}
}

const maxLatencySamples = 100_000

type latencyRecorder struct {
	samples []int64
	seen    int
}

func newLatencyRecorder() *latencyRecorder {
	return &latencyRecorder{samples: make([]int64, maxLatencySamples)}
}

func (recorder *latencyRecorder) Observe(duration time.Duration) {
	if recorder.seen < len(recorder.samples) {
		recorder.samples[recorder.seen] = duration.Nanoseconds()
	}
	recorder.seen++
}

func (recorder *latencyRecorder) Report(b *testing.B) {
	b.StopTimer()
	count := min(recorder.seen, len(recorder.samples))
	if count == 0 {
		return
	}
	samples := recorder.samples[:count]
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	b.ReportMetric(float64(samples[percentileIndex(count, .50)]), "p50-ns/op")
	b.ReportMetric(float64(samples[percentileIndex(count, .95)]), "p95-ns/op")
	b.ReportMetric(float64(samples[percentileIndex(count, .99)]), "p99-ns/op")
}

func percentileIndex(count int, percentile float64) int {
	index := int(float64(count-1) * percentile)
	return min(max(index, 0), count-1)
}
