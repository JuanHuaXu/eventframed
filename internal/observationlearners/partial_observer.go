package observationlearners

import (
	"errors"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// ForecastPartial integrates unobserved independent fair binary coordinates.
// This is an explicit synthetic-generator assumption, not a real-data default.
func (f *Forest) ForecastPartial(mask, values uint16) (float64, error) {
	if f == nil || mask >= 512 || values&^mask != 0 {
		return 0, errors.New("invalid partial forest input")
	}
	var visit func(*node) float64
	visit = func(n *node) float64 {
		if n.Feature < 0 {
			return float64(n.Yes+1) / float64(n.N+2)
		}
		bit := uint16(1 << n.Feature)
		if mask&bit == 0 {
			return .5 * (visit(n.Left) + visit(n.Right))
		}
		if values&bit == 0 {
			return visit(n.Left)
		}
		return visit(n.Right)
	}
	p := 0.
	for _, n := range f.Trees {
		p += visit(n) / 5
	}
	return p, nil
}

func binaryEntropy(p float64) float64 { return -p*math.Log(p) - (1-p)*math.Log1p(-p) }

// RunForestObserver mirrors the existing scope/depth budget contract while
// using the tree's conditional forecast to choose the next view. Hidden values
// can be accessed only through reader.Read on that chosen view.
func RunForestObserver(f *Forest, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if f == nil || reader == nil {
		return observation.Result{}, errors.New("missing observer inputs")
	}
	var mask, values, attempted uint16
	result := observation.Result{}
	result.Probability, _ = f.ForecastPartial(0, 0)
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
				expected := 0.
				for sub := added; ; sub = (sub - 1) & added {
					p, _ := f.ForecastPartial(mask|added, values|sub)
					expected += binaryEntropy(p) / float64(uint(1)<<cost)
					if sub == 0 {
						break
					}
				}
				gain := math.Max(0, binaryEntropy(result.Probability)-expected) / float64(cost)
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
		result.Probability, _ = f.ForecastPartial(mask, values)
		result.Trace = append(result.Trace, observation.Step{View: chosen, Attempted: attempted, Observed: mask, Values: values, Probability: result.Probability})
	}
	if result.Stop == "" {
		result.Stop = "budget"
	}
	result.Observed = bits.OnesCount16(mask)
	return result, nil
}
