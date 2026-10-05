package libravdbstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchvalidity"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type eagerStoreV42 struct {
	*warmStoreV37
	eager                                 bool
	attempts, prepared, preparationErrors atomic.Int64
	preparationOwners, preparationNS      atomic.Int64
}

// This optional acceleration runs AFTER the delegated durable operation.
// A successful native operation must not become a falsely uncommitted result
// merely because acceleration fails; warm Search still checks every head.
func (s *eagerStoreV42) refresh(ctx context.Context) error {
	if !s.eager {
		return nil
	}
	begin := time.Now()
	defer func() { s.preparationNS.Add(time.Since(begin).Nanoseconds()) }()
	s.attempts.Add(1)
	if e := ctx.Err(); e != nil {
		return e
	}
	a := s.authority
	a.owner.Lock()
	s.preparationOwners.Add(1)
	defer a.owner.Unlock()
	if e := ctx.Err(); e != nil {
		return e
	}
	native := a.gate.store.Snapshot(ctx)
	view := a.current.Load()
	if a.poison.Load() || a.state == nil || a.state.Head != native || view == nil || view.snapshot != native {
		return store.ErrStaleSnapshot
	}
	// Initial certificate binding changes read semantics without head motion.
	// Subsequent journal-only transitions change Hash/Journals, not read fields.
	if a.state.Certificate == nil {
		return nil
	}
	if c := s.core.Load(); c != nil && c.state.Head == native && c.state.Certificate != nil {
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
	encoded, e := json.Marshal(read)
	if e != nil {
		return e
	}
	var owned witnessStateV23
	if e = json.Unmarshal(encoded, &owned); e != nil {
		return e
	}
	// Owner excludes witness mutation; native/publication equality also catches
	// unwitnessed writes. Never publish speculative state to make a gate pass.
	if a.poison.Load() || a.gate.store.Snapshot(ctx) != owned.Head || a.current.Load() != view {
		return store.ErrStaleSnapshot
	}
	c := &readCoreV36{state: &owned, encoded: string(encoded), at: time.Now().UTC()}
	s.core.Store(c)
	s.prepared.Add(1)
	s.builds.Add(1)
	s.physicalBytes.Add(int64(len(encoded)))
	s.copiesBytes.Add(int64(len(encoded)))
	s.coreStoreV36.mu.Lock()
	s.proofs = append(s.proofs, coreBuildV36{Root: owned.Hash, Encoded: c.encoded, At: c.at})
	s.coreStoreV36.mu.Unlock()
	return nil
}
func (s *eagerStoreV42) afterDurability(ctx context.Context) {
	if e := s.refresh(ctx); e != nil {
		s.preparationErrors.Add(1)
	}
}
func (s *eagerStoreV42) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if err := s.warmStoreV37.PutBayesianJournal(ctx, e); err != nil {
		return err
	}
	s.afterDurability(ctx)
	return nil
}
func (s *eagerStoreV42) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, key, parent, digest string, weight float64, change bayes.ChangePolicy, group bayes.GroupPolicy, observation model.ResidualObservation, policy residual.Policy) (store.BayesianOutcomeResult, error) {
	r, e := s.warmStoreV37.ApplyBayesianOutcome(ctx, request, key, parent, digest, weight, change, group, observation, policy)
	if e == nil {
		s.afterDurability(ctx)
	}
	return r, e
}

type loadEagerV42 struct {
	*loadWarmV37
	eager *eagerStoreV42
}

func attachEagerV42(t *testing.T, f *witnessFixtureV23, enabled bool) *loadEagerV42 {
	t.Helper()
	inner := attachLoadWarmV37(t, f, 4)
	s := &eagerStoreV42{warmStoreV37: inner.warm, eager: enabled}
	before := s.authority.gate.store.Snapshot(context.Background())
	svc, e := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if e != nil || before != s.authority.gate.store.Snapshot(context.Background()) {
		t.Fatal("eager attachment changed authority", e)
	}
	f.svc = svc
	return &loadEagerV42{inner, s}
}
func (s *loadEagerV42) append(ctx context.Context, writes []ResearchEventWrite, interrupt bool) (model.Snapshot, error) {
	r, e := s.loadWarmV37.append(ctx, writes, interrupt)
	if e == nil {
		s.eager.afterDurability(ctx)
	}
	return r, e
}

