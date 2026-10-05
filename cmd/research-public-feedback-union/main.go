// Development-only frozen query expansion; no labels or fit operation.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	fb "github.com/JuanHuaXu/eventframed/internal/researchpublicfeedback"
	u "github.com/JuanHuaXu/eventframed/internal/researchpublicfrontierunion"
	idf "github.com/JuanHuaXu/eventframed/internal/researchpublicidf"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"io"
	"math"
	"os"
	"time"
)

type query struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Family string `json:"family"`
	Fold   int    `json:"fold"`
}
type input struct {
	Partition string     `json:"partition"`
	Epoch     string     `json:"epoch"`
	AsOf      time.Time  `json:"asOf"`
	Sources   []r.Source `json:"sources"`
	Queries   []query    `json:"queries"`
}
type frozen struct {
	Fold    int       `json:"fold"`
	Weights []float64 `json:"weights"`
}

func load(p string, v any) error {
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 128<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}
func models(v []frozen) ([]r.Model, error) {
	if len(v) != 5 {
		return nil, errors.New("five held-fold models required")
	}
	out := make([]r.Model, 5)
	for i, v := range v {
		if v.Fold != i || len(v.Weights) != 8 {
			return nil, errors.New("wrong model fold or dimension")
		}
		for j, w := range v.Weights {
			if math.IsNaN(w) || math.IsInf(w, 0) || math.Abs(w) > 4 || (j >= 6 && w != 0) {
				return nil, errors.New("invalid source-only weight")
			}
			out[i].Weights[j] = w
		}
	}
	return out, nil
}
func rows(ctx context.Context, s *idf.Sources, q query, epoch string, hits []fb.Hit) ([]r.Row, error) {
	out := make([]r.Row, len(hits))
	max := 0.
	for _, h := range hits {
		if h.Score > max {
			max = h.Score
		}
	}
	for j, h := range hits {
		f, e := s.Features(ctx, q.Text, h.ID, epoch, 0)
		if e != nil {
			return nil, e
		}
		base := 0.
		if max > 0 {
			base = h.Score / max
		}
		out[j] = r.Row{ID: h.ID, Base: base, Features: f}
	}
	return out, nil
}
func rank(ctx context.Context, m r.Model, rows []r.Row) ([]r.Scored, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if len(rows) == 0 {
		return []r.Scored{}, nil
	}
	return m.Rank(ctx, rows)
}
func run(args []string) error {
	if len(args) != 3 {
		return errors.New("usage: INPUT.json MODELS.json NEW-trace.ndjson")
	}
	began := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var in input
	var weights []frozen
	if e := load(args[0], &in); e != nil {
		return e
	}
	if e := load(args[1], &weights); e != nil {
		return e
	}
	model, e := models(weights)
	if e != nil {
		return e
	}
	if in.Partition != "fit" || len(in.Sources) != 5183 || len(in.Queries) != 351 {
		return errors.New("whole FIT required")
	}
	seen := map[string]bool{}
	for _, q := range in.Queries {
		h := sha256.Sum256([]byte(q.Family))
		if q.ID == "" || seen[q.ID] || q.Family == "" || q.Fold != int(binary.BigEndian.Uint32(h[:4])%5) {
			return errors.New("query/family/fold mismatch")
		}
		seen[q.ID] = true
	}
	docs := make([]fb.Document, len(in.Sources))
	for j, v := range in.Sources {
		docs[j] = fb.Document{ID: v.ID, Text: v.Body, AvailableAt: v.AvailableAt}
	}
	t := time.Now()
	x, e := fb.New(ctx, docs, in.AsOf)
	if e != nil {
		return e
	}
	buildNS := time.Since(t).Nanoseconds()
	if x.Stats().Documents != 5183 {
		return errors.New("source cutoff omitted rows")
	}
	t = time.Now()
	s, e := idf.NewSources(ctx, in.Sources, in.AsOf, in.Epoch)
	if e != nil {
		return e
	}
	featureBuildNS := time.Since(t).Nanoseconds()
	out, e := os.OpenFile(args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	if e = enc.Encode(map[string]any{"kind": "start", "unionRescue": true, "maxUnion": 400, "contract": fb.Contract, "partition": "fit", "epoch": in.Epoch, "asOf": in.AsOf, "documents": 5183, "queries": 351, "stats": x.Stats(), "models": weights, "buildNS": buildNS, "featureBuildNS": featureBuildNS, "setupNS": time.Since(began).Nanoseconds(), "labelsRead": false}); e != nil {
		return e
	}
	queryBegan := time.Now()
	writeNS := int64(0)
	for i, q := range in.Queries {
		t = time.Now()
		base, e := x.Search(ctx, q.Text, 200)
		if e != nil {
			return e
		}
		searchNS := time.Since(t).Nanoseconds()
		t = time.Now()
		terms, e := x.Feedback(ctx, q.Text, base)
		if e != nil {
			return e
		}
		feedbackNS := time.Since(t).Nanoseconds()
		t = time.Now()
		expanded, e := x.SearchWeighted(ctx, terms, 200)
		if e != nil {
			return e
		}
		expandedNS := time.Since(t).Nanoseconds()
		t = time.Now()
		union, e := u.OriginalScores(ctx, x, q.Text, base, expanded)
		if e != nil {
			return e
		}
		unionNS := time.Since(t).Nanoseconds()
		t = time.Now()
		br, e := rows(ctx, s, q, in.Epoch, base)
		if e != nil {
			return e
		}
		er, e := rows(ctx, s, q, in.Epoch, union)
		if e != nil {
			return e
		}
		featureNS := time.Since(t).Nanoseconds()
		t = time.Now()
		b, e := rank(ctx, r.Model{}, br)
		if e != nil {
			return e
		}
		bc, e := rank(ctx, model[q.Fold], br)
		if e != nil {
			return e
		}
		ex, e := u.Rank(ctx, r.Model{}, er, 200)
		if e != nil {
			return e
		}
		ec, e := u.Rank(ctx, model[q.Fold], er, 200)
		if e != nil {
			return e
		}
		rankNS := time.Since(t).Nanoseconds()
		t = time.Now()
		if e = enc.Encode(map[string]any{"kind": "query", "index": i, "id": q.ID, "text": q.Text, "family": q.Family, "fold": q.Fold, "epoch": in.Epoch, "baseline": base, "terms": terms, "expanded": expanded, "baseRows": br, "unionRows": er, "unionNominees": union, "plain": b, "plainIDF": bc, "union": ex, "unionIDF": ec, "searchNS": searchNS, "feedbackNS": feedbackNS, "expandedNS": expandedNS, "unionNS": unionNS, "featureNS": featureNS, "fourRankNS": rankNS}); e != nil {
			return e
		}
		writeNS += time.Since(t).Nanoseconds()
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = enc.Encode(map[string]any{"kind": "complete", "queries": 351, "labelsRead": false, "modelRefitted": false, "calibrationPredictions": 0, "confirmationPredictions": 0, "queryPhaseNS": time.Since(queryBegan).Nanoseconds(), "queryWriteNS": writeNS, "wholeBeforeFinalSyncNS": time.Since(began).Nanoseconds()}); e != nil {
		return e
	}
	return out.Sync()
}
func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
