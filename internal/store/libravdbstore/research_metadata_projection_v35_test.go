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
	"sort"
)

type metadataKeyV35 struct{}
type metadataPinV35 struct {
	owner      *projectionStoreV35
	state      *witnessStateV23
	active     atomic.Bool
	encoded    string
	capturedAt time.Time
	scope      *witnessScopeV23
	mu         sync.Mutex
	getters    []metadataGetterV35
}
type metadataGetterV35 struct {
	Method, Key string
	NativeOK    bool
}
type metadataReadV35 struct {
	Journal string
	Getters []metadataGetterV35
}
type metadataProofV35 struct {
	Journal, Root, SHA256, Encoded string
	Snapshot                       model.Snapshot
	CapturedAt                     time.Time
}
type metadataStatsV35 struct {
	Captured, ProjectedGets, NativeSuccess, LegacyOwnerResults int64
	CopiesBytes                                                int64
}
type projectionStoreV35 struct {
	store.EventStore
	authority                                                               *witnessStoreV23
	projected                                                               bool
	mu                                                                      sync.Mutex
	proofs                                                                  []metadataProofV35
	reads                                                                   []metadataReadV35
	captured, projectedGets, nativeSuccess, legacyOwnerResults, copiesBytes atomic.Int64
	afterProjection                                                         func(context.Context) // fixed technical-test callback only
}

// Copy only the read-relevant witness fields once under owner. Native getters
// still read their original records and compare them to this owned projection.
func (s *projectionStoreV35) Search(ctx context.Context, tenant string, vector []float32, asOf time.Time, k int) ([]store.SearchResult, error) {
	results, err := s.authority.Search(ctx, tenant, vector, asOf, k)
	if err != nil || !s.projected {
		return results, err
	}
	p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35)
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	view, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if p == nil || p.owner != s || scope == nil || view == nil || view.view == nil {
		return nil, errors.New("missing metadata request authority")
	}
	s.authority.owner.Lock()
	if s.authority.poison.Load() || s.authority.state.Head != scope.Snapshot || scope.Snapshot != view.view.snapshot || s.authority.gate.store.Snapshot(ctx) != scope.Snapshot {
		s.authority.owner.Unlock()
		return nil, store.ErrStaleSnapshot
	}
	state := s.authority.state
	read := struct {
		Head        model.Snapshot
		Hash        string
		Log         []researchvalidity.Mutation
		Sources     map[string]witnessSourceV23
		Certificate *witnessBindingV23
	}{state.Head, state.Hash, state.Log, state.Sources, state.Certificate}
	encoded, err := json.Marshal(read)
	if err == nil {
		var owned witnessStateV23
		err = json.Unmarshal(encoded, &owned)
		if err == nil {
			p.state, p.scope, p.encoded, p.capturedAt = &owned, scope, string(encoded), time.Now().UTC()
			p.active.Store(true)
		}
	}
	s.authority.owner.Unlock()
	if err != nil {
		return nil, err
	}
	s.captured.Add(1)
	s.copiesBytes.Add(int64(len(encoded)))
	if s.afterProjection != nil {
		s.afterProjection(ctx)
	}
	return results, nil
}

func (s *projectionStoreV35) pinV35(ctx context.Context) (*metadataPinV35, bool, error) {
	if !s.projected {
		return nil, false, nil
	}
	p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35)
	// Non-Recall maintenance/feedback calls retain sealed ordinary behavior.
	if p == nil {
		return nil, false, nil
	}
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	view, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if p.owner != s || !p.active.Load() || p.state == nil || p.scope != scope || scope == nil || scope.Snapshot != p.state.Head || view == nil || view.view == nil || view.view.snapshot != p.state.Head || s.authority.poison.Load() || s.authority.current.Load() == nil || time.Since(view.view.at) >= 250*time.Millisecond || s.authority.gate.store.Snapshot(ctx) != p.state.Head {
		return nil, true, errors.New("invalid or expired metadata projection")
	}
	return p, true, nil
}

// Same external validity predicate as sealed V23; no self-certification, no
// posterior/certificate durable retagging. This private view is request-local.
func (s *projectionStoreV35) compatibleV35(ctx context.Context, pin *metadataPinV35, record researchvalidity.Record, b witnessBindingV23) bool {
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if !s.authority.enabled || s.authority.poison.Load() || scope == nil || scope.Binding != b.Binding || scope.Vector != b.Vector || scope.Frontier != b.Frontier {
		return false
	}
	if scope.Snapshot != pin.state.Head || s.authority.gate.store.Snapshot(ctx) != pin.state.Head {
		return false
	}
	start := sort.Search(len(pin.state.Log), func(i int) bool { return pin.state.Log[i].RuntimeVersion > record.Base.RuntimeVersion })
	req := researchvalidity.Request{TenantID: record.TenantID, PosteriorKey: record.PosteriorKey, QueryDigest: scope.Binding, ModelID: scope.Vector, HorizonKey: record.HorizonKey, SourceID: record.SourceID, AsOf: scope.AsOf, FrontierDigest: scope.Frontier, Current: scope.Snapshot}
	return researchvalidity.Check(record, req, pin.state.Log[start:]).Status == researchvalidity.Compatible
}

