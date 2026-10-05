// public-git-fusion collects a prospectively frozen, public-only retrieval
// screen. It neither opens labels nor trains, mutates production or generates
// answers. Every arm/query gets a new service/store/session.
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
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchfusion"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const fixture = "research/public-task-pilot/git-fusion-v1/"
const digest = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"

var arms = []string{"baseline", "incumbent", "lexical", "rrf", "protected"}

type fact struct {
	ID   string `json:"fixture_id"`
	Text string `json:"text"`
}
type query struct {
	ID       string `json:"case_id"`
	Question string `json:"question"`
}
type item struct {
	ID    string             `json:"id"`
	Score float64            `json:"score"`
	Law   model.BernoulliLaw `json:"law"`
}
type record struct {
	Type                 string                   `json:"type"`
	Arm                  string                   `json:"arm"`
	Case                 string                   `json:"case"`
	Order                []item                   `json:"order"`
	Packed               []item                   `json:"packed"`
	Lexical              []float64                `json:"lexical"`
	HookScores           []float64                `json:"hook_scores"`
	HookCalls            int                      `json:"hook_calls"`
	HookNS               int64                    `json:"hook_ns"`
	RecallNS             int64                    `json:"recall_ns"`
	CaptureNS            int64                    `json:"capture_ns"`
	CorrelatedSuppressed int                      `json:"correlated_suppressed"`
	JournalStored        bool                     `json:"journal_stored"`
	MemoBefore           researchfusion.MemoStats `json:"memo_before"`
	MemoAfter            researchfusion.MemoStats `json:"memo_after"`
	Frames               map[string][6]string     `json:"frames,omitempty"`
}

