package libravdbstore

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

// Diagnostic workload: pseudo-random vectors, not semantic embeddings. Each
// operation measures sixteen writes after untimed corpus construction.
func BenchmarkResearchIndexScaling(b *testing.B) {
	for _, n := range []int{200, 800, 1600} {
		for _, flat := range []bool{false, true} {
			for _, size := range []int{1, 16} {
				b.Run(fmt.Sprintf("n%d/flat%t/batch%d", n, flat, size), func(b *testing.B) {
					var searchNS int64
					hits := 0
					for trial := 0; trial < b.N; trial++ {
						b.StopTimer()
						ctx := context.Background()
						config := Config{Path: b.TempDir() + "/scaling.libravdb", Dimension: 32, Quantization: "none", MemoryMapping: true, EmbeddingModel: "research:d32"}
						s, err := Open(config)
						if err != nil {
							b.Fatal(err)
						}
						if flat {
							_, err = s.db.EnsureCollection(ctx, collectionName("tenant-a", config.EmbeddingModel), 32, libra.WithMetric(libra.CosineDistance), libra.WithFlat(), libra.WithMemoryMapping(true))
							if err != nil {
								b.Fatal(err)
							}
						}
						writes := make([]ResearchEventWrite, n+16)
						for i := range writes {
							writes[i] = batchWrite(fmt.Sprintf("index-%d", i))
							h := sha256.Sum256([]byte(writes[i].Event.ID))
							writes[i].Vector = make([]float32, 32)
							for j, v := range h {
								writes[i].Vector[j] = float32(v) - 127.5
							}
						}
						for i := 0; i < n; i += 16 {
							if _, err := s.PutResearchEventBatch(ctx, writes[i:min(i+16, n)]); err != nil {
								b.Fatal(err)
							}
						}
						before := s.Snapshot(ctx)
						b.StartTimer()
						for i := n; i < n+16; i += size {
							if size == 1 {
								w := writes[i]
								_, err = s.Put(ctx, w.Event, w.Vector, w.Digest)
							} else {
								_, err = s.PutResearchEventBatch(ctx, writes[i:i+size])
							}
							if err != nil {
								b.Fatal(err)
							}
						}
						b.StopTimer()
						after := s.Snapshot(ctx)
						if after.RuntimeVersion != before.RuntimeVersion+16 || after.EvidenceEpoch != before.EvidenceEpoch+16 {
							b.Fatal("lost write")
						}
						for _, w := range writes[n:] {
							start := time.Now()
							found, err := s.Search(ctx, "tenant-a", w.Vector, w.Event.AvailableAt.Add(time.Hour), 10)
							searchNS += time.Since(start).Nanoseconds()
							if err != nil {
								b.Fatal(err)
							}
							for _, r := range found {
								if r.Event.ID == w.Event.ID {
									hits++
									break
								}
							}
						}
						if err := s.Close(); err != nil {
							b.Fatal(err)
						}
						s, err = Open(config)
						if err != nil {
							b.Fatal(err)
						}
						if s.Snapshot(ctx) != after {
							b.Fatal("snapshot lost on reopen")
						}
						for _, w := range writes[n:] {
							events, err := s.GetEvents(ctx, "tenant-a", []string{w.Event.ID}, w.Event.AvailableAt.Add(time.Hour))
							if err != nil || len(events) != 1 || events[0].FrameText() != w.Event.FrameText() {
								b.Fatal("reopen payload mismatch", err)
							}
						}
						if err := s.Close(); err != nil {
							b.Fatal(err)
						}
					}
					b.ReportMetric(float64(searchNS)/float64(b.N*16), "search-ns/query")
					b.ReportMetric(float64(hits)/float64(b.N*16), "self-recall@10")
				})
			}
		}
	}
}
