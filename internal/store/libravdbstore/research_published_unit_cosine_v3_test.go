package libravdbstore

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func researchUnitVectorV3(vector []float32) bool {
	if len(vector) == 0 {
		return false
	}
	var normSquared float64
	for _, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return false
		}
		normSquared += float64(value) * float64(value)
	}
	return math.Abs(normSquared-1) <= 1e-4
}

func TestResearchPublishedUnitCosineV3(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_UNIT_COSINE_V3") != "1" {
		t.Skip("opt-in unit-vector certified-LSN score probe")
	}
	ctx := context.Background()
	g := createPublishedPayloadGateV1(t, t.TempDir())
	defer g.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	queryVector := []float32{1, 0, 0, 0}
	angled := pinnedWrite("angled", second.Add(110*time.Millisecond))
	angled.Vector = []float32{0.6, 0.8, 0, 0}
	orthogonal := pinnedWrite("orthogonal", second.Add(111*time.Millisecond))
	orthogonal.Vector = []float32{0, 1, 0, 0}
	opposite := pinnedWrite("opposite", second.Add(112*time.Millisecond))
	opposite.Vector = []float32{-1, 0, 0, 0}
	for _, vector := range [][]float32{queryVector, angled.Vector, orthogonal.Vector, opposite.Vector} {
		if !researchUnitVectorV3(vector) {
			t.Fatal("positive contrast vector is not unit normalized", vector)
		}
	}
	if researchUnitVectorV3([]float32{2, 0, 0, 0}) || researchUnitVectorV3([]float32{3, 4, 0, 0}) || researchUnitVectorV3([]float32{0, 0, 0, 0}) {
		t.Fatal("nonunit or zero vector passed the proposed SQL score precondition")
	}
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
		"snapshot_lsn": int64(lsn), "query_vec": queryVector,
		"available_by_sort": researchAvailabilitySortKey(asOf),
	})
	if err != nil {
		t.Fatal(err)
	}
	ordinary, err := g.store.Search(ctx, "tenant-a", queryVector, asOf, 10)
	if err != nil {
		t.Fatal(err)
	}
	wantScores := make(map[string]float64, len(ordinary))
	for _, item := range ordinary {
		wantScores[item.Event.ID] = item.Similarity
	}
	if len(rows.Results) != 5 || len(wantScores) != 5 {
		t.Fatal("wrong restricted as-of result count", len(rows.Results), len(wantScores))
	}
	seen := make(map[string]bool, 5)
	previous := math.Inf(1)
	for _, row := range rows.Results {
		if seen[row.ID] || row.ID == "future125" {
			t.Fatal("duplicate or future row", row.ID)
		}
		seen[row.ID] = true
		event, err := decodeStoredEvent(row.Metadata, true)
		if err != nil || event.ID != row.ID || event.TenantID != "tenant-a" || event.AvailableAt.After(asOf) ||
			event.Content != "public vector search fixture" || row.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) {
			t.Fatal("restricted exact-LSN body mismatch", row.ID, err)
		}
		distance, numeric := publishedScoreNumberV1(row.Metadata["distance"])
		if !numeric || math.IsNaN(distance) || math.IsInf(distance, 0) {
			t.Fatal("missing or non-finite cosine distance", row.ID, row.Metadata["distance"])
		}
		similarity := math.Max(-1, math.Min(1, 1-distance))
		want, known := wantScores[row.ID]
		t.Logf("id=%s distance=%g converted_similarity=%g Store_similarity=%g", row.ID, distance, similarity, want)
		if !known || math.Abs(similarity-want) > 1e-5 || similarity > previous+1e-5 {
			t.Errorf("unit-vector SQL conversion disagrees with ordinary Search for %q", row.ID)
		}
		previous = similarity
	}
	for _, id := range []string{"past100", "at120", "angled", "orthogonal", "opposite"} {
		if !seen[id] {
			t.Error("missing restricted contrast row", id)
		}
	}
}
