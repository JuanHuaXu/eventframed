package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

// This boundary is deliberately test-only until old nonunit rows and the
// writer/query contract have a migration plan.
func normalizedVectorV4(vector []float32, dimension int) ([]float32, error) {
	if len(vector) != dimension {
		return nil, fmt.Errorf("vector dimension %d, want %d", len(vector), dimension)
	}
	var squared float64
	for _, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("nonfinite vector")
		}
		squared += float64(value) * float64(value)
	}
	if squared == 0 {
		return nil, fmt.Errorf("zero vector")
	}
	norm := math.Sqrt(squared)
	unit := make([]float32, dimension)
	for i, value := range vector {
		unit[i] = float32(float64(value) / norm)
	}
	if !researchUnitVectorV3(unit) {
		return nil, fmt.Errorf("rounded vector is not unit length")
	}
	return unit, nil
}

func appendNormalizedV4(ctx context.Context, g *incrementalSortGate, writes []ResearchEventWrite) error {
	converted := make([]ResearchEventWrite, len(writes))
	for i, write := range writes {
		unit, err := normalizedVectorV4(write.Vector, 4)
		if err != nil {
			return err
		}
		converted[i] = write
		converted[i].Vector = unit
	}
	return g.appendBatch(ctx, converted, "")
}

func queryNormalizedV4(ctx context.Context, g *incrementalSortGate, lsn uint64, asOf time.Time, rawQuery []float32) ([]store.SearchResult, error) {
	unit, err := normalizedVectorV4(rawQuery, 4)
	if err != nil {
		return nil, err
	}
	lease, err := g.store.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	query := "SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <=> $query_vec AS distance FROM " + g.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	rows, err := g.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(lsn), "query_vec": unit,
		"available_by_sort": researchAvailabilitySortKey(asOf),
	})
	if err != nil {
		return nil, err
	}
	results := make([]store.SearchResult, 0, len(rows.Results))
	for _, row := range rows.Results {
		event, err := decodeStoredEvent(row.Metadata, true)
		if err != nil {
			return nil, err
		}
		if event.ID != row.ID || event.TenantID != "tenant-a" || event.AvailableAt.After(asOf) ||
			row.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) {
			return nil, fmt.Errorf("published row identity or availability mismatch: %s", row.ID)
		}
		distance, ok := publishedScoreNumberV1(row.Metadata["distance"])
		if !ok || math.IsNaN(distance) || math.IsInf(distance, 0) {
			return nil, fmt.Errorf("invalid published distance: %s", row.ID)
		}
		results = append(results, store.SearchResult{Event: event, Similarity: math.Max(-1, math.Min(1, 1-distance))})
	}
	return results, nil
}

