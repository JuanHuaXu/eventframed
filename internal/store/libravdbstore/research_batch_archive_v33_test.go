package libravdbstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

var errArchiveCapacityV33 = errors.New("archive accepted-job capacity exceeded")

type archiveJobV33 struct {
	capture    *archiveCaptureV29
	done       chan error
	capturedAt time.Time
}
type archiveProofV33 struct {
	ID, Root, Wire string
	Encoded        string
	Snapshot       model.Snapshot
	CapturedAt     time.Time
}
type archiveQueueStatsV33 struct {
	Accepted, Finished, Rejected, Peak, Active, Capacity int
	Closed                                               bool
}
type batchArchiveV33 struct {
	*archiveWitnessV29
	mu      sync.Mutex
	jobs    chan *archiveJobV33
	end     chan struct{}
	stats   archiveQueueStatsV33
	api     sync.WaitGroup
	batches []joinedBatchV25
	proofs  []archiveProofV33
	barrier func() // before owner acquisition, only controlled lifecycle tests
}

func attachBatchArchiveV33(t *testing.T, f *witnessFixtureV23, capacity int) *batchArchiveV33 {
	t.Helper()
	if capacity < 1 || capacity > 128 {
		t.Fatal("archive capacity must be1..128")
	}
	base := &archiveWitnessV29{scheduledWitnessV24: attachScheduledV24(t, f, true)}
	if err := base.publishedRecallLoadStoreV16.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := base.gate.sidecar.Exec("CREATE TABLE IF NOT EXISTS archive_capture_v29(journal_id TEXT PRIMARY KEY, root TEXT NOT NULL, wire TEXT NOT NULL, snapshot TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	s := &batchArchiveV33{archiveWitnessV29: base, jobs: make(chan *archiveJobV33, capacity), end: make(chan struct{}), stats: archiveQueueStatsV33{Capacity: capacity}}
	go s.workerV33()
	before := s.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != s.gate.store.Snapshot(context.Background()) {
		t.Fatal("batch attachment moved source")
	}
	f.svc = svc
	return s
}

func (s *batchArchiveV33) recall(ctx context.Context, svc *service.Service, req model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	s.mu.Lock()
	if s.stats.Closed {
		s.mu.Unlock()
		return model.ContextPacket{}, nil, errors.New("archive closed")
	}
	s.api.Add(1)
	s.mu.Unlock()
	defer s.api.Done()
	return s.archiveWitnessV29.recall(ctx, svc, req)
}

func (s *batchArchiveV33) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	if s.stats.Closed {
		s.mu.Unlock()
		return errors.New("archive closed")
	}
	if s.stats.Active >= s.stats.Capacity {
		s.stats.Rejected++
		s.mu.Unlock()
		return errArchiveCapacityV33
	}
	s.stats.Active++
	if s.stats.Active > s.stats.Peak {
		s.stats.Peak = s.stats.Active
	}
	s.api.Add(1)
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.stats.Active--; s.mu.Unlock(); s.api.Done() }()
	c, err := s.capture(ctx, e)
	if err != nil {
		return err
	}
	job := &archiveJobV33{capture: c, done: make(chan error, 1), capturedAt: time.Now().UTC()}
	s.mu.Lock()
	if s.stats.Closed {
		s.mu.Unlock()
		return errors.New("archive closed before enqueue")
	}
	// Queue reservation precedes copying. Under the same close/enqueue mutex,
	// end the read lease and enqueue without an orphaning cancellation branch.
	ctx.Value(archiveLeaseKeyV29{}).(*archiveLeaseV29).finish()
	s.jobs <- job
	s.stats.Accepted++
	s.mu.Unlock()
	if err := <-job.done; err != nil {
		return err
	}
	return ctx.Err()
}

