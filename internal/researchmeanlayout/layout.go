// Package researchmeanlayout measures a complete proposed storage layout.
// It initializes the V73 prior, but is NOT a fitted learner or serving backend.
// Learning, delayed replay and observation queries must be integrated separately.
package researchmeanlayout

import (
	"errors"
	"math"
)

const Means, Families, Noises, MaxMembers, MaxTrials = 27, 3, 3, 200, 64

// The point family has no rate vector. Free has no unsupported spike atom.
// Noise odds and next means reserve publication caches, not a second rate table.
type member struct {
	mean     [Means]float64
	current  [Means][Noises][22]float64
	free     [Means][Noises][21]float64
	log      [Means * Families * Noises]float64
	noise    [Means * Families * Noises]float64
	nextMean [Means * Families * Noises]float64
	slots    [MaxTrials]int
	count    int
}

// Reserve the complete planned immutable issue/paired-observation journal.
// These fields are intentionally unpopulated: this is a layout preflight.
type record struct {
	member, ordinal                int
	at, auditAt                    int64
	clean, observed, auditForecast float64
	first, second                  uint8
	a, b                           bool
}

type Layout struct {
	base                []float64
	members             []member
	rows                []record
	joint               [Means * Families * Noises]float64
	finiteSum           [Means * Families * Noises]float64
	zero                [Means * Families * Noises]int
	epoch               uint64
	clock               int64
	cap, pending, count int
}

var meanPrior = [Means]float64{.1, .8, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004}
var noisePrior = [Noises]float64{.8, .1, .1}

func meanMaps(b float64, i, size int) (out [Means]float64) {
	out[0], out[1] = b, (.9*b-.05)/.8
	t := 2
	for _, a := range []float64{.1, .3, .5, .7, .9} {
		for _, c := range []float64{-.8, -.4, 0, .4, .8} {
			out[t] = math.Max(.02, math.Min(.98, a+c*float64(i)/float64(size-1)))
			t++
		}
	}
	return
}

func urn(p, strength float64) (out [21]float64) {
	alpha, beta := strength*p, strength*(1-p)
	out[0] = 1
	for n := 0; n < 20; n++ {
		out[0] *= (beta + float64(n)) / (strength + float64(n))
	}
	for z := 0; z < 20; z++ {
		out[z+1] = out[z] * float64(20-z) / float64(z+1) * (alpha + float64(z)) / (beta + float64(19-z))
	}
	total := 0.
	for _, x := range out {
		total += x
	}
	for z := range out {
		out[z] /= total
	}
	return
}

func New(base []float64) (*Layout, error) {
	if len(base) < 2 || len(base) > MaxMembers {
		return nil, errors.New("mean layout member cap")
	}
	for _, b := range base {
		if math.IsNaN(b) || math.IsInf(b, 0) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("mean layout baseline")
		}
	}
	m := &Layout{base: append([]float64(nil), base...), members: make([]member, len(base)), rows: make([]record, len(base)*MaxTrials), epoch: 1, cap: 2 * len(base) * MaxTrials}
	for t, p := range meanPrior {
		for a := 0; a < Families; a++ {
			for h, w := range noisePrior {
				m.joint[(t*Families+a)*Noises+h] = p / 3 * w
			}
		}
	}
	for i, b := range base {
		x := &m.members[i]
		x.mean = meanMaps(b, i, len(base))
		for t, p := range x.mean {
			current, free := urn(p, 1), urn(p, 2)
			for h, w := range noisePrior {
				x.current[t][h][0] = .8
				for z := range current {
					x.current[t][h][z+1] = .2 * current[z]
					x.free[t][h][z] = free[z]
				}
				for a := 0; a < Families; a++ {
					k := (t*Families+a)*Noises + h
					x.noise[k], x.nextMean[k] = w, p
				}
			}
		}
	}
	return m, nil
}
