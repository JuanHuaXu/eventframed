// Frozen cross-corpus retrieval research. No relevance path or fitter exists here.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	f "github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	h "github.com/JuanHuaXu/eventframed/internal/researchpublichybrid"
	idf "github.com/JuanHuaXu/eventframed/internal/researchpublicidf"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	p "github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type query struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type projection struct {
	Partition string  `json:"partition"`
	Queries   []query `json:"queries"`
}
type weights struct {
	Weights []float64 `json:"weights"`
}

func load(path string, v any) error {
	file, e := os.Open(path)
	if e != nil {
		return e
	}
	defer file.Close()
	d := json.NewDecoder(io.LimitReader(file, 128<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return errors.New("trailing input")
	}
	return nil
}
func model(w weights) (r.Model, error) {
	m := r.Model{}
	if len(w.Weights) != r.Dimension {
		return m, errors.New("exactly eight weights required")
	}
	for i, v := range w.Weights {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 4 || (i >= 6 && v != 0) {
			return r.Model{}, errors.New("invalid source-only frozen weights")
		}
		m.Weights[i] = v
	}
	return m, nil
}
func checkQueries(q projection) error {
	if q.Partition != "nfcorpus-official-test" || len(q.Queries) != 323 {
		return errors.New("whole NFCorpus test required")
	}
	seen := map[string]bool{}
	for _, v := range q.Queries {
		if v.ID == "" || seen[v.ID] || len(v.ID) > 256 || !utf8.ValidString(v.ID) || v.Text == "" || len(v.Text) > 4096 || !utf8.ValidString(v.Text) {
			return errors.New("invalid query identity/text")
		}
		seen[v.ID] = true
	}
	return nil
}
func encodeNew(path string, v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	b = append(b, '\n')
	file, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return nil, e
	}
	_, e = file.Write(b)
	if e == nil {
		e = file.Sync()
	}
	closeErr := file.Close()
	if e != nil {
		return nil, e
	}
	return b, closeErr
}

