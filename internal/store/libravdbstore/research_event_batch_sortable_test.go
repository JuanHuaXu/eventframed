package libravdbstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchSortableEventBatch(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/sortable-batch.libravdb"
	config := Config{Path: path, Dimension: 4, EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true}
	s, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if s != nil {
			s.Close()
		}
	}()
	name := collectionName("tenant-a", "research:d4")
	col, err := s.db.CreateCollection(ctx, name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
		}))
	if err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	early, late := second.Add(100*time.Millisecond), second.Add(150*time.Millisecond)
	writes := []ResearchEventWrite{pinnedWrite("early", early), pinnedWrite("late", late)}
	base := s.Snapshot(ctx)
	if _, err := s.PutResearchEventBatchSortable(ctx, writes); err != nil {
		t.Fatal("sortable insert", err)
	}
	after := s.Snapshot(ctx)
	if after.RuntimeVersion != base.RuntimeVersion+2 || after.EvidenceEpoch != base.EvidenceEpoch+2 ||
		!s.ingestMotion[base.RuntimeVersion+1].Equal(early) || !s.ingestMotion[base.RuntimeVersion+2].Equal(late) {
		t.Fatal("sortable batch lost per-event version or motion")
	}
	for _, w := range writes {
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got := record.Metadata["available_at_sort"]; got != researchAvailabilitySortKey(w.Event.AvailableAt) {
			t.Fatal("sort key differs from EventFrame time", w.Event.ID, got)
		}
	}
	if researchAvailabilitySortKey(early) != "2026-10-02T00:00:00.100000000Z" {
		t.Fatal("sort key is not fixed-width UTC")
	}
	query := "SELECT id FROM " + name + " AS OF LSN $snapshot_lsn WHERE available_at_sort <= $available_by_sort ORDER BY id"
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	check := func(db *libra.Database) {
		t.Helper()
		rows, err := db.QueryWithParams(ctx, query, libra.QueryParams{
			"snapshot_lsn": int64(lsn), "available_by_sort": researchAvailabilitySortKey(second.Add(120 * time.Millisecond)),
		})
		if err != nil || len(rows.Results) != 1 || rows.Results[0].ID != "early" {
			t.Fatal("sortable as-of query", rows, err)
		}
	}
	check(s.db)
	retries, err := s.PutResearchEventBatchSortable(ctx, writes)
	if err != nil || len(retries) != 2 || !retries[0].Duplicate || !retries[1].Duplicate || s.Snapshot(ctx) != after {
		t.Fatal("sortable exact retry", retries, err)
	}
	legacy := pinnedWrite("legacy", second.Add(90*time.Millisecond))
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{legacy}); err != nil {
		t.Fatal("legacy row", err)
	}
	beforeConflict := s.Snapshot(ctx)
	if _, err := s.PutResearchEventBatchSortable(ctx, []ResearchEventWrite{legacy}); !errors.Is(err, store.ErrIdempotencyConflict) || s.Snapshot(ctx) != beforeConflict {
		t.Fatal("missing sort key did not fail closed", err)
	}
	a := pinnedWrite("same", early)
	b := pinnedWrite("same", late)
	if _, err := s.PutResearchEventBatchSortable(ctx, []ResearchEventWrite{a, b}); !errors.Is(err, store.ErrIdempotencyConflict) || s.Snapshot(ctx) != beforeConflict {
		t.Fatal("conflicting intra-batch availability did not fail closed", err)
	}
	forged := pinnedWrite("forged", early)
	stored := pinnedWrite("forged", late)
	payload, err := encodeStoredEvent(stored.Event)
	if err != nil {
		t.Fatal(err)
	}
	if err := col.Insert(ctx, forged.Event.ID, forged.Vector, map[string]interface{}{
		"event_json": string(payload), "content_digest": forged.Digest,
		"available_at":      stored.Event.AvailableAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"available_at_sort": researchAvailabilitySortKey(forged.Event.AvailableAt),
	}); err != nil {
		t.Fatal("private payload/key mismatch control", err)
	}
	if _, err := s.PutResearchEventBatchSortable(ctx, []ResearchEventWrite{forged}); !errors.Is(err, store.ErrIdempotencyConflict) || s.Snapshot(ctx) != beforeConflict {
		t.Fatal("sort key certified a mismatched stored EventFrame", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = nil
	reopened, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	check(reopened.db)
}
