package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
)

const mixedV22DesignSeed int64 = 2026102203
const mixedV22ConfirmationSeed int64 = 2026102204

type mixedV22Candidate struct {
	EventID               string
	Probability           float64
	TrainingUseful        bool
	Before, Base, Learned float64
	AfterWrite            float64
	LearnedBelief         bool
	AfterWriteBelief      bool
	PackedBefore          bool
	PackedLearned         bool
	PackedAfterWrite      bool
}

type mixedV22Row struct {
	Kind, Split         string
	World               int
	Seed                int64
	Candidates          []mixedV22Candidate
	Positive, Negative  int
	InitialEpoch        uint64
	LearnedEpoch        uint64
	AfterWriteEpoch     uint64
	InitialCertified    bool
	LearnedCertified    bool
	AfterWriteCertified bool
	BeforeBrier         float64
	BaseBrier           float64
	LearnedBrier        float64
	AfterWriteBrier     float64
	InitialRecallNS     int64
	LearnedRecallNS     int64
	AfterWriteRecallNS  int64
	OutcomeNS           []int64
	WriteNS             int64
	RefreshNS           int64
}

func mixedV22ExpectedBrier(prediction, truth float64) float64 {
	return truth*(1-prediction)*(1-prediction) + (1-truth)*prediction*prediction
}

func mixedV22Decision(packet model.ContextPacket, id string) (model.BayesianDecision, bool) {
	for _, decision := range packet.BayesianShadow.Decisions {
		if decision.EventID == id {
			return decision, true
		}
	}
	return model.BayesianDecision{}, false
}

func mixedV22Packed(packet model.ContextPacket, id string) bool {
	for _, candidate := range packet.Candidates {
		if candidate.Event.ID == id {
			return true
		}
	}
	return false
}

