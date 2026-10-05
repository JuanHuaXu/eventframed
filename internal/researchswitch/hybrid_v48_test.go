package researchswitch

import (
	"math"
	"math/rand"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchhybridref"
)

func TestHybridV48CompactEquivalent(t *testing.T) {
	for _, prior := range [][]float64{{.9, .1}, {.8, .1, .1}} {
		for _, alpha := range []float64{0, 1. / 2400, 1. / 150, 1} {
			cfg := Config{Prior: prior, Hazard: alpha, Trials: 64, Pending: 64}
			m, err := New(cfg, 1)
			if err != nil {
				t.Fatal(err)
			}
			c, err := newCompactV48(compactConfigV48(cfg), 1)
			if err != nil {
				t.Fatal(err)
			}
			rng := rand.New(rand.NewSource(2026104897))
			var mt []Ticket
			var ct []compactTicketV48
			var tape [][]float64
			known := map[int]bool{}
			pending := []int{}
			for j := 0; j < 64; j++ {
				q := make([]float64, len(prior))
				for h := range q {
					q[h] = .01 + .98*rng.Float64()
				}
				at := int64(2 * j)
				a, err := m.Issue(q, at)
				if err != nil {
					t.Fatal(err)
				}
				b, err := c.Issue(q, at)
				if err != nil || a.Forecast() != b.Forecast() {
					t.Fatal("compact changed original scored forecast", err)
				}
				mt, ct = append(mt, a), append(ct, b)
				tape = append(tape, q)
				pending = append(pending, j)
				if j%3 != 0 {
					index := rng.Intn(len(pending))
					slot := pending[index]
					pending = append(pending[:index], pending[index+1:]...)
					if j%7 == 0 {
						if err = m.Cancel(mt[slot], at+1); err != nil {
							t.Fatal(err)
						}
						if err = c.Cancel(ct[slot], at+1); err != nil {
							t.Fatal(err)
						}
					} else {
						y := rng.Intn(2) == 1
						r, err := m.Resolve(mt[slot], y, at+1)
						if err != nil {
							t.Fatal(err)
						}
						s, err := c.Resolve(ct[slot], y, at+1)
						if err != nil || r != Receipt(s) {
							t.Fatal("compact changed receipt", err)
						}
						known[slot] = y
					}
				}
				w, err := researchhybridref.End(prior, alpha, tape, known)
				if err != nil {
					t.Fatal(err)
				}
				got := c.probabilities(c.logsAt(j))
				for h := range w {
					near(t, got[h], w[h])
				}
			}
		}
	}
}

func TestHybridV48CompactExtremeRevival(t *testing.T) {
	c, err := newCompactV48(compactConfigV48{Prior: []float64{.5, .5}, Hazard: 0, Trials: 128, Pending: 128}, 1)
	if err != nil {
		t.Fatal(err)
	}
	for j := 0; j < 128; j++ {
		ticket, err := c.Issue([]float64{ProbabilityFloor, 1 - ProbabilityFloor}, int64(j))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = c.Resolve(ticket, j < 64, int64(j)); err != nil {
			t.Fatal(err)
		}
	}
	w := c.probabilities(c.nextLogs())
	// Resolve the actual Float64 advice, not idealized decimal symmetry.
	q0, q1 := ProbabilityFloor, 1-ProbabilityFloor
	l0 := 64 * (math.Log(q0) + math.Log1p(-q0))
	l1 := 64 * (math.Log(q1) + math.Log1p(-q1))
	w0 := 1 / (1 + math.Exp(l1-l0))
	near(t, w[0], w0)
	near(t, w[1], 1-w0)
}

