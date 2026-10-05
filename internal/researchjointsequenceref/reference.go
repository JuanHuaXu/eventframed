// Package researchjointsequenceref rebuilds a finite joint model using dense
// transition matrices and independently constructed likelihoods/prior masses.
package researchjointsequenceref

import (
	"errors"
	"math"
)

const states = 66

type distribution [states]float64
type record struct {
	first, second, requested bool
	a, b                     bool
}
type member struct {
	base    float64
	initial distribution
	matrix  [states][states]float64
	records []record
	forward []distribution
}
type Reference struct {
	members []member
	hazard  float64
}

func New(base []float64, hazard float64) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || math.IsNaN(hazard) || hazard < 0 || hazard > 1 {
		return nil, errors.New("reference contract")
	}
	r := &Reference{members: make([]member, len(base)), hazard: hazard}
	for i, b := range base {
		m := &r.members[i]
		m.base = b
		// Independent Polya-urn DP, not adjacent-mass recurrence.
		weights := [21]float64{1}
		for n := 0; n < 20; n++ {
			var next [21]float64
			for z := 0; z <= n; z++ {
				next[z+1] += weights[z] * (b + float64(z)) / (1 + float64(n))
				next[z] += weights[z] * (1 - b + float64(n-z)) / (1 + float64(n))
			}
			weights = next
		}
		prior := [22]float64{.8}
		for z, p := range weights {
			prior[z+1] = .2 * p
		}
		for h, w := range []float64{.8, .1, .1} {
			for z, p := range prior {
				m.initial[h*22+z] = w * p
			}
		}
		for from := 0; from < states; from++ {
			for to := 0; to < states; to++ {
				if from/22 != to/22 {
					continue
				}
				m.matrix[from][to] = hazard * prior[to%22]
				if from == to {
					m.matrix[from][to] += 1 - hazard
				}
			}
		}
	}
	return r, nil
}
func truth(base float64, state int) float64 {
	if state%22 == 0 {
		return base
	}
	return float64(state%22-1) / 20
}
func likelihood(base float64, state int, x record) float64 {
	if !x.first {
		return 1
	}
	eta := []float64{0, .1, .2}[state/22]
	sum := 0.
	for y := 0; y < 2; y++ {
		p := truth(base, state)
		if y == 0 {
			p = 1 - p
		}
		a := eta
		if (y == 1) == x.a {
			a = 1 - eta
		}
		p *= a
		if x.second {
			b := eta
			if (y == 1) == x.b {
				b = 1 - eta
			}
			p *= b
		}
		sum += p
	}
	return sum
}
func normalize(p distribution) (distribution, float64, error) {
	s := 0.
	for _, v := range p {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return p, 0, errors.New("reference mass")
		}
		s += v
	}
	if s <= 0 {
		return p, 0, errors.New("reference zero")
	}
	for j := range p {
		p[j] /= s
	}
	return p, s, nil
}
func (m *member) transition(p distribution) distribution {
	var out distribution
	for from, w := range p {
		for to := 0; to < states; to++ {
			out[to] += w * m.matrix[from][to]
		}
	}
	return out
}
func (m *member) rebuild(n int, replacement record) ([]distribution, float64, error) {
	out := make([]distribution, len(m.records))
	p := m.initial
	logEvidence := 0.
	for j, x := range m.records {
		if j > 0 {
			p = m.transition(p)
		}
		if j == n {
			x = replacement
		}
		for state := range p {
			p[state] *= likelihood(m.base, state, x)
		}
		var s float64
		var e error
		p, s, e = normalize(p)
		if e != nil {
			return nil, 0, e
		}
		logEvidence += math.Log(s)
		out[j] = p
	}
	return out, logEvidence, nil
}
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= len(r.members) || len(r.members[i].records) >= 64 {
		return errors.New("reference issue")
	}
	m := &r.members[i]
	m.records = append(m.records, record{})
	f, _, e := m.rebuild(-1, record{})
	m.forward = f
	return e
}
func (r *Reference) Observe(i, n, measurement int, value bool) error {
	if i < 0 || i >= len(r.members) || n < 1 || n > len(r.members[i].records) {
		return errors.New("reference origin")
	}
	m := &r.members[i]
	x := m.records[n-1]
	if measurement == 1 {
		if x.first {
			return errors.New("reference first replay")
		}
		x.first, x.a = true, value
	} else if measurement == 2 {
		if !x.first || !x.requested || x.second {
			return errors.New("reference second phase")
		}
		x.second, x.b = true, value
	} else {
		return errors.New("reference measurement")
	}
	f, _, e := m.rebuild(n-1, x)
	if e != nil {
		return e
	}
	m.records[n-1] = x
	m.forward = f
	return nil
}
func (r *Reference) Audit(i, n int) error {
	if i < 0 || i >= len(r.members) || n < 1 || n > len(r.members[i].records) {
		return errors.New("reference audit origin")
	}
	x := &r.members[i].records[n-1]
	if !x.first || x.requested {
		return errors.New("reference audit phase")
	}
	x.requested = true
	return nil
}
func (m *member) next(p distribution) (float64, float64) {
	p = m.transition(p)
	q, o := 0., 0.
	for state, w := range p {
		clean := truth(m.base, state)
		q += w * clean
		eta := []float64{0, .1, .2}[state/22]
		o += w * (eta + (1-2*eta)*clean)
	}
	return q, o
}
func (r *Reference) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(r.members) {
		return 0, 0, errors.New("reference member")
	}
	m := &r.members[i]
	p := m.initial
	if len(m.forward) > 0 {
		p = m.forward[len(m.forward)-1]
	}
	q, o := m.next(p)
	return q, o, nil
}

