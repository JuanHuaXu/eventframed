// grid-belief-experiment is a synthetic prequential experiment, not a serving
// benchmark. It prints all trajectory metrics so summaries are auditable.
package main

import (
	"encoding/json"
	"flag"
	"math"
	"math/rand"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

type metrics struct{ Brier, LogLoss, ComposedBrier float64 }
type trajectory struct {
	Scenario, Trajectory int
	Old, Grid, Beta      metrics
}
type summary struct {
	Scenario                                                     int
	Metric                                                       string
	Gain, Lower95, Upper95, LowerSimultaneous, UpperSimultaneous float64
}

func main() {
	seed := flag.Int64("seed", 2026090601, "frozen seed base")
	flag.Parse()
	var rows []trajectory
	old, grid := bayes.DefaultWorkingPolicy(), bayes.GridWorkingPolicy()
	for scenario := 0; scenario < 8; scenario++ {
		for rep := 0; rep < 64; rep++ {
			rng := rand.New(rand.NewSource(*seed + int64(scenario)*100000 + int64(rep)*1000))
			var a, b *model.WorkingBelief
			alpha, beta := 1., 1.
			r := trajectory{Scenario: scenario, Trajectory: rep}
			for n := 0; n < 1000; n++ {
				prob := []float64{.01, .2, .5, .8, .99, .99, .9, .1}[scenario]
				if scenario == 5 && n >= 500 {
					prob = .01
				}
				if scenario == 6 && (n/100)%2 == 1 {
					prob = .1
				}
				if scenario == 7 {
					prob = .1 + .8*float64(n)/999
				}
				// Predictions are computed before drawing or applying the outcome.
				pa := bayes.PredictiveMean(model.BayesianPosterior{WorkingBelief: a}, old)
				pb := bayes.PredictiveMean(model.BayesianPosterior{WorkingBelief: b}, grid)
				pc := alpha / (alpha + beta)
				y := rng.Float64() < prob
				add(&r.Old, pa, y)
				add(&r.Grid, pb, y)
				add(&r.Beta, pc, y)
				a = bayes.UpdateWorking(a, y, 1, false, old)
				b = bayes.UpdateWorking(b, y, 1, false, grid)
				if y {
					alpha++
				} else {
					beta++
				}
			}
			rows = append(rows, r)
		}
	}
	var sums []summary
	for s := 0; s < 8; s++ {
		for k, name := range []string{"Brier", "LogLoss", "ComposedBrier"} {
			var values []float64
			for _, r := range rows {
				if r.Scenario == s {
					a := []float64{r.Old.Brier, r.Old.LogLoss, r.Old.ComposedBrier}
					b := []float64{r.Grid.Brier, r.Grid.LogLoss, r.Grid.ComposedBrier}
					values = append(values, a[k]-b[k])
				}
			}
			mean := 0.
			for _, v := range values {
				mean += v / 64
			}
			variance := 0.
			for _, v := range values {
				variance += (v - mean) * (v - mean) / 63
			}
			se := math.Sqrt(variance / 64)
			sums = append(sums, summary{s, name, mean, mean - 1.96*se, mean + 1.96*se, mean - 3.08*se, mean + 3.08*se})
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Seed         int64
		Protocol     string
		Trajectories []trajectory
		Summaries    []summary
	}{*seed, "docs/grid-belief-protocol.md", rows, sums}); err != nil {
		panic(err)
	}
}
func add(m *metrics, p float64, y bool) {
	v := 0.
	likelihood := 1 - p
	if y {
		v = 1
		likelihood = p
	}
	m.Brier += (p - v) * (p - v) / 1000
	m.LogLoss -= math.Log(likelihood) / 1000
	composed := .45 + .1*p
	m.ComposedBrier += (composed - v) * (composed - v) / 1000
}
