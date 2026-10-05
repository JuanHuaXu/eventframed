// Package researchpublicidf tests source-discriminative rank features only.
// Corpus statistics are bound to an immutable source epoch, never to outcomes.
package researchpublicidf

import (
	"context"
	"errors"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
)

const Contract = "public-source-idf-features-v1"

type terms struct {
	title, body map[string]bool
	order       []string
}
type Sources struct {
	raw  *r.Sources
	docs map[string]terms
	df   map[string]int
	n    int
}

func tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsDigit(c) })
}
func set(ts []string) map[string]bool {
	s := map[string]bool{}
	for _, t := range ts {
		s[t] = true
	}
	return s
}

func NewSources(ctx context.Context, docs []r.Source, asOf time.Time, epoch string) (*Sources, error) {
	// The base contract rejects every malformed/future source before any df use.
	raw, e := r.NewSources(ctx, docs, asOf, epoch)
	if e != nil {
		return nil, e
	}
	s := &Sources{raw: raw, docs: map[string]terms{}, df: map[string]int{}, n: len(docs)}
	for _, d := range docs {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		t := tokens(d.Title)
		v := terms{set(t), set(tokens(d.Body)), t}
		s.docs[d.ID] = v
		union := map[string]bool{}
		for t := range v.title {
			union[t] = true
		}
		for t := range v.body {
			union[t] = true
		}
		for t := range union {
			s.df[t]++
		}
	}
	return s, ctx.Err()
}

func (s *Sources) weight(t string) float64 {
	df := float64(s.df[t])
	return math.Log1p((float64(s.n) - df + .5) / (df + .5))
}

// Weighted overlap retains absent query terms in the denominator. Native cues
// are not features; they can still nominate candidates outside this package.
func (s *Sources) Features(ctx context.Context, query, id, epoch string, nativeRank int) (r.Vector, error) {
	if s == nil {
		return r.Vector{}, errors.New("nil IDF source index")
	}
	v, e := s.raw.Features(ctx, query, id, epoch, nativeRank)
	if e != nil {
		return r.Vector{}, e
	}
	v[0], v[1], v[2], v[3], v[6], v[7] = 0, 0, 0, 0, 0, 0
	d := s.docs[id]
	o := tokens(query)
	u := set(o)
	ordered := make([]string, 0, len(u))
	for t := range u {
		ordered = append(ordered, t)
	}
	sort.Strings(ordered)
	total, numeric := 0., 0.
	for _, t := range ordered {
		w := s.weight(t)
		total += w
		if d.title[t] {
			v[0] += w
		}
		if d.body[t] {
			v[1] += w
		}
		for _, c := range t {
			if unicode.IsDigit(c) {
				numeric += w
				if d.body[t] {
					v[3] += w
				}
				break
			}
		}
	}
	v[0] /= total
	v[1] /= total
	if numeric > 0 {
		v[3] /= numeric
	}
	bigram := 0.
	for i := 0; i+1 < len(o); i++ {
		w := (s.weight(o[i]) + s.weight(o[i+1])) / 2
		bigram += w
		for j := 0; j+1 < len(d.order); j++ {
			if o[i] == d.order[j] && o[i+1] == d.order[j+1] {
				v[2] += w
				break
			}
		}
	}
	if bigram > 0 {
		v[2] /= bigram
	}
	if e := ctx.Err(); e != nil {
		return r.Vector{}, e
	}
	return v, nil
}
