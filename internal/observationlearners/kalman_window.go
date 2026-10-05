package observationlearners

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const kalmanDim = 11

// KalmanResidual is a bounded linear-Gaussian working approximation for a
// Bernoulli forecast residual. It is not a calibrated Bernoulli posterior.
type KalmanResidual struct {
	Mean [kalmanDim]float64
	Cov  [kalmanDim][kalmanDim]float64
}

func NewKalmanResidual() KalmanResidual {
	var k KalmanResidual
	for i := range k.Cov {
		k.Cov[i][i] = 1
	}
	return k
}

func kalmanFeatures(x uint16, base float64) [kalmanDim]float64 {
	var h [kalmanDim]float64
	h[0] = 1
	h[1] = 2 * (base - .5)
	for i := 0; i < 9; i++ {
		if x&(1<<i) != 0 {
			h[i+2] = 1. / 3
		} else {
			h[i+2] = -1. / 3
		}
	}
	return h
}

func (k *KalmanResidual) Advance() {
	for i := range k.Cov {
		k.Cov[i][i] += .0005
	}
}

func (k *KalmanResidual) Predict(base float64, h [kalmanDim]float64) float64 {
	p := base
	for i := range h {
		p += h[i] * k.Mean[i]
	}
	return math.Max(.01, math.Min(.99, p))
}

func (k *KalmanResidual) Observe(base float64, h [kalmanDim]float64, y bool) error {
	var ph [kalmanDim]float64
	s := .25
	mu := 0.
	for i := range h {
		mu += h[i] * k.Mean[i]
		for j := range h {
			ph[i] += k.Cov[i][j] * h[j]
		}
		s += h[i] * ph[i]
	}
	if !isFinite(s) || s <= 0 {
		return fmt.Errorf("invalid innovation variance %g", s)
	}
	target := -base
	if y {
		target++
	}
	innovation := target - mu
	for i := range h {
		k.Mean[i] += ph[i] / s * innovation
		for j := i; j < len(h); j++ {
			v := k.Cov[i][j] - ph[i]*ph[j]/s
			k.Cov[i][j], k.Cov[j][i] = v, v
		}
		if !isFinite(k.Mean[i]) || !isFinite(k.Cov[i][i]) || k.Cov[i][i] < -1e-9 {
			return fmt.Errorf("invalid filter state at %d", i)
		}
	}
	return nil
}

func isFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

type KalmanTick struct {
	Outcome, Audit, Missing bool
	Predictions             [4]float64
	Delivered               []int `json:",omitempty"`
}

type KalmanMeasure struct {
	N        int
	Brier    float64
	Accuracy float64
}

type KalmanWindowRecord struct {
	Split, Scenario     string
	Fit, Stream         int
	Ticks               []KalmanTick
	Full, Tail, Early   [4]KalmanMeasure
	Audits, Available   int
	Pending, Fits, Cuts int
	Updates, StateBytes int
	UpdateP99NS         int64
}

func kalmanMeasure(ticks []KalmanTick, arm, start, end int) KalmanMeasure {
	m := KalmanMeasure{N: end - start}
	for _, tick := range ticks[start:end] {
		p := tick.Predictions[arm]
		y := 0.
		if tick.Outcome {
			y = 1
		}
		m.Brier += (p - y) * (p - y)
		if (p >= .5) == tick.Outcome {
			m.Accuracy++
		}
	}
	m.Brier /= float64(m.N)
	m.Accuracy /= float64(m.N)
	return m
}

type kalmanPacket struct {
	Origin, Due int
	X           uint16
	Y, Audit    bool
	Base        float64
	Features    [kalmanDim]float64
}

// RunKalmanWindowStream mirrors the existing observation learner's feedback
// order: forecast, generate outcome, then deliver only due packets.
func RunKalmanWindowStream(split string, scenario, fit, stream int) (KalmanWindowRecord, error) {
	if scenario < 0 || scenario >= len(Scenarios) || fit < 0 || stream < 0 || (split != "design" && split != "confirmation") {
		return KalmanWindowRecord{}, fmt.Errorf("invalid stream coordinates")
	}
	fitOffset := 0
	if split == "confirmation" {
		fitOffset = 100
	}
	base, err := Base(Scenarios[scenario], scenario, fit+fitOffset)
	if err != nil {
		return KalmanWindowRecord{}, err
	}
	return RunKalmanWindowWithBase(base, split, scenario, fit, stream)
}

