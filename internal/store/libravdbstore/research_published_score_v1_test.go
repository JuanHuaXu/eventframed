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

func publishedScoreNumberV1(value interface{}) (float64, bool) {
	switch n := value.(type) {
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func TestResearchPublishedScoreV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_SCORE_V1") != "1" {
		t.Skip("opt-in exact-LSN score contrast probe")
	}
	ctx := context.Background()
	root := t.TempDir()
	g, err := openSortPublicationGate(root, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()
	_, err = g.store.db.CreateCollection(ctx, g.name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true), libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
			"event_json": libra.StringField, "corpus_text": libra.StringField,
			"raw_content": libra.StringField,
		}))
	if err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	writes := []ResearchEventWrite{
		pinnedWrite("aligned", second.Add(100*time.Millisecond)),
		pinnedWrite("orthogonal", second.Add(120*time.Millisecond)),
		pinnedWrite("future", second.Add(125*time.Millisecond)),
	}
	writes[1].Vector = []float32{0, 1, 0, 0}
	if _, receipt, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || receipt == 0 {
		t.Fatal("contrast writes", receipt, err)
	}
	if err := g.publish(ctx); err != nil {
		t.Fatal(err)
	}
	gate := &incrementalSortGate{g}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	lsn, ok := gate.capture(ctx)
	if !ok {
		t.Fatal("declared contrast LSN is not certified")
	}
	query := "SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <-> $query_vec AS distance FROM " + g.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	rows, err := g.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(asOf),
	})
	if err != nil {
		t.Fatal(err)
	}
	ordinary, err := g.store.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 10)
	if err != nil {
		t.Fatal(err)
	}
	ordinaryScores := make(map[string]float64, len(ordinary))
	for _, item := range ordinary {
		ordinaryScores[item.Event.ID] = item.Similarity
	}
	if len(rows.Results) != 2 || len(ordinaryScores) != 2 || rows.Results[0].ID != "aligned" || rows.Results[1].ID != "orthogonal" {
		t.Fatal("contrast retrieval set or order", rows.Results, ordinaryScores)
	}
	distances := make(map[string]float64)
	for _, row := range rows.Results {
		event, err := decodeStoredEvent(row.Metadata, true)
		if err != nil || event.ID != row.ID || event.Content != "public vector search fixture" {
			t.Fatal("projected EventFrame body mismatch", row.ID, err)
		}
		distance, numeric := publishedScoreNumberV1(row.Metadata["distance"])
		if !numeric || math.IsNaN(distance) || math.IsInf(distance, 0) {
			t.Fatal("projected distance is missing or non-finite", row.ID, row.Metadata["distance"])
		}
		distances[row.ID] = distance
		t.Logf("id=%s SQL_score=%.9f SQL_distance=%.9f Store_similarity=%.9f", row.ID, row.Score, distance, ordinaryScores[row.ID])
	}
	if distances["aligned"] >= distances["orthogonal"] {
		t.Fatal("SQL distance did not distinguish aligned from orthogonal")
	}
	for _, row := range rows.Results {
		if math.Abs(float64(row.Score)-ordinaryScores[row.ID]) > 1e-5 {
			t.Errorf("SQL SearchResult.Score is not ordinary Store similarity for %q: score=%g similarity=%g", row.ID, row.Score, ordinaryScores[row.ID])
		}
	}
	if len(distances) != 2 {
		t.Fatal(fmt.Errorf("incomplete contrast distances: %v", distances))
	}
}
