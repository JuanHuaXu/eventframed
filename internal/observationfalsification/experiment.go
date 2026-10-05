// Package observationfalsification tests evidence selection, not production serving.
package observationfalsification

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	prev "github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

const FitSeed int64 = 2026092001
const DesignSeed int64 = 2026092002
const ConfirmationSeed int64 = 2026092003

var Arms = []string{"random", "single_mmm", "paired_mmm"}

func label(x uint16, local, null bool, r *rand.Rand) bool {
	if null {
		return r.Intn(2) == 1
	}
	y := (x&(1<<6) != 0) != (x&(1<<7) != 0)
	y = y != (x&(1<<8) != 0)
	if local {
		y = x&(1<<2) != 0
	}
	if r.Float64() < .05 {
		y = !y
	}
	return y
}
func Base(null bool) (*observation.Model, error) {
	r := rand.New(rand.NewSource(FitSeed))
	a := make([]observation.Sample, 4096)
	for i := range a {
		x := uint16(r.Intn(512))
		a[i] = observation.Sample{Bits: x, Outcome: label(x, false, null, r)}
	}
	return observation.Fit(a)
}
func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}
func divergence(p, q float64) float64 { return math.Max(0, entropy((p+q)/2)-(entropy(p)+entropy(q))/2) }

// Choice uses only committed forecasts and ex-ante random coins, never labels.
func Choice(arm string, forecasts [4][2]float64, explore bool, u float64) (int, float64) {
	var scores [4]float64
	for i, p := range forecasts {
		if arm == "single_mmm" {
			scores[i] = entropy(p[0])
		} else if arm == "paired_mmm" {
			scores[i] = divergence(p[0], p[1])
		}
	}
	best := scores[0]
	for _, v := range scores {
		best = math.Max(best, v)
	}
	var ties []int
	for i, v := range scores {
		if math.Abs(v-best) <= 1e-12 {
			ties = append(ties, i)
		}
	}
	i := min(3, int(u*4))
	if arm != "random" && !explore {
		i = ties[min(len(ties)-1, int(u*float64(len(ties))))]
	}
	prob := .25
	if arm != "random" {
		prob = .25 / 4
		for _, j := range ties {
			if i == j {
				prob += .75 / float64(len(ties))
			}
		}
	}
	return i, prob
}

type Probe struct {
	Forecasts                 [4][2]float64
	ObservedMasks             [4][2]uint16
	InspectionCosts           [4][2]int
	Choice                    int
	Probability               float64
	Bits                      uint16
	Outcome, ReferenceOutcome bool
}
type Tick struct {
	Outcome, ReferenceOutcome, Audit, Authorized bool
	Pool                                         [4]uint16
	Predictions                                  [3]prev.Prediction
	Probes                                       [3]Probe
	Nomination                                   string
	EvidenceMax                                  float64
}
type Metrics struct {
	N                        int
	Brier, LogLoss, Accuracy float64
	ConfidentErrors          int
}

func score(ticks []Tick, arm, start int) Metrics {
	m := Metrics{N: len(ticks) - start}
	for _, t := range ticks[start:] {
		p := t.Predictions[arm].P
		y := 0.
		if t.Outcome {
			y = 1
			m.LogLoss -= math.Log(p)
		} else {
			m.LogLoss -= math.Log1p(-p)
		}
		m.Brier += (p - y) * (p - y)
		if (p >= .5) == t.Outcome {
			m.Accuracy++
		} else if p <= .1 || p >= .9 {
			m.ConfidentErrors++
		}
	}
	m.Brier /= float64(m.N)
	m.LogLoss /= float64(m.N)
	m.Accuracy /= float64(m.N)
	return m
}

type Record struct {
	Split, Scenario                          string
	Index                                    int
	Ticks                                    []Tick `json:",omitempty"`
	Full, Post, Late                         [3]Metrics
	Foreground                               [3]int
	Monitor, ProbeCost, OutcomeQueries, Fits int
	FitNS                                    int64
	SplitAt                                  int
	Recovery                                 [3]int
}
type learner struct {
	state     *prev.State
	models    prev.Models
	live, ref []observation.Sample
	audits    int
}

