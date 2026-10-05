package libravdbstore

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func queryPublishedTopkV5(ctx context.Context, g *incrementalSortGate, lsn uint64, asOf time.Time, k int) ([]store.SearchResult, error) {
	unit, err := normalizedVectorV4([]float32{2, 0, 0, 0}, 4)
	if err != nil {
		return nil, err
	}
	lease, err := g.store.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	query := fmt.Sprintf("SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <=> $query_vec AS distance FROM %s AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT %d", g.name, k)
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
		if err != nil || event.ID != row.ID || event.TenantID != "tenant-a" || event.AvailableAt.After(asOf) ||
			row.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) {
			return nil, fmt.Errorf("bad published payload %s: %w", row.ID, err)
		}
		distance, ok := publishedScoreNumberV1(row.Metadata["distance"])
		if !ok || math.IsNaN(distance) || math.IsInf(distance, 0) {
			return nil, fmt.Errorf("invalid published distance for %s", row.ID)
		}
		results = append(results, store.SearchResult{Event: event, Similarity: math.Max(-1, math.Min(1, 1-distance))})
	}
	return results, nil
}

func checkPublishedTopkV5(t *testing.T, k int, published, ordinary []store.SearchResult) {
	t.Helper()
	if len(published) != k || len(ordinary) != k {
		t.Errorf("k=%d result count published=%d ordinary=%d", k, len(published), len(ordinary))
	}
	want := make(map[string]float64, len(ordinary))
	for _, item := range ordinary {
		if strings.HasPrefix(item.Event.ID, "future") {
			t.Errorf("k=%d ordinary Search included future %s", k, item.Event.ID)
		}
		if _, exists := want[item.Event.ID]; exists {
			t.Errorf("k=%d ordinary Search duplicated %s", k, item.Event.ID)
		}
		want[item.Event.ID] = item.Similarity
	}
	seen := make(map[string]bool, len(published))
	previous := math.Inf(1)
	for _, item := range published {
		id := item.Event.ID
		if seen[id] || strings.HasPrefix(id, "future") || item.Similarity > previous+1e-5 {
			t.Errorf("k=%d duplicate/future/misordered published row %s", k, id)
		}
		seen[id] = true
		previous = item.Similarity
		score, exists := want[id]
		if !exists || math.Abs(score-item.Similarity) > 1e-5 {
			t.Errorf("k=%d published row %s missing or score mismatch: published=%g ordinary=%g", k, id, item.Similarity, score)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("k=%d ordinary row %s missing from published results", k, id)
		}
	}
}

func TestResearchPublishedTopkV5(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_TOPK_V5") != "1" {
		t.Skip("opt-in certified-LSN bounded top-k transfer screen")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	g := createPublishedPayloadGateV1(t, t.TempDir())
	defer g.close()
	appendAngles := func(prefix string, count int, baseAngle float64, at time.Time) {
		t.Helper()
		for start := 0; start < count; start += 16 {
			batch := make([]ResearchEventWrite, 0, min(16, count-start))
			for i := start; i < min(count, start+16); i++ {
				angle := baseAngle * float64(i+1)
				write := pinnedWrite(fmt.Sprintf("%s-%03d", prefix, i), at)
				write.Vector = []float32{float32(2 * math.Cos(angle)), float32(2 * math.Sin(angle)), 0, 0}
				batch = append(batch, write)
			}
			if err := appendNormalizedV4(ctx, g, batch); err != nil {
				t.Fatal(err)
			}
		}
	}
	appendAngles("eligible", 198, 0.005, second.Add(100*time.Millisecond))
	appendAngles("future", 16, 0.001, asOf.Add(5*time.Millisecond))
	r, err := newSortPublishedReader(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	lsn := r.current.Load().lsn
	query := []float32{2, 0, 0, 0}
	for _, k := range []int{10, 50, 200} {
		t.Run(fmt.Sprintf("k%d", k), func(t *testing.T) {
			for warm := 0; warm < 4; warm++ {
				published, err := queryPublishedTopkV5(ctx, g, lsn, asOf, k)
				if err != nil {
					t.Fatal(err)
				}
				ordinary, err := g.store.Search(ctx, "tenant-a", query, asOf, k)
				if err != nil {
					t.Fatal(err)
				}
				checkPublishedTopkV5(t, k, published, ordinary)
			}
			publishedNS := make([]int64, 0, 32)
			ordinaryNS := make([]int64, 0, 32)
			for trial := 0; trial < 32; trial++ {
				publishedCall := func() []store.SearchResult {
					start := time.Now()
					results, err := queryPublishedTopkV5(ctx, g, lsn, asOf, k)
					publishedNS = append(publishedNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					return results
				}
				ordinaryCall := func() []store.SearchResult {
					start := time.Now()
					results, err := g.store.Search(ctx, "tenant-a", query, asOf, k)
					ordinaryNS = append(ordinaryNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					return results
				}
				var published, ordinary []store.SearchResult
				if trial%2 == 0 {
					published, ordinary = publishedCall(), ordinaryCall()
				} else {
					ordinary, published = ordinaryCall(), publishedCall()
				}
				checkPublishedTopkV5(t, k, published, ordinary)
			}
			t.Logf("k=%d published_p50=%s published_p99=%s ordinary_p50=%s ordinary_p99=%s", k,
				pinnedPercentile(publishedNS, .50), pinnedPercentile(publishedNS, .99),
				pinnedPercentile(ordinaryNS, .50), pinnedPercentile(ordinaryNS, .99))
		})
	}
}
