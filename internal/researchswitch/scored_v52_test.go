package researchswitch

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchscoreref"
)

func TestScoredV52HybridLegacyExactAndFencing(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 32, Pending: 32}
	for _, alpha := range []float64{0, 1. / 2400, 1. / 150} {
		old, err := NewMemoHybridV49([]float64{.3, .8}, cfg, 16, 1./2400, alpha, 1)
		if err != nil {
			t.Fatal(err)
		}
		next, err := NewScoredHybridV52([]float64{.3, .8}, cfg, 16, 1./2400, alpha, 1, "log_mean")
		if err != nil {
			t.Fatal(err)
		}
		var x [32]HybridTicket
		var y [32]ScoredHybridTicketV52
		for j := range x {
			x[j], err = old.Issue(j%2, int64(j))
			if err != nil {
				t.Fatal(err)
			}
			y[j], err = next.Issue(j%2, int64(j))
			if err != nil || x[j].Forecast() != y[j].Forecast() {
				t.Fatal("served legacy law", err)
			}
		}
		for j := 0; j < 32; j++ {
			id := j * 13 % 32
			at := int64(32 + j)
			if id%7 == 0 {
				if err = old.Cancel(x[id], at); err != nil {
					t.Fatal(err)
				}
				if err = next.Cancel(y[id], at); err != nil {
					t.Fatal(err)
				}
			} else {
				a, e := old.Resolve(x[id], id%3 != 0, at)
				b, f := next.Resolve(y[id], id%3 != 0, at)
				if e != nil || f != nil || a != b {
					t.Fatal("original receipt", a, b, e, f)
				}
			}
			for m := 0; m < 2; m++ {
				a, e := old.Predict(m)
				b, f := next.Predict(m)
				if e != nil || f != nil || a != b {
					t.Fatal("delayed whole law", e, f)
				}
			}
		}
	}
	for _, mode := range scoreModesV51 {
		for _, failure := range []string{"issue", "resolve", "cancel"} {
			p, err := NewScoredHybridV52([]float64{.3, .8}, cfg, 16, 1./2400, 1./150, 1, mode)
			if err != nil {
				t.Fatal(err)
			}
			if failure == "issue" {
				p.scope.used = len(p.scope.rows)
				_, err = p.Issue(0, 0)
			} else {
				ticket, e := p.Issue(0, 0)
				if e != nil {
					t.Fatal(e)
				}
				p.scope.epoch = 99
				if failure == "resolve" {
					_, err = p.Resolve(ticket, true, 1)
				} else {
					err = p.Cancel(ticket, 1)
				}
			}
			if err == nil || !p.local.poison {
				t.Fatal("partial bundle not fenced", mode, failure)
			}
			if _, err = p.Predict(0); err == nil {
				t.Fatal("poisoned law served")
			}
			if err = p.BeginEpoch(2, 2); err != nil || p.local.poison || p.local.mixes[0].mode != mode || p.global.mode != mode || p.scope.mode != mode {
				t.Fatal("epoch mode recovery", err)
			}
		}
	}
}

func TestScoredV52AllHeadDenseDelayedReference(t *testing.T) {
	const steps = 64
	checks := 0
	for _, mode := range scoreModesV51 {
		cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: steps, Pending: steps}
		p, err := NewScoredHybridV52([]float64{.3, .45, .65, .8}, cfg, 16, 1./2400, 1./150, 1, mode)
		if err != nil {
			t.Fatal(err)
		}
		var tickets [steps]ScoredHybridTicketV52
		for j := range tickets {
			tickets[j], err = p.Issue(j%4, int64(j))
			if err != nil {
				t.Fatal(err)
			}
		}
		known := map[int]bool{}
		locals := make([]map[int]bool, 4)
		for m := range locals {
			locals[m] = map[int]bool{}
		}
		verify := func(model *scoreModelV51, observed map[int]bool, advice []float64) {
			t.Helper()
			tape := make([][]float64, model.used)
			for j := range tape {
				tape[j] = append([]float64(nil), model.rows[j].advice[:model.n]...)
			}
			w, e := researchscoreref.End(model.prior[:model.n], model.hazard, tape, observed, mode)
			if e != nil {
				t.Fatal(e)
			}
			got := model.probabilities(model.logsAt(model.used - 1))
			for h := range w {
				if math.Abs(got[h]-w[h]) > 3e-12 {
					t.Fatal("dense state", mode, h)
				}
				w[h] = (1-model.hazard)*w[h] + model.hazard*model.prior[h]
			}
			q, e := model.Predict(advice)
			r, f := researchscoreref.Forecast(w, advice, mode)
			if e != nil || f != nil || math.Abs(q-r) > 3e-12 {
				t.Fatal("independent head law", mode, q, r, e, f)
			}
			checks++
		}
		for j := range tickets {
			id := j * 37 % steps
			at := int64(steps + j)
			if id%11 == 0 {
				err = p.Cancel(tickets[id], at)
			} else {
				y := id%3 != 0
				r, e := p.Resolve(tickets[id], y, at)
				err = e
				if r.Forecast != tickets[id].Forecast() || r.TrialOrdinal != id+1 || r.Member != id%4 || r.MemberOrdinal != id/4+1 {
					t.Fatal("original served receipt")
				}
				known[id] = y
				locals[id%4][id/4] = y
			}
			if err != nil {
				t.Fatal(err)
			}
			verify(p.global, known, []float64{.3, .55, .8})
			verify(p.scope, known, []float64{.35, .75})
			for m := range locals {
				verify(p.local.mixes[m], locals[m], []float64{.3, .55, .8})
			}
		}
	}
	t.Log("independent all-head state and next law cases", checks)
}

func TestScoredV52OwnershipCapsAndLocalClock(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 4, Pending: 2}
	for _, mode := range scoreModesV51 {
		p, err := NewScoredHybridV52([]float64{.3, .8}, cfg, 2, 1./2400, 1./150, 1, mode)
		if err != nil {
			t.Fatal(err)
		}
		other, err := NewScoredHybridV52([]float64{.3, .8}, cfg, 2, 1./2400, 1./150, 1, mode)
		if err != nil {
			t.Fatal(err)
		}
		x, err := p.Issue(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		foreign, err := other.Issue(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		state := append([]scoreRowV51(nil), p.global.rows...)
		if _, err = p.Resolve(foreign, true, 1); err == nil || !reflect.DeepEqual(state, p.global.rows) {
			t.Fatal("foreign changed state")
		}
		clock := p.local.mixes[1].used
		if _, err = p.Issue(0, 1); err != nil || p.local.mixes[1].used != clock {
			t.Fatal("another member advanced local clock", err)
		}
		if _, err = p.Issue(1, 2); err == nil {
			t.Fatal("pending cap missed")
		}
		if _, err = p.Resolve(x, true, 2); err != nil {
			t.Fatal(err)
		}
		if _, err = p.Resolve(x, false, 3); err == nil {
			t.Fatal("duplicate accepted")
		}
		if _, err = p.Issue(0, 3); err == nil {
			t.Fatal("member cap missed")
		}
		if err = p.BeginEpoch(2, 3); err != nil {
			t.Fatal(err)
		}
		if _, err = p.Resolve(x, true, 4); err == nil {
			t.Fatal("stale accepted")
		}
	}
}
