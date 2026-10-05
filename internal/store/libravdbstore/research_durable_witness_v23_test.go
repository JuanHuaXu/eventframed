package libravdbstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchvalidity"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// These adapters exist only in a test binary. Transport is request-local and
// never modifies durable posterior/certificate epochs. The single writer owns
// witness publication before releasing the same admission lease as Recall.
type witnessScopeV23 struct {
	Binding, Vector, Frontier string
	AsOf                      time.Time
	Snapshot                  model.Snapshot
}
type witnessContextV23 struct{}
type witnessBindingV23 struct {
	Binding, Vector, Frontier string
	Snapshot                  model.Snapshot
	SourceAt                  time.Time
}
type witnessSourceV23 struct {
	Record    researchvalidity.Record
	Binding   witnessBindingV23
	Posterior model.BayesianPosterior
}
type witnessTransitionV23 struct {
	Before, After model.Snapshot
	Mutations     []researchvalidity.Mutation
	StateHash     string
}
type witnessStateV23 struct {
	Head        model.Snapshot
	Hash        string
	Log         []researchvalidity.Mutation
	Sources     map[string]witnessSourceV23
	Journals    map[string]witnessBindingV23
	Certificate *witnessBindingV23
	Base        model.Snapshot
}
type witnessStoreV23 struct {
	*publishedOutcomeStoreV17
	enabled    bool
	state      *witnessStateV23 // owner mutex; admission also excludes runtime writes
	poison     atomic.Bool
	transports atomic.Int64
}

