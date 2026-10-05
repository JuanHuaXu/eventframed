package researchmeanjoint

import (
	"errors"
	"math"
)

// Derived caches are prepared with their belief, never changed by a query.
func (m *Model) derive(i int, b *belief) error {
	var ip [States]float64
	for t, mu := range m.members[i].mean {
		for a := 0; a < 3; a++ {
			if !m.active(t, a) {
				continue
			}
			pi := prior(mu, a)
			k := index(t, a, 0)
			hw, z, e := noiseWeights([3]float64{b.log[k], b.log[k+1], b.log[k+2]})
			if e != nil {
				return e
			}
			if math.IsInf(z, -1) {
				return errors.New("mean joint unsupported member")
			}
			b.marginal[t*3+a] = z
			for h, w := range hw {
				b.noise[k+h] = w
				ip[k+h] = m.meanPrior[t] * m.familyPrior[a] * noisePrior[h]
				if math.IsInf(b.log[k+h], -1) {
					b.next[k+h] = 0
					continue
				}
				p, _, e := normalize(m.move(b.load(t, a, h), pi))
				if e != nil {
					return e
				}
				q := 0.
				for z := 0; z < width(a); z++ {
					q += p[z] * rate(mu, a, z)
				}
				if !finite(q) || q < 0 || q > 1+2e-12 {
					return errors.New("mean joint next mean")
				}
				b.next[k+h] = q
			}
		}
	}
	w, z, e := weights(b.log, ip)
	if e != nil {
		return e
	}
	if math.IsInf(z, -1) {
		return errors.New("mean joint unsupported local model")
	}
	b.individual = w
	return nil
}
func (m *Model) validLogs(i int, b *belief) error {
	for j := range m.members {
		logs := &m.members[j].latest.log
		if j == i && b != nil {
			logs = &b.log
		}
		for k, x := range logs {
			if m.active(k/9, (k/3)%3) && (math.IsNaN(x) || math.IsInf(x, 1)) {
				return errors.New("mean joint invalid source evidence")
			}
		}
	}
	return nil
}
func (m *Model) rebuildOdds(i int, b *belief) (out odds, err error) {
	if err = m.validLogs(i, b); err != nil {
		return
	}
	if m.cfg.Mode == "individual" {
		return
	}
	var logs, pr [States]float64
	for t, p := range m.meanPrior {
		for a, q := range m.familyPrior {
			if p*q == 0 {
				continue
			}
			k := index(t, a, 0)
			if m.cfg.Mode == "noise" {
				for h, w := range noisePrior {
					pr[k+h] = p * q * w
				}
				for j := range m.members {
					l := &m.members[j].latest.log
					if j == i && b != nil {
						l = &b.log
					}
					for h := 0; h < 3; h++ {
						v := l[k+h]
						if math.IsInf(v, -1) {
							out.zero[k+h]++
						} else {
							out.finiteSum[k+h] += v
						}
					}
				}
			} else {
				pr[k] = p * q
				for j := range m.members {
					v := m.members[j].latest.marginal[t*3+a]
					if j == i && b != nil {
						v = b.marginal[t*3+a]
					}
					if !finite(v) {
						return out, errors.New("mean joint invalid member marginal")
					}
					out.finiteSum[k] += v
				}
			}
		}
	}
	for k, v := range out.finiteSum {
		logs[k] = v
		if out.zero[k] > 0 {
			logs[k] = math.Inf(-1)
		}
	}
	w, z, e := weights(logs, pr)
	if e != nil {
		return out, e
	}
	if math.IsInf(z, -1) {
		return out, errors.New("mean joint unsupported global model")
	}
	out.joint = w
	return
}

type view struct {
	joint       [States]float64
	member      int
	replacement *belief
}

