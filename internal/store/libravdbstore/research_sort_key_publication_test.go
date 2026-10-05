package libravdbstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
	_ "modernc.org/sqlite"
)

type sortPublicationMarker struct {
	phase  string
	lsn    uint64
	count  int
	digest string
}

type sortPublicationGate struct {
	store       *Store
	sidecar     *sql.DB
	name        string
	wantRows    int
	mu          sync.RWMutex
	verified    string
	verifiedLSN uint64
}

func openSortPublicationGate(root string, create bool) (_ *sortPublicationGate, err error) {
	path := filepath.Join(root, "sort-publication.sqlite")
	if !create {
		if _, statErr := os.Stat(path); statErr != nil {
			return nil, statErr
		}
	}
	store, err := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			store.Close()
		}
	}()
	if create {
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if createErr != nil {
			return nil, createErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return nil, closeErr
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000"} {
		if _, err := db.Exec(pragma); err != nil {
			return nil, err
		}
	}
	if create {
		if _, err := db.Exec(`CREATE TABLE marker(
			id INTEGER PRIMARY KEY CHECK(id=1), phase TEXT NOT NULL,
			lsn INTEGER NOT NULL, row_count INTEGER NOT NULL, digest TEXT NOT NULL);
			INSERT INTO marker(id,phase,lsn,row_count,digest) VALUES(1,'PENDING',0,0,'')`); err != nil {
			return nil, err
		}
	}
	g := &sortPublicationGate{store: store, sidecar: db,
		name: collectionName("tenant-a", "research:d4"), wantRows: 3}
	marker, err := g.readMarker(context.Background())
	if err != nil {
		return nil, err
	}
	if marker.phase == "READY" && marker.count == g.wantRows {
		lsn, lsnErr := g.store.db.LatestCommitLSN(context.Background())
		if lsnErr != nil {
			return nil, lsnErr
		}
		if lsn == marker.lsn {
			digest, scanErr := g.scan(context.Background())
			if scanErr == nil && digest == marker.digest {
				g.verified = digest
				g.verifiedLSN = lsn
			}
		}
	}
	return g, nil
}

func (g *sortPublicationGate) close() error {
	return errors.Join(g.sidecar.Close(), g.store.Close())
}

func (g *sortPublicationGate) readMarker(ctx context.Context) (sortPublicationMarker, error) {
	var marker sortPublicationMarker
	var lsn int64
	err := g.sidecar.QueryRowContext(ctx,
		"SELECT phase,lsn,row_count,digest FROM marker WHERE id=1").Scan(&marker.phase, &lsn, &marker.count, &marker.digest)
	if err != nil {
		return sortPublicationMarker{}, fmt.Errorf("read durable migration marker: %w", err)
	}
	if lsn < 0 {
		return sortPublicationMarker{}, fmt.Errorf("negative migration marker LSN: %d", lsn)
	}
	marker.lsn = uint64(lsn)
	return marker, nil
}

func (g *sortPublicationGate) setMarker(ctx context.Context, marker sortPublicationMarker) error {
	result, err := g.sidecar.ExecContext(ctx,
		"UPDATE marker SET phase=?,lsn=?,row_count=?,digest=? WHERE id=1",
		marker.phase, int64(marker.lsn), marker.count, marker.digest)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("migration marker update affected rows: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("migration marker update affected %d rows", count)
	}
	return nil
}

func (g *sortPublicationGate) begin(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.setMarker(ctx, sortPublicationMarker{phase: "PENDING"}); err != nil {
		return err
	}
	g.verified = ""
	g.verifiedLSN = 0
	return nil
}

