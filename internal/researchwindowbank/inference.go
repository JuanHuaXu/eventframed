// Package researchlocal tests pooled and member-local corrections. This is
// isolated research, not an Anti-Pigeon certificate or production learner.
package researchwindowbank

import (
	"errors"
	parent "github.com/JuanHuaXu/eventframed/internal/researchlocal"
	paired "github.com/JuanHuaXu/eventframed/internal/researchpaired"
	"math"
	"sort"
)

const Atoms, MaxDepth, MaxNodes, MaxTrials = 21, 7, 255, 64
const LocalAtoms = 22

var noise = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}

type Config = parent.Config
type node struct {
	counts                          [6]int
	finite, compensation, posterior [3][Atoms]float64
	zeros                           [3][Atoms]int
	leaf, tree, stop                [3]float64
	localFinite, localComp          [3]float64
	localZeros                      [3]int
}
type individual struct {
	counts                          [6]int
	finite, compensation, posterior [3][LocalAtoms]float64
	zeros                           [3][LocalAtoms]int
	evidence, mean                  [3]float64
}
type trial struct {
	seq, member, ordinal           int
	first, second                  uint8
	w1, w2                         bool
	at, auditAt                    int64
	clean, observed, auditForecast float64
}
type Ticket struct {
	owner  *Model
	epoch  uint64
	slot   int
	second bool
	q      float64
}

func (t Ticket) Forecast() float64 { return t.q }

type Receipt = paired.Receipt
type Option = paired.Option
type PredictiveOption struct{ Observed, Value float64 }
type Model struct {
	base              []float64
	cfg               Config
	epoch             uint64
	clock             int64
	cap, pending, seq int
	issued, ranks     []int
	trials            []trial
	seqSlots          []int
	rates             [][Atoms]float64
	factors           [][3][Atoms][6]float64
	individuals       []individual
	localRates        [][LocalAtoms]float64
	localPriors       [][LocalAtoms]float64
	localFactors      [][3][LocalAtoms][6]float64
	leafMembers       [MaxNodes]int
	nodes             [MaxNodes]node
	weights           [3]float64
}

// Only the changed root-to-leaf path is copied. Index0 means the unchanged
// published node; scratch values never escape into committed state on failure.
type plan struct {
	nodes    [MaxDepth + 1]node
	indices  [MaxNodes]uint8
	order    [MaxDepth + 1]int
	size     int
	weights  [3]float64
	memberID int
	member   individual
}