// Compare sink-visible wire data. EvidenceGroupKey is explicitly json:"-";
// its absence after persistence is not a lost forecast or an early ack.
func durableWireV42(saved model.BayesianJournalEntry, packet model.ContextPacket) error {
	if saved.Snapshot != packet.Snapshot || len(packet.Candidates) != 10 || len(saved.Report.Decisions) != 150 || len(packet.BayesianShadow.Decisions) != 150 || saved.ID == "" || saved.ID != packet.BayesianShadow.JournalID || !saved.Report.JournalDurable || !packet.BayesianShadow.JournalDurable {
		return errors.New("durable wire identity/count mismatch")
	}
	got, err := json.Marshal(saved.Report)
	if err != nil {
		return err
	}
	want, err := json.Marshal(packet.BayesianShadow)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return errors.New("durable report wire mismatch")
	}
	return nil
}

func TestResearchEagerCoreWireControlsV42(t *testing.T) {
	p := model.ContextPacket{Candidates: make([]model.Candidate, 10), BayesianShadow: model.BayesianShadowReport{JournalID: "wire-control", JournalDurable: true, Decisions: make([]model.BayesianDecision, 150)}}
	p.BayesianShadow.Decisions[0].EvidenceGroupKey = "private-not-wire"
	encoded, err := json.Marshal(p.BayesianShadow)
	if err != nil {
		t.Fatal(err)
	}
	j := model.BayesianJournalEntry{ID: p.BayesianShadow.JournalID, Snapshot: p.Snapshot}
	if err = json.Unmarshal(encoded, &j.Report); err != nil || durableWireV42(j, p) != nil {
		t.Fatal("valid JSON round trip rejected", err)
	}
	for _, mutation := range []string{"forecast", "snapshot", "omission", "durability", "journal", "report", "nonfinite"} {
		t.Run(mutation, func(t *testing.T) {
			var report model.BayesianShadowReport
			if err := json.Unmarshal(encoded, &report); err != nil {
				t.Fatal(err)
			}
			bad := j
			bad.Report = report
			switch mutation {
			case "forecast":
				bad.Report.Decisions[0].Forecast.CorrectedLaw.Useful = .75
			case "snapshot":
				bad.Snapshot.EvidenceEpoch++
			case "omission":
				bad.Report.Decisions = bad.Report.Decisions[:149]
			case "durability":
				bad.Report.JournalDurable = false
			case "journal":
				bad.ID = "foreign"
			case "report":
				bad.Report.SelectionSupportCertified = true
			case "nonfinite":
				bad.Report.Decisions[0].ActivationScore = math.NaN()
			}
			if durableWireV42(bad, p) == nil {
				t.Fatal("wire corruption accepted")
			}
		})
	}
}

