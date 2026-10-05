package libravdbstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type joinedJobV25 struct {
	entry   model.BayesianJournalEntry
	binding witnessBindingV23
	done    chan error
}
type joinedBatchV25 struct {
	IDs           []string
	Begin, End    time.Time
	Before, After string
	Error         string
}
type joinedWitnessV25 struct {
	*scheduledWitnessV24
	joined  bool
	mu      sync.Mutex
	closed  bool
	jobs    chan *joinedJobV25
	end     chan struct{}
	batches []joinedBatchV25 // worker owner; read only after serving drained
	stopAt  string           // injection fixed before any operation
	barrier func()           // cancellation test only, nil in load
}

// One owner performs native wire readback and commits all witness bindings
// with the SQLite marker. This is NOT atomic across the two databases: an
// interrupted native commit must remain unacknowledged and fail closed.
func (s *joinedWitnessV25) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if !s.joined {
		return s.scheduledWitnessV24.PutBayesianJournal(ctx, entry)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if scope == nil || scope.Binding == "" || scope.Vector == "" || scope.Frontier == "" || scope.Snapshot != entry.Snapshot || scope.AsOf != entry.AsOf {
		return errors.New("missing/mismatched joined scope")
	}
	job := &joinedJobV25{entry: entry, binding: witnessBindingV23{Binding: scope.Binding, Vector: scope.Vector, Frontier: scope.Frontier, Snapshot: entry.Snapshot, SourceAt: scope.AsOf}, done: make(chan error, 1)}
	// Own the submitted value; the worker never reads mutable caller fields.
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(encoded, &job.entry); err != nil {
		return err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors.New("joined worker closed")
	}
	select {
	case s.jobs <- job:
	case <-ctx.Done():
		s.mu.Unlock()
		return ctx.Err()
	}
	s.mu.Unlock()
	// After successful enqueue, keep the read lease until terminal durability.
	// A canceled call does not receive a packet, but cannot orphan its handoff.
	if err := <-job.done; err != nil {
		return err
	}
	return ctx.Err()
}

func (s *joinedWitnessV25) prepare(batch []*joinedJobV25) (*witnessStateV23, string, string, error) {
	if s.poison.Load() || s.state.Head != s.gate.store.Snapshot(context.Background()) {
		return nil, "", "", store.ErrStaleSnapshot
	}
	data, err := json.Marshal(s.state)
	if err != nil {
		return nil, "", "", err
	}
	var next witnessStateV23
	if err = json.Unmarshal(data, &next); err != nil {
		return nil, "", "", err
	}
	for _, job := range batch {
		if previous, ok := next.Journals[job.entry.ID]; ok && previous != job.binding {
			return nil, "", "", store.ErrJournalConflict
		}
		next.Journals[job.entry.ID] = job.binding
		if next.Certificate == nil {
			b := job.binding
			b.Snapshot = next.Base
			next.Certificate = &b
		}
	}
	if len(next.Journals) > 512 || len(next.Sources) > 200 || len(next.Log) > 10000 {
		return nil, "", "", errors.New("joined witness cap")
	}
	tr := witnessTransitionV23{Before: next.Head, After: next.Head, StateHash: witnessStateHashV23(next)}
	encoded, err := json.Marshal(tr)
	if err != nil {
		return nil, "", "", err
	}
	payload := string(encoded)
	digest := witnessHashV23([]string{s.state.Hash, payload})
	next.Hash = digest
	return &next, payload, digest, nil
}

func (s *joinedWitnessV25) apply(batch []*joinedJobV25) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	if s.barrier != nil {
		s.barrier()
	}
	next, payload, digest, err := s.prepare(batch)
	if err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return err
	}
	entries := make([]model.BayesianJournalEntry, len(batch))
	for i, job := range batch {
		entries[i] = job.entry
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	prior := s.state.Hash
	err = s.gate.appendJoinedJournalV25(ctx, entries, s.stopAt, 0, func(tx *sql.Tx) error {
		if s.stopAt == "before_witness" {
			return errors.New("injected witness rollback")
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO witness_v23(payload,prior,digest) VALUES(?,?,?)", payload, prior, digest); err != nil {
			return err
		}
		encoded, err := json.Marshal(next)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, "UPDATE witness_state_v23 SET payload=? WHERE id=1", string(encoded))
		if err != nil {
			return err
		}
		if count, err := result.RowsAffected(); err != nil || count != 1 {
			return errors.New("missing witness state row")
		}
		return nil
	})
	if err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return err
	}
	s.state = next
	if err = s.publishLocked(ctx); err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
	}
	return err
}

