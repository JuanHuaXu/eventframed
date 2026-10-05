package researchswitch

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchdispersion"
	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

func testPool(t *testing.T, alpha float64) *Pool {
	t.Helper()
	p, err := NewPool([]float64{.85, .35}, Config{Prior: []float64{.8, .1, .1}, Hazard: alpha, Trials: 64, Pending: 64}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPoolActualExpertsAndDelayedScoredLaw(t *testing.T) {
	for _, alpha := range []float64{0, 1. / 2400, 1. / 150, 1} {
		p := testPool(t, alpha)
		f, err := researchdispersion.NewWindowObserver(p.base, 1, 64, "full")
		if err != nil {
			t.Fatal(err)
		}
		a, err := researchdispersion.NewWindowObserver(p.base, 1, 64, "adaptive")
		if err != nil {
			t.Fatal(err)
		}
		r, err := researchmoment.New(p.base, 1, 64, researchmoment.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true})
		if err != nil {
			t.Fatal(err)
		}
		rng := rand.New(rand.NewSource(2026104301))
		var qs [][]float64
		known := map[int]bool{}
		var pt []PoolTicket
		var ft, at []researchdispersion.DelayedTicket
		var rt []researchmoment.Ticket
		var pending []int
		for j := 0; j < 32; j++ {
			clock := int64(2 * j)
			qf, _ := f.Predict(j % 2)
			qa, _ := a.Predict(j % 2)
			qr, _ := r.Predict(j % 2)
			advice := []float64{qf, qa, qr}
			want := referencePrediction(t, p.mix, qs, known, advice)
			ticket, e := p.Issue(j%2, clock)
			if e != nil {
				t.Fatal(e)
			}
			near(t, ticket.Forecast(), want)
			qs = append(qs, advice)
			pt = append(pt, ticket)
			ff, e := f.Issue(j%2, clock)
			if e != nil {
				t.Fatal(e)
			}
			ft = append(ft, ff)
			aa, e := a.Issue(j%2, clock)
			if e != nil {
				t.Fatal(e)
			}
			at = append(at, aa)
			rr, e := r.Issue(j%2, clock)
			if e != nil {
				t.Fatal(e)
			}
			rt = append(rt, rr)
			pending = append(pending, j)
			if j%3 != 0 {
				index := rng.Intn(len(pending))
				slot := pending[index]
				pending = append(pending[:index], pending[index+1:]...)
				if j%7 == 0 {
					for _, e := range []error{p.Cancel(pt[slot], clock+1), f.Cancel(ft[slot], clock+1), a.Cancel(at[slot], clock+1), r.Cancel(rt[slot], clock+1)} {
						if e != nil {
							t.Fatal(e)
						}
					}
				} else {
					y := rng.Intn(2) == 1
					receipt, e := p.Resolve(pt[slot], y, clock+1)
					if e != nil || receipt.Forecast != pt[slot].Forecast() || receipt.Member != slot%2 || receipt.MemberOrdinal != slot/2+1 {
						t.Fatal("pool original law identity", e)
					}
					if _, e = f.Resolve(ft[slot], y, clock+1); e != nil {
						t.Fatal(e)
					}
					if _, e = a.Resolve(at[slot], y, clock+1); e != nil {
						t.Fatal(e)
					}
					if _, e = r.Resolve(rt[slot], y, clock+1); e != nil {
						t.Fatal(e)
					}
					known[slot] = y
				}
			}
			if p.Pending() != len(pending) || p.full.Pending() != p.Pending() || p.adaptive.Pending() != p.Pending() || p.moment.Pending() != p.Pending() {
				t.Fatal("expert pending conservation")
			}
			for member := 0; member < 2; member++ {
				qf, _ := f.Predict(member)
				qa, _ := a.Predict(member)
				qr, _ := r.Predict(member)
				got, e := p.Predict(member)
				if e != nil {
					t.Fatal(e)
				}
				near(t, got, referencePrediction(t, p.mix, qs, known, []float64{qf, qa, qr}))
			}
		}
	}
}

func TestPoolRejectsBeforeMutationAndFencesPartialFailure(t *testing.T) {
	p := testPool(t, .1)
	first, e := p.Issue(0, 1)
	if e != nil {
		t.Fatal(e)
	}
	foreign := testPool(t, .1)
	if _, e = foreign.Resolve(first, true, 2); e == nil || foreign.mix.used != 0 {
		t.Fatal("foreign pool input")
	}
	before := append([]poolRow(nil), p.rows...)
	if _, e = p.Resolve(first, true, 0); e == nil || !reflect.DeepEqual(p.rows, before) || p.poison {
		t.Fatal("invalid input mutated pool")
	}
	// Deliberately corrupt one retained private child handle. Full/Adaptive can
	// already have changed when the third child rejects. Fencing is required;
	// no false transaction/rollback guarantee is asserted for external APIs.
	p.rows[0].moment = researchmoment.Ticket{}
	if _, e = p.Resolve(first, true, 2); e == nil || !p.poison {
		t.Fatal("partial child failure not fenced")
	}
	if _, e = p.Predict(0); e == nil {
		t.Fatal("poisoned bundle emitted law")
	}
	if _, e = p.Issue(1, 3); e == nil {
		t.Fatal("poisoned bundle admitted trial")
	}
	if e = p.Cancel(first, 3); e == nil {
		t.Fatal("poisoned bundle accepted cancel")
	}
	if e = p.BeginEpoch(2, 1); e == nil {
		t.Fatal("epoch replacement moved before a partially committed child operation")
	}
	if e = p.BeginEpoch(2, 3); e != nil {
		t.Fatal(e)
	}
	if _, e = p.Resolve(first, true, 4); e == nil {
		t.Fatal("old pool epoch accepted")
	}
	if _, e = p.Issue(1, 2); e == nil {
		t.Fatal("reset lost logical time fence")
	}
	if _, e = p.Issue(1, 4); e != nil {
		t.Fatal("explicit epoch failed to restore healthy pool", e)
	}
}

func TestPoolUntouchedFutureDoesNotChangePrefix(t *testing.T) {
	a, b := testPool(t, .1), testPool(t, .1)
	for j := 0; j < 20; j++ {
		ta, e := a.Issue(j%2, int64(2*j))
		if e != nil {
			t.Fatal(e)
		}
		tb, e := b.Issue(j%2, int64(2*j))
		if e != nil {
			t.Fatal(e)
		}
		if ta.Forecast() != tb.Forecast() {
			t.Fatal("future label altered an issued prefix")
		}
		if j < 19 {
			if _, e = a.Resolve(ta, j%3 == 0, int64(2*j+1)); e != nil {
				t.Fatal(e)
			}
			if _, e = b.Resolve(tb, j%3 == 0, int64(2*j+1)); e != nil {
				t.Fatal(e)
			}
		} else {
			if _, e = a.Resolve(ta, true, 39); e != nil {
				t.Fatal(e)
			}
			if _, e = b.Resolve(tb, false, 39); e != nil {
				t.Fatal(e)
			}
		}
	}
	qa, e := a.Predict(1)
	if e != nil {
		t.Fatal(e)
	}
	qb, e := b.Predict(1)
	if e != nil {
		t.Fatal(e)
	}
	if qa == qb {
		t.Fatal("revealed label did not enter scored output")
	}
}
