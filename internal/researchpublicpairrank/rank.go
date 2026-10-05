// Package researchpublicpairrank is a bounded research-only rank residual.
// Scores order source records; they are not confidence or forecast probabilities.
package researchpublicpairrank

import (
	"context"
	"encoding/hex"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const Contract = "public-source-pairrank-v1"
const Dimension = 8

type Vector [Dimension]float64
type Source struct {
	ID, Title, Body string
	AvailableAt     time.Time
}
type terms struct {
	title, body map[string]bool
	titleOrder  []string
	lengths     [2]int
}
type Sources struct {
	epoch   string
	records map[string]terms
}

func tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
func set(t []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range t {
		m[s] = true
	}
	return m
}
func NewSources(ctx context.Context, docs []Source, asOf time.Time, epoch string) (*Sources, error) {
	b, e := hex.DecodeString(epoch)
	if ctx == nil || asOf.IsZero() || e != nil || len(b) != 32 || len(docs) == 0 || len(docs) > 100000 {
		return nil, errors.New("invalid source snapshot")
	}
	s := &Sources{epoch: epoch, records: map[string]terms{}}
	for _, d := range docs {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if d.ID == "" || len(d.ID) > 256 || !utf8.ValidString(d.Title) || !utf8.ValidString(d.Body) || len(d.Title) == 0 || len(d.Title) > 16384 || len(d.Body) == 0 || len(d.Body) > 16384 || d.AvailableAt.IsZero() || d.AvailableAt.After(asOf) {
			return nil, errors.New("invalid or future source")
		}
		if _, ok := s.records[d.ID]; ok {
			return nil, errors.New("duplicate source")
		}
		a, b := tokens(d.Title), tokens(d.Body)
		if len(a) == 0 || len(b) == 0 {
			return nil, errors.New("source tokens empty")
		}
		s.records[d.ID] = terms{set(a), set(b), a, [2]int{len(a), len(b)}}
	}
	return s, ctx.Err()
}
func (s *Sources) Features(ctx context.Context, query, id, epoch string, nativeRank int) (Vector, error) {
	v := Vector{}
	if ctx == nil || s == nil || epoch != s.epoch || len(query) == 0 || len(query) > 4096 || !utf8.ValidString(query) || nativeRank < 0 || nativeRank > 200 {
		return v, errors.New("invalid feature context")
	}
	if e := ctx.Err(); e != nil {
		return v, e
	}
	d, ok := s.records[id]
	if !ok {
		return v, errors.New("unknown source")
	}
	order := tokens(query)
	q := set(order)
	if len(q) == 0 || len(q) > 256 {
		return v, errors.New("query tokens invalid")
	}
	numeric := 0
	for t := range q {
		if d.title[t] {
			v[0]++
		}
		if d.body[t] {
			v[1]++
		}
		hasDigit := false
		for _, r := range t {
			hasDigit = hasDigit || unicode.IsDigit(r)
		}
		if hasDigit {
			numeric++
			if d.body[t] {
				v[3]++
			}
		}
	}
	v[0] /= float64(len(q))
	v[1] /= float64(len(q))
	if numeric > 0 {
		v[3] /= float64(numeric)
	}
	for i := 0; i+1 < len(order); i++ {
		for j := 0; j+1 < len(d.titleOrder); j++ {
			if order[i] == d.titleOrder[j] && order[i+1] == d.titleOrder[j+1] {
				v[2]++
				break
			}
		}
	}
	if len(order) > 1 {
		v[2] /= float64(len(order) - 1)
	}
	v[4] = math.Min(1, math.Log1p(float64(d.lengths[0]))/math.Log1p(8192))
	v[5] = math.Min(1, math.Log1p(float64(d.lengths[1]))/math.Log1p(8192))
	if nativeRank > 0 {
		v[6] = 1
		v[7] = 60 / float64(60+nativeRank)
	}
	return v, ctx.Err()
}

type Row struct {
	ID       string  `json:"id"`
	Base     float64 `json:"base"`
	Features Vector  `json:"features"`
}
type Case struct {
	ID       string          `json:"id"`
	Family   string          `json:"family"`
	Rows     []Row           `json:"rows"`
	Positive map[string]bool `json:"positive"`
}
type Model struct {
	Weights Vector `json:"weights"`
}
type FitStats struct {
	Cases       int `json:"cases"`
	Families    int `json:"families"`
	Pairs       int `json:"pairs"`
	NoPairCases int `json:"noPairCases"`
	Steps       int `json:"steps"`
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func validate(rows []Row) error {
	if len(rows) == 0 || len(rows) > 200 {
		return errors.New("invalid rank frontier")
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if r.ID == "" || seen[r.ID] || !finite(r.Base) || r.Base < 0 || r.Base > 1 {
			return errors.New("invalid rank row")
		}
		seen[r.ID] = true
		for _, v := range r.Features {
			if !finite(v) || v < 0 || v > 1 {
				return errors.New("invalid feature")
			}
		}
	}
	return nil
}
func (m Model) valid() bool {
	for _, w := range m.Weights {
		if !finite(w) || math.Abs(w) > 4 {
			return false
		}
	}
	return true
}
func (m Model) value(r Row) (float64, Vector) {
	z := 0.
	for j, v := range r.Features {
		z += m.Weights[j] * v
	}
	t := math.Tanh(z)
	d := Vector{}
	for j, v := range r.Features {
		d[j] = .25 * (1 - t*t) * v
	}
	return r.Base + .25*t, d
}
func (m Model) Score(r Row) (float64, error) {
	if !m.valid() {
		return 0, errors.New("invalid model")
	}
	if e := validate([]Row{r}); e != nil {
		return 0, e
	}
	v, _ := m.value(r)
	return v, nil
}

type Scored struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

func (m Model) Rank(ctx context.Context, rows []Row) ([]Scored, error) {
	if ctx == nil || !m.valid() {
		return nil, errors.New("invalid rank context")
	}
	if e := validate(rows); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	out := make([]Scored, len(rows))
	for i, r := range rows {
		v, _ := m.value(r)
		out[i] = Scored{r.ID, v}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].ID < out[j].ID
		}
		return out[i].Score > out[j].Score
	})
	return out, ctx.Err()
}

