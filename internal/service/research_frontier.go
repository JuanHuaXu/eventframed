package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type ResearchFrontierCandidate struct {
	EventID  string
	Features uint16
	Baseline float64
}

// ResearchFrontierObservation is a committed pre-packing diagnostic, not proof
// the caller or an agent consumed those candidates. IDs bind explicit feedback;
// they must not be exported as a public dataset without a separate privacy audit.
type ResearchFrontierObservation struct {
	JournalID  string
	Snapshot   model.Snapshot
	AsOf       time.Time
	Candidates []ResearchFrontierCandidate
	queuedAt   time.Time
}

// ResearchFrontierTap provides a bounded nonblocking handoff. The owner manages
// its lifetime and any external learner. It confers no serving/model authority.
type ResearchFrontierTap struct {
	mu      sync.Mutex
	tenant  string
	queue   chan ResearchFrontierObservation
	closed  bool
	dropped uint64
}

func NewResearchFrontierTap(tenant string, capacity int) (*ResearchFrontierTap, error) {
	if tenant == "" || capacity < 1 || capacity > 256 {
		return nil, errors.New("invalid research frontier tap")
	}
	return &ResearchFrontierTap{tenant: tenant, queue: make(chan ResearchFrontierObservation, capacity)}, nil
}

func (p *ResearchFrontierTap) Take(ctx context.Context) (ResearchFrontierObservation, error) {
	if err := ctx.Err(); err != nil {
		return ResearchFrontierObservation{}, err
	}
	select {
	case <-ctx.Done():
		return ResearchFrontierObservation{}, ctx.Err()
	case in, ok := <-p.queue:
		if !ok {
			return ResearchFrontierObservation{}, errors.New("research frontier tap closed")
		}
		return in, nil
	}
}

func (p *ResearchFrontierTap) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		p.closed = true
		close(p.queue)
	}
}
func (p *ResearchFrontierTap) Dropped() uint64 { p.mu.Lock(); defer p.mu.Unlock(); return p.dropped }

func (s *Service) tapResearchFrontier(request model.RecallRequest, journal model.BayesianJournalEntry, candidates []model.Candidate) {
	p := s.config.ResearchFrontier
	if p == nil || p.tenant != request.TenantID {
		return
	}
	// Backlog is not a reason to spend serving time preparing work we cannot
	// accept. Admission is checked again below because callers can race here.
	p.mu.Lock()
	if p.closed || len(p.queue) == cap(p.queue) {
		p.dropped++
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()
	// No whole text, attributes or embedding copies cross the research boundary.
	in := ResearchFrontierObservation{JournalID: journal.ID, Snapshot: journal.Snapshot, AsOf: request.AsOf}
	valid := len(candidates) <= 200
	if valid {
		in.Candidates = make([]ResearchFrontierCandidate, len(candidates))
		for i, c := range candidates {
			baseline := c.Forecast.PreResidualLaw.Useful
			features, err := researchmemory.Extract(request.Query, c.Event, baseline)
			if err != nil {
				valid = false
				break
			}
			in.Candidates[i] = ResearchFrontierCandidate{c.Event.ID, features, baseline}
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || !valid {
		p.dropped++
		return
	}
	select {
	case p.queue <- stampedResearchFrontier(in):
	default:
		p.dropped++
	}
}

func stampedResearchFrontier(in ResearchFrontierObservation) ResearchFrontierObservation {
	in.queuedAt = time.Now()
	return in
}
