// Package researchmomentref independently integrates complete as-of histories.
// It does not call the candidate's prior solver, predictor or replay code.
package researchmomentref

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

type Reference struct {
	cfg        researchmoment.Config
	h          int
	prior      []float64
	issued     []int
	known, yes []uint64
	logs, next []float64
	weights    [28]float64
}

// Prior uses safeguarded Newton steps with the rate variance derivative,
// independently of the candidate's fixed-iteration bisection.
func Prior(mu, s float64, mode string) ([21]float64, error) {
	var out, exponent [21]float64
	if math.IsNaN(mu) || math.IsNaN(s) || math.IsInf(s, 0) || mu <= .5/21 || mu >= 20.5/21 || s <= 0 || s > 32 || (mode != "density" && mode != "moment") {
		return out, errors.New("reference prior contract")
	}
	for z := range exponent {
		a := (float64(z) + .5) / 21
		exponent[z] = (s*mu-1)*math.Log(a) + (s*(1-mu)-1)*math.Log(1-a)
	}
	evaluate := func(lambda float64) (float64, float64) {
		max := math.Inf(-1)
		for z, v := range exponent {
			max = math.Max(max, v+lambda*(float64(z)+.5)/21)
		}
		sum, mean, second := 0., 0., 0.
		for z, v := range exponent {
			out[z] = math.Exp(v + lambda*(float64(z)+.5)/21 - max)
			sum += out[z]
		}
		for z := range out {
			out[z] /= sum
			a := (float64(z) + .5) / 21
			mean += out[z] * a
			second += out[z] * a * a
		}
		return mean, second - mean*mean
	}
	if mode == "density" {
		evaluate(0)
		return out, nil
	}
	lo, hi, lambda := -4096., 4096., 0.
	l, _ := evaluate(lo)
	u, _ := evaluate(hi)
	if !(l < mu && u > mu) {
		return out, errors.New("reference mean bracket")
	}
	for k := 0; k < 160; k++ {
		mean, variance := evaluate(lambda)
		if math.Abs(mean-mu) < 2e-14 {
			return out, nil
		}
		if mean < mu {
			lo = lambda
		} else {
			hi = lambda
		}
		next := lambda + (mu-mean)/variance
		if variance <= 0 || math.IsNaN(next) || next <= lo || next >= hi {
			next = (lo + hi) / 2
		}
		lambda = next
	}
	return out, errors.New("reference mean nonconvergence")
}
func New(base []float64, cfg researchmoment.Config) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || (cfg.Family != "narrow" && cfg.Family != "rich") || math.IsNaN(cfg.Hazard) || math.IsInf(cfg.Hazard, 0) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("reference contract")
	}
	h := 3
	if cfg.Family == "rich" {
		h = 28
	}
	r := &Reference{cfg: cfg, h: h, prior: make([]float64, len(base)*h*21), issued: make([]int, len(base)), known: make([]uint64, len(base)), yes: make([]uint64, len(base)), logs: make([]float64, len(base)*h), next: make([]float64, len(base)*h), weights: [28]float64{.8, .1, .1}}
	if h == 28 {
		r.weights[0], r.weights[1], r.weights[2] = .72, .09, .09
		for j := 3; j < h; j++ {
			r.weights[j] = .1 / 25
		}
	}
	for i, b := range base {
		if math.IsNaN(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("reference baseline")
		}
		for j := 0; j < h; j++ {
			mu := b
			if j == 1 {
				mu = 1 - b
			} else if j == 2 {
				mu = .5
			} else if j >= 3 {
				mu = math.Max(.025, math.Min(.975, .1+.2*float64((j-3)/5)+(-.8+.4*float64((j-3)%5))*float64(i)/float64(len(base)-1)))
			}
			p, e := Prior(mu, cfg.Strength, cfg.Prior)
			if e != nil {
				return nil, e
			}
			copy(r.prior[(i*h+j)*21:], p[:])
		}
		r.refresh(i)
	}
	return r, nil
}

// Unnormalized full-history integration: pending/censored outcomes have unit
// likelihoods. Global family evidence is summed afresh, not incrementally
// subtracted from the candidate's cache. The bounded 64-step masses stay finite.
func (r *Reference) refresh(i int) {
	for h := 0; h < r.h; h++ {
		var mass [21]float64
		p := r.prior[(i*r.h+h)*21 : (i*r.h+h+1)*21]
		copy(mass[:], p)
		for j := 0; j < r.issued[i]; j++ {
			if j > 0 {
				total := 0.
				for _, v := range mass {
					total += v
				}
				for z := range mass {
					mass[z] = (1-r.cfg.Hazard)*mass[z] + r.cfg.Hazard*p[z]*total
				}
			}
			if r.known[i]&(1<<j) != 0 {
				for z := range mass {
					a := (float64(z) + .5) / 21
					if r.yes[i]&(1<<j) == 0 {
						a = 1 - a
					}
					mass[z] *= a
				}
			}
		}
		total, first, initial := 0., 0., 0.
		for z, v := range mass {
			a := (float64(z) + .5) / 21
			total += v
			first += v * a
			initial += p[z] * a
		}
		r.logs[i*r.h+h] = math.Log(total)
		r.next[i*r.h+h] = first / total
		if r.issued[i] > 0 {
			r.next[i*r.h+h] = (1-r.cfg.Hazard)*r.next[i*r.h+h] + r.cfg.Hazard*initial
		}
	}
}
func (r *Reference) Predict(i int) (float64, error) {
	if i < 0 || i >= len(r.issued) {
		return 0, errors.New("reference member")
	}
	var logs, weights [28]float64
	if r.cfg.Shared {
		for k := range r.issued {
			for h := 0; h < r.h; h++ {
				logs[h] += r.logs[k*r.h+h]
			}
		}
	} else {
		copy(logs[:r.h], r.logs[i*r.h:(i+1)*r.h])
	}
	maximum := math.Inf(-1)
	for h := 0; h < r.h; h++ {
		weights[h] = math.Log(r.weights[h]) + logs[h]
		maximum = math.Max(maximum, weights[h])
	}
	sum, q := 0., 0.
	for h := 0; h < r.h; h++ {
		weights[h] = math.Exp(weights[h] - maximum)
		sum += weights[h]
	}
	for h := 0; h < r.h; h++ {
		q += weights[h] / sum * r.next[i*r.h+h]
	}
	if math.IsNaN(q) || q <= 0 || q >= 1 {
		return 0, errors.New("reference probability")
	}
	return q, nil
}
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= len(r.issued) || r.issued[i] >= 64 {
		return errors.New("reference issue")
	}
	r.issued[i]++
	r.refresh(i)
	return nil
}
func (r *Reference) Resolve(i, ordinal int, y bool) error {
	if i < 0 || i >= len(r.issued) || ordinal < 1 || ordinal > r.issued[i] || r.known[i]&(1<<(ordinal-1)) != 0 {
		return errors.New("reference receipt")
	}
	r.known[i] |= 1 << (ordinal - 1)
	if y {
		r.yes[i] |= 1 << (ordinal - 1)
	}
	r.refresh(i)
	return nil
}
