package observationlearners

import (
	"errors"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type stackingEvidence struct {
	generic, parity        [512]float64
	loo                    [][2]float64
	numerator, denominator float64
}

func looPredict(evidence, predictions [512]float64) float64 {
	logZ := familyLogEvidence(evidence)
	result := 0.
	for mask, e := range evidence {
		k := float64(bits.OnesCount(uint(mask)))
		logWeight := e + k*math.Log(1./3) + (9-k)*math.Log(2./3) - logZ
		result += math.Exp(logWeight) * predictions[mask]
	}
	return result
}

// fitStackingEvidence computes exact within-window LOO forecasts. The excluded
// label is removed from subset evidence as well as its parameter counts. This
// is not an as-of forecast for that observation's historical arrival time.
func fitStackingEvidence(samples []observation.Sample) (*stackingEvidence, error) {
	if len(samples) < 2 || len(samples) > 256 {
		return nil, errors.New("invalid stacking sample count")
	}
	g, err := fitSubset(samples)
	if err != nil {
		return nil, err
	}
	p, err := fitBooleanSpecialist(samples)
	if err != nil {
		return nil, err
	}
	var counts [19683]subsetCell
	var agreements [512]int
	for _, sample := range samples {
		for mask := uint16(0); mask < 512; mask++ {
			c := &counts[partialIndex(mask, sample.Bits&mask)]
			c.n++
			if sample.Outcome {
				c.yes++
			}
			if (bits.OnesCount16(sample.Bits&mask)%2 == 1) == sample.Outcome {
				agreements[mask]++
			}
		}
	}
	m := &stackingEvidence{generic: g.predictions, parity: p.predictions, loo: make([][2]float64, len(samples))}
	for i, sample := range samples {
		y := 0
		if sample.Outcome {
			y = 1
		}
		var ge, pe, gp, pp [512]float64
		for mask := uint16(0); mask < 512; mask++ {
			c := counts[partialIndex(mask, sample.Bits&mask)]
			gp[mask] = (float64(c.yes-y) + .5) / float64(c.n)
			rule := bits.OnesCount16(sample.Bits&mask)%2 == 1
			count := agreements[mask]
			if rule == sample.Outcome {
				count--
			}
			pp[mask] = (float64(count) + .5) / float64(len(samples))
			if !rule {
				pp[mask] = 1 - pp[mask]
			}
			gl, pl := gp[mask], pp[mask]
			if !sample.Outcome {
				gl, pl = 1-gl, 1-pl
			}
			// L(D) = L(D without i) * predictive(y_i | D without i).
			ge[mask] = g.evidence[mask] - math.Log(gl)
			pe[mask] = p.evidence[mask] - math.Log(pl)
		}
		m.loo[i] = [2]float64{looPredict(ge, gp), looPredict(pe, pp)}
		d := m.loo[i][1] - m.loo[i][0]
		m.numerator += d * (float64(y) - m.loo[i][0])
		m.denominator += d * d
	}
	return m, nil
}

func compileStacking(m *stackingEvidence, lambda float64) (*ConditionalForest, float64, error) {
	if m == nil || len(m.loo) < 2 || math.IsNaN(lambda) || math.IsInf(lambda, 0) || lambda < 0 {
		return nil, 0, errors.New("invalid stacking penalty or evidence")
	}
	a := 0.
	if denominator := m.denominator + lambda; denominator > 0 {
		a = math.Max(0, math.Min(1, m.numerator/denominator))
	}
	result := new(ConditionalForest)
	for x, g := range m.generic {
		result.cells[partialIndex(511, uint16(x))] = conditionalCell{1, (1-a)*g + a*m.parity[x]}
	}
	for i := len(result.cells) - 1; i >= 0; i-- {
		v, place := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := result.cells[i+place], result.cells[i+2*place]
				result.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			place *= 3
		}
	}
	return result, a, nil
}

// NewStackedConditional fits predictive weights, not posterior family odds.
// Only already eligible evidence belongs here; no retrieval or admission gate
// is implemented by this bounded research constructor.
func NewStackedConditional(samples []observation.Sample, lambda float64) (*ConditionalForest, error) {
	if math.IsNaN(lambda) || math.IsInf(lambda, 0) || lambda < 0 {
		return nil, errors.New("invalid stacking penalty")
	}
	m, err := fitStackingEvidence(samples)
	if err != nil {
		return nil, err
	}
	result, _, err := compileStacking(m, lambda)
	return result, err
}