func (s *joinedWitnessV25) worker() {
	defer close(s.end)
	for first := range s.jobs {
		batch := []*joinedJobV25{first}
		timer := time.NewTimer(time.Millisecond)
		closed := false
	collect:
		for len(batch) < 4 {
			select {
			case next, ok := <-s.jobs:
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
		row := joinedBatchV25{Begin: time.Now().UTC()}
		for _, job := range batch {
			row.IDs = append(row.IDs, job.entry.ID)
		}
		// Reads already hold admission, so writers cannot advance witness state
		// until the entire batch has acknowledged and all reader owners release.
		row.Before = s.state.Hash
		err := s.apply(batch)
		row.After = s.state.Hash
		row.End = time.Now().UTC()
		if err != nil {
			row.Error = err.Error()
		}
		s.batches = append(s.batches, row)
		for _, job := range batch {
			job.done <- err
		}
		if closed {
			return
		}
	}
}

func (s *joinedWitnessV25) Close() error {
	if !s.joined {
		return s.scheduledWitnessV24.Close()
	}
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.jobs)
	}
	s.mu.Unlock()
	<-s.end
	s.scheduler.Close()
	return nil
}

func attachJoinedV25(t *testing.T, f *witnessFixtureV23, joined bool) *joinedWitnessV25 {
	t.Helper()
	s := &joinedWitnessV25{scheduledWitnessV24: attachScheduledV24(t, f, false), joined: joined, jobs: make(chan *joinedJobV25, 128), end: make(chan struct{})}
	if joined {
		// No requests exist at attachment. Stop only the isolated idle V16
		// worker; the database, immutable pins and published state stay open.
		if err := s.publishedRecallLoadStoreV16.Close(); err != nil {
			t.Fatal(err)
		}
		go s.worker()
	}
	before := f.adapter.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != f.adapter.gate.store.Snapshot(context.Background()) {
		t.Fatal("attachment moved source")
	}
	f.svc = svc
	return s
}

func TestResearchJoinedWitnessLifecycleV25(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOINED_WITNESS_V25") != "1" {
		t.Skip("isolated joined integration")
	}
	ctx := context.Background()
	for _, visible := range []bool{false, true} {
		t.Run(fmt.Sprint("visible-", visible), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachJoinedV25(t, f, true)
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			response, err := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "first", false)
			if err != nil {
				t.Fatal(err)
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite("joined-event", at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err := s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			after, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "after"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong transport", err)
			}
			p, err := s.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if err != nil || !reflect.DeepEqual(p, response.Posterior) {
				t.Fatal("retagged posterior", err)
			}
			for _, alter := range []string{"query", "vector", "selection", "old"} {
				r := f.request(time.Now().UTC(), "changed")
				switch alter {
				case "query":
					r.Query = "different"
				case "vector":
					r.Embedding = denseRowV6(f.query, 11, .1)
				case "selection":
					r.PackK = 9
				case "old":
					r.AsOf = p.UpdatedAt.Add(-time.Nanosecond)
				}
				other, _, err := s.recall(ctx, f.svc, r)
				if err != nil || witnessBeliefsV23(other) != 0 {
					t.Fatal("changed request reused", alter, err)
				}
			}
			f.reopen(t)
			s = attachJoinedV25(t, f, true)
			if s.poison.Load() {
				t.Fatal("reopen poisoned")
			}
			after, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong reopened transport", err)
			}
		})
	}
	for _, stop := range []string{"after_db", "before_witness", "after_sqlite"} {
		t.Run(stop, func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachJoinedV25(t, f, true)
			s.stopAt = stop
			if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "interrupt")); err == nil {
				t.Fatal("interruption acknowledged")
			}
			if !s.poison.Load() || s.current.Load() != nil {
				t.Fatal("failure not blocked")
			}
			if stop == "after_sqlite" {
				f.reopen(t)
				s = attachJoinedV25(t, f, true)
				if s.poison.Load() {
					t.Fatal("durable joint commit lost")
				}
				if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "after")); err != nil {
					t.Fatal(err)
				}
			} else {
				// Reopen requires matching native LSN/SQLite marker. A DB-only
				// commit or rolled-back witness cannot invent a new READY marker.
				f.close()
				g, err := openDenseOutcomeGateV17(f.root)
				if err != nil {
					t.Fatal(err)
				}
				defer g.close()
				if _, ready := g.capture(ctx); ready {
					t.Fatal("incomplete cross-store publication became READY")
				}
				base := &publishedRecallLoadStoreV16{gate: g}
				if err := base.publishLocked(ctx); err == nil || base.current.Load() != nil {
					t.Fatal("incomplete publication served")
				}
			}
		})
	}
}

