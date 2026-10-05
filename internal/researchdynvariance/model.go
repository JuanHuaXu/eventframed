// Package researchdynvariance is an isolated finite joint-model experiment.
// A static dispersion hyperstate can connect independent member rates; that is
// model-conditional borrowing, not Anti-Pigeon authority or authenticated truth.
// All methods require one serialized owner. No daemon or persistence is wired.
package researchdynvariance

import (
	"errors"
	"math"
)

const MaxMembers, MaxTrials, RateAtoms, States = 200, 64, 22, 9

type vector [RateAtoms]float64
type belief struct {
	p   [States]vector
	log [States]float64
}
type member struct {
	slots  [MaxTrials]int
	count  int
	latest belief
}
type row struct {
	member, ordinal                int
	at, auditAt                    int64
	clean, observed, auditForecast float64
	first, second                  uint8
	a, b                           bool
}
type Config struct {
	Mode, Family string
	Hazard       float64
}
type Model struct {
	base                []float64
	cfg                 Config
	epoch               uint64
	clock               int64
	cap, pending, count int
	priors              [][3]vector
	factors             [][3][RateAtoms][6]float64
	members             []member
	rows                []row
	familyPrior         [3]float64
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

var eta = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}

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
func normalize(p vector) (vector, float64, error) {
	s := 0.
	for _, x := range p {
		if !finite(x) || x < 0 {
			return p, 0, errors.New("dispersion invalid mass")
		}
		s += x
	}
	if !finite(s) {
		return p, 0, errors.New("dispersion invalid total")
	}
	if s == 0 {
		return vector{}, 0, nil
	}
	for z := range p {
		p[z] /= s
	}
	return p, s, nil
}
func weights(logs, prior [States]float64) (out [States]float64, total float64, err error) {
	maximum := math.Inf(-1)
	for k, w := range prior {
		if w == 0 {
			continue
		}
		if !finite(w) || w < 0 || math.IsNaN(logs[k]) || math.IsInf(logs[k], 1) {
			return out, 0, errors.New("dispersion invalid evidence")
		}
		maximum = math.Max(maximum, math.Log(w)+logs[k])
	}
	if math.IsInf(maximum, -1) {
		return out, maximum, nil
	}
	s := 0.
	for k, w := range prior {
		if w > 0 {
			out[k] = math.Exp(math.Log(w) + logs[k] - maximum)
			s += out[k]
		}
	}
	for k := range out {
		out[k] /= s
	}
	return out, maximum + math.Log(s), nil
}
func noiseWeights(logs [3]float64) (out [3]float64, total float64, err error) {
	var l, p [States]float64
	copy(l[:], logs[:])
	copy(p[:], noisePrior[:])
	w, z, e := weights(l, p)
	copy(out[:], w[:3])
	return out, z, e
}
func emission(p, noise float64) (out [6]float64) {
	q := noise + (1-2*noise)*p
	out[0], out[1] = 1-q, q
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			pa, pb := noise, noise
			if a == 1 {
				pa = 1 - noise
			}
			if b == 1 {
				pb = 1 - noise
			}
			out[2+2*a+b] = p*pa*pb + (1-p)*(1-pa)*(1-pb)
		}
	}
	return
}
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > MaxMembers || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard > 1 || (cfg.Mode != "local" && cfg.Mode != "noise" && cfg.Mode != "individual") {
		return nil, errors.New("dispersion constructor")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("dispersion baseline")
		}
	}
	prior := [3]float64{}
	switch cfg.Family {
	case "learn":
		prior = [3]float64{1. / 3, 1. / 3, 1. / 3}
	case "baseline":
		prior[0] = 1
	case "current":
		prior[1] = 1
	case "free":
		prior[2] = 1
	default:
		return nil, errors.New("dispersion family")
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, familyPrior: prior, priors: make([][3]vector, len(base)), factors: make([][3][RateAtoms][6]float64, len(base)), members: make([]member, len(base)), rows: make([]row, len(base)*MaxTrials)}
	for i, b := range base {
		m.priors[i][0][0] = 1
		for a := 1; a < 3; a++ {
			strength, spike := 1., .8
			if a == 2 {
				strength, spike = 2, 0
			}
			alpha, beta := strength*b, strength*(1-b)
			w := [21]float64{1}
			for n := 0; n < 20; n++ {
				w[0] *= (beta + float64(n)) / (strength + float64(n))
			}
			for z := 0; z < 20; z++ {
				w[z+1] = w[z] * float64(20-z) / float64(z+1) * (alpha + float64(z)) / (beta + float64(19-z))
			}
			s := 0.
			for _, x := range w {
				s += x
			}
			m.priors[i][a][0] = spike
			for z, x := range w {
				m.priors[i][a][z+1] = (1 - spike) * x / s
			}
		}
		for h, e := range eta {
			for z := 0; z < RateAtoms; z++ {
				m.factors[i][h][z] = emission(rate(b, z), e)
			}
			for a := 0; a < 3; a++ {
				m.members[i].latest.p[3*a+h] = m.priors[i][a]
			}
		}
	}
	return m, nil
}
func (m *Model) transition(i, a int, p vector) (out vector) {
	s := 0.
	for _, x := range p {
		s += x
	}
	for z, w := range m.priors[i][a] {
		out[z] = (1-m.cfg.Hazard)*p[z] + m.cfg.Hazard*s*w
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
