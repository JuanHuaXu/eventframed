package service_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

// Paired complete internal requests against an isolated on-disk database.
// Both modes retain ordinary certificates, residual checks, and journals.
func BenchmarkEvidenceInternalRequests(b *testing.B) {
	for _, enabled := range []bool{false, true} {
		for _, workers := range []int{1, 4} {
			for _, mixed := range []bool{false, true} {
				b.Run(fmt.Sprintf("upgrade=%t/workers=%d/mixed=%t", enabled, workers, mixed), func(b *testing.B) {
					cfg, key := evidenceConfig(b, enabled)
					if !enabled {
						key = nil
					}
					db, err := libravdbstore.Open(libravdbstore.Config{Path: b.TempDir() + "/internal.libravdb", Dimension: 32, EmbeddingModel: "test-hash:d32", Quantization: "none", MemoryMapping: true})
					if err != nil {
						b.Fatal(err)
					}
					runtime := evidenceRuntime(b, db, cfg)
					defer runtime.Close()
					seedEvidence(b, runtime, 50)
					certifyEvidenceFixture(b, runtime, db)
					packet := evidenceRecall(b, runtime)
					warm := signedFeedback(b, packet, "warm", key)
					if _, err := runtime.ObserveBayesianOutcome(context.Background(), warm); err != nil {
						b.Fatal(err)
					}
					requests := make([]model.BayesianOutcomeRequest, b.N)
					for i := range requests {
						requests[i] = signedFeedback(b, packet, fmt.Sprintf("observation-%d", i), key)
					}
					latencies := make([]time.Duration, b.N)
					var cursor atomic.Int64
					var wg sync.WaitGroup
					b.ReportAllocs()
					b.ResetTimer()
					for worker := 0; worker < workers; worker++ {
						wg.Add(1)
						go func() {
							defer wg.Done()
							for {
								i := int(cursor.Add(1) - 1)
								if i >= b.N {
									return
								}
								start := time.Now()
								if mixed && i%4 == 0 {
									_, err := runtime.ObserveBayesianOutcome(context.Background(), requests[i])
									if err != nil {
										b.Error(err)
									}
								} else {
									_, err := runtime.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "benchmark", Query: "public evidence", AsOf: time.Now().UTC(), RecallK: 50, PackK: 10, TokenBudget: 2000})
									if err != nil {
										b.Error(err)
									}
								}
								latencies[i] = time.Since(start)
							}
						}()
					}
					wg.Wait()
					b.StopTimer()
					sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
					b.ReportMetric(float64(latencies[(len(latencies)-1)*50/100].Nanoseconds()), "p50-ns")
					b.ReportMetric(float64(latencies[(len(latencies)-1)*99/100].Nanoseconds()), "p99-ns")
					b.ReportMetric(float64(latencies[len(latencies)-1].Nanoseconds()), "max-ns")
				})
			}
		}
	}
}
