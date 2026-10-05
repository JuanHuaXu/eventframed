package researchretentionlaw

import (
	retention "github.com/JuanHuaXu/eventframed/internal/researchretention"
	"math"
	"testing"
)

func fixture() [3]Joint {
	var out [3]Joint
	for k, q := range [3]float64{.15, .5, .85} {
		for y := 0; y < 2; y++ {
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					mass := q
					if y == 0 {
						mass = 1 - q
					}
					one, two := .1, .1
					if a == y {
						one = .9
					}
					if b == y {
						two = .9
					}
					out[k][4*y+2*a+b] = mass * one * two
				}
			}
		}
	}
	return out
}
func closeValue(t *testing.T, a, b float64) {
	t.Helper()
	if math.Abs(a-b) > 2e-12 {
		t.Fatal(a, b)
	}
}
func TestSameJointCleanAndMeasurementMarginals(t *testing.T) {
	joints := fixture()
	f, e := Forecasts(joints)
	if e != nil {
		t.Fatal(e)
	}
	for k, q := range [3]float64{.15, .5, .85} {
		closeValue(t, f[k].Clean, q)
		closeValue(t, f[k].Joint[1], .09)
		closeValue(t, f[k].Joint[2], .09)
		closeValue(t, f[k].Joint[2]+f[k].Joint[3], .1+.8*q)
	}
	s, e := retention.New(1, 1)
	if e != nil {
		t.Fatal(e)
	}
	one, e := s.Issue(0, 0, f)
	if e != nil {
		t.Fatal(e)
	}
	closeValue(t, one.Forecast(), .5)
	receipt, e := s.Resolve(one, true, 1)
	if e != nil {
		t.Fatal(e)
	}
	closeValue(t, receipt.Forecast, .5)
	second, e := s.RequestSecond(one, 2)
	if e != nil {
		t.Fatal(e)
	}
	closeValue(t, second.Forecast(), .82)
	if _, e = s.Resolve(second, false, 3); e != nil {
		t.Fatal(e)
	}
	w, e := s.Weights(0)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range w {
		closeValue(t, p, 1./3)
	}
}
func TestInvalidAndZeroSupportJoints(t *testing.T) {
	good := fixture()
	for fault := 0; fault < 5; fault++ {
		bad := good
		switch fault {
		case 0:
			bad[0][0] = math.NaN()
		case 1:
			bad[0][0] = math.Inf(1)
		case 2:
			bad[1][0] = -.01
		case 3:
			bad[2][0] += .01
		case 4:
			bad[2] = Joint{}
		}
		f, e := Forecasts(bad)
		if e == nil || f != (retention.Forecasts{}) {
			t.Fatal("partial invalid publication", fault)
		}
	}
	for _, index := range []int{0, 3, 4, 7} {
		var j [3]Joint
		for k := range j {
			j[k][index] = 1
		}
		f, e := Forecasts(j)
		if e != nil {
			t.Fatal(e)
		}
		want := 0.
		if index >= 4 {
			want = 1
		}
		for _, x := range f {
			closeValue(t, x.Clean, want)
			closeValue(t, x.Joint[index%4], 1)
		}
	}
}
func TestLatentLawNotAuthenticatedTruth(t *testing.T) {
	// Two different declared latent laws may yield exactly the same observed law.
	// This adapter proves internal coherence, not identification of physical truth.
	var first, second [3]Joint
	for k := range first {
		first[k][0] = 1
		second[k][4] = 1
	}
	a, e := Forecasts(first)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Forecasts(second)
	if e != nil {
		t.Fatal(e)
	}
	for k := range a {
		if a[k].Joint != b[k].Joint {
			t.Fatal("invalid nonidentification control")
		}
		if a[k].Clean == b[k].Clean {
			t.Fatal("vacuous latent control")
		}
	}
}
