package observationlearners

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestSpectralAsOfIsolation(t *testing.T) {
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
	base, e := runSpectral(r)
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
	got, e := runSpectral(poison)
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
	got, e = runSpectral(future)
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

func BenchmarkSpectralPredict(b *testing.B) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%3 == 0}
	}
	for mode := 0; mode < 3; mode++ {
		m, e := fitSpectral(s, mode)
		if e != nil {
			b.Fatal(e)
		}
		b.Run([]string{"linear", "screened", "full"}[mode], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, e := m.predict(uint16(i % 512)); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
