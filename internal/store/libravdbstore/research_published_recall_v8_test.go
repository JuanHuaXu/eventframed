package libravdbstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func createUnpublishedDenseGateV8(t *testing.T, root string, query []float32) *incrementalSortGate {
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
	return &incrementalSortGate{g}
}

type publishedRecallViewV8 struct {
	lsn      uint64
	snapshot model.Snapshot
	at       time.Time
}

type publishedRecallPinV8 struct{ view *publishedRecallViewV8 }
type publishedRecallPinKeyV8 struct{}

func withPublishedRecallPinV8(ctx context.Context) context.Context {
	return context.WithValue(ctx, publishedRecallPinKeyV8{}, &publishedRecallPinV8{})
}

type publishedRecallStoreV8 struct {
	store.EventStore
	gate        *incrementalSortGate
	owner       sync.Mutex
	current     atomic.Pointer[publishedRecallViewV8]
	searches    int
	staleReject int
	afterSearch func() error
}

func (s *publishedRecallStoreV8) publishLocked(ctx context.Context) error {
	lsn, ready := s.gate.capture(ctx)
	if !ready {
		s.current.Store(nil)
		return fmt.Errorf("journaled EventFrame view is not READY")
	}
	s.current.Store(&publishedRecallViewV8{lsn: lsn, snapshot: s.gate.store.Snapshot(ctx), at: time.Now()})
	return nil
}

func (s *publishedRecallStoreV8) appendEvent(ctx context.Context, write ResearchEventWrite) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	if err := appendDenseV6(ctx, s.gate, []ResearchEventWrite{write}); err != nil {
		s.current.Store(nil)
		return err
	}
	return s.publishLocked(ctx)
}

func (s *publishedRecallStoreV8) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	if tenantID != "tenant-a" {
		return nil, fmt.Errorf("unexpected tenant %q", tenantID)
	}
	pin, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if pin == nil {
		return nil, fmt.Errorf("Recall has no per-call published view pin")
	}
	view := s.current.Load()
	if view == nil || time.Since(view.at) >= 250*time.Millisecond {
		return nil, fmt.Errorf("published view unavailable or expired")
	}
	results, err := queryDensePublishedV6(ctx, s.gate, view.lsn, availableBy, vector, limit)
	if err != nil {
		return nil, err
	}
	if time.Since(view.at) >= 250*time.Millisecond || s.current.Load() == nil {
		return nil, fmt.Errorf("published view expired during Search")
	}
	pin.view = view
	s.searches++
	if s.afterSearch != nil {
		after := s.afterSearch
		s.afterSearch = nil
		if err := after(); err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (s *publishedRecallStoreV8) Snapshot(ctx context.Context) model.Snapshot {
	if pin, ok := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8); ok && pin.view != nil {
		return pin.view.snapshot
	}
	return s.gate.store.Snapshot(ctx)
}

func (s *publishedRecallStoreV8) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	current := s.gate.store.Snapshot(ctx)
	if !store.JournalSnapshotCompatible(entry.Snapshot, current, entry.AsOf, s.gate.store.ingestMotion) {
		s.staleReject++
		return store.ErrStaleSnapshot
	}
	if err := s.gate.appendJournal(ctx, entry, ""); err != nil {
		s.current.Store(nil)
		return err
	}
	return s.publishLocked(ctx)
}

func (s *publishedRecallStoreV8) Close() error { return nil }

func checkPublishedRecallV8(t *testing.T, ctx context.Context, gate *incrementalSortGate, packet model.ContextPacket, wantSnapshot model.Snapshot, wantIDs []string) {
	t.Helper()
	journal, err := gate.store.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Snapshot != wantSnapshot || journal.Snapshot != packet.Snapshot ||
		packet.Recalled != len(wantIDs) || len(journal.Report.Decisions) != len(wantIDs) {
		t.Fatalf("packet/journal snapshot or nominated count mismatch: packet=%+v journal=%+v", packet.Snapshot, journal.Snapshot)
	}
	want := make(map[string]bool, len(wantIDs))
	for _, id := range wantIDs {
		want[id] = true
	}
	for _, decision := range journal.Report.Decisions {
		if !want[decision.EventID] {
			t.Errorf("unexpected or future nominated event %q", decision.EventID)
		}
		delete(want, decision.EventID)
	}
	for id := range want {
		t.Errorf("missing nominated event %q", id)
	}
	if _, ready := gate.capture(ctx); !ready {
		t.Error("authorized Recall journal left publication gate unavailable")
	}
}

