package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/epistemic"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/researchbeta"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type packetV26Candidate struct {
	ID                             string
	Probability                    float64
	Monitored, Useful              bool
	Alpha, Beta                    float64
	Base, Served, Flat, Context    float64
	Law                            float64
	PriorBase, FlatLaw, ContextLaw float64
	Belief                         bool
}

type packetV26Control struct {
	Name        string
	IDs         []string
	Utility     float64
	FalseItems  float64
	PacketBrier float64
}

type packetV26Row struct {
	Kind, Split, Geometry, Regime                                    string
	AngleStep                                                        float64
	World                                                            int
	Seed                                                             int64
	InitialEpoch, LearnedEpoch                                       uint64
	InitialCertified, LearnedCertified                               bool
	InitialPacket, ActualPacket                                      []string
	Candidates                                                       []packetV26Candidate
	Controls                                                         []packetV26Control
	PopulationBrier                                                  float64
	PopulationBaseBrier, PopulationFlatBrier, PopulationContextBrier float64
	RecallNS, OutcomeNS                                              []int64
	ScoreNS                                                          int64
}

// Both challengers consume only an accepted one-outcome posterior and the
// prefeedback baseline. Hidden utility is deliberately absent from this API.
func packetV26Innovation(base, priorBase, flatMean, contextualMean float64) (flat, contextual float64) {
	clamp := func(v float64) float64 { return math.Min(1, math.Max(0, v)) }
	flat = clamp(base + .1*(flatMean-.5))
	contextual = clamp(base + .1*(contextualMean-priorBase))
	return
}

func TestResearchPacketInnovationV26Controls(t *testing.T) {
	for _, base := range []float64{.1, .5, .925} {
		flat, contextual := packetV26Innovation(base, base, .5, base)
		if flat != base || math.Abs(contextual-base) > 1e-15 {
			t.Fatal("a zero-evidence posterior moved its baseline")
		}
		upFlat, upContext := packetV26Innovation(base, base, 2.0/3, (2*base+1)/3)
		downFlat, downContext := packetV26Innovation(base, base, 1.0/3, 2*base/3)
		if upFlat <= base || upContext <= base || downFlat >= base || downContext >= base {
			t.Fatal("innovation direction does not follow observed evidence")
		}
	}
}

func packetV26IDs(candidates []model.Candidate) []string {
	ids := make([]string, len(candidates))
	for i, candidate := range candidates {
		ids[i] = candidate.Event.ID
	}
	return ids
}

