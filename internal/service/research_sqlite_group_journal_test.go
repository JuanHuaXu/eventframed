package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type researchGroupJournalRequest struct {
	entry   model.BayesianJournalEntry
	encoded []byte
	done    chan error
}

type researchSQLiteGroupJournalStore struct {
	*researchSQLiteJournalStore
	batchGuard interface {
		WithResearchAsOfSnapshotsWait(context.Context, []researchpublicationstore.AsOfSnapshot, func() error) error
	}
	jobs         chan *researchGroupJournalRequest
	workerEnd    chan struct{}
	maxBatch     int
	maxWait      time.Duration
	mu           sync.Mutex
	closed       bool
	statsMu      sync.Mutex
	batchSize    []int
	guardRejects int
	closeErr     error
}

func newResearchSQLiteGroupJournalStore(base *researchSQLiteJournalStore, maxBatch int, maxWait time.Duration) (*researchSQLiteGroupJournalStore, error) {
	if base == nil || maxBatch < 1 || maxBatch > 8 || maxWait <= 0 {
		return nil, errors.New("invalid research journal group size or dwell")
	}
	guard, ok := base.EventStore.(interface {
		WithResearchAsOfSnapshotsWait(context.Context, []researchpublicationstore.AsOfSnapshot, func() error) error
	})
	if !ok {
		return nil, errors.New("research backend lacks batch as-of guard")
	}
	g := &researchSQLiteGroupJournalStore{researchSQLiteJournalStore: base, batchGuard: guard, jobs: make(chan *researchGroupJournalRequest, 128), workerEnd: make(chan struct{}), maxBatch: maxBatch, maxWait: maxWait}
	go g.run()
	return g, nil
}

func (g *researchSQLiteGroupJournalStore) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if entry.TenantID == "" || entry.ID == "" || entry.AsOf.IsZero() {
		return errors.New("invalid research group journal entry")
	}
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	prior, err := g.readRaw(ctx, entry.TenantID, entry.ID)
	if err == nil {
		if bytes.Equal(prior, encoded) {
			return nil
		}
		return store.ErrJournalConflict
	}
	if !errors.Is(err, store.ErrJournalNotFound) {
		return err
	}
	request := &researchGroupJournalRequest{entry: entry, encoded: encoded, done: make(chan error, 1)}
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return errors.New("research group journal closed")
	}
	select {
	case g.jobs <- request:
	case <-ctx.Done():
		g.mu.Unlock()
		return ctx.Err()
	}
	g.mu.Unlock()
	select {
	case result := <-request.done:
		return result
	case <-ctx.Done():
		// A queued commit may finish after this uncertain acknowledgement.
		return ctx.Err()
	}
}

