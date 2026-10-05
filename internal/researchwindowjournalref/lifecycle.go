// Ledger lifecycle retained mechanically; inference rebuilt independently.
package researchwindowjournalref

import (
	"errors"
	tree "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"
	"math"
)

func (r *Reference) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(r.base) {
		return 0, 0, errors.New("reference member")
	}
	q, o := r.forecast(i, &r.nodes, r.weights)
	return q, o, nil
}
func (r *Reference) NoiseWeights() ([3]float64, error) { return r.weights, nil }
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= len(r.base) || r.issued[i] >= 64 {
		return errors.New("reference issue")
	}
	r.issued[i]++
	r.records = append(r.records, event{member: i, ordinal: r.issued[i]})
	if len(r.records) > r.cfg.Window {
		old := r.records[len(r.records)-r.cfg.Window-1]
		if cat(old) >= 0 {
			r.path(old.member, &r.nodes, -1, event{})
			r.weights = r.normalize(&r.nodes)
		}
	}
	return nil
}
func (r *Reference) position(i, ordinal int) int {
	for k, e := range r.records {
		if e.member == i && e.ordinal == ordinal {
			return k
		}
	}
	return -1
}
func (r *Reference) Observe(i, ordinal, measurement int, y bool) error {
	k := r.position(i, ordinal)
	if k < 0 || (measurement != 1 && measurement != 2) {
		return errors.New("reference observe")
	}
	x := &r.records[k]
	if measurement == 1 {
		if x.first {
			return errors.New("reference replay")
		}
		x.first, x.a = true, y
	} else {
		if !x.first || !x.audited || x.second {
			return errors.New("reference second")
		}
		x.second, x.b = true, y
	}
	if k >= len(r.records)-r.cfg.Window {
		r.path(i, &r.nodes, -1, event{})
		r.weights = r.normalize(&r.nodes)
	}
	return nil
}
func (r *Reference) Audit(i, ordinal int) error {
	k := r.position(i, ordinal)
	if k < max(0, len(r.records)-r.cfg.Window) || !r.records[k].first || r.records[k].audited {
		return errors.New("reference audit")
	}
	r.records[k].audited = true
	return nil
}
func hbinary(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log(1-p)
}
func (r *Reference) hypothetical(i, ordinal int) ([255]summary, [255]summary, float64, [3]float64, error) {
	yes, no := r.nodes, r.nodes
	probabilities := [3]float64{}
	k := r.position(i, ordinal)
	if k < max(0, len(r.records)-r.cfg.Window) || !r.records[k].first || r.records[k].audited {
		return yes, no, 0, probabilities, errors.New("reference option")
	}
	x := r.records[k]
	x.second, x.b = true, true
	r.path(i, &yes, k, x)
	x.b = false
	r.path(i, &no, k, x)
	q := 0.
	for h, w := range r.weights {
		if w == 0 {
			continue
		}
		p := math.Exp(yes[0].evidence[h] - r.nodes[0].evidence[h])
		pn := math.Exp(no[0].evidence[h] - r.nodes[0].evidence[h])
		if math.Abs(p+pn-1) > 1e-9 {
			return yes, no, 0, probabilities, errors.New("reference normalization")
		}
		probabilities[h] = math.Min(1, p)
		q += w * probabilities[h]
	}
	return yes, no, math.Min(1, q), probabilities, nil
}
func (r *Reference) Options(i, ordinal int) (tree.Option, error) {
	_, _, q, py, e := r.hypothetical(i, ordinal)
	if e != nil {
		return tree.Option{}, e
	}
	o := tree.Option{Observed: q, Uncertainty: hbinary(q), Information: hbinary(q)}
	for h, w := range r.weights {
		o.Information -= w * hbinary(py[h])
		one, two := 0., 0.
		if q > 0 {
			one = py[h] * py[h] / q
		}
		if q < 1 {
			two = (1 - py[h]) * (1 - py[h]) / (1 - q)
		}
		o.EdgeCut += w * w * (one + two - 1)
	}
	o.Information = math.Max(0, o.Information)
	o.EdgeCut = math.Max(0, o.EdgeCut)
	return o, nil
}
func (r *Reference) PredictionValue(i, ordinal int, targets []float64) (float64, float64, error) {
	yes, no, q, _, e := r.hypothetical(i, ordinal)
	if e != nil {
		return 0, 0, e
	}
	wy, wn := r.normalize(&yes), r.normalize(&no)
	value := 0.
	for j, w := range targets {
		before, _ := r.forecast(j, &r.nodes, r.weights)
		a, b := before, before
		if q > 0 {
			a, _ = r.forecast(j, &yes, wy)
		}
		if q < 1 {
			b, _ = r.forecast(j, &no, wn)
		}
		if math.Abs(q*a+(1-q)*b-before) > 1e-9 {
			return 0, 0, errors.New("reference tower")
		}
		value += w * (q*(a-before)*(a-before) + (1-q)*(b-before)*(b-before))
	}
	return q, value, nil
}

// RebuildAll is an independent full-suffix reconstruction fence for the
// incremental reference itself. Small-depth tests additionally enumerate trees.
func (r *Reference) RebuildAll() error {
	saved := r.nodes
	savedIndividuals := append([]memberSummary(nil), r.individuals...)
	for i := range r.base {
		r.contributions[i] = r.contribution(i, -1, event{})
		r.individuals[i] = r.individual(i, -1, event{})
		for h := 0; h < 3; h++ {
			for z := 0; z < 22; z++ {
				if math.Abs(savedIndividuals[i].posterior[h][z]-r.individuals[i].posterior[h][z]) > 1e-12 {
					return errors.New("reference individual reconstruction")
				}
			}
			for _, p := range [][2]float64{{savedIndividuals[i].mean[h], r.individuals[i].mean[h]}, {savedIndividuals[i].evidence[h], r.individuals[i].evidence[h]}} {
				if p[0] != p[1] && math.Abs(p[0]-p[1]) > 1e-12 {
					return errors.New("reference individual evidence reconstruction")
				}
			}
		}
	}
	for n := (1 << (r.cfg.Depth + 1)) - 2; n >= 0; n-- {
		r.rebuild(&r.nodes, n, -1, [3][21]float64{}, memberSummary{})
	}
	for n := 0; n < (1<<(r.cfg.Depth+1))-1; n++ {
		for h := 0; h < 3; h++ {
			a, b := saved[n], r.nodes[n]
			for z := 0; z < 21; z++ {
				if math.Abs(a.posterior[h][z]-b.posterior[h][z]) > 1e-12 {
					return errors.New("reference conditional reconstruction")
				}
			}
			for _, p := range [][2]float64{{a.evidence[h], b.evidence[h]}, {a.stopping[h], b.stopping[h]}} {
				if p[0] != p[1] && math.Abs(p[0]-p[1]) > 1e-12 {
					return errors.New("reference incremental reconstruction")
				}
			}
			for member, v := range a.local {
				if math.Abs(v[h]-b.local[member][h]) > 1e-12 {
					return errors.New("reference terminal reconstruction")
				}
			}
		}
	}
	r.weights = r.normalize(&r.nodes)
	return nil
}
