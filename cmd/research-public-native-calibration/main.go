// A fresh owned native retrieval trial. No label file, default endpoint, or
// production source is accessible through the command's declared inputs.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	f "github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	h "github.com/JuanHuaXu/eventframed/internal/researchpublichybrid"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	p "github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	mask "github.com/JuanHuaXu/eventframed/internal/researchpublicrankmask"
	resume "github.com/JuanHuaXu/eventframed/internal/researchpublicresume"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"io"
	"os"
	"path/filepath"
	"time"
)

type query struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type models struct {
	Primary   r.Model `json:"primary"`
	Secondary r.Model `json:"secondary"`
}

type projection struct {
	Partition string  `json:"partition"`
	Queries   []query `json:"queries"`
}

func strictJSON(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}

func decodeProjection(b []byte) (projection, error) {
	var input projection
	if e := strictJSON(b, &input); e != nil {
		return input, e
	}
	if input.Partition != "calibration" || len(input.Queries) != 180 {
		return input, errors.New("full calibration required")
	}
	seen := map[string]bool{}
	for _, q := range input.Queries {
		if q.ID == "" || q.Text == "" || len(q.Text) > 4096 || seen[q.ID] {
			return input, errors.New("bad query")
		}
		seen[q.ID] = true
	}
	return input, nil
}

func decodeModels(b []byte) (models, error) {
	var input struct {
		Primary struct {
			Weights []float64 `json:"weights"`
		} `json:"primary"`
		Secondary struct {
			Weights []float64 `json:"weights"`
		} `json:"secondary"`
	}
	var m models
	if e := strictJSON(b, &input); e != nil {
		return m, e
	}
	if len(input.Primary.Weights) != 8 || len(input.Secondary.Weights) != 8 {
		return m, errors.New("exact eight weights required for both arms")
	}
	copy(m.Primary.Weights[:], input.Primary.Weights)
	copy(m.Secondary.Weights[:], input.Secondary.Weights)
	for _, model := range []r.Model{m.Primary, m.Secondary} {
		if model.Weights[6] != 0 || model.Weights[7] != 0 {
			return m, errors.New("native features not masked in model")
		}
		if _, e := model.Score(r.Row{ID: "validation"}); e != nil {
			return m, e
		}
	}
	return m, nil
}

