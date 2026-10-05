// Package researchpriorref independently integrates complete as-of histories.
// It is offline research verification, not the candidate's cached suffix code.
package researchpriorref

import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/researchprior"
	"math"
)

type Reference struct {
	cfg        researchprior.Config
	prior      [][3][21]float64
	issued     []int
	known, yes []uint64
	logs, next [][3]float64
}

func New(base []float64, cfg researchprior.Config) *Reference {
	r := &Reference{cfg: cfg, prior: make([][3][21]float64, len(base)), issued: make([]int, len(base)), known: make([]uint64, len(base)), yes: make([]uint64, len(base)), logs: make([][3]float64, len(base)), next: make([][3]float64, len(base))}
	for i, b := range base {
		mean := b
		if cfg.Center == "inverse" {
			mean = (.9*b - .05) / .8
		}
		for h, p := range []float64{mean, 1 - mean, .5} {
			sum := 0.
			for z := 0; z < 21; z++ {
				q := (float64(z) + .5) / 21
				v := math.Exp((cfg.Strength*p-1)*math.Log(q) + (cfg.Strength*(1-p)-1)*math.Log1p(-q))
				r.prior[i][h][z] = v
				sum += v
			}
			for z := 0; z < 21; z++ {
				r.prior[i][h][z] /= sum
			}
		}
		r.refresh(i)
	}
	return r
}
func (r *Reference) refresh(i int) {
	for h := 0; h < 3; h++ {
		mass := r.prior[i][h]
		for j := 0; j < r.issued[i]; j++ {
			if j > 0 {
				total := 0.
				for _, v := range mass {
					total += v
				}
				for z := range mass {
					mass[z] = (1-r.cfg.Hazard)*mass[z] + r.cfg.Hazard*r.prior[i][h][z]*total
				}
			}
			if r.known[i]&(1<<j) != 0 {
				for z := range mass {
					q := (float64(z) + .5) / 21
					if r.yes[i]&(1<<j) == 0 {
						q = 1 - q
					}
					mass[z] *= q
				}
			}
		}
		total, moment, priorMean := 0., 0., 0.
		for z, v := range mass {
			q := (float64(z) + .5) / 21
			total += v
			moment += v * q
			priorMean += r.prior[i][h][z] * q
		}
		r.logs[i][h] = math.Log(total)
		r.next[i][h] = moment / total
		if r.issued[i] > 0 {
			r.next[i][h] = (1-r.cfg.Hazard)*r.next[i][h] + r.cfg.Hazard*priorMean
		}
	}
}
func (r *Reference) Predict(i int) (float64, error) {
	if i < 0 || i >= len(r.issued) {
		return 0, errors.New("reference member")
	}
	l := r.logs[i]
	if r.cfg.Shared {
		l = [3]float64{}
		for _, v := range r.logs {
			for h := range l {
				l[h] += v[h]
			}
		}
	}
	w := [3]float64{.8, .1, .1}
	maximum := math.Inf(-1)
	for h := range w {
		w[h] = math.Log(w[h]) + l[h]
		maximum = math.Max(maximum, w[h])
	}
	sum, q := 0., 0.
	for h := range w {
		w[h] = math.Exp(w[h] - maximum)
		sum += w[h]
	}
	for h := range w {
		q += w[h] / sum * r.next[i][h]
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
