package researchclasssequence

import (
	"errors"
	"math"
)

func weights3(logs, prior [3]float64) (out [3]float64, total float64, err error) {
	maximum := math.Inf(-1)
	for h, w := range prior {
		if w == 0 {
			continue
		}
		if math.IsNaN(logs[h]) || math.IsInf(logs[h], 1) {
			return out, 0, errors.New("shared invalid log evidence")
		}
		maximum = math.Max(maximum, math.Log(w)+logs[h])
	}
	if math.IsInf(maximum, -1) {
		return out, maximum, nil
	}
	sum := 0.
	for h, w := range prior {
		if w > 0 {
			out[h] = math.Exp(math.Log(w) + logs[h] - maximum)
			sum += out[h]
		}
	}
	for h := range out {
		out[h] /= sum
	}
	return out, maximum + math.Log(sum), nil
}
func (m *Model) prepareLocal(i, slot int, replacement row) (out conditional, err error) {
	x := &m.members[i]
	for h := 0; h < 3; h++ {
		p := m.priors[i]
		log := 0.
		for n := 0; n < x.count; n++ {
			if n > 0 {
				p = m.localTransition(i, p)
			}
			r := m.rows[x.slots[n]]
			if x.slots[n] == slot {
				r = replacement
			}
			c := category(r)
			if c >= 0 {
				for z := range p {
					p[z] *= m.localFactors[i][h*RateAtoms+z][c]
				}
			}
			var s float64
			p, s, err = normalizeLocal(p)
			if err != nil {
				return out, err
			}
			if s == 0 {
				p = localVector{}
				log = math.Inf(-1)
				break
			}
			if c >= 0 {
				log += math.Log(s)
			}
		}
		out.p[h], out.log[h] = p, log
	}
	return
}

type sharedPrepared struct {
	latest      checkpoint
	blocks      [MaxBlocks]checkpoint
	first, last int
}

func (m *Model) prepareShared(slot int, replacement row) (out sharedPrepared, err error) {
	if m.cfg.Hazard == 0 {
		return m.prepareStaticShared(slot, replacement)
	}
	start := slot / Block
	current := m.checkpoints[start]
	out.first = start
	out.last = (m.count - 1) / Block
	for n := start * Block; n < m.count; n++ {
		if n%Block == 0 {
			out.blocks[n/Block] = current
		}
		if n > 0 {
			current.p = m.sharedTransition(current.p)
		}
		r := m.rows[n]
		if n == slot {
			r = replacement
		}
		c := category(r)
		if c >= 0 {
			for z := range current.p {
				current.p[z] *= m.sharedFactors[r.member][z][c]
			}
		}
		var s [3]float64
		current.p, s, err = normalizeShared(current.p)
		if err != nil {
			return out, err
		}
		if c >= 0 {
			for h, x := range s {
				if x == 0 {
					current.log[h] = math.Inf(-1)
				} else {
					current.log[h] += math.Log(x)
				}
			}
		}
	}
	out.latest = current
	return
}

// With a static field there is no reset floor. Rebuild log masses from the
// journal so an earlier underflowed field atom can regain posterior support.
func (m *Model) prepareStaticShared(slot int, replacement row) (out sharedPrepared, err error) {
	var logs sharedVector
	for h := 0; h < 3; h++ {
		for z, w := range fieldPrior {
			logs[h*FieldAtoms+z] = math.Log(w)
		}
	}
	for n := 0; n < m.count; n++ {
		r := m.rows[n]
		if n == slot {
			r = replacement
		}
		c := category(r)
		if c < 0 {
			continue
		}
		for z := range logs {
			x := m.sharedFactors[r.member][z][c]
			if !finite(x) || x < 0 {
				return out, errors.New("shared static invalid emission")
			}
			logs[z] += math.Log(x)
		}
	}
	for h := 0; h < 3; h++ {
		maximum := math.Inf(-1)
		for z := 0; z < FieldAtoms; z++ {
			maximum = math.Max(maximum, logs[h*FieldAtoms+z])
		}
		if math.IsInf(maximum, -1) {
			out.latest.log[h] = maximum
			continue
		}
		sum := 0.
		for z := 0; z < FieldAtoms; z++ {
			k := h*FieldAtoms + z
			out.latest.p[k] = math.Exp(logs[k] - maximum)
			sum += out.latest.p[k]
		}
		for z := 0; z < FieldAtoms; z++ {
			out.latest.p[h*FieldAtoms+z] /= sum
		}
		out.latest.log[h] = maximum + math.Log(sum)
	}
	if _, l, e := weights3(out.latest.log, noisePrior); e != nil || math.IsInf(l, -1) {
		return out, errors.New("shared static unsupported")
	}
	out.first, out.last = 1, 0
	return
}

type view struct {
	model, noise [3]float64
	replacement  *conditional
	member       int
	shared       checkpoint
}