func TestResearchPublishedRecallV8(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_RECALL_V8") != "1" {
		t.Skip("opt-in published-LSN full Service Recall interleaving probe")
	}
	ctx := context.Background()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createUnpublishedDenseGateV8(t, t.TempDir(), query)
	defer gate.close()
	adapter := &publishedRecallStoreV8{EventStore: gate.store, gate: gate}
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{DefaultRecallK: 3, DefaultPackK: 3, DefaultTokenBudget: 10000})
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()
	if err := gate.publish(ctx); err != nil {
		t.Fatal("publish after service policy bind", err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal("initialize journal after service policy bind", err)
	}
	if err := adapter.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "private-lab",
		Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
		AsOf: asOf, RecallK: 3, PackK: 3, TokenBudget: 10000}
	beforeVisible := adapter.current.Load().snapshot
	visible := pinnedWrite("visible110", second.Add(110*time.Millisecond))
	visible.Vector = denseRowV6(query, 7, 0.01)
	adapter.afterSearch = func() error { return adapter.appendEvent(ctx, visible) }
	packet, err := svc.Recall(withPublishedRecallPinV8(ctx), request)
	if err != nil {
		t.Fatal("visible interleave Recall", err)
	}
	afterVisible := adapter.current.Load().snapshot
	if adapter.searches != 2 || adapter.staleReject != 1 || afterVisible.RuntimeVersion <= beforeVisible.RuntimeVersion {
		t.Fatal("visible interleave did not reject old pin and retry", adapter.searches, adapter.staleReject, beforeVisible, afterVisible)
	}
	checkPublishedRecallV8(t, ctx, gate, packet, afterVisible, []string{"past100", "at120", "visible110"})
	t.Logf("visible_searches=%d stale_rejects=%d packet_version=%d", adapter.searches, adapter.staleReject, packet.Snapshot.RuntimeVersion)
	beforeFuture := adapter.current.Load().snapshot
	future := pinnedWrite("future130", asOf.Add(5*time.Millisecond))
	future.Vector = denseRowV6(query, 8, 0.001)
	adapter.afterSearch = func() error { return adapter.appendEvent(ctx, future) }
	packet, err = svc.Recall(withPublishedRecallPinV8(ctx), request)
	if err != nil {
		t.Fatal("future-only interleave Recall", err)
	}
	if adapter.searches != 3 || adapter.staleReject != 1 ||
		adapter.current.Load().snapshot.RuntimeVersion != beforeFuture.RuntimeVersion+1 {
		t.Fatal("future-only write incorrectly forced retry or failed to publish", adapter.searches, adapter.staleReject)
	}
	checkPublishedRecallV8(t, ctx, gate, packet, beforeFuture, []string{"past100", "at120", "visible110"})
	t.Logf("future_searches=%d stale_rejects=%d packet_version=%d current_version=%d", adapter.searches, adapter.staleReject,
		packet.Snapshot.RuntimeVersion, adapter.current.Load().snapshot.RuntimeVersion)
	packet, err = svc.Recall(withPublishedRecallPinV8(ctx), request)
	if err != nil {
		t.Fatal("new-view Recall after future-only write", err)
	}
	checkPublishedRecallV8(t, ctx, gate, packet, adapter.current.Load().snapshot, []string{"past100", "at120", "visible110"})
	if adapter.searches != 4 {
		t.Fatal("unexpected later Recall search count", adapter.searches)
	}
}
