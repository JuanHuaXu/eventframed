package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	"github.com/JuanHuaXu/eventframed/internal/service"
)

type packetV23Candidate struct {
	ID                           string
	Probability                  float64
	Monitored, Useful            bool
	Alpha, Beta                  float64
	Base, Current, Flat, Context float64
	Law                          float64
	Belief                       bool
}

type packetV23Control struct {
	Name        string
	IDs         []string
	Utility     float64
	FalseItems  float64
	PacketBrier float64
}

type packetV23Row struct {
	Kind, Split, Regime                string
	World                              int
	Seed                               int64
	InitialEpoch, LearnedEpoch         uint64
	InitialCertified, LearnedCertified bool
	InitialPacket, ActualPacket        []string
	Candidates                         []packetV23Candidate
	Controls                           []packetV23Control
	PopulationBrier                    float64
	RecallNS, OutcomeNS                []int64
	ScoreNS                            int64
}

// Both challengers consume only an accepted one-outcome posterior and the
// prefeedback baseline. Hidden utility is deliberately absent from this API.
func packetV23Innovation(base, alpha, beta float64) (flat, contextual float64) {
	clamp := func(v float64) float64 { return math.Min(1, math.Max(0, v)) }
	flat = clamp(base + .1*(alpha/(alpha+beta)-.5))
	contextualMean := (2*base + alpha - 1) / (alpha + beta)
	contextual = clamp(base + .1*(contextualMean-base))
	return
}

func TestResearchPacketInnovationV23Controls(t *testing.T) {
	for _, base := range []float64{.1, .5, .925} {
		flat, contextual := packetV23Innovation(base, 1, 1)
		if flat != base || math.Abs(contextual-base) > 1e-15 {
			t.Fatal("a zero-evidence posterior moved its baseline")
		}
		upFlat, upContext := packetV23Innovation(base, 2, 1)
		downFlat, downContext := packetV23Innovation(base, 1, 2)
		if upFlat <= base || upContext <= base || downFlat >= base || downContext >= base {
			t.Fatal("innovation direction does not follow observed evidence")
		}
	}
}

func packetV23IDs(candidates []model.Candidate) []string {
	ids := make([]string, len(candidates))
	for i, candidate := range candidates {
		ids[i] = candidate.Event.ID
	}
	return ids
}

