package observationlearners

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestSpikeCadenceClockAsOf(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256), Fits: make([]softV120Fit, 8)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%3 == 0, Delay: i % 16, Missing: i%5 == 0}
	}
	for _, start := range []int{0, 32, 64, 96, 128, 160, 192, 224} {
		r.Fits[start/32].Origins[0] = spikeCadenceOrigins(r, start)
		base, err := runSpikeCadenceAt(r, start)
		if err != nil {
			t.Fatal(err)
		}
		if (start == 128 && base.Clock != nil) || (start != 128 && (base.Clock == nil || *base.Clock != start)) {
			t.Fatal("ambiguous clock")
		}
		frozen, err := runSpikePilotAt(r, start)
		if err != nil || base.P[0] != frozen.P[0][0] || !reflect.DeepEqual(base.Fits[0].Trace, frozen.Trace) {
			t.Fatal("initial frozen mismatch", err)
		}
		poison := r
		poison.Steps = append([]softV120Step(nil), r.Steps...)
		for i := range poison.Steps {
			s := &poison.Steps[i]
			s.Q = math.NaN()
			s.P = [15]float64{math.NaN()}
			if i >= start+8 || s.Missing || i+s.Delay > start+8 {
				s.Y = !s.Y
			}
		}
		got, err := runSpikeCadenceAt(poison, start)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= 8; i++ {
			if got.P[i] != base.P[i] || got.Used[i] != base.Used[i] {
				t.Fatal("unavailable label leak", start, i)
			}
		}
		if !reflect.DeepEqual(base.Fits[:base.Used[8]+1], got.Fits[:got.Used[8]+1]) {
			t.Fatal("fit prefix leak")
		}
		if base.P == got.P {
			t.Fatal("available future changes had no effect")
		}
	}
	for _, bad := range []int{-1, 1, 225, 256} {
		if _, err := runSpikeCadenceAt(r, bad); err == nil {
			t.Fatal("invalid clock accepted", bad)
		}
	}
}
