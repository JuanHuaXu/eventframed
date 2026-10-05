//go:build researchpriority

// public-nara-fusion records an untouched public retrieval transfer screen.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/researchpacket"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const root = "research/public-task-pilot/"
const modelDigest = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"

type fact struct {
	ID   string `json:"fixture_id"`
	Text string `json:"text"`
}
type query struct {
	ID       string `json:"case_id"`
	Question string `json:"question"`
}
type candidate struct {
	Fixture string
	Score   float64
	Law     model.BernoulliLaw
}
type result struct {
	Arm, Case      string
	Candidates     []candidate
	Frontier       int
	RecallNS       int64
	Before, After  []retrieval.Candidate
	Laws           map[string]model.BernoulliLaw
	OriginalScores map[string]float64
	JournalMatched bool
	Explanation    string
	PacketStatus   string
}
type traceRanker struct{ before, after []retrieval.Candidate }

func (t *traceRanker) ContractName() string { return (retrieval.PassthroughRanker{}).ContractName() }
func (t *traceRanker) RankCandidates(ctx context.Context, req retrieval.RankRequest) ([]retrieval.Candidate, error) {
	t.before = append([]retrieval.Candidate(nil), req.Candidates...)
	out, err := (retrieval.PassthroughRanker{}).RankCandidates(ctx, req)
	t.after = append([]retrieval.Candidate(nil), out...)
	return out, err
}

func readJSON(path string, value any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, value)
}

func checkModelDigest(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:11434/api/tags", nil)
	if err != nil {
		return err
	}
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("embedding inventory status %d", response.StatusCode)
	}
	var inventory struct {
		Models []struct {
			Name   string
			Digest string
		}
	}
	if err := json.NewDecoder(response.Body).Decode(&inventory); err != nil {
		return err
	}
	for _, m := range inventory.Models {
		if m.Name == "nomic-embed-text:latest" {
			if m.Digest != modelDigest {
				return errors.New("embedding model digest changed")
			}
			return nil
		}
	}
	return errors.New("declared embedding model unavailable")
}

func runCase(ctx context.Context, em embed.Embedder, facts []fact, q query, arm string, now time.Time) (result, error) {
	config := service.Config{MaxRankDelta: .25, DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000}
	config.ResearchTemporalPriority = arm == "priority"
	config.PackingPolicy = packing.DefaultPolicy()
	config.PackingPolicy.DiversityEnabled = true
	trace := &traceRanker{}
	config.CandidateRanker = trace
	config.CandidateRankerRequired = true
	memory := memorystore.New()
	s, err := service.New(memory, em, config)
	if err != nil {
		return result{}, err
	}
	defer s.Close()
	ids := make(map[string]string, len(facts))
	for i, f := range facts {
		id := fmt.Sprintf("public-record-%03d", i)
		captured, err := s.CaptureTurn(ctx, model.CaptureTurnRequest{
			ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id,
			Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "seed", Sequence: uint64(i + 1), UserText: f.Text, AssistantText: "Recorded.", OccurredAt: now, ObservedAt: now, AvailableAt: now},
		})
		if err != nil {
			return result{}, err
		}
		ids[captured.EventID] = f.ID
	}
	start := time.Now()
	packet, err := s.Recall(researchcalendar.Bind(ctx, "research", q.Question), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query",
		Query: q.Question, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 10, TokenBudget: 10000,
	})
	if err != nil {
		return result{}, err
	}
	duration := time.Since(start).Nanoseconds()
	journal, err := memory.GetBayesianJournal(ctx, "research", packet.BayesianShadow.JournalID)
	if err != nil {
		return result{}, err
	}
	out := result{Arm: arm, Case: q.ID, Frontier: len(trace.before), RecallNS: duration,
		Before: trace.before, After: trace.after, Laws: make(map[string]model.BernoulliLaw), OriginalScores: make(map[string]float64),
		Explanation: packet.TemporalPriority, PacketStatus: packet.PacketCalibrationStatus,
		JournalMatched: journal.TemporalPriority == packet.TemporalPriority && journal.PacketCalibrationStatus == packet.PacketCalibrationStatus}
	for _, c := range trace.before {
		fixture, ok := ids[c.ID]
		if !ok {
			return result{}, errors.New("ranker candidate has no fixture identity")
		}
		out.OriginalScores[fixture] = c.Score
	}
	for _, c := range packet.Candidates {
		fixture, ok := ids[c.Event.ID]
		if !ok {
			return result{}, errors.New("packed event has no fixture identity")
		}
		out.Candidates = append(out.Candidates, candidate{Fixture: fixture, Score: c.Score, Law: c.Forecast.CorrectedLaw})
	}
	for _, decision := range packet.BayesianShadow.Decisions {
		fixture, ok := ids[decision.EventID]
		if !ok {
			return result{}, errors.New("forecast decision has no fixture identity")
		}
		out.Laws[fixture] = decision.Forecast.CorrectedLaw
	}
	return out, nil
}

