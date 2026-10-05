package researchpublicationstore

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func BenchmarkResearchEventContinuity(b *testing.B) {
	for _, durable := range []bool{false, true} {
		name := "memory"
		if durable {
			name = "sqlite-sidecar"
		}
		b.Run(name, func(b *testing.B) {
			var s *Store
			var err error
			if durable {
				s, err = NewWithDurableLineage(memorystore.New(), filepath.Join(b.TempDir(), "lineage.sqlite"), true)
			} else {
				s, err = New(memorystore.New())
			}
			if err != nil {
				b.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			if _, err := s.Put(ctx, testutil.Event("a", "public", now), []float32{1, 0, 0, 0}, "a"); err != nil {
				b.Fatal(err)
			}
			from := s.Snapshot(ctx)
			if _, err := s.Put(ctx, testutil.Event("b", "public", now), []float32{1, 0, 0, 0}, "b"); err != nil {
				b.Fatal(err)
			}
			target := s.Snapshot(ctx)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ok, err := s.ResearchEventUnchanged(ctx, from, target, "tenant-a", "a")
				if err != nil || !ok {
					b.Fatalf("continuity failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkResearchEventPut(b *testing.B) {
	for _, durable := range []bool{false, true} {
		name := "memory"
		if durable {
			name = "sqlite-sidecar"
		}
		b.Run(name, func(b *testing.B) {
			var s *Store
			var err error
			if durable {
				s, err = NewWithDurableLineage(memorystore.New(), filepath.Join(b.TempDir(), "lineage.sqlite"), true)
			} else {
				s, err = New(memorystore.New())
			}
			if err != nil {
				b.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				id := strconv.Itoa(i)
				if _, err := s.Put(ctx, testutil.Event(id, "public", now), []float32{1, 0, 0, 0}, id); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