// Empty frontiers are real no-answer cases, never discarded or padded.
func rank(ctx context.Context, m r.Model, rows []r.Row) ([]r.Scored, []r.Scored, error) {
	if ctx == nil {
		return nil, nil, errors.New("nil rank context")
	}
	if e := ctx.Err(); e != nil {
		return nil, nil, e
	}
	if len(rows) == 0 {
		return []r.Scored{}, []r.Scored{}, nil
	}
	base, e := (r.Model{}).Rank(ctx, rows)
	if e != nil {
		return nil, nil, e
	}
	learned, e := m.Rank(ctx, rows)
	return base, learned, e
}
func run(args []string) error {
	if len(args) != 4 {
		return errors.New("usage: corpus.json test-only.json frozen-model.json NEW-output-directory")
	}
	began := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	input, e := os.Open(args[0])
	if e != nil {
		return e
	}
	inputBytes, e := io.ReadAll(io.LimitReader(input, f.MaxInputBytes+1))
	closeErr := input.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	records, e := f.Decode(inputBytes)
	if e != nil {
		return e
	}
	if len(records) != 3633 {
		return errors.New("whole NFCorpus corpus required")
	}
	var q projection
	var w weights
	if e = load(args[1], &q); e != nil {
		return e
	}
	if e = checkQueries(q); e != nil {
		return e
	}
	if e = load(args[2], &w); e != nil {
		return e
	}
	m, e := model(w)
	if e != nil {
		return e
	}
	asOf := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	config := f.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: asOf}
	t := time.Now()
	docs, e := f.Convert(ctx, records, config)
	if e != nil {
		return e
	}
	convertNS := time.Since(t).Nanoseconds()
	t = time.Now()
	reg, e := p.New(ctx, records, config, "public-scientific")
	if e != nil {
		return e
	}
	poolNS := time.Since(t).Nanoseconds()
	entries := reg.Entries()
	// Seal actual framed payloads, not import acknowledgments or full-text proxies.
	if _, e = encodeNew(filepath.Join(args[3], "frames.json"), docs); e != nil {
		return e
	}
	b, e := encodeNew(filepath.Join(args[3], "pool.json"), entries)
	if e != nil {
		return e
	}
	digest := sha256.Sum256(b)
	epoch := hex.EncodeToString(digest[:])
	hd := make([]h.Document, len(entries))
	sd := make([]r.Source, len(entries))
	byID := map[string]p.Entry{}
	for i, v := range entries {
		hd[i] = h.Document{ID: v.Candidate.ID, Text: v.Candidate.Text, AvailableAt: v.AvailableAt}
		sd[i] = r.Source{ID: v.Candidate.ID, Title: records[i].Title, Body: v.Candidate.Text, AvailableAt: v.AvailableAt}
		byID[v.Candidate.ID] = v
	}
	t = time.Now()
	index, e := h.New(ctx, hd, asOf)
	if e != nil {
		return e
	}
	indexNS := time.Since(t).Nanoseconds()
	t = time.Now()
	features, e := idf.NewSources(ctx, sd, asOf, epoch)
	if e != nil {
		return e
	}
	featureIndexNS := time.Since(t).Nanoseconds()
	out, e := os.OpenFile(filepath.Join(args[3], "predictions.ndjson"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	if e = enc.Encode(map[string]any{"kind": "start", "dataset": "nfcorpus", "partition": q.Partition, "contract": idf.Contract, "epoch": epoch, "asOf": asOf, "model": m, "documents": len(records), "queries": len(q.Queries), "stats": index.Stats(), "conversionNS": convertNS, "poolNS": poolNS, "indexNS": indexNS, "featureIndexNS": featureIndexNS, "setupNS": time.Since(began).Nanoseconds(), "labelsRead": false, "nativeDatabase": false}); e != nil {
		return e
	}
	queryBegan := time.Now()
	writeNS := int64(0)
	for i, v := range q.Queries {
		t = time.Now()
		hits, e := index.Search(ctx, v.Text, 200)
		if e != nil {
			return e
		}
		searchNS := time.Since(t).Nanoseconds()
		t = time.Now()
		candidates := make([]retrieval.Candidate, len(hits))
		max := 0.
		for j, h := range hits {
			entry, ok := byID[h.ID]
			if !ok {
				return errors.New("unbound nominee")
			}
			candidates[j] = entry.Candidate
			candidates[j].Score = h.Score
			if h.Score > max {
				max = h.Score
			}
		}
		if _, e = reg.Bind(ctx, candidates, asOf, 200); e != nil {
			return e
		}
		hydrateNS := time.Since(t).Nanoseconds()
		t = time.Now()
		rows := make([]r.Row, len(hits))
		for j, hit := range hits {
			fv, e := features.Features(ctx, v.Text, hit.ID, epoch, 0)
			if e != nil {
				return e
			}
			base := 0.
			if max > 0 {
				base = hit.Score / max
			}
			rows[j] = r.Row{ID: hit.ID, Base: base, Features: fv}
		}
		featureNS := time.Since(t).Nanoseconds()
		t = time.Now()
		base, learned, e := rank(ctx, m, rows)
		if e != nil {
			return e
		}
		rankNS := time.Since(t).Nanoseconds()
		t = time.Now()
		e = enc.Encode(map[string]any{"kind": "query", "index": i, "id": v.ID, "text": v.Text, "epoch": epoch, "lexical": hits, "rows": rows, "baseline": base, "learned": learned, "searchNS": searchNS, "hydrateNS": hydrateNS, "featureNS": featureNS, "rankNS": rankNS})
		writeNS += time.Since(t).Nanoseconds()
		if e != nil {
			return e
		}
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = enc.Encode(map[string]any{"kind": "complete", "queries": 323, "queryPhaseNS": time.Since(queryBegan).Nanoseconds(), "wholeBeforeFinalSyncNS": time.Since(began).Nanoseconds(), "queryWriteNS": writeNS, "labelsRead": false, "fittedOnNF": false}); e != nil {
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
