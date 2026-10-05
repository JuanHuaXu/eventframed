package researchpublicationstore

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestDurableMotionRestoresOnlyFutureIngestion(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true}
	path := filepath.Join(root, "lineage.sqlite")
	open := func(create bool) *Store {
		t.Helper()
		backend, err := libravdbstore.Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewWithDurableLineage(backend, path, create)
		if err != nil {
			_ = backend.Close()
			t.Fatal(err)
		}
		return s
	}
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	s := open(true)
	if _, err := s.Put(ctx, testutil.Event("old", "public", now.Add(-time.Minute)), []float32{1, 0, 0, 0}, "old"); err != nil {
		t.Fatal(err)
	}
	from := s.Snapshot(ctx)
	if _, err := s.Put(ctx, testutil.Event("future", "public", now.Add(time.Hour)), []float32{1, 0, 0, 0}, "future"); err != nil {
		t.Fatal(err)
	}
	if !s.ResearchPublicationCompatible(ctx, from, now) {
		t.Fatal("pre-restart future-only motion failed")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	if !s.ResearchPublicationCompatible(ctx, from, now) {
		t.Fatal("durable wrapper forgot future-only motion on restart")
	}
	if s.ResearchPublicationCompatible(ctx, from, now.Add(time.Hour)) {
		t.Fatal("at-availability query accepted old snapshot")
	}
	if _, err := s.Put(ctx, testutil.Event("past", "public", now.Add(-time.Second)), []float32{1, 0, 0, 0}, "past"); err != nil {
		t.Fatal(err)
	}
	if s.ResearchPublicationCompatible(ctx, from, now) {
		t.Fatal("past ingestion retained old authority")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	defer s.Close()
	if s.ResearchPublicationCompatible(ctx, from, now) {
		t.Fatal("past ingestion became valid after restart")
	}
}

func TestDurableMotionGeneralMutationBlocksOldSnapshot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true}
	path := filepath.Join(root, "lineage.sqlite")
	open := func(create bool) *Store {
		t.Helper()
		backend, err := libravdbstore.Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewWithDurableLineage(backend, path, create)
		if err != nil {
			_ = backend.Close()
			t.Fatal(err)
		}
		return s
	}
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	s := open(true)
	from := s.Snapshot(ctx)
	if _, err := s.Put(ctx, testutil.Event("future", "public", now.Add(time.Hour)), []float32{1, 0, 0, 0}, "future"); err != nil {
		t.Fatal(err)
	}
	if !s.ResearchPublicationCompatible(ctx, from, now) {
		t.Fatal("future-only control did not pass")
	}
	if _, err := s.BindBayesianPolicy(ctx, "policy-a"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false)
	defer s.Close()
	if s.ResearchPublicationCompatible(ctx, from, now) || !s.ResearchPublicationCompatible(ctx, s.Snapshot(ctx), now) {
		t.Fatal("general mutation either retained old authority or denied current state")
	}
}

func TestDurableMotionMissingAndMalformedRowsFailClosed(t *testing.T) {
	for _, mode := range []string{"missing-row", "malformed-row", "legacy-table"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true}
			path := filepath.Join(root, "lineage.sqlite")
			backend, err := libravdbstore.Open(cfg)
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewWithDurableLineage(backend, path, true)
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			from := s.Snapshot(ctx)
			if _, err := s.Put(ctx, testutil.Event("future", "public", now.Add(time.Hour)), []float32{1, 0, 0, 0}, "future"); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			statement := "DELETE FROM lineage_motion"
			switch mode {
			case "malformed-row":
				statement = "UPDATE lineage_motion SET available_at='not-a-time'"
			case "legacy-table":
				statement = "DROP TABLE lineage_motion"
			}
			if _, err := db.Exec(statement); err != nil {
				t.Fatal(err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			backend, err = libravdbstore.Open(cfg)
			if err != nil {
				t.Fatal(err)
			}
			s, err = NewWithDurableLineage(backend, path, false)
			if mode == "malformed-row" {
				if err == nil {
					_ = s.Close()
					t.Fatal("malformed motion row admitted")
				}
				_ = backend.Close()
				return
			}
			if err != nil {
				_ = backend.Close()
				t.Fatal(err)
			}
			defer s.Close()
			if s.ResearchPublicationCompatible(ctx, from, now) || !s.ResearchPublicationCompatible(ctx, s.Snapshot(ctx), now) {
				t.Fatal("missing prior motion was invented or current authority lost")
			}
		})
	}
}
