package observationgate

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

type forestDelayFrame struct {
	X, RX                 uint16
	Y, RY, Audit, Missing bool
	Arrival               int
	Seed                  int64
	Baseline, Reference   float64
	Predictions           [2]observationpreserved.Prediction
}
type forestDelayFit struct {
	Clock   int
	Origins []int
}
type forestDelayResult struct {
	Arms                     [2]integrationArm
	Frames                   []forestDelayFrame
	Fits                     []forestDelayFit
	Applied, Stale, Censored [2]int
	MonitorCost, AuditCost   int
	Released                 []int
}

// External Anti-Pigeon authority acts on its own fixed-baseline evidence,
// independently of whether a fitted selector's old forecast is still current.
// No synthetic label or selector update is used to apply a structural split.
func forestDelayRevoke(j *subsetDelayJournal) {
	if j.state.split {
		return
	}
	j.state.split = true
	if j.state.mix.Weights == [4]float64{} {
		j.state.mix.Weights = [4]float64{.7, .1, .1, .1}
	}
	j.state.mix.Weights[2] = math.Min(.1, j.state.mix.Weights[2])
	sum := 0.
	for _, w := range j.state.mix.Weights {
		sum += w
	}
	for i := range j.state.mix.Weights {
		j.state.mix.Weights[i] /= sum
	}
	j.generation++
}