func runPacketV23World(t *testing.T, split, regime string, world int, seed int64) packetV23Row {
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
	adapter := &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
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
			write.Vector = denseRowV6(query, i, .005*float64(i+1))
			writes = append(writes, write)
		}
		if _, receipt, err := gate.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || receipt == 0 {
			t.Fatal("seed eligible", receipt, err)
		}
	}
	for i := 0; i < 16; i++ {
		write := pinnedWrite(fmt.Sprintf("future-%03d", i), origin.Add(time.Hour))
		write.Vector = denseRowV6(query, i, .001*float64(i+1))
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
	row := packetV23Row{Kind: "trial", Split: split, Regime: regime, World: world, Seed: seed}
	recall := func(asOf time.Time) model.ContextPacket {
		pin := &publishedRecallPinV8{}
		callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
		request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a",
			SessionID: fmt.Sprintf("packet-v23-%s-%s-%d", split, regime, world), Query: "public vector search fixture",
			Embedding: rawQuery, EmbeddingModel: em.ModelKey(), AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 10000}
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
	row.InitialPacket = packetV23IDs(before.Candidates)
	if !row.InitialCertified || len(before.BayesianShadow.Decisions) != 150 {
		t.Fatal("invalid initial frontier")
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
		default:
			t.Fatal("unknown regime")
		}
		candidate := packetV23Candidate{ID: decision.EventID, Probability: p, Monitored: i < 32}
		if candidate.Monitored {
			candidate.Useful = labelRNG.Float64() < p
		}
		row.Candidates = append(row.Candidates, candidate)
	}
	for i := 0; i < 32; i++ {
		candidate := row.Candidates[i]
		observed := origin.Add(3*time.Second + time.Duration(i)*time.Millisecond)
		request := model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion,
			IdempotencyKey: fmt.Sprintf("packet-v23-%s-%s-%d-%02d", split, regime, world, i),
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
	row.ActualPacket = packetV23IDs(learned.Candidates)
	if !row.LearnedCertified || row.LearnedEpoch != row.InitialEpoch || len(learned.BayesianShadow.Decisions) != 150 {
		t.Fatal("learned frontier or epoch changed")
	}
	ids := make([]string, len(row.Candidates))
	for i, c := range row.Candidates {
		ids[i] = c.ID
	}
	events, err := gate.store.GetEvents(ctx, "tenant-a", ids, asOf)
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
		c.Base, c.Current, c.Law = decision.Forecast.BaseLaw.Useful, decision.Forecast.RankScore, decision.Forecast.CorrectedLaw.Useful
		c.Belief = decision.Forecast.BeliefLaw != nil
		c.Flat, c.Context = c.Base, c.Base
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
			if c.Alpha != wantAlpha || c.Beta != wantBeta || math.Abs(decision.Forecast.BeliefLaw.Useful-c.Alpha/(c.Alpha+c.Beta)) > 1e-12 {
				t.Fatal("posterior is not one ordinary outcome")
			}
			started := time.Now()
			c.Flat, c.Context = packetV23Innovation(c.Base, c.Alpha, c.Beta)
			row.ScoreNS += time.Since(started).Nanoseconds()
		}
		row.PopulationBrier += mixedV22ExpectedBrier(c.Law, c.Probability) / 150
		event := eventsByID[c.ID]
		replayCandidates[i] = model.Candidate{Event: event, EstimatedTokens: max(1, (utf8.RuneCountInString(event.Content)+3)/4),
			Forecast: decision.Forecast, EvidenceGroupKey: epistemic.Describe(event).EvidenceGroupKey}
		postKeys[c.ID] = decision.PosteriorKey
	}
	for _, name := range []string{"baseline", "current", "flat", "context"} {
		candidates := append([]model.Candidate(nil), replayCandidates...)
		for i := range candidates {
			c := row.Candidates[i]
			switch name {
			case "baseline":
				candidates[i].Score = c.Base
			case "current":
				candidates[i].Score = c.Current
			case "flat":
				candidates[i].Score = c.Flat
			case "context":
				candidates[i].Score = c.Context
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
		policy := packing.DefaultPolicy()
		packed := packing.Select(candidates[:50], postKeys, 10, 50, 10000, policy)
		control := packetV23Control{Name: name, IDs: packetV23IDs(packed.Candidates)}
		if len(control.IDs) != 10 {
			t.Fatal("control packet has fewer than ten items")
		}
		if name == "current" && !reflect.DeepEqual(control.IDs, row.ActualPacket) {
			t.Fatal("current reconstruction differs from actual packet", control.IDs, row.ActualPacket)
		}
		byID := make(map[string]packetV23Candidate, len(row.Candidates))
		for _, c := range row.Candidates {
			byID[c.ID] = c
		}
		for _, candidate := range packed.Candidates {
			c := byID[candidate.Event.ID]
			control.Utility += c.Probability / 10
			control.FalseItems += 1 - c.Probability
			control.PacketBrier += mixedV22ExpectedBrier(c.Law, c.Probability) / 10
			if !reflect.DeepEqual(candidate.Forecast, replayCandidates[indexPacketV23(ids, c.ID)].Forecast) {
				t.Fatal("rank proposal changed scored law")
			}
		}
		row.Controls = append(row.Controls, control)
	}
	return row
}

func indexPacketV23(ids []string, id string) int {
	for i, candidate := range ids {
		if candidate == id {
			return i
		}
	}
	panic("packet contains an unknown ID")
}

func TestResearchPacketInnovationV23(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_PACKET_V23_OUT"), os.Getenv("EVENTFRAME_PACKET_V23_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in all-candidate packet screen")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seedBase := int64(2026102303)
	if split == "confirmation" {
		seedBase = 2026102304
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-packet-innovation-v23-protocol.md",
		"internal/store/libravdbstore/research_packet_innovation_v23_test.go",
		"internal/store/libravdbstore/research_mixed_outcome_v22_test.go",
		"internal/store/libravdbstore/research_published_outcome_load_v18_test.go",
		"internal/store/libravdbstore/research_published_outcome_gate_v17_test.go",
		"internal/store/libravdbstore/research_published_certificate_refresh_v19_test.go",
		"internal/service/service.go", "internal/packing/packing.go",
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
	if err := encoder.Encode(map[string]any{"kind": "manifest", "split": split, "seedBase": seedBase, "worldsPerRegime": 8, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for regimeIndex, regime := range []string{"independent", "aligned", "reversed"} {
		for world := 0; world < 8; world++ {
			row := runPacketV23World(t, split, regime, world, seedBase+int64(regimeIndex)*1000000+int64(world)*1000)
			if err := encoder.Encode(row); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
