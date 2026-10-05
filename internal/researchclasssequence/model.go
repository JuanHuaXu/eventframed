// Package researchclasssequence tests coherent contextual sharing in isolation.
// Its three model alternatives are hypotheses, not Anti-Pigeon merge authority.
// A single owner serializes all operations; this is not a production service.
package researchclasssequence

import (
	"errors"
	"math"
)

const MaxMembers, MaxTrials, RateAtoms, FieldAtoms, SharedStates, Block = 200, 64, 22, 16, 48, 128
const MaxBlocks = MaxMembers * MaxTrials / Block

type localVector [RateAtoms]float64
type sharedVector [SharedStates]float64
type checkpoint struct {
	p   sharedVector
	log [3]float64
}
type conditional struct {
	p   [3]localVector
	log [3]float64
}
type member struct {
	slots  [MaxTrials]int
	count  int
	latest conditional
}
type row struct {
	member, ordinal                int
	at, auditAt                    int64
	clean, observed, auditForecast float64
	first, second                  uint8
	a, b                           bool
}
type Config struct {
	Mode   string
	Hazard float64
}
type Model struct {
	base                []float64
	cfg                 Config
	epoch               uint64
	clock               int64
	cap, pending, count int
	priors              []localVector
	localFactors        [][66][6]float64
	fields              [][FieldAtoms]float64
	sharedFactors       [][SharedStates][6]float64
	members             []member
	rows                []row
	checkpoints         []checkpoint
	latest              checkpoint
}
type Ticket struct {
	owner    *Model
	epoch    uint64
	slot     int
	second   bool
	forecast float64
}

func (t Ticket) Forecast() float64 { return t.forecast }

type Receipt struct {
	Member, Ordinal, Measurement int
	Epoch                        uint64
	IssuedAt, ArrivedAt          int64
	Forecast                     float64
	Value                        bool
}
type Option struct{ Observed, Uncertainty, Information, EdgeCut, Value, ClassGain float64 }

var noise = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}
var fieldPrior = [FieldAtoms]float64{.72, .0012, .0108, .096, .0108, .0012, .0008, .0072, .064, .0072, .0008, .0008, .0072, .064, .0072, .0008}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func bit(x bool) int {
	if x {
		return 1
	}
	return 0
}
func rate(base float64, z int) float64 {
	if z == 0 {
		return base
	}
	return float64(z-1) / 20
}
func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }
func loading(f int, b float64) float64 {
	switch f {
	case 0:
		return 1
	case 1:
		return 2*b - 1
	default:
		return 2*(2*b-1)*(2*b-1) - 1
	}
}
func normalizeLocal(p localVector) (localVector, float64, error) {
	s := 0.
	for _, v := range p {
		if !finite(v) || v < 0 {
			return p, 0, errors.New("shared invalid local mass")
		}
		s += v
	}
	if !finite(s) {
		return p, 0, errors.New("shared local total")
	}
	if s == 0 {
		return localVector{}, 0, nil
	}
	for z := range p {
		p[z] /= s
	}
	return p, s, nil
}