func TestHybridV48DelayedHeadsAndOriginalServedLaw(t *testing.T) {
	base := []float64{.85, .35, .7, .45}
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 64, Pending: 64}
	for _, hazard := range []float64{0, 1. / 2400, 1. / 150} {
		p, err := NewHybridPool(base, cfg, 16, 1./2400, hazard, 1)
		if err != nil {
			t.Fatal(err)
		}
		gc := cfg
		gc.Hazard = 1. / 2400
		g, err := NewPool(base, gc, 1)
		if err != nil {
			t.Fatal(err)
		}
		l, err := NewLocalPool(base, cfg, 16, 1)
		if err != nil {
			t.Fatal(err)
		}
		var tickets []HybridTicket
		var gt []PoolTicket
		var lt []LocalTicket
		var heads [][]float64
		known := map[int]bool{}
		pending := []int{}
		rng := rand.New(rand.NewSource(2026104899))
		for j := 0; j < 64; j++ {
			at, m := int64(2*j), j%4
			qg, e := g.Predict(m)
			if e != nil {
				t.Fatal(e)
			}
			ql, e := l.Predict(m)
			if e != nil {
				t.Fatal(e)
			}
			w, e := researchhybridref.End([]float64{.9, .1}, hazard, heads, known)
			if e != nil {
				t.Fatal(e)
			}
			if j > 0 {
				w[0], w[1] = (1-hazard)*w[0]+hazard*.9, (1-hazard)*w[1]+hazard*.1
			}
			ticket, e := p.Issue(m, at)
			if e != nil {
				t.Fatal(e)
			}
			near(t, ticket.Forecast(), w[0]*qg+w[1]*ql)
			q, ah, e := p.advice(m)
			if e != nil || len(q) != 3 || len(ah) != 2 {
				t.Fatal(e)
			}
			tickets = append(tickets, ticket)
			heads = append(heads, []float64{qg, ql})
			tg, e := g.Issue(m, at)
			if e != nil {
				t.Fatal(e)
			}
			tl, e := l.Issue(m, at)
			if e != nil {
				t.Fatal(e)
			}
			gt, lt = append(gt, tg), append(lt, tl)
			pending = append(pending, j)
			if j%3 != 0 {
				index := rng.Intn(len(pending))
				slot := pending[index]
				pending = append(pending[:index], pending[index+1:]...)
				if j%7 == 0 {
					for _, e := range []error{p.Cancel(tickets[slot], at+1), g.Cancel(gt[slot], at+1), l.Cancel(lt[slot], at+1)} {
						if e != nil {
							t.Fatal(e)
						}
					}
				} else {
					y := rng.Intn(2) == 1
					r, e := p.Resolve(tickets[slot], y, at+1)
					if e != nil || r.Forecast != tickets[slot].Forecast() || r.TrialOrdinal != slot+1 || r.Member != slot%4 || r.MemberOrdinal != slot/4+1 {
						t.Fatal("original served law replaced by a head", e)
					}
					if _, e = g.Resolve(gt[slot], y, at+1); e != nil {
						t.Fatal(e)
					}
					if _, e = l.Resolve(lt[slot], y, at+1); e != nil {
						t.Fatal(e)
					}
					known[slot] = y
				}
			}
			if p.Pending() != len(pending) || p.global.Pending() != len(pending) || p.scope.Pending() != len(pending) {
				t.Fatal("head/scope pending mismatch")
			}
			for m := 0; m < 4; m++ {
				_, got, e := p.advice(m)
				if e != nil {
					t.Fatal(e)
				}
				x, e := g.Predict(m)
				if e != nil {
					t.Fatal(e)
				}
				y, e := l.Predict(m)
				if e != nil {
					t.Fatal(e)
				}
				near(t, got[0], x)
				near(t, got[1], y)
			}
		}
	}
}

func TestHybridV48OwnershipCapsAndGlobalFence(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: .1, Trials: 4, Pending: 2}
	p, err := NewHybridPool([]float64{.85, .35}, cfg, 2, .1, .1, 1)
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.Issue(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	f, err := NewHybridPool(p.local.base, cfg, 2, .1, .1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Resolve(first, true, 2); err == nil || f.local.used != 0 {
		t.Fatal("foreign handle")
	}
	if _, err = p.Resolve(first, true, 0); err == nil || p.local.pending != 1 || p.local.poison {
		t.Fatal("backward arrival mutated state")
	}
	second, err := p.Issue(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Issue(1, 2); err == nil || p.local.used != 2 || p.local.poison {
		t.Fatal("pending cap")
	}
	if err = p.Cancel(second, 2); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Issue(0, 2); err == nil || p.local.used != 2 {
		t.Fatal("local cap")
	}
	if err = p.Cancel(second, 2); err == nil {
		t.Fatal("duplicate cancellation")
	}
	p.scope.rows[0].advice[0] = 0
	if _, err = p.Resolve(first, true, 9); err == nil || !p.local.poison {
		t.Fatal("partial scope failure not fenced")
	}
	if _, err = p.Predict(1); err == nil {
		t.Fatal("partial bundle emitted")
	}
	if err = p.BeginEpoch(2, 8); err == nil {
		t.Fatal("reset lost failed-operation clock")
	}
	if err = p.BeginEpoch(2, 9); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Resolve(first, true, 10); err == nil {
		t.Fatal("old epoch")
	}
	if _, err = p.Issue(1, 8); err == nil {
		t.Fatal("reset lost logical clock")
	}
}