func TestResearchEagerCoreOwnerBoundaryV42(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_EAGER_CORE_V42") != "1" {
		t.Skip("explicit isolated native preflight")
	}
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "lazy", true: "eager"}[enabled], func(t *testing.T) {
			ctx := context.Background()
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachEagerV42(t, f, enabled)
			defer s.closeV37()
			prime, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if e != nil {
				t.Fatal(e)
			}
			if enabled && (s.eager.prepared.Load() != 1 || s.shared.core.Load() == nil) {
				t.Fatal("durable first binding did not publish core")
			}
			if !enabled && s.shared.core.Load() != nil {
				t.Fatal("nil-certificate lazy core cached")
			}
			if _, e = f.adapter.gate.store.GetBayesianJournal(ctx, "tenant-a", prime.BayesianShadow.JournalID); e != nil {
				t.Fatal("prime not durable", e)
			}
			entered := make(chan struct{}, 1)
			resume := make(chan struct{})
			var once sync.Once
			s.metadata.afterProjection = func(context.Context) { entered <- struct{}{}; <-resume }
			s.eager.authority.owner.Lock()
			held := true
			release := func() {
				if held {
					held = false
					s.eager.authority.owner.Unlock()
				}
				once.Do(func() { close(resume) })
			}
			type response struct {
				p model.ContextPacket
				e error
			}
			done := make(chan response, 1)
			go func() { p, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "held")); done <- response{p, e} }()
			terminal := false
			defer func() {
				release()
				if !terminal {
					select {
					case <-done:
					case <-time.After(3 * time.Second):
						t.Error("failed operation did not drain before fixture close")
					}
				}
			}()
			wait := 50 * time.Millisecond
			if enabled {
				wait = 3 * time.Second
			}
			select {
			case <-entered:
				if !enabled {
					t.Fatal("lazy path bypassed owner")
				}
			case <-time.After(wait):
				if enabled {
					t.Fatal("fresh eager path blocked on owner")
				}
			}
			select {
			case r := <-done:
				terminal = true
				t.Fatal("packet before full durability", r.e)
			default:
			}
			release()
			select {
			case r := <-done:
				terminal = true
				if r.e != nil {
					t.Fatal(r.e)
				}
				saved, e := f.adapter.gate.store.GetBayesianJournal(ctx, "tenant-a", r.p.BayesianShadow.JournalID)
				if e != nil {
					t.Fatal(e)
				}
				if e = durableWireV42(saved, r.p); e != nil {
					t.Fatal("original durable wire", e)
				}
				if enabled && s.warm.fast.Load() != 1 {
					t.Fatal("fresh head not used")
				}
				if enabled && (s.eager.attempts.Load() != 2 || s.eager.preparationOwners.Load() != 2 || s.eager.prepared.Load() != 1 || s.eager.preparationNS.Load() <= 0) {
					t.Fatal("publication work not separately accounted")
				}
				t.Logf("eager=%v attempts=%d preparationOwners=%d prepared=%d errors=%d preparationNS=%d (technical race run, not latency benchmark)", enabled, s.eager.attempts.Load(), s.eager.preparationOwners.Load(), s.eager.prepared.Load(), s.eager.preparationErrors.Load(), s.eager.preparationNS.Load())
			case <-time.After(3 * time.Second):
				t.Fatal("held operation failed to drain")
			}
		})
	}
}
func TestResearchEagerCoreMutationV42(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_EAGER_CORE_V42") != "1" {
		t.Skip("explicit isolated native preflight")
	}
	for _, visible := range []bool{false, true} {
		t.Run(map[bool]string{false: "future", true: "visible"}[visible], func(t *testing.T) {
			ctx := context.Background()
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachEagerV42(t, f, true)
			defer s.closeV37()
			prime, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if e != nil {
				t.Fatal(e)
			}
			r, e := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "eager-label", false)
			if e != nil {
				t.Fatal(e)
			}
			prior := s.shared.core.Load()
			if prior == nil || prior.state.Head != r.Snapshot || len(prior.state.Sources) != 1 {
				t.Fatal("outcome core missing source")
			}
			encoded := prior.encoded
			var original witnessStateV23
			if e = json.Unmarshal([]byte(encoded), &original); e != nil {
				t.Fatal(e)
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite("eager-insert", at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			head, e := s.append(ctx, []ResearchEventWrite{w}, false)
			if e != nil {
				t.Fatal(e)
			}
			c := s.shared.core.Load()
			if c == nil || c == prior || c.state.Head != head || prior.encoded != encoded || !reflect.DeepEqual(*prior.state, original) {
				t.Fatal("eager core not owned/current")
			}
			before := s.warm.fast.Load()
			p, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "after"))
			if e != nil || s.warm.fast.Load() != before+1 || (witnessBeliefsV23(p) > 0) == visible {
				t.Fatal("eager mutation transport", e)
			}
			durable, e := f.adapter.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if e != nil || durable.EvidenceEpoch != r.Posterior.EvidenceEpoch {
				t.Fatal("source retag", e)
			}
			if s.eager.preparationErrors.Load() != 0 {
				t.Fatal("unexpected preparation errors")
			}
		})
	}
}
func TestResearchEagerCoreGapAndCancellationV42(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_EAGER_CORE_V42") != "1" {
		t.Skip("explicit isolated native preflight")
	}
	for _, interrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "unwitnessed", true: "interrupted"}[interrupt], func(t *testing.T) {
			ctx := context.Background()
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachEagerV42(t, f, true)
			defer s.closeV37()
			if _, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime")); e != nil {
				t.Fatal(e)
			}
			old := s.shared.core.Load()
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if e := s.eager.refresh(cancelled); !errors.Is(e, context.Canceled) || s.shared.core.Load() != old {
				t.Fatal("cancelled preparation published", e)
			}
			w := pinnedWrite("eager-gap", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 9, .001)
			if interrupt {
				if _, e := s.append(ctx, []ResearchEventWrite{w}, true); e == nil {
					t.Fatal("no injected interruption")
				}
			} else {
				if e := appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); e != nil {
					t.Fatal(e)
				}
			}
			if e := s.eager.refresh(ctx); e == nil || s.shared.core.Load() != old {
				t.Fatal("gapped authority published", e)
			}
			if _, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); e == nil {
				t.Fatal("uncommitted head served")
			}
		})
	}
}

var _ store.EventStore = (*eagerStoreV42)(nil)