func RunKalmanWindowWithBase(base *observation.Model, split string, scenario, fit, stream int) (KalmanWindowRecord, error) {
	if base == nil || scenario < 0 || scenario >= len(Scenarios) || fit < 0 || stream < 0 || (split != "design" && split != "confirmation") {
		return KalmanWindowRecord{}, fmt.Errorf("invalid stream coordinates or base")
	}
	s := Scenarios[scenario]
	seedBase := int64(20261101)
	if split == "confirmation" {
		seedBase = 20261102
	}
	rng := rand.New(rand.NewSource(Seed(seedBase, scenario, fit, stream, 0)))
	auditRNG := rand.New(rand.NewSource(Seed(seedBase, scenario, fit, stream, 1)))
	missingRNG := rand.New(rand.NewSource(Seed(seedBase, scenario, fit, stream, 2)))
	r := KalmanWindowRecord{Split: split, Scenario: s.Name, Fit: fit, Stream: stream, StateBytes: int(unsafe.Sizeof(KalmanResidual{}))}
	k := NewKalmanResidual()
	var short, adaptive *observation.Model
	var window Window
	var audits []auditSample
	var queue []kalmanPacket
	var durations []int64
	var err error
	fitAdaptive := func() error {
		retained := make([]observation.Sample, 0, len(audits))
		for _, v := range audits {
			if v.Origin >= window.Cutoff {
				retained = append(retained, v.Sample)
			}
		}
		if len(retained) < 32 {
			adaptive = nil
			return nil
		}
		var err error
		adaptive, err = observation.Fit(retained)
		if err == nil {
			r.Fits++
		}
		return err
	}
	for t := 0; t < 512; t++ {
		k.Advance()
		x := uint16(rng.Intn(512))
		b := forecast(base, x)
		h := kalmanFeatures(x, b)
		tick := KalmanTick{Audit: auditRNG.Float64() < .25, Missing: missingRNG.Float64() < s.Missing}
		tick.Predictions = [4]float64{b, forecast(short, x), forecast(adaptive, x), k.Predict(b, h)}
		tick.Outcome = truth(x, t, s, rng)
		if !tick.Missing {
			queue = append(queue, kalmanPacket{Origin: t, Due: t + s.Delay, X: x, Y: tick.Outcome, Audit: tick.Audit, Base: b, Features: h})
		}
		for len(queue) > 0 && queue[0].Due <= t {
			q := queue[0]
			queue = queue[1:]
			r.Available++
			tick.Delivered = append(tick.Delivered, q.Origin)
			y := 0.
			if q.Y {
				y = 1
			}
			cut := window.Add((q.Base-y)*(q.Base-y), q.Origin)
			refit := false
			if q.Audit {
				r.Audits++
				start := time.Now()
				if err := k.Observe(q.Base, q.Features, q.Y); err != nil {
					return r, err
				}
				durations = append(durations, time.Since(start).Nanoseconds())
				audits = append(audits, auditSample{Sample: observation.Sample{Bits: q.X, Outcome: q.Y}, Origin: q.Origin})
				if len(audits) > 256 {
					audits = audits[1:]
				}
				refit = r.Audits >= 32 && r.Audits%16 == 0
				if refit {
					all := make([]observation.Sample, len(audits))
					for i, v := range audits {
						all[i] = v.Sample
					}
					short, err = observation.Fit(all[max(0, len(all)-64):])
					if err != nil {
						return r, err
					}
					r.Fits++
				}
			}
			if refit || cut {
				if err := fitAdaptive(); err != nil {
					return r, err
				}
			}
		}
		r.Ticks = append(r.Ticks, tick)
	}
	r.Pending, r.Cuts, r.Updates = len(queue), window.Cuts, len(durations)
	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		r.UpdateP99NS = durations[int(math.Ceil(.99*float64(len(durations))))-1]
	}
	for a := 0; a < 4; a++ {
		r.Full[a] = kalmanMeasure(r.Ticks, a, 0, 512)
		r.Tail[a] = kalmanMeasure(r.Ticks, a, 384, 512)
		start := s.Change
		if start >= 512 {
			start = 0
		}
		r.Early[a] = kalmanMeasure(r.Ticks, a, start, min(start+64, 512))
	}
	return r, nil
}