// Models fit only arrived audit labels. Gate updates wait for the origin prefix;
// missing origins expire at age32. The evaluator retains all hidden labels but
// the learning transition receives them only through the delivery branch.
func forestDelayRun(base *observation.Model, cfg forestDependenceCase, ci, index int, seed int64, delayed bool) (forestDelayResult, error) {
	r := forestDelayResult{Arms: [2]integrationArm{{Arm: "uniform", SplitAt: -1}, {Arm: "forest", SplitAt: -1}}}
	var rng [8]*rand.Rand
	for role := range rng {
		rng[role] = rand.New(rand.NewSource(observationpreserved.Seed(seed, ci, index, role)))
	}
	js := [2]*subsetDelayJournal{{state: subsetState{base: base, enabled: true}}, {state: subsetState{base: base, enabled: true}}}
	var models observationpreserved.Models
	for _, j := range js {
		if err := j.publish(models, nil); err != nil {
			return r, err
		}
	}
	var gate pairedInvestigator
	var monitor observationrescue.Monitor
	var arrived, expired [512]bool
	var origins []int
	head := 0
	for clock := 0; clock < 544; clock++ {
		if clock < 512 {
			audit := rng[2].Float64() < .25
			x := forestDependenceInput(uint16(rng[0].Intn(512)), cfg.Pattern, clock, rng[5])
			rx := forestDependenceInput(uint16(rng[1].Intn(512)), cfg.Pattern, clock, rng[6])
			rd, rr := observationexperiment.Frames(x, "delayed-live"), observationexperiment.Frames(rx, "delayed-ref")
			bp, err := observation.Run(base, rd, 1, "mmm", 0)
			if err != nil {
				return r, err
			}
			rp, err := observation.Run(base, rr, 1, "mmm", 0)
			if err != nil {
				return r, err
			}
			r.MonitorCost += bp.Cost + rp.Cost
			f := forestDelayFrame{X: x, RX: rx, Audit: audit, Arrival: clock, Seed: rng[3].Int63(), Baseline: bp.Probability, Reference: rp.Probability}
			for a, j := range js {
				p, err := j.predict(clock, rd, f.Seed)
				if err != nil {
					return r, err
				}
				f.Predictions[a] = p
			}
			if audit {
				for _, reader := range []observation.Reader{rd, rr} {
					var mask uint16
					for scope := 0; scope < 3; scope++ {
						m, _, err := reader.Read(observation.View{Scope: scope, Depth: 2})
						if err != nil {
							return r, err
						}
						mask |= m
					}
					if mask != 511 {
						return r, fmt.Errorf("incomplete delayed audit")
					}
				}
				r.AuditCost += 18
			}
			local := (cfg.Mode == "member_shift" || cfg.Mode == "common_shift") && clock >= 256
			if cfg.Mode == "recurring" {
				local = (clock/128)%2 == 1
			}
			f.Y = breadthLabel(x, local, cfg.Mode == "null", rng[0], cfg.Target)
			f.RY = breadthLabel(rx, cfg.Mode == "common_shift" && clock >= 256, cfg.Mode == "null", rng[1], cfg.Target)
			if delayed {
				f.Arrival += rng[7].Intn(32)
				f.Missing = rng[7].Float64() < .2
			}
			for a, p := range f.Predictions {
				r.Arms[a].Full.add(p, f.Y)
				if clock >= 256 {
					r.Arms[a].Post.add(p, f.Y)
				}
			}
			r.Frames = append(r.Frames, f)
		}
		// Stable origin ordering resolves simultaneous arrivals reproducibly.
		var fitNeeded bool
		for origin, f := range r.Frames {
			if !arrived[origin] && !expired[origin] && !f.Missing && f.Arrival == clock {
				arrived[origin] = true
				for _, j := range js {
					if err := j.deliver(origin, f.Y, false); err != nil {
						return r, err
					}
				}
				if f.Audit {
					origins = append(origins, origin)
					sort.Ints(origins)
					if len(origins) >= 32 && (len(origins)-32)%16 == 0 {
						fitNeeded = true
					}
				}
			}
		}
		cutoff := max(0, clock-31)
		for origin := head; origin < min(cutoff, len(r.Frames)); origin++ {
			if !arrived[origin] {
				expired[origin] = true
			}
		}
		for _, j := range js {
			if cutoff > j.head {
				if err := j.expireBefore(min(cutoff, len(r.Frames))); err != nil {
					return r, err
				}
			}
		}
		for head < len(r.Frames) && (arrived[head] || expired[head]) {
			if arrived[head] {
				f := r.Frames[head]
				rc, lc := (f.Reference >= .5) == f.RY, (f.Baseline >= .5) == f.Y
				revision, _ := monitor.Observe(rc, lc, true)
				allow, err := gate.observe(gate.next, rc, lc)
				if err != nil {
					return r, err
				}
				if allow && bayes.RevisionSplits(revision.Action) {
					for a, j := range js {
						if !j.state.split {
							forestDelayRevoke(j)
							r.Arms[a].SplitAt = clock
						}
					}
				}
				r.Released = append(r.Released, head)
			}
			head++
		}
		if fitNeeded {
			ids := origins[max(0, len(origins)-256):]
			ls, rs := make([]observation.Sample, len(ids)), make([]observation.Sample, len(ids))
			for i, id := range ids {
				f := r.Frames[id]
				if f.Missing || f.Arrival > clock {
					return r, fmt.Errorf("future fit")
				}
				ls[i] = observation.Sample{Bits: f.X, Outcome: f.Y}
				rs[i] = observation.Sample{Bits: f.RX, Outcome: f.RY}
			}
			n := len(ls)
			var err error
			models.Short, err = observation.Fit(ls[max(0, n-64):])
			if err != nil {
				return r, err
			}
			models.Local, err = observation.Fit(ls)
			if err != nil {
				return r, err
			}
			pooled := append(append([]observation.Sample{}, ls[max(0, n-128):]...), rs[max(0, n-128):]...)
			models.Pooled, err = observation.Fit(pooled)
			if err != nil {
				return r, err
			}
			models.Version++
			for a, j := range js {
				trial := newSubsetTrial(base, 1)
				if a == 0 {
					err = trial.fit(ls[max(0, n-64):])
				} else {
					err = forestInputFit(trial, ls[max(0, n-64):])
				}
				if err != nil {
					return r, err
				}
				if err = j.publish(models, trial.model); err != nil {
					return r, err
				}
			}
			r.Fits = append(r.Fits, forestDelayFit{Clock: clock, Origins: append([]int{}, ids...)})
		}
	}
	for a, j := range js {
		r.Applied[a], r.Stale[a], r.Censored[a] = j.applied, j.stale, j.censored
		if j.head != 512 || j.applied+j.stale+j.censored != 512 {
			return r, fmt.Errorf("unsettled delayed journal")
		}
	}
	return r, nil
}
