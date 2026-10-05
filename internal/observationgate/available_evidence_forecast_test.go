package observationgate

import (
	"fmt"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type availableEvidenceFrame struct {
	Origin                                                                                   int
	RequestedMask, RequestedValues, MonitorMask, MonitorValues, ConsumedMask, ConsumedValues uint16
	Original, Available                                                                      float64
	Experts                                                                                  [2][4]float64
	OuterWeights, InnerWeights                                                               [4]float64
}

// Read-only: the complete law is evaluated from the issue-time models and
// weights. This function cannot acquire coordinates or update any learner.
func availableEvidenceForecast(j *subsetDelayJournal, origin int, mask, values uint16) (float64, [4]float64, error) {
	var experts [4]float64
	if j == nil || origin < 0 || mask >= 512 || values&^mask != 0 {
		return 0, experts, fmt.Errorf("invalid available view")
	}
	e := j.entries[origin%subsetDelayCapacity]
	if !e.active || e.origin != origin || e.forecast.p.Version != j.models.Version {
		return 0, experts, fmt.Errorf("unmatched issued model")
	}
	long := j.models.Pooled
	if e.forecast.p.Split {
		long = j.models.Local
	}
	experts = [4]float64{.5, .5, .5, .5}
	for i, m := range [3]*observation.Model{j.state.base, j.models.Short, long} {
		if m != nil {
			p, err := m.ForecastObserved(mask, values)
			if err != nil {
				return 0, experts, err
			}
			experts[i] = p
		}
	}
	if e.forecast.available {
		if j.model == nil {
			return 0, experts, fmt.Errorf("missing issued subset model")
		}
		p, err := j.model.Forecast(mask, values)
		if err != nil {
			return 0, experts, err
		}
		inner := bayes.ForecastMix{Weights: e.forecast.innerWeights}
		experts[1] = inner.Forecast([4]float64{experts[1], p, p, p})
	}
	outer := bayes.ForecastMix{Weights: e.forecast.p.Weights}
	return outer.Forecast(experts), experts, nil
}

func availableEvidenceCapture(j *subsetDelayJournal, origin int, monitorMask, monitorValues uint16) (availableEvidenceFrame, error) {
	var r availableEvidenceFrame
	if j == nil || origin < 0 || monitorMask >= 512 || monitorValues&^monitorMask != 0 {
		return r, fmt.Errorf("invalid monitor view")
	}
	e := j.entries[origin%subsetDelayCapacity]
	if !e.active || e.origin != origin {
		return r, fmt.Errorf("missing issuance")
	}
	p := e.forecast.p
	if (p.Values^monitorValues)&p.Mask&monitorMask != 0 {
		return r, fmt.Errorf("conflicting paid coordinates")
	}
	r = availableEvidenceFrame{Origin: origin, RequestedMask: p.Mask, RequestedValues: p.Values, MonitorMask: monitorMask, MonitorValues: monitorValues, ConsumedMask: p.Mask | monitorMask, ConsumedValues: p.Values | monitorValues, OuterWeights: p.Weights, InnerWeights: e.forecast.innerWeights}
	var err error
	r.Original, r.Experts[0], err = availableEvidenceForecast(j, origin, p.Mask, p.Values)
	if err != nil {
		return r, err
	}
	if math.Abs(r.Original-p.P) > 1e-14 {
		return r, fmt.Errorf("original law mismatch")
	}
	for i, v := range r.Experts[0] {
		if math.Abs(v-p.Experts[i]) > 1e-14 {
			return r, fmt.Errorf("original expert mismatch")
		}
	}
	r.Available, r.Experts[1], err = availableEvidenceForecast(j, origin, r.ConsumedMask, r.ConsumedValues)
	return r, err
}
