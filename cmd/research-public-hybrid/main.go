// This offline predictor reads only public source text and frozen FIT queries.
// Native nominations are replayed; annotations are evaluator-only inputs.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublichybrid"
)

type entry struct {
	SourceID    string                    `json:"source_id"`
	AvailableAt time.Time                 `json:"available_at"`
	Candidate   struct{ ID, Text string } `json:"candidate"`
}
type query struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type projection struct {
	Partition string  `json:"partition"`
	Queries   []query `json:"queries"`
}
type native struct {
	Kind   string `json:"kind"`
	Index  int    `json:"index"`
	ID     string `json:"id"`
	Text   string `json:"text"`
	Search []struct {
		ID, Text string
		Score    float64
	} `json:"search"`
}

func load(p string, v any) error {
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 256<<20))
	if e = d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}
func run(args []string) error {
	if len(args) != 4 {
		return errors.New("usage: research-public-hybrid pool.json fit-only.json native-trace.ndjson NEW-output.ndjson")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var pool []entry
	var fit projection
	if e := load(args[0], &pool); e != nil {
		return e
	}
	if e := load(args[1], &fit); e != nil {
		return e
	}
	if len(pool) != 5183 || fit.Partition != "fit" || len(fit.Queries) != 351 {
		return errors.New("whole corpus/fit partition required")
	}
	asOf := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	docs := make([]researchpublichybrid.Document, len(pool))
	byID := map[string]entry{}
	for i, p := range pool {
		docs[i] = researchpublichybrid.Document{ID: p.Candidate.ID, Text: p.Candidate.Text, AvailableAt: p.AvailableAt}
		byID[p.Candidate.ID] = p
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	began := time.Now()
	x, e := researchpublichybrid.New(ctx, docs, asOf)
	if e != nil {
		return e
	}
	buildNS := time.Since(began).Nanoseconds()
	runtime.ReadMemStats(&after)
	if x.Stats().Documents != 5183 {
		return errors.New("incomplete as-of corpus")
	}
	in, e := os.Open(args[2])
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.OpenFile(args[3], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	if e = enc.Encode(map[string]any{"kind": "start", "contract": researchpublichybrid.Contract, "partition": "fit", "documents": 5183, "queries": 351, "stats": x.Stats(), "buildNS": buildNS, "buildAllocatedBytes": after.TotalAlloc - before.TotalAlloc}); e != nil {
		return e
	}
	scan := bufio.NewScanner(in)
	scan.Buffer(make([]byte, 64<<10), 16<<20)
	i := 0
	complete := false
	began = time.Now()
	for scan.Scan() {
		if e = ctx.Err(); e != nil {
			return e
		}
		if complete {
			return errors.New("native data after completion")
		}
		var n native
		if e = json.Unmarshal(scan.Bytes(), &n); e != nil {
			return e
		}
		if n.Kind == "start" || n.Kind == "verify" {
			if i != 0 {
				return errors.New("native phase after query")
			}
			continue
		}
		if n.Kind == "complete" {
			if i != 351 {
				return errors.New("incomplete native queries")
			}
			complete = true
			continue
		}
		if n.Kind != "query" || i >= 351 || n.Index != i || n.ID != fit.Queries[i].ID || n.Text != fit.Queries[i].Text {
			return errors.New("native fit query mismatch")
		}
		nom := make([]researchpublichybrid.Hit, len(n.Search))
		for j, h := range n.Search {
			p, ok := byID[h.ID]
			if !ok || h.Text != p.Candidate.Text {
				return errors.New("native source/text mismatch")
			}
			nom[j] = researchpublichybrid.Hit{ID: h.ID, Score: h.Score}
		}
		t := time.Now()
		lex, e := x.Search(ctx, n.Text, 200)
		if e != nil {
			return e
		}
		lexNS := time.Since(t).Nanoseconds()
		t = time.Now()
		rerank, e := x.Rerank(ctx, n.Text, nom)
		if e != nil {
			return e
		}
		rerankNS := time.Since(t).Nanoseconds()
		t = time.Now()
		fused, e := x.Fuse(ctx, nom, lex, 200)
		if e != nil {
			return e
		}
		fusionNS := time.Since(t).Nanoseconds()
		if e = enc.Encode(map[string]any{"kind": "query", "index": i, "id": n.ID, "text": n.Text, "native": nom, "lexical": lex, "withinNative": rerank, "fused": fused, "lexicalNS": lexNS, "rerankNS": rerankNS, "fusionNS": fusionNS}); e != nil {
			return e
		}
		i++
	}
	if e = scan.Err(); e != nil {
		return e
	}
	if !complete {
		return errors.New("native completion absent")
	}
	if e = enc.Encode(map[string]any{"kind": "complete", "queries": i, "queryPhaseNS": time.Since(began).Nanoseconds(), "labelsRead": false, "calibrationPredictions": 0, "confirmationPredictions": 0}); e != nil {
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