func TestResearchJoinedWitnessCancellationV25(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOINED_WITNESS_V25") != "1" {
		t.Skip("isolated joined cancellation")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachJoinedV25(t, f, true)
	entered, release := make(chan struct{}), make(chan struct{})
	s.barrier = func() { close(entered); <-release }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "cancel")); done <- err }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("no handoff")
	}
	cancel()
	select {
	case err := <-done:
		close(release)
		t.Fatal("early terminal", err)
	case <-time.After(10 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no terminal")
	}
	if len(s.state.Journals) != 1 {
		t.Fatal("lost binding")
	}
	f.reopen(t)
	s = attachJoinedV25(t, f, true)
	if s.poison.Load() || len(s.state.Journals) != 1 {
		t.Fatal("canceled durability lost")
	}
}

func TestResearchJoinedWitnessBatchV25(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOINED_WITNESS_V25") != "1" {
		t.Skip("isolated joined batch contract")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachJoinedV25(t, f, true)
	p, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "prime"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := s.gate.store.GetBayesianJournal(context.Background(), "tenant-a", p.BayesianShadow.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	b := s.state.Journals[e.ID]
	jobs := make([]*joinedJobV25, 4)
	for i := range jobs {
		clone := e
		clone.ID = fmt.Sprintf("batch-%d", i)
		clone.Report.JournalID = clone.ID
		jobs[i] = &joinedJobV25{entry: clone, binding: b}
	}
	before := len(s.state.Journals)
	if err := s.apply(jobs); err != nil {
		t.Fatal(err)
	}
	if len(s.state.Journals) != before+4 {
		t.Fatal("partial witness batch")
	}
	for _, job := range jobs {
		saved, err := s.gate.store.GetBayesianJournal(context.Background(), "tenant-a", job.entry.ID)
		got, _ := json.Marshal(saved)
		want, _ := json.Marshal(job.entry)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("wire mismatch", err)
		}
	}
	if err := s.apply(jobs); err != nil {
		t.Fatal("duplicate batch", err)
	}
	conflict := *jobs[0]
	conflict.binding.Binding = "wrong"
	if err := s.apply([]*joinedJobV25{&conflict}); !errors.Is(err, store.ErrJournalConflict) {
		t.Fatal("conflicting scope accepted", err)
	}
	if len(s.state.Journals) != before+4 {
		t.Fatal("conflict mutated state")
	}
}

func TestResearchJoinedWitnessGapV25(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOINED_WITNESS_V25") != "1" {
		t.Skip("isolated joined fail-closed controls")
	}
	ctx := context.Background()
	for _, interrupt := range []bool{false, true} {
		t.Run(fmt.Sprint("runtime-", interrupt), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachJoinedV25(t, f, true)
			w := pinnedWrite("unaccounted", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 3, .0001)
			if interrupt {
				if _, err := s.append(ctx, []ResearchEventWrite{w}, true); err == nil {
					t.Fatal("interruption missing")
				}
			} else {
				if err := appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); err != nil {
					t.Fatal(err)
				}
			}
			if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatal("unaccounted runtime served", err)
			}
			f.reopen(t)
			s = attachJoinedV25(t, f, true)
			if !s.poison.Load() {
				t.Fatal("reopened runtime gap repaired")
			}
			if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatal("reopened gap served", err)
			}
		})
	}
	t.Run("wire-conflict-no-partial-batch", func(t *testing.T) {
		f := createWitnessFixtureV23(t, true)
		defer f.close()
		s := attachJoinedV25(t, f, true)
		p, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
		if err != nil {
			t.Fatal(err)
		}
		e, err := s.gate.store.GetBayesianJournal(ctx, "tenant-a", p.BayesianShadow.JournalID)
		if err != nil {
			t.Fatal(err)
		}
		b := s.state.Journals[e.ID]
		fresh := e
		fresh.ID = "never-partial"
		fresh.Report.JournalID = fresh.ID
		conflict := e
		conflict.QueryDigest = "different-wire-query"
		before := s.state.Hash
		if err := s.apply([]*joinedJobV25{{entry: fresh, binding: b}, {entry: conflict, binding: b}}); !errors.Is(err, store.ErrJournalConflict) {
			t.Fatal("wire conflict accepted", err)
		}
		if s.state.Hash != before {
			t.Fatal("rollback changed witness")
		}
		if _, err := s.gate.store.GetBayesianJournal(ctx, "tenant-a", fresh.ID); !errors.Is(err, store.ErrJournalNotFound) {
			t.Fatal("partial native batch", err)
		}
	})
	t.Run("missing-scope-and-canceled", func(t *testing.T) {
		f := createWitnessFixtureV23(t, true)
		defer f.close()
		s := attachJoinedV25(t, f, true)
		if err := s.PutBayesianJournal(ctx, model.BayesianJournalEntry{}); err == nil {
			t.Fatal("missing scope accepted")
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if err := s.PutBayesianJournal(c, model.BayesianJournalEntry{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if len(s.state.Journals) != 0 {
			t.Fatal("invalid call persisted")
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "closed")); err == nil {
			t.Fatal("closed worker accepted")
		}
	})
}