func (g *sortPublicationGate) scan(ctx context.Context) (string, error) {
	ready, err := researchSortableReady(ctx, g.store, g.name, g.wantRows)
	if err != nil {
		return "", fmt.Errorf("collection is not fully sortable: %w", err)
	}
	if !ready {
		return "", errors.New("collection is not fully sortable")
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		return "", err
	}
	records, err := col.ListAll(ctx)
	if err != nil {
		return "", err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	type digestRow struct {
		ID, Payload, SortKey string
		Vector               []float32
	}
	rows := make([]digestRow, 0, len(records))
	for _, record := range records {
		payload, payloadOK := record.Metadata["event_json"].(string)
		key, keyOK := record.Metadata["available_at_sort"].(string)
		if !payloadOK || !keyOK || payload == "" {
			return "", fmt.Errorf("incomplete row %q", record.ID)
		}
		rows = append(rows, digestRow{record.ID, payload, key, record.Vector})
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func (g *sortPublicationGate) publish(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	before, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil {
		return err
	}
	digest, err := g.scan(ctx)
	if err != nil {
		return err
	}
	after, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil {
		return fmt.Errorf("LibraVDB moved during readiness scan: %w", err)
	}
	if before != after {
		return fmt.Errorf("LibraVDB moved during readiness scan: %d -> %d", before, after)
	}
	marker := sortPublicationMarker{phase: "READY", lsn: after, count: g.wantRows, digest: digest}
	if err := g.setMarker(ctx, marker); err != nil {
		return err
	}
	g.verified = digest
	g.verifiedLSN = after
	return nil
}

func (g *sortPublicationGate) capture(ctx context.Context) (uint64, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.verified == "" {
		return 0, false
	}
	first, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil {
		return 0, false
	}
	marker, err := g.readMarker(ctx)
	if err != nil {
		return 0, false
	}
	second, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || first != second || marker.phase != "READY" || marker.lsn != first ||
		g.verifiedLSN != first ||
		marker.count != g.wantRows || marker.digest != g.verified {
		return 0, false
	}
	return first, true
}

func (g *sortPublicationGate) search(ctx context.Context, at time.Time) ([]string, error) {
	lsn, ok := g.capture(ctx)
	if !ok {
		return nil, errors.New("sort-key migration is not published")
	}
	lease, err := g.store.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + g.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	rows, err := g.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(at),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(rows.Results))
	for i, row := range rows.Results {
		ids[i] = row.ID
	}
	return ids, nil
}

func TestResearchSortKeyPublicationV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_KEY_PUBLICATION_V1") != "1" {
		t.Skip("opt-in private sort-key publication state-machine test")
	}
	ctx := context.Background()
	root := t.TempDir()
	g, err := openSortPublicationGate(root, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if g != nil {
			g.close()
		}
	}()
	reopen := func() {
		t.Helper()
		if err := g.close(); err != nil {
			t.Fatal("close phase", err)
		}
		g, err = openSortPublicationGate(root, false)
		if err != nil {
			t.Fatal("reopen phase", err)
		}
	}
	assertDenied := func(stage string) {
		t.Helper()
		if _, ok := g.capture(ctx); ok {
			t.Fatal(stage, "published an unsafe collection")
		}
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writes := []ResearchEventWrite{
		pinnedWrite("past100", second.Add(100*time.Millisecond)),
		pinnedWrite("at120", second.Add(120*time.Millisecond)),
		pinnedWrite("future125", second.Add(125*time.Millisecond)),
	}
	if _, err := g.store.PutResearchEventBatch(ctx, writes); err != nil {
		t.Fatal("legacy insert", err)
	}
	if err := g.begin(ctx); err != nil {
		t.Fatal("begin migration", err)
	}
	reopen()
	assertDenied("pending before ALTER")
	if _, err := g.store.db.Query(ctx, "ALTER TABLE "+g.name+" ADD COLUMN available_at_sort TEXT"); err != nil {
		t.Fatal("ALTER", err)
	}
	backfill := func(w ResearchEventWrite) {
		t.Helper()
		col, err := g.store.db.GetCollection(g.name)
		if err != nil {
			t.Fatal(err)
		}
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil {
			t.Fatal(err)
		}
		event, err := decodeStoredEvent(record.Metadata, true)
		if err != nil || event.ID != w.Event.ID {
			t.Fatal("decode durable backfill source", err)
		}
		if _, err := g.store.db.QueryWithParams(ctx,
			"UPDATE "+g.name+" SET available_at_sort = $key WHERE id = $id",
			libra.QueryParams{"key": researchAvailabilitySortKey(event.AvailableAt), "id": w.Event.ID}); err != nil {
			t.Fatal("backfill", w.Event.ID, err)
		}
	}
	backfill(writes[0])
	reopen()
	assertDenied("partial backfill")
	for _, w := range writes[1:] {
		backfill(w)
	}
	if _, err := g.scan(ctx); err != nil {
		t.Fatal("complete prepublication scan", err)
	}
	reopen()
	assertDenied("complete but unpublished")
	if err := g.publish(ctx); err != nil {
		t.Fatal("publish READY", err)
	}
	ids, err := g.search(ctx, second.Add(120*time.Millisecond))
	if err != nil || len(ids) != 2 || ids[0] == "future125" || ids[1] == "future125" {
		t.Fatal("published as-of result", ids, err)
	}
	seen := map[string]bool{ids[0]: true, ids[1]: true}
	if !seen["past100"] || !seen["at120"] {
		t.Fatal("published as-of missed past rows", ids)
	}
	reopen()
	if _, ok := g.capture(ctx); !ok {
		t.Fatal("READY did not survive reopen")
	}
	readyMarker, err := g.readMarker(ctx)
	if err != nil {
		t.Fatal("read READY marker", err)
	}
	badDigest := readyMarker
	badDigest.digest = "stale-digest"
	if err := g.setMarker(ctx, badDigest); err != nil {
		t.Fatal("private digest-mismatch control", err)
	}
	assertDenied("digest-mismatched READY in current process")
	reopen()
	assertDenied("digest-mismatched READY after reopen")
	if err := g.setMarker(ctx, readyMarker); err != nil {
		t.Fatal("restore private READY marker", err)
	}
	reopen()
	if _, ok := g.capture(ctx); !ok {
		t.Fatal("verified READY did not recover after valid marker restore")
	}
	if _, err := openSortPublicationGate(filepath.Join(t.TempDir(), "missing"), false); err == nil {
		t.Fatal("missing sidecar was accepted")
	}
	latencies := make([]int64, 0, 1000)
	for i := 0; i < 1000; i++ {
		start := time.Now()
		if _, ok := g.capture(ctx); !ok {
			t.Fatal("quiet READY gate rejected")
		}
		latencies = append(latencies, time.Since(start).Nanoseconds())
	}
	t.Logf("ready gate p50=%s p99=%s", pinnedPercentile(latencies, .5), pinnedPercentile(latencies, .99))
	if pinnedPercentile(latencies, .99) > time.Millisecond {
		t.Error("ready gate exceeded frozen isolated 1 ms p99")
	}
	if _, err := g.store.PutResearchEventBatch(ctx,
		[]ResearchEventWrite{pinnedWrite("legacy-new", second.Add(110*time.Millisecond))}); err != nil {
		t.Fatal("legacy post-READY write", err)
	}
	assertDenied("post-READY unauthenticated write")
	marker, err := g.readMarker(ctx)
	if err != nil {
		t.Fatal("read published marker", err)
	}
	latest, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal("read post-write LSN", err)
	}
	marker.lsn = latest
	if err := g.setMarker(ctx, marker); err != nil {
		t.Fatal("private stale-digest marker control", err)
	}
	assertDenied("new LSN with digest verified at old LSN")
	reopen()
	assertDenied("new LSN with stale digest after reopen")
	if err := g.begin(ctx); err != nil {
		t.Fatal("restart migration", err)
	}
	if err := g.publish(ctx); err == nil {
		t.Fatal("unkeyed legacy row was republished")
	}
	reopen()
	assertDenied("unkeyed legacy row after reopen")
}
