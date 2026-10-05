package researchdynvarcache

import (
	"errors"
	"math"
)

func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}
func (m *Model) Query(t Ticket, mode string) (out Option, err error) {
	if mode != "forecast" && mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" && mode != "model_class" && mode != "noise_class" {
		return out, errors.New("dispersion query mode")
	}
	if t.second || t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= m.count {
		return out, errors.New("dispersion query owner/epoch")
	}
	r := m.rows[t.slot]
	if r.first != 2 || r.second != 0 {
		return out, errors.New("dispersion query phase")
	}
	v, err := m.currentView(-1, nil)
	if err != nil {
		return out, err
	}
	w, err := m.memberWeights(r.member, v)
	if err != nil {
		return out, err
	}
	origin, err := m.origin(t.slot)
	if err != nil {
		return out, err
	}
	var mass, yes [States * RateAtoms]float64
	var classMass, classYes [States]float64
	total := 0.
	for k, p := range origin {
		for z, x := range p {
			j := k*RateAtoms + z
			mass[j] = w[k] * x
			f := m.factors[r.member][k%3][z]
			if f[bit(r.a)] > 0 {
				yes[j] = f[2+2*bit(r.a)+1] / f[bit(r.a)]
			} else if mass[j] > 0 {
				return out, errors.New("dispersion unsupported first")
			}
			if !finite(mass[j]) || !finite(yes[j]) || mass[j] < 0 || yes[j] < 0 || yes[j] > 1+2e-12 {
				return out, errors.New("dispersion query mass")
			}
			total += mass[j]
			out.Observed += mass[j] * yes[j]
			c := k
			if mode == "model_class" {
				c = k / 3
			}
			classMass[c] += mass[j]
			classYes[c] += mass[j] * yes[j]
		}
	}
	if math.Abs(total-1) > 2e-10 || !finite(out.Observed) || out.Observed < 0 || out.Observed > 1+2e-12 {
		return out, errors.New("dispersion query probability")
	}
	out.Observed = math.Min(1, out.Observed)
	if mode == "forecast" {
		return
	}
	out.Uncertainty = entropy(out.Observed)
	if mode == "uncertainty" {
		return
	}
	if mode == "model_class" || mode == "noise_class" {
		p := out.Observed
		if p > 0 && p < 1 {
			for c, q := range classMass {
				d := classYes[c] - p*q
				out.ClassGain += d * d / (p * (1 - p))
			}
		}
		return
	}
	if mode == "information" || mode == "falsification" {
		out.Information = out.Uncertainty
		for j, q := range mass {
			if q == 0 {
				continue
			}
			out.Information -= q * entropy(yes[j])
			if out.Observed > 0 && out.Observed < 1 {
				d := yes[j] - out.Observed
				out.EdgeCut += q * q * d * d / (out.Observed * (1 - out.Observed))
			}
		}
		if out.Information < -2e-10 || !finite(out.EdgeCut) {
			return out, errors.New("dispersion information")
		}
		out.Information = math.Max(0, out.Information)
		return
	}
	var before [MaxMembers]float64
	for i := range m.base {
		before[i], _, err = m.predictOne(i, v)
		if err != nil {
			return out, err
		}
	}
	var after [2][MaxMembers]float64
	for y := 0; y < 2; y++ {
		prob := out.Observed
		if y == 0 {
			prob = 1 - prob
		}
		if prob == 0 {
			after[y] = before
			continue
		}
		x := r
		x.second, x.b = 2, y == 1
		b, e := m.prepare(r.member, t.slot, x)
		if e != nil {
			return out, e
		}
		v, e := m.currentView(r.member, &b)
		if e != nil {
			return out, e
		}
		for i := range m.base {
			after[y][i], _, err = m.predictOne(i, v)
			if err != nil {
				return out, err
			}
		}
	}
	for i := range m.base {
		p := out.Observed
		if math.Abs(p*after[1][i]+(1-p)*after[0][i]-before[i]) > 2e-10 {
			return out, errors.New("dispersion predictive tower")
		}
		out.Value += (p*(after[1][i]-before[i])*(after[1][i]-before[i]) + (1-p)*(after[0][i]-before[i])*(after[0][i]-before[i])) / float64(len(m.base))
	}
	return
}
