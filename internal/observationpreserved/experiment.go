package observationpreserved

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

const FitSeed int64 = 2026091701
const DesignSeed int64 = 2026091702
const ConfirmationSeed int64 = 2026091703
const Streams = 32
const Steps = 512

var Scenarios = []string{"stable", "member_shift", "common_shift", "recurring", "null"}

func Seed(base int64, scenario, stream, role int) int64 {
	return base*1000000 + int64(scenario)*100000 + int64(stream)*100 + int64(role)
}
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
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(r.Intn(512))
		samples[i] = observation.Sample{Bits: x, Outcome: label(x, false, null, r)}
	}
	return observation.Fit(samples)
}

type Metrics struct {
	N                        int `json:"n"`
	Brier, LogLoss, Accuracy float64
	ConfidentErrors          int
}

func metrics(p []Prediction, y []bool, start int) Metrics {
	m := Metrics{N: len(y) - start}
	for i := start; i < len(y); i++ {
		v := p[i].P
		truth := 0.
		if y[i] {
			truth = 1
			m.LogLoss -= math.Log(v)
		} else {
			m.LogLoss -= math.Log1p(-v)
		}
		m.Brier += (v - truth) * (v - truth)
		if (v >= .5) == y[i] {
			m.Accuracy++
		} else if v <= .1 || v >= .9 {
			m.ConfidentErrors++
		}
	}
	m.Brier /= float64(m.N)
	m.LogLoss /= float64(m.N)
	m.Accuracy /= float64(m.N)
	return m
}

type ArmRecord struct {
	Arm         string
	Predictions []Prediction `json:",omitempty"`
	Full, Post  Metrics
	Foreground  int
	SplitAt     int
}
type Tick struct {
	Outcome, ReferenceOutcome, Audit bool
	Nomination                       string
	ChangePoint, Authorized          bool
	EvidenceMax                      float64
	Fit                              bool
}
type Record struct {
	Split, Scenario              string
	Index                        int
	Ticks                        []Tick `json:",omitempty"`
	Arms                         []ArmRecord
	MonitorCost, AuditCost, Fits int
	FitNS                        int64
}

