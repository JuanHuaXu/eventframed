package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikePilotRecord struct {
	Phase, Case, Index, Schedule  int
	Origins                       []int
	Inclusion, Mean, Variance, Xi []float64
	LogOdds, Exclusion            []float64
	Trace                         [][3]float64
	Motion                        []float64
	Stop                          string
	X                             [32]uint16
	Q                             [32]float64
	Y                             [32]bool
	P                             [32][2]float64
	Control                       [32][15]float64
	Stats                         [32]spikePredictionStats
	LogitMean, LogitVariance      [32]float64
}

func runSpikePilot(r softV120Record) (spikePilotRecord, error) {
	return runSpikePilotAt(r, 128)
}

func runSpikePilotAt(r softV120Record, clock int) (spikePilotRecord, error) {
	return runSpikePilotOrdered(r, clock, false)
}

func runSpikePilotOrdered(r softV120Record, clock int, reverse bool) (spikePilotRecord, error) {
	out := spikePilotRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}
	if len(r.Steps) != 256 || len(r.Fits) != 8 || clock < 0 || clock > 224 || clock%32 != 0 {
		return out, fmt.Errorf("spike pilot dimensions")
	}
	for i := -16; i < clock; i++ {
		if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
			out.Origins = append(out.Origins, i)
		}
	}
	if len(out.Origins) > 64 {
		out.Origins = out.Origins[len(out.Origins)-64:]
	}
	if !reflect.DeepEqual(out.Origins, r.Fits[clock/32].Origins[0]) {
		return out, fmt.Errorf("spike pilot origins")
	}
	samples := make([]observation.Sample, len(out.Origins))
	for j, i := range out.Origins {
		if i < 0 {
			samples[j] = r.Initial[i+16]
		} else {
			samples[j] = observation.Sample{Bits: r.Steps[i].X, Outcome: r.Steps[i].Y}
		}
	}
	var masks []uint16
	for m := uint16(1); m < 512; m++ {
		if bits.OnesCount16(m) <= 4 {
			masks = append(masks, m)
		}
	}
	if reverse {
		for i, j := 0, len(masks)-1; i < j; i, j = i+1, j-1 {
			masks[i], masks[j] = masks[j], masks[i]
		}
	}
	f, err := fitSpikeFixed(samples, masks, 1.0/255, 1, 1024)
	if err != nil {
		return out, err
	}
	out.LogOdds, out.Mean, out.Variance = f.q.logOdds, f.q.mean, f.q.variance
	out.Inclusion, out.Exclusion = make([]float64, len(masks)), make([]float64, len(masks))
	for j, odds := range f.q.logOdds {
		out.Inclusion[j], out.Exclusion[j], _, _ = spikeBernoulli(odds)
	}
	out.Xi, out.Trace, out.Motion, out.Stop = f.profile.xi, f.trace, f.motion, f.stop
	for i := 0; i < 32; i++ {
		x := r.Steps[clock+i].X
		p, st, err := f.predict(x)
		if err != nil {
			return out, err
		}
		l, err := f.law(x)
		if err != nil {
			return out, err
		}
		_, _, mu, err := l.validate()
		if err != nil {
			return out, err
		}
		out.X[i], out.P[i], out.Stats[i] = x, [2]float64{p, math.Max(1e-12, math.Min(1-1e-12, ridgeSigmoid(mu)))}, st
		variance := l.variance
		for j, g := range l.inclusion {
			variance += g*l.slabVariance[j] + g*l.excluded(j)*l.mean[j]*l.mean[j]
		}
		out.LogitMean[i], out.LogitVariance[i] = mu, variance
	}
	// Evaluator fields cannot affect the fit or any forecast in this block.
	for i := 0; i < 32; i++ {
		s := r.Steps[clock+i]
		out.Q[i], out.Y[i], out.Control[i] = s.Q, s.Y, s.P
	}
	return out, nil
}

func TestSpikePilotAsOf(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256), Fits: make([]softV120Fit, 8)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%3 == 0, Q: .3, Delay: i % 32, Missing: i%5 == 0}
	}
	for i := -16; i < 128; i++ {
		if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= 128) {
			r.Fits[4].Origins[0] = append(r.Fits[4].Origins[0], i)
		}
	}
	n := len(r.Fits[4].Origins[0])
	r.Fits[4].Origins[0] = r.Fits[4].Origins[0][n-64:]
	base, err := runSpikePilot(r)
	if err != nil {
		t.Fatal(err)
	}
	poison := r
	poison.Steps = append([]softV120Step(nil), r.Steps...)
	poison.Case = 999
	poison.Seeds = [5]int64{-1, -2, -3, -4, -5}
	poison.Rules = [2]uint16{511, 511}
	for i := range poison.Steps {
		s := &poison.Steps[i]
		s.Q = math.NaN()
		s.P = [15]float64{math.NaN()}
		if i >= 128 || s.Missing || i+s.Delay > 128 {
			s.Y = !s.Y
		}
	}
	got, err := runSpikePilot(poison)
	if err != nil {
		t.Fatal(err)
	}
	if base.P != got.P || !reflect.DeepEqual(base.Trace, got.Trace) || !reflect.DeepEqual(base.Inclusion, got.Inclusion) {
		t.Fatal("unavailable evidence leak")
	}
	for _, i := range base.Origins {
		if i < 0 {
			poison.Initial[i+16].Outcome = !r.Initial[i+16].Outcome
		} else {
			poison.Steps[i].Y = !r.Steps[i].Y
		}
	}
	got, err = runSpikePilot(poison)
	if err != nil {
		t.Fatal(err)
	}
	if got.P == base.P {
		t.Fatal("admitted labels ignored")
	}
	for i := range base.P {
		if math.Abs(base.P[i][0]+got.P[i][0]-1) > 1e-8 {
			t.Fatal("complement")
		}
	}
}

func TestSpikePilotCollect(t *testing.T) {
	src, dst := os.Getenv("EVENTFRAME_SPIKE_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_OUTPUT")
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
		t.Fatal("source header", err)
	}
	count := 0
	for {
		var r softV120Record
		err = dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index != 0 {
			continue
		}
		v, err := runSpikePilot(r)
		if err != nil {
			t.Fatalf("fit %d phase%d case%d schedule%d: %v", count, r.Phase, r.Case, r.Schedule, err)
		}
		if err = enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 84 {
		t.Fatal("fit count", count)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("fits=%d forecasts=%d", count, count*32)
}
