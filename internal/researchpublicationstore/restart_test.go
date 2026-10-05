package researchpublicationstore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestOrderlyRestartDoesNotInventPublicationHistory(t *testing.T) {
	ctx := context.Background()
	cfg := libravdbstore.Config{Path: t.TempDir() + "/restart.libravdb", Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true}
	open := func() *Store {
		t.Helper()
		db, e := libravdbstore.Open(cfg)
		if e != nil {
			t.Fatal(e)
		}
		s, e := New(db)
		if e != nil {
			db.Close()
			t.Fatal(e)
		}
		return s
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	first := open()
	ev := testutil.Event("future", "public fixture", now.Add(time.Hour))
	before := first.Snapshot(ctx)
	if _, e := first.Put(ctx, ev, []float32{1, 0, 0, 0}, "future-digest"); e != nil {
		first.Close()
		t.Fatal(e)
	}
	committed := first.Snapshot(ctx)
	if !first.ResearchPublicationCompatible(ctx, before, now) {
		first.Close()
		t.Fatal("future proof absent before restart")
	}
	if e := first.Close(); e != nil {
		t.Fatal(e)
	}
	if first.ResearchPublicationCompatible(ctx, committed, now) {
		t.Fatal("closed adapter retained authority")
	}

	second := open()
	if second.Snapshot(ctx) != committed || !second.ResearchPublicationCompatible(ctx, committed, now) {
		second.Close()
		t.Fatal("committed state not restored")
	}
	if second.ResearchPublicationCompatible(ctx, before, now) {
		second.Close()
		t.Fatal("invented pre-restart motion proof")
	}
	visible, e := second.GetEvents(ctx, ev.TenantID, []string{ev.ID}, now.Add(time.Hour))
	if e != nil || len(visible) != 1 || visible[0].ID != ev.ID {
		second.Close()
		t.Fatal("event not durable", e)
	}
	// Reopened stores expose vectors independently of their publication history.
	vectors, e := second.EventStore.(store.VectorEventStore).GetEventsWithVectors(ctx, ev.TenantID, []string{ev.ID}, now.Add(time.Hour))
	if e != nil || len(vectors) != 1 || !reflect.DeepEqual(vectors[0].Embedding, []float32{1, 0, 0, 0}) {
		second.Close()
		t.Fatal("vector not durable", e)
	}
	later := testutil.Event("later", "public second fixture", now.Add(2*time.Hour))
	if _, e = second.Put(ctx, later, []float32{0, 1, 0, 0}, "later-digest"); e != nil {
		second.Close()
		t.Fatal(e)
	}
	if !second.ResearchPublicationCompatible(ctx, committed, now) {
		second.Close()
		t.Fatal("new lifetime history not recorded")
	}
	// This is a known availability limitation, not a claim of duplicate recovery.
	if _, e = second.Put(ctx, later, []float32{0, 1, 0, 0}, "later-digest"); e == nil {
		second.Close()
		t.Fatal("duplicate unexpectedly resolved")
	}
	final := second.Snapshot(ctx)
	if second.ResearchPublicationCompatible(ctx, final, now) {
		second.Close()
		t.Fatal("quarantine ignored")
	}
	if e = second.Close(); e != nil {
		t.Fatal(e)
	}

	third := open()
	defer third.Close()
	if third.Snapshot(ctx) != final || !third.ResearchPublicationCompatible(ctx, final, now) {
		t.Fatal("clean reopen failed")
	}
	if third.ResearchPublicationCompatible(ctx, committed, now) {
		t.Fatal("quarantine reopen invented motion history")
	}
	// The new lifetime trusts only the backend's completed open. This controlled
	// duplicate case is not a simulation of power loss or uncertain disk writes.
	if _, e = third.BindBayesianPolicy(ctx, "after-reopen"); e != nil {
		t.Fatal(e)
	}
	if third.ResearchPublicationCompatible(ctx, final, now) {
		t.Fatal("policy update retained stale authority")
	}
}
