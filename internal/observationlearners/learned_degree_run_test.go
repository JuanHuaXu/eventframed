package observationlearners

import (
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
)

// Deliberately detached copy of the frozen replay adapter; only the fitter changes.
type learnedDegreeRecord struct {
	spectralRecord
	Optimization [8][4]learnedDegreeStats
}

func runLearnedDegree(r softV120Record) (learnedDegreeRecord, error) {
	out := learnedDegreeRecord{spectralRecord: spectralRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}}
	if len(r.Steps) != 256 || len(r.Fits) != 8 {
		return out, fmt.Errorf("source shape")
	}
	var models [6]*spectralModel
	for clock, s := range r.Steps {
		if clock%32 == 0 {
			var available []int
			for i := -16; i < clock; i++ {
				if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
					available = append(available, i)
				}
			}
			for w, cap := range []int{64, 32} {
				origins := available
				if len(origins) > cap {
					origins = origins[len(origins)-cap:]
				}
				if !reflect.DeepEqual(origins, r.Fits[clock/32].Origins[w]) {
					return out, fmt.Errorf("as-of mismatch")
				}
				samples := make([]observation.Sample, len(origins))
				for j, i := range origins {
					if i < 0 {
						samples[j] = r.Initial[i+16]
					} else {
						samples[j] = observation.Sample{Bits: r.Steps[i].X, Outcome: r.Steps[i].Y}
					}
				}
				for mode := 0; mode < 3; mode++ {
					var m *spectralModel
					var e error
					if mode == 0 {
						m, e = fitSpectral(samples, 0)
					} else {
						var stats learnedDegreeStats
						m, stats, e = fitLearnedDegree(samples, mode == 2)
						out.Optimization[clock/32][2*(mode-1)+w] = stats
					}
					if e != nil {
						return out, e
					}
					k := 2*mode + w
					models[k] = m
					out.MaxResidual = math.Max(out.MaxResidual, m.Residual)
					out.FeatureTotal[k] += len(m.Features)
					out.Fits++
				}
			}
		}
		for k, m := range models {
			p, e := m.predict(s.X)
			if e != nil {
				return out, e
			}
			out.P[clock][k] = p
		}
	}
	return out, nil
}

func TestLearnedDegreeRun(t *testing.T) {
	path := os.Getenv("EVENTFRAME_LEARNED_DEGREE_SOURCE")
	if path == "" {
		t.Skip("explicit research source required")
	}
	dst := os.Getenv("EVENTFRAME_LEARNED_DEGREE_OUTPUT")
	if dst == "" {
		t.Fatal("output required")
	}
	in, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var header softV120Artifact
	if e = dec.Decode(&header); e != nil || header.Version != "soft-learners-v120" {
		t.Fatalf("source header: %v", e)
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
		v, e := runLearnedDegree(r)
		if e != nil {
			t.Fatalf("record%d: %v", count, e)
		}
		if e = enc.Encode(v); e != nil {
			t.Fatal(e)
		}
		count++
	}
	if count != 2688 {
		t.Fatalf("records%d", count)
	}
	if e = out.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("records=%d fits=%d", count, count*48)
}

func TestLearnedDegreeAsOfIsolation(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%4 < 2, Q: .3, Delay: []int{0, 7, 31}[i%3], Missing: i%5 == 0}
	}
	for clock := 0; clock < 256; clock += 32 {
		fit := softV120Fit{Clock: clock}
		all := make([]int, 0, 272)
		for i := -16; i < clock; i++ {
			if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
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
	base, e := runLearnedDegree(r)
	if e != nil {
		t.Fatal(e)
	}
	poison := r
	poison.Steps = append([]softV120Step(nil), r.Steps...)
	poison.Rules = [2]uint16{511, 511}
	poison.Seeds = [5]int64{-1, -2, -3, -4, -5}
	for i := range poison.Steps {
		poison.Steps[i].Q = math.NaN()
		poison.Steps[i].P = [15]float64{math.NaN()}
		if poison.Steps[i].Missing {
			poison.Steps[i].Y = !poison.Steps[i].Y
		}
	}
	got, e := runLearnedDegree(poison)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(base.P, got.P) {
		t.Fatal("unavailable labels or evaluator metadata affected forecasts")
	}
	future := r
	future.Steps = append([]softV120Step(nil), r.Steps...)
	for i := 128; i < 256; i++ {
		future.Steps[i].Y = !future.Steps[i].Y
	}
	got, e = runLearnedDegree(future)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 160; i++ {
		if base.P[i] != got.P[i] {
			t.Fatalf("future label affected pre-publication forecast %d", i)
		}
	}
	if reflect.DeepEqual(base.P, got.P) {
		t.Fatal("eligible later evidence never affected forecast")
	}
}