func run(args []string) error {
	if len(args) != 4 {
		return errors.New("usage: CORPUS CALIBRATION_ONLY OWNED_ROOT FROZEN_MODELS")
	}
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	root, e := filepath.Abs(args[2])
	if e != nil {
		return e
	}
	reader, e := resume.Open(root, cwd)
	if e != nil {
		return e
	}
	defer reader.Close()
	b, e := os.ReadFile(args[0])
	if e != nil {
		return e
	}
	records, e := f.Decode(b)
	if e != nil {
		return e
	}
	if len(records) != 5183 {
		return errors.New("whole corpus required")
	}
	b, e = os.ReadFile(args[1])
	if e != nil {
		return e
	}
	input, e := decodeProjection(b)
	if e != nil {
		return e
	}
	b, e = os.ReadFile(args[3])
	if e != nil {
		return e
	}
	m, e := decodeModels(b)
	if e != nil {
		return e
	}
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	reg, e := p.New(ctx, records, f.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: clock}, "public-scientific")
	if e != nil {
		return e
	}
	client, e := retrieval.OpenLibraVDBContractsWithConfig(retrieval.ContractClientConfig{Endpoint: "unix:" + filepath.Join(root, "native.sock"), TLSMode: "insecure", MaxConcurrent: 1, RequestTimeout: 30 * time.Second, MaxAttempts: 1})
	if e != nil {
		return e
	}
	defer client.Close()
	if e = client.Ready(ctx); e != nil {
		return e
	}
	trace := filepath.Join(root, "trace.ndjson")
	out, e := os.OpenFile(trace, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	write := func(v any) error {
		if e := enc.Encode(v); e != nil {
			return e
		}
		return out.Sync()
	}
	started := time.Now()
	if e = write(map[string]any{"kind": "start", "partition": "calibration", "documents": 5183, "queries": 180, "labelsRead": false}); e != nil {
		return e
	}
	entries := reg.Entries()
	ids := make([]string, len(entries))
	cache := map[string]retrieval.Candidate{}
	title := map[string]string{}
	for _, d := range records {
		title[d.ID] = d.Title
	}
	for i, v := range entries {
		ids[i] = v.Candidate.ID
	}
	if e = resume.Verify(ctx, reg, ids, clock, reader, func(i int, v p.Entry, rows []retrieval.Candidate, ns int64) error {
		cache[v.Candidate.ID] = rows[0]
		return write(map[string]any{"kind": "verify", "index": i, "id": v.Candidate.ID, "source": v.SourceID, "rows": rows, "ns": ns, "ok": true})
	}); e != nil {
		return e
	}
	prefix, e := os.ReadFile(trace)
	if e != nil {
		return e
	}
	digest := sha256.Sum256(prefix)
	epoch := hex.EncodeToString(digest[:])
	docs := make([]h.Document, len(entries))
	sources := make([]r.Source, len(entries))
	for i, v := range entries {
		row := cache[v.Candidate.ID]
		docs[i] = h.Document{ID: row.ID, Text: row.Text, AvailableAt: v.AvailableAt}
		sources[i] = r.Source{ID: row.ID, Title: title[v.SourceID], Body: row.Text, AvailableAt: v.AvailableAt}
	}
	t := time.Now()
	index, e := h.New(ctx, docs, clock)
	if e != nil {
		return e
	}
	features, e := r.NewSources(ctx, sources, clock, epoch)
	if e != nil {
		return e
	}
	indexNS := time.Since(t).Nanoseconds()
	if e = write(map[string]any{"kind": "sealed", "epoch": epoch, "indexNS": indexNS, "documents": len(cache), "stats": index.Stats()}); e != nil {
		return e
	}
	for i, q := range input.Queries {
		t = time.Now()
		native, e := client.SearchTextCollections(ctx, retrieval.SearchRequest{Collections: []string{reg.Collection()}, QueryText: q.Text, K: 200})
		if e != nil {
			return e
		}
		searchNS := time.Since(t).Nanoseconds()
		t = time.Now()
		if _, e = reg.Bind(ctx, native, clock, 200); e != nil {
			return e
		}
		bindNS := time.Since(t).Nanoseconds()
		nom := make([]h.Hit, len(native))
		nr := map[string]int{}
		for j, v := range native {
			nom[j] = h.Hit{ID: v.ID, Score: v.Score}
			nr[v.ID] = j + 1
		}
		t = time.Now()
		lex, e := index.Search(ctx, q.Text, 200)
		if e != nil {
			return e
		}
		fused, e := index.Fuse(ctx, nom, lex, 200)
		if e != nil {
			return e
		}
		lexical, e := index.Rerank(ctx, q.Text, fused)
		if e != nil {
			return e
		}
		lexNS := time.Since(t).Nanoseconds()
		if len(fused) != 200 {
			return errors.New("complete fused200 required")
		}
		t = time.Now()
		payload := make([]retrieval.Candidate, len(fused))
		for j, v := range fused {
			c, ok := cache[v.ID]
			if !ok {
				return errors.New("unverified source hydration")
			}
			c.Metadata = append([]byte(nil), c.Metadata...)
			c.Score = v.Score
			payload[j] = c
		}
		if _, e = reg.Bind(ctx, payload, clock, 200); e != nil {
			return e
		}
		hydrateNS := time.Since(t).Nanoseconds()
		t = time.Now()
		max := lexical[0].Score
		byScore := map[string]float64{}
		for _, v := range lexical {
			byScore[v.ID] = v.Score
		}
		rows := make([]r.Row, len(fused))
		for j, v := range fused {
			fv, e := features.Features(ctx, q.Text, v.ID, epoch, nr[v.ID])
			if e != nil {
				return e
			}
			base := 0.
			if max > 0 {
				base = byScore[v.ID] / max
			}
			rows[j] = r.Row{ID: v.ID, Base: base, Features: fv}
		}
		rows, e = mask.SourceOnly(ctx, rows)
		if e != nil {
			return e
		}
		featureNS := time.Since(t).Nanoseconds()
		t = time.Now()
		baseline, e := (r.Model{}).Rank(ctx, rows)
		if e != nil {
			return e
		}
		primary, e := m.Primary.Rank(ctx, rows)
		if e != nil {
			return e
		}
		secondary, e := m.Secondary.Rank(ctx, rows)
		if e != nil {
			return e
		}
		rankNS := time.Since(t).Nanoseconds()
		if e = write(map[string]any{"kind": "query", "index": i, "id": q.ID, "text": q.Text, "epoch": epoch, "native": native, "lexical": lex, "fused": fused, "hydrated": payload, "features": rows, "baseline": baseline, "primary": primary, "secondary": secondary, "searchNS": searchNS, "bindNS": bindNS, "lexicalNS": lexNS, "hydrateNS": hydrateNS, "featureNS": featureNS, "rankNS": rankNS}); e != nil {
			return e
		}
		if (i+1)%30 == 0 {
			fmt.Printf("queried %d/180\n", i+1)
		}
	}
	return write(map[string]any{"kind": "complete", "queries": 180, "documents": 5183, "labelsRead": false, "confirmationPredictions": 0, "totalNS": time.Since(started).Nanoseconds()})
}
func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
