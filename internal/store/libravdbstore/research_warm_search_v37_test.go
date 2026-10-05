package libravdbstore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type warmKeyV37 struct{}
type warmUseV37 struct {
	Journal, Root string
	Fast          bool
	CheckedAt     time.Time
}
type warmStatsV37 struct{ Fast, Cold int64 }
type warmStoreV37 struct {
	*coreStoreV36
	enabled    bool
	fast, cold atomic.Int64
	mu         sync.Mutex
	uses       []warmUseV37
}

// Only a committed immutable core at the current native/published head can
// replace the preliminary owner-protected Head check. Native nomination and
// every metadata getter still run. Missing/gapped heads take the sealed path.
func (s *warmStoreV37) Search(ctx context.Context, tenant string, vector []float32, asOf time.Time, k int) ([]store.SearchResult, error) {
	u, _ := ctx.Value(warmKeyV37{}).(*warmUseV37)
	if u == nil {
		return nil, errors.New("missing warm-search request")
	}
	c := s.core.Load()
	u.CheckedAt = time.Now().UTC()
	current := s.authority.current.Load()
	native := s.authority.gate.store.Snapshot(ctx)
	if !s.enabled || c == nil || current == nil || c.state.Head != native || current.snapshot != native || s.authority.poison.Load() {
		s.cold.Add(1)
		return s.coreStoreV36.Search(ctx, tenant, vector, asOf, k)
	}
	p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35)
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if p == nil || p.owner != s.projectionStoreV35 || scope == nil || p.scope != nil && !p.active.Load() {
		return nil, errors.New("foreign or expired warm-search request")
	}
	results, err := s.authority.publishedRecallLoadStoreV16.Search(ctx, tenant, vector, asOf, k)
	if err != nil {
		return nil, err
	}
	// These are exactly the sealed Search binding operations, derived from actual
	// nomination rather than the caller's requested k or a guessed candidate set.
	scope.Vector = witnessHashV23(vector)
	scope.AsOf = asOf
	scope.Snapshot = s.authority.Snapshot(ctx)
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.Event.ID
	}
	scope.Frontier = witnessIDsV23(ids)
	view, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if view == nil || view.view == nil || scope.Snapshot != c.state.Head || view.view.snapshot != c.state.Head || s.authority.poison.Load() || s.authority.current.Load() == nil || s.authority.gate.store.Snapshot(ctx) != c.state.Head {
		return nil, store.ErrStaleSnapshot
	}
	p.state, p.scope, p.encoded, p.capturedAt = c.state, scope, c.encoded, time.Now().UTC()
	p.active.Store(true)
	s.hits.Add(1)
	s.captured.Add(1)
	s.fast.Add(1)
	u.Fast = true
	if s.afterProjection != nil {
		s.afterProjection(ctx)
	}
	return results, nil
}
func (s *warmStoreV37) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	u, _ := ctx.Value(warmKeyV37{}).(*warmUseV37)
	p, _, err := s.pinV35(ctx)
	if err != nil {
		return err
	}
	if u == nil || p == nil || p.state == nil {
		return errors.New("missing warm-search handoff")
	}
	u.Journal, u.Root = e.ID, p.state.Hash
	s.mu.Lock()
	s.uses = append(s.uses, *u)
	s.mu.Unlock()
	// The original request token expires and the original full durability path
	// completes below. This wrapper cannot manufacture an early serving receipt.
	return s.projectionStoreV35.PutBayesianJournal(ctx, e)
}

type loadWarmV37 struct {
	*loadCoreV36
	warm *warmStoreV37
}

func attachLoadWarmV37(t *testing.T, f *witnessFixtureV23, mode int) *loadWarmV37 {
	t.Helper()
	enabled := mode == -2 || mode == 4
	coreMode := -2
	if mode == 3 || mode == 4 {
		coreMode = 4
	} else if mode != -1 && mode != -2 {
		t.Fatal("invalid warm-search factorial")
	}
	inner := attachLoadCoreV36(t, f, coreMode)
	w := &warmStoreV37{coreStoreV36: inner.shared, enabled: enabled}
	before := w.authority.gate.store.Snapshot(context.Background())
	svc, err := service.New(w, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil || before != w.authority.gate.store.Snapshot(context.Background()) {
		t.Fatal("warm attachment", err)
	}
	f.svc = svc
	return &loadWarmV37{loadCoreV36: inner, warm: w}
}
func (s *loadWarmV37) recall(ctx context.Context, svc *service.Service, r model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	return s.loadCoreV36.recall(context.WithValue(ctx, warmKeyV37{}, &warmUseV37{}), svc, r)
}
func (s *loadWarmV37) closeV37() error { return s.closeV36() }

type warmTrialV37 struct {
	coreTrialV36
	WarmEnabled bool
	WarmStats   warmStatsV37
	WarmUses    []warmUseV37
}

func finishWarmTrialV37(s *loadWarmV37, r joinedTrialV25) warmTrialV37 {
	w := s.warm
	w.mu.Lock()
	uses := append([]warmUseV37(nil), w.uses...)
	w.mu.Unlock()
	return warmTrialV37{coreTrialV36: finishCoreTrialV36(s.loadCoreV36, r), WarmEnabled: w.enabled, WarmStats: warmStatsV37{Fast: w.fast.Load(), Cold: w.cold.Load()}, WarmUses: uses}
}

var _ store.EventStore = (*warmStoreV37)(nil)
