package observationlearners

import (
	"math"
	"math/bits"
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type boundedReader struct {
	x     uint16
	bad   bool
	reads []observation.View
}

func (r *boundedReader) Epoch() uint64 { return 1 }
func (r *boundedReader) Read(v observation.View) (uint16, uint16, error) {
	r.reads = append(r.reads, v)
	if r.bad {
		return 511, r.x, nil
	}
	return v.Mask(), r.x & v.Mask(), nil
}
func fittedTestForest() *Forest {
	f := NewForest(41)
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 512; i++ {
		x := uint16(r.Intn(512))
		f.Update(x, x&4 != 0)
	}
	return f
}
func TestPartialMarginal(t *testing.T) {
	f := fittedTestForest()
	for mask := uint16(0); mask < 512; mask++ {
		values := uint16(301) & mask
		got, e := f.ForecastPartial(mask, values)
		if e != nil {
			t.Fatal(e)
		}
		want, n := 0., 0
		for x := uint16(0); x < 512; x++ {
			if x&mask == values {
				want += f.Predict(x)
				n++
			}
		}
		if math.Abs(got-want/float64(n)) > 1e-12 {
			t.Fatal("marginal used hidden values")
		}
	}
	if _, e := f.ForecastPartial(1, 2); e == nil {
		t.Fatal("accepted hidden bits")
	}
}
func TestPartialObserverHiddenInvariance(t *testing.T) {
	f := fittedTestForest()
	for x := uint16(0); x < 512; x += 17 {
		r := &boundedReader{x: x}
		a, e := RunForestObserver(f, r, 1)
		if e != nil {
			t.Fatal(e)
		}
		mask := a.Trace[len(a.Trace)-1].Observed
		if a.Cost > 6 || bits.OnesCount16(mask) > 6 || len(r.reads) != len(a.Trace) {
			t.Fatal("observation budget")
		}
		other := &boundedReader{x: x ^ (511 &^ mask)}
		b, e := RunForestObserver(f, other, 1)
		if e != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("unobserved value affected decision")
		}
	}
	if _, e := RunForestObserver(f, &boundedReader{bad: true}, 1); e == nil {
		t.Fatal("reader leaked undeclared fields")
	}
}
func TestPartialStreamBudgetAndAvailability(t *testing.T) {
	j := 7
	b, e := Base(Scenarios[j], j, 0)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunPartialStream(b, "unit", j, 0, 0, 2026092901)
	if e != nil {
		t.Fatal(e)
	}
	for i, v := range r.Views {
		for a, view := range v {
			if view.Cost > 6 || view.Observed > 6 || len(view.Trace) == 0 {
				t.Fatal("invalid live budget")
			}
			if r.Ticks[i].Predictions[a].P <= 0 || r.Ticks[i].Predictions[a].P >= 1 {
				t.Fatal("invalid law")
			}
		}
		for _, origin := range r.Ticks[i].Delivered {
			if origin+16 != i || r.Ticks[origin].Missing {
				t.Fatal("feedback leak")
			}
		}
	}
}
