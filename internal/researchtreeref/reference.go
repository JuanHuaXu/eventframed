// Package researchtreeref independently reconstructs retained evidence from an
// issued-position ledger. It never calls the candidate's inference routines.
package researchtreeref

import (
	"errors"
	"math"
	"sort"

	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
	tree "github.com/JuanHuaXu/eventframed/internal/researchtree"
)

type event struct {
	member, ordinal              int
	first, second, audited, a, b bool
}
type summary struct{ evidence, mean, stopping [3]float64 }
type Reference struct {
	base          []float64
	cfg           tree.Config
	ranks, issued []int
	records       []event
	priors        [255][21]float64
	likelihood    [3][21][6]float64
	nodes         [255]summary
	weights       [3]float64
}

func New(base []float64, cfg tree.Config) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || cfg.Depth < 0 || cfg.Depth > 7 || cfg.Window < 1 || cfg.Window > 64*len(base) || cfg.Strength <= 0 || cfg.Strength > 32 || math.IsNaN(cfg.Strength) {
		return nil, errors.New("reference contract")
	}
	r := &Reference{base: append([]float64(nil), base...), cfg: cfg, ranks: make([]int, len(base)), issued: make([]int, len(base))}
	ordered := append([]float64(nil), base...)
	sort.Float64s(ordered)
	for i, b := range base {
		if b < .25 || b > .925+1e-12 || math.IsNaN(b) {
			return nil, errors.New("reference feature")
		}
		r.ranks[i] = sort.SearchFloat64s(ordered, b) * (1 << cfg.Depth) / len(base)
	}
	for n := 0; n < (1<<(cfg.Depth+1))-1; n++ {
		values := []float64{}
		for i, b := range base {
			for p := (1 << cfg.Depth) - 1 + r.ranks[i]; ; p = (p - 1) / 2 {
				if p == n {
					values = append(values, b)
					break
				}
				if p == 0 {
					break
				}
			}
		}
		if len(values) == 0 {
			values = append(values, ordered...)
		}
		sort.Float64s(values)
		mu := 0.
		for _, b := range values {
			mu += b / float64(len(values))
		}
		p, e := prior.Prior(mu, cfg.Strength, "moment")
		if e != nil {
			return nil, e
		}
		r.priors[n] = p
	}
	// Explicit latent-Y sum: each observer is conditionally independent given
	// ONE original outcome, not another independent outcome-rate draw.
	for h, eta := range []float64{0, .1, .2} {
		for z := 0; z < 21; z++ {
			p := (float64(z) + .5) / 21
			for k := 0; k < 6; k++ {
				sum := 0.
				for y := 0; y < 2; y++ {
					mass := 1 - p
					if y == 1 {
						mass = p
					}
					a := k
					if k >= 2 {
						a = (k - 2) / 2
					}
					if a == y {
						mass *= 1 - eta
					} else {
						mass *= eta
					}
					if k >= 2 {
						b := (k - 2) % 2
						if b == y {
							mass *= 1 - eta
						} else {
							mass *= eta
						}
					}
					sum += mass
				}
				r.likelihood[h][z][k] = math.Log(sum)
			}
		}
	}
	for n := (1 << (cfg.Depth + 1)) - 2; n >= 0; n-- {
		r.rebuild(&r.nodes, n, -1, event{})
	}
	r.weights = r.normalize(&r.nodes)
	return r, nil
}
func cat(e event) int {
	if !e.first {
		return -1
	}
	a, b := 0, 0
	if e.a {
		a = 1
	}
	if e.b {
		b = 1
	}
	if e.second {
		return 2 + 2*a + b
	}
	return a
}
func lse(v []float64) float64 {
	maximum := math.Inf(-1)
	for _, x := range v {
		maximum = math.Max(maximum, x)
	}
	if math.IsInf(maximum, -1) {
		return maximum
	}
	sum := 0.
	for _, x := range v {
		sum += math.Exp(x - maximum)
	}
	return maximum + math.Log(sum)
}
func (r *Reference) rebuild(nodes *[255]summary, n, override int, extra event) {
	counts := [6]int{}
	for k := max(0, len(r.records)-r.cfg.Window); k < len(r.records); k++ {
		e := r.records[k]
		if k == override {
			e = extra
		}
		c := cat(e)
		if c < 0 {
			continue
		}
		for p := (1 << r.cfg.Depth) - 1 + r.ranks[e.member]; ; p = (p - 1) / 2 {
			if p == n {
				counts[c]++
				break
			}
			if p == 0 {
				break
			}
		}
	}
	for h := 0; h < 3; h++ {
		v := make([]float64, 21)
		for z := range v {
			v[z] = math.Log(r.priors[n][z])
			for k, c := range counts {
				if c > 0 {
					v[z] += float64(c) * r.likelihood[h][z][k]
				}
			}
		}
		leaf := lse(v)
		mean := 0.
		if math.IsInf(leaf, -1) {
			for z, p := range r.priors[n] {
				mean += p * (float64(z) + .5) / 21
			}
		} else {
			for z, x := range v {
				mean += math.Exp(x-leaf) * (float64(z) + .5) / 21
			}
		}
		out := &nodes[n]
		out.mean[h] = mean
		out.evidence[h] = leaf
		out.stopping[h] = 1
		if n < (1<<r.cfg.Depth)-1 {
			out.evidence[h] = lse([]float64{math.Log(.5) + leaf, math.Log(.5) + nodes[2*n+1].evidence[h] + nodes[2*n+2].evidence[h]})
			if !math.IsInf(out.evidence[h], -1) {
				out.stopping[h] = math.Exp(math.Log(.5) + leaf - out.evidence[h])
			}
		}
	}
}
func (r *Reference) normalize(nodes *[255]summary) [3]float64 {
	p := [3]float64{.8, .1, .1}
	logs := []float64{}
	for h, v := range p {
		logs = append(logs, math.Log(v)+nodes[0].evidence[h])
	}
	total := lse(logs)
	for h := range p {
		p[h] = math.Exp(logs[h] - total)
	}
	return p
}
func (r *Reference) path(member int, nodes *[255]summary, override int, extra event) {
	for n := (1 << r.cfg.Depth) - 1 + r.ranks[member]; ; n = (n - 1) / 2 {
		r.rebuild(nodes, n, override, extra)
		if n == 0 {
			break
		}
	}
}
func (r *Reference) forecast(member int, nodes *[255]summary, w [3]float64) (float64, float64) {
	clean, observed := 0., 0.
	leaf := (1 << r.cfg.Depth) - 1 + r.ranks[member]
	path := []int{}
	for n := leaf; ; n = (n - 1) / 2 {
		path = append(path, n)
		if n == 0 {
			break
		}
	}
	for h, eta := range []float64{0, .1, .2} {
		if w[h] == 0 {
			continue
		}
		conditional, survival := 0., 1.
		for k := len(path) - 1; k >= 0; k-- {
			n := path[k]
			s := nodes[n].stopping[h]
			conditional += survival * s * nodes[n].mean[h]
			survival *= 1 - s
		}
		clean += w[h] * conditional
		observed += w[h] * (eta + (1-2*eta)*conditional)
	}
	return clean, observed
}
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
	for n := (1 << (r.cfg.Depth + 1)) - 2; n >= 0; n-- {
		r.rebuild(&r.nodes, n, -1, event{})
	}
	for n := 0; n < (1<<(r.cfg.Depth+1))-1; n++ {
		for h := 0; h < 3; h++ {
			a, b := saved[n], r.nodes[n]
			for _, p := range [][2]float64{{a.evidence[h], b.evidence[h]}, {a.mean[h], b.mean[h]}, {a.stopping[h], b.stopping[h]}} {
				if p[0] != p[1] && math.Abs(p[0]-p[1]) > 1e-12 {
					return errors.New("reference incremental reconstruction")
				}
			}
		}
	}
	r.weights = r.normalize(&r.nodes)
	return nil
}
