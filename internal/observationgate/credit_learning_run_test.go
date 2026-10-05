package observationgate

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func creditLearningRun(base *observation.Model, cfg forestDependenceCase, ci, index int, seed int64, delayed bool, outcome func(uint16, int, bool, *rand.Rand) bool, useCredit bool) (innerArrivalResult, []monitorCreditAcquisition, error) {
	var acquisition []monitorCreditAcquisition
	credit := 0
	r := innerArrivalResult{Arms: [4]integrationArm{{Arm: "original", SplitAt: -1}, {Arm: "coupled_outer", SplitAt: -1}, {Arm: "coupled_full", SplitAt: -1}, {Arm: "fixed_view", SplitAt: -1}}}
	var rng [8]*rand.Rand
	for role := range rng {
		rng[role] = rand.New(rand.NewSource(observationpreserved.Seed(seed, ci, index, role)))
	}
	js := [3]*subsetDelayJournal{{state: subsetState{base: base, enabled: true}}, {state: subsetState{base: base, enabled: true}}, {state: subsetState{base: base, enabled: true}}}
	var models observationpreserved.Models
	var coupledMix, fixedMix, fullMix, fullInner bayes.ForecastMix
	splitClock := -1
	for _, j := range js {
		if err := j.publish(models, nil); err != nil {
			return r, acquisition, err
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
			var rd, rr observation.Reader = observationexperiment.Frames(x, "delayed-live"), observationexperiment.Frames(rx, "delayed-ref")
			var cached [2]*monitorCreditReader
			if useCredit {
				cached = [2]*monitorCreditReader{newMonitorCreditReader(rd), newMonitorCreditReader(rr)}
				rd, rr = cached[0], cached[1]
			}
			bp, err := observation.Run(base, rd, 1, "mmm", 0)
			if err != nil {
				return r, acquisition, err
			}
			rp, err := observation.Run(base, rr, 1, "mmm", 0)
			if err != nil {
				return r, acquisition, err
			}
			if useCredit {
				bounded := bp.Cost + rp.Cost
				full, charge, next, e := monitorCreditStep(credit, bounded, audit)
				if e != nil {
					return r, acquisition, e
				}
				if full {
					for _, reader := range cached {
						for scope := 0; scope < 3; scope++ {
							if _, _, e := reader.Read(observation.View{Scope: scope, Depth: 2}); e != nil {
								return r, acquisition, e
							}
						}
						if reader.mask != 511 {
							return r, acquisition, fmt.Errorf("incomplete credited view")
						}
					}
					bp.Probability, e = base.ForecastObserved(cached[0].mask, cached[0].values)
					if e != nil {
						return r, acquisition, e
					}
					rp.Probability, e = base.ForecastObserved(cached[1].mask, cached[1].values)
					if e != nil {
						return r, acquisition, e
					}
				}
				if cached[0].charged+cached[1].charged != charge {
					return r, acquisition, fmt.Errorf("credit read accounting")
				}
				allowance := bounded
				if audit {
					allowance += 18
				}
				credit = next
				acquisition = append(acquisition, monitorCreditAcquisition{bounded, allowance, charge, credit, audit, full})
				r.MonitorCost += charge
				if audit {
					r.MonitorCost -= 18
				}
			} else {
				r.MonitorCost += bp.Cost + rp.Cost
			}
			f := innerArrivalFrame{X: x, RX: rx, Audit: audit, Arrival: clock, Seed: rng[3].Int63(), Baseline: bp.Probability, Reference: rp.Probability}
			for a, j := range js {
				if a == 1 {
					j.state.mix = coupledMix
					j.state.inner = js[0].state.inner
				}
				if a == 2 {
					j.state.mix = fullMix
					j.state.inner = fullInner
				}
				p, err := j.predict(clock, rd, f.Seed)
				if err != nil {
					return r, acquisition, err
				}
				f.Predictions[a] = p
			}
			f.Predictions[3] = f.Predictions[0]
			f.Predictions[3].P = fixedMix.Forecast(f.Predictions[0].Experts)
			f.Predictions[3].Weights = fixedMix.Weights
			if audit {
				for _, reader := range []observation.Reader{rd, rr} {
					var mask uint16
					for scope := 0; scope < 3; scope++ {
						m, _, err := reader.Read(observation.View{Scope: scope, Depth: 2})
						if err != nil {
							return r, acquisition, err
						}
						mask |= m
					}
					if mask != 511 {
						return r, acquisition, fmt.Errorf("incomplete delayed audit")
					}
				}
				r.AuditCost += 18
			}
			local := (cfg.Mode == "member_shift" || cfg.Mode == "common_shift") && clock >= 256
			if cfg.Mode == "recurring" {
				local = (clock/128)%2 == 1
			}
			if outcome == nil {
				f.Y = breadthLabel(x, local, cfg.Mode == "null", rng[0], cfg.Target)
				f.RY = breadthLabel(rx, cfg.Mode == "common_shift" && clock >= 256, cfg.Mode == "null", rng[1], cfg.Target)
			} else {
				f.Y = outcome(x, clock, false, rng[0])
				f.RY = outcome(rx, clock, true, rng[1])
			}
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
				if splitClock < 0 || origin > splitClock {
					coupledMix = coupledMix.Observe(f.Predictions[1].Experts, f.Y, 1)
					fullMix = fullMix.Observe(f.Predictions[2].Experts, f.Y, 1)
					e := js[2].entries[origin%subsetDelayCapacity]
					if !e.active || e.origin != origin {
						return r, acquisition, fmt.Errorf("missing issued inner advice")
					}
					if e.forecast.available {
						fullInner = fullInner.Observe(e.forecast.inner, f.Y, 1)
					}
					fixedMix = fixedMix.Observe(f.Predictions[0].Experts, f.Y, 1)
				}
				for _, j := range js {
					if err := j.deliver(origin, f.Y, false); err != nil {
						return r, acquisition, err
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
					return r, acquisition, err
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
					return r, acquisition, err
				}
				if allow && bayes.RevisionSplits(revision.Action) {
					if splitClock < 0 {
						splitClock = clock
						r.Arms[3].SplitAt = clock
						coupledMix = coupledArrivalRevoke(coupledMix)
						fullMix = coupledArrivalRevoke(fullMix)
						fixedMix = coupledArrivalRevoke(fixedMix)
					}
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
					return r, acquisition, fmt.Errorf("future fit")
				}
				ls[i] = observation.Sample{Bits: f.X, Outcome: f.Y}
				rs[i] = observation.Sample{Bits: f.RX, Outcome: f.RY}
			}
			n := len(ls)
			var err error
			models.Short, err = observation.Fit(ls[max(0, n-64):])
			if err != nil {
				return r, acquisition, err
			}
			models.Local, err = observation.Fit(ls)
			if err != nil {
				return r, acquisition, err
			}
			pooled := append(append([]observation.Sample{}, ls[max(0, n-128):]...), rs[max(0, n-128):]...)
			models.Pooled, err = observation.Fit(pooled)
			if err != nil {
				return r, acquisition, err
			}
			models.Version++
			for _, j := range js {
				trial := newSubsetTrial(base, 1)
				err = trial.fit(ls[max(0, n-64):])
				if err != nil {
					return r, acquisition, err
				}
				if err = j.publish(models, trial.model); err != nil {
					return r, acquisition, err
				}
			}
			r.Fits = append(r.Fits, forestDelayFit{Clock: clock, Origins: append([]int{}, ids...)})
		}
	}
	for a, j := range js {
		r.Applied[a], r.Stale[a], r.Censored[a] = j.applied, j.stale, j.censored
		if j.head != 512 || j.applied+j.stale+j.censored != 512 {
			return r, acquisition, fmt.Errorf("unsettled delayed journal")
		}
	}
	return r, acquisition, nil
}