type Option struct{ Observed, Uncertainty, Information, EdgeCut, Value float64 }

func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log(1-p)
}
func (r *Reference) Query(i, n int) (Option, error) {
	if i < 0 || i >= len(r.members) || n < 1 || n > len(r.members[i].records) {
		return Option{}, errors.New("reference query origin")
	}
	m := &r.members[i]
	x := m.records[n-1]
	if !x.first || x.requested {
		return Option{}, errors.New("reference query phase")
	}
	_, logBase, e := m.rebuild(-1, record{})
	if e != nil {
		return Option{}, e
	}
	var yes, no []distribution
	var ey, en float64
	x.second, x.b = true, true
	yes, ey, e = m.rebuild(n-1, x)
	if e != nil {
		return Option{}, e
	}
	x.b = false
	no, en, e = m.rebuild(n-1, x)
	if e != nil {
		return Option{}, e
	}
	q := math.Exp(ey - logBase)
	qn := math.Exp(en - logBase)
	if math.Abs(q+qn-1) > 1e-10 {
		return Option{}, errors.New("reference conditional mass")
	}
	q = math.Min(q, 1)
	out := Option{Observed: q, Uncertainty: entropy(q), Information: entropy(q)}
	back := distribution{}
	for state := range back {
		back[state] = 1
	}
	for j := len(m.records) - 1; j >= n; j-- {
		var next distribution
		for from := 0; from < states; from++ {
			for to := 0; to < states; to++ {
				next[from] += m.matrix[from][to] * likelihood(m.base, to, m.records[j]) * back[to]
			}
		}
		back, _, e = normalize(next)
		if e != nil {
			return Option{}, e
		}
	}
	p := m.forward[n-1]
	for state := range p {
		p[state] *= back[state]
	}
	p, _, e = normalize(p)
	if e != nil {
		return Option{}, e
	}
	first := m.records[n-1]
	paired := first
	paired.second, paired.b = true, true
	for state, w := range p {
		if w == 0 {
			continue
		}
		py := likelihood(m.base, state, paired) / likelihood(m.base, state, first)
		out.Information -= w * entropy(py)
		term := -1.
		if q > 0 {
			term += py * py / q
		}
		if q < 1 {
			term += (1 - py) * (1 - py) / (1 - q)
		}
		out.EdgeCut += w * w * term
	}
	b, _, _ := r.Predict(i)
	a, _ := m.next(yes[len(yes)-1])
	c, _ := m.next(no[len(no)-1])
	if math.Abs(q*a+(1-q)*c-b) > 1e-10 {
		return Option{}, errors.New("reference tower")
	}
	out.Value = (q*(a-b)*(a-b) + (1-q)*(c-b)*(c-b)) / float64(len(r.members))
	out.Information = math.Max(0, out.Information)
	out.EdgeCut = math.Max(0, out.EdgeCut)
	return out, nil
}
