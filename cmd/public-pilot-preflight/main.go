// public-pilot-preflight exercises post-contract ingestion and pilot retrieval.
// It neither calls an LLM nor trains a learner. Oracle data is opened only after
// all requested retrievals finish. Consumed tasks are not fresh confirmation.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
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
	Split    string `json:"split"`
	Question string `json:"question"`
}
type row struct {
	Fixture  string               `json:"fixture"`
	Features uint16               `json:"features"`
	Score    float64              `json:"score"`
	Frame    string               `json:"frame"`
	Backend  float64              `json:"backend_score"`
	Delta    float64              `json:"research_delta"`
	Law      model.ForecastBundle `json:"law_bundle"`
}
type result struct {
	Case              string  `json:"case"`
	Candidates        []row   `json:"candidates"`
	SupportRank       int     `json:"support_rank"`
	FeatureCollisions int     `json:"support_feature_collisions"`
	Frontier          int     `json:"callback_frontier"`
	Certainty         float64 `json:"packet_certainty"`
}

func read(path string, v any) {
	b, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, v); e != nil {
		panic(e)
	}
}
func main() {
	root := flag.String("input", "research/public-task-pilot/prepared-v2", "fixture directory")
	out := flag.String("output", "", "new output file (required)")
	split := flag.String("split", "design", "design or confirmation")
	ranking := flag.String("ranking", "baseline", "baseline, lexical or sparse")
	weights := flag.String("weights", "research/public-task-pilot/sparse-go-golden.json", "frozen weights export")
	pack := flag.Int("pack", 10, "packed candidate count")
	flag.Parse()
	if *ranking != "baseline" && *ranking != "lexical" && *ranking != "sparse" {
		panic("invalid ranking")
	}
	if *split != "design" && *split != "confirmation" {
		panic("invalid split")
	}
	if *out == "" {
		panic("output required")
	}
	var facts []fact
	var qs []query
	read(filepath.Join(*root, "corpus.json"), &facts)
	read(filepath.Join(*root, "queries.json"), &qs)
	var fitted researchsparse.Frozen
	if *ranking == "sparse" {
		var export struct{ Weights []float64 }
		read(*weights, &export)
		var e error
		fitted, e = researchsparse.New(export.Weights)
		if e != nil {
			panic(e)
		}
	}
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	results := []result{}
	for _, q := range qs {
		if q.Split != *split {
			continue
		}
		em, e := embed.NewHashEmbedder(256)
		if e != nil {
			panic(e)
		}
		frontier := 0
		config := service.Config{DefaultRecallK: 50, DefaultPackK: *pack, DefaultTokenBudget: 10000}
		if *ranking != "baseline" {
			config.ResearchRanking = service.ResearchRankingPolicy{TenantID: "research", SparseASCII: true, DirectOrder: true, Score: func(ctx context.Context, in service.ResearchRankInput) ([]float64, error) {
				frontier = len(in.Candidates)
				scores := make([]float64, frontier)
				for i, c := range in.Candidates {
					if e := ctx.Err(); e != nil {
						return nil, e
					}
					if *ranking == "lexical" {
						scores[i] = c.Sparse.Values()[7]
					} else {
						p, e := fitted.Score(*c.Sparse)
						if e != nil {
							return nil, e
						}
						scores[i] = p
					}
				}
				return scores, nil
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
		p, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query", Query: q.Question, AsOf: now.Add(time.Minute), RecallK: 50, PackK: *pack, TokenBudget: 10000})
		if e != nil {
			panic(e)
		}
		r := result{Case: q.ID, Frontier: frontier, Certainty: p.PacketAnswerCertainty}
		for _, c := range p.Candidates {
			f, e := researchmemory.Extract(q.Question, c.Event, c.Forecast.PreResidualLaw.Useful)
			if e != nil {
				panic(e)
			}
			law := c.Forecast
			law.RankScore = 0
			r.Candidates = append(r.Candidates, row{Fixture: ids[c.Event.ID], Features: f, Score: c.Score, Frame: c.Event.FrameText(), Backend: c.RetrievalScore, Delta: c.ResearchRankDelta, Law: law})
		}
		results = append(results, r)
		s.Close()
	}
	// Evaluation only: oracle labels never enter capture, query or scoring.
	var oracle map[string]struct {
		Support []string `json:"support"`
	}
	read(filepath.Join(*root, "oracle.json"), &oracle)
	for i := range results {
		r := &results[i]
		for j, c := range r.Candidates {
			for _, support := range oracle[r.Case].Support {
				if support == c.Fixture {
					r.SupportRank = j + 1
					for _, other := range r.Candidates {
						if other.Fixture != support && other.Features == c.Features {
							r.FeatureCollisions++
						}
					}
				}
			}
		}
	}
	f, e := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		panic(e)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if e = enc.Encode(results); e != nil {
		panic(e)
	}
	if e = f.Close(); e != nil {
		panic(e)
	}
	for _, r := range results {
		fmt.Printf("%s packed=%d support_rank=%d feature_collisions=%d\n", r.Case, len(r.Candidates), r.SupportRank, r.FeatureCollisions)
	}
}
