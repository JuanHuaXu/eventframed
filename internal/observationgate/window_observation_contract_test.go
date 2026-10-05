package observationgate

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

// This counterexample prevents treating mask-local Laplace estimates as the
// marginals of the subset expert's uniform-input joint distribution. The count
// estimator is unchanged; it may still provide a declared acquisition proxy.
func TestWindowCountJointBoundary(t *testing.T) {
	m, e := observation.Fit([]observation.Sample{{Bits: 0, Outcome: true}, {Bits: 0, Outcome: true}})
	if e != nil {
		t.Fatal(e)
	}
	parent, _ := m.ForecastObserved(0, 0)
	left, _ := m.ForecastObserved(1, 0)
	right, _ := m.ForecastObserved(1, 1)
	average := (left + right) / 2
	if parent != .75 || left != .75 || right != .5 || average != .625 {
		t.Fatal("counterexample changed", parent, left, right, average)
	}
	t.Logf("count uniform marginal defect=%g", parent-average)
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	coherent, e := observationlearners.NewSubsetConditional([]observation.Sample{{Bits: 0, Outcome: true}, {Bits: 0, Outcome: true}}, weights)
	if e != nil {
		t.Fatal(e)
	}
	for mask := uint16(0); mask < 512; mask++ {
		for values := uint16(0); values < 512; values++ {
			if values&^mask != 0 {
				continue
			}
			p, e := coherent.Forecast(mask, values)
			if e != nil {
				t.Fatal(e)
			}
			for bit := uint16(1); bit < 512; bit <<= 1 {
				if bit&mask == 0 {
					l, _ := coherent.Forecast(mask|bit, values)
					r, _ := coherent.Forecast(mask|bit, values|bit)
					if math.Abs(p-(l+r)/2) > 1e-12 {
						t.Fatal("coherent control failed", mask, values, bit)
					}
				}
			}
		}
	}
}

type windowGuideModels struct {
	Base, LabelCount, EventCount, Long *observation.Model
	LabelSubset, EventSubset           *observationlearners.ConditionalForest
}
type windowGuideAdvice struct {
	Guide                     string
	Mask, Values              uint16
	Cost                      int
	P                         float64
	OldInner, NewInner, Outer [4]float64
}

func windowGuideWeights(m bayes.ForecastMix) [4]float64 {
	if m.Weights == [4]float64{} {
		return [4]float64{.7, .1, .1, .1}
	}
	return m.Weights
}

// The chosen model is explicitly an acquisition proxy. The scored law is the
// complete nested forecast evaluated on the mask actually acquired by it.
// No hypothetical inspection updates weights or contributes training evidence.
func windowGuidePredict(models windowGuideModels, oldInner, newInner, outer bayes.ForecastMix, reader observation.Reader, seed int64) (windowGuideAdvice, error) {
	var a windowGuideAdvice
	if models.Base == nil || reader == nil {
		return a, fmt.Errorf("missing base model or reader")
	}
	for _, mix := range []bayes.ForecastMix{oldInner, newInner, outer} {
		if mix.Weights == [4]float64{} {
			continue
		}
		sum := 0.
		for _, w := range mix.Weights {
			if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
				return a, fmt.Errorf("invalid guide weights")
			}
			sum += w
		}
		if math.Abs(sum-1) > 1e-9 {
			return a, fmt.Errorf("unnormalized guide weights")
		}
	}
	ow, nw, iw := windowGuideWeights(outer), windowGuideWeights(newInner), windowGuideWeights(oldInner)
	guide := 0
	if models.LabelCount != nil && ow[1] > ow[guide] {
		guide = 1
	}
	if models.Long != nil && ow[2] > ow[guide] {
		guide = 2
	}
	count, subset := models.Base, (*observationlearners.ConditionalForest)(nil)
	a.Guide = "base"
	if guide == 2 {
		count = models.Long
		a.Guide = "long"
	}
	if guide == 1 {
		count = models.LabelCount
		a.Guide = "label-count"
		if models.LabelSubset != nil && iw[1]+iw[2]+iw[3] > iw[0] {
			subset = models.LabelSubset
			a.Guide = "label-subset"
		}
		branch, weight := 0, nw[0]
		if models.EventCount != nil && nw[1] > weight {
			branch = 1
			weight = nw[1]
		}
		if models.EventSubset != nil && nw[2]+nw[3] > weight {
			branch = 2
		}
		if branch == 1 {
			count = models.EventCount
			subset = nil
			a.Guide = "event-count"
		}
		if branch == 2 {
			subset = models.EventSubset
			a.Guide = "event-subset"
		}
	}
	var r observation.Result
	var err error
	if subset != nil {
		r, err = observationlearners.RunConditionalObserver(subset, reader, reader.Epoch())
	} else {
		r, err = observation.Run(count, reader, reader.Epoch(), "mmm", seed)
	}
	if err != nil {
		return a, err
	}
	last := r.Trace[len(r.Trace)-1]
	a.Mask, a.Values, a.Cost = last.Observed, last.Values, r.Cost
	a.Outer = [4]float64{.5, .5, .5, .5}
	a.OldInner = [4]float64{.5, .5, .5, .5}
	for i, m := range []*observation.Model{models.Base, models.LabelCount, models.Long} {
		if m != nil {
			a.Outer[i], err = m.ForecastObserved(a.Mask, a.Values)
			if err != nil {
				return a, err
			}
		}
	}
	a.OldInner[0] = a.Outer[1]
	old := a.Outer[1]
	if models.LabelCount != nil && models.LabelSubset != nil {
		p, e := models.LabelSubset.Forecast(a.Mask, a.Values)
		if e != nil {
			return a, e
		}
		a.OldInner[1], a.OldInner[2], a.OldInner[3] = p, p, p
		old = oldInner.Forecast(a.OldInner)
	}
	a.NewInner = [4]float64{old, .5, .5, .5}
	if models.EventCount != nil {
		a.NewInner[1], err = models.EventCount.ForecastObserved(a.Mask, a.Values)
		if err != nil {
			return a, err
		}
	}
	if models.EventSubset != nil {
		p, e := models.EventSubset.Forecast(a.Mask, a.Values)
		if e != nil {
			return a, e
		}
		a.NewInner[2], a.NewInner[3] = p, p
	}
	a.Outer[1] = newInner.Forecast(a.NewInner)
	a.P = outer.Forecast(a.Outer)
	return a, nil
}

