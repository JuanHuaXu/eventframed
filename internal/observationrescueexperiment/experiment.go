package observationrescueexperiment

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

const FitSeed int64 = 2026091401
const DesignSeed int64 = 2026091402
const ConfirmationSeed int64 = 2026091403
const Streams = 32
const Steps = 512

var Scenarios = []string{"stable", "member_shift", "common_shift", "recurring", "null"}

func outcome(x uint16, local, null bool, rng *rand.Rand) bool {
	if null {
		return rng.Intn(2) == 1
	}
	y := (x&(1<<6) != 0) != (x&(1<<7) != 0)
	y = y != (x&(1<<8) != 0)
	if local {
		y = x&(1<<2) != 0
	}
	if rng.Float64() < .05 {
		y = !y
	}
	return y
}
func Base(null bool) (*observation.Model, error) {
	rng := rand.New(rand.NewSource(FitSeed))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(rng.Intn(observation.Universe))
		samples[i] = observation.Sample{Bits: x, Outcome: outcome(x, false, null, rng)}
	}
	return observation.Fit(samples)
}

type Tick struct {
	P          float64                      `json:"p"`
	Raw        float64                      `json:"raw"`
	Epoch      uint64                       `json:"epoch"`
	Mask       uint16                       `json:"mask"`
	Values     uint16                       `json:"values"`
	Cost       int                          `json:"cost"`
	Support    uint32                       `json:"support"`
	Transition observationrescue.Transition `json:"transition"`
}
type Metrics struct {
	N               int     `json:"n"`
	Brier           float64 `json:"brier"`
	LogLoss         float64 `json:"log_loss"`
	Accuracy        float64 `json:"accuracy"`
	ConfidentErrors int     `json:"confident_errors"`
}

func (m *Metrics) add(p float64, y bool) {
	m.N++
	x := 0.
	if y {
		x = 1
	}
	m.Brier += (p - x) * (p - x)
	if y {
		m.LogLoss -= math.Log(p)
	} else {
		m.LogLoss -= math.Log1p(-p)
	}
	if (p >= .5) == y {
		m.Accuracy++
	} else if p <= .1 || p >= .9 {
		m.ConfidentErrors++
	}
}
func (m *Metrics) finish() {
	if m.N > 0 {
		n := float64(m.N)
		m.Brier /= n
		m.LogLoss /= n
		m.Accuracy /= n
	}
}

type PolicyRecord struct {
	Arm               string  `json:"arm"`
	Ticks             []Tick  `json:"ticks,omitempty"`
	Full              Metrics `json:"full"`
	Post              Metrics `json:"post"`
	Late              Metrics `json:"late"`
	FirstInvalidation int     `json:"first_invalidation"`
	FirstReplacement  int     `json:"first_replacement"`
	ForegroundCost    int     `json:"foreground_cost"`
	FitCount          int     `json:"fit_count"`
	FitNanoseconds    int64   `json:"fit_nanoseconds"`
}
type StreamRecord struct {
	Split       string         `json:"split"`
	Scenario    string         `json:"scenario"`
	Index       int            `json:"index"`
	Outcomes    []bool         `json:"outcomes,omitempty"`
	Audit       []bool         `json:"audit,omitempty"`
	MonitorCost int            `json:"monitor_cost"`
	AuditCost   int            `json:"audit_cost"`
	Policies    []PolicyRecord `json:"policies"`
}