func RunStream(base *observation.Model, split string, scenario, index int, splitSeed int64) (Record, error) {
	name := Scenarios[scenario]
	r := Record{Split: split, Scenario: name, Index: index}
	live := rand.New(rand.NewSource(Seed(splitSeed, scenario, index, 0)))
	ref := rand.New(rand.NewSource(Seed(splitSeed, scenario, index, 1)))
	audit := rand.New(rand.NewSource(Seed(splitSeed, scenario, index, 2)))
	random := rand.New(rand.NewSource(Seed(splitSeed, scenario, index, 3)))
	states := make([]*State, len(Arms))
	for j, a := range Arms {
		s, e := New(base, a)
		if e != nil {
			return r, e
		}
		states[j] = s
		r.Arms = append(r.Arms, ArmRecord{Arm: a, SplitAt: -1})
	}
	var models Models
	var monitor observationrescue.Monitor
	var evidence Evidence
	var liveSamples, refSamples []observation.Sample
	var outcomes []bool
	audits := 0
	for step := 0; step < Steps; step++ {
		auditNow := audit.Float64() < .25
		x, rx := uint16(live.Intn(512)), uint16(ref.Intn(512))
		reader := observationexperiment.Frames(x, fmt.Sprintf("v3-live-%s-%s-%d-%d", split, name, index, step))
		reference := observationexperiment.Frames(rx, fmt.Sprintf("v3-ref-%s-%s-%d-%d", split, name, index, step))
		old, err := observation.Run(base, reader, 1, "mmm", 0)
		if err != nil {
			return r, err
		}
		oldRef, err := observation.Run(base, reference, 1, "mmm", 0)
		if err != nil {
			return r, err
		}
		r.MonitorCost += old.Cost + oldRef.Cost
		randSeed := random.Int63()
		for j, s := range states {
			p, e := s.Predict(reader, models, step, randSeed)
			if e != nil {
				return r, e
			}
			r.Arms[j].Predictions = append(r.Arms[j].Predictions, p)
			r.Arms[j].Foreground += p.Cost
		}
		// Shadow observation takes place after all forecasts and before labels.
		var a, b observation.Sample
		if auditNow {
			r.AuditCost += 18
			for j, rd := range []observation.Reader{reader, reference} {
				var mask, values uint16
				for scope := 0; scope < 3; scope++ {
					m, v, e := rd.Read(observation.View{Scope: scope, Depth: 2})
					if e != nil {
						return r, e
					}
					mask |= m
					values |= v
				}
				if mask != 511 {
					return r, fmt.Errorf("incomplete synthetic audit")
				}
				if j == 0 {
					a.Bits = values
				} else {
					b.Bits = values
				}
			}
		}
		local := (name == "member_shift" || name == "common_shift") && step >= 256
		if name == "recurring" {
			local = (step/128)%2 == 1
		}
		y, ry := label(x, local, name == "null", live), label(rx, name == "common_shift" && step >= 256, name == "null", ref)
		rc, lc := (oldRef.Probability >= .5) == ry, (old.Probability >= .5) == y
		revision, cp := monitor.Observe(rc, lc, true)
		valid := evidence.Observe(rc, lc)
		authorized := valid && bayes.RevisionSplits(revision.Action)
		tick := Tick{Outcome: y, ReferenceOutcome: ry, Audit: auditNow, Nomination: string(revision.Action), ChangePoint: cp, Authorized: authorized, EvidenceMax: evidence.Max}
		for j, s := range states {
			if e := s.Observe(step, y, authorized); e != nil {
				return r, e
			}
			if s.split && r.Arms[j].SplitAt < 0 {
				r.Arms[j].SplitAt = step
			}
		}
		if auditNow {
			a.Outcome = y
			b.Outcome = ry
			liveSamples = append(liveSamples, a)
			refSamples = append(refSamples, b)
			audits++
			if len(liveSamples) > 256 {
				liveSamples = liveSamples[1:]
				refSamples = refSamples[1:]
			}
			if audits >= 32 && (audits-32)%16 == 0 {
				start := time.Now()
				n := len(liveSamples)
				models.Short, err = observation.Fit(liveSamples[max(0, n-64):])
				if err != nil {
					return r, err
				}
				models.Local, err = observation.Fit(liveSamples)
				if err != nil {
					return r, err
				}
				pooled := append([]observation.Sample{}, liveSamples[max(0, n-128):]...)
				pooled = append(pooled, refSamples[max(0, n-128):]...)
				models.Pooled, err = observation.Fit(pooled)
				if err != nil {
					return r, err
				}
				models.Version++
				r.Fits += 3
				r.FitNS += int64(time.Since(start))
				tick.Fit = true
			}
		}
		outcomes = append(outcomes, y)
		r.Ticks = append(r.Ticks, tick)
	}
	post := 256
	if name == "recurring" {
		post = 128
	}
	for j := range r.Arms {
		a := &r.Arms[j]
		a.Full = metrics(a.Predictions, outcomes, 0)
		a.Post = metrics(a.Predictions, outcomes, post)
	}
	return r, nil
}

type Summary struct {
	Split, Scenario, Arm                                                                    string
	FullBrier, PostBrier, PostLogLoss, PostAccuracy, MeanForeground, MeanMonitor, MeanAudit float64
	ConfidentErrors, Splits, EarlySplits                                                    int
	MeanDelay                                                                               *float64
	SplitWilson                                                                             [2]float64
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
	Summaries                          []Summary
	Comparisons                        []Comparison
	ShiftPass, StablePass, OverallPass bool
}

