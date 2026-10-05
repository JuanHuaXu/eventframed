package researchpairedvalue

import (
	"errors"
	"math"
)

// PredictiveOption values an immediately revealed second measurement for the
// next clean prediction at each member. It is model utility, not a certificate
// of external improvement, delay-aware value, or source independence.
type PredictiveOption struct {
	Observed float64
	Value    float64
}

// PredictionValues evaluates one single-owner snapshot without committing any
// hypothetical evidence. Target weights are declared before outcomes, not
// supplied from hidden future labels. Every target is the next issued clean Y.
// This dense reference implementation charges two full integrations per origin.
func (m *Model) PredictionValues(origins []Ticket, targetWeights []float64) ([]PredictiveOption, error) {
	if len(origins) == 0 || len(origins) > len(m.base) || len(targetWeights) != len(m.base) {
		return nil, errors.New("prediction value shape")
	}
	sum := 0.
	for _, w := range targetWeights {
		if !finite(w) || w < 0 || w > 1 {
			return nil, errors.New("prediction value target weight")
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, errors.New("prediction value target normalization")
	}
	seen := make(map[int]bool, len(origins))
	for _, t := range origins {
		if t.second || t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) {
			return nil, errors.New("prediction value origin")
		}
		x := m.trials[t.slot]
		if x.first != 2 || x.second != 0 || seen[t.slot] {
			return nil, errors.New("prediction value availability")
		}
		seen[t.slot] = true
	}

	// Other members' state is unchanged by a hypothetical local factor, but their
	// forecasts still move when its evidence changes the shared component weights.
	means := make([][Components]float64, len(m.base))
	current := make([]float64, len(m.base))
	for i := range means {
		for c, w := range m.weights {
			for z := 0; z < Atoms; z++ {
				means[i][c] += m.next(i, c, z) * atom(z)
			}
			current[i] += w * means[i][c]
		}
		if !finite(current[i]) || current[i] < 0 || current[i] > 1 {
			return nil, errors.New("prediction value current law")
		}
	}
	out := make([]PredictiveOption, len(origins))
	for k, t := range origins {
		i := t.slot / MaxTrials
		x := m.trials[t.slot]
		x.second, x.w2 = 2, true
		yesRows, yesLogs, e := m.integrate(i, t.slot, x)
		if e != nil {
			return nil, e
		}
		x.w2 = false
		noRows, noLogs, e := m.integrate(i, t.slot, x)
		if e != nil {
			return nil, e
		}
		yes, no := 0., 0.
		for c, w := range m.weights {
			if w == 0 {
				continue
			}
			old := m.logs[i*Components+c]
			if !finite(old) {
				return nil, errors.New("prediction value live support")
			}
			py, pn := math.Exp(yesLogs[c]-old), math.Exp(noLogs[c]-old)
			if !finite(py) || !finite(pn) || py < 0 || pn < 0 || math.Abs(py+pn-1) > 2e-10 {
				return nil, errors.New("prediction value conditional support")
			}
			yes += w * py
			no += w * pn
		}
		if !finite(yes) || !finite(no) || math.Abs(yes+no-1) > 2e-10 || yes > 1+2e-10 || no > 1+2e-10 {
			return nil, errors.New("prediction value outcome law")
		}
		yes, no = math.Min(1, yes), math.Min(1, no)
		out[k].Observed = yes
		var tower = make([]float64, len(m.base))
		for branch := 0; branch < 2; branch++ {
			probability, rows, logs := yes, &yesRows, yesLogs
			if branch == 1 {
				probability, rows, logs = no, &noRows, noLogs
			}
			// An impossible branch has no conditional law and contributes zero.
			if probability == 0 {
				continue
			}
			weights, e := m.normalized(i, logs)
			if e != nil {
				return nil, e
			}
			var own [Components]float64
			for c := range own {
				for z := 0; z < Atoms; z++ {
					p := m.prior[(i*Families+c%Families)*Atoms+z]
					own[c] += ((1-m.cfg.Hazard)*rows[c][z] + m.cfg.Hazard*p) * atom(z)
				}
			}
			for j := range means {
				conditional := 0.
				for c, w := range weights {
					mu := means[j][c]
					if j == i {
						mu = own[c]
					}
					conditional += w * mu
				}
				if !finite(conditional) || conditional < 0 || conditional > 1 {
					return nil, errors.New("prediction value conditional law")
				}
				d := conditional - current[j]
				out[k].Value += probability * targetWeights[j] * d * d
				tower[j] += probability * conditional
			}
		}
		for j := range current {
			if math.Abs(tower[j]-current[j]) > 2e-10 {
				return nil, errors.New("prediction value tower defect")
			}
		}
		if !finite(out[k].Value) || out[k].Value < 0 || out[k].Value > .25+2e-10 {
			return nil, errors.New("prediction value risk bound")
		}
	}
	return out, nil
}
