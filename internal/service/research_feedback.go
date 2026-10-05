package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type researchFeedbackKey struct{ journal, event string }

// ResearchFeedbackBridge is a bounded, in-process research consumer. Callers
// supply externally verified usefulness; this API does NOT authenticate truth.
// One fixed dependency snapshot/tenant is allowed for the entire lifetime.
// It cannot alter served rankings, and never treats an absent label as false.
type ResearchFeedbackBridge struct {
	mu         sync.Mutex
	service    *Service
	tenant     string
	snapshot   model.Snapshot
	worker     *researchmemory.Background
	seen       map[string]bool
	pending    map[researchFeedbackKey]uint64
	closed     bool
	temporal   bool
	latestAsOf time.Time
}

func NewResearchFeedbackBridge(s *Service, tenant string, seed int64) (*ResearchFeedbackBridge, error) {
	return newResearchFeedbackBridge(s, tenant, seed, false)
}

// Temporal reuse is opt-in and requires the store's bounded, fail-closed motion
// proof. A future arrival must be later than EVERY admitted query, not merely
// the query attached to the next feedback label.
func NewResearchTemporalFeedbackBridge(s *Service, tenant string, seed int64) (*ResearchFeedbackBridge, error) {
	return newResearchFeedbackBridge(s, tenant, seed, true)
}

func newResearchFeedbackBridge(s *Service, tenant string, seed int64, temporal bool) (*ResearchFeedbackBridge, error) {
	if s == nil || tenant == "" {
		return nil, errors.New("invalid research bridge")
	}
	w, err := researchmemory.NewBackground(1, seed, 256)
	if err != nil {
		return nil, err
	}
	return &ResearchFeedbackBridge{service: s, tenant: tenant, snapshot: s.store.Snapshot(context.Background()), worker: w, seen: make(map[string]bool), pending: make(map[researchFeedbackKey]uint64), temporal: temporal}, nil
}

func (b *ResearchFeedbackBridge) compatible(ctx context.Context, captured model.Snapshot, at time.Time) bool {
	if ctx.Err() != nil {
		return false
	}
	// A research-only store adapter may supply one coherent publication proof.
	// Do not acquire the ordinary snapshot lock before asking that adapter.
	if b.temporal && !at.IsZero() {
		if db, ok := b.service.store.(interface {
			ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
		}); ok {
			return db.ResearchPublicationCompatible(ctx, captured, at)
		}
	}
	if b.service.store.Snapshot(ctx) == captured {
		return true
	}
	if !b.temporal || at.IsZero() {
		return false
	}
	db, ok := b.service.store.(interface {
		ResearchSnapshotCompatible(context.Context, model.Snapshot, time.Time) bool
	})
	return ok && db.ResearchSnapshotCompatible(ctx, captured, at)
}

func (b *ResearchFeedbackBridge) cutoff(at time.Time) time.Time {
	if b.latestAsOf.After(at) {
		return b.latestAsOf
	}
	return at
}

func (b *ResearchFeedbackBridge) valid(ctx context.Context, at time.Time) bool {
	return !b.closed && b.compatible(ctx, b.snapshot, b.cutoff(at))
}

// Admit accepts trusted tap output after its journal commit. The durable journal
// validates event membership/baselines, not the lexical features themselves.
// The caller must not construct or modify tap features from untrusted input.
// Saturation fails closed rather than evicting replay tombstones.
func (b *ResearchFeedbackBridge) Admit(ctx context.Context, in ResearchFrontierObservation) ([]researchmemory.Prediction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.valid(ctx, in.AsOf) || !b.compatible(ctx, in.Snapshot, b.cutoff(in.AsOf)) || in.JournalID == "" || b.seen[in.JournalID] || len(b.seen) >= 256 || len(in.Candidates) == 0 || len(in.Candidates) > 200 || len(b.pending)+len(in.Candidates) > 256 {
		return nil, errors.New("stale, repeated or saturated research observation")
	}
	j, err := b.service.store.GetBayesianJournal(ctx, b.tenant, in.JournalID)
	if err != nil {
		return nil, err
	}
	if j.TenantID != b.tenant || j.Snapshot != in.Snapshot || !j.AsOf.Equal(in.AsOf) {
		return nil, errors.New("research journal binding mismatch")
	}
	baselines := make(map[string]float64, len(j.Report.Decisions))
	for _, d := range j.Report.Decisions {
		baselines[d.EventID] = d.Forecast.PreResidualLaw.Useful
	}
	unique := make(map[string]bool, len(in.Candidates))
	for _, c := range in.Candidates {
		base, ok := baselines[c.EventID]
		if !ok || c.EventID == "" || unique[c.EventID] || base != c.Baseline {
			return nil, errors.New("research candidate binding mismatch")
		}
		unique[c.EventID] = true
	}
	var predictions []researchmemory.Prediction
	rollback := func() {
		for _, p := range predictions {
			b.worker.Discard(p.ID)
		}
	}
	for _, c := range in.Candidates {
		p, e := b.worker.Predict(c.Features, c.Baseline, 1, in.AsOf)
		if e != nil {
			rollback()
			return nil, e
		}
		predictions = append(predictions, p)
	}
	if !b.valid(ctx, in.AsOf) || !b.compatible(ctx, in.Snapshot, b.cutoff(in.AsOf)) {
		rollback()
		return nil, errors.New("research dependencies changed during admission")
	}
	for i, c := range in.Candidates {
		b.pending[researchFeedbackKey{in.JournalID, c.EventID}] = predictions[i].ID
	}
	b.seen[in.JournalID] = true
	b.latestAsOf = b.cutoff(in.AsOf)
	return predictions, nil
}

func (b *ResearchFeedbackBridge) Feedback(ctx context.Context, journal, event string, useful bool, available time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := researchFeedbackKey{journal, event}
	id, ok := b.pending[key]
	if !b.valid(ctx, b.latestAsOf) || !ok {
		return errors.New("stale or unbound research feedback")
	}
	if err := b.worker.Feedback(id, useful, 1, available); err != nil {
		return err
	}
	delete(b.pending, key)
	return nil
}

func (b *ResearchFeedbackBridge) Discard(journal, event string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := researchFeedbackKey{journal, event}
	if id, ok := b.pending[key]; ok {
		b.worker.Discard(id)
		delete(b.pending, key)
	}
}

// Score checks the dependency version both sides of an immutable score. It is
// still research output, not authorization to publish it into the served law.
func (b *ResearchFeedbackBridge) Score(ctx context.Context, features uint16, baseline float64, at time.Time) (float64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.valid(ctx, at) {
		return 0, errors.New("stale research bridge")
	}
	p, err := b.worker.Snapshot().Score(features, baseline, 1, at)
	if err != nil {
		return 0, err
	}
	if !b.valid(ctx, at) {
		return 0, errors.New("research dependencies changed during score")
	}
	return p, nil
}

func (b *ResearchFeedbackBridge) Close() {
	b.mu.Lock()
	b.closed = true
	clear(b.pending)
	b.mu.Unlock()
	b.worker.Close()
}