func (s *projectionStoreV35) markOwnerResultV35(err error) {
	if err == nil {
		s.legacyOwnerResults.Add(1)
	}
}
func (s *projectionStoreV35) traceGetterV35(ctx context.Context, method, key string, err error) {
	p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35)
	if p == nil {
		return
	}
	p.mu.Lock()
	p.getters = append(p.getters, metadataGetterV35{Method: method, Key: key, NativeOK: err == nil})
	p.mu.Unlock()
}
func (s *projectionStoreV35) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if p, _ := ctx.Value(metadataKeyV35{}).(*metadataPinV35); p != nil {
		p.mu.Lock()
		getters := append([]metadataGetterV35(nil), p.getters...)
		p.mu.Unlock()
		s.mu.Lock()
		s.reads = append(s.reads, metadataReadV35{Journal: e.ID, Getters: getters})
		s.mu.Unlock()
	}
	if s.projected {
		p, selected, err := s.pinV35(ctx)
		if err != nil {
			return err
		}
		if !selected {
			return errors.New("journal lacks projected Recall authority")
		}
		if p.state.Head != e.Snapshot {
			return store.ErrStaleSnapshot
		}
		s.mu.Lock()
		s.proofs = append(s.proofs, metadataProofV35{Journal: e.ID, Root: p.state.Hash, SHA256: witnessHashV23(json.RawMessage(p.encoded)), Encoded: p.encoded, Snapshot: e.Snapshot, CapturedAt: p.capturedAt})
		s.mu.Unlock()
		// Expire BEFORE forwarding handoff. Capture/archival persistence uses its
		// own sealed authority, and projection getters cannot outlive the read phase.
		p.active.Store(false)
	}
	return s.EventStore.PutBayesianJournal(ctx, e)
}

type loadProjectionV35 struct {
	*loadArchiveV34
	metadata *projectionStoreV35
}

func attachLoadProjectionV35(t *testing.T, f *witnessFixtureV23, mode int) *loadProjectionV35 {
	t.Helper()
	projected := mode == -2 || mode == 4
	baseMode := mode
	if mode == -2 {
		baseMode = -1
	}
	if mode == 4 {
		baseMode = 3
	}
	if baseMode != -1 && baseMode != 3 {
		t.Fatal("invalid projection/storage factorial")
	}
	inner := attachLoadArchiveV34(t, f, baseMode)
	var backend store.EventStore = inner.joinedWitnessV25
	if inner.archive != nil {
		backend = inner.archive
	}
	m := &projectionStoreV35{EventStore: backend, authority: inner.witnessStoreV23, projected: projected}
	before := m.authority.gate.store.Snapshot(context.Background())
	svc, err := service.New(m, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != m.authority.gate.store.Snapshot(context.Background()) {
		t.Fatal("projection attachment moved native head")
	}
	f.svc = svc
	return &loadProjectionV35{loadArchiveV34: inner, metadata: m}
}
func (s *loadProjectionV35) recall(ctx context.Context, svc *service.Service, r model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	p := &metadataPinV35{owner: s.metadata}
	ctx = context.WithValue(ctx, metadataKeyV35{}, p)
	defer p.active.Store(false)
	return s.loadArchiveV34.recall(ctx, svc, r)
}
func (s *loadProjectionV35) closeV35() error { return s.closeV34() }

type projectionTrialV35 struct {
	archiveTrialV34
	Projected      bool
	Metadata       metadataStatsV35
	MetadataProofs []metadataProofV35
	MetadataReads  []metadataReadV35
}

func finishProjectionTrialV35(s *loadProjectionV35, r joinedTrialV25) projectionTrialV35 {
	m := s.metadata
	m.mu.Lock()
	proofs := append([]metadataProofV35(nil), m.proofs...)
	reads := append([]metadataReadV35(nil), m.reads...)
	m.mu.Unlock()
	stats := metadataStatsV35{Captured: m.captured.Load(), ProjectedGets: m.projectedGets.Load(), NativeSuccess: m.nativeSuccess.Load(), LegacyOwnerResults: m.legacyOwnerResults.Load(), CopiesBytes: m.copiesBytes.Load()}
	return projectionTrialV35{archiveTrialV34: finishArchiveTrialV34(s.loadArchiveV34, r), Projected: m.projected, Metadata: stats, MetadataProofs: proofs, MetadataReads: reads}
}