// Normalize CONDITIONAL field states separately. Noise evidence stays in logs,
// so a rare but supported noise hypothesis is not irreversibly rounded away.
func normalizeShared(p sharedVector) (sharedVector, [3]float64, error) {
	var sums [3]float64
	for z, v := range p {
		if !finite(v) || v < 0 {
			return p, sums, errors.New("shared invalid field mass")
		}
		sums[z/FieldAtoms] += v
	}
	total := 0.
	for _, s := range sums {
		if !finite(s) {
			return p, sums, errors.New("shared field total")
		}
		total += s
	}
	if total <= 0 {
		return p, sums, errors.New("shared zero field support")
	}
	for z := range p {
		if sums[z/FieldAtoms] > 0 {
			p[z] /= sums[z/FieldAtoms]
		}
	}
	return p, sums, nil
}
func (m *Model) needsLocal() bool  { return m.cfg.Mode != "shared" }
func (m *Model) needsShared() bool { return m.cfg.Mode == "shared" || m.cfg.Mode == "hybrid" }
func emissions(p, eta float64) [6]float64 {
	q := eta + (1-2*eta)*p
	out := [6]float64{1 - q, q}
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			one, two := eta, eta
			if a == 1 {
				one = 1 - eta
			}
			if b == 1 {
				two = 1 - eta
			}
			out[2+2*a+b] = p*one*two + (1-p)*(1-one)*(1-two)
		}
	}
	return out
}
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > MaxMembers || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard > 1 {
		return nil, errors.New("shared constructor")
	}
	if cfg.Mode != "local" && cfg.Mode != "noise" && cfg.Mode != "shared" && cfg.Mode != "hybrid" {
		return nil, errors.New("shared mode")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("shared baseline")
		}
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, priors: make([]localVector, len(base)), localFactors: make([][66][6]float64, len(base)), fields: make([][FieldAtoms]float64, len(base)), sharedFactors: make([][SharedStates][6]float64, len(base)), members: make([]member, len(base)), rows: make([]row, len(base)*MaxTrials), checkpoints: make([]checkpoint, (len(base)*MaxTrials+Block-1)/Block)}
	grid := [5]float64{-8, -4, 0, 4, 8}
	gw := [5]float64{.01, .09, .8, .09, .01}
	for i, b := range base {
		w := [21]float64{1}
		for n := 0; n < 20; n++ {
			w[0] *= (1 - b + float64(n)) / (1 + float64(n))
		}
		for z := 0; z < 20; z++ {
			w[z+1] = w[z] * float64(20-z) / float64(z+1) * (b + float64(z)) / (1 - b + float64(19-z))
		}
		s := 0.
		for _, x := range w {
			s += x
		}
		m.priors[i][0] = .8
		for z, x := range w {
			m.priors[i][z+1] = .2 * x / s
		}
		m.fields[i][0] = b
		for f := 0; f < 3; f++ {
			lo, hi := -64., 64.
			for n := 0; n < 100; n++ {
				c := (lo + hi) / 2
				mean := 0.
				for z, x := range grid {
					mean += gw[z] * sigmoid(c+loading(f, b)*x)
				}
				if mean < b {
					lo = c
				} else {
					hi = c
				}
			}
			for z, x := range grid {
				m.fields[i][1+5*f+z] = sigmoid((lo+hi)/2 + loading(f, b)*x)
			}
		}
		mean := 0.
		for z, x := range fieldPrior {
			mean += x * m.fields[i][z]
		}
		if math.Abs(mean-b) > 3e-14 {
			return nil, errors.New("shared field prior mean")
		}
		for h, e := range noise {
			for z := 0; z < RateAtoms; z++ {
				m.localFactors[i][h*RateAtoms+z] = emissions(rate(b, z), e)
			}
			for z, p := range m.fields[i] {
				m.sharedFactors[i][h*FieldAtoms+z] = emissions(p, e)
			}
			m.members[i].latest.p[h] = m.priors[i]
		}
	}
	m.latest.p = m.initialShared()
	return m, nil
}
func (m *Model) initialShared() (p sharedVector) {
	for h := range noisePrior {
		for z, x := range fieldPrior {
			p[h*FieldAtoms+z] = x
		}
	}
	return
}
func (m *Model) localTransition(i int, p localVector) (out localVector) {
	s := 0.
	for _, x := range p {
		s += x
	}
	for z, w := range m.priors[i] {
		out[z] = (1-m.cfg.Hazard)*p[z] + m.cfg.Hazard*s*w
	}
	return
}
func (m *Model) sharedTransition(p sharedVector) (out sharedVector) {
	hazard := m.cfg.Hazard / float64(len(m.base))
	for h := 0; h < 3; h++ {
		s := 0.
		for z := 0; z < FieldAtoms; z++ {
			s += p[h*FieldAtoms+z]
		}
		for z, w := range fieldPrior {
			out[h*FieldAtoms+z] = (1-hazard)*p[h*FieldAtoms+z] + hazard*s*w
		}
	}
	return
}
func category(r row) int {
	if r.first != 2 {
		return -1
	}
	if r.second == 2 {
		return 2 + 2*bit(r.a) + bit(r.b)
	}
	return bit(r.a)
}
func (m *Model) validate(t Ticket, at int64, known bool) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= m.count || at < m.clock {
		return errors.New("shared owner/epoch/time")
	}
	r := m.rows[t.slot]
	if t.second {
		if r.second != 1 || at < r.auditAt {
			return errors.New("shared second replay")
		}
	} else {
		want := uint8(1)
		if known {
			want = 2
		}
		if r.first != want || at < r.at {
			return errors.New("shared first replay")
		}
	}
	return nil
}
func (m *Model) Pending() int { return m.pending }
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("shared epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}
