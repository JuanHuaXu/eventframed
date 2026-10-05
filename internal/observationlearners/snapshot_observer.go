package observationlearners

import (
	"errors"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// This archived-policy clone uses a read-only mixture view without modifying
// the earlier observer or its frozen artifacts. Hidden values enter only via Read.
func runSnapshotObserver(f snapshotRoutedLaw, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if (f.base.banks[0] == nil && f.base.banks[1] == nil) || reader == nil {
		return observation.Result{}, errors.New("missing conditional observer")
	}
	var mask, values, attempted uint16
	p, e := f.Forecast(0, 0)
	if e != nil {
		return observation.Result{}, e
	}
	result := observation.Result{Probability: p}
	for result.Cost < 6 {
		if reader.Epoch() != epoch {
			return observation.Result{}, errors.New("stale observation snapshot")
		}
		if result.Cost > 0 && (result.Probability <= .1 || result.Probability >= .9) {
			result.Stop = "confidence"
			break
		}
		best := -1.
		var chosen observation.View
		found := false
		for scope := 0; scope < 3; scope++ {
			for depth := 0; depth < 3; depth++ {
				v := observation.View{Scope: scope, Depth: depth}
				added := v.Mask() &^ attempted
				cost := bits.OnesCount16(added)
				if cost == 0 || cost > 6-result.Cost || (result.Cost == 0 && v != (observation.View{Scope: 0, Depth: 0})) {
					continue
				}
				parent := f.cell(mask, values)
				expected := 0.
				for sub := added; ; sub = (sub - 1) & added {
					child := f.cell(mask|added, values|sub)
					expected += child.mass / parent.mass * jointObservationEntropy(child.weighted/child.mass)
					if sub == 0 {
						break
					}
				}
				gain := math.Max(0, jointObservationEntropy(result.Probability)-expected) / float64(cost)
				if !found || gain > best+1e-12 {
					best = gain
					chosen = v
					found = true
				}
			}
		}
		if !found {
			result.Stop = "no_affordable_view"
			break
		}
		m, v, e := reader.Read(chosen)
		if e != nil {
			return observation.Result{}, e
		}
		if reader.Epoch() != epoch || m != chosen.Mask() || v&^m != 0 || (v^values)&(m&mask) != 0 {
			return observation.Result{}, errors.New("invalid observation snapshot")
		}
		result.Cost += bits.OnesCount16(chosen.Mask() &^ attempted)
		attempted |= chosen.Mask()
		mask |= m
		values |= v
		result.Probability, e = f.Forecast(mask, values)
		if e != nil {
			return observation.Result{}, e
		}
		result.Trace = append(result.Trace, observation.Step{View: chosen, Attempted: attempted, Observed: mask, Values: values, Probability: result.Probability})
	}
	if result.Stop == "" {
		result.Stop = "budget"
	}
	result.Observed = bits.OnesCount16(mask)
	return result, nil
}
