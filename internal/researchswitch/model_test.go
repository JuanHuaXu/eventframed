package researchswitch

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchswitchref"
)

func testModel(t *testing.T, hazard float64, trials, pending int) *Model {
	t.Helper()
	m, e := New(Config{Prior: []float64{.6, .3, .1}, Hazard: hazard, Trials: trials, Pending: pending}, 1)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func near(t *testing.T, x, y float64) {
	t.Helper()
	if math.IsNaN(x) || math.Abs(x-y) > 3e-12 {
		t.Fatalf("got %.17g want %.17g", x, y)
	}
}

func TestCachedTransitionExactlyMatchesOriginal(t *testing.T) {
	// This regression deliberately keeps the pre-cache expression independent
	// of the candidate's cached fields, including both endpoint branches.
	for _, alpha := range []float64{0, math.SmallestNonzeroFloat64, 1. / 150, .23, math.Nextafter(1, 0), 1} {
		m := testModel(t, alpha, 64, 64)
		for j := 0; j < 200; j++ {
			previous := [MaxExperts]float64{-float64(j) * 13.8, -float64(j+1) * .51, -float64(j+7) * 3.2}
			want := previous
			if alpha == 1 {
				want = m.logPrior
			} else if alpha > 0 {
				want = [MaxExperts]float64{}
				for i := 0; i < m.n; i++ {
					a, b := math.Log1p(-m.hazard)+previous[i], math.Log(m.hazard)+m.logPrior[i]
					maximum := math.Max(a, b)
					want[i] = maximum + math.Log(math.Exp(a-maximum)+math.Exp(b-maximum))
				}
			}
			got := m.transitionLogs(previous)
			if got != want {
				t.Fatalf("cached transition changed: alpha=%g row=%d got=%v want=%v", alpha, j, got, want)
			}
		}
		if err := m.BeginEpoch(2, 1); err != nil {
			t.Fatal(err)
		}
		if alpha > 0 && alpha < 1 && (m.logStay != math.Log1p(-alpha) || m.logReset[0] != math.Log(alpha)+m.logPrior[0]) {
			t.Fatal("epoch reset did not rebuild constants")
		}
	}
}
func referencePrediction(t *testing.T, m *Model, advice [][]float64, known map[int]bool, q []float64) float64 {
	t.Helper()
	u, e := researchswitchref.Filter(m.prior[:m.n], m.hazard, advice, known)
	if e != nil {
		t.Fatal(e)
	}
	if len(advice) > 0 {
		for i := range u {
			u[i] = (1-m.hazard)*u[i] + m.hazard*m.prior[i]
		}
	}
	out := 0.
	for i := range u {
		out += u[i] * q[i]
	}
	return out
}

func TestOriginalPositionBatchAndPaths(t *testing.T) {
	for _, alpha := range []float64{0, 1. / 150, .23, 1} {
		m := testModel(t, alpha, 5, 5)
		qs := [][]float64{{.8, .3, .4}, {.2, .6, .9}, {.7, .5, .1}, {.1, .2, .95}, {.9, .1, .6}}
		tickets := make([]Ticket, len(qs))
		known := map[int]bool{}
		for j, q := range qs {
			want := referencePrediction(t, m, qs[:j], known, q)
			ticket, e := m.Issue(q, int64(j))
			if e != nil {
				t.Fatal(e)
			}
			tickets[j] = ticket
			near(t, ticket.Forecast(), want)
		}
		for k, j := range []int{3, 0, 4, 1, 2} {
			y := j%2 == 0
			known[j] = y
			r, e := m.Resolve(tickets[j], y, int64(5+k))
			if e != nil {
				t.Fatal(e)
			}
			if r.Forecast != tickets[j].Forecast() || r.TrialOrdinal != j+1 || r.IssuedAt != int64(j) {
				t.Fatal("original receipt changed")
			}
			want, e := researchswitchref.Filter(m.prior[:m.n], alpha, qs, known)
			if e != nil {
				t.Fatal(e)
			}
			var exact [MaxExperts]float64
			var walk func(int, int, float64)
			walk = func(pos, prev int, mass float64) {
				if pos == len(qs) {
					exact[prev] += mass
					return
				}
				for h := 0; h < 3; h++ {
					p := m.prior[h]
					if pos > 0 {
						p = alpha * m.prior[h]
						if h == prev {
							p += 1 - alpha
						}
					}
					if value, ok := known[pos]; ok {
						if value {
							p *= qs[pos][h]
						} else {
							p *= 1 - qs[pos][h]
						}
					}
					walk(pos+1, h, mass*p)
				}
			}
			walk(0, -1, 1)
			sum := exact[0] + exact[1] + exact[2]
			for h := 0; h < 3; h++ {
				near(t, math.Exp(m.logsAt(4)[h]), want[h])
				near(t, want[h], exact[h]/sum)
			}
		}
	}
}

func TestInterleavedDelayCancellationAndPrivateAdvice(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		m := testModel(t, .07, 64, 64)
		known := map[int]bool{}
		var qs [][]float64
		var tickets []Ticket
		var pending []int
		clock := int64(0)
		for j := 0; j < 64; j++ {
			q := []float64{.1 + .8*rng.Float64(), .1 + .8*rng.Float64(), .1 + .8*rng.Float64()}
			want := referencePrediction(t, m, qs, known, q)
			ticket, e := m.Issue(q, clock)
			if e != nil {
				t.Fatal(e)
			}
			near(t, ticket.Forecast(), want)
			qs = append(qs, append([]float64(nil), q...))
			q[0] = .999 // Must not change retained advice or original law.
			tickets = append(tickets, ticket)
			pending = append(pending, j)
			clock++
			if j%3 != 0 {
				index := rng.Intn(len(pending))
				slot := pending[index]
				pending = append(pending[:index], pending[index+1:]...)
				if j%7 == 0 {
					if e = m.Cancel(tickets[slot], clock); e != nil {
						t.Fatal(e)
					}
				} else {
					y := rng.Intn(2) == 1
					known[slot] = y
					r, e := m.Resolve(tickets[slot], y, clock)
					if e != nil || r.Forecast != tickets[slot].Forecast() {
						t.Fatal("private receipt", e)
					}
				}
				clock++
			}
			wantWeights, e := researchswitchref.Filter(m.prior[:m.n], m.hazard, qs, known)
			if e != nil {
				t.Fatal(e)
			}
			for h := range wantWeights {
				near(t, math.Exp(m.logsAt(m.used - 1)[h]), wantWeights[h])
			}
			if m.Pending() != len(pending) {
				t.Fatal("pending conservation")
			}
		}
	}
}