func (m *Model) currentView(i int, b *belief) (out view, err error) {
	out.member, out.replacement = i, b
	if b == nil {
		if err = m.validLogs(-1, nil); err != nil {
			return
		}
		out.joint = m.odds.joint
		return
	}
	w, e := m.rebuildOdds(i, b)
	out.joint, err = w.joint, e
	return
}
func (m *Model) memberWeights(i int, v view) (out [States]float64) {
	if m.cfg.Mode == "noise" {
		return v.joint
	}
	b := &m.members[i].latest
	if v.member == i && v.replacement != nil {
		b = v.replacement
	}
	if m.cfg.Mode == "individual" {
		return b.individual
	}
	for t := 0; t < Means; t++ {
		for a := 0; a < 3; a++ {
			k := index(t, a, 0)
			for h := 0; h < 3; h++ {
				out[k+h] = v.joint[k] * b.noise[k+h]
			}
		}
	}
	return
}
func (m *Model) predictOne(i int, v view) (clean, observed float64, err error) {
	w := m.memberWeights(i, v)
	b := &m.members[i].latest
	if v.member == i && v.replacement != nil {
		b = v.replacement
	}
	for k, mass := range w {
		if mass == 0 {
			continue
		}
		p := b.next[k]
		clean += mass * p
		observed += mass * (eta[k%3] + (1-2*eta[k%3])*p)
	}
	if !finite(clean) || !finite(observed) || clean < 0 || observed < 0 || clean > 1+2e-12 || observed > 1+2e-12 {
		return 0, 0, errors.New("mean joint forecast bounds")
	}
	return math.Min(1, clean), math.Min(1, observed), nil
}
func (m *Model) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("mean joint member")
	}
	v, e := m.currentView(-1, nil)
	if e != nil {
		return 0, 0, e
	}
	return m.predictOne(i, v)
}
func (m *Model) Posterior(i int) ([States]float64, error) {
	if i < 0 || i >= len(m.base) {
		return [States]float64{}, errors.New("mean joint posterior member")
	}
	v, e := m.currentView(-1, nil)
	if e != nil {
		return [States]float64{}, e
	}
	return m.memberWeights(i, v), nil
}
func (m *Model) ModelWeights() (out [3]float64, err error) {
	v, e := m.currentView(-1, nil)
	if e != nil {
		return out, e
	}
	if m.cfg.Mode == "individual" {
		for i := range m.base {
			w := m.memberWeights(i, v)
			for k, p := range w {
				out[(k/3)%3] += p / float64(len(m.base))
			}
		}
	} else {
		for k, p := range v.joint {
			out[(k/3)%3] += p
		}
	}
	return
}

// Late pairs replace the original factor; arrival never advances rate dynamics.
func (m *Model) prepare(i, slot int, replacement row) (out belief, err error) {
	x := &m.members[i]
	for t, mu := range x.mean {
		for a := 0; a < 3; a++ {
			if !m.active(t, a) {
				continue
			}
			pi := prior(mu, a)
			for h := 0; h < 3; h++ {
				p, log := pi, 0.
				for n := 0; n < x.count; n++ {
					if n > 0 {
						p = m.move(p, pi)
					}
					r := m.rows[x.slots[n]]
					if x.slots[n] == slot {
						r = replacement
					}
					c := category(r)
					if c >= 0 {
						for z := 0; z < width(a); z++ {
							p[z] *= m.factor(mu, a, h, z, c)
						}
					}
					var s float64
					p, s, err = normalize(p)
					if err != nil {
						return
					}
					if s == 0 {
						log = math.Inf(-1)
						break
					}
					if c >= 0 {
						log += math.Log(s)
					}
				}
				out.store(t, a, h, p)
				out.log[index(t, a, h)] = log
			}
		}
	}
	err = m.derive(i, &out)
	return
}

// Smoothed original rate uses only revealed factors, including later member ones.
func (m *Model) origin(slot int) (out [States]vector, err error) {
	r := m.rows[slot]
	x := &m.members[r.member]
	for t, mu := range x.mean {
		for a := 0; a < 3; a++ {
			if !m.active(t, a) {
				continue
			}
			pi := prior(mu, a)
			for h := 0; h < 3; h++ {
				p := pi
				var at vector
				supported := true
				for n := 0; n < x.count; n++ {
					if n > 0 {
						p = m.move(p, pi)
					}
					c := category(m.rows[x.slots[n]])
					if c >= 0 {
						for z := 0; z < width(a); z++ {
							p[z] *= m.factor(mu, a, h, z, c)
						}
					}
					var s float64
					p, s, err = normalize(p)
					if err != nil {
						return
					}
					if s == 0 {
						supported = false
						break
					}
					if n == r.ordinal {
						at = p
					}
				}
				if !supported {
					continue
				}
				var back vector
				for z := 0; z < width(a); z++ {
					back[z] = 1
				}
				for n := x.count - 1; n > r.ordinal; n-- {
					c := category(m.rows[x.slots[n]])
					if c >= 0 {
						for z := 0; z < width(a); z++ {
							back[z] *= m.factor(mu, a, h, z, c)
						}
					}
					s := 0.
					for z, w := range pi {
						s += w * back[z]
					}
					for z := 0; z < width(a); z++ {
						back[z] = (1-m.cfg.Hazard)*back[z] + m.cfg.Hazard*s
					}
					back, _, err = normalize(back)
					if err != nil {
						return
					}
				}
				for z := range at {
					at[z] *= back[z]
				}
				out[index(t, a, h)], _, err = normalize(at)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
