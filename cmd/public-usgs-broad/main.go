//go:build researchpriority

// public-usgs-broad runs an untouched public, broad-location retrieval check.
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
	"reflect"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const root = "research/public-task-pilot/"
const set = root + "usgs-broad-confirmation-v1/"
const modelDigest = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"

type fact struct {
	ID   string `json:"fixture_id"`
	Text string `json:"text"`
}
type query struct {
	ID       string `json:"case_id"`
	Question string `json:"question"`
	Split    string `json:"split"`
}
type candidate struct {
	Fixture string
	Score   float64
	Law     model.BernoulliLaw
}
type result struct {
	Arm, Case      string
	Candidates     []candidate
	Frontier       []string
	RecallNS       int64
	Before, After  []retrieval.Candidate
	Laws           map[string]model.BernoulliLaw
	OriginalScores map[string]float64
	JournalMatched bool
}
type traceRanker struct {
	mode          string
	before, after []retrieval.Candidate
}

func (t *traceRanker) ContractName() string {
	if t.mode == "task-lexical" {
		return "research/task-lexical-rank-v1"
	}
	return (retrieval.PassthroughRanker{}).ContractName()
}
func (t *traceRanker) RankCandidates(ctx context.Context, req retrieval.RankRequest) ([]retrieval.Candidate, error) {
	t.before = append([]retrieval.Candidate(nil), req.Candidates...)
	var out []retrieval.Candidate
	var err error
	if t.mode == "task-lexical" {
		var plan researchcalendar.TaskPlan
		plan, err = researchcalendar.PlanTaskRole(ctx, req)
		if err == nil {
			var lexical researchcalendar.LexicalResult
			lexical, err = researchcalendar.LexicalOrder(ctx, req.QueryText, req.Candidates, &plan.PriorityPlan)
			if err == nil {
				out = make([]retrieval.Candidate, len(req.Candidates))
				for rank, index := range lexical.Order {
					out[rank] = req.Candidates[index]
					out[rank].Score = 1 - float64(rank)/float64(len(out)+1)
				}
			}
		}
	} else {
		out, err = (retrieval.PassthroughRanker{}).RankCandidates(ctx, req)
	}
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
	config.PackingPolicy = packing.DefaultPolicy()
	config.PackingPolicy.DiversityEnabled = true
	trace := &traceRanker{mode: arm}
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
			Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "seed", Sequence: uint64(i + 1),
				UserText: f.Text, AssistantText: "Recorded.", OccurredAt: now, ObservedAt: now, AvailableAt: now},
		})
		if err != nil {
			return result{}, err
		}
		ids[captured.EventID] = f.ID
	}
	start := time.Now()
	packet, err := s.Recall(researchcalendar.Bind(ctx, "research", q.Question), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query",
		Query: q.Question, AsOf: now.Add(time.Second), RecallK: 50, PackK: 10, TokenBudget: 10000,
	})
	if err != nil {
		return result{}, err
	}
	duration := time.Since(start).Nanoseconds()
	journal, err := memory.GetBayesianJournal(ctx, "research", packet.BayesianShadow.JournalID)
	if err != nil {
		return result{}, err
	}
	out := result{Arm: arm, Case: q.ID, RecallNS: duration, Before: trace.before, After: trace.after,
		Laws: make(map[string]model.BernoulliLaw), OriginalScores: make(map[string]float64),
		JournalMatched: reflect.DeepEqual(journal.Report, packet.BayesianShadow) && reflect.DeepEqual(journal.Snapshot, packet.Snapshot)}
	for _, c := range trace.before {
		fixture, ok := ids[c.ID]
		if !ok {
			return result{}, errors.New("ranker candidate has no fixture identity")
		}
		out.Frontier = append(out.Frontier, fixture)
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
	if err := readJSON(set+"corpus.json", &facts); err != nil {
		return err
	}
	if err := readJSON(set+"queries.json", &queries); err != nil {
		return err
	}
	if len(facts) != 24 || len(queries) != 72 {
		return errors.New("frozen broad USGS fixture shape changed")
	}
	var metadata struct {
		CapturedAt string `json:"captured_at"`
	}
	if err := readJSON(set+"source-metadata.json", &metadata); err != nil {
		return err
	}
	now, err := time.Parse(time.RFC3339Nano, metadata.CapturedAt)
	if err != nil {
		return fmt.Errorf("invalid frozen source capture clock: %w", err)
	}
	em, err := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
		URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768,
		Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: ",
	})
	if err != nil {
		return err
	}
	results := make([]result, 0, 144)
	for _, arm := range []string{"focus", "task-lexical"} {
		for _, q := range queries {
			caseCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			r, err := runCase(caseCtx, em, facts, q, arm, now)
			cancel()
			if err != nil {
				return fmt.Errorf("%s/%s: %w", arm, q.ID, err)
			}
			results = append(results, r)
			fmt.Println("completed", arm, q.ID)
		}
	}
	if len(results) != 144 {
		return errors.New("incomplete broad USGS confirmation")
	}
	paths := []string{
		root + "USGS_TADINE_CURRENT_CONTRACT.md", root + "USGS_BROAD_CONFIRMATION_PROTOCOL.md",
		root + "prepare_usgs_broad.py", set + "source.geojson", set + "source-metadata.json",
		set + "captured-at.txt", set + "corpus.json", set + "queries.json",
		"cmd/public-usgs-broad/main.go", "internal/researchcalendar/task_plan.go",
		"internal/researchcalendar/lexical.go", "internal/researchcalendar/priority.go",
		"internal/service/service.go", "internal/embed/openai.go",
	}
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(raw)
		hashes[path] = hex.EncodeToString(digest[:])
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
	}{em.ModelKey(), modelDigest, hashes, results}); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func main() {
	if len(os.Args) != 2 {
		panic("usage: public-usgs-broad NEW.json")
	}
	if err := run(os.Args[1]); err != nil {
		panic(err)
	}
}