// PairGradient includes the derivative of the bounded tanh correction. Base
// scores are frozen controls; no gradient or evidence update is applied to them.
func PairGradient(m Model, a, b Row) (float64, Vector, error) {
	if !m.valid() {
		return 0, Vector{}, errors.New("invalid model")
	}
	if e := validate([]Row{a, b}); e != nil {
		return 0, Vector{}, e
	}
	sa, da := m.value(a)
	sb, db := m.value(b)
	margin := sa - sb
	loss := math.Log1p(math.Exp(-margin))
	factor := -1 / (1 + math.Exp(margin))
	g := Vector{}
	for j := range g {
		g[j] = factor * (da[j] - db[j])
	}
	return loss, g, nil
}
func Fit(ctx context.Context, cases []Case) (Model, FitStats, error) {
	m := Model{}
	stats := FitStats{Cases: len(cases), Steps: 200}
	if ctx == nil || len(cases) == 0 || len(cases) > 4096 {
		return m, stats, errors.New("invalid fit cases")
	}
	families := map[string]int{}
	seen := map[string]bool{}
	type pairCase struct {
		c        Case
		pos, neg []int
	}
	groups := make([]pairCase, 0, len(cases))
	for _, c := range cases {
		if e := ctx.Err(); e != nil {
			return m, stats, e
		}
		if c.ID == "" || c.Family == "" || seen[c.ID] {
			return m, stats, errors.New("invalid fit identity")
		}
		seen[c.ID] = true
		if e := validate(c.Rows); e != nil {
			return m, stats, e
		}
		families[c.Family]++
		p := pairCase{c: c}
		for i, r := range c.Rows {
			if c.Positive[r.ID] {
				p.pos = append(p.pos, i)
			} else {
				p.neg = append(p.neg, i)
			}
		}
		if len(p.pos) == 0 || len(p.neg) == 0 {
			stats.NoPairCases++
		}
		stats.Pairs += len(p.pos) * len(p.neg)
		groups = append(groups, p)
	}
	stats.Families = len(families)
	for step := 0; step < 200; step++ {
		g := Vector{}
		for j, w := range m.Weights {
			g[j] = .001 * w
		}
		for _, p := range groups {
			if e := ctx.Err(); e != nil {
				return m, stats, e
			}
			if len(p.pos) == 0 || len(p.neg) == 0 {
				continue
			}
			weight := 1 / float64(len(families)*families[p.c.Family]*len(p.pos)*len(p.neg))
			for _, a := range p.pos {
				sa, da := m.value(p.c.Rows[a])
				for _, b := range p.neg {
					sb, db := m.value(p.c.Rows[b])
					f := -weight / (1 + math.Exp(sa-sb))
					for j := range g {
						g[j] += f * (da[j] - db[j])
					}
				}
			}
		}
		for j := range g {
			m.Weights[j] = math.Max(-4, math.Min(4, m.Weights[j]-.2*g[j]))
		}
	}
	return m, stats, ctx.Err()
}