func run(output string) error {
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		return errors.New("output already exists or cannot be checked")
	}
	ctx := context.Background()
	if err := checkModelDigest(ctx); err != nil {
		return err
	}
	var facts []fact
	var queries []query
	if err := readJSON(root+"nara-fusion-v1/corpus.json", &facts); err != nil {
		return err
	}
	if err := readJSON(root+"nara-fusion-v1/queries.json", &queries); err != nil {
		return err
	}
	if len(facts) != 28 || len(queries) != 32 {
		return errors.New("frozen National Archives transfer shape changed")
	}
	em, err := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
		URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768,
		Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: ",
	})
	if err != nil {
		return err
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	results := make([]result, 0, 64)
	byArm := map[string]map[string]result{"focus": {}, "priority": {}}
	for _, arm := range []string{"focus", "priority"} {
		for _, q := range queries {
			caseCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			r, err := runCase(caseCtx, em, facts, q, arm, now)
			cancel()
			if err != nil {
				return fmt.Errorf("%s/%s: %w", arm, q.ID, err)
			}
			results = append(results, r)
			byArm[arm][q.ID] = r
			fmt.Println("completed", arm, q.ID)
		}
	}
	fusions := make(map[string][]string, len(queries))
	for _, q := range queries {
		base := byArm["focus"][q.ID]
		priority := byArm["priority"][q.ID]
		baseIDs := make([]string, 0, len(base.Candidates))
		priorityIDs := make([]string, 0, len(priority.Candidates))
		for _, c := range base.Candidates {
			baseIDs = append(baseIDs, c.Fixture)
		}
		for _, c := range priority.Candidates {
			priorityIDs = append(priorityIDs, c.Fixture)
		}
		fused, err := researchpacket.PromoteWithinBaseline(baseIDs, priorityIDs)
		if err != nil {
			return fmt.Errorf("fusion %s: %w", q.ID, err)
		}
		fusions[q.ID] = fused
	}
	paths := []string{
		root + "NARA_FUSION_PROTOCOL.md", root + "nara-fusion-v1/facts.json", root + "nara-fusion-v1/corpus.json", root + "nara-fusion-v1/queries.json",
		root + "prepare_nara_fusion.py", "cmd/public-nara-fusion/main.go", "internal/researchpacket/fusion.go",
		root + "task-lexical-overlay-v1/overlay.json", root + "task-lexical-overlay-v1/sources.json",
		root + "task-lexical-overlay-v1/source-0.go.txt", root + "task-lexical-overlay-v1/source-1.go.txt", root + "task-lexical-overlay-v1/source-2.go.txt",
		"internal/service/calendar_task_lexical_overlay.go", "internal/researchcalendar/packing_lexical.go", "internal/researchcalendar/task_plan.go",
		"internal/researchcalendar/lexical.go", "internal/service/research_rank.go", "internal/embed/openai.go",
	}
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(raw)
		hashes[path] = hex.EncodeToString(hash[:])
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(struct {
		Model   string
		Digest  string
		Hashes  map[string]string
		Results []result
		Fusions map[string][]string
	}{em.ModelKey(), modelDigest, hashes, results, fusions}); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func main() {
	if len(os.Args) != 2 {
		panic("usage: public-nara-fusion NEW.json")
	}
	if err := run(os.Args[1]); err != nil {
		panic(err)
	}
}
