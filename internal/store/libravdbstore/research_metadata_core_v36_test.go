package libravdbstore

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchvalidity"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type readCoreV36 struct {
	state   *witnessStateV23
	encoded string
	at      time.Time
}
type coreBuildV36 struct {
	Root, Encoded string
	At            time.Time
}
type coreStatsV36 struct {
	Builds, Hits, OwnerAcquisitions, PhysicalBytes int64
}
type coreStoreV36 struct {
	*projectionStoreV35
	enabled                             bool
	core                                atomic.Pointer[readCoreV36]
	builds, hits, owners, physicalBytes atomic.Int64
	mu                                  sync.Mutex
	proofs                              []coreBuildV36
}

// A core is deep-copied, then never modified. It may survive journal-only
// publications, not runtime-head motion. The nil initial certificate is never
// cached: its first binding changes read semantics without moving runtime head.
func (s *coreStoreV36) Search(ctx context.Context, tenant string, vector []float32, asOf time.Time, k int) ([]store.SearchResult, error) {
	if !s.enabled {
		return s.projectionStoreV35.Search(ctx, tenant, vector, asOf, k)
	}
	results, err := s.authority.Search(ctx, tenant, vector, asOf, k)
	if err != nil {
		return nil, err
	}
	p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35)
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	view, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if p == nil || p.owner != s.projectionStoreV35 || scope == nil || view == nil || view.view == nil || scope.Snapshot != view.view.snapshot || s.authority.poison.Load() || s.authority.gate.store.Snapshot(ctx) != scope.Snapshot {
		return nil, errors.New("invalid shared-core request authority")
	}
	c := s.core.Load()
	if c == nil || c.state.Head != scope.Snapshot {
		s.authority.owner.Lock()
		s.owners.Add(1)
		if s.authority.poison.Load() || s.authority.state.Head != scope.Snapshot || s.authority.gate.store.Snapshot(ctx) != scope.Snapshot {
			s.authority.owner.Unlock()
			return nil, store.ErrStaleSnapshot
		}
		// Concurrent readers may have published this exact head while we waited.
		c = s.core.Load()
		if c == nil || c.state.Head != scope.Snapshot {
			state := s.authority.state
			read := struct {
				Head        model.Snapshot
				Hash        string
				Log         []researchvalidity.Mutation
				Sources     map[string]witnessSourceV23
				Certificate *witnessBindingV23
			}{state.Head, state.Hash, state.Log, state.Sources, state.Certificate}
			encoded, e := json.Marshal(read)
			if e != nil {
				s.authority.owner.Unlock()
				return nil, e
			}
			var owned witnessStateV23
			if e = json.Unmarshal(encoded, &owned); e != nil {
				s.authority.owner.Unlock()
				return nil, e
			}
			c = &readCoreV36{state: &owned, encoded: string(encoded), at: time.Now().UTC()}
			s.builds.Add(1)
			s.physicalBytes.Add(int64(len(encoded)))
			s.copiesBytes.Add(int64(len(encoded)))
			s.mu.Lock()
			s.proofs = append(s.proofs, coreBuildV36{Root: owned.Hash, Encoded: c.encoded, At: c.at})
			s.mu.Unlock()
			if owned.Certificate != nil {
				s.core.Store(c)
			}
		} else {
			s.hits.Add(1)
		}
		s.authority.owner.Unlock()
	} else {
		s.hits.Add(1)
	}
	p.state, p.scope, p.encoded, p.capturedAt = c.state, scope, c.encoded, time.Now().UTC()
	p.active.Store(true)
	s.captured.Add(1)
	if s.afterProjection != nil {
		s.afterProjection(ctx)
	}
	return results, nil
}

type loadCoreV36 struct {
	*loadProjectionV35
	shared *coreStoreV36
}

func attachLoadCoreV36(t *testing.T, f *witnessFixtureV23, mode int) *loadCoreV36 {
	t.Helper()
	enabled := mode == -2 || mode == 4
	storageMode := -2
	if mode == 3 || mode == 4 {
		storageMode = 4
	} else if mode != -1 && mode != -2 {
		t.Fatal("invalid core factorial")
	}
	inner := attachLoadProjectionV35(t, f, storageMode)
	s := &coreStoreV36{projectionStoreV35: inner.metadata, enabled: enabled}
	before := s.authority.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil || before != s.authority.gate.store.Snapshot(context.Background()) {
		t.Fatal("shared-core attachment", err)
	}
	f.svc = svc
	return &loadCoreV36{loadProjectionV35: inner, shared: s}
}
func (s *loadCoreV36) closeV36() error { return s.closeV35() }

type coreTrialV36 struct {
	projectionTrialV35
	CoreEnabled bool
	CoreStats   coreStatsV36
	CoreBuilds  []coreBuildV36
}

func finishCoreTrialV36(s *loadCoreV36, r joinedTrialV25) coreTrialV36 {
	c := s.shared
	c.mu.Lock()
	builds := append([]coreBuildV36(nil), c.proofs...)
	c.mu.Unlock()
	stats := coreStatsV36{Builds: c.builds.Load(), Hits: c.hits.Load(), OwnerAcquisitions: c.owners.Load(), PhysicalBytes: c.physicalBytes.Load()}
	return coreTrialV36{projectionTrialV35: finishProjectionTrialV35(s.loadProjectionV35, r), CoreEnabled: c.enabled, CoreStats: stats, CoreBuilds: builds}
}
