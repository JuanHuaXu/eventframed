// public-calendar-margin is an isolated, public-only embedding ablation.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
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
	PacketCertainty       float64
	Arm, Case             string
	Original, Effective   string
	Candidates            []candidate
	Frontier, SupportRank int
	RecallNS              int64
	Before, After         []retrieval.Candidate
	Laws                  map[string]model.BernoulliLaw
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
	if len(os.Args) != 3 {
		panic("usage: public-calendar-margin SET NEW.json")
	}
	if _, e := os.Stat(os.Args[2]); !os.IsNotExist(e) {
		panic("output exists or inaccessible")
	}
	root := "research/public-task-pilot/"
	set := os.Args[1]
	if set != "esa-date-v1" && set != "esa-iso-v1" && set != "w3c-focus-v1" {
		panic("unknown dataset")
	}
	var facts []fact
	var queries []query
	read(root+set+"/corpus.json", &facts)
	read(root+set+"/queries.json", &queries)
	em, e := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768, Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: "})
	if e != nil {
		panic(e)
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	results := []result{}
	for _, arm := range []string{"focus", "calendar", "margin"} {
		for _, q := range queries {
			frontier := 0
			config := service.Config{MaxRankDelta: .25, DefaultRecallK: 50, DefaultPackK: 1, DefaultTokenBudget: 10000}
			effective := q.Question
			effective = focusQuery(effective)
			trace := &traceRanker{enabled: arm != "focus", margin: arm == "margin"}
			config.CandidateRanker = trace
			config.CandidateRankerRequired = true
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
			p, e := s.Recall(researchcalendar.Bind(ctx, "research", q.Question), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query", Query: effective, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 1, TokenBudget: 10000})
			if e != nil {
				panic(e)
			}
			r := result{PacketCertainty: p.PacketAnswerCertainty, Before: trace.before, After: trace.after, Laws: map[string]model.BernoulliLaw{}, Arm: arm, Case: q.ID, Original: q.Question, Effective: effective, Frontier: frontier, RecallNS: time.Since(start).Nanoseconds()}
			for _, c := range p.Candidates {
				r.Candidates = append(r.Candidates, candidate{ids[c.Event.ID], c.Score, c.Forecast.CorrectedLaw})
			}
			for _, d := range p.BayesianShadow.Decisions {
				r.Laws[ids[d.EventID]] = d.Forecast.CorrectedLaw
			}
			results = append(results, r)
			s.Close()
			fmt.Println("completed", arm, q.ID)
		}
	}
	var oracle map[string]struct {
		Support []string `json:"support"`
	}
	read(root+set+"/oracle.json", &oracle)
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
	for _, p := range []string{root + "CALENDAR_MARGIN_PROTOCOL.md", root + set + "/corpus.json", root + set + "/queries.json", root + set + "/oracle.json", "cmd/public-calendar-margin/main.go", "internal/researchcalendar/ranker.go", "internal/researchcalendar/margin.go", "internal/researchcalendar/ranker_test.go", "internal/frame/query.go", "internal/frame/turn.go", "internal/model/event.go", "internal/embed/openai.go", "internal/service/research_rank.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	f, e := os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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

// This research ablation drops an explicit alternative, not a logical exclusion engine.
func focusQuery(q string) string {
	const separator = ", rather than "
	if strings.Count(q, separator) != 1 {
		return q
	}
	prefix, suffix, _ := strings.Cut(q, separator)
	if strings.TrimSpace(prefix) == "" || strings.TrimSpace(suffix) == "" {
		return q
	}
	return strings.TrimRight(strings.TrimSpace(prefix), "?") + "?"
}

// Each experiment query owns one trace; the production adapter is stateless.
type traceRanker struct {
	enabled       bool
	margin        bool
	before, after []retrieval.Candidate
}

func (t *traceRanker) ContractName() string {
	if t.margin {
		return (researchcalendar.MarginRanker{MaxDelta: .25}).ContractName()
	}
	if t.enabled {
		return (researchcalendar.Ranker{}).ContractName()
	}
	return (retrieval.PassthroughRanker{}).ContractName()
}
func (t *traceRanker) RankCandidates(ctx context.Context, r retrieval.RankRequest) ([]retrieval.Candidate, error) {
	t.before = append([]retrieval.Candidate(nil), r.Candidates...)
	var out []retrieval.Candidate
	var err error
	if t.margin {
		out, err = (researchcalendar.MarginRanker{MaxDelta: .25}).RankCandidates(ctx, r)
	} else if t.enabled {
		out, err = (researchcalendar.Ranker{}).RankCandidates(ctx, r)
	} else {
		out, err = (retrieval.PassthroughRanker{}).RankCandidates(ctx, r)
	}
	t.after = append([]retrieval.Candidate(nil), out...)
	return out, err
}
