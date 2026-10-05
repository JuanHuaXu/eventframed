// public-ecma-magnitude never opens task labels or fits inside Recall. Each
// query/arm gets a fresh memory store and service; only embedding vectors share.
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
	"github.com/JuanHuaXu/eventframed/internal/researchmagnitude"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const fixture = "research/public-task-pilot/ecmascript-v1/"
const digest = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"

type fact struct {
	ID    string                                            `json:"id"`
	Frame struct{ Who, What, Where, When, Why, How string } `json:"frame"`
}
type query struct{ ID, Split, Wording, Text string }
type arm struct {
	Name   string
	Weight float64
}
type item struct {
	ID        string             `json:"id"`
	Score     float64            `json:"score"`
	Retrieval float64            `json:"retrieval"`
	Law       model.BernoulliLaw `json:"law"`
}
type record struct {
	Type, Arm, Case, Split     string
	Weight                     float64
	Order, Packed              []item
	Lexical, HookScores        []float64
	HookCalls                  int
	HookNS, RecallNS, ImportNS int64
	MemoBefore, MemoAfter      researchfusion.MemoStats
	JournalStored              bool
	Frames                     map[string][6]string `json:",omitempty"`
}

func read(p string, dst any) error {
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, dst)
}
func verify(files map[string]string) error {
	for p, want := range files {
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			return fmt.Errorf("source freeze changed %s", p)
		}
	}
	return nil
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
		return errors.New("embedding inventory failure")
	}
	var data struct {
		Models []struct{ Name, Digest string }
	}
	if e = json.NewDecoder(r.Body).Decode(&data); e != nil {
		return e
	}
	for _, m := range data.Models {
		if m.Name == "nomic-embed-text:latest" && m.Digest == digest {
			return nil
		}
	}
	return errors.New("frozen embedding unavailable")
}
func runCase(ctx context.Context, em *researchfusion.Memo, facts []fact, q query, a arm, saveFrames bool) (record, error) {
	out := record{Type: "case", Arm: a.Name, Case: q.ID, Split: q.Split, Weight: a.Weight}
	tap, e := service.NewResearchFrontierTap("research", 1)
	if e != nil {
		return out, e
	}
	defer tap.Close()
	config := service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, ResearchFrontier: tap}
	if a.Name != "baseline" {
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
				scores, e := researchmagnitude.Scores(ctx, x, a.Weight)
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
	// This sealed public import is already structured. Do not silently replace
	// it with the text extractor or treat the WHEN string as availability time.
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	ids := map[string]string{}
	eventIDs := []string{}
	start := time.Now()
	for i, f := range facts {
		id := fmt.Sprintf("ecma-record-%03d", i)
		field := func(x string) model.Field { return model.Field{Value: x, Source: model.SourceObserved, Confidence: 1} }
		v := model.Event{ID: id, TenantID: "research", SessionID: "public-import", Sequence: uint64(i + 1), Kind: "public_program",
			Content: f.Frame.How, OccurredAt: now, ObservedAt: now, AvailableAt: now, Priority: .5,
			Who: field(f.Frame.Who), What: field(f.Frame.What), Where: field(f.Frame.Where), When: field(f.Frame.When), Why: field(f.Frame.Why), How: field(f.Frame.How), Provenance: model.Provenance{Producer: "public-ecma-import"}}
		r, e := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: v})
		if e != nil {
			return out, e
		}
		ids[r.EventID] = f.ID
		eventIDs = append(eventIDs, r.EventID)
	}
	out.ImportNS = time.Since(start).Nanoseconds()
	if saveFrames {
		ev, e := db.GetEvents(ctx, "research", eventIDs, now.Add(time.Minute))
		if e != nil {
			return out, e
		}
		out.Frames = map[string][6]string{}
		for _, v := range ev {
			out.Frames[ids[v.ID]] = [6]string{v.Who.Value, v.What.Value, v.Where.Value, v.When.Value, v.Why.Value, v.How.Value}
		}
	}
	if _, e = embed.QueryContext(ctx, em, frame.QueryText(q.Text)); e != nil {
		return out, e
	}
	out.MemoBefore = em.Stats()
	start = time.Now()
	p, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "fresh-query", Query: q.Text, AsOf: now.Add(time.Minute), RecallK: 50, PackK: 10, TokenBudget: 10000})
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
		return out, errors.New("frontier binding")
	}
	decisions := make(map[string]model.ForecastBundle, len(p.BayesianShadow.Decisions))
	for _, d := range p.BayesianShadow.Decisions {
		decisions[d.EventID] = d.Forecast
	}
	for _, c := range full.Candidates {
		forecast, ok := decisions[c.EventID]
		if ids[c.EventID] == "" || !ok {
			return out, errors.New("unknown frontier ID")
		}
		out.Order = append(out.Order, item{ID: ids[c.EventID], Score: forecast.RankScore, Law: forecast.CorrectedLaw})
	}
	for _, c := range p.Candidates {
		if ids[c.Event.ID] == "" {
			return out, errors.New("unknown packet ID")
		}
		out.Packed = append(out.Packed, item{ids[c.Event.ID], c.Score, c.RetrievalScore, c.Forecast.CorrectedLaw})
	}
	return out, nil
}
func run(phase, modelPath, output string) error {
	var files map[string]string
	if e := read(fixture+"magnitude-runtime-freeze.json", &files); e != nil {
		return e
	}
	if e := verify(files); e != nil {
		return e
	}
	var facts []fact
	var queries []query
	if e := read(fixture+"prepared/corpus.json", &facts); e != nil {
		return e
	}
	if e := read(fixture+"prepared/queries.json", &queries); e != nil {
		return e
	}
	if len(facts) != 54 || len(queries) != 108 {
		return errors.New("fixture count changed")
	}
	arms := []arm{{"baseline", 0}, {"incumbent", 0}, {"w025", .25}, {"w050", .5}, {"w075", .75}, {"w100", 1}}
	modelSHA := ""
	if phase == "heldout" {
		var m struct {
			Weight  float64 `json:"weight"`
			FitOnly bool    `json:"fit_only"`
		}
		if e := read(modelPath, &m); e != nil {
			return e
		}
		if !m.FitOnly {
			return errors.New("model not fit-only")
		}
		b, _ := os.ReadFile(modelPath)
		h := sha256.Sum256(b)
		modelSHA = hex.EncodeToString(h[:])
		arms = []arm{{"baseline", 0}, {"incumbent", 0}, {"selected", m.Weight}}
	} else if phase != "fit" {
		return errors.New("bad phase")
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
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"type": "header", "files": files, "digest": digest, "phase": phase, "model_sha256": modelSHA, "time": time.Now().UTC()}); e != nil {
		return e
	}
	count := 0
	for qi, q := range queries {
		if (phase == "fit") != (q.Split == "fit") {
			continue
		}
		for ai := range arms {
			a := arms[(ai+qi)%len(arms)]
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			r, e := runCase(ctx, em, facts, q, a, count == 0)
			cancel()
			if e != nil {
				return fmt.Errorf("%s/%s: %w", q.ID, a.Name, e)
			}
			if e = enc.Encode(r); e != nil {
				return e
			}
			if e = f.Sync(); e != nil {
				return e
			}
			count++
		}
	}
	if count != 216 {
		return fmt.Errorf("incomplete cases %d", count)
	}
	if e = verify(files); e != nil {
		return e
	}
	if e = inventory(context.Background()); e != nil {
		return e
	}
	if phase == "heldout" {
		b, e := os.ReadFile(modelPath)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != modelSHA {
			return errors.New("model changed")
		}
	}
	if e = enc.Encode(map[string]any{"type": "footer", "complete": count, "memo": em.Stats(), "hashes_verified": true}); e != nil {
		return e
	}
	return f.Sync()
}
func main() {
	if len(os.Args) != 4 {
		panic("usage: public-ecma-magnitude fit UNUSED OUTPUT or heldout MODEL OUTPUT")
	}
	if e := run(os.Args[1], os.Args[2], os.Args[3]); e != nil {
		panic(e)
	}
}
