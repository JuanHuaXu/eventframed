// Package researchpairedref independently sums unnormalized full histories.
package researchpairedref

import (
	"errors"
	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
	p "github.com/JuanHuaXu/eventframed/internal/researchpaired"
	"math"
)

type trial struct{ known1, known2, y1, y2, requested bool }
type Reference struct {
	cfg               p.Config
	n                 int
	issued            []int
	trials            []trial
	prior, logs, next []float64
}

func New(base []float64, cfg p.Config) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || math.IsNaN(cfg.Hazard) || math.IsInf(cfg.Hazard, 0) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("reference contract")
	}
	r := &Reference{cfg: cfg, n: len(base), issued: make([]int, len(base)), trials: make([]trial, len(base)*64), prior: make([]float64, len(base)*28*21), logs: make([]float64, len(base)*84), next: make([]float64, len(base)*84)}
	for i, b := range base {
		if math.IsNaN(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("reference base")
		}
		for h := 0; h < 28; h++ {
			mu := b
			switch h {
			case 1:
				mu = 1 - b
			case 2:
				mu = .5
			default:
				if h >= 3 {
					j := h - 3
					mu = math.Max(.025, math.Min(.975, .1+.2*float64(j/5)+(-.8+.4*float64(j%5))*float64(i)/float64(len(base)-1)))
				}
			}
			mass, e := prior.Prior(mu, cfg.Strength, "moment")
			if e != nil {
				return nil, e
			}
			copy(r.prior[(i*28+h)*21:], mass[:])
		}
		r.refresh(i)
	}
	return r, nil
}
func componentPrior(c int) float64 {
	eta := []float64{.8, .1, .1}[c/28]
	family := .004
	switch c % 28 {
	case 0:
		family = .72
	case 1, 2:
		family = .09
	}
	return eta * family
}
func (r *Reference) mass(i, c, override int, replacement trial) (float64, float64) {
	mu := r.prior[(i*28+c%28)*21 : (i*28+c%28+1)*21]
	var mass [21]float64
	copy(mass[:], mu)
	eta := []float64{0, .1, .2}[c/28]
	for j := 0; j < r.issued[i]; j++ {
		if j > 0 {
			total := 0.
			for _, v := range mass {
				total += v
			}
			for z := range mass {
				mass[z] = (1-r.cfg.Hazard)*mass[z] + r.cfg.Hazard*mu[z]*total
			}
		}
		slot := i*64 + j
		x := r.trials[slot]
		if slot == override {
			x = replacement
		}
		if x.known1 {
			for z := range mass {
				rate := (float64(z) + .5) / 21
				factor := 0.
				for _, y := range []bool{false, true} {
					prob := rate
					if !y {
						prob = 1 - rate
					}
					a := eta
					if x.y1 == y {
						a = 1 - eta
					}
					prob *= a
					if x.known2 {
						b := eta
						if x.y2 == y {
							b = 1 - eta
						}
						prob *= b
					}
					factor += prob
				}
				mass[z] *= factor
			}
		}
	}
	total, first, initial := 0., 0., 0.
	for z, v := range mass {
		rate := (float64(z) + .5) / 21
		total += v
		first += v * rate
		initial += mu[z] * rate
	}
	if total == 0 {
		return math.Inf(-1), initial
	}
	q := first / total
	if r.issued[i] > 0 {
		q = (1-r.cfg.Hazard)*q + r.cfg.Hazard*initial
	}
	return math.Log(total), q
}
func (r *Reference) refresh(i int) {
	for c := 0; c < 84; c++ {
		r.logs[i*84+c], r.next[i*84+c] = r.mass(i, c, -1, trial{})
	}
}
func (r *Reference) Weights() ([84]float64, error) {
	var w [84]float64
	maximum := math.Inf(-1)
	for c := range w {
		v := math.Log(componentPrior(c))
		for i := 0; i < r.n; i++ {
			v += r.logs[i*84+c]
		}
		w[c] = v
		maximum = math.Max(maximum, v)
	}
	if math.IsNaN(maximum) || math.IsInf(maximum, 0) {
		return w, errors.New("reference global support")
	}
	sum := 0.
	for c := range w {
		w[c] = math.Exp(w[c] - maximum)
		sum += w[c]
	}
	for c := range w {
		w[c] /= sum
	}
	return w, nil
}
func (r *Reference) NoiseWeights() ([3]float64, error) {
	w, e := r.Weights()
	var out [3]float64
	for c, v := range w {
		out[c/28] += v
	}
	return out, e
}
func (r *Reference) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= r.n {
		return 0, 0, errors.New("reference member")
	}
	w, e := r.Weights()
	if e != nil {
		return 0, 0, e
	}
	q, o := 0., 0.
	for c, v := range w {
		eta := []float64{0, .1, .2}[c/28]
		q += v * r.next[i*84+c]
		o += v * (eta + (1-2*eta)*r.next[i*84+c])
	}
	return q, o, nil
}
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= r.n || r.issued[i] >= 64 {
		return errors.New("reference issue")
	}
	r.issued[i]++
	r.refresh(i)
	return nil
}
func (r *Reference) Observe(i, ordinal, measurement int, y bool) error {
	if i < 0 || i >= r.n || ordinal < 1 || ordinal > r.issued[i] {
		return errors.New("reference identity")
	}
	k := i*64 + ordinal - 1
	x := r.trials[k]
	if measurement == 1 {
		if x.known1 {
			return errors.New("reference first replay")
		}
		x.known1, x.y1 = true, y
	} else if measurement == 2 {
		if !x.known1 || !x.requested || x.known2 {
			return errors.New("reference second state")
		}
		x.known2, x.y2 = true, y
	} else {
		return errors.New("reference measurement")
	}
	r.trials[k] = x
	r.refresh(i)
	return nil
}
func (r *Reference) Audit(i, ordinal int) error {
	if i < 0 || i >= r.n || ordinal < 1 || ordinal > r.issued[i] {
		return errors.New("reference audit identity")
	}
	k := i*64 + ordinal - 1
	if !r.trials[k].known1 || r.trials[k].requested {
		return errors.New("reference audit state")
	}
	r.trials[k].requested = true
	return nil
}
func entropy(q float64) float64 {
	if q == 0 || q == 1 {
		return 0
	}
	return -q*math.Log(q) - (1-q)*math.Log(1-q)
}
func (r *Reference) Options(i, ordinal int) (p.Option, error) {
	if i < 0 || i >= r.n || ordinal < 1 || ordinal > r.issued[i] {
		return p.Option{}, errors.New("reference options identity")
	}
	k := i*64 + ordinal - 1
	x := r.trials[k]
	if !x.known1 || x.requested {
		return p.Option{}, errors.New("reference options state")
	}
	weights, e := r.Weights()
	if e != nil {
		return p.Option{}, e
	}
	q, h := 0., 0.
	var probabilities [84]float64
	x.known2 = true
	for c, w := range weights {
		if w == 0 {
			continue
		}
		x.y2 = true
		yes, _ := r.mass(i, c, k, x)
		x.y2 = false
		no, _ := r.mass(i, c, k, x)
		py, pn := math.Exp(yes-r.logs[i*84+c]), math.Exp(no-r.logs[i*84+c])
		if math.Abs(py+pn-1) > 2e-10 {
			return p.Option{}, errors.New("reference options normalization")
		}
		py = math.Min(1, py)
		probabilities[c] = py
		q += w * py
		h += w * entropy(py)
	}
	q = math.Min(1, q)
	edge := 0.
	for c, w := range weights {
		v := probabilities[c]
		term := -1.
		if q > 0 {
			term += v * v / q
		}
		if q < 1 {
			term += (1 - v) * (1 - v) / (1 - q)
		}
		edge += w * w * term
	}
	return p.Option{Observed: q, Uncertainty: entropy(q), Information: math.Max(0, entropy(q)-h), EdgeCut: math.Max(0, edge)}, nil
}
