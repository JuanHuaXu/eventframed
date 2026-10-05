package observationgate

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

func windowProxyPredict(m windowGuideModels, oldInner, newInner, outer bayes.ForecastMix, r observation.Reader, seed int64, enabled bool) (windowGuideAdvice, error) {
	// Delegate invalid state to the original validator, before any observation.
	valid := m.Base != nil && r != nil
	for _, s := range []bayes.ForecastMix{oldInner, newInner, outer} {
		if s.Weights == [4]float64{} {
			continue
		}
		sum := 0.
		for _, w := range s.Weights {
			if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
				valid = false
			}
			sum += w
		}
		if math.Abs(sum-1) > 1e-9 {
			valid = false
		}
	}
	ow, nw := windowGuideWeights(outer), windowGuideWeights(newInner)
	guide := 0
	if m.LabelCount != nil && ow[1] > ow[guide] {
		guide = 1
	}
	if m.Long != nil && ow[2] > ow[guide] {
		guide = 2
	}
	eventCountWins := guide == 1 && m.EventCount != nil && nw[1] > nw[0] && (m.EventSubset == nil || nw[2]+nw[3] <= nw[1])
	if !enabled || !valid || !eventCountWins || m.EventSubset == nil {
		return windowGuidePredict(m, oldInner, newInner, outer, r, seed)
	}
	view, e := observationlearners.RunConditionalObserver(m.EventSubset, r, r.Epoch())
	if e != nil {
		return windowGuideAdvice{}, e
	}
	last := view.Trace[len(view.Trace)-1]
	// Acquisition changed, not the forecast definition. In particular EventCount
	// remains in NewInner[1]; no outcome or simulator rule selects the proxy.
	a, e := windowFixedAdvice(m, [3]bayes.ForecastMix{oldInner, newInner, outer}, last.Observed, last.Values, view.Cost)
	a.Guide = "event-count/subset-proxy"
	return a, e
}

func TestWindowProxyContracts(t *testing.T) {
	f := make([]innerArrivalFrame, 64)
	ids := make([]int, 64)
	for i := range f {
		f[i] = innerArrivalFrame{X: uint16(i * 37 % 512), Y: i%3 != 0}
		ids[i] = i
	}
	models, e := fitEventWindow(f, ids)
	if e != nil {
		t.Fatal(e)
	}
	m := windowGuideModels{models.count, models.count, models.count, models.count, models.subset, models.subset}
	focus := func(i int) bayes.ForecastMix { var v bayes.ForecastMix; v.Weights[i] = 1; return v }
	differences := 0
	for _, outer := range []bayes.ForecastMix{focus(0), focus(1), focus(2)} {
		for _, inner := range []bayes.ForecastMix{focus(0), focus(1), focus(2)} {
			for x := uint16(0); x < 512; x++ {
				old := focus(0)
				a, e := windowGuidePredict(m, old, inner, outer, observationexperiment.Frames(x, "original"), 0)
				if e != nil {
					t.Fatal(e)
				}
				b, e := windowProxyPredict(m, old, inner, outer, observationexperiment.Frames(x, "disabled"), 0, false)
				if e != nil || a != b {
					t.Fatal("disabled drift", e)
				}
				c, e := windowProxyPredict(m, old, inner, outer, observationexperiment.Frames(x, "enabled"), 0, true)
				if e != nil {
					t.Fatal(e)
				}
				if a.Guide != "event-count" {
					if a != c {
						t.Fatal("unrelated guide changed")
					}
					continue
				}
				if c.Guide != "event-count/subset-proxy" || c.Cost > 6 || c.Values != x&c.Mask {
					t.Fatal("proxy contract")
				}
				q, _ := m.EventCount.ForecastObserved(c.Mask, c.Values)
				s, _ := m.EventSubset.Forecast(c.Mask, c.Values)
				if c.NewInner[1] != q {
					t.Fatal("count removed")
				}
				if q != s {
					differences++
				}
				expected, e := windowFixedAdvice(m, [3]bayes.ForecastMix{old, inner, outer}, c.Mask, c.Values, c.Cost)
				if e != nil {
					t.Fatal(e)
				}
				expected.Guide = c.Guide
				if !reflect.DeepEqual(expected, c) {
					t.Fatal("law changed beyond mask")
				}
			}
		}
	}
	if differences == 0 {
		t.Fatal("count retention check vacuous")
	}
	m.EventSubset = nil
	a, e := windowGuidePredict(m, focus(0), focus(1), focus(1), observationexperiment.Frames(0, "missing"), 0)
	b, err := windowProxyPredict(m, focus(0), focus(1), focus(1), observationexperiment.Frames(0, "missing"), 0, true)
	if e != nil || err != nil || a != b {
		t.Fatal("missing model fallback")
	}
}
