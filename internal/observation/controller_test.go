package observation

import (
	"sync"
	"testing"
)

type testReader struct {
	value   uint16
	version uint64
	drift   bool
	illegal bool
}

func (r *testReader) Epoch() uint64 { return r.version }
func (r *testReader) Read(v View) (uint16, uint16, error) {
	if r.drift {
		r.version++
	}
	if r.illegal {
		return 511, 511, nil
	}
	return v.Mask(), r.value & v.Mask(), nil
}
func fitted(t *testing.T) *Model {
	t.Helper()
	samples := make([]Sample, Universe)
	for i := range samples {
		samples[i] = Sample{uint16(i), i&256 != 0}
	}
	m, err := Fit(samples)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestAdaptiveViewUsesTrainingNotHiddenValues(t *testing.T) {
	m := fitted(t)
	a, err := Run(m, &testReader{value: 0, version: 1}, 1, "mmm", 7)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Run(m, &testReader{value: 256, version: 1}, 1, "mmm", 7)
	if err != nil {
		t.Fatal(err)
	}
	if a.Trace[1].View != (View{2, 2}) || a.Trace[1].View != b.Trace[1].View {
		t.Fatal("wrong or outcome-aware selection")
	}
	if a.Probability >= .1 || b.Probability <= .9 {
		t.Fatal("did not use inspected mechanism/history")
	}
	if a.Cost > 6 || b.Cost > 6 || m.Support() != Universe {
		t.Fatal("budget/support changed")
	}
}
func TestAsOfEpochAndReaderBoundary(t *testing.T) {
	m := fitted(t)
	for _, r := range []*testReader{{version: 2}, {version: 1, drift: true}, {version: 1, illegal: true}} {
		if _, err := Run(m, r, 1, "mmm", 0); err == nil {
			t.Fatal("accepted invalid snapshot/reader")
		}
	}
	if _, err := Run(m, &testReader{version: 1}, 1, "bad", 0); err == nil {
		t.Fatal("accepted invalid policy")
	}
}
func TestPoliciesBoundedAndImmutableConcurrent(t *testing.T) {
	m := fitted(t)
	before := *m
	var wg sync.WaitGroup
	for _, p := range Policies {
		wg.Add(1)
		go func(policy string) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				r, err := Run(m, &testReader{value: uint16(i), version: 1}, 1, policy, int64(i))
				if err != nil {
					t.Error(err)
					return
				}
				cap := 6
				if policy == "exhaustive" {
					cap = 9
				}
				if r.Cost > cap || len(r.Trace) > 9 {
					t.Error("unbounded job")
				}
			}
		}(p)
	}
	wg.Wait()
	if m.Support() != Universe {
		t.Fatal("support changed")
	}
	if *m != before {
		t.Fatal("conditional fitting table mutated during inference")
	}
}
func TestInvalidFittingAndMissingIsUnknown(t *testing.T) {
	if _, err := Fit(nil); err == nil {
		t.Fatal("empty accepted")
	}
	if _, err := Fit([]Sample{{Bits: Universe}}); err == nil {
		t.Fatal("invalid bits accepted")
	}
	m := fitted(t)
	if m.predict(0, 0) != .5 {
		t.Fatal("unknown confused with zero")
	}
}
