// Package researchregimefrozen is an isolated changing-explanation preflight.
// No daemon, authority gate, store or plugin imports this package.
package researchregimefrozen

import (
	"errors"
	"math"
	"sort"
)

const Classes = 9
const MaxMembers = 200
const MaxRows = MaxMembers * 64

var classPrior = [Classes]float64{.08, .01, .01, .64, .08, .08, .08, .01, .01}
var noise = [3]float64{0, .1, .2}

type Config struct {
	Reset, Hazard float64
	Cap           int // Zero retains every reset component (small histories only).
}

// -1 denotes unavailable evidence, not a negative observation.
type Row struct {
	Member        int
	First, Second int
}

type component struct {
	weight float64
	h      int
	start  int
	hi     [MaxMembers]float64
}

type Result struct {
	Components, MaxComponents int
	LogEvidence               float64
	Discard, Envelope         float64
	Forecast                  []float64
	Support                   [][]Key
	parts                     []component
}

type Key struct {
	Class, Start int
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func atom(bit int) float64  { return .02 + .96*float64(bit) }
func high(base float64, h int) float64 {
	mu := base
	switch h / 3 {
	case 1:
		mu = (.9*base - .05) / .8
	case 2:
		mu = 1 - base
	}
	return (math.Max(.02, math.Min(.98, mu)) - .02) / .96
}

// Sum over one latent Y. The two measurements do not get independent Y draws.
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

func initial(base []float64) []component {
	parts := make([]component, Classes)
	for h, w := range classPrior {
		parts[h] = component{weight: w, h: h, start: -1}
		for i, p := range base {
			parts[h].hi[i] = high(p, h)
		}
	}
	return parts
}

// Run-length components retain the most recent reset and current class. Earlier
// resets affect their prefix probability, not their conditional member filters.
func Run(base []float64, cfg Config, rows []Row) (out Result, err error) {
	return run(base, cfg, rows, nil)
}

// Conditional freezes every historical latent support set before comparing
// hypothetical observations. Reranking inside a branch is a different law.
func Conditional(base []float64, cfg Config, rows []Row, support [][]Key) (Result, error) {
	if len(support) != len(rows) {
		return Result{}, errors.New("regime support history length")
	}
	return run(base, cfg, rows, support)
}

func run(base []float64, cfg Config, rows []Row, frozen [][]Key) (out Result, err error) {
	if err = validate(base, cfg, rows); err != nil {
		return
	}
	parts := initial(base)
	out.MaxComponents = len(parts)
	for t, r := range rows {
		next := make([]component, 0, len(parts)+Classes)
		if cfg.Reset < 1 {
			for _, c := range parts {
				c.weight *= 1 - cfg.Reset
				c.hi[r.Member] = (1-cfg.Hazard)*c.hi[r.Member] + cfg.Hazard*high(base[r.Member], c.h)
				next = append(next, c)
			}
		}
		if cfg.Reset > 0 {
			for _, c := range initial(base) {
				c.weight *= cfg.Reset
				c.start = t
				next = append(next, c)
			}
		}
		z, minFactor, maxFactor := 0., 1., 0.
		for j := range next {
			c := &next[j]
			low := likelihood(atom(0), c.h, r)
			hi := likelihood(atom(1), c.h, r)
			minFactor, maxFactor = math.Min(minFactor, math.Min(low, hi)), math.Max(maxFactor, math.Max(low, hi))
			f := (1-c.hi[r.Member])*low + c.hi[r.Member]*hi
			c.weight *= f
			if f > 0 {
				c.hi[r.Member] *= hi / f
			}
			z += c.weight
		}
		if !finite(z) || z <= 0 {
			return out, errors.New("regime evidence support")
		}
		out.LogEvidence += math.Log(z)
		for j := range next {
			next[j].weight /= z
		}
		// Conservative TV envelope: transition contracts previous error by
		// 1-kappa, then conditioning may amplify it. Zero likelihood floors
		// yield a vacuous envelope, never an invented finite certificate.
		e := (1 - cfg.Reset) * out.Envelope
		if e > 0 {
			if minFactor == 0 {
				e = 1
			} else {
				e = math.Min(1, 2*maxFactor*e/minFactor)
			}
		}
		out.MaxComponents = max(out.MaxComponents, len(next))
		if frozen != nil {
			keys := frozen[t]
			if len(keys) == 0 || (cfg.Cap > 0 && len(keys) > cfg.Cap) {
				return out, errors.New("regime support count")
			}
			selected := make([]component, 0, len(keys))
			seen := map[Key]bool{}
			retained := 0.
			for _, key := range keys {
				if seen[key] || key.Class < 0 || key.Class >= Classes || key.Start < -1 || key.Start > t {
					return out, errors.New("regime support identity")
				}
				seen[key] = true
				found := false
				for _, c := range next {
					if c.h == key.Class && c.start == key.Start {
						selected = append(selected, c)
						retained += c.weight
						found = true
						break
					}
				}
				if !found {
					return out, errors.New("regime support path binding")
				}
			}
			if retained <= 0 || !finite(retained) {
				return out, errors.New("regime conditional evidence support")
			}
			out.LogEvidence += math.Log(retained)
			for j := range selected {
				selected[j].weight /= retained
			}
			d := math.Max(0, 1-retained)
			out.Discard += d
			e = math.Min(1, e+d)
			next = selected
		} else if cfg.Cap > 0 && len(next) > cfg.Cap {
			sort.SliceStable(next, func(i, j int) bool { return next[i].weight > next[j].weight })
			d := 0.
			for _, c := range next[cfg.Cap:] {
				d += c.weight
			}
			retained := 0.
			for _, c := range next[:cfg.Cap] {
				retained += c.weight
			}
			if retained <= 0 || !finite(retained) {
				return out, errors.New("regime cap support")
			}
			// Cross-history evidence must include the probability of every
			// retained latent-path restriction, not just observation factors.
			out.LogEvidence += math.Log(retained)
			next = next[:cfg.Cap]
			for j := range next {
				next[j].weight /= retained
			}
			out.Discard += d
			e = math.Min(1, e+d)
		}
		keys := make([]Key, len(next))
		for j, c := range next {
			keys[j] = Key{c.h, c.start}
		}
		out.Support = append(out.Support, keys)
		out.Envelope, parts = e, next
	}
	out.parts, out.Components = parts, len(parts)
	out.Forecast = make([]float64, len(base))
	for i := range base {
		for _, c := range parts {
			p := (1-cfg.Hazard)*c.hi[i] + cfg.Hazard*high(base[i], c.h)
			out.Forecast[i] += (1 - cfg.Reset) * c.weight * atomProbability(p)
		}
		for h, w := range classPrior {
			out.Forecast[i] += cfg.Reset * w * atomProbability(high(base[i], h))
		}
	}
	return
}

func atomProbability(high float64) float64 { return .02 + .96*high }

// Joint reconstructs a SMALL-state posterior for independent oracle comparison.
// The scalable run-length inference does not use this exponential expansion.
func (r Result) Joint(members int) ([]float64, error) {
	if members < 1 || members > 4 {
		return nil, errors.New("regime joint expansion limit")
	}
	width := 1 << members
	out := make([]float64, Classes*width)
	for _, c := range r.parts {
		for mask := 0; mask < width; mask++ {
			p := c.weight
			for i := 0; i < members; i++ {
				if mask&(1<<i) != 0 {
					p *= c.hi[i]
				} else {
					p *= 1 - c.hi[i]
				}
			}
			out[c.h*width+mask] += p
		}
	}
	return out, nil
}

// Journal is a single-owner research ledger, not a concurrent serving API.
// Revealing is explicit; Run callers must supply an already-as-of evidence set.
type Journal struct {
	base  []float64
	cfg   Config
	rows  []Row
	clock int64
	state Result
}

func New(base []float64, cfg Config) (*Journal, error) {
	r, e := Run(base, cfg, nil)
	if e != nil {
		return nil, e
	}
	return &Journal{base: append([]float64(nil), base...), cfg: cfg, state: r, clock: -1}, nil
}

func (m *Journal) Predict(member int) (float64, error) {
	if member < 0 || member >= len(m.base) {
		return 0, errors.New("regime member")
	}
	if !finite(m.state.Forecast[member]) || m.state.Forecast[member] < 0 || m.state.Forecast[member] > 1 {
		return 0, errors.New("regime forecast bounds")
	}
	return m.state.Forecast[member], nil
}

func (m *Journal) Issue(member int, at int64) (ordinal int, forecast float64, err error) {
	if at < 0 || at < m.clock || len(m.rows) >= MaxRows {
		return 0, 0, errors.New("regime issue clock/cap")
	}
	forecast, err = m.Predict(member)
	if err != nil {
		return
	}
	rows := append(append([]Row(nil), m.rows...), Row{member, -1, -1})
	r, e := Run(m.base, m.cfg, rows)
	if e != nil {
		return 0, 0, e
	}
	m.rows, m.state, m.clock = rows, r, at
	return len(rows) - 1, forecast, nil
}

func (m *Journal) Reveal(ordinal, which, value int, at int64) error {
	if at < 0 || at < m.clock || ordinal < 0 || ordinal >= len(m.rows) || which < 1 || which > 2 || value < 0 || value > 1 {
		return errors.New("regime reveal binding")
	}
	row := m.rows[ordinal]
	if (which == 1 && row.First != -1) || (which == 2 && (row.First == -1 || row.Second != -1)) {
		return errors.New("regime reveal phase")
	}
	rows := append([]Row(nil), m.rows...)
	if which == 1 {
		rows[ordinal].First = value
	} else {
		rows[ordinal].Second = value
	}
	r, e := Run(m.base, m.cfg, rows)
	if e != nil {
		return e
	}
	m.rows, m.state, m.clock = rows, r, at
	return nil
}
