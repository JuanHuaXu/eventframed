package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeCadenceFit struct {
	Clock                       int
	Origins                     []int
	LogOdds, Mean, Variance, Xi []float64
	Trace                       [][3]float64
	Motion                      []float64
	Stop                        string
}
type spikeCadenceRecord struct {
	Clock                          *int `json:"Clock,omitempty"`
	Phase, Case, Index, Schedule   int
	Fits                           []spikeCadenceFit
	Used                           [32]int
	X                              [32]uint16
	P, Q, LogitMean, LogitVariance [32]float64
	Y                              [32]bool
	Control                        [32][15]float64
	Stats                          [32]spikePredictionStats
}

func spikeCadenceOrigins(r softV120Record, clock int) []int {
	var origins []int
	for i := -16; i < clock; i++ {
		if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
			origins = append(origins, i)
		}
	}
	if len(origins) > 64 {
		origins = origins[len(origins)-64:]
	}
	return origins
}

func runSpikeCadence(r softV120Record) (spikeCadenceRecord, error) {
	return runSpikeCadenceAt(r, 128)
}

func runSpikeCadenceAt(r softV120Record, start int) (spikeCadenceRecord, error) {
	out := spikeCadenceRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}
	if start < 0 || start > 224 || start%32 != 0 {
		return out, fmt.Errorf("cadence start clock")
	}
	if start != 128 {
		out.Clock = &start
	}
	if len(r.Steps) != 256 || len(r.Fits) != 8 {
		return out, fmt.Errorf("cadence dimensions")
	}
	var masks []uint16
	for m := uint16(1); m < 512; m++ {
		if bits.OnesCount16(m) <= 4 {
			masks = append(masks, m)
		}
	}
	var fit *spikeFit
	var previous []int
	for t := 0; t < 32; t++ {
		clock := start + t
		origins := spikeCadenceOrigins(r, clock)
		if t == 0 && !reflect.DeepEqual(origins, r.Fits[start/32].Origins[0]) {
			return out, fmt.Errorf("initial cadence origins")
		}
		if fit == nil || !reflect.DeepEqual(origins, previous) {
			samples := make([]observation.Sample, len(origins))
			for j, i := range origins {
				if i < 0 {
					samples[j] = r.Initial[i+16]
				} else {
					samples[j] = observation.Sample{Bits: r.Steps[i].X, Outcome: r.Steps[i].Y}
				}
			}
			var err error
			fit, err = fitSpikeFixed(samples, masks, 1.0/255, 1, 1024)
			if err != nil {
				return out, fmt.Errorf("clock%d: %w", clock, err)
			}
			previous = origins
			out.Fits = append(out.Fits, spikeCadenceFit{clock, origins, fit.q.logOdds, fit.q.mean, fit.q.variance, fit.profile.xi, fit.trace, fit.motion, fit.stop})
		}
		out.Used[t] = len(out.Fits) - 1
		x := r.Steps[clock].X
		p, st, err := fit.predict(x)
		if err != nil {
			return out, err
		}
		l, err := fit.law(x)
		if err != nil {
			return out, err
		}
		_, _, mu, err := l.validate()
		if err != nil {
			return out, err
		}
		variance := l.variance
		for j, g := range l.inclusion {
			variance += g*l.slabVariance[j] + g*l.excluded(j)*l.mean[j]*l.mean[j]
		}
		out.X[t], out.P[t], out.Stats[t], out.LogitMean[t], out.LogitVariance[t] = x, p, st, mu, variance
	}
	// Outcomes below are evaluator output only; none are fed back before arrival.
	for t := 0; t < 32; t++ {
		s := r.Steps[start+t]
		out.Q[t], out.Y[t], out.Control[t] = s.Q, s.Y, s.P
	}
	return out, nil
}

func TestSpikeCadenceAsOf(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256), Fits: make([]softV120Fit, 8)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%3 == 0, Delay: i % 32, Missing: i%5 == 0}
	}
	r.Fits[4].Origins[0] = spikeCadenceOrigins(r, 128)
	base, err := runSpikeCadence(r)
	if err != nil {
		t.Fatal(err)
	}
	old, err := runSpikePilot(r)
	if err != nil {
		t.Fatal(err)
	}
	if base.P[0] != old.P[0][0] || !reflect.DeepEqual(base.Fits[0].Trace, old.Trace) {
		t.Fatal("initial state differs")
	}
	poison := r
	poison.Steps = append([]softV120Step(nil), r.Steps...)
	for i := range poison.Steps {
		s := &poison.Steps[i]
		s.Q = math.NaN()
		s.P = [15]float64{math.NaN()}
		if i >= 144 || s.Missing || i+s.Delay > 144 {
			s.Y = !s.Y
		}
	}
	got, err := runSpikeCadence(poison)
	if err != nil {
		t.Fatal(err)
	}
	for t0 := 0; t0 <= 16; t0++ {
		if base.P[t0] != got.P[t0] || base.Used[t0] != got.Used[t0] {
			t.Fatal("prefix leaks future", t0)
		}
	}
	if !reflect.DeepEqual(base.Fits[:base.Used[16]+1], got.Fits[:got.Used[16]+1]) {
		t.Fatal("fit prefix leak")
	}
	if got.P == base.P {
		t.Fatal("later admitted changes had no effect")
	}
	// With no retained-window change, one fit must serve the whole block.
	for i := range r.Steps {
		r.Steps[i].Delay = 0
		r.Steps[i].Missing = i >= 128
	}
	r.Fits[4].Origins[0] = spikeCadenceOrigins(r, 128)
	quiet, err := runSpikeCadence(r)
	if err != nil || len(quiet.Fits) != 1 {
		t.Fatal("unnecessary quiet refit", err)
	}
	frozen, err := runSpikePilot(r)
	if err != nil {
		t.Fatal(err)
	}
	for i := range quiet.P {
		if quiet.Used[i] != 0 || quiet.P[i] != frozen.P[i][0] {
			t.Fatal("quiet law changed")
		}
	}
}

func TestSpikeCadenceCollect(t *testing.T) {
	start := 128
	if value := os.Getenv("EVENTFRAME_SPIKE_CADENCE_CLOCK"); value != "" {
		var err error
		start, err = strconv.Atoi(value)
		if err != nil || (start != 0 && start != 128 && start != 224) {
			t.Fatal("invalid collection clock", value)
		}
	}
	indices := 1
	if os.Getenv("EVENTFRAME_SPIKE_CADENCE_BREADTH") == "eight" {
		indices = 8
	} else if os.Getenv("EVENTFRAME_SPIKE_CADENCE_BREADTH") != "" {
		t.Fatal("unknown breadth")
	}
	src, dst := os.Getenv("EVENTFRAME_SPIKE_CADENCE_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_CADENCE_OUTPUT")
	if src == "" {
		t.Skip("explicit research source")
	}
	if dst == "" {
		t.Fatal("output required")
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var h softV120Artifact
	if err = dec.Decode(&h); err != nil || h.Version != "soft-learners-v120" {
		t.Fatal("header", err)
	}
	count, fits := 0, 0
	for {
		var r softV120Record
		err = dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index < 0 || r.Index >= indices {
			continue
		}
		v, err := runSpikeCadenceAt(r, start)
		if err != nil {
			t.Fatalf("record%d: %v", count, err)
		}
		if err = enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		count++
		fits += len(v.Fits)
	}
	if count != 84*indices {
		t.Fatal("count", count)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("records=%d fits=%d forecasts=%d", count, fits, count*32)
}
