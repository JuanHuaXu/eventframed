package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type sparsePilotRecord struct {
	Phase, Case, Index, Schedule, Clock, Window int
	Origins                                     []int
	UsedXi, FinalXi, Rates                      []float64
	Mean, Diagonal, PriorVariance               [256]float64
	Trace                                       [][3]float64
	Stop                                        string
	Residual, Logdet                            float64
	X                                           [32]uint16
	Q                                           [32]float64
	Y                                           [32]bool
	P                                           [32][2]float64
	Control                                     [32][15]float64
	LogitMean, LogitVariance                    [32]float64
}

func runSparsePilot(r softV120Record, clock, window int) (sparsePilotRecord, error) {
	return runSparsePilotBudget(r, clock, window, 64)
}

func runSparsePilotBudget(r softV120Record, clock, window, limit int) (sparsePilotRecord, error) {
	out := sparsePilotRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule, Clock: clock, Window: window}
	if len(r.Steps) != 256 || len(r.Fits) != 8 || (clock != 0 && clock != 128 && clock != 224) || (window != 32 && window != 64) {
		return out, fmt.Errorf("sparse pilot dimensions")
	}
	for i := -16; i < clock; i++ {
		if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
			out.Origins = append(out.Origins, i)
		}
	}
	if len(out.Origins) > window {
		out.Origins = out.Origins[len(out.Origins)-window:]
	}
	w := 0
	if window == 32 {
		w = 1
	}
	if !reflect.DeepEqual(out.Origins, r.Fits[clock/32].Origins[w]) {
		return out, fmt.Errorf("sparse pilot origins")
	}
	samples := make([]observation.Sample, len(out.Origins))
	for j, i := range out.Origins {
		if i < 0 {
			samples[j] = r.Initial[i+16]
		} else {
			samples[j] = observation.Sample{Bits: r.Steps[i].X, Outcome: r.Steps[i].Y}
		}
	}
	f, e := fitSparseVRVMBudget(samples, limit)
	if e != nil {
		return out, e
	}
	out.UsedXi, out.FinalXi, out.Rates = f.gaussianXi, f.xi, f.rates
	out.Mean, out.Diagonal, out.PriorVariance = f.gaussian.mean, f.gaussian.diagonal, f.gaussian.priorVariance
	out.Trace, out.Stop, out.Residual, out.Logdet = f.trace, f.stop, f.gaussian.residual, f.gaussian.logdet
	for i := 0; i < 32; i++ {
		x := r.Steps[clock+i].X
		p, e := f.predict(x)
		if e != nil {
			return out, e
		}
		mu, v, e := f.gaussian.moments(x)
		if e != nil {
			return out, e
		}
		out.X[i] = x
		out.P[i] = [2]float64{p, math.Max(1e-12, math.Min(1-1e-12, ridgeSigmoid(mu)))}
		out.LogitMean[i], out.LogitVariance[i] = mu, v
	}
	// Evaluation-only fields: read after all forecasts from the frozen fit.
	for i := 0; i < 32; i++ {
		s := r.Steps[clock+i]
		out.Q[i], out.Y[i], out.Control[i] = s.Q, s.Y, s.P
	}
	return out, nil
}

func TestSparsePilotAsOf(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%3 == 0, Q: .3, Delay: []int{0, 7, 31}[i%3], Missing: i%5 == 0}
	}
	for c := 0; c < 256; c += 32 {
		fit := softV120Fit{Clock: c}
		var all []int
		for i := -16; i < c; i++ {
			if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= c) {
				all = append(all, i)
			}
		}
		for w, cap := range []int{64, 32} {
			v := all
			if len(v) > cap {
				v = v[len(v)-cap:]
			}
			fit.Origins[w] = append([]int(nil), v...)
		}
		r.Fits = append(r.Fits, fit)
	}
	base, e := runSparsePilot(r, 128, 32)
	if e != nil {
		t.Fatal(e)
	}
	poison := r
	poison.Steps = append([]softV120Step(nil), r.Steps...)
	for i := range poison.Steps {
		s := &poison.Steps[i]
		s.Q = math.NaN()
		s.P = [15]float64{math.NaN()}
		if i >= 128 || s.Missing || i+s.Delay > 128 {
			s.Y = !s.Y
		}
	}
	poison.Case = 999
	poison.Seeds = [5]int64{-1, -2, -3, -4, -5}
	poison.Rules = [2]uint16{511, 511}
	got, e := runSparsePilot(poison, 128, 32)
	if e != nil {
		t.Fatal(e)
	}
	if base.P != got.P || base.Mean != got.Mean || !reflect.DeepEqual(base.Trace, got.Trace) {
		t.Fatal("unavailable/evaluator evidence leak")
	}
	changed := r
	changed.Initial = r.Initial
	changed.Steps = append([]softV120Step(nil), r.Steps...)
	for _, i := range base.Origins {
		if i < 0 {
			changed.Initial[i+16].Outcome = !changed.Initial[i+16].Outcome
		} else {
			changed.Steps[i].Y = !changed.Steps[i].Y
		}
	}
	got, e = runSparsePilot(changed, 128, 32)
	if e != nil {
		t.Fatal(e)
	}
	if got.P == base.P {
		t.Fatal("admitted labels had no effect")
	}
	for i := range base.P {
		if math.Abs(base.P[i][0]+got.P[i][0]-1) > 1e-8 {
			t.Fatal("label complement symmetry")
		}
	}
	masks := []uint16{}
	f, e := fitSparseVRVM([]observation.Sample{{Bits: 0, Outcome: true}, {Bits: 7, Outcome: false}})
	if e != nil {
		t.Fatal(e)
	}
	masks = append(masks, f.gaussian.masks[:f.gaussian.p]...)
	precision := make([]float64, len(masks))
	for j := range precision {
		precision[j] = 1 / f.gaussian.priorVariance[j]
	}
	rebuilt, e := sparseVRVMGaussian([]observation.Sample{{Bits: 0, Outcome: true}, {Bits: 7, Outcome: false}}, masks, precision, f.gaussianXi)
	if e != nil {
		t.Fatal(e)
	}
	for j := range rebuilt.mean {
		if math.Abs(rebuilt.mean[j]-f.gaussian.mean[j]) > 1e-10 {
			t.Fatal("Gaussian snapshot reconstruction")
		}
	}
}

func TestSparsePilotCollect(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SPARSE_SOURCE")
	if path == "" {
		t.Skip("explicit research source")
	}
	dst := os.Getenv("EVENTFRAME_SPARSE_OUTPUT")
	if dst == "" {
		t.Fatal("output required")
	}
	in, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var header softV120Artifact
	if e = dec.Decode(&header); e != nil || header.Version != "soft-learners-v120" {
		t.Fatal("header", e)
	}
	count := 0
	for {
		var r softV120Record
		e = dec.Decode(&r)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		if r.Index != 0 {
			continue
		}
		for _, clock := range []int{0, 128, 224} {
			for _, window := range []int{64, 32} {
				v, e := runSparsePilot(r, clock, window)
				if e != nil {
					t.Fatalf("record%d: %v", count, e)
				}
				if e = enc.Encode(v); e != nil {
					t.Fatal(e)
				}
				count++
			}
		}
	}
	if count != 504 {
		t.Fatal("count", count)
	}
	if e = out.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("fits=%d forecasts=%d", count, count*32)
}