func TestWindowGuideContracts(t *testing.T) {
	frames := make([]innerArrivalFrame, 64)
	ids := make([]int, 64)
	for i := range frames {
		frames[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
		ids[i] = i
	}
	m, e := fitEventWindow(frames, ids)
	if e != nil {
		t.Fatal(e)
	}
	models := windowGuideModels{m.count, m.count, m.count, m.count, m.subset, m.subset}
	countBefore, subsetBefore := *m.count, *m.subset
	zero := bayes.ForecastMix{}
	focus := func(i int) bayes.ForecastMix { var m bayes.ForecastMix; m.Weights[i] = 1; return m }
	tests := []struct {
		old, new, outer bayes.ForecastMix
		want            string
	}{
		{zero, zero, focus(0), "base"},
		{zero, zero, focus(2), "long"},
		{focus(0), focus(0), focus(1), "label-count"},
		{focus(1), focus(0), focus(1), "label-subset"},
		{focus(1), focus(1), focus(1), "event-count"},
		{focus(0), focus(2), focus(1), "event-subset"},
	}
	for _, tt := range tests {
		for x := uint16(0); x < 512; x++ {
			before := tt
			a, e := windowGuidePredict(models, tt.old, tt.new, tt.outer, observationexperiment.Frames(x, "guide"), 0)
			if e != nil || a.Guide != tt.want || a.Cost > 6 || a.Values != x&a.Mask || a.P < 0 || a.P > 1 {
				t.Fatal("guide contract", tt.want, x, a, e)
			}
			if !reflect.DeepEqual(before, tt) || a.P != tt.outer.Forecast(a.Outer) || a.Outer[1] != tt.new.Forecast(a.NewInner) {
				t.Fatal("mutated state or wrong law")
			}
		}
	}
	// Unavailable refreshed models must not cause unrecorded full-text fallback.
	models.EventCount = nil
	models.EventSubset = nil
	a, e := windowGuidePredict(models, focus(0), focus(2), focus(1), observationexperiment.Frames(0, "missing"), 0)
	if e != nil || a.Guide != "label-count" || a.NewInner[1] != .5 || a.NewInner[2] != .5 {
		t.Fatal("missing model fallback", a, e)
	}
	if !reflect.DeepEqual(countBefore, *m.count) || !reflect.DeepEqual(subsetBefore, *m.subset) {
		t.Fatal("observation mutated a published model")
	}
	for _, bad := range [][4]float64{{-1, 1, 1, 0}, {math.NaN(), 0, 0, 0}, {.1, .1, .1, .1}} {
		if _, e := windowGuidePredict(models, bayes.ForecastMix{Weights: bad}, zero, zero, observationexperiment.Frames(0, "invalid"), 0); e == nil {
			t.Fatal("invalid weight state accepted")
		}
	}
}

var windowGuideBenchSink windowGuideAdvice

func BenchmarkWindowGuidePredict(b *testing.B) {
	frames := make([]innerArrivalFrame, 64)
	ids := make([]int, 64)
	for i := range frames {
		frames[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
		ids[i] = i
	}
	m, e := fitEventWindow(frames, ids)
	if e != nil {
		b.Fatal(e)
	}
	models := windowGuideModels{m.count, m.count, m.count, m.count, m.subset, m.subset}
	for _, subset := range []bool{false, true} {
		b.Run(fmt.Sprintf("subset-%t", subset), func(b *testing.B) {
			outer := bayes.ForecastMix{Weights: [4]float64{0, 1, 0, 0}}
			inner := bayes.ForecastMix{Weights: [4]float64{0, 1, 0, 0}}
			if subset {
				inner.Weights = [4]float64{0, 0, 1, 0}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a, e := windowGuidePredict(models, bayes.ForecastMix{}, inner, outer, observationexperiment.Frames(uint16(i%512), "bench"), 0)
				if e != nil {
					b.Fatal(e)
				}
				windowGuideBenchSink = a
			}
		})
	}
}
