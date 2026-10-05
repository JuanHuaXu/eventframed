package researchdynvarcache

import (
	"errors"
	"math"
)

// Replay member-ordinal transitions, replacing only the original trial factor.
func (m *Model) prepare(i, slot int, replacement row) (out belief, err error) {
	x := &m.members[i]
	for k := 0; k < States; k++ {
		a, h := k/3, k%3
		p, log := m.priors[i][a], 0.
		for n := 0; n < x.count; n++ {
			if n > 0 {
				p = m.transition(i, a, p)
			}
			r := m.rows[x.slots[n]]
			if x.slots[n] == slot {
				r = replacement
			}
			c := category(r)
			if c >= 0 {
				for z := range p {
					p[z] *= m.factors[i][h][z][c]
				}
			}
			var s float64
			p, s, err = normalize(p)
			if err != nil {
				return out, err
			}
			if s == 0 {
				log = math.Inf(-1)
				break
			}
			if c >= 0 {
				log += math.Log(s)
			}
		}
		out.p[k], out.log[k] = p, log
	}
	return
}

type view struct {
	joint       [States]float64
	replacement *belief
	member      int
}

func (m *Model) currentView(i int, c *belief) (out view, err error) {
	out.member, out.replacement = i, c
	if c == nil {
		if err = m.validLogs(-1, nil); err != nil {
			return out, err
		}
		out.joint = m.odds.joint
		return out, nil
	}
	w, e := m.rebuildOdds(i, c)
	out.joint, err = w.joint, e
	return
}
func (m *Model) memberWeights(i int, v view) (out [States]float64, err error) {
	if m.cfg.Mode == "noise" {
		return v.joint, nil
	}
	b := m.members[i].latest
	if v.member == i && v.replacement != nil {
		b = *v.replacement
	}
	if m.cfg.Mode == "individual" {
		if v.member != i || v.replacement == nil {
			return m.members[i].derived.individual, nil
		}
		var prior [States]float64
		for k := range prior {
			prior[k] = m.familyPrior[k/3] * noisePrior[k%3]
		}
		w, z, e := weights(b.log, prior)
		if e == nil && math.IsInf(z, -1) {
			e = errors.New("dispersion unsupported individual")
		}
		return w, e
	}
	for a := 0; a < 3; a++ {
		if v.joint[3*a] == 0 {
			continue
		}
		var hw [3]float64
		copy(hw[:], m.members[i].derived.noise[3*a:3*a+3])
		if v.member == i && v.replacement != nil {
			var e error
			hw, _, e = noiseWeights([3]float64{b.log[3*a], b.log[3*a+1], b.log[3*a+2]})
			if e != nil {
				return out, e
			}
		}
		for h, w := range hw {
			out[3*a+h] = v.joint[3*a] * w
		}
	}
	return
}
func (m *Model) predictOne(i int, v view) (clean, observed float64, err error) {
	w, err := m.memberWeights(i, v)
	if err != nil {
		return 0, 0, err
	}
	b := m.members[i].latest
	if v.member == i && v.replacement != nil {
		b = *v.replacement
	}
	for k, mass := range w {
		if mass == 0 {
			continue
		}
		p := m.members[i].derived.next[k]
		if v.member == i && v.replacement != nil {
			var e error
			p, _, e = normalize(m.transition(i, k/3, b.p[k]))
			if e != nil {
				return 0, 0, e
			}
		}
		for z, q := range p {
			clean += mass * q * rate(m.base[i], z)
			observed += mass * q * m.factors[i][k%3][z][1]
		}
	}
	if !finite(clean) || !finite(observed) || clean < 0 || observed < 0 || clean > 1+2e-12 || observed > 1+2e-12 {
		return 0, 0, errors.New("dispersion forecast bound")
	}
	return math.Min(1, clean), math.Min(1, observed), nil
}
func (m *Model) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("dispersion member")
	}
	v, e := m.currentView(-1, nil)
	if e != nil {
		return 0, 0, e
	}
	return m.predictOne(i, v)
}
func (m *Model) ModelWeights() (out [3]float64, err error) {
	v, err := m.currentView(-1, nil)
	if err != nil {
		return out, err
	}
	if m.cfg.Mode == "individual" {
		// Descriptive mean of member-local family posteriors, not a global belief.
		for i := range m.base {
			w, e := m.memberWeights(i, v)
			if e != nil {
				return out, e
			}
			for k, p := range w {
				out[k/3] += p / float64(len(m.base))
			}
		}
		return out, nil
	}
	for k, w := range v.joint {
		out[k/3] += w
	}
	return
}

// Smoothed rate at the original trial, conditioned only on revealed factors.
func (m *Model) origin(slot int) (out [States]vector, err error) {
	r := m.rows[slot]
	x := &m.members[r.member]
	for k := 0; k < States; k++ {
		a, h := k/3, k%3
		p, back := m.priors[r.member][a], vector{}
		var origin vector
		supported := true
		for n := 0; n < x.count; n++ {
			if n > 0 {
				p = m.transition(r.member, a, p)
			}
			c := category(m.rows[x.slots[n]])
			if c >= 0 {
				for z := range p {
					p[z] *= m.factors[r.member][h][z][c]
				}
			}
			var s float64
			p, s, err = normalize(p)
			if err != nil {
				return out, err
			}
			if s == 0 {
				supported = false
				break
			}
			if n == r.ordinal {
				origin = p
			}
		}
		if !supported {
			continue
		}
		for z := range back {
			back[z] = 1
		}
		for n := x.count - 1; n > r.ordinal; n-- {
			c := category(m.rows[x.slots[n]])
			if c >= 0 {
				for z := range back {
					back[z] *= m.factors[r.member][h][z][c]
				}
			}
			s := 0.
			for z, w := range m.priors[r.member][a] {
				s += w * back[z]
			}
			for z := range back {
				back[z] = (1-m.cfg.Hazard)*back[z] + m.cfg.Hazard*s
			}
			back, _, err = normalize(back)
			if err != nil {
				return out, err
			}
		}
		for z := range origin {
			origin[z] *= back[z]
		}
		out[k], _, err = normalize(origin)
		if err != nil {
			return out, err
		}
	}
	return
}
