package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func batchWrite(id string) ResearchEventWrite {
	return ResearchEventWrite{Event: testutil.Event(id, "Curiosity landed on Mars in 2012", time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)), Vector: []float32{1, 0, 0, 0}, Digest: id}
}

func batchStore(t testing.TB) (*Store, Config) {
	t.Helper()
	c := Config{Path: t.TempDir() + "/batch.libravdb", Dimension: 4, Quantization: "none", MemoryMapping: true, EmbeddingModel: "research:d4"}
	s, err := Open(c)
	if err != nil {
		t.Fatal(err)
	}
	return s, c
}

func TestResearchEventBatchParityReopen(t *testing.T) {
	ctx := context.Background()
	s, config := batchStore(t)
	control, _ := batchStore(t)
	defer control.Close()
	writes := []ResearchEventWrite{batchWrite("a"), batchWrite("b"), batchWrite("a"), batchWrite("c")}
	base := s.Snapshot(ctx)
	results, err := s.PutResearchEventBatch(ctx, writes)
	if err != nil {
		t.Fatal(err)
	}
	for i, w := range writes {
		r, err := control.Put(ctx, w.Event, w.Vector, w.Digest)
		if err != nil || r.Duplicate != results[i].Duplicate {
			t.Fatalf("duplicate parity: %v %v", r, err)
		}
		if results[i].Snapshot != results[0].Snapshot {
			t.Fatal("intermediate snapshot returned")
		}
	}
	if s.Snapshot(ctx) != control.Snapshot(ctx) {
		t.Fatal("snapshot parity")
	}
	for i, id := range []string{"a", "b", "c"} {
		if s.ingestMotion[base.RuntimeVersion+uint64(i)+1] != batchWrite(id).Event.AvailableAt {
			t.Fatal("missing motion")
		}
	}
	wantSnapshot := s.Snapshot(ctx)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Snapshot(ctx) != wantSnapshot {
		t.Fatal("snapshot not durable")
	}
	ids := []string{"a", "b", "c"}
	got, err := s.GetEventsWithVectors(ctx, "tenant-a", ids, time.Now().Add(365*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	want, err := control.GetEventsWithVectors(ctx, "tenant-a", ids, time.Now().Add(365*24*time.Hour))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("payload parity: %v", err)
	}
	results, err = s.PutResearchEventBatch(ctx, writes)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if !r.Duplicate || r.Snapshot != wantSnapshot {
			t.Fatal("retry changed state")
		}
	}
}

func TestResearchEventBatchRejectsBeforeWrites(t *testing.T) {
	for _, kind := range []string{"conflict", "existing-conflict", "dimension", "tenant", "cancel", "over-cap"} {
		t.Run(kind, func(t *testing.T) {
			s, _ := batchStore(t)
			defer s.Close()
			ctx := context.Background()
			writes := []ResearchEventWrite{batchWrite("a"), batchWrite("b")}
			switch kind {
			case "conflict":
				writes = append(writes, batchWrite("a"))
				writes[2].Digest = "different"
			case "existing-conflict":
				w := batchWrite("b")
				if _, err := s.Put(ctx, w.Event, w.Vector, "different"); err != nil {
					t.Fatal(err)
				}
			case "dimension":
				writes[1].Vector = nil
			case "tenant":
				writes[1].Event.TenantID = "tenant-b"
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "over-cap":
				for len(writes) < 17 {
					writes = append(writes, batchWrite(fmt.Sprint(len(writes))))
				}
			}
			before := s.Snapshot(context.Background())
			if r, err := s.PutResearchEventBatch(ctx, writes); err == nil || r != nil {
				t.Fatal("invalid batch accepted")
			}
			if s.Snapshot(context.Background()) != before {
				t.Fatal("rejected batch changed snapshot")
			}
			_, err := s.GetEvents(context.Background(), "tenant-a", []string{"a"}, time.Now().Add(365*24*time.Hour))
			if !errors.Is(err, store.ErrEventNotFound) {
				t.Fatalf("rejected batch leaked event: %v", err)
			}
		})
	}
}

func TestResearchEventBatchConcurrentPut(t *testing.T) {
	s, _ := batchStore(t)
	defer s.Close()
	ctx := context.Background()
	before := s.Snapshot(ctx)
	var group sync.WaitGroup
	errs := make(chan error, 16)
	for i := range 16 {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			w := batchWrite(fmt.Sprint(i))
			if i%2 == 0 {
				_, err := s.Put(ctx, w.Event, w.Vector, w.Digest)
				errs <- err
			} else {
				_, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{w, w})
				errs <- err
			}
		}(i)
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	after := s.Snapshot(ctx)
	if after.RuntimeVersion != before.RuntimeVersion+16 || after.EvidenceEpoch != before.EvidenceEpoch+16 {
		t.Fatal("lost update")
	}
}

// Each measured operation is 128 already-available events, not an arrival stream.
// Setup/close are excluded. No queue-delay or service-throughput claim follows.
func BenchmarkResearchEventBatch(b *testing.B) {
	for _, size := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("batch%d", size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				s, _ := batchStore(b)
				writes := make([]ResearchEventWrite, 128)
				for i := range writes {
					writes[i] = batchWrite(fmt.Sprint(i))
				}
				b.StartTimer()
				for i := 0; i < len(writes); i += size {
					var err error
					if size == 1 {
						w := writes[i]
						_, err = s.Put(context.Background(), w.Event, w.Vector, w.Digest)
					} else {
						_, err = s.PutResearchEventBatch(context.Background(), writes[i:i+size])
					}
					if err != nil {
						b.Fatal(err)
					}
				}
				b.StopTimer()
				if err := s.Close(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
