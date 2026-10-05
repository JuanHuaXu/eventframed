// Package researchpublicfeedback is an isolated, frozen nomination experiment.
// It never reads labels, changes beliefs, or calls a production memory service.
package researchpublicfeedback

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const Contract = "public-source-feedback-bm25-v1"
const MaxFrontier = 200

type Document struct {
	ID          string
	Text        string
	AvailableAt time.Time
}
type Hit struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}
type posting struct{ doc, tf int }
type Stats struct {
	Documents int `json:"documents"`
	Terms     int `json:"terms"`
	Postings  int `json:"postings"`
	Tokens    int `json:"tokens"`
}
type Index struct {
	ids      []string
	lengths  []int
	byID     map[string]int
	postings map[string][]posting
	stats    Stats
	freqs    []map[string]int
}

func tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// New excludes future records before computing df and mean document length.
// The fixed as-of index can be queried concurrently, but is not an online store.
func New(ctx context.Context, docs []Document, asOf time.Time) (*Index, error) {
	if ctx == nil || asOf.IsZero() || len(docs) == 0 || len(docs) > 100000 {
		return nil, errors.New("invalid hybrid corpus")
	}
	x := &Index{byID: map[string]int{}, postings: map[string][]posting{}}
	seen := map[string]bool{}
	eligible := make([]Document, 0, len(docs))
	for _, d := range docs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if d.ID == "" || len(d.ID) > 256 || seen[d.ID] || d.AvailableAt.IsZero() || len(d.Text) == 0 || len(d.Text) > 16384 || !utf8.ValidString(d.Text) {
			return nil, errors.New("invalid hybrid document")
		}
		seen[d.ID] = true
		if !d.AvailableAt.After(asOf) {
			eligible = append(eligible, d)
		}
	}
	if len(eligible) == 0 {
		return nil, errors.New("no available hybrid documents")
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].ID < eligible[j].ID })
	for i, d := range eligible {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		t := tokens(d.Text)
		if len(t) == 0 {
			return nil, errors.New("empty hybrid document tokens")
		}
		x.ids = append(x.ids, d.ID)
		x.byID[d.ID] = i
		x.lengths = append(x.lengths, len(t))
		x.stats.Tokens += len(t)
		freq := map[string]int{}
		for _, term := range t {
			freq[term]++
		}
		x.freqs = append(x.freqs, freq)
		for term, tf := range freq {
			x.postings[term] = append(x.postings[term], posting{i, tf})
			x.stats.Postings++
		}
	}
	x.stats.Documents = len(x.ids)
	x.stats.Terms = len(x.postings)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return x, nil
}
func (x *Index) Stats() Stats {
	if x == nil {
		return Stats{}
	}
	return x.stats
}

func (x *Index) scores(ctx context.Context, query string) (map[int]float64, error) {
	if ctx == nil || x == nil || len(query) == 0 || len(query) > 4096 || !utf8.ValidString(query) {
		return nil, errors.New("invalid hybrid query")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	unique := map[string]bool{}
	for _, t := range tokens(query) {
		unique[t] = true
	}
	if len(unique) == 0 || len(unique) > 256 {
		return nil, errors.New("invalid hybrid query tokens")
	}
	terms := make([]string, 0, len(unique))
	for t := range unique {
		terms = append(terms, t)
	}
	sort.Strings(terms)
	values := map[int]float64{}
	avg := float64(x.stats.Tokens) / float64(len(x.ids))
	for _, term := range terms {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		p := x.postings[term]
		df := float64(len(p))
		idf := math.Log1p((float64(len(x.ids)) - df + .5) / (df + .5))
		for j, v := range p {
			if j%256 == 0 {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
			}
			tf := float64(v.tf)
			values[v.doc] += idf * tf * 2.2 / (tf + 1.2*(.25+.75*float64(x.lengths[v.doc])/avg))
		}
	}
	return values, nil
}

func order(h []Hit) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].Score == h[j].Score {
			return h[i].ID < h[j].ID
		}
		return h[i].Score > h[j].Score
	})
}
func validCap(ctx context.Context, cap int) error {
	if ctx == nil || cap < 1 || cap > MaxFrontier {
		return errors.New("invalid hybrid frontier cap")
	}
	return ctx.Err()
}
func (x *Index) validate(h []Hit) error {
	if x == nil || len(h) > MaxFrontier {
		return errors.New("invalid hybrid frontier")
	}
	seen := map[string]bool{}
	for _, v := range h {
		_, ok := x.byID[v.ID]
		if !ok || seen[v.ID] || math.IsNaN(v.Score) || math.IsInf(v.Score, 0) {
			return errors.New("invalid hybrid hit")
		}
		seen[v.ID] = true
	}
	return nil
}
func (x *Index) Search(ctx context.Context, query string, cap int) ([]Hit, error) {
	if err := validCap(ctx, cap); err != nil {
		return nil, err
	}
	s, err := x.scores(ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]Hit, 0, len(s))
	for i, v := range s {
		out = append(out, Hit{x.ids[i], v})
	}
	order(out)
	if len(out) > cap {
		out = out[:cap]
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Rerank preserves the entire nominated set, even rows with zero lexical score.
// It measures ordering separately from the changed nomination universe.
func (x *Index) Rerank(ctx context.Context, query string, nominated []Hit) ([]Hit, error) {
	if err := x.validate(nominated); err != nil {
		return nil, err
	}
	s, err := x.scores(ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]Hit, 0, len(nominated))
	for _, v := range nominated {
		out = append(out, Hit{v.ID, s[x.byID[v.ID]]})
	}
	order(out)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Fuse uses one-based ranks, not probabilities or incomparable native scores.
// Both complete input lists are inspected before the bounded union is truncated.
func (x *Index) Fuse(ctx context.Context, a, b []Hit, cap int) ([]Hit, error) {
	if err := validCap(ctx, cap); err != nil {
		return nil, err
	}
	if err := x.validate(a); err != nil {
		return nil, err
	}
	if err := x.validate(b); err != nil {
		return nil, err
	}
	s := map[string]float64{}
	for _, list := range [][]Hit{a, b} {
		for i, v := range list {
			s[v.ID] += 1 / float64(60+i+1)
		}
	}
	out := make([]Hit, 0, len(s))
	for id, v := range s {
		out = append(out, Hit{id, v})
	}
	order(out)
	if len(out) > cap {
		out = out[:cap]
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
