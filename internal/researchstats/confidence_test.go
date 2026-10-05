package researchstats

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func logJSONRecord(t *testing.T, kind string, record map[string]any) {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s=%s", kind, raw)
}

func newSequence(t *testing.T, alpha float64, names ...string) *ConfidenceSequence {
	t.Helper()
	s, err := NewConfidenceSequence(alpha, names)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func near(t *testing.T, got, want float64) {
	t.Helper()
	if math.IsNaN(got) || math.Abs(got-want) > 1e-13*math.Max(1, math.Abs(want)) {
		t.Fatalf("got %.17g, want %.17g", got, want)
	}
}

func TestInvalidConfiguration(t *testing.T) {
	for _, alpha := range []float64{-1, 0, 1, 2, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := NewConfidenceSequence(alpha, []string{"gain"}); err == nil {
			t.Errorf("accepted alpha %g", alpha)
		}
	}
	for _, names := range [][]string{nil, {}, {""}, {" \t"}, {"gain", "gain"}} {
		if _, err := NewConfidenceSequence(.05, names); err == nil {
			t.Errorf("accepted names %q", names)
		}
	}
}

func TestFrozenFamily(t *testing.T) {
	names := []string{"member", "stable"}
	s := newSequence(t, .05, names...)
	names[0] = "changed"
	alpha, copyNames, err := s.Configuration()
	if err != nil || alpha != .05 || !reflect.DeepEqual(copyNames, []string{"member", "stable"}) {
		t.Fatalf("configuration changed: %g %q %v", alpha, copyNames, err)
	}
	copyNames[1] = "also changed"
	_, again, err := s.Configuration()
	if err != nil || again[1] != "stable" {
		t.Fatal("configuration read exposes internal names")
	}
	if err := s.AddStream("changed", .1); err == nil {
		t.Fatal("added comparison outside frozen family")
	}
	for _, read := range []func(string) (Interval, error){s.Anytime, s.FixedSample} {
		if _, err := read("unknown"); err == nil {
			t.Fatal("unknown comparison returned an interval")
		}
	}
}

func TestUninitializedSequence(t *testing.T) {
	for _, s := range []*ConfidenceSequence{nil, {}} {
		if err := s.AddStream("gain", 0); err == nil {
			t.Fatal("uninitialized addition accepted")
		}
		if _, _, err := s.Configuration(); err == nil {
			t.Fatal("uninitialized configuration accepted")
		}
		if _, err := s.Anytime("gain"); err == nil {
			t.Fatal("uninitialized CS accepted")
		}
		if _, err := s.FixedSample("gain"); err == nil {
			t.Fatal("uninitialized fixed-sample interval accepted")
		}
		if err := s.AddPairedStream("gain", []BrierPair{{}}); err == nil {
			t.Fatal("uninitialized paired stream accepted")
		}
	}
}

func TestEmptyAndAtomicInvalidUpdates(t *testing.T) {
	s := newSequence(t, .05, "gain", "guard")
	empty := Interval{Radius: 2, Lower: -1, Upper: 1}
	for _, read := range []func(string) (Interval, error){s.Anytime, s.FixedSample} {
		got, err := read("gain")
		if err != nil || got != empty {
			t.Fatalf("empty interval: %+v %v", got, err)
		}
	}
	if err := s.AddStream("gain", .25); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Anytime("gain")
	for _, x := range []float64{math.Nextafter(-1, -2), math.Nextafter(1, 2), math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := s.AddStream("gain", x); err == nil {
			t.Fatalf("accepted invalid stream %g", x)
		}
		after, _ := s.Anytime("gain")
		if after != before {
			t.Fatalf("failed update mutated state: %+v -> %+v", before, after)
		}
	}
	guard, _ := s.Anytime("guard")
	if guard != empty {
		t.Fatal("one comparison changed another")
	}
}

func TestHoeffdingFormulasAndFamilyMultiplicity(t *testing.T) {
	s := newSequence(t, .05, "a", "b", "c")
	for _, x := range []float64{-1, 1, .5, .5} {
		if err := s.AddStream("a", x); err != nil {
			t.Fatal(err)
		}
	}
	cs, err := s.Anytime("a")
	if err != nil || cs.N != 4 {
		t.Fatalf("CS count/error: %+v %v", cs, err)
	}
	near(t, cs.Mean, .25)
	near(t, cs.Radius, math.Sqrt(2*math.Log(2*3*4*5/.05)/4))
	near(t, cs.Lower, math.Max(-1, cs.Mean-cs.Radius))
	near(t, cs.Upper, math.Min(1, cs.Mean+cs.Radius))
	fixed, err := s.FixedSample("a")
	if err != nil {
		t.Fatal(err)
	}
	near(t, fixed.Radius, math.Sqrt(2*math.Log(2*3/.05)/4))
	if fixed.N != cs.N || fixed.Mean != cs.Mean || fixed.Radius >= cs.Radius {
		t.Fatal("fixed-sample bound was not kept distinct")
	}
	one := newSequence(t, .05, "a")
	for i := 0; i < 4; i++ {
		if err := one.AddStream("a", .25); err != nil {
			t.Fatal(err)
		}
	}
	individual, _ := one.Anytime("a")
	if individual.Radius >= cs.Radius {
		t.Fatal("larger family did not widen radius")
	}
	lowerAlpha := newSequence(t, .01, "a", "b", "c")
	for i := 0; i < 4; i++ {
		if err := lowerAlpha.AddStream("a", .25); err != nil {
			t.Fatal(err)
		}
	}
	conservative, _ := lowerAlpha.Anytime("a")
	if conservative.Radius <= cs.Radius {
		t.Fatal("smaller alpha did not widen radius")
	}
}

func TestSpendingIdentity(t *testing.T) {
	s := newSequence(t, .05, "a", "b")
	for _, n := range []uint64{1, 2, 32, 1024, 1000000, maxUnits} {
		s.states["a"].n = n
		cs, err := s.Anytime("a")
		if err != nil {
			t.Fatal(err)
		}
		// The two-tail bound is exactly the allocated look budget in real arithmetic.
		logFailure := math.Log(2) - float64(n)*cs.Radius*cs.Radius/2
		want := math.Log(.05) - math.Log(2) - math.Log(float64(n)) - math.Log1p(float64(n))
		near(t, logFailure, want)
	}
	spent := 0.
	for n := 1.; n <= 10000; n++ {
		spent += 1 / (n * (n + 1))
	}
	near(t, spent, 1-1./10001)
}

func TestAppendixCFormula(t *testing.T) {
	for _, k := range []int{1, 3, 100} {
		names := make([]string, k)
		for c := range names {
			names[c] = fmt.Sprintf("comparison-%d", c)
		}
		for _, alphaFamily := range []float64{.01, .05} {
			s := newSequence(t, alphaFamily, names...)
			for _, j := range []uint64{1, 2, 32, 4096, 1000000} {
				s.states[names[0]].n = j
				interval, err := s.Anytime(names[0])
				if err != nil {
					t.Fatal(err)
				}
				alphaJC := alphaFamily / (float64(k) * float64(j) * (float64(j) + 1))
				near(t, interval.Radius, math.Sqrt(2*math.Log(2/alphaJC)/float64(j)))
			}
		}
	}
	t.Log("Appendix C match: K={1,3,100}, alphaFamily={0.01,0.05}, j={1,2,32,4096,1000000}; no betting method")
}

func TestNumericalEdgesAndCountLimit(t *testing.T) {
	for _, alpha := range []float64{math.SmallestNonzeroFloat64, math.Nextafter(1, 0)} {
		s := newSequence(t, alpha, "gain")
		s.states["gain"].n = maxUnits - 1
		if err := s.AddStream("gain", 1); err != nil {
			t.Fatal(err)
		}
		before, err := s.Anytime("gain")
		if err != nil || before.N != maxUnits || !finite(before.Radius) || before.Radius <= 0 {
			t.Fatalf("unstable extreme interval: %+v %v", before, err)
		}
		if err := s.AddStream("gain", 0); err == nil {
			t.Fatal("count beyond supported range accepted")
		}
		after, _ := s.Anytime("gain")
		if after != before {
			t.Fatal("overflow rejection changed state")
		}
	}
	s := newSequence(t, .05, "gain")
	for _, x := range []float64{1, 1e-16, -1} {
		if err := s.AddStream("gain", x); err != nil {
			t.Fatal(err)
		}
	}
	i, _ := s.Anytime("gain")
	if math.Abs(i.Mean-1e-16/3) > 1e-30 {
		t.Fatalf("lost small paired effect: %.17g", i.Mean)
	}
}

func TestMeanAndClippedSupport(t *testing.T) {
	if (meanState{}).mean() != 0 {
		t.Fatal("empty mean placeholder changed")
	}
	for _, x := range []float64{-1, 1} {
		s := newSequence(t, .05, "endpoint")
		for n := 0; n < 4096; n++ {
			if err := s.AddStream("endpoint", x); err != nil {
				t.Fatal(err)
			}
		}
		for _, read := range []func(string) (Interval, error){s.Anytime, s.FixedSample} {
			i, err := read("endpoint")
			if err != nil || i.Mean != x || i.Lower < -1 || i.Upper > 1 || i.Lower > x || i.Upper < x {
				t.Fatalf("endpoint not retained inside clipped support: %+v %v", i, err)
			}
			if x == -1 && i.Lower != -1 || x == 1 && i.Upper != 1 {
				t.Fatal("endpoint clipping failed")
			}
		}
	}
}

func TestSyntheticSequences(t *testing.T) {
	s := newSequence(t, .05, "benefit", "null", "harm")
	for n := 1; n <= 4096; n++ {
		for name, x := range map[string]float64{"benefit": .25, "null": 0, "harm": -.25} {
			if err := s.AddStream(name, x); err != nil {
				t.Fatal(err)
			}
			i, err := s.Anytime(name)
			if err != nil || i.N != uint64(n) || i.Lower > x || i.Upper < x {
				t.Fatalf("constant-mean synthetic path excluded truth: %+v %v", i, err)
			}
			near(t, i.Mean, x)
		}
	}
	benefit, _ := s.Anytime("benefit")
	null, _ := s.Anytime("null")
	harm, _ := s.Anytime("harm")
	if benefit.Lower <= 0 || harm.Upper >= 0 || null.Lower >= 0 || null.Upper <= 0 {
		t.Fatal("synthetic sign controls failed")
	}
}
