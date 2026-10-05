package libravdbstore

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchSortableEventBatchReceipt(t *testing.T) {
	ctx := context.Background()
	config := Config{Path: t.TempDir() + "/receipt.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true}
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
	early := pinnedWrite("early", second.Add(100*time.Millisecond))
	late := pinnedWrite("late", second.Add(150*time.Millisecond))
	middle := pinnedWrite("middle", second.Add(110*time.Millisecond))
	base := s.Snapshot(ctx)
	results, firstLSN, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{early, late})
	if err != nil || firstLSN == 0 || len(results) != 2 || results[0].Duplicate || results[1].Duplicate {
		t.Fatal("first batch receipt", firstLSN, results, err)
	}
	latest, err := s.db.LatestCommitLSN(ctx)
	if err != nil || latest != firstLSN {
		t.Fatal("receipt does not name exact latest commit", latest, firstLSN, err)
	}
	after := s.Snapshot(ctx)
	if after.RuntimeVersion != base.RuntimeVersion+2 || after.EvidenceEpoch != base.EvidenceEpoch+2 ||
		!s.ingestMotion[base.RuntimeVersion+1].Equal(early.Event.AvailableAt) ||
		!s.ingestMotion[base.RuntimeVersion+2].Equal(late.Event.AvailableAt) {
		t.Fatal("receipt batch lost runtime version, epoch or motion")
	}
	for _, w := range []ResearchEventWrite{early, late} {
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil || record.Metadata["available_at_sort"] != researchAvailabilitySortKey(w.Event.AvailableAt) {
			t.Fatal("receipt batch lost sort key", w.Event.ID, err)
		}
	}
	queryAt := func(db *libra.Database, lsn uint64, want []string) {
		t.Helper()
		rows, err := db.QueryWithParams(ctx,
			"SELECT id FROM "+name+" AS OF LSN $snapshot_lsn WHERE available_at_sort <= $available_by_sort ORDER BY id",
			libra.QueryParams{"snapshot_lsn": int64(lsn),
				"available_by_sort": researchAvailabilitySortKey(second.Add(120 * time.Millisecond))})
		if err != nil {
			t.Fatal("receipt exact-LSN query", err)
		}
		got := make([]string, len(rows.Results))
		for i, row := range rows.Results {
			got[i] = row.ID
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("receipt as-of IDs", got, want)
		}
	}
	queryAt(s.db, firstLSN, []string{"early"})
	retries, retryLSN, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{early, late})
	if err != nil || retryLSN != 0 || len(retries) != 2 || !retries[0].Duplicate || !retries[1].Duplicate || s.Snapshot(ctx) != after {
		t.Fatal("exact retry changed state", retryLSN, retries, err)
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != firstLSN {
		t.Fatal("exact retry committed", latest, err)
	}
	mixed, mixedLSN, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{early, middle})
	if err != nil || mixedLSN <= firstLSN || len(mixed) != 2 || !mixed[0].Duplicate || mixed[1].Duplicate {
		t.Fatal("mixed duplicate/new receipt", mixedLSN, mixed, err)
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != mixedLSN ||
		s.Snapshot(ctx).RuntimeVersion != after.RuntimeVersion+1 ||
		!s.ingestMotion[after.RuntimeVersion+1].Equal(middle.Event.AvailableAt) {
		t.Fatal("mixed batch did not commit exactly one event", latest, err)
	}
	queryAt(s.db, mixedLSN, []string{"early", "middle"})
	conflict := pinnedWrite("early", second.Add(200*time.Millisecond))
	beforeConflict := s.Snapshot(ctx)
	if _, lsn, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{conflict}); !errors.Is(err, store.ErrIdempotencyConflict) || lsn != 0 || s.Snapshot(ctx) != beforeConflict {
		t.Fatal("conflicting duplicate changed state", lsn, err)
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != mixedLSN {
		t.Fatal("conflicting duplicate committed", latest, err)
	}
	legacy := pinnedWrite("legacy", second.Add(90*time.Millisecond))
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{legacy}); err != nil {
		t.Fatal("legacy control write", err)
	}
	beforeLegacyConflict := s.Snapshot(ctx)
	legacyLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, lsn, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{legacy}); !errors.Is(err, store.ErrIdempotencyConflict) || lsn != 0 || s.Snapshot(ctx) != beforeLegacyConflict {
		t.Fatal("unkeyed legacy row was accepted", lsn, err)
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != legacyLSN {
		t.Fatal("unkeyed legacy conflict committed", latest, err)
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
	if latest, err := reopened.db.LatestCommitLSN(ctx); err != nil || latest != legacyLSN {
		t.Fatal("reopened commit boundary moved", latest, legacyLSN, err)
	}
	queryAt(reopened.db, mixedLSN, []string{"early", "middle"})
}

func TestResearchSortableEventBatchRequiresDeclaredKey(t *testing.T) {
	ctx := context.Background()
	s, err := Open(Config{Path: t.TempDir() + "/undeclared.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	initial := s.Snapshot(ctx)
	write := pinnedWrite("undeclared", time.Date(2026, 10, 2, 0, 0, 0, 100000000, time.UTC))
	col, err := s.collection(ctx, write.Event.TenantID)
	if err != nil {
		t.Fatal(err)
	}
	if _, declared := col.Config().MetadataSchema["available_at_sort"]; declared {
		t.Fatal("negative control unexpectedly declared the sort key")
	}
	beforeLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, lsn, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write}); err == nil || lsn != 0 {
		t.Fatal("receipt writer accepted an undeclared SQL key", lsn, err)
	}
	if _, err := s.PutResearchEventBatchSortable(ctx, []ResearchEventWrite{write}); err == nil {
		t.Fatal("existing sortable writer accepted an undeclared SQL key")
	}
	if s.Snapshot(ctx) != initial {
		t.Fatal("undeclared-schema attempt changed runtime state")
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != beforeLSN {
		t.Fatal("undeclared-schema attempt committed", latest, beforeLSN, err)
	}
	if _, err := col.Get(ctx, write.Event.ID); !errors.Is(err, libra.ErrRecordNotFound) {
		t.Fatal("undeclared-schema attempt stored an event", err)
	}
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{write}); err != nil {
		t.Fatal("legacy writer was unnecessarily blocked", err)
	}
}

func TestResearchSortableEventBatchRejectsWrongKeyType(t *testing.T) {
	ctx := context.Background()
	s, err := Open(Config{Path: t.TempDir() + "/wrong-type.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	name := collectionName("tenant-a", "research:d4")
	_, err = s.db.CreateCollection(ctx, name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{"available_at_sort": libra.IntField}))
	if err != nil {
		t.Fatal(err)
	}
	beforeLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	write := pinnedWrite("wrong-type", time.Date(2026, 10, 2, 0, 0, 0, 100000000, time.UTC))
	if _, receipt, err := s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write}); err == nil || receipt != 0 {
		t.Fatal("wrong-type key was accepted", receipt, err)
	}
	if latest, err := s.db.LatestCommitLSN(ctx); err != nil || latest != beforeLSN {
		t.Fatal("wrong-type attempt committed", latest, beforeLSN, err)
	}
}
