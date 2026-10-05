// Offline cross-fit research; only already-consumed FIT outcomes are permitted.
// Held-fold judgments are excluded from Fit, not hidden from this process.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	l "github.com/JuanHuaXu/eventframed/internal/researchpubliclambdarank"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"io"
	"os"
	"time"
)

type row struct {
	ID         string  `json:"id"`
	Base       float64 `json:"base"`
	NativeRank int     `json:"nativeRank"`
}
type query struct {
	ID     string `json:"id"`
	Family string `json:"family"`
	Fold   int    `json:"fold"`
	Text   string `json:"text"`
	Rows   []row  `json:"rows"`
}
type input struct {
	Partition string     `json:"partition"`
	Epoch     string     `json:"epoch"`
	AsOf      time.Time  `json:"asOf"`
	Sources   []r.Source `json:"sources"`
	Queries   []query    `json:"queries"`
}
type outcome struct {
	Query    string   `json:"query"`
	Positive []string `json:"positive"`
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
		return errors.New("trailing input")
	}
	return nil
}
func save(p string, v any) error {
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(v); e != nil {
		return e
	}
	return f.Sync()
}
func run(args []string) error {
	if len(args) != 4 {
		return errors.New("usage: INPUT.json FIT_LABELS.json NEW_FEATURES.json NEW_RESULT.json")
	}
	var in input
	if e := load(args[0], &in); e != nil {
		return e
	}
	if in.Partition != "fit" || len(in.Sources) != 5183 || len(in.Queries) != 351 {
		return errors.New("whole fit corpus required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	began := time.Now()
	s, e := r.NewSources(ctx, in.Sources, in.AsOf, in.Epoch)
	if e != nil {
		return e
	}
	sourceNS := time.Since(began).Nanoseconds()
	cases := make([]r.Case, len(in.Queries))
	seen := map[string]bool{}
	featureNS := make([]int64, len(cases))
	for i, q := range in.Queries {
		if q.ID == "" || seen[q.ID] || q.Family == "" {
			return errors.New("invalid query")
		}
		seen[q.ID] = true
		h := sha256.Sum256([]byte(q.Family))
		if q.Fold != int(binary.BigEndian.Uint32(h[:4])%5) {
			return errors.New("wrong family fold")
		}
		c := r.Case{ID: q.ID, Family: q.Family, Rows: make([]r.Row, len(q.Rows))}
		t := time.Now()
		for j, v := range q.Rows {
			f, e := s.Features(ctx, q.Text, v.ID, in.Epoch, v.NativeRank)
			if e != nil {
				return e
			}
			c.Rows[j] = r.Row{ID: v.ID, Base: v.Base, Features: f}
		}
		if _, e := (r.Model{}).Rank(ctx, c.Rows); e != nil {
			return e
		}
		featureNS[i] = time.Since(t).Nanoseconds()
		cases[i] = c
	}
	if e = save(args[2], map[string]any{"contract": r.Contract, "epoch": in.Epoch, "cases": cases, "sourceBuildNS": sourceNS, "featureNS": featureNS, "labelsRead": false}); e != nil {
		return e
	}
	var labels []outcome
	if e = load(args[1], &labels); e != nil {
		return e
	}
	if len(labels) != 351 {
		return errors.New("fit labels count")
	}
	for i, l := range labels {
		if l.Query != cases[i].ID || len(l.Positive) == 0 {
			return errors.New("fit label projection mismatch")
		}
		p := map[string]bool{}
		for _, id := range l.Positive {
			if p[id] {
				return errors.New("duplicate target")
			}
			p[id] = true
		}
		cases[i].Positive = p
	}
	type fitted struct {
		Fold     int        `json:"fold"`
		Model    r.Model    `json:"model"`
		Stats    r.FitStats `json:"stats"`
		TrainIDs []string   `json:"trainIDs"`
		NS       int64      `json:"ns"`
	}
	models := make([]fitted, 0, 6)
	for fold := -1; fold < 5; fold++ {
		train := []r.Case{}
		ids := []string{}
		for i, c := range cases {
			if in.Queries[i].Fold != fold {
				train = append(train, c)
				ids = append(ids, c.ID)
			}
		}
		t := time.Now()
		m, stats, e := l.Fit(ctx, train)
		if e != nil {
			return e
		}
		models = append(models, fitted{fold, m, stats, ids, time.Since(t).Nanoseconds()})
	}
	pred := make([]map[string]any, len(cases))
	for i, c := range cases {
		t := time.Now()
		baseline, e := (r.Model{}).Rank(ctx, c.Rows)
		if e != nil {
			return e
		}
		learned, e := models[in.Queries[i].Fold+1].Model.Rank(ctx, c.Rows)
		if e != nil {
			return e
		}
		pred[i] = map[string]any{"id": c.ID, "family": c.Family, "fold": in.Queries[i].Fold, "baseline": baseline, "learned": learned, "rankNS": time.Since(t).Nanoseconds()}
	}
	return save(args[3], map[string]any{"contract": r.Contract, "epoch": in.Epoch, "models": models, "predictions": pred, "fitLabelsRead": true, "heldFoldLabelsExcludedFromFit": true, "calibrationPredictions": 0, "confirmationPredictions": 0})
}
func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
