package researchsharedsequence

import (
	"errors"
	"math"
)

const probeStates = 66 + 66 + SharedStates

type probe struct{ mass, yes [probeStates]float64 }

func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}
func (m *Model) queryProbe(t Ticket) (out probe, v view, err error) {
	if t.second || t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= m.count {
		return out, v, errors.New("shared query owner/epoch")
	}
	r := m.rows[t.slot]
	if r.first != 2 || r.second != 0 {
		return out, v, errors.New("shared query phase")
	}
	v, err = m.currentView(-1, nil, nil)
	if err != nil {
		return out, v, err
	}
	a := bit(r.a)
	if v.model[0] > 0 || v.model[1] > 0 {
		origin, e := m.localOrigin(t.slot)
		if e != nil {
			return out, v, e
		}
		w, _, e := weights3(m.members[r.member].latest.log, noisePrior)
		if e != nil {
			return out, v, e
		}
		for h := 0; h < 3; h++ {
			for z, p := range origin[h] {
				k := h*RateAtoms + z
				f := m.localFactors[r.member][k]
				ratio := 0.
				if f[a] > 0 {
					ratio = f[2+2*a+1] / f[a]
				} else if p > 0 && (w[h]*v.model[0]+v.noise[h]*v.model[1]) > 0 {
					return out, v, errors.New("shared unsupported first")
				}
				out.mass[k] = v.model[0] * w[h] * p
				out.mass[66+k] = v.model[1] * v.noise[h] * p
				out.yes[k], out.yes[66+k] = ratio, ratio
			}
		}
	}
	if v.model[2] > 0 {
		origin, e := m.sharedOrigin(t.slot)
		if e != nil {
			return out, v, e
		}
		w, _, e := weights3(v.shared.log, noisePrior)
		if e != nil {
			return out, v, e
		}
		for z, p := range origin {
			f := m.sharedFactors[r.member][z]
			ratio := 0.
			if f[a] > 0 {
				ratio = f[2+2*a+1] / f[a]
			} else if p > 0 && w[z/FieldAtoms] > 0 {
				return out, v, errors.New("shared unsupported field first")
			}
			out.mass[132+z], out.yes[132+z] = v.model[2]*w[z/FieldAtoms]*p, ratio
		}
	}
	sum := 0.
	for _, p := range out.mass {
		sum += p
	}
	if math.Abs(sum-1) > 2e-10 {
		return out, v, errors.New("shared probe normalization")
	}
	return
}

// Predictive value includes all affected targets. It is not the local-only
// shortcut: noise and model evidence can move other members' forecasts.
func (m *Model) Query(t Ticket, mode string) (out Option, err error) {
	if mode != "forecast" && mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" {
		return out, errors.New("shared query mode")
	}
	p, v, err := m.queryProbe(t)
	if err != nil {
		return out, err
	}
	for z, w := range p.mass {
		out.Observed += w * p.yes[z]
	}
	if !finite(out.Observed) || out.Observed < 0 || out.Observed > 1+2e-12 {
		return out, errors.New("shared query probability")
	}
	out.Observed = math.Min(1, out.Observed)
	if mode == "forecast" {
		return
	}
	out.Uncertainty = entropy(out.Observed)
	if mode == "uncertainty" {
		return
	}
	if mode == "information" || mode == "falsification" {
		out.Information = out.Uncertainty
		for z, w := range p.mass {
			if w == 0 {
				continue
			}
			y := p.yes[z]
			out.Information -= w * entropy(y)
			gain := -1.
			if out.Observed > 0 {
				gain += y * y / out.Observed
			}
			if out.Observed < 1 {
				gain += (1 - y) * (1 - y) / (1 - out.Observed)
			}
			out.EdgeCut += w * w * gain
		}
		if out.Information < -2e-10 || out.EdgeCut < -2e-10 {
			return out, errors.New("shared negative information")
		}
		out.Information = math.Max(0, out.Information)
		out.EdgeCut = math.Max(0, out.EdgeCut)
		return
	}
	before := [MaxMembers]float64{}
	for i := range m.base {
		before[i], _, err = m.predictOne(i, v)
		if err != nil {
			return out, err
		}
	}
	branch := [2][MaxMembers]float64{}
	for y := 0; y < 2; y++ {
		weight := out.Observed
		if y == 0 {
			weight = 1 - weight
		}
		if weight == 0 {
			branch[y] = before
			continue
		}
		r := m.rows[t.slot]
		r.second, r.b = 2, y == 1
		var c conditional
		var s sharedPrepared
		var cp *conditional
		var sp *checkpoint
		if m.needsLocal() {
			c, err = m.prepareLocal(r.member, t.slot, r)
			if err != nil {
				return out, err
			}
			cp = &c
		}
		if m.needsShared() {
			s, err = m.prepareShared(t.slot, r)
			if err != nil {
				return out, err
			}
			sp = &s.latest
		}
		after, e := m.currentView(r.member, cp, sp)
		if e != nil {
			return out, e
		}
		for i := range m.base {
			branch[y][i], _, err = m.predictOne(i, after)
			if err != nil {
				return out, err
			}
		}
	}
	for i := range m.base {
		one, zero := branch[1][i], branch[0][i]
		q := out.Observed
		if math.Abs(q*one+(1-q)*zero-before[i]) > 2e-10 {
			return out, errors.New("shared predictive tower")
		}
		out.Value += (q*(one-before[i])*(one-before[i]) + (1-q)*(zero-before[i])*(zero-before[i])) / float64(len(m.base))
	}
	return
}
