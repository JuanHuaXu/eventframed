package researchswitch

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
	"github.com/JuanHuaXu/eventframed/internal/researchswitchref"
)

func localReference(t *testing.T, alpha float64, tape [][3]float64, known map[int]bool, advice [3]float64) float64 {
	t.Helper()
	prior := []float64{.8, .1, .1}
	w, err := researchswitchref.EndWeights(prior, alpha, tape, known)
	if err != nil {
		t.Fatal(err)
	}
	q := 0.
	for h := range w {
		if len(tape) > 0 {
			w[h] = (1-alpha)*w[h] + alpha*prior[h]
		}
		q += w[h] * advice[h]
	}
	return q
}

func TestLocalV47DelayedReferenceAndSameExperts(t *testing.T) {
	for _, alpha := range []float64{0, 1. / 16, 1} {
		base := []float64{.85, .35, .7, .45}
		cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: alpha, Trials: 64, Pending: 64}
		p, err := NewLocalPool(base, cfg, 16, 1)
		if err != nil {
			t.Fatal(err)
		}
		g, err := NewPool(base, cfg, 1)
		if err != nil {
			t.Fatal(err)
		}
		var tapes [4][][3]float64
		var known [4]map[int]bool
		for m := range known {
			known[m] = map[int]bool{}
		}
		var tickets []LocalTicket
		var global []PoolTicket
		pending := []int{}
		rng := rand.New(rand.NewSource(2026104797))
		for j := 0; j < 64; j++ {
			m, at := j%4, int64(2*j)
			q, e := p.advice(m)
			if e != nil {
				t.Fatal(e)
			}
			gq, e := g.advice(m)
			if e != nil || q != gq {
				t.Fatal("shared expert advice diverged", e)
			}
			want := localReference(t, alpha, tapes[m], known[m], q)
			ticket, e := p.Issue(m, at)
			if e != nil {
				t.Fatal(e)
			}
			near(t, ticket.Forecast(), want)
			tickets = append(tickets, ticket)
			gt, e := g.Issue(m, at)
			if e != nil {
				t.Fatal(e)
			}
			global = append(global, gt)
			tapes[m] = append(tapes[m], q)
			pending = append(pending, j)
			if j%3 != 0 {
				index := rng.Intn(len(pending))
				slot := pending[index]
				pending = append(pending[:index], pending[index+1:]...)
				if j%7 == 0 {
					if e = p.Cancel(tickets[slot], at+1); e != nil {
						t.Fatal(e)
					}
					if e = g.Cancel(global[slot], at+1); e != nil {
						t.Fatal(e)
					}
				} else {
					y := rng.Intn(2) == 1
					r, e := p.Resolve(tickets[slot], y, at+1)
					if e != nil || r.TrialOrdinal != slot+1 || r.Member != slot%4 || r.MemberOrdinal != slot/4+1 || r.Forecast != tickets[slot].Forecast() {
						t.Fatal("original local/global receipt", e)
					}
					if _, e = g.Resolve(global[slot], y, at+1); e != nil {
						t.Fatal(e)
					}
					known[slot%4][slot/4] = y
				}
			}
			for m := range tapes {
				q, e := p.advice(m)
				if e != nil {
					t.Fatal(e)
				}
				got, e := p.Predict(m)
				if e != nil {
					t.Fatal(e)
				}
				near(t, got, localReference(t, alpha, tapes[m], known[m], q))
			}
			localPending := 0
			for _, mix := range p.mixes {
				localPending += mix.Pending()
			}
			if p.Pending() != len(pending) || localPending != p.Pending() || p.full.Pending() != p.Pending() || p.adaptive.Pending() != p.Pending() || p.moment.Pending() != p.Pending() {
				t.Fatal("pending conservation")
			}
		}
	}
}

func TestLocalV47InactiveTaskDoesNotAge(t *testing.T) {
	p, err := NewLocalPool([]float64{.85, .35}, Config{Prior: []float64{.8, .1, .1}, Hazard: .5, Trials: 32, Pending: 32}, 16, 1)
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := p.Issue(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Resolve(ticket, false, 0); err != nil {
		t.Fatal(err)
	}
	before := p.mixes[0].nextLogs()
	for j := 1; j <= 16; j++ {
		ticket, e := p.Issue(1, int64(j))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = p.Resolve(ticket, j%2 == 0, int64(j)); e != nil {
			t.Fatal(e)
		}
		if before != p.mixes[0].nextLogs() || p.issued[0] != 1 {
			t.Fatal("inactive task reset/aged")
		}
	}
}

func TestLocalV47CapsOwnershipAndPartialFence(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: .1, Trials: 4, Pending: 2}
	p, err := NewLocalPool([]float64{.85, .35}, cfg, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.Issue(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := NewLocalPool(p.base, cfg, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []int64{0, -1} {
		if _, err = p.Resolve(first, true, at); err == nil || p.poison || p.pending != 1 {
			t.Fatal("backward outcome mutated state")
		}
	}
	if _, err = foreign.Resolve(first, true, 2); err == nil || foreign.used != 0 {
		t.Fatal("foreign handle accepted")
	}
	second, err := p.Issue(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Issue(1, 2); err == nil || p.used != 2 || p.poison {
		t.Fatal("global pending cap")
	}
	if err = p.Cancel(second, 2); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Issue(0, 2); err == nil || p.used != 2 {
		t.Fatal("task history cap")
	}
	if err = p.Cancel(second, 2); err == nil {
		t.Fatal("duplicate cancel")
	}
	// A valid third-child fault must fence ALL tasks and preserve the global
	// failed-operation clock. Replacing only this task would expose mixed state.
	p.rows[0].moment = researchmoment.Ticket{}
	if _, err = p.Resolve(first, true, 9); err == nil || !p.poison {
		t.Fatal("partial mutation not fenced")
	}
	if _, err = p.Predict(1); err == nil {
		t.Fatal("other task escaped global fence")
	}
	if err = p.BeginEpoch(2, 8); err == nil {
		t.Fatal("epoch lost failed-operation clock")
	}
	if err = p.BeginEpoch(2, 9); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Resolve(first, true, 10); err == nil {
		t.Fatal("old epoch handle")
	}
	if _, err = p.Issue(1, 8); err == nil {
		t.Fatal("old clock issue")
	}
	if _, err = p.Issue(1, 10); err != nil {
		t.Fatal("new epoch unhealthy", err)
	}
}

func TestLocalV47ConstructorOwnsContracts(t *testing.T) {
	base := []float64{.85, .35}
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: .1, Trials: 4, Pending: 4}
	p, err := NewLocalPool(base, cfg, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	base[0], cfg.Prior[0] = .01, .01
	if !reflect.DeepEqual(p.base, []float64{.85, .35}) || p.cfg.Prior[0] != .8 {
		t.Fatal("mutable caller contract alias")
	}
	for _, cap := range []int{0, 65} {
		if _, err = NewLocalPool(p.base, p.cfg, cap, 1); err == nil {
			t.Fatal("invalid local history cap")
		}
	}
}
