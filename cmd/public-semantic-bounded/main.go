// public-semantic-bounded is an isolated, public-only embedding ablation.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchsparse"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

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
	Arm, Case             string
	Candidates            []candidate
	Frontier, SupportRank int
	RecallNS              int64
}

func read(p string, v any) {
	b, e := os.ReadFile(p)
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, v); e != nil {
		panic(e)
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("usage: public-semantic-bounded NEW.json")
	}
	if _, e := os.Stat(os.Args[1]); !os.IsNotExist(e) {
		panic("output exists or inaccessible")
	}
	root := "research/public-task-pilot/"
	var facts []fact
	var queries []query
	var weights struct{ Weights []float64 }
	read(root+"nobel-v1/corpus.json", &facts)
	read(root+"nobel-v1/queries.json", &queries)
	read(root+"sparse-go-golden.json", &weights)
	frozen, e := researchsparse.New(weights.Weights)
	if e != nil {
		panic(e)
	}
	em, e := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768, Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: "})
	if e != nil {
		panic(e)
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	results := []result{}
	for _, arm := range []string{"baseline", "sparse", "bounded"} {
		for _, q := range queries {
			frontier := 0
			config := service.Config{DefaultRecallK: 50, DefaultPackK: 20, DefaultTokenBudget: 10000}
			if arm != "baseline" {
				config.ResearchRanking = service.ResearchRankingPolicy{TenantID: "research", SparseASCII: true, DirectOrder: arm == "sparse", Score: func(ctx context.Context, in service.ResearchRankInput) ([]float64, error) {
					frontier = len(in.Candidates)
					out := make([]float64, frontier)
					for i, c := range in.Candidates {
						if e := ctx.Err(); e != nil {
							return nil, e
						}
						p, e := frozen.Score(*c.Sparse)
						if e != nil {
							return nil, e
						}
						if arm == "bounded" {
							out[i] = math.Max(0, math.Min(1, c.Baseline+.01*(p-.5)))
						} else {
							out[i] = p
						}
					}
					return out, nil
				}}
			}
			s, e := service.New(memorystore.New(), em, config)
			if e != nil {
				panic(e)
			}
			ids := map[string]string{}
			for i, f := range facts {
				id := fmt.Sprintf("public-record-%03d", i)
				r, e := s.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "seed", Sequence: uint64(i + 1), UserText: f.Text, AssistantText: "Recorded.", OccurredAt: now, ObservedAt: now, AvailableAt: now}})
				if e != nil {
					panic(e)
				}
				ids[r.EventID] = f.ID
			}
			start := time.Now()
			p, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query", Query: q.Question, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 20, TokenBudget: 10000})
			if e != nil {
				panic(e)
			}
			r := result{Arm: arm, Case: q.ID, Frontier: frontier, RecallNS: time.Since(start).Nanoseconds()}
			for _, c := range p.Candidates {
				r.Candidates = append(r.Candidates, candidate{ids[c.Event.ID], c.Score, c.Forecast.CorrectedLaw})
			}
			results = append(results, r)
			s.Close()
			fmt.Println("completed", arm, q.ID)
		}
	}
	var oracle map[string]struct {
		Support []string `json:"support"`
	}
	read(root+"nobel-v1/oracle.json", &oracle)
	for i := range results {
		for j, c := range results[i].Candidates {
			for _, id := range oracle[results[i].Case].Support {
				if c.Fixture == id {
					results[i].SupportRank = j + 1
				}
			}
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{root + "BOUNDED_SEMANTIC_PROTOCOL.md", root + "nobel-v1/corpus.json", root + "nobel-v1/queries.json", root + "nobel-v1/oracle.json", root + "sparse-go-golden.json", "cmd/public-semantic-bounded/main.go", "internal/researchsparse/sparse.go", "internal/embed/openai.go", "internal/service/research_rank.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	f, e := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if e = enc.Encode(struct {
		Model   string
		Hashes  map[string]string
		Results []result
	}{em.ModelKey(), hashes, results}); e != nil {
		panic(e)
	}
	if e = f.Close(); e != nil {
		panic(e)
	}
}
