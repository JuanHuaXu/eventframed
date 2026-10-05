package observationlearners

import (
	"fmt"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

type AlternateTick struct {
	Step, Audits int
	Views        [4]observation.Result
	Outcome      bool
	Predictions  [4]float64
}
type AlternateRecord struct {
	Generator, Scenario, Split string
	Fit, Stream, BuildCount    int
	BuildNS                    int64
	Ticks                      []AlternateTick
}

// RunAlternateRecord lets each expert acquire its own bounded view. The full
// input is accessible through a reader only, and labels arrive after forecasts.
func RunAlternateRecord(r DependentRecord) (AlternateRecord, error) {
	o := AlternateRecord{Generator: r.Generator, Scenario: r.Scenario, Split: r.Split, Fit: r.Fit, Stream: r.Stream}
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
	var short *observation.Model
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

			rd := observationexperiment.Frames(r.Inputs[t], "alternate-diagnostic")
			var views [4]observation.Result
			views[0], e = observation.Run(short, rd, rd.Epoch(), "mmm", 0)
			if e != nil {
				return o, e
			}
			views[1], e = RunForestObserver(forest, rd, rd.Epoch())
			if e != nil {
				return o, e
			}
			views[2], e = RunConditionalObserver(empirical, rd, rd.Epoch())
			if e != nil {
				return o, e
			}
			views[3], e = RunConditionalObserver(oracle, rd, rd.Epoch())
			if e != nil {
				return o, e
			}
			var predictions [4]float64
			for arm, view := range views {
				predictions[arm] = view.Probability
			}
			o.Ticks = append(o.Ticks, AlternateTick{t, n, views, tick.Outcome, predictions})

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
			short, e = observation.Fit(audits[max(0, len(audits)-64):])
			if e != nil {
				return o, e
			}
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
