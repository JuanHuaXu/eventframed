package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func researchSortableReady(ctx context.Context, s *Store, name string, want int) (bool, error) {
	col, err := s.db.GetCollection(name)
	if err != nil {
		return false, err
	}
	if err := researchSortableColumnBindable(ctx, s, name); err != nil {
		return false, nil
	}
	records, err := col.ListAll(ctx)
	if err != nil {
		return false, err
	}
	if len(records) != want {
		return false, nil
	}
	for _, record := range records {
		event, err := decodeStoredEvent(record.Metadata, true)
		if err != nil || event.ID != record.ID || event.TenantID != "tenant-a" {
			return false, nil
		}
		key, ok := record.Metadata["available_at_sort"].(string)
		if !ok || key != researchAvailabilitySortKey(event.AvailableAt) {
			return false, nil
		}
	}
	return true, nil
}

func TestResearchSortableReadyRequiresDeclaredSchema(t *testing.T) {
	ctx := context.Background()
	s, err := Open(Config{Path: t.TempDir() + "/undeclared-ready.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	write := pinnedWrite("undeclared-ready", time.Date(2026, 10, 2, 0, 0, 0, 100000000, time.UTC))
	col, err := s.collection(ctx, write.Event.TenantID)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := encodeStoredEvent(write.Event)
	if err != nil {
		t.Fatal(err)
	}
	if err := col.Insert(ctx, write.Event.ID, write.Vector, map[string]interface{}{
		"event_json": string(payload), "corpus_text": write.Event.FrameText(),
		"available_at_sort": researchAvailabilitySortKey(write.Event.AvailableAt),
	}); err != nil {
		t.Fatal(err)
	}
	ready, err := researchSortableReady(ctx, s, collectionName(write.Event.TenantID, s.config.EmbeddingModel), 1)
	if err != nil || ready {
		t.Fatal("undeclared SQL key passed migration readiness", ready, err)
	}
}

func TestResearchSortKeyMigrationV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_KEY_MIGRATION_V1") != "1" {
		t.Skip("opt-in private legacy collection migration probe")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := Config{Path: t.TempDir() + "/migration.libravdb", Dimension: 4,
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
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writes := []ResearchEventWrite{
		pinnedWrite("past100", second.Add(100*time.Millisecond)),
		pinnedWrite("at120", second.Add(120*time.Millisecond)),
		pinnedWrite("future125", second.Add(125*time.Millisecond)),
	}
	writes[1].Vector, writes[2].Vector = []float32{0, 1, 0, 0}, []float32{0, 0, 1, 0}
	if _, err := s.PutResearchEventBatch(ctx, writes); err != nil {
		t.Fatal("legacy insert", err)
	}
	name := collectionName("tenant-a", "research:d4")
	col, err := s.db.GetCollection(name)
	if err != nil {
		t.Fatal(err)
	}
	type saved struct {
		payload string
		vector  []float32
		version uint64
	}
	beforeRows := make(map[string]saved, len(writes))
	for _, w := range writes {
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil {
			t.Fatal(err)
		}
		payload, _ := record.Metadata["event_json"].(string)
		if payload == "" {
			t.Fatal("legacy event payload absent", w.Event.ID)
		}
		beforeRows[w.Event.ID] = saved{payload, append([]float32(nil), record.Vector...), record.Version}
	}
	beforeSnapshot := s.Snapshot(ctx)
	beforeLSN, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	params := libra.QueryParams{
		"snapshot_lsn": int64(beforeLSN), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(second.Add(120 * time.Millisecond)),
	}
	if _, err := s.db.QueryWithParams(ctx, query, params); err == nil {
		t.Fatal("legacy collection unexpectedly bound undeclared sort key")
	}
	if _, err := s.db.Query(ctx, "ALTER TABLE "+name+" ADD COLUMN available_at_sort TEXT"); err != nil {
		t.Fatal("ALTER legacy collection", err)
	}
	ready, err := researchSortableReady(ctx, s, name, len(writes))
	if err != nil || ready {
		t.Fatal("unbackfilled collection marked ready", ready, err)
	}
	backfill := func(w ResearchEventWrite) {
		t.Helper()
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil {
			t.Fatal("read durable event for backfill", w.Event.ID, err)
		}
		event, err := decodeStoredEvent(record.Metadata, true)
		if err != nil || event.ID != w.Event.ID {
			t.Fatal("decode durable event for backfill", w.Event.ID, err)
		}
		if _, err := s.db.QueryWithParams(ctx,
			"UPDATE "+name+" SET available_at_sort = $key WHERE id = $id",
			libra.QueryParams{"key": researchAvailabilitySortKey(event.AvailableAt), "id": w.Event.ID}); err != nil {
			t.Fatal("backfill", w.Event.ID, err)
		}
	}
	backfill(writes[0])
	ready, err = researchSortableReady(ctx, s, name, len(writes))
	if err != nil || ready {
		t.Fatal("partly backfilled collection marked ready", ready, err)
	}
	for _, w := range writes[1:] {
		backfill(w)
	}
	ready, err = researchSortableReady(ctx, s, name, len(writes))
	if err != nil || !ready {
		t.Fatal("complete backfill not ready", ready, err)
	}
	checkQuery := func(db *libra.Database, stage string) {
		t.Helper()
		lsn, err := db.LatestCommitLSN(ctx)
		if err != nil {
			t.Fatal(stage, err)
		}
		params["snapshot_lsn"] = int64(lsn)
		rows, err := db.QueryWithParams(ctx, query, params)
		if err != nil {
			t.Fatal(stage, "SQL", err)
		}
		ids := make(map[string]bool)
		for _, row := range rows.Results {
			ids[row.ID] = true
		}
		if len(ids) != 2 || !ids["past100"] || !ids["at120"] || ids["future125"] {
			t.Fatal(stage, "as-of IDs", ids)
		}
		t.Logf("stage=%s LSN=%d IDs=%v", stage, lsn, ids)
	}
	checkQuery(s.db, "backfilled")
	if got := s.Snapshot(ctx); got != beforeSnapshot {
		t.Fatal("metadata backfill changed runtime snapshot", got, beforeSnapshot)
	}
	if !s.ResearchSnapshotCompatible(ctx, beforeSnapshot, second.Add(120*time.Millisecond)) {
		t.Fatal("metadata-only backfill unexpectedly invalidated runtime snapshot")
	}
	t.Log("runtime snapshot compatibility remained true across metadata backfill")
	for i, w := range writes {
		record, err := col.Get(ctx, w.Event.ID)
		if err != nil {
			t.Fatal(err)
		}
		old := beforeRows[w.Event.ID]
		if record.Metadata["event_json"] != old.payload || !reflect.DeepEqual(record.Vector, old.vector) ||
			!s.ingestMotion[beforeSnapshot.RuntimeVersion-uint64(len(writes)-i)+1].Equal(w.Event.AvailableAt) {
			t.Fatal("backfill changed event, vector or motion", w.Event.ID)
		}
		t.Logf("id=%s record_version=%d->%d", w.Event.ID, old.version, record.Version)
	}
	if err := s.Close(); err != nil {
		t.Fatal("close migrated store", err)
	}
	s = nil
	reopened, err := Open(config)
	if err != nil {
		t.Fatal("reopen migrated store", err)
	}
	defer reopened.Close()
	checkQuery(reopened.db, "reopened-before-ready")
	ready, err = researchSortableReady(ctx, reopened, name, len(writes))
	if err != nil || !ready {
		t.Fatal("reopened migration not ready", ready, err)
	}
	checkQuery(reopened.db, "reopened")
	if _, err := reopened.db.QueryWithParams(ctx,
		"UPDATE "+name+" SET available_at_sort = $key WHERE id = $id",
		libra.QueryParams{"key": "0000-01-01T00:00:00.000000000Z", "id": writes[0].Event.ID}); err != nil {
		t.Fatal("private corrupt-key control", err)
	}
	ready, err = researchSortableReady(ctx, reopened, name, len(writes))
	if err != nil || ready {
		t.Fatal(fmt.Sprintf("corrupt key was marked ready: ready=%v err=%v", ready, err))
	}
}
