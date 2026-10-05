package observationlearners

import (
	"fmt"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type ConditionalTick struct {
	Step, Audits int
	Mask, Values uint16
	Outcome      bool
	Predictions  [3]float64
}
type ConditionalRecord struct {
	Generator, Scenario, Split string
	Fit, Stream, BuildCount    int
	BuildNS                    int64
	Ticks                      []ConditionalTick
}

func conditionalOracle(generator int) [512]float64 {
	var weights [512]float64
	for x := range weights {
		switch generator {
		case 0:
			weights[x] = 1. / 512
		case 1:
			p := 1.
			for b := 0; b < 9; b++ {
				q := .2
				if b%2 == 1 {
					q = .8
				}
				if x&(1<<b) == 0 {
					q = 1 - q
				}
				p *= q
			}
			weights[x] = p
		case 2:
			p, q := .5, .5
			for b := 0; b < 9; b++ {
				if x&(1<<b) != 0 {
					p *= .1
					q *= .9
				} else {
					p *= .9
					q *= .1
				}
			}
			weights[x] = p + q
		}
	}
	return weights
}

// RunConditionalRecord deliberately reuses acquired masks. It isolates the
// integration calculation, not the adaptive consequences of new forecasts.
func RunConditionalRecord(r DependentRecord) (ConditionalRecord, error) {
	o := ConditionalRecord{Generator: r.Generator, Scenario: r.Scenario, Split: r.Split, Fit: r.Fit, Stream: r.Stream}
	g := -1
	for i, name := range []string{"fair", "biased", "clustered"} {
		if name == r.Generator {
			g = i
		}
	}
	j := -1
	for i, s := range Scenarios {
		if s.Name == r.Scenario {
			j = i
		}
	}
	seed := int64(2026104202)
	if r.Split == "confirmation" {
		seed++
	} else if r.Split != "design" {
		return o, fmt.Errorf("unknown split")
	}
	if g < 0 || j < 0 || len(r.Inputs) != 512 || len(r.Ticks) != 512 || len(r.Views) != 512 || len(r.Inner) != 512 {
		return o, fmt.Errorf("invalid diagnostic record")
	}
	var empirical, oracle *ConditionalForest
	var forest *Forest
	var audits []observation.Sample
	n := 0
	for t, tick := range r.Ticks {
		if forest != nil {
			trace := r.Views[t][2].Trace
			if len(trace) == 0 {
				return o, fmt.Errorf("missing view")
			}
			last := trace[len(trace)-1]
			p, e := forest.ForecastPartial(last.Observed, last.Values)
			if e != nil {
				return o, e
			}
			if math.Abs(p-r.Inner[t][2][1]) > 1e-12 {
				return o, fmt.Errorf("uniform forecast mismatch at %d", t)
			}
			q, e := empirical.Forecast(last.Observed, last.Values)
			if e != nil {
				return o, e
			}
			v, e := oracle.Forecast(last.Observed, last.Values)
			if e != nil {
				return o, e
			}
			o.Ticks = append(o.Ticks, ConditionalTick{t, n, last.Observed, last.Values, tick.Outcome, [3]float64{p, q, v}})
		}
		for _, origin := range tick.Delivered {
			if origin < 0 || origin > t {
				return o, fmt.Errorf("invalid delivery")
			}
			if !r.Ticks[origin].Audit {
				continue
			}
			n++
			audits = append(audits, observation.Sample{Bits: r.Inputs[origin], Outcome: r.Ticks[origin].Outcome})
			if len(audits) > 256 {
				audits = audits[1:]
			}
			if n < 32 || n%16 != 0 {
				continue
			}
			forest = NewForest(Seed(seed, 10*g+j, r.Fit, r.Stream, 4))
			for _, v := range audits[max(0, len(audits)-64):] {
				forest.Update(v.Bits, v.Outcome)
			}
			var weights [512]float64
			for x := range weights {
				weights[x] = 1. / 512
			}
			for _, v := range audits {
				weights[v.Bits]++
			}
			start := time.Now()
			var e error
			empirical, e = NewConditionalForest(forest, weights)
			if e != nil {
				return o, e
			}
			oracle, e = NewConditionalForest(forest, conditionalOracle(g))
			if e != nil {
				return o, e
			}
			o.BuildNS += int64(time.Since(start))
			o.BuildCount += 2
		}
	}
	if n != r.Audits || o.BuildCount != r.Fits {
		return o, fmt.Errorf("fit/audit mismatch")
	}
	return o, nil
}
