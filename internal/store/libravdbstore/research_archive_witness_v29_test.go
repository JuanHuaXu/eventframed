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
	"github.com/JuanHuaXu/eventframed/internal/researchadmission"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type archiveLeaseKeyV29 struct{}
type archiveLeaseV29 struct {
	once    sync.Once
	release func()
}

func (l *archiveLeaseV29) finish() { l.once.Do(l.release) }

// A capture is process-local authority, not a claim that its forecast is current.
// Its wire and dependency provenance are owned before the read lease ends.
type archiveCaptureV29 struct {
	entry            model.BayesianJournalEntry
	binding          witnessBindingV23
	root, wire, seal string
	owner            *archiveWitnessV29
}
type archiveWitnessV29 struct {
	*scheduledWitnessV24
	afterCapture func(*archiveCaptureV29) // test barrier, fixed before admission
	stopAt       string
}

func attachArchiveV29(t *testing.T, f *witnessFixtureV23) *archiveWitnessV29 {
	t.Helper()
	s := &archiveWitnessV29{scheduledWitnessV24: attachScheduledV24(t, f, true)}
	if err := s.publishedRecallLoadStoreV16.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := s.gate.sidecar.Exec(`CREATE TABLE archive_capture_v29(journal_id TEXT PRIMARY KEY, root TEXT NOT NULL, wire TEXT NOT NULL, snapshot TEXT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	before := s.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != s.gate.store.Snapshot(context.Background()) {
		t.Fatal("attachment moved source")
	}
	f.svc = svc
	return s
}

func (s *archiveWitnessV29) recall(ctx context.Context, svc *service.Service, req model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	if req.TenantID != "tenant-a" || req.Query == "" || req.EmbeddingModel == "" {
		return model.ContextPacket{}, nil, errors.New("invalid archive request")
	}
	release, err := s.acquire(ctx, researchadmission.Recall)
	if err != nil {
		return model.ContextPacket{}, nil, err
	}
	s.admission.RLock()
	l := &archiveLeaseV29{release: func() { s.admission.RUnlock(); release() }}
	defer l.finish()
	selection := req
	selection.Embedding, selection.AsOf = nil, time.Time{}
	scope := &witnessScopeV23{Binding: witnessHashV23(selection)}
	pin := &publishedRecallPinV8{}
	ctx = context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
	ctx = context.WithValue(ctx, witnessContextV23{}, scope)
	ctx = context.WithValue(ctx, archiveLeaseKeyV29{}, l)
	packet, err := svc.Recall(ctx, req)
	return packet, pin.view, err
}

func (s *archiveWitnessV29) capture(ctx context.Context, e model.BayesianJournalEntry) (*archiveCaptureV29, error) {
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	pin, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	l, _ := ctx.Value(archiveLeaseKeyV29{}).(*archiveLeaseV29)
	if l == nil || scope == nil || pin == nil || pin.view == nil || scope.Binding == "" || scope.Vector == "" || scope.Frontier == "" || scope.Snapshot != e.Snapshot || scope.AsOf != e.AsOf || pin.view.snapshot != e.Snapshot {
		return nil, errors.New("missing owned capture scope")
	}
	encoded, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	c := &archiveCaptureV29{owner: s, wire: witnessHashV23(e), binding: witnessBindingV23{Binding: scope.Binding, Vector: scope.Vector, Frontier: scope.Frontier, Snapshot: e.Snapshot, SourceAt: e.AsOf}}
	if err := json.Unmarshal(encoded, &c.entry); err != nil {
		return nil, err
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	if s.poison.Load() || s.state.Head != e.Snapshot || s.gate.store.Snapshot(ctx) != e.Snapshot {
		return nil, store.ErrStaleSnapshot
	}
	c.root = s.state.Hash
	c.seal = witnessHashV23([]any{c.entry, c.binding, c.root})
	return c, nil
}

// Called with owner held. Recheck the complete committed witness chain, including
// the capture point and the latest state hash. Unknown/missing history rejects.
func (s *archiveWitnessV29) validate(ctx context.Context, c *archiveCaptureV29, current model.Snapshot) error {
	if c == nil || c.owner != s || c.root == "" || c.wire != witnessHashV23(c.entry) || c.seal != witnessHashV23([]any{c.entry, c.binding, c.root}) || c.binding.Snapshot != c.entry.Snapshot || c.binding.SourceAt != c.entry.AsOf || s.poison.Load() || current != s.state.Head || current != s.gate.store.Snapshot(ctx) {
		return errors.New("invalid archive authority")
	}
	head := s.state.Base
	root := witnessGenesisV23(head).Hash
	found := root == c.root && head == c.entry.Snapshot
	stateHash := root
	rows, err := s.gate.sidecar.QueryContext(ctx, "SELECT payload,prior,digest FROM witness_v23 ORDER BY seq")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		if count > 11000 {
			return errors.New("archive history cap")
		}
		var payload, prior, digest string
		if err := rows.Scan(&payload, &prior, &digest); err != nil {
			return err
		}
		var tr witnessTransitionV23
		if err := json.Unmarshal([]byte(payload), &tr); err != nil {
			return err
		}
		if prior != root || tr.Before != head || digest != witnessHashV23([]string{prior, payload}) {
			return errors.New("archive chain gap")
		}
		root, head, stateHash = digest, tr.After, tr.StateHash
		if root == c.root && head == c.entry.Snapshot {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !found || head != current || root != s.state.Hash || stateHash != witnessStateHashV23(*s.state) {
		return errors.New("archive ancestry mismatch")
	}
	return nil
}

func (s *archiveWitnessV29) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c, err := s.capture(ctx, e)
	if err != nil {
		return err
	}
	l := ctx.Value(archiveLeaseKeyV29{}).(*archiveLeaseV29)
	l.finish()
	if s.afterCapture != nil {
		s.afterCapture(c)
	}
	// Caller cancellation cannot abandon an accepted capture. The call returns no
	// packet on cancellation, but waits for the terminal persistence result.
	work, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := s.commit(work, c); err != nil {
		return err
	}
	return ctx.Err()
}

func (s *archiveWitnessV29) commit(ctx context.Context, c *archiveCaptureV29) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	if err := s.validate(ctx, c, s.state.Head); err != nil {
		return err
	}
	// Reuse sealed state preparation, but retain the ORIGINAL captured binding.
	prep := &joinedWitnessV25{scheduledWitnessV24: s.scheduledWitnessV24}
	job := &joinedJobV25{entry: c.entry, binding: c.binding}
	next, payload, digest, err := prep.prepare([]*joinedJobV25{job})
	if err != nil {
		return err
	}
	prior := s.state.Hash
	err = s.gate.appendArchiveJournalV29(ctx, []model.BayesianJournalEntry{c.entry}, s.stopAt, 0, func(tx *sql.Tx) error {
		if s.stopAt == "before_witness" {
			return errors.New("injected archive witness rollback")
		}
		snapshot, err := json.Marshal(c.entry.Snapshot)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO archive_capture_v29(journal_id,root,wire,snapshot) VALUES(?,?,?,?) ON CONFLICT(journal_id) DO NOTHING", c.entry.ID, c.root, c.wire, string(snapshot))
		if err != nil {
			return err
		}
		var root, wire, snap string
		if err := tx.QueryRowContext(ctx, "SELECT root,wire,snapshot FROM archive_capture_v29 WHERE journal_id=?", c.entry.ID).Scan(&root, &wire, &snap); err != nil {
			return err
		}
		if root != c.root || wire != c.wire || snap != string(snapshot) {
			return store.ErrJournalConflict
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
			return fmt.Errorf("archive state rows=%d: %v", n, err)
		}
		return nil
	}, func(entry model.BayesianJournalEntry, current model.Snapshot) bool {
		return current == s.state.Head && witnessHashV23(entry) == c.wire && entry.Snapshot == c.entry.Snapshot
	})
	if err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return err
	}
	s.state = next
	if err := s.publishLocked(ctx); err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return err
	}
	return nil
}

func (s *archiveWitnessV29) Close() error { s.scheduler.Close(); return nil }
