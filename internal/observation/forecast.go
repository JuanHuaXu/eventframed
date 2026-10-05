package observation

import "errors"

// ForecastObserved lets research experts share one acquired observation budget.
// It cannot inspect any coordinate outside the caller's recorded observed mask.
func (m *Model) ForecastObserved(mask, values uint16) (float64, error) {
	if m == nil || mask >= Universe || values&^mask != 0 {
		return 0, errors.New("invalid observed forecast input")
	}
	return m.predict(mask, values), nil
}