func (g *researchSQLiteGroupJournalStore) run() {
	defer close(g.workerEnd)
	for first := range g.jobs {
		batch := []*researchGroupJournalRequest{first}
		timer := time.NewTimer(g.maxWait)
		closed := false
	collect:
		for len(batch) < g.maxBatch {
			select {
			case next, ok := <-g.jobs:
				if !ok {
					closed = true
					break collect
				}
				batch = append(batch, next)
			case <-timer.C:
				break collect
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		g.process(batch)
		if closed {
			return
		}
	}
}

func (g *researchSQLiteGroupJournalStore) process(batch []*researchGroupJournalRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	pending := make([]*researchGroupJournalRequest, 0, len(batch))
	for _, request := range batch {
		prior, err := g.readRaw(ctx, request.entry.TenantID, request.entry.ID)
		if err == nil {
			if bytes.Equal(prior, request.encoded) {
				request.done <- nil
			} else {
				request.done <- store.ErrJournalConflict
			}
			continue
		}
		if !errors.Is(err, store.ErrJournalNotFound) {
			request.done <- err
			continue
		}
		pending = append(pending, request)
	}
	if len(pending) == 0 {
		return
	}
	checks := make([]researchpublicationstore.AsOfSnapshot, 0, len(pending))
	for _, request := range pending {
		checks = append(checks, researchpublicationstore.AsOfSnapshot{Captured: request.entry.Snapshot, AsOf: request.entry.AsOf})
	}
	g.statsMu.Lock()
	g.batchSize = append(g.batchSize, len(pending))
	g.statsMu.Unlock()
	results := make([]error, len(pending))
	callbackEntered := false
	err := g.batchGuard.WithResearchAsOfSnapshotsWait(ctx, checks, func() error {
		callbackEntered = true
		tx, beginErr := g.db.BeginTx(ctx, nil)
		if beginErr != nil {
			return beginErr
		}
		defer tx.Rollback()
		for i, request := range pending {
			result, insertErr := tx.ExecContext(ctx, `INSERT INTO journals(tenant,id,payload) VALUES(?,?,?) ON CONFLICT(tenant,id) DO NOTHING`, request.entry.TenantID, request.entry.ID, request.encoded)
			if insertErr != nil {
				return insertErr
			}
			inserted, insertErr := result.RowsAffected()
			if insertErr != nil {
				return insertErr
			}
			if inserted == 0 {
				var current []byte
				if readErr := tx.QueryRowContext(ctx, `SELECT payload FROM journals WHERE tenant=? AND id=?`, request.entry.TenantID, request.entry.ID).Scan(&current); readErr != nil {
					return readErr
				}
				if !bytes.Equal(current, request.encoded) {
					results[i] = store.ErrJournalConflict
				}
			}
		}
		return tx.Commit()
	})
	if err != nil && !callbackEntered {
		g.statsMu.Lock()
		g.guardRejects++
		g.statsMu.Unlock()
	}
	for i, request := range pending {
		if err != nil {
			request.done <- err
		} else {
			request.done <- results[i]
		}
	}
}

func (g *researchSQLiteGroupJournalStore) closeGroup() error {
	g.mu.Lock()
	if !g.closed {
		g.closed = true
		close(g.jobs)
	}
	g.mu.Unlock()
	<-g.workerEnd
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closeErr == nil {
		g.closeErr = g.closeJournal()
	}
	return g.closeErr
}

func (g *researchSQLiteGroupJournalStore) batchSizes() []int {
	g.statsMu.Lock()
	defer g.statsMu.Unlock()
	return append([]int(nil), g.batchSize...)
}

func (g *researchSQLiteGroupJournalStore) guardRejections() int {
	g.statsMu.Lock()
	defer g.statsMu.Unlock()
	return g.guardRejects
}

func TestResearchSQLiteGroupJournalContractV22(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	base, err := researchpublicationstore.New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	path := t.TempDir() + "/group.sqlite"
	sidecar, err := openResearchSQLiteJournalStore(base, path, "group-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	group, err := newResearchSQLiteGroupJournalStore(sidecar, 8, 8*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	snapshot := base.Snapshot(ctx)
	start := make(chan struct{})
	var workers sync.WaitGroup
	outcomes := make(chan error, 16)
	for i := range 16 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			entry := model.BayesianJournalEntry{ID: fmt.Sprintf("journal-%d", i), TenantID: "tenant-a", SessionID: "session-a", AsOf: asOf, Snapshot: snapshot}
			outcomes <- group.PutBayesianJournal(ctx, entry)
		}()
	}
	close(start)
	workers.Wait()
	close(outcomes)
	for result := range outcomes {
		if result != nil {
			t.Fatal(result)
		}
	}
	if err := group.closeGroup(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openResearchSQLiteJournalStore(base, path, "group-owner", false)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.closeJournal()
	if count, err := reopened.count(ctx); err != nil || count != 16 {
		t.Fatalf("group journal lost acknowledged rows: count=%d err=%v", count, err)
	}
	for i := range 16 {
		id := fmt.Sprintf("journal-%d", i)
		if _, err := reopened.GetBayesianJournal(ctx, "tenant-a", id); err != nil {
			t.Fatalf("missing acknowledged %s: %v", id, err)
		}
	}
	entry := model.BayesianJournalEntry{ID: "journal-0", TenantID: "tenant-a", SessionID: "session-a", AsOf: asOf, Snapshot: snapshot}
	if err := reopened.PutBayesianJournal(ctx, entry); err != nil {
		t.Fatalf("uncertain acknowledgement exact retry: %v", err)
	}
	entry.SessionID = "changed"
	if err := reopened.PutBayesianJournal(ctx, entry); !errors.Is(err, store.ErrJournalConflict) {
		t.Fatalf("changed retry: got %v", err)
	}
}

func TestResearchSQLiteGroupJournalMixedHorizonRejectsV22(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	base, err := researchpublicationstore.New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	path := t.TempDir() + "/mixed.sqlite"
	sidecar, err := openResearchSQLiteJournalStore(base, path, "mixed-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	group, err := newResearchSQLiteGroupJournalStore(sidecar, 8, 8*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer group.closeGroup()
	at := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	first := base.Snapshot(ctx)
	// The second captured snapshot precedes a write that is future for the
	// first horizon but invalid for the second horizon.
	fromFirst := model.BayesianJournalEntry{ID: "first", TenantID: "tenant-a", SessionID: "a", AsOf: at, Snapshot: first}
	fromSecond := model.BayesianJournalEntry{ID: "second", TenantID: "tenant-a", SessionID: "b", AsOf: at.Add(2 * time.Hour)}
	// Fill the second snapshot and intervening write using the existing store.
	if _, err := base.Put(ctx, testutil.Event("future", "public batch fixture", at.Add(time.Hour)), make([]float32, 8), "digest"); err != nil {
		t.Fatal(err)
	}
	fromSecond.Snapshot = base.Snapshot(ctx)
	if _, err := base.Put(ctx, testutil.Event("intervening", "public batch fixture", at.Add(90*time.Minute)), make([]float32, 8), "digest"); err != nil {
		t.Fatal(err)
	}
	requests := []*researchGroupJournalRequest{
		{entry: fromFirst, done: make(chan error, 1)},
		{entry: fromSecond, done: make(chan error, 1)},
	}
	for _, request := range requests {
		request.encoded, err = json.Marshal(request.entry)
		if err != nil {
			t.Fatal(err)
		}
	}
	group.process(requests)
	for _, request := range requests {
		if err := <-request.done; err == nil {
			t.Fatal("mixed invalid batch acknowledged")
		}
	}
	if count, err := group.count(ctx); err != nil || count != 0 {
		t.Fatalf("mixed invalid batch committed: count=%d err=%v", count, err)
	}
}

func TestResearchSQLiteGroupJournalStatsDoNotBlockFullQueueV22(t *testing.T) {
	base, err := researchpublicationstore.New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	sidecar, err := openResearchSQLiteJournalStore(base, t.TempDir()+"/pressure.sqlite", "pressure-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	defer sidecar.closeJournal()
	g := &researchSQLiteGroupJournalStore{researchSQLiteJournalStore: sidecar, jobs: make(chan *researchGroupJournalRequest, 1)}
	g.jobs <- &researchGroupJournalRequest{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- g.PutBayesianJournal(ctx, model.BayesianJournalEntry{ID: "pressure", TenantID: "tenant-a", SessionID: "session-a", AsOf: time.Now(), Snapshot: base.Snapshot(ctx)})
	}()
	locked := false
	for i := 0; i < 100; i++ {
		if !g.mu.TryLock() {
			locked = true
			break
		}
		g.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
	if !locked {
		t.Fatal("producer did not block on the full queue")
	}
	statsDone := make(chan struct{}, 1)
	go func() {
		_ = g.batchSizes()
		statsDone <- struct{}{}
	}()
	select {
	case <-statsDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("statistics blocked behind a full enqueue")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("blocked enqueue returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled enqueue did not release the close mutex")
	}
}