func (s *batchArchiveV33) applyV33(batch []*archiveJobV33) (joinedBatchV25, error) {
	if s.barrier != nil {
		s.barrier()
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	row := joinedBatchV25{Begin: time.Now().UTC(), Before: s.state.Hash}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	jobs := make([]*joinedJobV25, len(batch))
	entries := make([]model.BayesianJournalEntry, len(batch))
	caps := map[string]*archiveCaptureV29{}
	for i, job := range batch {
		c := job.capture
		row.IDs = append(row.IDs, c.entry.ID)
		if err := s.validate(ctx, c, s.state.Head); err != nil {
			return row, err
		}
		if prior, ok := caps[c.entry.ID]; ok && (prior.wire != c.wire || prior.binding != c.binding) {
			return row, store.ErrJournalConflict
		}
		caps[c.entry.ID] = c
		jobs[i] = &joinedJobV25{entry: c.entry, binding: c.binding}
		entries[i] = c.entry
	}
	prep := &joinedWitnessV25{scheduledWitnessV24: s.scheduledWitnessV24}
	canonical := map[string]string{}
	for id, c := range caps {
		var root, wire, snap string
		err := s.gate.sidecar.QueryRowContext(ctx, "SELECT root,wire,snapshot FROM archive_capture_v29 WHERE journal_id=?", id).Scan(&root, &wire, &snap)
		if errors.Is(err, sql.ErrNoRows) {
			canonical[id] = c.root
			continue
		}
		if err != nil {
			return row, err
		}
		snapshot, _ := json.Marshal(c.entry.Snapshot)
		if wire != c.wire || snap != string(snapshot) {
			return row, store.ErrJournalConflict
		}
		old := *c
		old.root = root
		old.seal = witnessHashV23([]any{old.entry, old.binding, old.root})
		if err := s.validate(ctx, &old, s.state.Head); err != nil {
			return row, err
		}
		canonical[id] = root
	}
	next, payload, digest, err := prep.prepare(jobs)
	if err != nil {
		return row, err
	}
	prior := s.state.Hash
	err = s.gate.appendArchiveJournalV29(ctx, entries, s.stopAt, 0, func(tx *sql.Tx) error {
		if s.stopAt == "before_witness" {
			return errors.New("injected batch archive witness rollback")
		}
		for _, c := range caps {
			snapshot, err := json.Marshal(c.entry.Snapshot)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO archive_capture_v29(journal_id,root,wire,snapshot) VALUES(?,?,?,?) ON CONFLICT(journal_id) DO NOTHING", c.entry.ID, c.root, c.wire, string(snapshot)); err != nil {
				return err
			}
			var root, wire, snap string
			if err := tx.QueryRowContext(ctx, "SELECT root,wire,snapshot FROM archive_capture_v29 WHERE journal_id=?", c.entry.ID).Scan(&root, &wire, &snap); err != nil {
				return err
			}
			// Identical retries may capture a later journal-only head. Native full
			// wire equality plus the original witness binding retain the FIRST provenance;
			// retries do not rewrite that root, snapshot, or forecast.
			if root != canonical[c.entry.ID] || wire != c.wire || snap != string(snapshot) {
				return store.ErrJournalConflict
			}
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
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			return fmt.Errorf("archive batch state rows=%d: %v", n, err)
		}
		return nil
	}, func(e model.BayesianJournalEntry, current model.Snapshot) bool {
		c := caps[e.ID]
		return c != nil && current == s.state.Head && witnessHashV23(e) == c.wire && e.Snapshot == c.entry.Snapshot
	})
	if err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return row, err
	}
	s.state = next
	if err = s.publishLocked(ctx); err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return row, err
	}
	row.After = s.state.Hash
	row.End = time.Now().UTC()
	return row, nil
}

func (s *batchArchiveV33) workerV33() {
	defer close(s.end)
	for first := range s.jobs {
		batch := []*archiveJobV33{first}
		timer := time.NewTimer(time.Millisecond)
	collect:
		for len(batch) < 4 {
			select {
			case next, ok := <-s.jobs:
				if !ok {
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
		row, err := s.applyV33(batch)
		if err != nil {
			row.Error = err.Error()
			row.End = time.Now().UTC()
		}
		s.mu.Lock()
		s.batches = append(s.batches, row)
		s.stats.Finished += len(batch)
		for _, job := range batch {
			c := job.capture
			encoded, _ := json.Marshal(c.entry)
			s.proofs = append(s.proofs, archiveProofV33{ID: c.entry.ID, Root: c.root, Wire: c.wire, Encoded: string(encoded), Snapshot: c.entry.Snapshot, CapturedAt: job.capturedAt})
		}
		s.mu.Unlock()
		for _, job := range batch {
			job.done <- err
		}
	}
}

func (s *batchArchiveV33) Close() error {
	s.mu.Lock()
	if !s.stats.Closed {
		s.stats.Closed = true
		close(s.jobs)
	}
	s.mu.Unlock()
	s.scheduler.Close()
	<-s.end
	s.api.Wait()
	return nil
}

// Both loaded arms use the sealed fixed scheduler and durable native publication.
// Only the archived arm ends read admission at the owned handoff.
type loadArchiveV33 struct {
	*joinedWitnessV25
	archive *batchArchiveV33
}

func (s *loadArchiveV33) closeV33() error {
	if s.archive != nil {
		return s.archive.Close()
	}
	return s.joinedWitnessV25.Close()
}
func attachLoadArchiveV33(t *testing.T, f *witnessFixtureV23, archived bool) *loadArchiveV33 {
	if !archived {
		return &loadArchiveV33{joinedWitnessV25: attachCombinedV26(t, f, true)}
	}
	a := attachBatchArchiveV33(t, f, 128)
	return &loadArchiveV33{joinedWitnessV25: &joinedWitnessV25{scheduledWitnessV24: a.scheduledWitnessV24, joined: true}, archive: a}
}
func (s *loadArchiveV33) recall(ctx context.Context, svc *service.Service, r model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	if s.archive != nil {
		return s.archive.recall(ctx, svc, r)
	}
	return s.joinedWitnessV25.recall(ctx, svc, r)
}

type archiveTrialV33 struct {
	joinedTrialV25
	Archived     bool
	Captures     []archiveProofV33
	ArchiveQueue archiveQueueStatsV33
}

func finishArchiveTrialV33(s *loadArchiveV33, r joinedTrialV25) archiveTrialV33 {
	if s.archive == nil {
		return archiveTrialV33{joinedTrialV25: r}
	}
	s.archive.mu.Lock()
	defer s.archive.mu.Unlock()
	r.Batches = append([]joinedBatchV25(nil), s.archive.batches...)
	return archiveTrialV33{joinedTrialV25: r, Archived: true, Captures: append([]archiveProofV33(nil), s.archive.proofs...), ArchiveQueue: s.archive.stats}
}
