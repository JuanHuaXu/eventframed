// Package researchpublication models coherent research-only version publication.
// It is NOT connected to the persistent store. Every backend mutation would need
// to begin here before it can affect data, with authoritative commit resolution.
package researchpublication

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type Kind uint8

const (
	General Kind = iota
	Ingestion
)

type state struct {
	snapshot    model.Snapshot
	motion      map[uint64]time.Time
	ticket      uint64
	kind        Kind
	available   time.Time
	quarantined bool
}

type Publisher struct {
	mu   sync.Mutex
	next uint64
	view atomic.Pointer[state]
}

func New(snapshot model.Snapshot) *Publisher {
	p := &Publisher{}
	p.view.Store(&state{snapshot: snapshot, motion: map[uint64]time.Time{}})
	return p
}

// NewWithMotion accepts only checked, bounded history from a sidecar whose
// checkpoint equals snapshot. Missing versions stay missing, not inferred.
func NewWithMotion(snapshot model.Snapshot, motion map[uint64]time.Time) (*Publisher, error) {
	copyMotion := make(map[uint64]time.Time, len(motion))
	for version, at := range motion {
		if version == 0 || version > snapshot.RuntimeVersion || snapshot.RuntimeVersion-version >= 4096 || at.IsZero() {
			return nil, errors.New("invalid restored publication motion")
		}
		copyMotion[version] = at
	}
	p := &Publisher{}
	p.view.Store(&state{snapshot: snapshot, motion: copyMotion})
	return p, nil
}

// Begin must precede ANY backend-visible mutation. The caller must hold backend
// write ownership; this token does not acquire a database transaction for it.
func (p *Publisher) Begin(kind Kind, available time.Time) (uint64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	old := p.view.Load()
	if old.quarantined || old.ticket != 0 || p.next == ^uint64(0) || kind > Ingestion || (kind == Ingestion && available.IsZero()) {
		return 0, errors.New("invalid publication begin")
	}
	next := *old
	p.next++
	next.ticket = p.next
	next.kind = kind
	next.available = available
	p.view.Store(&next)
	return next.ticket, nil
}

// Compatible reads one immutable state without the writer mutex. For a declared
// ingestion, only query times strictly before the pending event may pass. That
// privilege depends on the writer actually obeying the ingestion-only contract.
func (p *Publisher) Compatible(captured model.Snapshot, asOf time.Time) bool {
	s := p.view.Load()
	if asOf.IsZero() || s.quarantined {
		return false
	}
	if s.ticket != 0 && (s.kind != Ingestion || !s.available.After(asOf)) {
		return false
	}
	return store.ResearchSnapshotCompatible(captured, s.snapshot, asOf, s.motion)
}

// CommittedSnapshotMatches checks agreement with an externally locked backend.
// It is not a reservation; callers must exclude backend/publication mutation.
func (p *Publisher) CommittedSnapshotMatches(snapshot model.Snapshot) bool {
	s := p.view.Load()
	return !s.quarantined && s.ticket == 0 && s.snapshot == snapshot
}

// Commit validates the declared transition. A mismatch quarantines the view;
// it must not pretend the backend rolled back after a possibly durable commit.
func (p *Publisher) Commit(ticket uint64, snapshot model.Snapshot) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	old := p.view.Load()
	if ticket == 0 || old.ticket != ticket || old.quarantined {
		return errors.New("invalid publication ticket")
	}
	next := *old
	valid := snapshot.RuntimeVersion > old.snapshot.RuntimeVersion
	if old.kind == Ingestion {
		expected := old.snapshot
		valid = expected.RuntimeVersion != ^uint64(0) && expected.EvidenceEpoch != ^uint64(0)
		expected.RuntimeVersion++
		expected.EvidenceEpoch++
		valid = valid && snapshot == expected
	}
	if !valid {
		next.quarantined = true
		p.view.Store(&next)
		return errors.New("unexpected committed transition")
	}
	// Old readers retain their map. Never mutate a published history in place.
	next.motion = make(map[uint64]time.Time, len(old.motion)+1)
	for version, at := range old.motion {
		if snapshot.RuntimeVersion-version < 4096 {
			next.motion[version] = at
		}
	}
	if old.kind == Ingestion {
		next.motion[snapshot.RuntimeVersion] = old.available
	}
	next.snapshot = snapshot
	next.ticket = 0
	p.view.Store(&next)
	return nil
}

// Abort(false) is allowed only after proving that no durable mutation occurred.
// Unknown commit outcomes quarantine until an external authoritative recovery
// constructs a new publisher; this model supplies no optimistic reset operation.
func (p *Publisher) Abort(ticket uint64, uncertain bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	old := p.view.Load()
	if ticket == 0 || ticket != old.ticket || old.quarantined {
		return errors.New("invalid publication ticket")
	}
	next := *old
	next.ticket = 0
	next.quarantined = uncertain
	p.view.Store(&next)
	return nil
}
