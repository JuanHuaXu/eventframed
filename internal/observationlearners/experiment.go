package observationlearners

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

const FitBase int64 = 2026092301
const DesignBase int64 = 2026092302
const ConfirmationBase int64 = 2026092303

var Arms = []string{"frozen", "short_only", "adaptive_only", "fixed_mix", "adaptive_mix", "forest_mix", "combined_mix"}

type Scenario struct {
	Name          string
	Noise         float64
	Change, Delay int
	Missing       float64
}

var Scenarios = []Scenario{{"stable05", .05, 512, 0, 0}, {"stable20", .20, 512, 0, 0}, {"shift128", .05, 128, 0, 0}, {"shift256", .05, 256, 0, 0}, {"shift384", .05, 384, 0, 0}, {"gradual", .05, 128, 0, 0}, {"recurring", .05, 128, 0, 0}, {"delayed_missing", .05, 256, 16, .25}, {"interaction", .05, 256, 0, 0}, {"null", 0, 512, 0, 0}}

func Seed(base int64, scenario, fit, stream, role int) int64 {
	return base*1000000 + int64(scenario)*10000 + int64(fit)*1000 + int64(stream)*10 + int64(role)
}
func truth(x uint16, t int, s Scenario, r *rand.Rand) bool {
	if s.Name == "null" {
		return r.Intn(2) == 1
	}
	local := t >= s.Change
	if s.Name == "gradual" {
		local = r.Float64() < math.Max(0, math.Min(1, float64(t-128)/256))
	}
	if s.Name == "recurring" {
		local = t >= 0 && (t/128)%2 == 1
	}
	y := (x&(1<<6) != 0) != (x&(1<<7) != 0)
	y = y != (x&(1<<8) != 0)
	if local {
		y = x&(1<<2) != 0
		if s.Name == "interaction" {
			y = (x&1 != 0) != (x&2 != 0)
			y = y != (x&4 != 0)
		}
	}
	if r.Float64() < s.Noise {
		y = !y
	}
	return y
}
func Base(s Scenario, j, fit int) (*observation.Model, error) {
	r := rand.New(rand.NewSource(FitBase*1000000 + int64(j)*1000 + int64(fit)))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(r.Intn(512))
		samples[i] = observation.Sample{Bits: x, Outcome: truth(x, -1, s, r)}
	}
	return observation.Fit(samples)
}

type auditSample struct {
	Sample observation.Sample
	Origin int
}
type Prediction struct {
	P       float64
	Experts [4]float64
}
type Tick struct {
	Outcome, Missing, Audit           bool
	Predictions                       [7]Prediction
	Delivered                         []int
	Cutoff, Cuts, Audits, ForestNodes int
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
	Split, Scenario                               string
	Fit, Stream                                   int
	Ticks                                         []Tick `json:",omitempty"`
	Full, Post                                    [7]Metrics
	Recovery                                      [7]int
	Available, Audits, Pending, Cuts, Nodes, Fits int
	FitNS                                         int64
}
type packet struct {
	Origin, Due int
	X           uint16
	Y, Audit    bool
	Predictions [7]Prediction
}

func forecast(m *observation.Model, x uint16) float64 {
	if m == nil {
		return .5
	}
	p, e := m.ForecastObserved(511, x)
	if e != nil {
		panic(e)
	}
	return p
}
func RunStream(base *observation.Model, split string, j, fit, stream int, seed int64) (Record, error) {
	s := Scenarios[j]
	r := Record{Split: split, Scenario: s.Name, Fit: fit, Stream: stream}
	for i := range r.Recovery {
		r.Recovery[i] = -1
	}
	rng := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 0)))
	ar := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 1)))
	mr := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 2)))
	forest := NewForest(Seed(seed, j, fit, stream, 3))
	var short, long, adaptive *observation.Model
	var mix [7]bayes.ForecastMix
	var window Window
	var audits []auditSample
	var queue []packet
	fitAdaptive := func() error {
		var a []observation.Sample
		for _, v := range audits {
			if v.Origin >= window.Cutoff {
				a = append(a, v.Sample)
			}
		}
		if len(a) < 32 {
			adaptive = nil
			return nil
		}
		start := time.Now()
		var e error
		adaptive, e = observation.Fit(a)
		r.FitNS += int64(time.Since(start))
		r.Fits++
		return e
	}
	for t := 0; t < 512; t++ {
		tick := Tick{Audit: ar.Float64() < .25, Missing: mr.Float64() < s.Missing}
		x := uint16(rng.Intn(512))
		rd := observationexperiment.Frames(x, fmt.Sprintf("v5-%s-%s-%d-%d-%d", split, s.Name, fit, stream, t))
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
			return r, fmt.Errorf("incomplete full-view fixture")
		}
		b, sh, lo, ad, tr := forecast(base, values), forecast(short, values), forecast(long, values), forecast(adaptive, values), forest.Predict(values)
		for i, p := range []float64{b, sh, ad} {
			tick.Predictions[i] = Prediction{P: p}
		}
		bundles := [4][4]float64{{b, sh, lo, .5}, {b, ad, lo, .5}, {b, tr, lo, .5}, {b, ad, tr, .5}}
		for k, e := range bundles {
			i := k + 3
			tick.Predictions[i] = Prediction{P: mix[i].Forecast(e), Experts: e}
		}
		tick.Outcome = truth(x, t, s, rng)
		if !tick.Missing {
			queue = append(queue, packet{Origin: t, Due: t + s.Delay, X: values, Y: tick.Outcome, Audit: tick.Audit, Predictions: tick.Predictions})
		}
		for len(queue) > 0 && queue[0].Due <= t {
			q := queue[0]
			queue = queue[1:]
			r.Available++
			tick.Delivered = append(tick.Delivered, q.Origin)
			for i := 3; i < 7; i++ {
				mix[i] = mix[i].Observe(q.Predictions[i].Experts, q.Y, 1)
			}
			y := 0.
			if q.Y {
				y = 1
			}
			loss := (q.Predictions[0].P - y) * (q.Predictions[0].P - y)
			cut := window.Add(loss, q.Origin)
			refit := false
			if q.Audit {
				r.Audits++
				audits = append(audits, auditSample{observation.Sample{Bits: q.X, Outcome: q.Y}, q.Origin})
				if len(audits) > 256 {
					audits = audits[1:]
				}
				forest.Update(q.X, q.Y)
				refit = r.Audits >= 32 && r.Audits%16 == 0
				if refit {
					a := make([]observation.Sample, len(audits))
					for i, v := range audits {
						a[i] = v.Sample
					}
					start := time.Now()
					var e error
					short, e = observation.Fit(a[max(0, len(a)-64):])
					if e != nil {
						return r, e
					}
					long, e = observation.Fit(a)
					if e != nil {
						return r, e
					}
					r.FitNS += int64(time.Since(start))
					r.Fits += 2
				}
			}
			if refit || cut {
				if e := fitAdaptive(); e != nil {
					return r, e
				}
			}
		}
		tick.Cutoff, tick.Cuts, tick.Audits, tick.ForestNodes = window.Cutoff, window.Cuts, r.Audits, forest.Nodes()
		r.Ticks = append(r.Ticks, tick)
	}
	r.Pending = len(queue)
	r.Cuts = window.Cuts
	r.Nodes = forest.Nodes()
	post := s.Change
	if post >= 512 {
		post = 256
	}
	for i := 0; i < 7; i++ {
		r.Full[i] = score(r.Ticks, i, 0)
		r.Post[i] = score(r.Ticks, i, post)
		if s.Name == "shift128" || s.Name == "shift256" || s.Name == "shift384" {
			for end := s.Change + 63; end < 512; end++ {
				correct := 0
				for k := end - 63; k <= end; k++ {
					if (r.Ticks[k].Predictions[i].P >= .5) == r.Ticks[k].Outcome {
						correct++
					}
				}
				if correct >= 52 {
					r.Recovery[i] = end - s.Change
					break
				}
			}
		}
	}
	return r, nil
}