func RunStream(base *observation.Model, split, scenario string, index int, seed int64) (StreamRecord, error) {
	r := StreamRecord{Split: split, Scenario: scenario, Index: index}
	liveRNG := rand.New(rand.NewSource(seed + int64(index)*100))
	refRNG := rand.New(rand.NewSource(seed + int64(index)*100 + 1))
	auditRNG := rand.New(rand.NewSource(seed + int64(index)*100 + 2))
	states := make([]*observationrescue.State, len(observationrescue.Arms))
	for i, arm := range observationrescue.Arms {
		state, err := observationrescue.New(base, arm)
		if err != nil {
			return r, err
		}
		states[i] = state
		r.Policies = append(r.Policies, PolicyRecord{Arm: arm, FirstInvalidation: -1, FirstReplacement: -1})
	}
	for step := 0; step < Steps; step++ {
		audit := auditRNG.Float64() < .25 // ex ante, independent of values and outcomes
		local := (scenario == "member_shift" || scenario == "common_shift") && step >= 256
		if scenario == "recurring" {
			local = (step/128)%2 == 1
		}
		x := uint16(liveRNG.Intn(observation.Universe))
		refX := uint16(refRNG.Intn(observation.Universe))
		reader := observationexperiment.Frames(x, fmt.Sprintf("v2-live-%s-%s-%d-%d", split, scenario, index, step))
		refReader := observationexperiment.Frames(refX, fmt.Sprintf("v2-reference-%s-%s-%d-%d", split, scenario, index, step))
		old, err := observation.Run(base, reader, 1, "mmm", 0)
		if err != nil {
			return r, err
		}
		ref, err := observation.Run(base, refReader, 1, "mmm", 0)
		if err != nil {
			return r, err
		}
		r.MonitorCost += old.Cost + ref.Cost
		forecasts := make([]observationrescue.Forecast, len(states))
		for i, state := range states {
			f, err := state.Predict(reader, step)
			if err != nil {
				return r, err
			}
			forecasts[i] = f
		}
		// The full-view shadow read is charged and never supplied to the just
		// emitted forecast. Incomplete observations cannot form a training sample.
		var sample *observation.Sample
		if audit {
			r.AuditCost += 9
			var mask, values uint16
			for scope := 0; scope < 3; scope++ {
				m, v, err := reader.Read(observation.View{Scope: scope, Depth: 2})
				if err != nil {
					return r, err
				}
				mask |= m
				values |= v
			}
			if mask == observation.Universe-1 {
				sample = &observation.Sample{Bits: values}
			}
		}
		y := outcome(x, local, scenario == "null", liveRNG)
		refY := outcome(refX, scenario == "common_shift" && step >= 256, scenario == "null", refRNG)
		if sample != nil {
			sample.Outcome = y
		}
		r.Outcomes = append(r.Outcomes, y)
		r.Audit = append(r.Audit, audit)
		for i, state := range states {
			f := forecasts[i]
			transition, err := state.Observe(observationrescue.Feedback{Sequence: step, Epoch: f.Epoch,
				ReferenceCorrect: (ref.Probability >= .5) == refY, LiveCorrect: (old.Probability >= .5) == y,
				ValidationEligible: true, Audit: sample})
			if err != nil {
				return r, err
			}
			p := &r.Policies[i]
			last := f.Inspection.Trace[len(f.Inspection.Trace)-1]
			p.Ticks = append(p.Ticks, Tick{f.Probability, f.Inspection.Probability, f.Epoch, last.Observed, last.Values, f.Inspection.Cost, f.Support, transition})
			if transition.Invalidated && p.FirstInvalidation < 0 {
				p.FirstInvalidation = step
			}
			if transition.Refitted && p.FirstReplacement < 0 {
				p.FirstReplacement = step
			}
		}
	}
	for i, state := range states {
		p := &r.Policies[i]
		p.FitCount = state.FitCount
		p.FitNanoseconds = int64(state.FitDuration)
		score(p, r.Outcomes, scenario)
	}
	return r, nil
}

func score(p *PolicyRecord, outcomes []bool, scenario string) {
	p.Full, p.Post, p.Late = Metrics{}, Metrics{}, Metrics{}
	p.ForegroundCost = 0
	post := 256
	if scenario == "recurring" {
		post = 128
	}
	for t, tick := range p.Ticks {
		p.Full.add(tick.P, outcomes[t])
		if t >= post {
			p.Post.add(tick.P, outcomes[t])
		}
		if t >= 384 {
			p.Late.add(tick.P, outcomes[t])
		}
		p.ForegroundCost += tick.Cost
	}
	p.Full.finish()
	p.Post.finish()
	p.Late.finish()
}

