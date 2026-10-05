package libravdbstore

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func denseQueryV6() []float32 {
	query := make([]float32, 256)
	for j := range query {
		x := float64(j + 1)
		query[j] = float32(math.Sin(0.17*x) + 0.3*math.Cos(0.071*x))
	}
	unit, err := normalizedVectorV4(query, 256)
	if err != nil {
		panic(err)
	}
	return unit
}

func denseRowV6(query []float32, i int, angle float64) []float32 {
	noise := make([]float64, 256)
	var dot float64
	for j := range noise {
		x := float64(j + 1)
		noise[j] = math.Sin(0.113*float64(i+1)*x) + math.Cos(0.047*float64(i+3)*x)
		dot += noise[j] * float64(query[j])
	}
	var squared float64
	for j := range noise {
		noise[j] -= dot * float64(query[j])
		squared += noise[j] * noise[j]
	}
	result := make([]float32, 256)
	for j := range result {
		result[j] = float32(2 * (math.Cos(angle)*float64(query[j]) + math.Sin(angle)*noise[j]/math.Sqrt(squared)))
	}
	return result
}

func createDensePublishedGateV6(t *testing.T, root string, query []float32) *incrementalSortGate {
	t.Helper()
	ctx := context.Background()
	s, err := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 256,
		EmbeddingModel: "research:d256", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "sort-publication.sqlite")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		s.Close()
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000"} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			s.Close()
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE TABLE marker(
		id INTEGER PRIMARY KEY CHECK(id=1), phase TEXT NOT NULL,
		lsn INTEGER NOT NULL, row_count INTEGER NOT NULL, digest TEXT NOT NULL);
		INSERT INTO marker(id,phase,lsn,row_count,digest) VALUES(1,'PENDING',0,0,'')`); err != nil {
		db.Close()
		s.Close()
		t.Fatal(err)
	}
	g := &sortPublicationGate{store: s, sidecar: db,
		name: collectionName("tenant-a", "research:d256"), wantRows: 3}
	_, err = s.db.CreateCollection(ctx, g.name, libra.WithDimension(256),
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
	for i := range writes {
		raw := denseRowV6(query, i, 0)
		writes[i].Vector, err = normalizedVectorV4(raw, 256)
		if err != nil {
			g.close()
			t.Fatal(err)
		}
	}
	if _, lsn, err := s.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || lsn == 0 {
		g.close()
		t.Fatal("dense genesis commit", lsn, err)
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

func appendDenseV6(ctx context.Context, gate *incrementalSortGate, writes []ResearchEventWrite) error {
	converted := make([]ResearchEventWrite, len(writes))
	for i, write := range writes {
		unit, err := normalizedVectorV4(write.Vector, 256)
		if err != nil {
			return err
		}
		converted[i] = write
		converted[i].Vector = unit
	}
	return gate.appendBatch(ctx, converted, "")
}

func queryDensePublishedV6(ctx context.Context, gate *incrementalSortGate, lsn uint64, asOf time.Time, rawQuery []float32, k int) ([]store.SearchResult, error) {
	unit, err := normalizedVectorV4(rawQuery, 256)
	if err != nil {
		return nil, err
	}
	lease, err := gate.store.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	query := fmt.Sprintf("SELECT id,event_json,corpus_text,raw_content,available_at_sort,embedding <=> $query_vec AS distance FROM %s AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT %d", gate.name, k)
	rows, err := gate.store.db.QueryWithParams(ctx, query, libra.QueryParams{
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
			return nil, fmt.Errorf("invalid dense published payload %s: %w", row.ID, err)
		}
		distance, ok := publishedScoreNumberV1(row.Metadata["distance"])
		if !ok || math.IsNaN(distance) || math.IsInf(distance, 0) {
			return nil, fmt.Errorf("invalid dense published distance for %s", row.ID)
		}
		results = append(results, store.SearchResult{Event: event, Similarity: math.Max(-1, math.Min(1, 1-distance))})
	}
	return results, nil
}

func TestResearchPublishedDenseV6(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_DENSE_V6") != "1" {
		t.Skip("opt-in dense 256D certified-LSN top-k transfer test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createDensePublishedGateV6(t, t.TempDir(), query)
	defer gate.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	appendAngles := func(prefix string, count int, multiplier float64, at time.Time) {
		t.Helper()
		for start := 0; start < count; start += 16 {
			batch := make([]ResearchEventWrite, 0, min(16, count-start))
			for i := start; i < min(count, start+16); i++ {
				write := pinnedWrite(fmt.Sprintf("%s-%03d", prefix, i), at)
				write.Vector = denseRowV6(query, i, multiplier*float64(i+1))
				batch = append(batch, write)
			}
			if err := appendDenseV6(ctx, gate, batch); err != nil {
				t.Fatal(err)
			}
		}
	}
	appendAngles("eligible", 198, 0.005, second.Add(100*time.Millisecond))
	appendAngles("future", 16, 0.001, asOf.Add(5*time.Millisecond))
	reader, err := newSortPublishedReader(ctx, gate)
	if err != nil {
		t.Fatal(err)
	}
	lsn := reader.current.Load().lsn
	for _, k := range []int{10, 50, 200} {
		t.Run(fmt.Sprintf("k%d", k), func(t *testing.T) {
			for warm := 0; warm < 4; warm++ {
				published, err := queryDensePublishedV6(ctx, gate, lsn, asOf, rawQuery, k)
				if err != nil {
					t.Fatal(err)
				}
				ordinary, err := gate.store.Search(ctx, "tenant-a", rawQuery, asOf, k)
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
					items, err := queryDensePublishedV6(ctx, gate, lsn, asOf, rawQuery, k)
					publishedNS = append(publishedNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					return items
				}
				ordinaryCall := func() []store.SearchResult {
					start := time.Now()
					items, err := gate.store.Search(ctx, "tenant-a", rawQuery, asOf, k)
					ordinaryNS = append(ordinaryNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					return items
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