func Run(progress func(string)) (Output, error) {
	out := Output{Protocol: "docs/experiments/mmm-preserved-v3-protocol.md", Seeds: [3]int64{FitSeed, DesignSeed, ConfirmationSeed}}
	for k, split := range []string{"design", "confirmation"} {
		seed := DesignSeed
		if k == 1 {
			seed = ConfirmationSeed
		}
		for scenario, name := range Scenarios {
			base, e := Base(name == "null")
			if e != nil {
				return out, e
			}
			for i := 0; i < Streams; i++ {
				r, e := RunStream(base, split, scenario, i, seed)
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
func wilson(k int) [2]float64 {
	n := float64(Streams)
	p := float64(k) / n
	z := 1.95996398454
	d := 1 + z*z/n
	c := (p + z*z/(2*n)) / d
	r := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / d
	return [2]float64{math.Max(0, c-r), math.Min(1, c+r)}
}
func Summarize(out *Output) {
	out.Summaries = nil
	out.Comparisons = nil
	out.ShiftPass = true
	out.StablePass = true
	for _, split := range []string{"design", "confirmation"} {
		for _, name := range Scenarios {
			for j, arm := range Arms {
				s := Summary{Split: split, Scenario: name, Arm: arm}
				delay, detected := 0., 0
				for _, r := range out.Records {
					if r.Split != split || r.Scenario != name {
						continue
					}
					a := r.Arms[j]
					s.FullBrier += a.Full.Brier / Streams
					s.PostBrier += a.Post.Brier / Streams
					s.PostLogLoss += a.Post.LogLoss / Streams
					s.PostAccuracy += a.Post.Accuracy / Streams
					s.ConfidentErrors += a.Post.ConfidentErrors
					s.MeanForeground += float64(a.Foreground) / (Streams * Steps)
					s.MeanMonitor += float64(r.MonitorCost) / (Streams * Steps)
					s.MeanAudit += float64(r.AuditCost) / (Streams * Steps)
					if a.SplitAt >= 0 {
						s.Splits++
						change := 256
						if name == "recurring" {
							change = 128
						}
						if name != "stable" && name != "null" {
							if a.SplitAt < change {
								s.EarlySplits++
							} else {
								delay += float64(a.SplitAt - change)
								detected++
							}
						}
					}
				}
				if detected > 0 {
					v := delay / float64(detected)
					s.MeanDelay = &v
				}
				s.SplitWilson = wilson(s.Splits)
				out.Summaries = append(out.Summaries, s)
				if split == "confirmation" && name == "stable" && arm == "mix_mmm_ap" {
					out.StablePass = out.StablePass && float64(s.Splits)/Streams <= .05
				}
			}
			for control := 0; control < 5; control++ {
				for _, window := range []string{"full", "post"} {
					var gains []float64
					for _, r := range out.Records {
						if r.Split != split || r.Scenario != name {
							continue
						}
						a, b := r.Arms[control], r.Arms[5]
						g := a.Full.Brier - b.Full.Brier
						if window == "post" {
							g = a.Post.Brier - b.Post.Brier
						}
						gains = append(gains, g)
					}
					mean := 0.
					for _, g := range gains {
						mean += g / Streams
					}
					variance := 0.
					for _, g := range gains {
						variance += (g - mean) * (g - mean) / (Streams - 1)
					}
					margin := 3.5 * math.Sqrt(variance/Streams)
					c := Comparison{split, name, Arms[control], window, mean, mean - margin, mean + margin}
					out.Comparisons = append(out.Comparisons, c)
					if split == "confirmation" && control == 0 {
						if name == "stable" && window == "full" {
							out.StablePass = out.StablePass && -c.Lower <= .01
						}
						if (name == "member_shift" || name == "common_shift") && window == "post" {
							out.ShiftPass = out.ShiftPass && c.Gain >= .05 && c.Lower > 0
						}
					}
				}
			}
		}
	}
	out.OverallPass = out.ShiftPass && out.StablePass
}