func witnessHashV23(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func witnessIDsV23(ids []string) string {
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	return witnessHashV23(ids)
}
func witnessStateHashV23(state witnessStateV23) string {
	state.Hash = ""
	return witnessHashV23(state)
}
func witnessGenesisV23(snapshot model.Snapshot) *witnessStateV23 {
	s := &witnessStateV23{Head: snapshot, Base: snapshot, Sources: map[string]witnessSourceV23{}, Journals: map[string]witnessBindingV23{}}
	s.Hash = witnessStateHashV23(*s)
	return s
}
func witnessRecordV23(key string, b witnessBindingV23, at time.Time) researchvalidity.Record {
	return researchvalidity.Record{TenantID: "tenant-a", PosteriorKey: key, QueryDigest: b.Binding, ModelID: b.Vector,
		HorizonKey: model.RetrievalUsefulnessHorizon, SourceID: key, SourceAt: at, FrontierDigest: b.Frontier, FrontierK: 150, CutoffLower: -1, Base: b.Snapshot}
}
func newWitnessStoreV23(ctx context.Context, base *publishedRecallLoadStoreV16, enabled bool) (*witnessStoreV23, error) {
	s := &witnessStoreV23{publishedOutcomeStoreV17: &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}, enabled: enabled}
	if _, err := base.gate.sidecar.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS witness_v23(seq INTEGER PRIMARY KEY,payload TEXT NOT NULL,prior TEXT NOT NULL,digest TEXT NOT NULL);
	CREATE TABLE IF NOT EXISTS witness_state_v23(id INTEGER PRIMARY KEY CHECK(id=1),payload TEXT NOT NULL)`); err != nil {
		return nil, err
	}
	var encoded string
	err := base.gate.sidecar.QueryRowContext(ctx, "SELECT payload FROM witness_state_v23 WHERE id=1").Scan(&encoded)
	if err == nil {
		if err = json.Unmarshal([]byte(encoded), &s.state); err != nil {
			return nil, err
		}
		if s.state == nil || s.state.Sources == nil || s.state.Journals == nil || len(s.state.Log) > 10000 || len(s.state.Sources) > 200 || len(s.state.Journals) > 512 {
			return nil, errors.New("invalid/bounded witness state")
		}
		rows, err := base.gate.sidecar.QueryContext(ctx, "SELECT payload,prior,digest FROM witness_v23 ORDER BY seq")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		head := s.state.Base
		root := witnessGenesisV23(head).Hash
		stateHash := root
		var log []researchvalidity.Mutation
		for rows.Next() {
			var payload, prior, digest string
			if err = rows.Scan(&payload, &prior, &digest); err != nil {
				return nil, err
			}
			var tr witnessTransitionV23
			if err = json.Unmarshal([]byte(payload), &tr); err != nil {
				return nil, err
			}
			if tr.Before != head || prior != root || digest != witnessHashV23([]string{prior, payload}) {
				s.poison.Store(true)
				return s, nil
			}
			head = tr.After
			root = digest
			stateHash = tr.StateHash
			log = append(log, tr.Mutations...)
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		if head != s.state.Head || root != s.state.Hash || stateHash != witnessStateHashV23(*s.state) || !reflect.DeepEqual(log, s.state.Log) || head != base.gate.store.Snapshot(ctx) {
			s.poison.Store(true)
		}
		return s, nil
	}
	// A non-empty/malformed state is never repaired by inventing genesis.
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var count int
	if err := base.gate.sidecar.QueryRowContext(ctx, "SELECT COUNT(*) FROM witness_v23").Scan(&count); err != nil {
		return nil, err
	}
	if count != 0 {
		return nil, errors.New("missing witness state with existing commit chain")
	}
	snap := base.gate.store.Snapshot(ctx)
	s.state = witnessGenesisV23(snap)
	encodedBytes, _ := json.Marshal(s.state)
	_, err = base.gate.sidecar.ExecContext(ctx, "INSERT INTO witness_state_v23(id,payload) VALUES(1,?)", string(encodedBytes))
	return s, err
}
func (s *witnessStoreV23) persistV23(ctx context.Context, after model.Snapshot, mutations []researchvalidity.Mutation, source *witnessSourceV23, journalID string, binding *witnessBindingV23) error {
	if s.poison.Load() {
		return store.ErrStaleSnapshot
	}
	// Copy through JSON so a failed transaction cannot mutate the published head.
	data, _ := json.Marshal(s.state)
	var next witnessStateV23
	if err := json.Unmarshal(data, &next); err != nil {
		return err
	}
	before := next.Head
	if after != next.Head {
		if len(mutations) == 0 || len(next.Log)+len(mutations) > 10000 {
			return errors.New("witness gap/cap")
		}
		next.Head = after
		next.Log = append(next.Log, mutations...)
	}
	if source != nil {
		next.Sources[source.Record.PosteriorKey] = *source
	}
	if binding != nil {
		next.Journals[journalID] = *binding
		if next.Certificate == nil {
			b := *binding
			b.Snapshot = next.Base
			next.Certificate = &b
		}
	}
	if len(next.Sources) > 200 || len(next.Journals) > 512 {
		return errors.New("source/journal witness cap")
	}
	// Bind provenance updates too, even when a journal does not move runtime.
	tr := witnessTransitionV23{Before: before, After: after, Mutations: mutations, StateHash: witnessStateHashV23(next)}
	b, _ := json.Marshal(tr)
	payload := string(b)
	digest := witnessHashV23([]string{s.state.Hash, payload})
	next.Hash = digest
	tx, err := s.gate.sidecar.BeginTx(ctx, nil)
	if err != nil {
		s.poison.Store(true)
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO witness_v23(payload,prior,digest) VALUES(?,?,?)", payload, s.state.Hash, digest); err != nil {
		s.poison.Store(true)
		return err
	}
	b, _ = json.Marshal(next)
	if _, err = tx.ExecContext(ctx, "UPDATE witness_state_v23 SET payload=? WHERE id=1", string(b)); err != nil {
		s.poison.Store(true)
		return err
	}
	if err = tx.Commit(); err != nil {
		s.poison.Store(true)
		return err
	}
	s.state = &next
	return nil
}
func (s *witnessStoreV23) appendV23(ctx context.Context, writes []ResearchEventWrite, interrupt bool) (model.Snapshot, error) {
	s.admission.Lock()
	defer s.admission.Unlock()
	s.owner.Lock()
	defer s.owner.Unlock()
	before := s.gate.store.Snapshot(ctx)
	if s.poison.Load() || before != s.state.Head {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	if err := appendDenseV6(ctx, s.gate, writes); err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return model.Snapshot{}, err
	}
	after := s.gate.store.Snapshot(ctx)
	if interrupt {
		s.poison.Store(true)
		s.current.Store(nil)
		return after, errors.New("injected backend-before-witness interruption")
	}
	if after.RuntimeVersion != before.RuntimeVersion+uint64(len(writes)) || after.EvidenceEpoch != before.EvidenceEpoch+uint64(len(writes)) {
		s.poison.Store(true)
		return after, errors.New("unaccounted insert transition")
	}
	mutations := make([]researchvalidity.Mutation, len(writes))
	for i, w := range writes {
		mutations[i] = researchvalidity.Mutation{Kind: researchvalidity.Insert, RuntimeVersion: before.RuntimeVersion + uint64(i+1), EvidenceEpochAfter: before.EvidenceEpoch + uint64(i+1), AvailableAt: w.Event.AvailableAt}
	}
	if err := s.persistV23(ctx, after, mutations, nil, "", nil); err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return after, err
	}
	if err := s.publishLocked(ctx); err != nil {
		s.poison.Store(true)
		return after, err
	}
	return after, nil
}
func (s *witnessStoreV23) Search(ctx context.Context, tenant string, vector []float32, asOf time.Time, k int) ([]store.SearchResult, error) {
	s.owner.Lock()
	valid := !s.poison.Load() && s.state.Head == s.gate.store.Snapshot(ctx)
	s.owner.Unlock()
	if !valid {
		return nil, store.ErrStaleSnapshot
	}
	results, err := s.publishedRecallLoadStoreV16.Search(ctx, tenant, vector, asOf, k)
	if err != nil {
		return nil, err
	}
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if scope == nil {
		return nil, errors.New("missing actual request binding")
	}
	scope.Vector = witnessHashV23(vector)
	scope.AsOf = asOf
	scope.Snapshot = s.Snapshot(ctx)
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.Event.ID
	}
	scope.Frontier = witnessIDsV23(ids)
	return results, nil
}
func (s *witnessStoreV23) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if err := s.publishedRecallLoadStoreV16.PutBayesianJournal(ctx, entry); err != nil {
		return err
	}
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if scope == nil {
		return errors.New("missing journal scope")
	}
	b := witnessBindingV23{Binding: scope.Binding, Vector: scope.Vector, Frontier: scope.Frontier, Snapshot: entry.Snapshot, SourceAt: scope.AsOf}
	// Journal-bound provenance updates need their own mutex: multiple read leases
	// can reach journal ack concurrently. The owner serializes with all publishes.
	s.owner.Lock()
	defer s.owner.Unlock()
	return s.persistV23(ctx, s.state.Head, nil, nil, entry.ID, &b)
}
func (s *witnessStoreV23) compatibleV23(ctx context.Context, record researchvalidity.Record, b witnessBindingV23) bool {
	scope, _ := ctx.Value(witnessContextV23{}).(*witnessScopeV23)
	if !s.enabled || s.poison.Load() || scope == nil || scope.Binding != b.Binding || scope.Vector != b.Vector || scope.Frontier != b.Frontier {
		return false
	}
	if scope.Snapshot != s.state.Head || s.gate.store.Snapshot(ctx) != s.state.Head {
		return false
	}
	start := sort.Search(len(s.state.Log), func(i int) bool { return s.state.Log[i].RuntimeVersion > record.Base.RuntimeVersion })
	req := researchvalidity.Request{TenantID: record.TenantID, PosteriorKey: record.PosteriorKey, QueryDigest: scope.Binding, ModelID: scope.Vector, HorizonKey: record.HorizonKey, SourceID: record.SourceID, AsOf: scope.AsOf, FrontierDigest: scope.Frontier, Current: scope.Snapshot}
	return researchvalidity.Check(record, req, s.state.Log[start:]).Status == researchvalidity.Compatible
}
func (s *witnessStoreV23) GetSelectionCertificate(ctx context.Context, tenant string) (model.SelectionSupportCertificate, error) {
	c, err := s.EventStore.GetSelectionCertificate(ctx, tenant)
	if err != nil {
		return c, err
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	if s.state.Certificate != nil {
		b := *s.state.Certificate
		r := witnessRecordV23("certificate", b, c.ValidFrom)
		if c.EvidenceEpoch == b.Snapshot.EvidenceEpoch && s.compatibleV23(ctx, r, b) {
			c.EvidenceEpoch = s.Snapshot(ctx).EvidenceEpoch
		}
	}
	return c, nil
}
func (s *witnessStoreV23) GetOmittedInfluenceCertificate(ctx context.Context, tenant string) (model.OmittedInfluenceCertificate, error) {
	c, err := s.EventStore.GetOmittedInfluenceCertificate(ctx, tenant)
	if err != nil {
		return c, err
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	if s.state.Certificate != nil {
		b := *s.state.Certificate
		r := witnessRecordV23("certificate", b, b.SourceAt)
		if c.EvidenceEpoch == b.Snapshot.EvidenceEpoch && s.compatibleV23(ctx, r, b) {
			c.EvidenceEpoch = s.Snapshot(ctx).EvidenceEpoch
		}
	}
	return c, nil
}
func (s *witnessStoreV23) GetBayesianPosterior(ctx context.Context, tenant, key string) (model.BayesianPosterior, error) {
	p, err := s.EventStore.GetBayesianPosterior(ctx, tenant, key)
	if err != nil {
		return p, err
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	if source, ok := s.state.Sources[key]; ok && reflect.DeepEqual(source.Posterior, p) && s.compatibleV23(ctx, source.Record, source.Binding) {
		if p.EvidenceEpoch != s.Snapshot(ctx).EvidenceEpoch {
			s.transports.Add(1)
		}
		p.EvidenceEpoch = s.Snapshot(ctx).EvidenceEpoch
	}
	return p, nil
}
func (s *witnessStoreV23) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, key, parent, digest string, weight float64, change bayes.ChangePolicy, group bayes.GroupPolicy, observation model.ResidualObservation, policy residual.Policy) (store.BayesianOutcomeResult, error) {
	if parent != "" {
		return store.BayesianOutcomeResult{}, errors.New("hierarchical affected-key closure is not implemented in this experiment")
	}
	s.admission.Lock()
	defer s.admission.Unlock()
	s.owner.Lock()
	defer s.owner.Unlock()
	before := s.gate.store.Snapshot(ctx)
	if s.poison.Load() || before != s.state.Head {
		return store.BayesianOutcomeResult{}, store.ErrStaleSnapshot
	}
	b, ok := s.state.Journals[request.JournalID]
	if !ok {
		return store.BayesianOutcomeResult{}, errors.New("outcome lacks committed query provenance")
	}
	result, err := s.gate.applyOutcomeV17(ctx, request, key, parent, digest, weight, change, group, observation, policy, "")
	if err != nil {
		s.poison.Store(true)
		s.current.Store(nil)
		return result, err
	}
	if result.Duplicate {
		return result, nil
	}
	after := result.Snapshot
	if after.RuntimeVersion != before.RuntimeVersion+1 || after.EvidenceEpoch != before.EvidenceEpoch {
		s.poison.Store(true)
		return result, errors.New("unaccounted outcome transition")
	}
	b.Snapshot = after
	r := witnessRecordV23(key, b, result.Posterior.UpdatedAt)
	r.SourceID = request.IdempotencyKey
	source := witnessSourceV23{r, b, result.Posterior}
	mutations := []researchvalidity.Mutation{{Kind: researchvalidity.Outcome, RuntimeVersion: after.RuntimeVersion, EvidenceEpochAfter: after.EvidenceEpoch, AvailableAt: request.AvailableAt, AffectedKnown: true, AffectedKeys: []string{key}}}
	if err = s.persistV23(ctx, after, mutations, &source, "", nil); err != nil {
		s.current.Store(nil)
		return result, err
	}
	err = s.publishLocked(ctx)
	return result, err
}

// This entry point constructs the binding from the actual request. The native
// read pin and shared lease remain held until journal durability acknowledges.
func (s *witnessStoreV23) recallV23(ctx context.Context, svc *service.Service, req model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	if req.TenantID != "tenant-a" {
		return model.ContextPacket{}, nil, errors.New("unexpected tenant")
	}
	if req.Query == "" || req.EmbeddingModel == "" {
		return model.ContextPacket{}, nil, errors.New("missing request")
	}
	// Selection parameters are part of the key, not hidden frozen defaults.
	selection := req
	selection.Embedding, selection.AsOf = nil, time.Time{}
	scope := &witnessScopeV23{Binding: witnessHashV23(selection)}
	pin := &publishedRecallPinV8{}
	ctx = context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
	ctx = context.WithValue(ctx, witnessContextV23{}, scope)
	s.admission.RLock()
	defer s.admission.RUnlock()
	packet, err := svc.Recall(ctx, req)
	return packet, pin.view, err
}

type witnessFixtureV23 struct {
	adapter *witnessStoreV23
	svc     *service.Service
	em      embed.Embedder
	query   []float32
	origin  time.Time
	root    string
}

func createWitnessFixtureV23(t *testing.T, enabled bool) *witnessFixtureV23 {
	t.Helper()
	ctx := context.Background()
	origin := time.Now().UTC().Add(-10 * time.Second)
	q := denseQueryV6()
	root := t.TempDir()
	g := createLiveDenseGateV18(t, root, q, origin)
	base := &publishedRecallLoadStoreV16{EventStore: g.store, gate: g, journalJobs: make(chan *batchJournalJobV16, 128), workerEnd: make(chan struct{})}
	a := &witnessStoreV23{publishedOutcomeStoreV17: &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}, enabled: enabled}
	em, _ := embed.NewHashEmbedder(256)
	svc, err := service.New(a, em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	for start := 0; start < 198; start += 16 {
		var writes []ResearchEventWrite
		for i := start; i < min(198, start+16); i++ {
			w := pinnedWrite(fmt.Sprintf("eligible-%03d", i), origin)
			w.Vector, err = normalizedVectorV4(denseRowV6(q, i, .005*float64(i+1)), 256)
			if err != nil {
				t.Fatal(err)
			}
			writes = append(writes, w)
		}
		if _, _, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 16; i++ {
		w := pinnedWrite(fmt.Sprintf("future-%03d", i), origin.Add(time.Hour))
		w.Vector, err = normalizedVectorV4(denseRowV6(q, i, .001*float64(i+1)), 256)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := g.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{w}); err != nil {
			t.Fatal(err)
		}
	}
	g.wantRows = 217
	selection, omitted := syntheticCertificatesV19(g.store.Snapshot(ctx), time.Now().UTC())
	if _, err := g.store.PublishSelectionCertificate(ctx, selection); err != nil {
		t.Fatal(err)
	}
	if _, err := g.store.PublishOmittedInfluenceCertificate(ctx, omitted); err != nil {
		t.Fatal(err)
	}
	if err := g.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := g.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := base.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	initialized, err := newWitnessStoreV23(ctx, base, enabled)
	if err != nil {
		t.Fatal(err)
	}
	a.state = initialized.state
	go base.runJournalWorker()
	return &witnessFixtureV23{a, svc, em, q, origin, root}
}
func (f *witnessFixtureV23) close() { f.svc.Close(); f.adapter.gate.close() }
func (f *witnessFixtureV23) reopen(t *testing.T) {
	t.Helper()
	enabled := f.adapter.enabled
	f.close()
	g, err := openDenseOutcomeGateV17(f.root)
	if err != nil {
		t.Fatal(err)
	}
	base := &publishedRecallLoadStoreV16{EventStore: g.store, gate: g, journalJobs: make(chan *batchJournalJobV16, 128), workerEnd: make(chan struct{})}
	if err := base.publishLocked(context.Background()); err != nil {
		g.close()
		t.Fatal(err)
	}
	a, err := newWitnessStoreV23(context.Background(), base, enabled)
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	svc, err := service.New(a, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	f.adapter, f.svc = a, svc
	go base.runJournalWorker()
}
func (f *witnessFixtureV23) request(at time.Time, _ string) model.RecallRequest {
	raw := make([]float32, len(f.query))
	for i, v := range f.query {
		raw[i] = 2 * v
	}
	return model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "witness-session", Query: "public vector search fixture", Embedding: raw, EmbeddingModel: f.em.ModelKey(), AsOf: at, RecallK: 50, PackK: 10, TokenBudget: 10000}
}
func (f *witnessFixtureV23) feedback(ctx context.Context, journal, event, id string, useful bool) (model.BayesianOutcomeResponse, error) {
	now := time.Now().UTC()
	return f.svc.ObserveBayesianOutcome(ctx, f.outcomeRequest(journal, event, id, useful, now))
}
func (f *witnessFixtureV23) outcomeRequest(journal, event, id string, useful bool, at time.Time) model.BayesianOutcomeRequest {
	return model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, TenantID: "tenant-a", JournalID: journal, EventID: event, Useful: useful, ObservedAt: at, AvailableAt: at, Source: model.OutcomeFullStream, InclusionProbability: 1}
}
func witnessBeliefsV23(p model.ContextPacket) int {
	n := 0
	for _, d := range p.BayesianShadow.Decisions {
		if d.Forecast.BeliefLaw != nil {
			n++
		}
	}
	return n
}
func TestResearchDurableWitnessLifecycleV23(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_DURABLE_WITNESS_V23") != "1" {
		t.Skip("isolated libravdb witness integration")
	}
	ctx := context.Background()
	for _, visible := range []bool{false, true} {
		t.Run(fmt.Sprintf("visible-%v", visible), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			a := f.adapter
			p, _, err := a.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			response, err := f.feedback(ctx, p.BayesianShadow.JournalID, "past100", "first", false)
			if err != nil {
				t.Fatal(err)
			}
			stored, err := a.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if err != nil {
				t.Fatal(err)
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite("new-witness", at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err = a.appendV23(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			after, _, err := a.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "after"))
			if err != nil {
				t.Fatal(err)
			}
			if (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong served transport", visible, witnessBeliefsV23(after))
			}
			original, err := a.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if err != nil || !reflect.DeepEqual(stored, original) || original.EvidenceEpoch != response.Posterior.EvidenceEpoch {
				t.Fatal("durable posterior retagged")
			}
			changed := f.request(time.Now().UTC(), "changed")
			changed.Query = "a different query"
			other, _, err := a.recallV23(ctx, f.svc, changed)
			if err != nil || witnessBeliefsV23(other) != 0 {
				t.Fatal("cross-query transport", err)
			}
			changed = f.request(time.Now().UTC(), "vector")
			changed.Embedding = denseRowV6(f.query, 11, .1)
			other, _, err = a.recallV23(ctx, f.svc, changed)
			if err != nil || witnessBeliefsV23(other) != 0 {
				t.Fatal("cross-vector transport", err)
			}
			changed = f.request(time.Now().UTC(), "selection")
			changed.PackK = 9
			other, _, err = a.recallV23(ctx, f.svc, changed)
			if err != nil || witnessBeliefsV23(other) != 0 {
				t.Fatal("cross-selection transport", err)
			}
			old, _, err := a.recallV23(ctx, f.svc, f.request(response.Posterior.UpdatedAt.Add(-time.Nanosecond), "old"))
			if err != nil || witnessBeliefsV23(old) != 0 {
				t.Fatal("future evidence", err)
			}
			f.reopen(t)
			if f.adapter.poison.Load() {
				t.Fatal("committed witness reload poisoned")
			}
			reopened, _, err := f.adapter.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "reopened"))
			if err != nil || (witnessBeliefsV23(reopened) > 0) == visible {
				t.Fatal("wrong served transport after actual reopen", err)
			}
		})
	}
	for _, interruption := range []bool{false, true} {
		t.Run(fmt.Sprintf("gap-%v", interruption), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			w := pinnedWrite("unwitnessed", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 5, .0004)
			if interruption {
				_, err := f.adapter.appendV23(ctx, []ResearchEventWrite{w}, true)
				if err == nil {
					t.Fatal("interruption missing")
				}
			} else {
				if err := appendDenseV6(ctx, f.adapter.gate, []ResearchEventWrite{w}); err != nil {
					t.Fatal(err)
				}
			}
			f.reopen(t)
			if !f.adapter.poison.Load() {
				t.Fatal("unaccounted backend state accepted")
			}
			if _, _, err := f.adapter.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "gap")); err == nil {
				t.Fatal("poisoned witness served a Recall")
			}
		})
	}
	for _, corrupt := range []string{"binding", "chain"} {
		t.Run("corrupt-"+corrupt, func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			if _, _, err := f.adapter.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "prime")); err != nil {
				t.Fatal(err)
			}
			db := f.adapter.gate.sidecar
			if corrupt == "binding" {
				payload, _ := json.Marshal(f.adapter.state)
				var modified witnessStateV23
				if err := json.Unmarshal(payload, &modified); err != nil {
					t.Fatal(err)
				}
				modified.Certificate.Binding = "corrupted"
				payload, _ = json.Marshal(modified)
				if _, err := db.ExecContext(ctx, "UPDATE witness_state_v23 SET payload=? WHERE id=1", string(payload)); err != nil {
					t.Fatal(err)
				}
			} else if _, err := db.ExecContext(ctx, "UPDATE witness_v23 SET prior='corrupted' WHERE seq=1"); err != nil {
				t.Fatal(err)
			}
			f.reopen(t)
			if !f.adapter.poison.Load() {
				t.Fatal("corrupted durable binding accepted")
			}
		})
	}
}

// Collector types retain actual candidate laws and original source timestamps,
// so the auditor can recompute observed freshness without assuming publication
// is the same event as serving or silently dropping never-used labels.
type witnessReadV23 struct {
	Index              int
	Offer, Start, End  time.Time
	Snapshot           model.Snapshot
	Journal            string
	Decisions          []model.BayesianDecision
	Packed             []model.Candidate
	ViewAt             time.Time
	Error              string
	Request            model.RecallRequest
	PinSnapshot        model.Snapshot
	Selection, Omitted bool
}
type witnessWriteV23 struct {
	Index      int
	Offer, Ack time.Time
	Snapshot   model.Snapshot
	Available  time.Time
	Error      string
}
type witnessOutcomeV23 struct {
	Index            int
	EventID          string
	Useful           bool
	Offer, Published time.Time
	Response         model.BayesianOutcomeResponse
	Request          model.BayesianOutcomeRequest
	Error            string
}
type witnessTrialV23 struct {
	Trial            int
	Enabled, Visible bool
	Reads            []witnessReadV23
	Writes           []witnessWriteV23
	Outcomes         []witnessOutcomeV23
	Errors           []string
	Transported      int64
	WitnessHash      string
	FinalSnapshot    model.Snapshot
	Metrics          map[string]int64
	Pass             bool
	Origin           time.Time
	InitialSnapshot  model.Snapshot
	FinalWitness     witnessStateV23
	DurableSources   []model.BayesianPosterior
	Commits          []witnessCommitV23
}
type witnessCommitV23 struct {
	Seq                    int
	Payload, Prior, Digest string
}

func runWitnessLoadV23(t *testing.T, trial int, enabled, visible bool) witnessTrialV23 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	f := createWitnessFixtureV23(t, enabled)
	defer f.close()
	a := f.adapter
	prime, _, err := a.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	for _, d := range prime.BayesianShadow.Decisions {
		if d.Activated {
			events = append(events, d.EventID)
			if len(events) == 16 {
				break
			}
		}
	}
	if len(events) != 16 {
		t.Fatal("prime activation")
	}
	start := time.Now().Add(20 * time.Millisecond)
	jobs := make(chan recallLoadOfferV16, 128)
	writeJobs := make(chan sortLoadWriteJob, 128)
	outcomeJobs := make(chan outcomeOfferV18, 16)
	reads := make(chan witnessReadV23, 128)
	writes := make(chan []witnessWriteV23, 1)
	outcomes := make(chan []witnessOutcomeV23, 1)
	go func() {
		defer close(jobs)
		for i := 0; i < 128; i++ {
			timer := time.NewTimer(time.Until(start.Add(time.Duration(i) * 4 * time.Millisecond)))
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
			jobs <- recallLoadOfferV16{index: i, offered: time.Now().UTC()}
		}
	}()
	go func() {
		defer close(writeJobs)
		for i := 0; i < 128; i++ {
			timer := time.NewTimer(time.Until(start.Add(time.Duration(i) * 4 * time.Millisecond)))
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite(fmt.Sprintf("write-%03d", i), at)
			w.Vector = denseRowV6(f.query, i, .00213*float64(i+1))
			writeJobs <- sortLoadWriteJob{write: w, offered: time.Now().UTC()}
		}
	}()
	go func() {
		var rows []witnessWriteV23
		for {
			group, closed := sortLoadTakeGroup(writeJobs, 16, 16*time.Millisecond)
			if len(group) == 0 {
				break
			}
			batch := make([]ResearchEventWrite, len(group))
			for j, job := range group {
				batch[j] = job.write
			}
			snap, err := a.appendV23(ctx, batch, false)
			ack := time.Now().UTC()
			for _, job := range group {
				r := witnessWriteV23{Index: len(rows), Offer: job.offered, Ack: ack, Snapshot: snap, Available: job.write.Event.AvailableAt}
				if err != nil {
					r.Error = err.Error()
				}
				rows = append(rows, r)
			}
			if closed || err != nil {
				break
			}
		}
		writes <- rows
	}()
	go func() {
		defer close(outcomeJobs)
		for i, event := range events {
			timer := time.NewTimer(time.Until(start.Add(20*time.Millisecond + time.Duration(i)*16*time.Millisecond)))
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
			offer := time.Now().UTC()
			useful := i%2 == 0
			outcomeJobs <- outcomeOfferV18{index: i, request: f.outcomeRequest(prime.BayesianShadow.JournalID, event, fmt.Sprintf("load-%d", i), useful, offer)}
		}
	}()
	go func() {
		var rows []witnessOutcomeV23
		for job := range outcomeJobs {
			response, err := f.svc.ObserveBayesianOutcome(ctx, job.request)
			r := witnessOutcomeV23{Index: job.index, EventID: job.request.EventID, Useful: job.request.Useful, Offer: job.request.AvailableAt, Published: time.Now().UTC(), Response: response, Request: job.request}
			if err != nil {
				r.Error = err.Error()
			}
			rows = append(rows, r)
			if err != nil {
				break
			}
		}
		outcomes <- rows
	}()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				began := time.Now().UTC()
				request := f.request(job.offered, fmt.Sprintf("load-%d", job.index))
				packet, view, err := a.recallV23(ctx, f.svc, request)
				r := witnessReadV23{Index: job.index, Offer: job.offered, Start: began, End: time.Now().UTC(), Snapshot: packet.Snapshot, Journal: packet.BayesianShadow.JournalID, Decisions: packet.BayesianShadow.Decisions, Packed: packet.Candidates}
				r.Request, r.Selection, r.Omitted = request, packet.BayesianShadow.SelectionSupportCertified, packet.BayesianShadow.OmittedInfluenceCertified
				if view != nil {
					r.ViewAt = view.at
					r.PinSnapshot = view.snapshot
				}
				if err != nil {
					r.Error = err.Error()
				}
				reads <- r
			}
		}()
	}
	go func() { wg.Wait(); close(reads) }()
	row := witnessTrialV23{Trial: trial, Enabled: enabled, Visible: visible, Metrics: map[string]int64{}, Origin: f.origin, InitialSnapshot: prime.Snapshot}
	for r := range reads {
		row.Reads = append(row.Reads, r)
	}
	row.Writes = <-writes
	row.Outcomes = <-outcomes
	row.Transported = a.transports.Load()
	row.WitnessHash = a.state.Hash
	row.FinalSnapshot = a.gate.store.Snapshot(ctx)
	row.FinalWitness = *a.state
	commitRows, err := a.gate.sidecar.QueryContext(ctx, "SELECT seq,payload,prior,digest FROM witness_v23 ORDER BY seq")
	if err != nil {
		t.Fatal(err)
	}
	for commitRows.Next() {
		var c witnessCommitV23
		if err := commitRows.Scan(&c.Seq, &c.Payload, &c.Prior, &c.Digest); err != nil {
			commitRows.Close()
			t.Fatal(err)
		}
		row.Commits = append(row.Commits, c)
	}
	if err := commitRows.Err(); err != nil {
		t.Fatal(err)
	}
	commitRows.Close()
	for _, event := range events {
		posterior, err := a.gate.store.GetBayesianPosterior(ctx, "tenant-a", event)
		if err != nil {
			row.Errors = append(row.Errors, "missing durable outcome source")
		}
		row.DurableSources = append(row.DurableSources, posterior)
	}
	if len(row.Reads) != 128 || len(row.Writes) != 128 || len(row.Outcomes) != 16 {
		row.Errors = append(row.Errors, "incomplete offered work")
	}
	var readNS, offerNS, viewNS, writeNS, outcomeNS []int64
	learned := 0
	for _, r := range row.Reads {
		readNS = append(readNS, r.End.Sub(r.Start).Nanoseconds())
		offerNS = append(offerNS, r.End.Sub(r.Offer).Nanoseconds())
		viewNS = append(viewNS, r.End.Sub(r.ViewAt).Nanoseconds())
		if r.Error != "" || r.Journal == "" || len(r.Decisions) != 150 {
			row.Errors = append(row.Errors, "invalid Recall/journal")
		}
		saved, err := a.gate.store.GetBayesianJournal(ctx, "tenant-a", r.Journal)
		if err != nil || saved.Snapshot != r.Snapshot || !reflect.DeepEqual(saved.Report.Decisions, r.Decisions) {
			row.Errors = append(row.Errors, "durable journal mismatch")
		}
		for _, d := range r.Decisions {
			if d.Forecast.BeliefLaw != nil {
				learned++
			}
			if len(d.EventID) >= 5 && d.EventID[:5] == "write" && !visible {
				row.Errors = append(row.Errors, "future nominee")
			}
		}
	}
	for _, w := range row.Writes {
		writeNS = append(writeNS, w.Ack.Sub(w.Offer).Nanoseconds())
		if w.Error != "" {
			row.Errors = append(row.Errors, "write: "+w.Error)
		}
	}
	for _, o := range row.Outcomes {
		outcomeNS = append(outcomeNS, o.Published.Sub(o.Offer).Nanoseconds())
		if o.Error != "" || o.Response.Duplicate {
			row.Errors = append(row.Errors, "outcome: "+o.Error)
		}
	}
	row.Metrics["learned_decisions"] = int64(learned)
	for name, values := range map[string][]int64{"call": readNS, "offer": offerNS, "view": viewNS, "write": writeNS, "outcome": outcomeNS} {
		row.Metrics[name+"_p99_ns"] = pinnedPercentile(values, .99).Nanoseconds()
		row.Metrics[name+"_max_ns"] = pinnedPercentile(values, 1).Nanoseconds()
	}
	if a.poison.Load() || a.state.Head != row.FinalSnapshot {
		row.Errors = append(row.Errors, "witness head invalid")
	}
	if _, ready := a.gate.capture(ctx); !ready {
		row.Errors = append(row.Errors, "marker not READY")
	}
	row.Pass = len(row.Errors) == 0 && row.Metrics["call_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["offer_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["write_p99_ns"] < int64(250*time.Millisecond) && row.Metrics["outcome_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["outcome_max_ns"] < int64(250*time.Millisecond) && row.Metrics["view_max_ns"] < int64(250*time.Millisecond)
	if enabled && !visible {
		row.Pass = row.Pass && learned > 0 && row.Transported > 0
	}
	if visible {
		row.Pass = row.Pass && row.Transported == 0
	}
	return row
}
func TestResearchDurableWitnessLoadV23(t *testing.T) {
	output := os.Getenv("EVENTFRAME_DURABLE_WITNESS_V23_OUTPUT")
	if output == "" {
		t.Skip("exclusive prospective output required")
	}
	manifest, err := os.ReadFile("../../../research/durable-witness-v23/freeze.json")
	if err != nil {
		t.Fatal("prospective freeze required", err)
	}
	manifestHash := sha256.Sum256(manifest)
	file, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	if err := enc.Encode(map[string]any{"Type": "header", "Protocol": "docs/experiments/mmm-durable-witness-v23-protocol.md", "Trials": 8, "Time": time.Now().UTC(), "FreezeSHA256": hex.EncodeToString(manifestHash[:])}); err != nil {
		t.Fatal(err)
	}
	for trial := 1; trial <= 2; trial++ {
		for _, visible := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				row := runWitnessLoadV23(t, trial, enabled, visible)
				if err := enc.Encode(row); err != nil {
					t.Fatal(err)
				}
				if err := file.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("trial=%d enabled=%v visible=%v pass=%v errors=%v metrics=%v transported=%d", trial, enabled, visible, row.Pass, row.Errors, row.Metrics, row.Transported)
			}
		}
	}
	if err := enc.Encode(map[string]any{"Type": "footer", "Trials": 8}); err != nil {
		t.Fatal(err)
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
