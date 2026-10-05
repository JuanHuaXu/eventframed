package researchpublicationstore

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestDurableLineageRestartsWithoutRevivingSources(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true}
	lineagePath := filepath.Join(root, "lineage.sqlite")
	open := func(create bool) *Store {
		t.Helper()
		backend, err := libravdbstore.Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewWithDurableLineage(backend, lineagePath, create)
		if err != nil {
			_ = backend.Close()
			t.Fatal(err)
		}
		return s
	}
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	a := testutil.Event("a", "public a", now)
	b := testutil.Event("b", "public b", now)
	s := open(true)
	for _, event := range []model.Event{a, b} {
		if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, event.ID); err != nil {
			t.Fatal(err)
		}
	}
	from := s.Snapshot(ctx)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	if _, err := s.BindBayesianPolicy(ctx, "policy-a"); err != nil {
		t.Fatal(err)
	}
	if same, err := s.ResearchEventUnchanged(ctx, from, s.Snapshot(ctx), "tenant-a", "b"); err != nil || !same {
		t.Fatalf("non-event version revoked b: same=%v err=%v", same, err)
	}
	if _, err := s.Delete(ctx, "tenant-a", "a"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	for _, tc := range []struct {
		id   string
		want bool
	}{{"a", false}, {"b", true}} {
		same, err := s.ResearchEventUnchanged(ctx, from, s.Snapshot(ctx), "tenant-a", tc.id)
		if err != nil || same != tc.want {
			t.Fatalf("after restart %s: same=%v err=%v", tc.id, same, err)
		}
	}
	if _, err := s.Delete(ctx, "tenant-a", "b"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Put(ctx, b, []float32{1, 0, 0, 0}, b.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	defer s.Close()
	if same, err := s.ResearchEventUnchanged(ctx, from, s.Snapshot(ctx), "tenant-a", "b"); err != nil || same {
		t.Fatalf("same-ID resurrection passed after restart: same=%v err=%v", same, err)
	}
}

func TestDurableLineageWriteFailureFaultsWrapper(t *testing.T) {
	ctx := context.Background()
	backend := memorystore.New()
	path := filepath.Join(t.TempDir(), "lineage.sqlite")
	s, err := NewWithDurableLineage(backend, path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	before := s.Snapshot(ctx)
	if err := s.lineage.Close(); err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("a", "public a", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, event.ID); err == nil {
		t.Fatal("post-backend lineage failure was acknowledged")
	}
	after := s.Snapshot(ctx)
	if after == before || !s.lineageFault.Load() {
		t.Fatal("backend commit or fault not visible")
	}
	if same, err := s.ResearchEventUnchanged(ctx, before, after, "tenant-a", "a"); err == nil || same {
		t.Fatal("faulted wrapper certified continuity")
	}
	if s.ResearchPublicationCompatible(ctx, after, event.AvailableAt) {
		t.Fatal("faulted wrapper retained publication authority")
	}
	if _, err := NewWithDurableLineage(backend, path, false); err == nil {
		t.Fatal("restart accepted unrecorded backend commit")
	}
}

func TestDurableLineageDuplicateStillQuarantines(t *testing.T) {
	ctx := context.Background()
	s, err := NewWithDurableLineage(memorystore.New(), filepath.Join(t.TempDir(), "lineage.sqlite"), true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	event := testutil.Event("a", "public", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, "a"); err != nil {
		t.Fatal(err)
	}
	before := s.Snapshot(ctx)
	if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, "a"); err == nil {
		t.Fatal("duplicate unexpectedly resolved")
	}
	if s.Snapshot(ctx) != before {
		t.Fatal("duplicate moved the backend snapshot")
	}
	if same, err := s.ResearchEventUnchanged(ctx, before, before, "tenant-a", "a"); err == nil || same {
		t.Fatal("quarantined duplicate retained source authority")
	}
}
