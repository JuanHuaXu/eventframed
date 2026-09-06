// Synthetic prequential test of complete-law aggregation. No corpus or secrets.
package main

import (
	"encoding/json"
	"flag"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"math/rand"
	"os"
)

type metric struct{ Brier, LogLoss float64 }
type row struct {
	Scenario, Baseline, Trajectory int
	Old, Grid, Beta, Rescue        metric
}
type summary struct {
	Scenario, Baseline                                           int
	Control, Metric                                              string
	Gain, Lower95, Upper95, LowerSimultaneous, UpperSimultaneous float64
}

func main() {
	seed := flag.Int64("seed", 2026090611, "seed base")
	flag.Parse()
	var rows []row
	for base := 0; base < 4; base++ {
		for scenario := 0; scenario < 8; scenario++ {
			for rep := 0; rep < 64; rep++ {
				r := row{Scenario: scenario, Baseline: base, Trajectory: rep}
				rng := rand.New(rand.NewSource(*seed + int64(base)*10000000 + int64(scenario)*100000 + int64(rep)*1000))
				var old, grid *model.WorkingBelief
				var mix bayes.ForecastMix
				alpha, beta := 1., 1.
				for n := 0; n < 1000; n++ {
					p := []float64{.01, .2, .5, .8, .99, .99, .9, .1}[scenario]
					if scenario == 5 && n >= 500 {
						p = .01
					}
					if scenario == 6 && (n/100)%2 == 1 {
						p = .1
					}
					if scenario == 7 {
						p = .1 + .8*float64(n)/999
					}
					baseline := []float64{.2, .5, .8, .2}[base]
					if base == 3 && (n/50)%2 == 1 {
						baseline = .8
					}
					po := bayes.PredictiveMean(model.BayesianPosterior{WorkingBelief: old}, bayes.DefaultWorkingPolicy())
					pg := bayes.PredictiveMean(model.BayesianPosterior{WorkingBelief: grid}, bayes.GridWorkingPolicy())
					pb := alpha / (alpha + beta)
					experts := [4]float64{baseline, .9*baseline + .1*pg, pg, pb}
					pr := mix.Forecast(experts)
					y := rng.Float64() < p
					add(&r.Old, .9*baseline+.1*po, y)
					add(&r.Grid, experts[1], y)
					add(&r.Beta, pb, y)
					add(&r.Rescue, pr, y)
					mix = mix.Observe(experts, y, 1)
					old = bayes.UpdateWorking(old, y, 1, false, bayes.DefaultWorkingPolicy())
					grid = bayes.UpdateWorking(grid, y, 1, false, bayes.GridWorkingPolicy())
					if y {
						alpha++
					} else {
						beta++
					}
				}
				rows = append(rows, r)
			}
		}
	}
	var sums []summary
	for base := 0; base < 4; base++ {
		for s := 0; s < 8; s++ {
			for _, ctrl := range []string{"Old", "Grid"} {
				for _, met := range []string{"Brier", "LogLoss"} {
					var gains []float64
					for _, r := range rows {
						if r.Scenario != s || r.Baseline != base {
							continue
						}
						c := r.Old
						if ctrl == "Grid" {
							c = r.Grid
						}
						v := c.Brier - r.Rescue.Brier
						if met == "LogLoss" {
							v = c.LogLoss - r.Rescue.LogLoss
						}
						gains = append(gains, v)
					}
					mean := 0.
					for _, v := range gains {
						mean += v / 64
					}
					variance := 0.
					for _, v := range gains {
						variance += (v - mean) * (v - mean) / 63
					}
					se := math.Sqrt(variance / 64)
					sums = append(sums, summary{s, base, ctrl, met, mean, mean - 1.96*se, mean + 1.96*se, mean - 3.6*se, mean + 3.6*se})
				}
			}
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Seed         int64
		Trajectories []row
		Summaries    []summary
	}{*seed, rows, sums}); err != nil {
		panic(err)
	}
}
func add(m *metric, p float64, y bool) {
	v := 0.
	l := 1 - p
	if y {
		v = 1
		l = p
	}
	m.Brier += (p - v) * (p - v) / 1000
	m.LogLoss -= math.Log(l) / 1000
}
