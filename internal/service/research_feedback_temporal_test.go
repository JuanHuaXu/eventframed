package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func temporalBridgeFixture(t *testing.T, persistent bool) (*Service, *ResearchFeedbackBridge, time.Time) {
	t.Helper()
	em, e := embed.NewHashEmbedder(8)
	if e != nil {
		t.Fatal(e)
	}
	var db store.EventStore = memorystore.New()
	if persistent {
		db, e = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/bridge.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
		if e != nil {
			t.Fatal(e)
		}
	}
	tap, e := NewResearchFrontierTap("tenant-a", 4)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(tap.Close)
	s, e := New(db, em, Config{ResearchFrontier: tap, DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	b, e := NewResearchTemporalFeedbackBridge(s, "tenant-a", 42)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(b.Close)
	return s, b, now
}

func putTemporalFixture(t *testing.T, s *Service, id string, at time.Time) {
	t.Helper()
	ev := testutil.Event(id, "public query fixture", at)
	if _, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: ev}); e != nil {
		t.Fatal(e)
	}
}

func takeTemporalFixture(t *testing.T, s *Service, at time.Time) ResearchFrontierObservation {
	t.Helper()
	shadowRecall(t, s, at)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	in, e := s.config.ResearchFrontier.Take(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return in
}

func TestResearchTemporalFeedbackBoundaries(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			s, b, now := temporalBridgeFixture(t, persistent)
			in := takeTemporalFixture(t, s, now)
			strict, e := NewResearchFeedbackBridge(s, "tenant-a", 42)
			if e != nil {
				t.Fatal(e)
			}
			defer strict.Close()
			if _, e := strict.Admit(context.Background(), in); e != nil {
				t.Fatal(e)
			}
			if _, e := b.Admit(context.Background(), in); e != nil {
				t.Fatal(e)
			}
			putTemporalFixture(t, s, "future", now.Add(time.Hour))
			if e := strict.Feedback(context.Background(), in.JournalID, in.Candidates[0].EventID, true, now.Add(time.Second)); e == nil {
				t.Fatal("default exact mode changed behavior")
			}
			if e := b.Feedback(context.Background(), in.JournalID, in.Candidates[0].EventID, true, now.Add(time.Second)); e != nil {
				t.Fatal("benign future ingestion rejected", e)
			}
			until := time.Now().Add(time.Second)
			for {
				n, failed, _, _ := b.worker.Counts()
				if failed != 0 {
					t.Fatal("future-compatible worker update failed")
				}
				if n == 1 {
					break
				}
				if time.Now().After(until) {
					t.Fatal("accepted label did not complete")
				}
				time.Sleep(time.Millisecond)
			}
			if _, e := b.Score(context.Background(), 1, .6, now.Add(2*time.Second)); e != nil {
				t.Fatal("future proof did not permit score", e)
			}
			if _, e := b.Score(context.Background(), 1, .6, now.Add(2*time.Hour)); e == nil {
				t.Fatal("future evidence became visible without invalidation")
			}
			// New journal has a newer physical snapshot, but unchanged as-of data.
			later := takeTemporalFixture(t, s, now.Add(3*time.Second))
			if _, e := b.Admit(context.Background(), later); e != nil {
				t.Fatal("new compatible journal rejected", e)
			}
			putTemporalFixture(t, s, "backfill", now)
			if e := b.Feedback(context.Background(), later.JournalID, later.Candidates[0].EventID, true, now.Add(4*time.Second)); e == nil {
				t.Fatal("backfill accepted")
			}
			if _, e := b.Score(context.Background(), 1, .6, now.Add(4*time.Second)); e == nil {
				t.Fatal("backfill score exposed")
			}
		})
	}
}

func TestResearchTemporalFeedbackLatestQueryAndFallback(t *testing.T) {
	s, b, now := temporalBridgeFixture(t, false)
	old := takeTemporalFixture(t, s, now)
	if _, e := b.Admit(context.Background(), old); e != nil {
		t.Fatal(e)
	}
	newer := takeTemporalFixture(t, s, now.Add(10*time.Minute))
	if _, e := b.Admit(context.Background(), newer); e != nil {
		t.Fatal(e)
	}
	putTemporalFixture(t, s, "between", now.Add(5*time.Minute))
	if e := b.Feedback(context.Background(), old.JournalID, old.Candidates[0].EventID, true, now.Add(11*time.Minute)); e == nil {
		t.Fatal("old label ignored newer admitted query")
	}
	s2, b2, now2 := temporalBridgeFixture(t, false)
	in := takeTemporalFixture(t, s2, now2)
	if _, e := b2.Admit(context.Background(), in); e != nil {
		t.Fatal(e)
	}
	putTemporalFixture(t, s2, "future", now2.Add(time.Hour))
	s2.store = struct{ store.EventStore }{s2.store}
	if e := b2.Feedback(context.Background(), in.JournalID, in.Candidates[0].EventID, true, now2); e == nil {
		t.Fatal("missing optional proof accepted")
	}
}
