package libravdbstore

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchvalidity"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type preparationV44 struct {
	Kind                             string
	OperationIDs                     []string
	Begin, OwnerAt, End, BuiltAt     time.Time
	Owner, Built, Reused, InitialNil bool
	Head                             model.Snapshot
	Root, Encoded, Error             string
}

// V44 preserves V42's optional AFTER-full-durability publication policy, but
// traces each preparation directly. Before/after aggregate counter deltas are
// unsafe under overlapping journal/write/outcome calls and are not used here.
type eagerStoreV44 struct {
	*warmStoreV37
	enabled       bool
	preparationMu sync.Mutex
	preparations  []preparationV44
}

func (s *eagerStoreV44) refresh(ctx context.Context, kind string, ids ...string) (err error) {
	if !s.enabled {
		return nil
	}
	trace := preparationV44{Kind: kind, OperationIDs: append([]string(nil), ids...), Begin: time.Now().UTC()}
	defer func() {
		trace.End = time.Now().UTC()
		if err != nil {
			trace.Error = err.Error()
		}
		s.preparationMu.Lock()
		s.preparations = append(s.preparations, trace)
		s.preparationMu.Unlock()
	}()
	if err = ctx.Err(); err != nil {
		return err
	}
	a := s.authority
	a.owner.Lock()
	trace.OwnerAt = time.Now().UTC()
	trace.Owner = true
	defer a.owner.Unlock()
	if err = ctx.Err(); err != nil {
		return err
	}
	native := a.gate.store.Snapshot(ctx)
	view := a.current.Load()
	if a.poison.Load() || a.state == nil || a.state.Head != native || view == nil || view.snapshot != native {
		return store.ErrStaleSnapshot
	}
	trace.Head = a.state.Head
	if a.state.Certificate == nil {
		trace.InitialNil = true
		return nil
	}
	if c := s.core.Load(); c != nil && c.state.Head == native && c.state.Certificate != nil {
		trace.Reused = true
		trace.Root = c.state.Hash
		return nil
	}
	state := a.state
	read := struct {
		Head        model.Snapshot
		Hash        string
		Log         []researchvalidity.Mutation
		Sources     map[string]witnessSourceV23
		Certificate *witnessBindingV23
	}{state.Head, state.Hash, state.Log, state.Sources, state.Certificate}
	encoded, err := json.Marshal(read)
	if err != nil {
		return err
	}
	var owned witnessStateV23
	if err = json.Unmarshal(encoded, &owned); err != nil {
		return err
	}
	if a.poison.Load() || a.gate.store.Snapshot(ctx) != owned.Head || a.current.Load() != view {
		return store.ErrStaleSnapshot
	}
	c := &readCoreV36{state: &owned, encoded: string(encoded), at: time.Now().UTC()}
	s.core.Store(c)
	s.builds.Add(1)
	s.physicalBytes.Add(int64(len(encoded)))
	s.copiesBytes.Add(int64(len(encoded)))
	s.coreStoreV36.mu.Lock()
	s.proofs = append(s.proofs, coreBuildV36{Root: owned.Hash, Encoded: c.encoded, At: c.at})
	s.coreStoreV36.mu.Unlock()
	trace.Built = true
	trace.Root = owned.Hash
	trace.Encoded = c.encoded
	trace.BuiltAt = c.at
	return nil
}

func (s *eagerStoreV44) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if err := s.warmStoreV37.PutBayesianJournal(ctx, e); err != nil {
		return err
	}
	_ = s.refresh(ctx, "journal", e.ID) // Optional acceleration cannot revoke a durable commit.
	return nil
}

func (s *eagerStoreV44) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, key, parent, digest string, weight float64, change bayes.ChangePolicy, group bayes.GroupPolicy, observation model.ResidualObservation, policy residual.Policy) (store.BayesianOutcomeResult, error) {
	r, err := s.warmStoreV37.ApplyBayesianOutcome(ctx, request, key, parent, digest, weight, change, group, observation, policy)
	if err == nil {
		_ = s.refresh(ctx, "outcome", request.IdempotencyKey)
	}
	return r, err
}

type loadEagerV44 struct {
	*loadWarmV37
	eager *eagerStoreV44
}

func attachLoadEagerV44(t *testing.T, f *witnessFixtureV23, mode int, enabled bool) *loadEagerV44 {
	t.Helper()
	if mode != -2 && mode != 4 {
		t.Fatal("eager load requires joined/archive with warm Search enabled")
	}
	inner := attachLoadWarmV37(t, f, mode)
	s := &eagerStoreV44{warmStoreV37: inner.warm, enabled: enabled}
	before := s.authority.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil || before != s.authority.gate.store.Snapshot(context.Background()) {
		t.Fatal("eager attachment changed authority", err)
	}
	f.svc = svc
	return &loadEagerV44{inner, s}
}
func (s *loadEagerV44) append(ctx context.Context, writes []ResearchEventWrite, interrupt bool) (model.Snapshot, error) {
	r, err := s.loadWarmV37.append(ctx, writes, interrupt)
	if err == nil {
		ids := make([]string, len(writes))
		for i, write := range writes {
			ids[i] = write.Event.ID
		}
		_ = s.eager.refresh(ctx, "write", ids...)
	}
	return r, err
}

type eagerTrialV44 struct {
	warmTrialV37
	EagerEnabled bool
	Preparations []preparationV44
}

func finishEagerTrialV44(s *loadEagerV44, r joinedTrialV25) eagerTrialV44 {
	s.eager.preparationMu.Lock()
	traces := append([]preparationV44(nil), s.eager.preparations...)
	s.eager.preparationMu.Unlock()
	return eagerTrialV44{warmTrialV37: finishWarmTrialV37(s.loadWarmV37, r), EagerEnabled: s.eager.enabled, Preparations: traces}
}

var _ store.EventStore = (*eagerStoreV44)(nil)
