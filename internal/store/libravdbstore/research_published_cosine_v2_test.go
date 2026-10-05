package libravdbstore

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchPublishedCosineV2(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_COSINE_V2") != "1" {
		t.Skip("opt-in certified-LSN cosine score conversion probe")
	}
	ctx := context.Background()
	g := createPublishedPayloadGateV1(t, t.TempDir())
	defer g.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	angled := pinnedWrite("angled", second.Add(110*time.Millisecond))
	angled.Vector = []float32{3, 4, 0, 0}
	orthogonal := pinnedWrite("orthogonal", second.Add(111*time.Millisecond))
	orthogonal.Vector = []float32{0, 2, 0, 0}
	opposite := pinnedWrite("opposite", second.Add(112*time.Millisecond))
	opposite.Vector = []float32{-1, 0, 0, 0}
	if err := g.appendBatch(ctx, []ResearchEventWrite{angled, orthogonal, opposite}, ""); err != nil {
		t.Fatal(err)
	}
	r, err := newSortPublishedReader(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	lsn := r.current.Load().lsn
	query := "SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <=> $query_vec AS distance FROM " + g.name +
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
	if len(rows.Results) != 5 || len(ordinaryScores) != 5 {
		t.Fatal("wrong as-of result count", len(rows.Results), len(ordinaryScores))
	}
	seen := make(map[string]bool, len(rows.Results))
	previous := math.Inf(1)
	for _, row := range rows.Results {
		if seen[row.ID] || row.ID == "future125" {
			t.Fatal("duplicate or future SQL row", row.ID)
		}
		seen[row.ID] = true
		event, err := decodeStoredEvent(row.Metadata, true)
		if err != nil || event.ID != row.ID || event.TenantID != "tenant-a" || event.AvailableAt.After(asOf) ||
			event.Content != "public vector search fixture" || row.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) {
			t.Fatal("SQL row body or availability mismatch", row.ID, err)
		}
		distance, numeric := publishedScoreNumberV1(row.Metadata["distance"])
		if !numeric || math.IsNaN(distance) || math.IsInf(distance, 0) {
			t.Fatal("non-finite or missing cosine distance", row.ID, row.Metadata["distance"])
		}
		similarity := math.Max(-1, math.Min(1, 1-distance))
		want, known := ordinaryScores[row.ID]
		t.Logf("id=%s SQL_score=%g cosine_distance=%g converted_similarity=%g Store_similarity=%g", row.ID, row.Score, distance, similarity, want)
		if !known || math.Abs(similarity-want) > 1e-5 || similarity > previous+1e-5 {
			t.Errorf("certified SQL cosine score/order disagrees with ordinary Search for %q", row.ID)
		}
		previous = similarity
	}
	for _, id := range []string{"past100", "at120", "angled", "orthogonal", "opposite"} {
		if !seen[id] {
			t.Error("missing as-of-visible contrast row", id)
		}
	}
}