func TestLifecycleCapsAndAtomicFailure(t *testing.T) {
	m := testModel(t, .2, 2, 1)
	q := []float64{.7, .3, .5}
	first, e := m.Issue(q, 1)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Issue(q, 2); e == nil {
		t.Fatal("pending cap ignored")
	}
	foreign := testModel(t, .2, 2, 1)
	if _, e = foreign.Resolve(first, true, 2); e == nil {
		t.Fatal("foreign ticket accepted")
	}
	if _, e = m.Resolve(first, true, 0); e == nil {
		t.Fatal("time reversal")
	}
	// Ticket forecast is not the source of truth even inside same-package tests.
	forged := first
	forged.q = .999
	r, e := m.Resolve(forged, true, 2)
	if e != nil || r.Forecast != first.Forecast() {
		t.Fatal("mutable ticket forecast", e)
	}
	if _, e = m.Resolve(first, false, 3); e == nil {
		t.Fatal("double evidence")
	}
	second, e := m.Issue(q, 3)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.Cancel(second, 4); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Issue(q, 5); e == nil {
		t.Fatal("history cap ignored")
	}
	if _, e = m.Resolve(second, true, 5); e == nil {
		t.Fatal("cancelled evidence accepted")
	}
	if e = m.BeginEpoch(1, 5); e == nil {
		t.Fatal("same epoch")
	}
	if e = m.BeginEpoch(2, 5); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(first, true, 6); e == nil {
		t.Fatal("old epoch accepted")
	}
	ticket, e := m.Issue(q, 6)
	if e != nil {
		t.Fatal(e)
	}
	m.rows[0].advice[1] = math.NaN()
	before := append([]row(nil), m.rows...)
	clock, pending := m.clock, m.pending
	if _, e = m.Resolve(ticket, true, 7); e == nil {
		t.Fatal("invalid emission normalized")
	}
	// NaN cannot be DeepEqual to itself; compare unaffected semantic fields.
	if m.clock != clock || m.pending != pending || m.rows[0].status != before[0].status || m.rows[0].logPosterior != before[0].logPosterior {
		t.Fatal("failed resolve published")
	}
}

func TestContractAndCallerIsolation(t *testing.T) {
	cfg := Config{Prior: []float64{.6, .3, .1}, Hazard: .1, Trials: 4, Pending: 4}
	m, e := New(cfg, 1)
	if e != nil {
		t.Fatal(e)
	}
	cfg.Prior[0] = .1
	near(t, m.prior[0], .6)
	for _, q := range [][]float64{{.2, .3}, {0, .3, .5}, {1, .3, .5}, {math.NaN(), .3, .5}, {math.Inf(1), .3, .5}} {
		if _, e = m.Issue(q, 0); e == nil || m.used != 0 {
			t.Fatal("invalid advice published")
		}
	}
	for _, bad := range []Config{{Prior: []float64{1}, Trials: 1, Pending: 1}, {Prior: []float64{.9, .2}, Trials: 1, Pending: 1}, {Prior: []float64{0, 1}, Trials: 1, Pending: 1}, {Prior: []float64{.5, .5}, Hazard: math.NaN(), Trials: 1, Pending: 1}, {Prior: []float64{.5, .5}, Hazard: 1.1, Trials: 1, Pending: 1}, {Prior: []float64{.5, .5}, Trials: MaxTrials + 1, Pending: 1}, {Prior: []float64{.5, .5}, Trials: 1, Pending: 2}} {
		if _, e = New(bad, 1); e == nil {
			t.Fatal("invalid contract")
		}
	}
}