func (m *Model) currentView(i int, c *conditional, s *checkpoint) (out view, err error) {
	out.member, out.replacement = i, c
	out.shared = m.latest
	if s != nil {
		out.shared = *s
	}
	var localLog, noiseLog [3]float64
	if m.needsLocal() {
		for j, x := range m.members {
			p := x.latest
			if j == i && c != nil {
				p = *c
			}
			_, l, e := weights3(p.log, noisePrior)
			if e != nil || math.IsInf(l, -1) {
				return out, errors.New("shared unsupported local model")
			}
			localLog[0] += l
			for h := range noiseLog {
				noiseLog[h] += p.log[h]
			}
		}
	}
	var sumNoise float64
	out.noise, sumNoise, err = weights3(noiseLog, noisePrior)
	if err != nil {
		return out, err
	}
	_, sumShared, e := weights3(out.shared.log, noisePrior)
	if e != nil {
		return out, e
	}
	logs := [3]float64{localLog[0], sumNoise, sumShared}
	prior := [3]float64{}
	switch m.cfg.Mode {
	case "local":
		prior[0] = 1
	case "noise":
		prior[1] = 1
	case "shared":
		prior[2] = 1
	case "hybrid":
		prior = [3]float64{1. / 3, 1. / 3, 1. / 3}
	}
	out.model, _, err = weights3(logs, prior)
	return
}
func (m *Model) predictOne(i int, v view) (clean, observed float64, err error) {
	p := m.members[i].latest
	if v.replacement != nil && v.member == i {
		p = *v.replacement
	}
	if v.model[0] > 0 || v.model[1] > 0 {
		w, _, e := weights3(p.log, noisePrior)
		if e != nil {
			return 0, 0, e
		}
		for h := 0; h < 3; h++ {
			weight := v.model[0]*w[h] + v.model[1]*v.noise[h]
			if weight == 0 {
				continue
			}
			next, _, e := normalizeLocal(m.localTransition(i, p.p[h]))
			if e != nil {
				return 0, 0, e
			}
			for z, q := range next {
				clean += weight * q * rate(m.base[i], z)
				observed += weight * q * m.localFactors[i][h*RateAtoms+z][1]
			}
		}
	}
	if v.model[2] > 0 {
		p, _, e := normalizeShared(m.sharedTransition(v.shared.p))
		if e != nil {
			return 0, 0, e
		}
		w, _, e := weights3(v.shared.log, noisePrior)
		if e != nil {
			return 0, 0, e
		}
		for z, q := range p {
			clean += v.model[2] * w[z/FieldAtoms] * q * m.fields[i][z%FieldAtoms]
			observed += v.model[2] * w[z/FieldAtoms] * q * m.sharedFactors[i][z][1]
		}
	}
	if !finite(clean) || !finite(observed) || clean < 0 || observed < 0 || clean > 1+2e-12 || observed > 1+2e-12 {
		return 0, 0, errors.New("shared forecast bound")
	}
	return math.Min(1, clean), math.Min(1, observed), nil
}
func (m *Model) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("shared member")
	}
	v, e := m.currentView(-1, nil, nil)
	if e != nil {
		return 0, 0, e
	}
	return m.predictOne(i, v)
}
func (m *Model) ModelWeights() ([3]float64, error) {
	v, e := m.currentView(-1, nil, nil)
	return v.model, e
}

// Conditional original-member state, with all presently visible local factors.
func (m *Model) localOrigin(slot int) (out [3]localVector, err error) {
	r := m.rows[slot]
	x := &m.members[r.member]
	for h := 0; h < 3; h++ {
		p := m.priors[r.member]
		var origin localVector
		supported := true
		for n := 0; n < x.count; n++ {
			if n > 0 {
				p = m.localTransition(r.member, p)
			}
			c := category(m.rows[x.slots[n]])
			if c >= 0 {
				for z := range p {
					p[z] *= m.localFactors[r.member][h*RateAtoms+z][c]
				}
			}
			var sum float64
			p, sum, err = normalizeLocal(p)
			if err != nil {
				return out, err
			}
			if sum == 0 {
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
		back := localVector{}
		for z := range back {
			back[z] = 1
		}
		for n := x.count - 1; n > r.ordinal; n-- {
			c := category(m.rows[x.slots[n]])
			if c >= 0 {
				for z := range back {
					back[z] *= m.localFactors[r.member][h*RateAtoms+z][c]
				}
			}
			sum := 0.
			for z, w := range m.priors[r.member] {
				sum += w * back[z]
			}
			for z := range back {
				back[z] = (1-m.cfg.Hazard)*back[z] + m.cfg.Hazard*sum
			}
			back, _, err = normalizeLocal(back)
			if err != nil {
				return out, err
			}
		}
		for z := range origin {
			origin[z] *= back[z]
		}
		out[h], _, err = normalizeLocal(origin)
		if err != nil {
			return out, err
		}
	}
	return
}
func (m *Model) sharedOrigin(slot int) (out sharedVector, err error) {
	if m.cfg.Hazard == 0 {
		return m.latest.p, nil
	}
	p := m.checkpoints[slot/Block].p
	for n := slot / Block * Block; n <= slot; n++ {
		if n > 0 {
			p = m.sharedTransition(p)
		}
		r := m.rows[n]
		c := category(r)
		if c >= 0 {
			for z := range p {
				p[z] *= m.sharedFactors[r.member][z][c]
			}
		}
		p, _, err = normalizeShared(p)
		if err != nil {
			return p, err
		}
	}
	back := sharedVector{}
	for z := range back {
		back[z] = 1
	}
	hazard := m.cfg.Hazard / float64(len(m.base))
	for n := m.count - 1; n > slot; n-- {
		r := m.rows[n]
		c := category(r)
		if c >= 0 {
			for z := range back {
				back[z] *= m.sharedFactors[r.member][z][c]
			}
		}
		for h := 0; h < 3; h++ {
			sum := 0.
			for z, w := range fieldPrior {
				sum += w * back[h*FieldAtoms+z]
			}
			for z := 0; z < FieldAtoms; z++ {
				k := h*FieldAtoms + z
				back[k] = (1-hazard)*back[k] + hazard*sum
			}
		}
		back, _, err = normalizeShared(back)
		if err != nil {
			return back, err
		}
	}
	for z := range p {
		p[z] *= back[z]
	}
	out, _, err = normalizeShared(p)
	return
}
