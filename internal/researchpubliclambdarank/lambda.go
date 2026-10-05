// Package researchpubliclambdarank tests a top-ten-aware ranking rescue.
// Dynamic swap weights are stop-gradient lambdas, not a convex-loss theorem.
package researchpubliclambdarank

import (
	"context"
	"errors"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"sort"
)

func Discount(rank int) float64 {
	if rank < 1 || rank > 10 {
		return 0
	}
	return 1 / math.Log2(float64(rank+1))
}
func value(m r.Model, row r.Row) (float64, r.Vector) {
	z := 0.
	for j, v := range row.Features {
		z += m.Weights[j] * v
	}
	t := math.Tanh(z)
	d := r.Vector{}
	for j, v := range row.Features {
		d[j] = .25 * (1 - t*t) * v
	}
	return row.Base + .25*t, d
}
func Fit(ctx context.Context, cases []r.Case) (r.Model, r.FitStats, error) {
	m := r.Model{}
	stats := r.FitStats{Cases: len(cases), Steps: 200}
	if ctx == nil || len(cases) == 0 || len(cases) > 4096 {
		return m, stats, errors.New("invalid lambda cases")
	}
	families := map[string]int{}
	seen := map[string]bool{}
	type group struct {
		c        r.Case
		pos, neg []int
	}
	groups := make([]group, 0, len(cases))
	for _, c := range cases {
		if c.ID == "" || c.Family == "" || seen[c.ID] {
			return m, stats, errors.New("invalid lambda identity")
		}
		seen[c.ID] = true
		if _, e := m.Rank(ctx, c.Rows); e != nil {
			return m, stats, e
		}
		families[c.Family]++
		p := group{c: c}
		for i, v := range c.Rows {
			if c.Positive[v.ID] {
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
		g := r.Vector{}
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
			n := len(p.c.Rows)
			s := make([]float64, n)
			d := make([]r.Vector, n)
			order := make([]int, n)
			for i, v := range p.c.Rows {
				s[i], d[i] = value(m, v)
				order[i] = i
			}
			sort.Slice(order, func(i, j int) bool {
				a, b := order[i], order[j]
				if s[a] == s[b] {
					return p.c.Rows[a].ID < p.c.Rows[b].ID
				}
				return s[a] > s[b]
			})
			discount := make([]float64, n)
			for i, j := range order {
				discount[j] = Discount(i + 1)
			}
			sum := 0.
			for _, a := range p.pos {
				for _, b := range p.neg {
					sum += math.Abs(discount[a] - discount[b])
				}
			}
			if sum == 0 {
				continue
			}
			norm := 1 / (float64(len(families)*families[p.c.Family]) * sum)
			for _, a := range p.pos {
				for _, b := range p.neg {
					f := -norm * math.Abs(discount[a]-discount[b]) / (1 + math.Exp(s[a]-s[b]))
					for j := range g {
						g[j] += f * (d[a][j] - d[b][j])
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
