package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchLSNAvailabilitySchemaV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_LSN_AVAILABILITY_SCHEMA_V1") != "1" {
		t.Skip("opt-in SQL availability-schema contract probe")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	path := t.TempDir() + "/availability.libravdb"
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
	if _, err := s.db.CreateCollection(ctx, name,
		libra.WithDimension(4), libra.WithMetric(libra.CosineDistance),
		libra.WithHNSW(16, 200, 100), libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{"available_at": libra.StringField})); err != nil {
		t.Fatal("create schema-backed tenant collection", err)
	}
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	initial := []ResearchEventWrite{
		pinnedWrite("past", asOf.Add(-time.Minute)),
		pinnedWrite("future", asOf.Add(time.Hour)),
	}
	if _, err := s.PutResearchEventBatch(ctx, initial); err != nil {
		t.Fatal("seed", err)
	}
	oldLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal("old LSN", err)
	}
	oldLease, err := s.db.SnapshotAtLSN(ctx, oldLSN)
	if err != nil {
		t.Fatal("old LSN lease", err)
	}
	defer oldLease.Close()
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{pinnedWrite("visible", asOf)}); err != nil {
		t.Fatal("visible write", err)
	}
	newLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil || newLSN <= oldLSN {
		t.Fatal("new LSN", oldLSN, newLSN, err)
	}
	newLease, err := s.db.SnapshotAtLSN(ctx, newLSN)
	if err != nil {
		t.Fatal("new LSN lease", err)
	}
	defer newLease.Close()
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + name +
		" AS OF LSN $snapshot_lsn d WHERE available_at <= $available_by ORDER BY distance LIMIT 10"
	queryIDs := func(db *libra.Database, lsn uint64) (map[string]bool, error) {
		rows, err := db.QueryWithParams(ctx, query, libra.QueryParams{
			"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
			"available_by": asOf.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		})
		if err != nil {
			return nil, err
		}
		ids := make(map[string]bool, len(rows.Results))
		for _, row := range rows.Results {
			if ids[row.ID] {
				return nil, fmt.Errorf("duplicate SQL row %s", row.ID)
			}
			ids[row.ID] = true
		}
		return ids, nil
	}
	assertIDs := func(label string, ids map[string]bool, want ...string) {
		t.Helper()
		if len(ids) != len(want) {
			t.Fatalf("%s IDs=%v, want %v", label, ids, want)
		}
		for _, id := range want {
			if !ids[id] {
				t.Fatalf("%s IDs=%v, missing %s", label, ids, id)
			}
		}
	}
	oldIDs, err := queryIDs(s.db, oldLSN)
	if err != nil {
		t.Fatal("old-LSN availability SQL", err)
	}
	assertIDs("old LSN", oldIDs, "past")
	newIDs, err := queryIDs(s.db, newLSN)
	if err != nil {
		t.Fatal("new-LSN availability SQL", err)
	}
	assertIDs("new LSN", newIDs, "past", "visible")
	current, err := s.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 10)
	if err != nil {
		t.Fatal("current Search", err)
	}
	currentIDs := make(map[string]bool, len(current))
	for _, row := range current {
		currentIDs[row.Event.ID] = true
	}
	assertIDs("current Search", currentIDs, "past", "visible")
	oldLease.Close()
	newLease.Close()
	stats, err := s.db.TemporalStats(ctx)
	if err != nil || stats.ActiveLeaseCount != 0 {
		t.Fatal("open temporal leases", stats, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal("close", err)
	}
	s = nil
	reopened, err := Open(config)
	if err != nil {
		t.Fatal("reopen", err)
	}
	defer reopened.Close()
	reopenedIDs, err := queryIDs(reopened.db, newLSN)
	if err != nil {
		t.Fatal("reopened availability SQL", err)
	}
	assertIDs("reopened new LSN", reopenedIDs, "past", "visible")
	t.Logf("availability schema passed old_lsn=%d new_lsn=%d old=%v new=%v reopened=%v", oldLSN, newLSN, oldIDs, newIDs, reopenedIDs)
}