type Summary struct {
	Split               string     `json:"split"`
	Scenario            string     `json:"scenario"`
	Arm                 string     `json:"arm"`
	Streams             int        `json:"streams"`
	FullBrier           float64    `json:"full_brier"`
	PostBrier           float64    `json:"post_brier"`
	LateBrier           float64    `json:"late_brier"`
	PostLogLoss         float64    `json:"post_log_loss"`
	PostAccuracy        float64    `json:"post_accuracy"`
	PostConfidentErrors int        `json:"post_confident_errors"`
	Invalidations       int        `json:"invalidations"`
	EarlyInvalidations  int        `json:"early_invalidations"`
	Detected            int        `json:"detected_after_change"`
	Missed              int        `json:"missed_after_change"`
	MeanDelay           *float64   `json:"mean_delay_detected_only"`
	InvalidationWilson  [2]float64 `json:"invalidation_wilson95"`
	MeanForegroundCost  float64    `json:"mean_foreground_cost"`
	MeanMonitorCost     float64    `json:"mean_monitor_cost"`
	MeanAuditCost       float64    `json:"mean_audit_cost"`
	Fits                int        `json:"fits"`
	FitNanoseconds      int64      `json:"fit_nanoseconds"`
}
type Comparison struct {
	Split    string  `json:"split"`
	Scenario string  `json:"scenario"`
	Control  string  `json:"control"`
	Metric   string  `json:"metric"`
	Gain     float64 `json:"gain"`
	Lower    float64 `json:"lower"`
	Upper    float64 `json:"upper"`
}
type Output struct {
	Protocol               string            `json:"protocol"`
	Runtime                string            `json:"runtime"`
	Hashes                 map[string]string `json:"sha256"`
	FitSeed                int64             `json:"fit_seed"`
	DesignSeed             int64             `json:"design_seed"`
	ConfirmationSeed       int64             `json:"confirmation_seed"`
	Records                []StreamRecord    `json:"records,omitempty"`
	Summaries              []Summary         `json:"summaries"`
	Comparisons            []Comparison      `json:"comparisons"`
	ShiftGainPassed        bool              `json:"shift_gain_passed"`
	StableProtectionPassed bool              `json:"stable_protection_passed"`
	OverallPassed          bool              `json:"overall_passed"`
}