func finite(x float64) bool     { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }
func prior(z int) float64 {
	if z == 0 {
		return .8
	}
	return .01
}
func offset(z int) float64 {
	if z <= 10 {
		return -.6 * float64(z)
	}
	return .6 * float64(z-10)
}
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || cfg.Depth < 0 || cfg.Depth > MaxDepth || cfg.Window < 1 || cfg.Window > len(base)*MaxTrials {
		return nil, errors.New("anchor contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("anchor baseline")
		}
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, issued: make([]int, len(base)), ranks: make([]int, len(base)), trials: make([]trial, len(base)*MaxTrials), seqSlots: make([]int, len(base)*MaxTrials), rates: make([][Atoms]float64, len(base)), factors: make([][3][Atoms][6]float64, len(base))}
	m.individuals = make([]individual, len(base))
	m.localRates = make([][LocalAtoms]float64, len(base))
	m.localPriors = make([][LocalAtoms]float64, len(base))
	m.localFactors = make([][3][LocalAtoms][6]float64, len(base))
	ordered := append([]float64(nil), base...)
	sort.Float64s(ordered)
	for i, b := range base {
		m.ranks[i] = sort.SearchFloat64s(ordered, b) * (1 << cfg.Depth) / len(base)
		m.leafMembers[(1<<cfg.Depth)-1+m.ranks[i]]++
		if e := m.initializeIndividual(i, b); e != nil {
			return nil, e
		}
		lo, hi := -40., 40.
		for k := 0; k < 100; k++ {
			a := (lo + hi) / 2
			mean := 0.
			for z := 1; z < Atoms; z++ {
				mean += sigmoid(a+offset(z)) / 20
			}
			if mean < b {
				lo = a
			} else {
				hi = a
			}
		}
		m.rates[i][0] = b
		mean := .8 * b
		for z := 1; z < Atoms; z++ {
			m.rates[i][z] = sigmoid((lo+hi)/2 + offset(z))
			mean += .01 * m.rates[i][z]
		}
		if math.Abs(mean-b) > 2e-14 {
			return nil, errors.New("anchor moment solve")
		}
		for h, eta := range noise {
			for z, p := range m.rates[i] {
				q := eta + (1-2*eta)*p
				m.factors[i][h][z][0], m.factors[i][h][z][1] = math.Log1p(-q), math.Log(q)
				for a := 0; a < 2; a++ {
					for b := 0; b < 2; b++ {
						one, two := eta, eta
						if a == 1 {
							one = 1 - eta
						}
						if b == 1 {
							two = 1 - eta
						}
						m.factors[i][h][z][2+2*a+b] = math.Log(p*one*two + (1-p)*(1-one)*(1-two))
					}
				}
			}
		}
	}
	for n := (1 << (cfg.Depth + 1)) - 2; n >= 0; n-- {
		if e := m.refresh(&m.nodes[n], n, nil); e != nil {
			return nil, e
		}
	}
	var e error
	m.weights, e = m.normalize(nil)
	return m, e
}
func (m *Model) nodeAt(n int, p *plan) *node {
	if p != nil && p.indices[n] > 0 {
		return &p.nodes[p.indices[n]-1]
	}
	return &m.nodes[n]
}
func logAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	hi, lo := math.Max(a, b), math.Min(a, b)
	return hi + math.Log1p(math.Exp(lo-hi))
}
func (m *Model) refresh(x *node, n int, p *plan) error {
	total := 0
	for _, c := range x.counts {
		if c < 0 {
			return errors.New("anchor counts")
		}
		total += c
	}
	if total > m.cfg.Window {
		return errors.New("anchor window count")
	}
	for h := 0; h < 3; h++ {
		maximum := math.Inf(-1)
		var terms [Atoms]float64
		for z := range terms {
			v := x.finite[h][z]
			if !finite(v) || !finite(x.compensation[h][z]) || x.zeros[h][z] < 0 || x.zeros[h][z] > total {
				return errors.New("anchor finite/support fault")
			}
			v += math.Log(prior(z))
			if x.zeros[h][z] > 0 {
				v = math.Inf(-1)
			}
			terms[z] = v
			maximum = math.Max(maximum, v)
		}
		x.leaf[h] = math.Inf(-1)
		for z := range terms {
			x.posterior[h][z] = prior(z)
		}
		if finite(maximum) {
			sum := 0.
			for z, v := range terms {
				x.posterior[h][z] = math.Exp(v - maximum)
				sum += x.posterior[h][z]
			}
			if !finite(sum) || sum <= 0 {
				return errors.New("anchor leaf normalization")
			}
			for z := range terms {
				x.posterior[h][z] /= sum
			}
			x.leaf[h] = maximum + math.Log(sum)
		}
		split := x.localFinite[h]
		if !finite(split) || !finite(x.localComp[h]) || x.localZeros[h] < 0 || x.localZeros[h] > m.leafMembers[n] {
			return errors.New("local terminal product fault")
		}
		if x.localZeros[h] > 0 {
			split = math.Inf(-1)
		}
		if n < (1<<m.cfg.Depth)-1 {
			split = m.nodeAt(2*n+1, p).tree[h] + m.nodeAt(2*n+2, p).tree[h]
		}
		leaf := x.leaf[h] - math.Ln2
		x.tree[h], x.stop[h] = logAdd(leaf, split-math.Ln2), 1
		if finite(x.tree[h]) {
			x.stop[h] = math.Exp(leaf - x.tree[h])
		}
		if math.IsNaN(x.tree[h]) || math.IsInf(x.tree[h], 1) || !finite(x.stop[h]) {
			return errors.New("anchor tree fault")
		}
	}
	return nil
}
func (m *Model) normalize(p *plan) ([3]float64, error) {
	var w [3]float64
	maximum := math.Inf(-1)
	root := m.nodeAt(0, p)
	for h := range w {
		w[h] = math.Log(noisePrior[h]) + root.tree[h]
		if math.IsNaN(w[h]) || math.IsInf(w[h], 1) {
			return w, errors.New("anchor root fault")
		}
		maximum = math.Max(maximum, w[h])
	}
	if !finite(maximum) {
		return w, errors.New("anchor root support")
	}
	sum := 0.
	for h := range w {
		w[h] = math.Exp(w[h] - maximum)
		sum += w[h]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("anchor root normalization")
	}
	for h := range w {
		w[h] /= sum
	}
	return w, nil
}
func (m *Model) predict(i int, p *plan, weights [3]float64) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("anchor member")
	}
	q, o := 0., 0.
	for h, w := range weights {
		if w == 0 {
			continue
		}
		n := (1 << m.cfg.Depth) - 1 + m.ranks[i]
		x := m.nodeAt(n, p)
		v := 0.
		for z, p := range x.posterior[h] {
			v += p * m.rates[i][z]
		}
		local := &m.individuals[i]
		if p != nil && p.memberID == i {
			local = &p.member
		}
		v = x.stop[h]*v + (1-x.stop[h])*local.mean[h]
		for n > 0 {
			n = (n - 1) / 2
			x = m.nodeAt(n, p)
			mean := 0.
			for z, p := range x.posterior[h] {
				mean += p * m.rates[i][z]
			}
			v = x.stop[h]*mean + (1-x.stop[h])*v
		}
		q += w * v
		o += w * (noise[h] + (1-2*noise[h])*v)
	}
	if !finite(q) || !finite(o) || q <= 0 || q >= 1 || o <= 0 || o >= 1 {
		return 0, 0, errors.New("anchor forecast")
	}
	return q, o, nil
}
func (m *Model) Predict(i int) (float64, float64, error) { return m.predict(i, nil, m.weights) }
func (m *Model) NoiseWeights() [3]float64                { return m.weights }
func (m *Model) Pending() int                            { return m.pending }
func category(x trial) int {
	if x.first != 2 {
		return -1
	}
	a, b := 0, 0
	if x.w1 {
		a = 1
	}
	if x.w2 {
		b = 1
	}
	if x.second == 2 {
		return 2 + 2*a + b
	}
	return a
}
func add(x *node, h, z int, v float64, sign int) error {
	if math.IsInf(v, -1) {
		x.zeros[h][z] += sign
		return nil
	}
	if !finite(v) {
		return errors.New("anchor factor fault")
	}
	// Compensated finite sums remain removable even when another factor has
	// zero support. Never add or subtract negative infinity from a log sum.
	y := float64(sign)*v - x.compensation[h][z]
	next := x.finite[h][z] + y
	x.compensation[h][z] = (next - x.finite[h][z]) - y
	x.finite[h][z] = next
	return nil
}
func (m *Model) prepare(i, old, next int) (plan, error) {
	var p plan
	if i < 0 || i >= len(m.base) || old < -1 || old >= 6 || next < -1 || next >= 6 {
		return p, errors.New("local prepare shape")
	}
	p.memberID, p.member = i, m.individuals[i]
	if e := m.prepareIndividual(&p.member, i, old, next); e != nil {
		return p, e
	}
	for n := (1 << m.cfg.Depth) - 1 + m.ranks[i]; ; n = (n - 1) / 2 {
		j := p.size
		p.size++
		p.indices[n] = uint8(j + 1)
		p.order[j] = n
		p.nodes[j] = m.nodes[n]
		x := &p.nodes[j]
		if j == 0 {
			for h := 0; h < 3; h++ {
				if e := addProduct(x, h, m.individuals[i].evidence[h], -1); e != nil {
					return p, e
				}
				if e := addProduct(x, h, p.member.evidence[h], 1); e != nil {
					return p, e
				}
			}
		}
		if old >= 0 {
			x.counts[old]--
		}
		if next >= 0 {
			x.counts[next]++
		}
		for h := 0; h < 3; h++ {
			for z := 0; z < Atoms; z++ {
				if old >= 0 {
					if e := add(x, h, z, m.factors[i][h][z][old], -1); e != nil {
						return p, e
					}
				}
				if next >= 0 {
					if e := add(x, h, z, m.factors[i][h][z][next], 1); e != nil {
						return p, e
					}
				}
			}
		}
		if e := m.refresh(x, n, &p); e != nil {
			return p, e
		}
		if n == 0 {
			break
		}
	}
	var e error
	p.weights, e = m.normalize(&p)
	return p, e
}
func (m *Model) commit(p *plan) {
	for j := 0; j < p.size; j++ {
		m.nodes[p.order[j]] = p.nodes[j]
	}
	m.weights = p.weights
	m.individuals[p.memberID] = p.member
}
