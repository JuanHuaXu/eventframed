// Package researchmeanjoint is an isolated finite joint mean/dispersion model.
// Hyperstates are static; member rates move on the member issue clock. Methods
// require one serialized owner. No daemon, truth authority or persistence is wired.
package researchmeananchor

import (
	"errors"
	"math"
)

const Means, Families, Noises, States, RateAtoms, MaxMembers, MaxTrials = 27, 3, 3, 243, 22, 200, 64

type vector [RateAtoms]float64
type belief struct {
	anchor                       int
	current                      [Means][Noises][22]float64
	free                         [Means][Noises][21]float64
	log, noise, next, individual [States]float64
	marginal                     [Means * Families]float64
}
type member struct {
	mean   [Means]float64
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
	Mode, Family, Means string
	Hazard              float64
}
type odds struct {
	joint, finiteSum [States]float64
	zero             [States]int
}
type Model struct {
	base                []float64
	cfg                 Config
	epoch               uint64
	clock               int64
	cap, pending, count int
	members             []member
	rows                []row
	meanPrior           [Means]float64
	familyPrior         [Families]float64
	grid                [Noises][21][6]float64
	odds                odds
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

var eta = [Noises]float64{0, .1, .2}
var noisePrior = [Noises]float64{.8, .1, .1}
var meanPrior = [Means]float64{.1, .8, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004, .004}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func bit(x bool) int {
	if x {
		return 1
	}
	return 0
}
func width(a int) int {
	if a == 0 {
		return 1
	}
	if a == 1 {
		return 22
	}
	return 21
}
func index(t, a, h int) int { return (t*Families+a)*Noises + h }
func rate(mu float64, a, z int) float64 {
	if a == 2 {
		return float64(z) / 20
	}
	if z == 0 {
		return mu
	}
	return float64(z-1) / 20
}
func emission(p, e float64) (out [6]float64) {
	q := e + (1-2*e)*p
	out[0], out[1] = 1-q, q
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			u, v := e, e
			if a == 1 {
				u = 1 - e
			}
			if b == 1 {
				v = 1 - e
			}
			out[2+2*a+b] = p*u*v + (1-p)*(1-u)*(1-v)
		}
	}
	return
}
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
func prior(mu float64, a int) (out vector) {
	if a == 0 {
		out[0] = 1
		return
	}
	strength, spike := 1., .8
	if a == 2 {
		strength, spike = 2, 0
	}
	alpha, beta := strength*mu, strength*(1-mu)
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
	if a == 1 {
		out[0] = spike
		for z, x := range w {
			out[z+1] = (1 - spike) * x / s
		}
	} else {
		for z, x := range w {
			out[z] = x / s
		}
	}
	return
}
func normalize(p vector) (vector, float64, error) {
	s := 0.
	for _, x := range p {
		if !finite(x) || x < 0 {
			return p, 0, errors.New("mean joint invalid mass")
		}
		s += x
	}
	if !finite(s) {
		return p, 0, errors.New("mean joint invalid total")
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
			return out, 0, errors.New("mean joint invalid evidence")
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
	maximum := math.Inf(-1)
	for h, x := range logs {
		if math.IsNaN(x) || math.IsInf(x, 1) {
			return out, 0, errors.New("mean joint invalid noise evidence")
		}
		maximum = math.Max(maximum, math.Log(noisePrior[h])+x)
	}
	if math.IsInf(maximum, -1) {
		return out, maximum, nil
	}
	s := 0.
	for h, x := range logs {
		out[h] = math.Exp(math.Log(noisePrior[h]) + x - maximum)
		s += out[h]
	}
	for h := range out {
		out[h] /= s
	}
	return out, maximum + math.Log(s), nil
}
func (m *Model) active(t, a int) bool { return m.meanPrior[t]*m.familyPrior[a] > 0 }
func (b *belief) load(t, a, h int) (out vector) {
	if a == 0 {
		out[0] = 1
	} else if a == 1 {
		out = b.current[t][h]
	} else {
		copy(out[:21], b.free[t][h][:])
	}
	return
}
func (b *belief) store(t, a, h int, p vector) {
	if a == 1 {
		b.current[t][h] = p
	} else if a == 2 {
		copy(b.free[t][h][:], p[:21])
	}
}
func (m *Model) move(p, pi vector) (out vector) {
	s := 0.
	for _, q := range p {
		s += q
	}
	for z, w := range pi {
		out[z] = (1-m.cfg.Hazard)*p[z] + m.cfg.Hazard*s*w
	}
	return
}
func (m *Model) factor(mu float64, a, h, z, c int) float64 {
	if a != 2 && z == 0 {
		return emission(mu, eta[h])[c]
	}
	if a == 1 {
		z--
	}
	return m.grid[h][z][c]
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

func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > MaxMembers || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard > 1 || (cfg.Mode != "noise" && cfg.Mode != "local" && cfg.Mode != "individual") {
		return nil, errors.New("mean joint constructor")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("mean joint baseline")
		}
	}
	fp := [Families]float64{}
	switch cfg.Family {
	case "learn":
		fp = [Families]float64{1. / 3, 1. / 3, 1. / 3}
	case "baseline":
		fp[0] = 1
	case "current":
		fp[1] = 1
	case "free":
		fp[2] = 1
	default:
		return nil, errors.New("mean joint family")
	}
	mp := [Means]float64{}
	switch cfg.Means {
	case "learn":
		mp = meanPrior
	case "baseline":
		mp[0] = 1
	default:
		return nil, errors.New("mean joint mean family")
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, meanPrior: mp, familyPrior: fp, members: make([]member, len(base)), rows: make([]row, len(base)*MaxTrials)}
	for h, e := range eta {
		for z := 0; z < 21; z++ {
			m.grid[h][z] = emission(float64(z)/20, e)
		}
	}
	for i, b := range base {
		x := &m.members[i]
		x.latest.anchor = -1
		x.mean = meanMaps(b, i, len(base))
		for t, mu := range x.mean {
			for a := 0; a < 3; a++ {
				if !m.active(t, a) {
					continue
				}
				p := prior(mu, a)
				for h := 0; h < 3; h++ {
					x.latest.store(t, a, h, p)
				}
			}
		}
		if e := m.derive(i, &x.latest, 0); e != nil {
			return nil, e
		}
	}
	w, e := m.rebuildOdds(-1, nil)
	if e != nil {
		return nil, e
	}
	m.odds = w
	return m, nil
}