func read(path string, out any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, out)
}
func inventory(ctx context.Context) error {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:11434/api/tags", nil)
	if e != nil {
		return e
	}
	r, e := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if e != nil {
		return e
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return errors.New("inventory HTTP failure")
	}
	var data struct {
		Models []struct {
			Name   string
			Digest string
		}
	}
	if e = json.NewDecoder(r.Body).Decode(&data); e != nil {
		return e
	}
	for _, m := range data.Models {
		if m.Name == "nomic-embed-text:latest" && m.Digest == digest {
			return nil
		}
	}
	return errors.New("frozen local embedding unavailable")
}
func verify(hashes map[string]string) error {
	for path, want := range hashes {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			return fmt.Errorf("freeze changed: %s", path)
		}
	}
	return nil
}
func runCase(ctx context.Context, em *researchfusion.Memo, facts []fact, q query, arm string, saveFrames bool) (record, error) {
	out := record{Type: "case", Arm: arm, Case: q.ID}
	tap, e := service.NewResearchFrontierTap("research", 1)
	if e != nil {
		return out, e
	}
	defer tap.Close()
	config := service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, ResearchFrontier: tap}
	if arm != "baseline" {
		config.ResearchRanking = service.ResearchRankingPolicy{TenantID: "research", SparseASCII: true, DirectOrder: true,
			Score: func(ctx context.Context, in service.ResearchRankInput) ([]float64, error) {
				start := time.Now()
				defer func() { out.HookNS += time.Since(start).Nanoseconds() }()
				out.HookCalls++
				x := make([]float64, len(in.Candidates))
				for i, c := range in.Candidates {
					if c.Sparse == nil {
						return nil, errors.New("missing sparse fields")
					}
					x[i] = c.Sparse.Values()[7]
				}
				out.Lexical = append([]float64(nil), x...)
				var scores []float64
				var e error
				switch arm {
				case "incumbent":
					scores = make([]float64, len(x))
					for i := range scores {
						scores[i] = 1 / float64(i+1)
					}
				case "lexical":
					scores = append([]float64(nil), x...)
				case "rrf":
					scores, e = researchfusion.Scores(ctx, x, 0)
				case "protected":
					scores, e = researchfusion.Scores(ctx, x, 10)
				default:
					return nil, errors.New("unknown arm")
				}
				out.HookScores = append([]float64(nil), scores...)
				return scores, e
			}}
	}
	db := memorystore.New()
	s, e := service.New(db, em, config)
	if e != nil {
		return out, e
	}
	defer s.Close()
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	ids := make(map[string]string, len(facts))
	eventIDs := make([]string, 0, len(facts))
	start := time.Now()
	for i, f := range facts {
		id := fmt.Sprintf("public-record-%03d", i)
		c, e := s.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id,
			Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "fixture", Sequence: uint64(i + 1), UserText: f.Text, AssistantText: "Recorded.", OccurredAt: now, ObservedAt: now, AvailableAt: now}})
		if e != nil {
			return out, e
		}
		ids[c.EventID] = f.ID
		eventIDs = append(eventIDs, c.EventID)
	}
	out.CaptureNS = time.Since(start).Nanoseconds()
	if saveFrames {
		events, e := db.GetEvents(ctx, "research", eventIDs, now.Add(time.Minute))
		if e != nil {
			return out, e
		}
		out.Frames = make(map[string][6]string, len(events))
		for _, v := range events {
			out.Frames[ids[v.ID]] = [6]string{v.Who.Value, v.What.Value, v.Where.Value, v.When.Value, v.Why.Value, v.How.Value}
		}
	}
	// Exact same query representation/prefix as Recall, outside timed serving.
	if _, e = embed.QueryContext(ctx, em, frame.QueryText(q.Question)); e != nil {
		return out, e
	}
	out.MemoBefore = em.Stats()
	start = time.Now()
	p, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query", Query: q.Question, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 10, TokenBudget: 10000})
	out.RecallNS = time.Since(start).Nanoseconds()
	out.MemoAfter = em.Stats()
	if e != nil {
		return out, e
	}
	j, e := db.GetBayesianJournal(ctx, "research", p.BayesianShadow.JournalID)
	if e != nil {
		return out, e
	}
	if !reflect.DeepEqual(j.Report, p.BayesianShadow) {
		return out, errors.New("packet/journal disagreement")
	}
	out.JournalStored = true
	full, e := tap.Take(ctx)
	if e != nil {
		return out, e
	}
	if full.JournalID != j.ID || full.Snapshot != p.Snapshot {
		return out, errors.New("tap journal mismatch")
	}
	decisions := make(map[string]model.ForecastBundle, len(p.BayesianShadow.Decisions))
	for _, d := range p.BayesianShadow.Decisions {
		decisions[d.EventID] = d.Forecast
	}
	for _, c := range full.Candidates {
		f, ok := decisions[c.EventID]
		if !ok || ids[c.EventID] == "" {
			return out, errors.New("missing frontier binding")
		}
		out.Order = append(out.Order, item{ids[c.EventID], f.RankScore, f.CorrectedLaw})
	}
	for _, c := range p.Candidates {
		if ids[c.Event.ID] == "" {
			return out, errors.New("unknown packet ID")
		}
		out.Packed = append(out.Packed, item{ids[c.Event.ID], c.Score, c.Forecast.CorrectedLaw})
	}
	out.CorrelatedSuppressed = p.CorrelatedSuppressed
	return out, nil
}
func run(path string) error {
	var hashes map[string]string
	if e := read(fixture+"freeze.json", &hashes); e != nil {
		return e
	}
	if e := verify(hashes); e != nil {
		return e
	}
	var facts []fact
	var queries []query
	if e := read(fixture+"corpus.json", &facts); e != nil {
		return e
	}
	if e := read(fixture+"queries.json", &queries); e != nil {
		return e
	}
	if len(facts) != 48 || len(queries) != 100 {
		return errors.New("frozen fixture shape changed")
	}
	if e := inventory(context.Background()); e != nil {
		return e
	}
	base, e := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: "http://127.0.0.1:11434/v1/embeddings", Model: "nomic-embed-text:latest", Dimension: 768, Timeout: 60 * time.Second, DocumentPrefix: "search_document: ", QueryPrefix: "search_query: "})
	if e != nil {
		return e
	}
	em, e := researchfusion.NewMemo(base, 512)
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"type": "header", "hashes": hashes, "digest": digest, "model": em.ModelKey(), "arms": arms, "cases": len(queries), "utc": time.Now().UTC()}); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	for qi, q := range queries {
		for ai := range arms {
			arm := arms[(ai+qi)%len(arms)]
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			r, e := runCase(ctx, em, facts, q, arm, qi == 0 && ai == 0)
			cancel()
			if e != nil {
				return fmt.Errorf("%s/%s: %w", arm, q.ID, e)
			}
			if e = enc.Encode(r); e != nil {
				return e
			}
			if e = f.Sync(); e != nil {
				return e
			}
			fmt.Println("completed", arm, q.ID)
		}
	}
	if e = verify(hashes); e != nil {
		return e
	}
	if e = inventory(context.Background()); e != nil {
		return e
	}
	if e = enc.Encode(map[string]any{"type": "footer", "complete": 500, "memo": em.Stats(), "hashes_verified": true}); e != nil {
		return e
	}
	return f.Sync()
}
func main() {
	if len(os.Args) != 2 {
		panic("usage: public-git-fusion NEW.jsonl")
	}
	if e := run(os.Args[1]); e != nil {
		panic(e)
	}
}
