package observationlearners

import (
	"math"
	"testing"
)

func TestConditionalEnumeration(t *testing.T) {
	f := NewForest(9)
	for i := 0; i < 400; i++ {
		f.Update(uint16(i%512), i%3 == 0)
	}
	var w, uniform [512]float64
	for x := range w {
		w[x] = float64(1 + x%11)
		uniform[x] = 1
	}
	m, err := NewConditionalForest(f, w)
	if err != nil {
		t.Fatal(err)
	}
	u, err := NewConditionalForest(f, uniform)
	if err != nil {
		t.Fatal(err)
	}
	for mask := uint16(0); mask < 512; mask++ {
		for value := mask; ; value = (value - 1) & mask {
			n, d := 0., 0.
			for x, weight := range w {
				if uint16(x)&mask == value {
					n += weight * f.Predict(uint16(x))
					d += weight
				}
			}
			got, e := m.Forecast(mask, value)
			if e != nil || math.Abs(got-n/d) > 1e-12 {
				t.Fatal("weighted enumeration", mask, value, got, n/d, e)
			}
			old, _ := f.ForecastPartial(mask, value)
			eq, _ := u.Forecast(mask, value)
			if math.Abs(old-eq) > 1e-12 {
				t.Fatal("uniform parity")
			}
			if mask == 511 && math.Abs(got-f.Predict(value)) > 1e-12 {
				t.Fatal("full observation changed")
			}
			if value == 0 {
				break
			}
		}
	}
	before, _ := m.Forecast(0, 0)
	f.Update(0, true)
	w[0] = 1e4
	after, _ := m.Forecast(0, 0)
	if before != after {
		t.Fatal("snapshot mutated")
	}
	for _, bad := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		w[0] = bad
		if _, e := NewConditionalForest(f, w); e == nil {
			t.Fatal("bad weight accepted")
		}
	}
	if _, e := m.Forecast(1, 2); e == nil {
		t.Fatal("unobserved value accepted")
	}
	if _, e := new(ConditionalForest).Forecast(0, 0); e == nil {
		t.Fatal("zero model accepted")
	}
}

func TestConditionalJournal(t *testing.T) {
	base, e := Base(Scenarios[2], 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunDependentStream(base, "design", 2, 0, 0, 2026104202, 2)
	if e != nil {
		t.Fatal(e)
	}
	a, e := RunConditionalRecord(r)
	if e != nil {
		t.Fatal(e)
	}
	if len(a.Ticks) == 0 {
		t.Fatal("empty diagnostic")
	}
	// The final outcome is not available for its own prediction, even if its
	// feedback happens to trigger a refit after that prediction.
	r.Ticks[511].Outcome = !r.Ticks[511].Outcome
	b, e := RunConditionalRecord(r)
	if e != nil {
		t.Fatal(e)
	}
	for i := range a.Ticks {
		if a.Ticks[i].Predictions != b.Ticks[i].Predictions {
			t.Fatal("current label leaked")
		}
	}
	for g := 0; g < 3; g++ {
		sum := 0.
		for _, v := range conditionalOracle(g) {
			sum += v
		}
		if math.Abs(sum-1) > 1e-12 {
			t.Fatal("oracle normalization")
		}
	}
}

var conditionalBenchmarkSink float64

func BenchmarkConditionalLookup(b *testing.B) {
	f := NewForest(3)
	var w [512]float64
	for x := range w {
		w[x] = 1
	}
	m, e := NewConditionalForest(f, w)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, e := m.Forecast(63, uint16(i)&63)
		if e != nil {
			b.Fatal(e)
		}
		conditionalBenchmarkSink = p
	}
}

func BenchmarkConditionalBuild(b *testing.B) {
	f := NewForest(3)
	var w [512]float64
	for x := range w {
		w[x] = 1
	}
	for i := 0; i < 256; i++ {
		f.Update(uint16(i), i%3 == 0)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := NewConditionalForest(f, w); e != nil {
			b.Fatal(e)
		}
	}
}