func runPacketV26World(t *testing.T, split, geometry, regime string, step float64, world int, seed int64) packetV26Row {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	origin := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createLiveDenseGateV18(t, t.TempDir(), query, origin)
	defer gate.close()
	base := &publishedRecallLoadStoreV16{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV16, 128), workerEnd: make(chan struct{})}
	adapter := &contextPriorStoreV26{EventStore: &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}}
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
		RankingPolicy: ranking.Policy{ContextualEnabled: true, PosteriorWeight: 1},
		ResidualMode:  service.ResidualModeDisabled, MaxRankDelta: 1,
		ElasticRankDelta: ranking.ElasticDeltaPolicy{MinScale: 1, MaxScale: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	go base.runJournalWorker()
	defer svc.Close()
	for start := 0; start < 198; start += 16 {
		writes := make([]ResearchEventWrite, 0, 16)
		for i := start; i < min(198, start+16); i++ {
			write := pinnedWrite(fmt.Sprintf("eligible-%03d", i), origin)
			write.Vector, err = normalizedVectorV4(denseRowV6(query, i, step*float64(i+1)), 256)
			if err != nil {
				t.Fatal("eligible normalization", err)
			}
			writes = append(writes, write)
		}
		if _, receipt, err := gate.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || receipt == 0 {
			t.Fatal("seed eligible", receipt, err)
		}
	}
	for i := 0; i < 16; i++ {
		write := pinnedWrite(fmt.Sprintf("future-%03d", i), origin.Add(time.Hour))
		write.Vector, err = normalizedVectorV4(denseRowV6(query, i, .001*float64(i+1)), 256)
		if err != nil {
			t.Fatal("future normalization", err)
		}
		if _, receipt, err := gate.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write}); err != nil || receipt == 0 {
			t.Fatal("seed future", receipt, err)
		}
	}
	gate.wantRows = 217
	selection, omitted := syntheticCertificatesV19(gate.store.Snapshot(ctx), time.Now().UTC())
	if _, err := gate.store.PublishSelectionCertificate(ctx, selection); err != nil {
		t.Fatal(err)
	}
	if _, err := gate.store.PublishOmittedInfluenceCertificate(ctx, omitted); err != nil {
		t.Fatal(err)
	}
	if err := gate.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := base.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	row := packetV26Row{Kind: "trial", Split: split, Geometry: geometry, Regime: regime, AngleStep: step, World: world, Seed: seed}
	recall := func(asOf time.Time) model.ContextPacket {
		pin := &publishedRecallPinV8{}
		callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
		request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a",
			SessionID: fmt.Sprintf("packet-v26-%s-%s-%d", split, regime, world), Query: "public vector search fixture",
			Embedding: rawQuery, EmbeddingModel: em.ModelKey(), AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 10000}
		callCtx, scopeErr := priorContextV26(callCtx, request)
		if scopeErr != nil {
			t.Fatal(scopeErr)
		}
		started := time.Now()
		packet, err := svc.Recall(callCtx, request)
		row.RecallNS = append(row.RecallNS, time.Since(started).Nanoseconds())
		if err != nil || pin.view == nil || packet.Snapshot != pin.view.snapshot || !packet.BayesianShadow.JournalDurable || packet.BayesianShadow.JournalID == "" {
			t.Fatal("invalid Recall/journal", err)
		}
		return packet
	}
	before := recall(origin.Add(2 * time.Second))
	row.InitialEpoch = before.Snapshot.EvidenceEpoch
	row.InitialCertified = before.BayesianShadow.SelectionSupportCertified && before.BayesianShadow.OmittedInfluenceCertified
	row.InitialPacket = packetV26IDs(before.Candidates)
	if !row.InitialCertified || len(before.BayesianShadow.Decisions) != 150 {
		t.Fatal("invalid initial frontier")
	}
	if err := adapter.bind(ctx, "tenant-a", before.BayesianShadow.JournalID); err != nil {
		t.Fatal("bind prefeedback journal", err)
	}
	truthRNG := rand.New(rand.NewSource(seed + 5000000000))
	labelRNG := rand.New(rand.NewSource(seed))
	permutation := truthRNG.Perm(150)
	keys := make(map[string]bool)
	for i, decision := range before.BayesianShadow.Decisions {
		if !decision.Activated || keys[decision.PosteriorKey] || strings.HasPrefix(decision.EventID, "future") {
			t.Fatal("invalid initial nominee identity")
		}
		keys[decision.PosteriorKey] = true
		p := .2
		switch regime {
		case "independent":
			if permutation[i] < 75 {
				p = .8
			}
		case "aligned":
			p = .9 - .8*float64(i)/149
		case "reversed":
			p = .1 + .8*float64(i)/149
		case "calibrated":
			p = decision.Forecast.BaseLaw.Useful
		default:
			t.Fatal("unknown regime")
		}
		candidate := packetV26Candidate{ID: decision.EventID, Probability: p, Monitored: i < 32, PriorBase: decision.Forecast.BaseLaw.Useful}
		if candidate.Monitored {
			candidate.Useful = labelRNG.Float64() < p
		}
		row.Candidates = append(row.Candidates, candidate)
	}
	for i := 0; i < 32; i++ {
		candidate := row.Candidates[i]
		observed := origin.Add(3*time.Second + time.Duration(i)*time.Millisecond)
		request := model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion,
			IdempotencyKey: fmt.Sprintf("packet-v26-%s-%s-%d-%02d", split, regime, world, i),
			TenantID:       "tenant-a", JournalID: before.BayesianShadow.JournalID, EventID: candidate.ID,
			Useful: candidate.Useful, Source: model.OutcomeFullStream, InclusionProbability: 1,
			ObservedAt: observed, AvailableAt: observed}
		started := time.Now()
		response, err := svc.ObserveBayesianOutcome(ctx, request)
		row.OutcomeNS = append(row.OutcomeNS, time.Since(started).Nanoseconds())
		if err != nil || response.Duplicate {
			t.Fatal("journaled outcome", i, err)
		}
	}
	asOf := origin.Add(5 * time.Second)
	learned := recall(asOf)
	row.LearnedEpoch = learned.Snapshot.EvidenceEpoch
	row.LearnedCertified = learned.BayesianShadow.SelectionSupportCertified && learned.BayesianShadow.OmittedInfluenceCertified
	row.ActualPacket = packetV26IDs(learned.Candidates)
	if !row.LearnedCertified || row.LearnedEpoch != row.InitialEpoch || len(learned.BayesianShadow.Decisions) != 150 {
		t.Fatal("learned frontier or epoch changed")
	}
	ids := make([]string, len(row.Candidates))
	for i, c := range row.Candidates {
		ids[i] = c.ID
	}
	events, err := gate.store.GetEventsWithVectors(ctx, "tenant-a", ids, asOf)
	if err != nil || len(events) != 150 {
		t.Fatal("incomplete event reconstruction", err)
	}
	eventsByID := make(map[string]model.Event, len(events))
	for _, event := range events {
		eventsByID[event.ID] = event
	}
	postKeys := make(map[string]string, len(events))
	replayCandidates := make([]model.Candidate, len(events))
	for i := range row.Candidates {
		c := &row.Candidates[i]
		decision, ok := mixedV22Decision(learned, c.ID)
		if !ok {
			t.Fatal("nominee disappeared", c.ID)
		}
		c.Base, c.Served, c.Law = decision.Forecast.BaseLaw.Useful, decision.Forecast.RankScore, decision.Forecast.CorrectedLaw.Useful
		c.Belief = decision.Forecast.BeliefLaw != nil
		if decision.Forecast.ResidualApplied {
			t.Fatal("residual reached contextual law")
		}
		c.Flat, c.Context = c.Base, c.Base
		c.FlatLaw, c.ContextLaw = c.Base, c.Base
		if c.Belief != c.Monitored {
			t.Fatal("accepted belief does not match monitored evidence", c.ID)
		}
		if c.Monitored {
			posterior, err := gate.store.GetBayesianPosterior(ctx, "tenant-a", decision.PosteriorKey)
			if err != nil || posterior.EvidenceEpoch != row.LearnedEpoch || !posterior.Certified || posterior.UpdatedAt.After(asOf) {
				t.Fatal("invalid accepted posterior", err)
			}
			c.Alpha, c.Beta = posterior.Alpha, posterior.Beta
			wantAlpha, wantBeta := 1.0, 2.0
			if c.Useful {
				wantAlpha, wantBeta = 2, 1
			}
			if c.Alpha != wantAlpha || c.Beta != wantBeta || math.Abs(decision.Forecast.BeliefLaw.Useful-(2*c.PriorBase+c.Alpha-1)/(c.Alpha+c.Beta)) > 1e-12 {
				t.Fatal("posterior is not one ordinary outcome")
			}
			started := time.Now()
			c.FlatLaw, err = researchbeta.Predict(.5, 2, uint64(c.Alpha-1), uint64(c.Beta-1))
			if err != nil {
				t.Fatal("flat joint model", err)
			}
			c.ContextLaw, err = researchbeta.Predict(c.PriorBase, 2, uint64(c.Alpha-1), uint64(c.Beta-1))
			if err != nil {
				t.Fatal("context joint model", err)
			}
			c.Flat, c.Context = packetV26Innovation(c.Base, c.PriorBase, c.FlatLaw, c.ContextLaw)
			row.ScoreNS += time.Since(started).Nanoseconds()
		}
		if math.Abs(c.Law-c.ContextLaw) > 1e-12 || math.Abs(c.Served-c.Law) > 1e-12 || math.Abs(decision.Forecast.PreResidualLaw.Useful-c.Law) > 1e-12 {
			t.Fatal("context prior did not reach served scored bundle", c.ID)
		}
		row.PopulationBrier += mixedV22ExpectedBrier(c.Law, c.Probability) / 150
		row.PopulationBaseBrier += mixedV22ExpectedBrier(c.Base, c.Probability) / 150
		row.PopulationFlatBrier += mixedV22ExpectedBrier(c.FlatLaw, c.Probability) / 150
		row.PopulationContextBrier += mixedV22ExpectedBrier(c.ContextLaw, c.Probability) / 150
		event := eventsByID[c.ID]
		angle := 0.0
		if strings.HasPrefix(c.ID, "eligible-") {
			var number int
			if _, err := fmt.Sscanf(c.ID, "eligible-%d", &number); err != nil {
				t.Fatal("bad eligible ID", err)
			}
			angle = step * float64(number+1)
		} else if c.ID != "past100" && c.ID != "at120" {
			t.Fatal("unexpected oracle ID", c.ID)
		}
		var squared, dot float64
		for j, value := range event.Embedding {
			squared += float64(value) * float64(value)
			dot += float64(value) * float64(query[j])
		}
		if len(event.Embedding) != 256 || math.Abs(math.Sqrt(squared)-1) > 1e-5 || math.Abs(dot/math.Sqrt(squared)-math.Cos(angle)) > 1e-5 {
			t.Fatal("durable unit/cosine oracle mismatch", c.ID)
		}
		expectedBase := .65*(math.Cos(angle)+1)/2 + .15 + .1*math.Exp(-asOf.Sub(event.AvailableAt).Hours()/(24*30)) + .025
		if math.Abs(c.Base-expectedBase) > 1e-6 {
			t.Fatal("service baseline disagrees with angle oracle", c.ID, c.Base, expectedBase)
		}
		replayCandidates[i] = model.Candidate{Event: event, EstimatedTokens: max(1, (utf8.RuneCountInString(event.Content)+3)/4),
			Forecast: decision.Forecast, EvidenceGroupKey: epistemic.Describe(event).EvidenceGroupKey}
		postKeys[c.ID] = decision.PosteriorKey
	}
	for _, name := range []string{"baseline", "served", "flat", "context"} {
		candidates := append([]model.Candidate(nil), replayCandidates...)
		for i := range candidates {
			c := row.Candidates[i]
			switch name {
			case "baseline":
				candidates[i].Score = c.Base
			case "served":
				candidates[i].Score = c.Served
			case "flat":
				candidates[i].Score = c.Flat
			case "context":
				candidates[i].Score = c.Context
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
		policy := packing.DefaultPolicy()
		packed := packing.Select(candidates[:50], postKeys, 10, 50, 10000, policy)
		control := packetV26Control{Name: name, IDs: packetV26IDs(packed.Candidates)}
		if len(control.IDs) != 10 {
			t.Fatal("control packet has fewer than ten items")
		}
		if name == "served" && !reflect.DeepEqual(control.IDs, row.ActualPacket) {
			t.Fatal("current reconstruction differs from actual packet", control.IDs, row.ActualPacket)
		}
		byID := make(map[string]packetV26Candidate, len(row.Candidates))
		for _, c := range row.Candidates {
			byID[c.ID] = c
		}
		for _, candidate := range packed.Candidates {
			c := byID[candidate.Event.ID]
			control.Utility += c.Probability / 10
			control.FalseItems += 1 - c.Probability
			control.PacketBrier += mixedV22ExpectedBrier(c.Law, c.Probability) / 10
			if !reflect.DeepEqual(candidate.Forecast, replayCandidates[indexPacketV26(ids, c.ID)].Forecast) {
				t.Fatal("rank proposal changed scored law")
			}
		}
		row.Controls = append(row.Controls, control)
	}
	saved, err := gate.store.GetBayesianJournal(ctx, "tenant-a", learned.BayesianShadow.JournalID)
	if err != nil {
		t.Fatal("read served durable journal", err)
	}
	// EvidenceGroupKey is intentionally json:"-"; compare the entire wire record.
	durableBytes, durableErr := json.Marshal(saved.Report.Decisions)
	servedBytes, servedErr := json.Marshal(learned.BayesianShadow.Decisions)
	if durableErr != nil || servedErr != nil || string(durableBytes) != string(servedBytes) {
		t.Fatal("served forecast differs from durable journal", err)
	}
	if split == "race-correctness" {
		checkPriorServiceScopeV26(t, ctx, adapter, em.ModelKey(), rawQuery, origin, before, learned)
	}
	return row
}

func indexPacketV26(ids []string, id string) int {
	for i, candidate := range ids {
		if candidate == id {
			return i
		}
	}
	panic("packet contains an unknown ID")
}

func TestResearchPacketInnovationV26(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_PACKET_V26_OUT"), os.Getenv("EVENTFRAME_PACKET_V26_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in all-candidate packet screen")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seedBase := int64(2026102603)
	if split == "confirmation" {
		seedBase = 2026102604
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-packet-prior-v26-protocol.md",
		"internal/store/libravdbstore/research_packet_prior_v26_test.go",
		"internal/store/libravdbstore/research_mixed_outcome_v22_test.go",
		"internal/store/libravdbstore/research_published_outcome_load_v18_test.go",
		"internal/store/libravdbstore/research_published_dense_v6_test.go",
		"internal/store/libravdbstore/research_published_recall_worker_v16_test.go",
		"internal/store/libravdbstore/research_event_batch.go",
		"internal/store/libravdbstore/store.go",
		"internal/ranking/elastic.go",
		"internal/store/libravdbstore/research_published_outcome_gate_v17_test.go",
		"internal/store/libravdbstore/research_published_certificate_refresh_v19_test.go",
		"internal/service/service.go", "internal/packing/packing.go",
		"internal/researchbeta/prior.go", "internal/ranking/ranking.go",
		"internal/store/libravdbstore/research_context_prior_v26_test.go",
		"internal/store/libravdbstore/research_context_prior_v26_controls_test.go",
		"internal/frame/query.go", "internal/frame/text.go",
	} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(content)
		hashes[name] = hex.EncodeToString(hash[:])
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(map[string]any{"kind": "manifest", "split": split, "seedBase": seedBase, "worldsPerGeometryRegime": 32, "geometries": []string{"tight", "wide"}, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for geometryIndex, geometry := range []struct {
		name string
		step float64
	}{{"tight", .005}, {"wide", .020}} {
		for regimeIndex, regime := range []string{"independent", "aligned", "reversed", "calibrated"} {
			for world := 0; world < 32; world++ {
				seed := seedBase + int64(geometryIndex)*10000000 + int64(regimeIndex)*1000000 + int64(world)*1000
				row := runPacketV26World(t, split, geometry.name, regime, geometry.step, world, seed)
				if err := encoder.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestResearchPacketPriorV26RaceWorlds(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PACKET_V26_RACE") != "1" {
		t.Skip("opt-in full-service correctness worlds")
	}
	for geometryIndex, geometry := range []struct {
		name string
		step float64
	}{{"tight", .005}, {"wide", .020}} {
		for regimeIndex, regime := range []string{"independent", "aligned", "calibrated"} {
			seed := int64(2026102609) + int64(geometryIndex)*10000000 + int64(regimeIndex)*1000000
			t.Run(geometry.name+"/"+regime, func(t *testing.T) {
				runPacketV26World(t, "race-correctness", geometry.name, regime, geometry.step, 0, seed)
			})
		}
	}
}

func checkPriorServiceScopeV26(t *testing.T, ctx context.Context, adapter *contextPriorStoreV26, modelKey string, vector []float32, origin time.Time, before, learned model.ContextPacket) {
	t.Helper()
	request := model.RecallRequest{TenantID: "tenant-a", Query: "public vector search fixture", Embedding: vector, EmbeddingModel: modelKey, AsOf: origin.Add(5 * time.Second)}
	scoped, err := priorContextV26(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	key := before.BayesianShadow.Decisions[0].PosteriorKey
	want, err := adapter.GetBayesianPosterior(scoped, "tenant-a", key)
	if err != nil {
		t.Fatal("accepted contextual posterior", err)
	}
	reconstructed := &contextPriorStoreV26{EventStore: adapter.EventStore}
	if err := reconstructed.bind(ctx, "tenant-a", before.BayesianShadow.JournalID); err != nil {
		t.Fatal("durable anchor reconstruction", err)
	}
	got, err := reconstructed.GetBayesianPosterior(scoped, "tenant-a", key)
	if err != nil || got.Alpha != want.Alpha || got.Beta != want.Beta {
		t.Fatal("reconstructed origin changed prediction", err)
	}
	if adapter.bind(ctx, "tenant-a", before.BayesianShadow.JournalID) == nil {
		t.Fatal("immutable origin replaced")
	}
	if _, err := adapter.GetBayesianPosterior(ctx, "tenant-a", key); !errors.Is(err, store.ErrPosteriorNotFound) {
		t.Fatal("unscoped posterior escaped", err)
	}
	request.Query = "different public query"
	wrong, err := priorContextV26(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.GetBayesianPosterior(wrong, "tenant-a", key); !errors.Is(err, store.ErrPosteriorNotFound) {
		t.Fatal("query-scoped evidence escaped", err)
	}
	request.Query = "public vector search fixture"
	request.AsOf = origin.Add(2 * time.Second)
	early, err := priorContextV26(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.GetBayesianPosterior(early, "tenant-a", key); !errors.Is(err, store.ErrPosteriorNotFound) {
		t.Fatal("future feedback escaped", err)
	}
	if _, err := adapter.GetBayesianPosterior(scoped, "other-tenant", key); !errors.Is(err, store.ErrPosteriorNotFound) {
		t.Fatal("tenant-scoped evidence escaped", err)
	}
	if _, err := adapter.GetBayesianPosterior(scoped, "tenant-a", "unknown-key"); !errors.Is(err, store.ErrPosteriorNotFound) {
		t.Fatal("unanchored posterior escaped", err)
	}
}