func (l *learner) fit(a, b observation.Sample) error {
	l.live = append(l.live, a)
	l.ref = append(l.ref, b)
	if len(l.live) > 256 {
		l.live = l.live[1:]
		l.ref = l.ref[1:]
	}
	l.audits++
	if l.audits < 32 || (l.audits-32)%16 != 0 {
		return nil
	}
	n := len(l.live)
	var e error
	l.models.Short, e = observation.Fit(l.live[max(0, n-64):])
	if e != nil {
		return e
	}
	l.models.Local, e = observation.Fit(l.live)
	if e != nil {
		return e
	}
	pool := append([]observation.Sample{}, l.live[max(0, n-128):]...)
	pool = append(pool, l.ref[max(0, n-128):]...)
	l.models.Pooled, e = observation.Fit(pool)
	l.models.Version++
	return e
}
func observers(base *observation.Model, m prev.Models, p prev.Prediction) [2]*observation.Model {
	long := m.Pooled
	if p.Split {
		long = m.Local
	}
	models := [3]*observation.Model{base, m.Short, long}
	order := []int{0, 1, 2}
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			if (models[order[i]] == nil && models[order[j]] != nil) || (models[order[j]] != nil && p.Weights[order[j]] > p.Weights[order[i]]) {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	return [2]*observation.Model{models[order[0]], models[order[1]]}
}
func RunStream(base *observation.Model, split string, scenario, index int, seed int64) (Record, error) {
	name := prev.Scenarios[scenario]
	r := Record{Split: split, Scenario: name, Index: index, SplitAt: -1, Recovery: [3]int{-1, -1, -1}}
	var rng [14]*rand.Rand
	for i := range rng {
		rng[i] = rand.New(rand.NewSource(prev.Seed(seed, scenario, index, i)))
	}
	var learners [3]learner
	for i := range learners {
		s, e := prev.New(base, "mix_mmm_ap")
		if e != nil {
			return r, e
		}
		learners[i].state = s
	}
	var monitor observationrescue.Monitor
	var evidence prev.Evidence
	for step := 0; step < 512; step++ {
		t := Tick{Audit: rng[2].Float64() < .25}
		explore, u := rng[4].Float64() < .25, rng[5].Float64()
		x, rx := uint16(rng[0].Intn(512)), uint16(rng[1].Intn(512))
		rd := observationexperiment.Frames(x, fmt.Sprintf("v4-live-%s-%s-%d-%d", split, name, index, step))
		rr := observationexperiment.Frames(rx, fmt.Sprintf("v4-ref-%s-%s-%d-%d", split, name, index, step))
		old, e := observation.Run(base, rd, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		oldRef, e := observation.Run(base, rr, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		r.Monitor += old.Cost + oldRef.Cost
		for i := range learners {
			p, e := learners[i].state.Predict(rd, learners[i].models, step, 0)
			if e != nil {
				return r, e
			}
			t.Predictions[i] = p
			r.Foreground[i] += p.Cost
		}
		if t.Audit {
			r.ProbeCost += 36
			r.OutcomeQueries += 2
			for k := 0; k < 4; k++ {
				probeReader := observationexperiment.Frames(uint16(rng[3].Intn(512)), fmt.Sprintf("v4-probe-%s-%s-%d-%d-%d", split, name, index, step, k))
				var bits, mask uint16
				for scope := 0; scope < 3; scope++ {
					m, v, e := probeReader.Read(observation.View{Scope: scope, Depth: 2})
					if e != nil {
						return r, e
					}
					mask |= m
					bits |= v
				}
				if mask != 511 {
					return r, fmt.Errorf("incomplete probe")
				}
				t.Pool[k] = bits
				for i := range learners {
					models := observers(base, learners[i].models, t.Predictions[i])
					for j, m := range models {
						p := .5
						if m != nil {
							result, e := observation.Run(m, probeReader, 1, "mmm", 0)
							if e != nil {
								return r, e
							}
							p = result.Probability
							t.Probes[i].ObservedMasks[k][j] = result.Trace[len(result.Trace)-1].Observed
							t.Probes[i].InspectionCosts[k][j] = result.Cost
						}
						t.Probes[i].Forecasts[k][j] = p
					}
				}
			}
			for i, arm := range Arms {
				choice, prob := Choice(arm, t.Probes[i].Forecasts, explore, u)
				t.Probes[i].Choice = choice
				t.Probes[i].Probability = prob
				t.Probes[i].Bits = t.Pool[choice]
			}
		}
		// Outcomes become available only after every live forecast and choice.
		local := (name == "member_shift" || name == "common_shift") && step >= 256
		if name == "recurring" {
			local = (step/128)%2 == 1
		}
		t.Outcome = label(x, local, name == "null", rng[0])
		t.ReferenceOutcome = label(rx, name == "common_shift" && step >= 256, name == "null", rng[1])
		rc, lc := (oldRef.Probability >= .5) == t.ReferenceOutcome, (old.Probability >= .5) == t.Outcome
		rev, _ := monitor.Observe(rc, lc, true)
		valid := evidence.Observe(rc, lc)
		t.Authorized = valid && bayes.RevisionSplits(rev.Action)
		t.Nomination = string(rev.Action)
		t.EvidenceMax = evidence.Max
		if t.Authorized && r.SplitAt < 0 {
			r.SplitAt = step
		}
		var py, pr [4]bool
		if t.Audit {
			for k, bits := range t.Pool {
				py[k] = label(bits, local, name == "null", rng[6+2*k])
				pr[k] = label(bits, name == "common_shift" && step >= 256, name == "null", rng[7+2*k])
			}
		}
		for i := range learners {
			l := &learners[i]
			if e := l.state.Observe(step, t.Outcome, t.Authorized); e != nil {
				return r, e
			}
			if t.Audit {
				p := &t.Probes[i]
				p.Outcome, p.ReferenceOutcome = py[p.Choice], pr[p.Choice]
				version := l.models.Version
				start := time.Now()
				if e := l.fit(observation.Sample{Bits: p.Bits, Outcome: p.Outcome}, observation.Sample{Bits: p.Bits, Outcome: p.ReferenceOutcome}); e != nil {
					return r, e
				}
				r.FitNS += int64(time.Since(start))
				if l.models.Version != version {
					r.Fits += 3
				}
			}
		}
		r.Ticks = append(r.Ticks, t)
	}
	post := 256
	if name == "recurring" {
		post = 128
	}
	for i := range learners {
		r.Full[i] = score(r.Ticks, i, 0)
		r.Post[i] = score(r.Ticks, i, post)
		r.Late[i] = score(r.Ticks, i, 384)
		if name == "member_shift" || name == "common_shift" {
			for end := 319; end < 512; end++ {
				correct := 0
				for j := end - 63; j <= end; j++ {
					if (r.Ticks[j].Predictions[i].P >= .5) == r.Ticks[j].Outcome {
						correct++
					}
				}
				if correct >= 52 {
					r.Recovery[i] = end - 256
					break
				}
			}
		}
	}
	return r, nil
}

type Comparison struct {
	Split, Scenario, Control, Window string
	Gain, Lower, Upper               float64
}
type Output struct {
	Protocol                           string
	Hashes                             map[string]string
	Seeds                              [3]int64
	Records                            []Record
	Comparisons                        []Comparison
	ShiftPass, StablePass, OverallPass bool
}

func Run(progress func(string)) (Output, error) {
	out := Output{Protocol: "docs/experiments/mmm-falsification-v4-protocol.md", Seeds: [3]int64{FitSeed, DesignSeed, ConfirmationSeed}}
	for k, split := range []string{"design", "confirmation"} {
		seed := DesignSeed
		if k == 1 {
			seed = ConfirmationSeed
		}
		for j, name := range prev.Scenarios {
			base, e := Base(name == "null")
			if e != nil {
				return out, e
			}
			for i := 0; i < 32; i++ {
				r, e := RunStream(base, split, j, i, seed)
				if e != nil {
					return out, e
				}
				out.Records = append(out.Records, r)
			}
			if progress != nil {
				progress(split + " " + name + " complete")
			}
		}
	}
	Summarize(&out)
	return out, nil
}
func Summarize(out *Output) {
	out.Comparisons = nil
	out.ShiftPass = true
	out.StablePass = true
	for _, split := range []string{"design", "confirmation"} {
		for _, name := range prev.Scenarios {
			for control := 0; control < 2; control++ {
				for _, window := range []string{"full", "post", "late"} {
					var g []float64
					for _, r := range out.Records {
						if r.Split != split || r.Scenario != name {
							continue
						}
						a := r.Full
						if window == "post" {
							a = r.Post
						}
						if window == "late" {
							a = r.Late
						}
						g = append(g, a[control].Brier-a[2].Brier)
					}
					mean := 0.
					for _, v := range g {
						mean += v / 32
					}
					variance := 0.
					for _, v := range g {
						variance += (v - mean) * (v - mean) / 31
					}
					margin := 3.5 * math.Sqrt(variance/32)
					c := Comparison{split, name, Arms[control], window, mean, mean - margin, mean + margin}
					out.Comparisons = append(out.Comparisons, c)
					if split == "confirmation" {
						if window == "post" && (name == "member_shift" || name == "common_shift") {
							out.ShiftPass = out.ShiftPass && c.Gain >= .005 && c.Lower > 0
						}
						if name == "stable" && window == "full" && control == 0 {
							out.StablePass = out.StablePass && -c.Lower <= .01
						}
					}
				}
			}
		}
	}
	n := 0
	for _, r := range out.Records {
		if r.Split == "confirmation" && r.Scenario == "stable" && r.SplitAt >= 0 {
			n++
		}
	}
	out.StablePass = out.StablePass && float64(n)/32 <= .05
	out.OverallPass = out.ShiftPass && out.StablePass
}
