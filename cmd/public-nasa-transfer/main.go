// public-nasa-transfer runs a frozen, public-only retrieval transfer screen.
// It never reads the oracle or trains on this fixture's labels.
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
	"github.com/JuanHuaXu/eventframed/internal/researchsparse"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const root = "research/public-task-pilot/"
const fixture = root + "nasa-transfer-v1/"
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
	Fixture string             `json:"fixture"`
	Score   float64            `json:"score"`
	Law     model.BernoulliLaw `json:"law"`
}

type result struct {
	Arm           string                        `json:"arm"`
	Case          string                        `json:"case"`
	Candidates    []candidate                   `json:"candidates"`
	Frontier      []string                      `json:"frontier"`
	Laws          map[string]model.BernoulliLaw `json:"laws"`
	RecallNS      int64                         `json:"recall_ns"`
	JournalStored bool                          `json:"journal_stored"`
}

func readJSON(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func checkModelDigest(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:11434/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("embedding inventory status %d", resp.StatusCode)
	}
	var inventory struct {
		Models []struct {
			Name   string
			Digest string
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&inventory); err != nil {
		return err
	}
	for _, item := range inventory.Models {
		if item.Name == "nomic-embed-text:latest" {
			if item.Digest != modelDigest {
				return errors.New("embedding model digest changed")
			}
			return nil
		}
	}
	return errors.New("declared embedding model unavailable")
}

func runCase(ctx context.Context, em embed.Embedder, facts []fact, q query, arm string, frozen researchsparse.Frozen, now time.Time) (result, error) {
	config := service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000}
	if arm != "baseline" {
		config.ResearchRanking = service.ResearchRankingPolicy{
			TenantID: "research", SparseASCII: true, DirectOrder: true,
			Score: func(ctx context.Context, in service.ResearchRankInput) ([]float64, error) {
				scores := make([]float64, len(in.Candidates))
				for i, c := range in.Candidates {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
					if arm == "lexical" {
						scores[i] = c.Sparse.Values()[7]
					} else {
						score, err := frozen.Score(*c.Sparse)
						if err != nil {
							return nil, err
						}
						scores[i] = score
					}
				}
				return scores, nil
			},
		}
	}
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
	packet, err := s.Recall(ctx, model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query",
		Query: q.Question, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 10, TokenBudget: 10000,
	})
	if err != nil {
		return result{}, err
	}
	out := result{Arm: arm, Case: q.ID, RecallNS: time.Since(start).Nanoseconds(), Laws: make(map[string]model.BernoulliLaw)}
	if _, err := memory.GetBayesianJournal(ctx, "research", packet.BayesianShadow.JournalID); err != nil {
		return result{}, err
	}
	out.JournalStored = true
	for _, decision := range packet.BayesianShadow.Decisions {
		id, ok := ids[decision.EventID]
		if !ok {
			return result{}, errors.New("unknown frontier event")
		}
		out.Frontier = append(out.Frontier, id)
		out.Laws[id] = decision.Forecast.CorrectedLaw
	}
	for _, c := range packet.Candidates {
		id, ok := ids[c.Event.ID]
		if !ok {
			return result{}, errors.New("unknown packed event")
		}
		out.Candidates = append(out.Candidates, candidate{id, c.Score, c.Forecast.CorrectedLaw})
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
	var export struct{ Weights []float64 }
	if err := readJSON(fixture+"corpus.json", &facts); err != nil {
		return err
	}
	if err := readJSON(fixture+"queries.json", &queries); err != nil {
		return err
	}
	if err := readJSON(root+"sparse-go-golden.json", &export); err != nil {
		return err
	}
	if len(facts) != 12 || len(queries) != 24 {
		return errors.New("frozen fixture shape changed")
	}
	frozen, err := researchsparse.New(export.Weights)
	if err != nil {
		return err
	}
	em, err := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
		URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768,
		Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: ",
	})
	if err != nil {
		return err
	}
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	results := make([]result, 0, 72)
	for _, arm := range []string{"baseline", "lexical", "sparse"} {
		for _, q := range queries {
			caseCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			r, err := runCase(caseCtx, em, facts, q, arm, frozen, now)
			cancel()
			if err != nil {
				return fmt.Errorf("%s/%s: %w", arm, q.ID, err)
			}
			results = append(results, r)
			fmt.Println("completed", arm, q.ID)
		}
	}
	paths := []string{
		root + "NASA_TRANSFER_V1_PROTOCOL.md", fixture + "corpus.json", fixture + "queries.json",
		root + "sparse-go-golden.json", "cmd/public-nasa-transfer/main.go",
		"internal/service/research_rank.go", "internal/researchsparse/sparse.go", "internal/embed/openai.go",
	}
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		hashes[path] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Model   string            `json:"model"`
		Digest  string            `json:"digest"`
		Hashes  map[string]string `json:"hashes"`
		Results []result          `json:"results"`
	}{em.ModelKey(), modelDigest, hashes, results}); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func main() {
	if len(os.Args) != 2 {
		panic("usage: public-nasa-transfer NEW.json")
	}
	if err := run(os.Args[1]); err != nil {
		panic(err)
	}
}