func runMixedV22World(t *testing.T, split string, world int, seed int64) mixedV22Row {
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
		writes := make([]ResearchEventWrite, 0, min(16, 198-start))
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
	snapshot := gate.store.Snapshot(ctx)
	selection, omitted := syntheticCertificatesV19(snapshot, time.Now().UTC())
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
	recall := func(asOf time.Time) (model.ContextPacket, int64) {
		pin := &publishedRecallPinV8{}
		callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
		request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a",
			SessionID: fmt.Sprintf("mixed-%s-%d", split, world), Query: "public vector search fixture",
			Embedding: rawQuery, EmbeddingModel: em.ModelKey(), AsOf: asOf,
			RecallK: 50, PackK: 10, TokenBudget: 10000}
		started := time.Now()
		packet, err := svc.Recall(callCtx, request)
		duration := time.Since(started).Nanoseconds()
		if err != nil || pin.view == nil || packet.Snapshot != pin.view.snapshot ||
			packet.BayesianShadow.JournalID == "" || !packet.BayesianShadow.JournalDurable {
			t.Fatal("invalid full Recall or journal", err)
		}
		return packet, duration
	}
	before, beforeNS := recall(origin.Add(2 * time.Second))
	row := mixedV22Row{Kind: "trial", Split: split, World: world, Seed: seed,
		InitialEpoch: before.Snapshot.EvidenceEpoch, InitialRecallNS: beforeNS,
		InitialCertified: before.BayesianShadow.SelectionSupportCertified && before.BayesianShadow.OmittedInfluenceCertified}
	if !row.InitialCertified || len(before.BayesianShadow.Decisions) != 150 {
		t.Fatal("initial frontier is not certified and complete")
	}
	rng := rand.New(rand.NewSource(seed))
	for _, decision := range before.BayesianShadow.Decisions {
		if !decision.Activated {
			continue
		}
		p := .8
		if len(row.Candidates)%2 == 1 {
			p = .2
		}
		useful := rng.Float64() < p
		if useful {
			row.Positive++
		} else {
			row.Negative++
		}
		row.Candidates = append(row.Candidates, mixedV22Candidate{
			EventID: decision.EventID, Probability: p, TrainingUseful: useful,
			Before:       decision.Forecast.CorrectedLaw.Useful,
			PackedBefore: mixedV22Packed(before, decision.EventID),
		})
		if len(row.Candidates) == 16 {
			break
		}
	}
	if len(row.Candidates) != 16 {
		t.Fatalf("only %d activated nominees", len(row.Candidates))
	}
	for i, candidate := range row.Candidates {
		observed := origin.Add(3*time.Second + time.Duration(i)*time.Millisecond)
		request := model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion,
			IdempotencyKey: fmt.Sprintf("mixed-v22-%s-%d-%02d", split, world, i),
			TenantID:       "tenant-a", JournalID: before.BayesianShadow.JournalID,
			EventID: candidate.EventID, Useful: candidate.TrainingUseful,
			Source: model.OutcomeFullStream, InclusionProbability: 1,
			ObservedAt: observed, AvailableAt: observed}
		started := time.Now()
		response, err := svc.ObserveBayesianOutcome(ctx, request)
		row.OutcomeNS = append(row.OutcomeNS, time.Since(started).Nanoseconds())
		if err != nil || response.Duplicate {
			t.Fatal("journaled mixed outcome", i, err)
		}
	}
	asOf := origin.Add(5 * time.Second)
	learned, learnedNS := recall(asOf)
	row.LearnedRecallNS = learnedNS
	row.LearnedEpoch = learned.Snapshot.EvidenceEpoch
	row.LearnedCertified = learned.BayesianShadow.SelectionSupportCertified && learned.BayesianShadow.OmittedInfluenceCertified
	for i := range row.Candidates {
		candidate := &row.Candidates[i]
		decision, ok := mixedV22Decision(learned, candidate.EventID)
		if !ok {
			t.Fatal("learned frontier lost selected event", candidate.EventID)
		}
		candidate.Base = decision.Forecast.BaseLaw.Useful
		candidate.Learned = decision.Forecast.CorrectedLaw.Useful
		candidate.LearnedBelief = decision.Forecast.BeliefLaw != nil
		candidate.PackedLearned = mixedV22Packed(learned, candidate.EventID)
	}
	writes := make([]ResearchEventWrite, 16)
	for i := range writes {
		writes[i] = pinnedWrite(fmt.Sprintf("visible-v22-%03d", i), origin.Add(time.Second))
		writes[i].Vector = denseRowV6(query, i, .00213*float64(i+1))
	}
	started := time.Now()
	if _, err := base.appendBatch(ctx, writes); err != nil {
		t.Fatal("visible append", err)
	}
	row.WriteNS = time.Since(started).Nanoseconds()
	selection, omitted = syntheticCertificatesV19(gate.store.Snapshot(ctx), time.Now().UTC())
	started = time.Now()
	if _, err := adapter.publishSyntheticCertificatesV19(ctx, selection, omitted); err != nil {
		t.Fatal("synthetic certificate refresh", err)
	}
	row.RefreshNS = time.Since(started).Nanoseconds()
	after, afterNS := recall(asOf)
	row.AfterWriteRecallNS = afterNS
	row.AfterWriteEpoch = after.Snapshot.EvidenceEpoch
	row.AfterWriteCertified = after.BayesianShadow.SelectionSupportCertified && after.BayesianShadow.OmittedInfluenceCertified
	for i := range row.Candidates {
		candidate := &row.Candidates[i]
		decision, ok := mixedV22Decision(after, candidate.EventID)
		if !ok {
			t.Fatal("postwrite frontier lost selected event", candidate.EventID)
		}
		candidate.AfterWrite = decision.Forecast.CorrectedLaw.Useful
		candidate.AfterWriteBelief = decision.Forecast.BeliefLaw != nil
		candidate.PackedAfterWrite = mixedV22Packed(after, candidate.EventID)
		row.BeforeBrier += mixedV22ExpectedBrier(candidate.Before, candidate.Probability) / 16
		row.BaseBrier += mixedV22ExpectedBrier(candidate.Base, candidate.Probability) / 16
		row.LearnedBrier += mixedV22ExpectedBrier(candidate.Learned, candidate.Probability) / 16
		row.AfterWriteBrier += mixedV22ExpectedBrier(candidate.AfterWrite, candidate.Probability) / 16
	}
	return row
}

func TestResearchMixedOutcomeV22(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_MIXED_V22_OUT"), os.Getenv("EVENTFRAME_MIXED_V22_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in mixed-outcome scored-law screen")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seedBase := mixedV22DesignSeed
	if split == "confirmation" {
		seedBase = mixedV22ConfirmationSeed
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-mixed-outcome-v22-protocol.md",
		"internal/store/libravdbstore/research_mixed_outcome_v22_test.go",
		"internal/store/libravdbstore/research_published_outcome_load_v18_test.go",
		"internal/store/libravdbstore/research_published_certificate_refresh_v19_test.go",
		"internal/service/service.go",
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
	if err := encoder.Encode(map[string]any{"kind": "manifest", "split": split,
		"seedBase": seedBase, "worlds": 4, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for world := 0; world < 4; world++ {
		row := runMixedV22World(t, split, world, seedBase+int64(world)*1000)
		if err := encoder.Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