func compareNormalizedV4(t *testing.T, g *incrementalSortGate, lsn uint64, asOf time.Time, query []float32, wantIDs []string, compareOrdinary bool) {
	t.Helper()
	ctx := context.Background()
	got, err := queryNormalizedV4(ctx, g, lsn, asOf, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(wantIDs) {
		t.Fatalf("wrong as-of result count: got %d want %d", len(got), len(wantIDs))
	}
	want := make(map[string]bool, len(wantIDs))
	for _, id := range wantIDs {
		want[id] = true
	}
	var ordinaryScores map[string]float64
	if compareOrdinary {
		ordinary, err := g.store.Search(ctx, "tenant-a", query, asOf, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(ordinary) != len(got) {
			t.Fatal("published and ordinary Search result count differs", len(got), len(ordinary))
		}
		ordinaryScores = make(map[string]float64, len(ordinary))
		for _, item := range ordinary {
			ordinaryScores[item.Event.ID] = item.Similarity
		}
	}
	seen := make(map[string]bool, len(got))
	previous := math.Inf(1)
	for _, item := range got {
		id := item.Event.ID
		if seen[id] || !want[id] || item.Similarity > previous+1e-5 {
			t.Errorf("unexpected, duplicated, or misordered published row %q", id)
		}
		seen[id] = true
		previous = item.Similarity
		if compareOrdinary {
			score, ok := ordinaryScores[id]
			if !ok || math.Abs(score-item.Similarity) > 1e-5 {
				t.Errorf("score mismatch %s: published=%g ordinary=%g", id, item.Similarity, score)
			}
			t.Logf("id=%s published_similarity=%g ordinary_similarity=%g", id, item.Similarity, score)
		}
	}
	for _, id := range wantIDs {
		if !seen[id] {
			t.Error("missing published row", id)
		}
	}
}

func TestResearchPublishedNormalizedBoundaryV4(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_NORMALIZED_BOUNDARY_V4") != "1" {
		t.Skip("opt-in normalized certified-LSN search boundary probe")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	rawQuery := []float32{2, 0, 0, 0}
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
	invalid := [][]float32{nil, {0, 0, 0, 0}, {1, 0, 0}, {float32(math.NaN()), 0, 0, 0}, {float32(math.Inf(1)), 0, 0, 0}}
	for _, vector := range invalid {
		if _, err := normalizedVectorV4(vector, 4); err == nil {
			t.Error("invalid vector passed", vector)
		}
		if _, err := queryNormalizedV4(ctx, g, oldLSN, asOf, vector); err == nil {
			t.Error("invalid query passed", vector)
		}
		bad := pinnedWrite("invalid", second.Add(110*time.Millisecond))
		bad.Vector = vector
		if err := appendNormalizedV4(ctx, g, []ResearchEventWrite{bad}); err == nil {
			t.Error("invalid write passed", vector)
		}
	}
	mixedGood := pinnedWrite("mixed-good", second.Add(109*time.Millisecond))
	mixedGood.Vector = []float32{2, 0, 0, 0}
	mixedBad := pinnedWrite("mixed-bad", second.Add(110*time.Millisecond))
	mixedBad.Vector = []float32{0, 0, 0, 0}
	if err := appendNormalizedV4(ctx, g, []ResearchEventWrite{mixedGood, mixedBad}); err == nil {
		t.Error("mixed valid/invalid batch passed pre-write normalization")
	}
	if _, err := g.store.GetEvents(ctx, "tenant-a", []string{"mixed-good"}, asOf); !errors.Is(err, store.ErrEventNotFound) {
		t.Fatal("mixed valid/invalid batch did not leave first event absent", err)
	}
	if current := r.current.Load().lsn; current != oldLSN {
		t.Fatal("invalid inputs moved the certified LSN", current, oldLSN)
	}
	angled := pinnedWrite("angled", second.Add(110*time.Millisecond))
	angled.Vector = []float32{3, 4, 0, 0}
	orthogonal := pinnedWrite("orthogonal", second.Add(111*time.Millisecond))
	orthogonal.Vector = []float32{0, 2, 0, 0}
	opposite := pinnedWrite("opposite", second.Add(112*time.Millisecond))
	opposite.Vector = []float32{-5, 0, 0, 0}
	if err := appendNormalizedV4(ctx, g, []ResearchEventWrite{angled, orthogonal, opposite}); err != nil {
		t.Fatal(err)
	}
	if err := r.publish(ctx); err != nil {
		t.Fatal(err)
	}
	newLSN := r.current.Load().lsn
	if newLSN <= oldLSN {
		t.Fatal("certified LSN did not advance", oldLSN, newLSN)
	}
	compareNormalizedV4(t, g, oldLSN, asOf, rawQuery, []string{"past100", "at120"}, false)
	compareNormalizedV4(t, g, newLSN, asOf, rawQuery, []string{"past100", "at120", "angled", "orthogonal", "opposite"}, true)
	stored, err := g.store.GetEventsWithVectors(ctx, "tenant-a", []string{"angled", "orthogonal", "opposite"}, asOf)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range stored {
		if !researchUnitVectorV3(event.Embedding) {
			t.Error("durable vector is not unit length", event.ID)
		}
	}
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	g = nil
	g, err = openIncrementalSortGate(root)
	if err != nil {
		t.Fatal(err)
	}
	if lsn, ok := g.capture(ctx); !ok || lsn != newLSN {
		t.Fatal("reopen lost certified LSN", lsn, ok, newLSN)
	}
	compareNormalizedV4(t, g, newLSN, asOf, rawQuery, []string{"past100", "at120", "angled", "orthogonal", "opposite"}, true)
}
