package libravdbstore

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func createPublishedPayloadGateV1(t *testing.T, root string) *incrementalSortGate {
	t.Helper()
	ctx := context.Background()
	g, err := openSortPublicationGate(root, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.store.db.CreateCollection(ctx, g.name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true), libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
			"event_json": libra.StringField, "corpus_text": libra.StringField,
			"raw_content": libra.StringField,
		}))
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writes := []ResearchEventWrite{
		pinnedWrite("past100", second.Add(100*time.Millisecond)),
		pinnedWrite("at120", second.Add(120*time.Millisecond)),
		pinnedWrite("future125", second.Add(125*time.Millisecond)),
	}
	if _, lsn, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || lsn == 0 {
		g.close()
		t.Fatal("declared-payload genesis batch", lsn, err)
	}
	if err := g.publish(ctx); err != nil {
		g.close()
		t.Fatal(err)
	}
	gate := &incrementalSortGate{g}
	if err := gate.initJournal(ctx); err != nil {
		g.close()
		t.Fatal(err)
	}
	return gate
}

func queryPublishedPayloadV1(ctx context.Context, g *incrementalSortGate, lsn uint64, asOf time.Time) (map[string]bool, error) {
	lease, err := g.store.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	query := "SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <-> $query_vec AS distance FROM " + g.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	rows, err := g.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(asOf),
	})
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(rows.Results))
	for _, row := range rows.Results {
		if seen[row.ID] || math.IsNaN(float64(row.Score)) || math.IsInf(float64(row.Score), 0) {
			return nil, fmt.Errorf("duplicate or non-finite result for %q", row.ID)
		}
		event, err := decodeStoredEvent(row.Metadata, true)
		if err != nil {
			return nil, fmt.Errorf("decode exact-LSN payload %q: %w", row.ID, err)
		}
		if event.ID != row.ID || event.TenantID != "tenant-a" || event.AvailableAt.After(asOf) ||
			row.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) ||
			event.Content != "public vector search fixture" {
			return nil, fmt.Errorf("exact-LSN body mismatch for %q", row.ID)
		}
		seen[row.ID] = true
	}
	return seen, nil
}

func TestResearchPublishedPayloadV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_PAYLOAD_V1") != "1" {
		t.Skip("opt-in certified-LSN EventFrame payload projection probe")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	legacy := createIncrementalSortGate(t, t.TempDir())
	legacyLSN, ok := legacy.capture(ctx)
	if !ok {
		t.Fatal("availability-only fixture is not published")
	}
	legacyIDs, legacyErr := queryPublishedPayloadV1(ctx, legacy, legacyLSN, asOf)
	t.Logf("availability_only_schema_decodable=%v error=%v ids=%v", legacyErr == nil, legacyErr, legacyIDs)
	if err := legacy.close(); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	g := createPublishedPayloadGateV1(t, root)
	defer func() {
		if g != nil {
			g.close()
		}
	}()
	r, err := newSortPublishedReader(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	oldLSN := r.current.Load().lsn
	newWrite := pinnedWrite("new110", second.Add(110*time.Millisecond))
	if err := g.appendBatch(ctx, []ResearchEventWrite{newWrite}, ""); err != nil {
		t.Fatal(err)
	}
	if err := r.publish(ctx); err != nil {
		t.Fatal(err)
	}
	newLSN := r.current.Load().lsn
	oldIDs, err := queryPublishedPayloadV1(ctx, g, oldLSN, asOf)
	if err != nil || len(oldIDs) != 2 || !oldIDs["past100"] || !oldIDs["at120"] {
		t.Fatal("old certified LSN returned wrong EventFrame bodies", oldIDs, err)
	}
	newIDs, err := queryPublishedPayloadV1(ctx, g, newLSN, asOf)
	if err != nil || len(newIDs) != 3 || !newIDs["past100"] || !newIDs["at120"] || !newIDs["new110"] {
		t.Fatal("new certified LSN returned wrong EventFrame bodies", newIDs, err)
	}
	t.Logf("old_lsn=%d new_lsn=%d old_ids=%v new_ids=%v", oldLSN, newLSN, oldIDs, newIDs)
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	g = nil
	g, err = openIncrementalSortGate(root)
	if err != nil {
		t.Fatal(err)
	}
	if reopened, ok := g.capture(ctx); !ok || reopened != newLSN {
		t.Fatal("reopened declared-payload gate lost certified LSN", reopened, ok)
	}
	reopenedIDs, err := queryPublishedPayloadV1(ctx, g, newLSN, asOf)
	if err != nil || len(reopenedIDs) != 3 || !reopenedIDs["new110"] {
		t.Fatal("reopened exact-LSN payload query failed", reopenedIDs, err)
	}
}
