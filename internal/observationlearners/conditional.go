package observationlearners

import (
	"errors"
	"math"
)

type conditionalCell struct{ mass, weighted float64 }

// ConditionalForest freezes both the fitted tree and an input distribution.
// Its 3^9 states are a finite research representation, not a scalable default.
type ConditionalForest struct{ cells [19683]conditionalCell }

func partialIndex(mask, values uint16) int {
	i, p := 0, 1
	for bit := uint16(1); bit < 512; bit <<= 1 {
		if mask&bit != 0 {
			i += p
			if values&bit != 0 {
				i += p
			}
		}
		p *= 3
	}
	return i
}

func NewConditionalForest(f *Forest, weights [512]float64) (*ConditionalForest, error) {
	if f == nil {
		return nil, errors.New("missing conditional forest")
	}
	m := new(ConditionalForest)
	for x, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w <= 0 || w > 1e6 {
			return nil, errors.New("invalid conditional weight")
		}
		m.cells[partialIndex(511, uint16(x))] = conditionalCell{w, w * f.Predict(uint16(x))}
	}
	// Every missing ternary coordinate has children with larger indices, so a
	// descending pass sums disjoint completions without double counting.
	for i := len(m.cells) - 1; i >= 0; i-- {
		v, p := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := m.cells[i+p], m.cells[i+2*p]
				m.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			p *= 3
		}
	}
	return m, nil
}

func (m *ConditionalForest) Forecast(mask, values uint16) (float64, error) {
	if m == nil || mask >= 512 || values&^mask != 0 {
		return 0, errors.New("invalid conditional observation")
	}
	c := m.cells[partialIndex(mask, values)]
	if c.mass <= 0 {
		return 0, errors.New("uninitialized conditional forecast")
	}
	return c.weighted / c.mass, nil
}