type Comparison struct {
	Split, Scenario, Arm, Window string
	Gain, Lower, Upper           float64
	PerFit                       [3]float64
}
type Verdict struct {
	Arm                                            string
	Shift, Protection, NoMeanHarm, SeedSigns, Pass bool
}
type Output struct {
	Protocol    string
	Hashes      map[string]string
	Records     []Record
	Comparisons []Comparison
	Verdicts    []Verdict
}

func Run(progress func(string)) (Output, error) {
	out := Output{Protocol: "docs/experiments/mmm-learners-v5-protocol.md"}
	for k, split := range []string{"design", "confirmation"} {
		seed := DesignBase
		if k == 1 {
			seed = ConfirmationBase
		}
		for j, s := range Scenarios {
			for fit := 0; fit < 3; fit++ {
				base, e := Base(s, j, fit)
				if e != nil {
					return out, e
				}
				for stream := 0; stream < 8; stream++ {
					r, e := RunStream(base, split, j, fit, stream, seed)
					if e != nil {
						return out, e
					}
					out.Records = append(out.Records, r)
				}
			}
			if progress != nil {
				progress(split + " " + s.Name + " complete")
			}
		}
	}
	Summarize(&out)
	return out, nil
}
func Summarize(out *Output) {
	out.Comparisons = nil
	out.Verdicts = nil
	for _, split := range []string{"design", "confirmation"} {
		for _, s := range Scenarios {
			for arm := 4; arm < 7; arm++ {
				for _, window := range []string{"full", "post"} {
					c := Comparison{Split: split, Scenario: s.Name, Arm: Arms[arm], Window: window}
					var gains []float64
					for _, r := range out.Records {
						if r.Split != split || r.Scenario != s.Name {
							continue
						}
						m := r.Full
						if window == "post" {
							m = r.Post
						}
						g := m[3].Brier - m[arm].Brier
						gains = append(gains, g)
						c.PerFit[r.Fit] += g / 8
						c.Gain += g / 24
					}
					v := 0.
					for _, g := range gains {
						v += (g - c.Gain) * (g - c.Gain) / 23
					}
					margin := 3.6 * math.Sqrt(v/24)
					c.Lower, c.Upper = c.Gain-margin, c.Gain+margin
					out.Comparisons = append(out.Comparisons, c)
				}
			}
		}
	}
	for arm := 4; arm < 7; arm++ {
		v := Verdict{Arm: Arms[arm], Shift: true, Protection: true, NoMeanHarm: true, SeedSigns: true}
		for _, c := range out.Comparisons {
			if c.Split != "confirmation" || c.Arm != v.Arm {
				continue
			}
			v.NoMeanHarm = v.NoMeanHarm && c.Gain >= -.01
			if (c.Scenario == "stable05" || c.Scenario == "stable20") && c.Window == "full" {
				v.Protection = v.Protection && -c.Lower <= .01
			}
			if (c.Scenario == "shift128" || c.Scenario == "shift256") && c.Window == "post" {
				v.Shift = v.Shift && c.Gain >= .005 && c.Lower > 0
				for _, g := range c.PerFit {
					v.SeedSigns = v.SeedSigns && g > 0
				}
			}
		}
		v.Pass = v.Shift && v.Protection && v.NoMeanHarm && v.SeedSigns
		out.Verdicts = append(out.Verdicts, v)
	}
}
