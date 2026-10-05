// Package researchregimelog is isolated research, not daemon authority.
package researchregimelog

import (
	"errors"
	"math"
)

const Classes = 9
const MaxMembers = 200
const MaxRows = MaxMembers * 64

var classPrior = [Classes]float64{.08, .01, .01, .64, .08, .08, .08, .01, .01}
var noise = [3]float64{0, .1, .2}

type Config struct {
	Reset, Hazard float64
	Cap           int
}
type Row struct {
	Member        int
	First, Second int
}
type Key struct{ Class, Start int }
type component struct {
	logWeight float64
	h, start  int
	odds      [MaxMembers]float64
}
type Result struct {
	Components, MaxComponents int
	LogEvidence               float64
	Discard, Envelope         float64
	Forecast                  []float64
	Support                   [][]Key
	parts                     []component
	members                   int
	logCompensation           float64
}

func finite(x float64) bool                { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func atom(bit int) float64                 { return .02 + .96*float64(bit) }
func atomProbability(high float64) float64 { return .02 + .96*high }
func high(base float64, h int) float64 {
	mu := base
	if h/3 == 1 {
		mu = (.9*base - .05) / .8
	}
	if h/3 == 2 {
		mu = 1 - base
	}
	return (math.Max(.02, math.Min(.98, mu)) - .02) / .96
}
func likelihood(r float64, h int, row Row) float64 {
	if row.First < 0 {
		return 1
	}
	eta := noise[h%3]
	sum := 0.
	for y := 0; y < 2; y++ {
		p := r
		if y == 0 {
			p = 1 - r
		}
		for _, w := range []int{row.First, row.Second} {
			if w < 0 {
				continue
			}
			if w == y {
				p *= 1 - eta
			} else {
				p *= eta
			}
		}
		sum += p
	}
	return sum
}
func validate(base []float64, cfg Config, rows []Row) error {
	if len(base) < 1 || len(base) > MaxMembers || len(rows) > MaxRows || !finite(cfg.Reset) || !finite(cfg.Hazard) || cfg.Reset < 0 || cfg.Reset > 1 || cfg.Hazard < 0 || cfg.Hazard > 1 || cfg.Cap < 0 || cfg.Cap > 256 || (cfg.Cap > 0 && cfg.Cap < Classes) || (cfg.Cap == 0 && (len(base) > 4 || len(rows) > 64)) {
		return errors.New("regime limits")
	}
	for _, p := range base {
		if !finite(p) || p < .25 || p > .925+1e-12 {
			return errors.New("regime baseline")
		}
	}
	for _, r := range rows {
		if r.Member < 0 || r.Member >= len(base) || r.First < -1 || r.First > 1 || r.Second < -1 || r.Second > 1 || (r.First == -1 && r.Second != -1) {
			return errors.New("regime row")
		}
	}
	return nil
}
func logAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	if b > a {
		a, b = b, a
	}
	return a + math.Log1p(math.Exp(b-a))
}
func logMass(parts []component) float64 {
	x := math.Inf(-1)
	for _, c := range parts {
		x = math.Max(x, c.logWeight)
	}
	if math.IsInf(x, -1) {
		return x
	}
	s := 0.
	for _, c := range parts {
		s += math.Exp(c.logWeight - x)
	}
	return x + math.Log(s)
}
func logit(p float64) float64 { return math.Log(p) - math.Log1p(-p) }
func logRates(z float64) (low, high float64) {
	if z >= 0 {
		s := math.Log1p(math.Exp(-z))
		return -z - s, -s
	}
	s := math.Log1p(math.Exp(z))
	return -s, z - s
}
func rate(z float64) float64 { _, h := logRates(z); return math.Exp(h) }
func movedOdds(z, p, hazard float64) float64 {
	if hazard == 0 {
		return z
	}
	if hazard == 1 {
		return logit(p)
	}
	l, h := logRates(z)
	a, b := math.Log1p(-hazard), math.Log(hazard)
	return logAdd(a+h, b+math.Log(p)) - logAdd(a+l, b+math.Log1p(-p))
}
func initial(base []float64) []component {
	parts := make([]component, Classes)
	for h, w := range classPrior {
		parts[h] = component{logWeight: math.Log(w), h: h, start: -1}
		for i, p := range base {
			parts[h].odds[i] = logit(high(p, h))
		}
	}
	return parts
}
func (r *Result) addEvidence(x float64) {
	// Keep a low part so branch ratios do not subtract two rounded large totals.
	s := r.LogEvidence + x
	b := s - r.LogEvidence
	err := (r.LogEvidence - (s - b)) + (x - b)
	lo := r.logCompensation + err
	r.LogEvidence = s + lo
	r.logCompensation = lo - (r.LogEvidence - s)
}
func Run(base []float64, cfg Config, rows []Row) (Result, error) { return run(base, cfg, rows, nil) }
func Conditional(base []float64, cfg Config, rows []Row, support [][]Key) (Result, error) {
	if len(support) != len(rows) {
		return Result{}, errors.New("regime support history length")
	}
	return run(base, cfg, rows, support)
}
func run(base []float64, cfg Config, rows []Row, support [][]Key) (Result, error) {
	if e := validate(base, cfg, rows); e != nil {
		return Result{}, e
	}
	r := Result{members: len(base), parts: initial(base), MaxComponents: Classes}
	r, _, e := resume(base, cfg, rows, support, prefix{state: r}, nil, false)
	return r, e
}
func (r Result) Joint(members int) ([]float64, error) {
	if members < 1 || members > 4 || members != r.members {
		return nil, errors.New("regime joint expansion limit")
	}
	width := 1 << members
	out := make([]float64, Classes*width)
	for _, c := range r.parts {
		for mask := 0; mask < width; mask++ {
			lp := c.logWeight
			for i := 0; i < members; i++ {
				l, h := logRates(c.odds[i])
				if mask&(1<<i) != 0 {
					lp += h
				} else {
					lp += l
				}
			}
			out[c.h*width+mask] += math.Exp(lp)
		}
	}
	return out, nil
}