func TestVisiblePrefixAndShareLimits(t *testing.T) {
	for _, alpha := range []float64{0, .17, 1} {
		a, b := testModel(t, alpha, 32, 32), testModel(t, alpha, 32, 32)
		for j := 0; j < 16; j++ {
			q := []float64{.85, .25, .5}
			ta, _ := a.Issue(q, int64(2*j))
			tb, _ := b.Issue(q, int64(2*j))
			if ta.Forecast() != tb.Forecast() {
				t.Fatal("future changed prefix")
			}
			if j < 8 {
				if _, e := a.Resolve(ta, j%2 == 0, int64(2*j+1)); e != nil {
					t.Fatal(e)
				}
				if _, e := b.Resolve(tb, j%2 == 0, int64(2*j+1)); e != nil {
					t.Fatal(e)
				}
			} else if j == 15 {
				if _, e := a.Resolve(ta, true, 31); e != nil {
					t.Fatal(e)
				}
				if _, e := b.Resolve(tb, false, 31); e != nil {
					t.Fatal(e)
				}
			}
		}
		if alpha == 1 {
			qa, _ := a.Predict([]float64{.85, .25, .5})
			qb, _ := b.Predict([]float64{.85, .25, .5})
			near(t, qa, .6*.85+.3*.25+.1*.5)
			near(t, qa, qb)
		} else if reflect.DeepEqual(a.rows[15].logPosterior, b.rows[15].logPosterior) {
			t.Fatal("different evidence did not affect posterior")
		}
	}
}

func TestExtremeEvidenceCanReviveAnExpert(t *testing.T) {
	m, e := New(Config{Prior: []float64{.5, .5}, Hazard: 0, Trials: 200, Pending: 1}, 1)
	if e != nil {
		t.Fatal(e)
	}
	q := []float64{ProbabilityFloor, 1 - ProbabilityFloor}
	for j := 0; j < 200; j++ {
		ticket, e := m.Issue(q, int64(2*j))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(ticket, j >= 100, int64(2*j+1)); e != nil {
			t.Fatal(e)
		}
	}
	// Resolve the ACTUAL Float64 advice, not idealized decimal symmetry:
	// 1-(1-1e-6) is not exactly1e-6. Use an independent batch log calculation.
	got, e := m.Predict(q)
	if e != nil {
		t.Fatal(e)
	}
	l0 := 100 * (math.Log(q[0]) + math.Log1p(-q[0]))
	l1 := 100 * (math.Log(q[1]) + math.Log1p(-q[1]))
	w0 := 1 / (1 + math.Exp(l1-l0))
	near(t, got, w0*q[0]+(1-w0)*q[1])
}

func TestDelayedAndSubnormalExpertRevival(t *testing.T) {
	q := []float64{ProbabilityFloor, 1 - ProbabilityFloor}
	m, e := New(Config{Prior: []float64{.5, .5}, Hazard: 0, Trials: 200, Pending: 200}, 1)
	if e != nil {
		t.Fatal(e)
	}
	var tickets []Ticket
	for j := 0; j < 200; j++ {
		ticket, e := m.Issue(q, int64(j))
		if e != nil {
			t.Fatal(e)
		}
		tickets = append(tickets, ticket)
	}
	for j := 0; j < 200; j++ {
		r, e := m.Resolve(tickets[j], j >= 100, int64(200+j))
		if e != nil || r.Forecast != tickets[j].Forecast() {
			t.Fatal("late original forecast", e)
		}
	}
	l0 := 100 * (math.Log(q[0]) + math.Log1p(-q[0]))
	l1 := 100 * (math.Log(q[1]) + math.Log1p(-q[1]))
	w0 := 1 / (1 + math.Exp(l1-l0))
	got, e := m.Predict(q)
	if e != nil {
		t.Fatal(e)
	}
	near(t, got, w0*q[0]+(1-w0)*q[1])
	for i := 0; i < 2; i++ {
		if !finite(m.rows[199].logPosterior[i]) {
			t.Fatal("positive latent state permanently lost")
		}
	}
	s, e := New(Config{Prior: []float64{1e-320, 1}, Hazard: 0, Trials: 81, Pending: 1}, 1)
	if e != nil {
		t.Fatal(e)
	}
	qr := []float64{1 - ProbabilityFloor, ProbabilityFloor}
	for j := 0; j < 81; j++ {
		ticket, e := s.Issue(qr, int64(2*j))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = s.Resolve(ticket, j > 0, int64(2*j+1)); e != nil {
			t.Fatal(e)
		}
	}
	x := math.Log(1e-320) + math.Log1p(-qr[0]) + 80*math.Log(qr[0])
	y := math.Log1p(-qr[1]) + 80*math.Log(qr[1])
	posterior := 1 / (1 + math.Exp(y-x))
	got, e = s.Predict(qr)
	if e != nil {
		t.Fatal(e)
	}
	near(t, got, posterior*qr[0]+(1-posterior)*qr[1])
}
