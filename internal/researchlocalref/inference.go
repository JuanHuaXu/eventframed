// Package researchlocalref independently rebuilds conditional likelihoods
// from the retained ledger. No candidate inference or cached sums are reused.
package researchlocalref

import (
	"errors"
	tree "github.com/JuanHuaXu/eventframed/internal/researchlocal"
	"math"
	"sort"
)

type event struct {
	member, ordinal              int
	first, second, audited, a, b bool
}
type summary struct {
	evidence, stopping [3]float64
	posterior          [3][21]float64
	local              map[int][3]float64 // Immutable after reconstruction.
}
type memberSummary struct {
	evidence, mean [3]float64
	posterior      [3][22]float64
}
type Reference struct {
	base                    []float64
	cfg                     tree.Config
	ranks, issued           []int
	records                 []event
	members                 [255][]int
	rates                   [][21]float64
	likelihood              [][3][21][6]float64
	contributions           [][3][21]float64
	individuals             []memberSummary
	localRates, localPriors [][22]float64
	localLikelihood         [][3][22][6]float64
	nodes                   [255]summary
	weights                 [3]float64
}

func atomPrior(z int) float64 {
	if z == 0 {
		return .8
	}
	return .01
}
func distance(z int) float64 {
	if z < 11 {
		return -.6 * float64(z)
	}
	return .6 * float64(z-10)
}
func logistic(x float64) float64 { return math.Exp(-math.Log1p(math.Exp(-x))) }
func solve(b float64) float64 {
	lo, hi, x := -40., 40., math.Log(b/(1-b))
	for k := 0; k < 160; k++ {
		mean, derivative := 0., 0.
		for z := 1; z <= 20; z++ {
			q := logistic(x + distance(z))
			mean += q / 20
			derivative += q * (1 - q) / 20
		}
		if math.Abs(mean-b) < 1e-15 {
			return x
		}
		if mean < b {
			lo = x
		} else {
			hi = x
		}
		next := x + (b-mean)/derivative
		if math.IsNaN(next) || next <= lo || next >= hi {
			next = (lo + hi) / 2
		}
		x = next
	}
	return x
}
func New(base []float64, cfg tree.Config) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || cfg.Depth < 0 || cfg.Depth > 7 || cfg.Window < 1 || cfg.Window > 64*len(base) {
		return nil, errors.New("anchor reference contract")
	}
	r := &Reference{base: append([]float64(nil), base...), cfg: cfg, ranks: make([]int, len(base)), issued: make([]int, len(base)), rates: make([][21]float64, len(base)), likelihood: make([][3][21][6]float64, len(base)), contributions: make([][3][21]float64, len(base))}
	r.individuals = make([]memberSummary, len(base))
	r.localRates, r.localPriors = make([][22]float64, len(base)), make([][22]float64, len(base))
	r.localLikelihood = make([][3][22][6]float64, len(base))
	ordered := append([]float64(nil), base...)
	sort.Float64s(ordered)
	for i, b := range base {
		if b < .25 || b > .925+1e-12 || math.IsNaN(b) {
			return nil, errors.New("reference feature")
		}
		r.ranks[i] = sort.SearchFloat64s(ordered, b) * (1 << cfg.Depth) / len(base)
		for n := (1 << cfg.Depth) - 1 + r.ranks[i]; ; n = (n - 1) / 2 {
			r.members[n] = append(r.members[n], i)
			if n == 0 {
				break
			}
		}
		x := solve(b)
		// Independent prior construction: dynamic distribution of a Polya urn,
		// not the candidate's adjacent-mass recurrence.
		weights := [21]float64{1}
		for count := 0; count < 20; count++ {
			var next [21]float64
			for z := 0; z <= count; z++ {
				next[z+1] += weights[z] * (b + float64(z)) / (1 + float64(count))
				next[z] += weights[z] * (1 - b + float64(count-z)) / (1 + float64(count))
			}
			weights = next
		}
		r.localRates[i][0], r.localPriors[i][0] = b, .8
		for z, p := range weights {
			r.localRates[i][z+1] = float64(z) / 20
			r.localPriors[i][z+1] = .2 * p
		}
		r.rates[i][0] = b
		mean := .8 * b
		for z := 1; z <= 20; z++ {
			r.rates[i][z] = logistic(x + distance(z))
			mean += .01 * r.rates[i][z]
		}
		if math.Abs(mean-b) > 1e-13 {
			return nil, errors.New("reference baseline solve")
		}
		// Different arithmetic: enumerate the hidden ONE Y for both observers.
		for h, eta := range []float64{0, .1, .2} {
			for z, p := range r.localRates[i] {
				for k := 0; k < 6; k++ {
					r.localLikelihood[i][h][z][k] = localLogFactor(p, eta, k)
				}
			}
			for z, p := range r.rates[i] {
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
					r.likelihood[i][h][z][k] = math.Log(sum)
				}
			}
		}
		r.individuals[i] = r.individual(i, -1, event{})
	}
	for n := (1 << (cfg.Depth + 1)) - 2; n >= 0; n-- {
		r.rebuild(&r.nodes, n, -1, [3][21]float64{}, memberSummary{})
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

// Recount the retained ledger for the affected member, not subtract/add cached
// evidence. Other members' independent reconstructions stay unchanged.
func (r *Reference) contribution(member, override int, extra event) [3][21]float64 {
	counts := [6]int{}
	for k := max(0, len(r.records)-r.cfg.Window); k < len(r.records); k++ {
		e := r.records[k]
		if k == override {
			e = extra
		}
		if e.member != member {
			continue
		}
		c := cat(e)
		if c >= 0 {
			counts[c]++
		}
	}
	var out [3][21]float64
	for h := 0; h < 3; h++ {
		for z := 0; z < 21; z++ {
			for k, c := range counts {
				if c > 0 {
					out[h][z] += float64(c) * r.likelihood[member][h][z][k]
				}
			}
		}
	}
	return out
}
func (r *Reference) rebuild(nodes *[255]summary, n, member int, extra [3][21]float64, local memberSummary) {
	if n >= (1<<r.cfg.Depth)-1 {
		nodes[n].local = make(map[int][3]float64, len(r.members[n]))
	}
	for h := 0; h < 3; h++ {
		terms := make([]float64, 21)
		for z := range terms {
			terms[z] = math.Log(atomPrior(z))
			for _, i := range r.members[n] {
				if i == member {
					terms[z] += extra[h][z]
				} else {
					terms[z] += r.contributions[i][h][z]
				}
			}
		}
		leaf := lse(terms)
		x := &nodes[n]
		for z := range terms {
			x.posterior[h][z] = atomPrior(z)
			if !math.IsInf(leaf, -1) {
				x.posterior[h][z] = math.Exp(terms[z] - leaf)
			}
		}
		split := 0.
		if n < (1<<r.cfg.Depth)-1 {
			split = nodes[2*n+1].evidence[h] + nodes[2*n+2].evidence[h]
		} else {
			for _, i := range r.members[n] {
				s := r.individuals[i]
				if i == member {
					s = local
				}
				split += s.evidence[h]
				x.local[i] = s.mean
			}
		}
		x.evidence[h], x.stopping[h] = lse([]float64{math.Log(.5) + leaf, math.Log(.5) + split}), 1
		if !math.IsInf(x.evidence[h], -1) {
			x.stopping[h] = math.Exp(math.Log(.5) + leaf - x.evidence[h])
		}
	}
}
func (r *Reference) normalize(nodes *[255]summary) [3]float64 {
	w := [3]float64{.8, .1, .1}
	logs := make([]float64, 3)
	for h, p := range w {
		logs[h] = math.Log(p) + nodes[0].evidence[h]
	}
	total := lse(logs)
	for h := range w {
		w[h] = math.Exp(logs[h] - total)
	}
	return w
}
func (r *Reference) path(member int, nodes *[255]summary, override int, extra event) {
	v := r.contribution(member, override, extra)
	local := r.individual(member, override, extra)
	if override < 0 {
		r.contributions[member] = v
		r.individuals[member] = local
	}
	for n := (1 << r.cfg.Depth) - 1 + r.ranks[member]; ; n = (n - 1) / 2 {
		r.rebuild(nodes, n, member, v, local)
		if n == 0 {
			break
		}
	}
}
func (r *Reference) forecast(member int, nodes *[255]summary, w [3]float64) (float64, float64) {
	path := []int{}
	for n := (1 << r.cfg.Depth) - 1 + r.ranks[member]; ; n = (n - 1) / 2 {
		path = append(path, n)
		if n == 0 {
			break
		}
	}
	clean, observed := 0., 0.
	for h, eta := range []float64{0, .1, .2} {
		if w[h] == 0 {
			continue
		}
		conditional, survival := 0., 1.
		for k := len(path) - 1; k >= 0; k-- {
			n := path[k]
			mean := 0.
			for z, weight := range nodes[n].posterior[h] {
				mean += weight * r.rates[member][z]
			}
			s := nodes[n].stopping[h]
			conditional += survival * s * mean
			survival *= 1 - s
		}
		conditional += survival * nodes[path[0]].local[member][h]
		clean += w[h] * conditional
		observed += w[h] * (eta + (1-2*eta)*conditional)
	}
	return clean, observed
}

func localLogFactor(p, eta float64, k int) float64 {
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
			if (k-2)%2 == y {
				mass *= 1 - eta
			} else {
				mass *= eta
			}
		}
		sum += mass
	}
	return math.Log(sum)
}
func (r *Reference) individual(member, override int, extra event) memberSummary {
	var out memberSummary
	counts := [6]int{}
	for k := max(0, len(r.records)-r.cfg.Window); k < len(r.records); k++ {
		e := r.records[k]
		if k == override {
			e = extra
		}
		if e.member == member && cat(e) >= 0 {
			counts[cat(e)]++
		}
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	for h := 0; h < 3; h++ {
		terms := make([]float64, 22)
		for z := range terms {
			terms[z] = math.Log(r.localPriors[member][z])
			for k, c := range counts {
				if c > 0 {
					terms[z] += float64(c) * r.localLikelihood[member][h][z][k]
				}
			}
		}
		out.evidence[h] = lse(terms)
		for z := range terms {
			out.posterior[h][z] = r.localPriors[member][z]
			if !math.IsInf(out.evidence[h], -1) {
				out.posterior[h][z] = math.Exp(terms[z] - out.evidence[h])
			}
			out.mean[h] += out.posterior[h][z] * r.localRates[member][z]
		}
		if total == 0 {
			out.evidence[h] = 0
		}
	}
	return out
}