func Run(progress func(string)) (Output, error) {
	out := Output{Protocol: "docs/experiments/mmm-antipigeon-v2-protocol.md", FitSeed: FitSeed, DesignSeed: DesignSeed, ConfirmationSeed: ConfirmationSeed}
	for _, split := range []string{"design", "confirmation"} {
		for j, scenario := range Scenarios {
			base, err := Base(scenario == "null")
			if err != nil {
				return out, err
			}
			seed := DesignSeed
			if split == "confirmation" {
				seed = ConfirmationSeed
			}
			seed = seed*1000000 + int64(j)*100000
			for i := 0; i < Streams; i++ {
				r, err := RunStream(base, split, scenario, i, seed)
				if err != nil {
					return out, err
				}
				out.Records = append(out.Records, r)
			}
			if progress != nil {
				progress(split + " " + scenario + " complete")
			}
		}
	}
	Summarize(&out)
	return out, nil
}
func wilson(k, n int) [2]float64 {
	if n == 0 {
		return [2]float64{0, 1}
	}
	p := float64(k) / float64(n)
	z := 1.959963984540054
	den := 1 + z*z/float64(n)
	c := (p + z*z/(2*float64(n))) / den
	r := z * math.Sqrt(p*(1-p)/float64(n)+z*z/(4*float64(n*n))) / den
	return [2]float64{math.Max(0, c-r), math.Min(1, c+r)}
}
func Summarize(out *Output) {
	out.Summaries = nil
	out.Comparisons = nil
	for _, split := range []string{"design", "confirmation"} {
		for _, scenario := range Scenarios {
			for armIndex, arm := range observationrescue.Arms {
				s := Summary{Split: split, Scenario: scenario, Arm: arm}
				delay := 0.
				for _, r := range out.Records {
					if r.Split != split || r.Scenario != scenario {
						continue
					}
					p := r.Policies[armIndex]
					s.Streams++
					s.FullBrier += p.Full.Brier
					s.PostBrier += p.Post.Brier
					s.LateBrier += p.Late.Brier
					s.PostLogLoss += p.Post.LogLoss
					s.PostAccuracy += p.Post.Accuracy
					s.PostConfidentErrors += p.Post.ConfidentErrors
					s.MeanForegroundCost += float64(p.ForegroundCost) / Steps
					s.MeanMonitorCost += float64(r.MonitorCost) / Steps
					s.MeanAuditCost += float64(r.AuditCost) / Steps
					s.Fits += p.FitCount
					s.FitNanoseconds += p.FitNanoseconds
					if p.FirstInvalidation >= 0 {
						s.Invalidations++
					}
					if scenario != "stable" && scenario != "null" {
						change := 256
						if scenario == "recurring" {
							change = 128
						}
						if p.FirstInvalidation >= change {
							s.Detected++
							delay += float64(p.FirstInvalidation - change)
						} else {
							s.Missed++
							if p.FirstInvalidation >= 0 {
								s.EarlyInvalidations++
							}
						}
					}
				}
				n := float64(s.Streams)
				s.FullBrier /= n
				s.PostBrier /= n
				s.LateBrier /= n
				s.PostLogLoss /= n
				s.PostAccuracy /= n
				s.MeanForegroundCost /= n
				s.MeanMonitorCost /= n
				s.MeanAuditCost /= n
				if s.Detected > 0 {
					d := delay / float64(s.Detected)
					s.MeanDelay = &d
				}
				s.InvalidationWilson = wilson(s.Invalidations, s.Streams)
				out.Summaries = append(out.Summaries, s)
			}
			for _, control := range []int{0, 1, 4, 3} {
				for _, metric := range []string{"full", "post"} {
					var gains []float64
					for _, r := range out.Records {
						if r.Split != split || r.Scenario != scenario {
							continue
						}
						a, b := r.Policies[control], r.Policies[5]
						g := a.Full.Brier - b.Full.Brier
						if metric == "post" {
							g = a.Post.Brier - b.Post.Brier
						}
						gains = append(gains, g)
					}
					n := float64(len(gains))
					mean := 0.
					for _, g := range gains {
						mean += g / n
					}
					variance := 0.
					for _, g := range gains {
						variance += (g - mean) * (g - mean) / (n - 1)
					}
					margin := 3.5 * math.Sqrt(variance/n)
					out.Comparisons = append(out.Comparisons, Comparison{split, scenario, observationrescue.Arms[control], metric, mean, mean - margin, mean + margin})
				}
			}
		}
	}
	out.ShiftGainPassed = true
	out.StableProtectionPassed = true
	for _, c := range out.Comparisons {
		if c.Split != "confirmation" || c.Control != "frozen" {
			continue
		}
		if (c.Scenario == "member_shift" || c.Scenario == "common_shift") && c.Metric == "post" {
			out.ShiftGainPassed = out.ShiftGainPassed && c.Gain >= .05 && c.Lower > 0
		}
		if c.Scenario == "stable" && c.Metric == "full" {
			out.StableProtectionPassed = out.StableProtectionPassed && -c.Lower <= .01
		}
	}
	for _, s := range out.Summaries {
		if s.Split == "confirmation" && s.Scenario == "stable" && s.Arm == "ap_audit" {
			out.StableProtectionPassed = out.StableProtectionPassed && float64(s.Invalidations)/float64(s.Streams) <= .05
		}
	}
	out.OverallPassed = out.ShiftGainPassed && out.StableProtectionPassed
}
