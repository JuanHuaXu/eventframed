package researchpublicfeedback

import (
	"context"
	"errors"
	"math"
	"sort"
	"unicode/utf8"
)

type Term struct {
	Text   string  `json:"text"`
	Weight float64 `json:"weight"`
}

// Feedback is bounded nomination only. Positive retrieval scores are not
// authenticated observations, truth probabilities or posterior support.
func (x *Index) Feedback(ctx context.Context, query string, initial []Hit) ([]Term, error) {
	if ctx == nil || x == nil {
		return nil, errors.New("invalid feedback context")
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if e := x.validate(initial); e != nil {
		return nil, e
	}
	// Callers cannot nominate a truncated or reordered set as feedback evidence.
	expected, e := x.Search(ctx, query, 200)
	if e != nil {
		return nil, e
	}
	if len(initial) != len(expected) {
		return nil, errors.New("feedback nomination mismatch")
	}
	for i, h := range initial {
		if h != expected[i] {
			return nil, errors.New("feedback nomination mismatch")
		}
	}
	original := map[string]bool{}
	for _, t := range tokens(query) {
		original[t] = true
	}
	feedback := map[string]float64{}
	limit := len(initial)
	if limit > 10 {
		limit = 10
	}
	for _, h := range initial[:limit] {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		freq := x.freqs[x.byID[h.ID]]
		terms := []Term{}
		for t, n := range freq {
			length := utf8.RuneCountInString(t)
			if length >= 2 && length <= 20 && float64(len(x.postings[t]))/float64(len(x.ids)) <= .1 {
				terms = append(terms, Term{t, float64(n)})
			}
		}
		termOrder(terms)
		if len(terms) > 10 {
			terms = terms[:10]
		}
		norm := 0.
		for _, v := range terms {
			norm += v.Weight
		}
		if norm == 0 {
			continue
		}
		for _, v := range terms {
			feedback[v.Text] += v.Weight / norm * h.Score
		}
	}
	selected := []Term{}
	for t, w := range feedback {
		selected = append(selected, Term{t, w})
	}
	termOrder(selected)
	if len(selected) > 10 {
		selected = selected[:10]
	}
	norm := 0.
	for _, v := range selected {
		norm += v.Weight
	}
	mixed := map[string]float64{}
	origWeight := 1.
	if norm > 0 {
		origWeight = .5
	}
	for t := range original {
		mixed[t] = origWeight / float64(len(original))
	}
	if norm > 0 {
		for _, v := range selected {
			mixed[v.Text] += .5 * v.Weight / norm
		}
	}
	out := make([]Term, 0, len(mixed))
	for t, w := range mixed {
		out = append(out, Term{t, w})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Text < out[j].Text })
	return out, ctx.Err()
}
func termOrder(v []Term) {
	sort.Slice(v, func(i, j int) bool {
		if v[i].Weight == v[j].Weight {
			return v[i].Text < v[j].Text
		}
		return v[i].Weight > v[j].Weight
	})
}

// SearchWeighted scores the complete eligible posting universe. The cap is
// applied after weighting, never separately per expansion term.
func (x *Index) SearchWeighted(ctx context.Context, terms []Term, cap int) ([]Hit, error) {
	if e := validCap(ctx, cap); e != nil {
		return nil, e
	}
	if x == nil || len(terms) == 0 || len(terms) > 266 {
		return nil, errors.New("invalid weighted query")
	}
	seen := map[string]bool{}
	sum := 0.
	for _, v := range terms {
		t := tokens(v.Text)
		if len(t) != 1 || t[0] != v.Text || seen[v.Text] || !utf8.ValidString(v.Text) || math.IsNaN(v.Weight) || math.IsInf(v.Weight, 0) || v.Weight <= 0 || v.Weight > 1 {
			return nil, errors.New("invalid weighted term")
		}
		seen[v.Text] = true
		sum += v.Weight
	}
	if math.Abs(sum-1) > 1e-10 {
		return nil, errors.New("term mass must sum to one")
	}
	sorted := append([]Term(nil), terms...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Text < sorted[j].Text })
	scores := map[int]float64{}
	avg := float64(x.stats.Tokens) / float64(len(x.ids))
	for _, term := range sorted {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		p := x.postings[term.Text]
		df := float64(len(p))
		idf := math.Log1p((float64(len(x.ids)) - df + .5) / (df + .5))
		for j, v := range p {
			if j%256 == 0 {
				if e := ctx.Err(); e != nil {
					return nil, e
				}
			}
			tf := float64(v.tf)
			scores[v.doc] += term.Weight * idf * tf * 2.2 / (tf + 1.2*(.25+.75*float64(x.lengths[v.doc])/avg))
		}
	}
	out := make([]Hit, 0, len(scores))
	for id, score := range scores {
		out = append(out, Hit{x.ids[id], score})
	}
	order(out)
	if len(out) > cap {
		out = out[:cap]
	}
	return out, ctx.Err()
}
