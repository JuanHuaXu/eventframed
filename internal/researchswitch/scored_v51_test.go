package researchswitch

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchscoreref"
)

var scoreModesV51 = []string{"log_mean", "log_strong", "brier_mean", "brier_strong"}

func TestScoredV51DenseDelayedReference(t *testing.T) {
	const steps = 96
	checks := 0
	for _, mode := range scoreModesV51 {
		for _, alpha := range []float64{0, 1. / 4096, 1. / 16, 1} {
			for _, prior := range [][]float64{{.9, .1}, {.8, .1, .1}} {
				m, err := newScoreV51(scoreConfigV51{Prior: prior, Hazard: alpha, Trials: steps, Pending: steps, Mode: mode}, 1)
				if err != nil {
					t.Fatal(err)
				}
				tape := make([][]float64, steps)
				tickets := make([]scoreTicketV51, steps)
				seen := map[int]bool{}
				for j := 0; j < steps; j++ {
					advice := make([]float64, len(prior))
					for h := range advice {
						advice[h] = .01 + .98*float64((j*17+h*31)%101)/100
					}
					tape[j] = append([]float64(nil), advice...)
					tickets[j], err = m.Issue(advice, int64(j))
					if err != nil {
						t.Fatal(err)
					}
					advice[0] = .99 // The retained nomination must own the old advice.
				}
				for j := 0; j < steps; j++ {
					id := (j * 37) % steps
					y := id%3 != 0
					r, err := m.Resolve(tickets[id], y, int64(steps+j))
					if err != nil || r.Forecast != tickets[id].Forecast() || r.TrialOrdinal != id+1 {
						t.Fatal("original served receipt", mode, err)
					}
					seen[id] = y
					want, err := researchscoreref.End(prior, alpha, tape, seen, mode)
					if err != nil {
						t.Fatal(err)
					}
					got := m.probabilities(m.logsAt(steps - 1))
					for h := range prior {
						if math.Abs(got[h]-want[h]) > 3e-12 {
							t.Fatal("dense original-position state", mode, alpha, j, h, got[h], want[h])
						}
						want[h] = (1-alpha)*want[h] + alpha*prior[h]
					}
					q, err := m.Predict(tape[0])
					ref, refErr := researchscoreref.Forecast(want, tape[0], mode)
					if err != nil || refErr != nil || math.Abs(q-ref) > 3e-12 {
						t.Fatal("independent next served law", mode, alpha, q, ref, err, refErr)
					}
					checks++
				}
			}
		}
	}
	t.Log("dense delayed state and independent next-law checks", checks)
}

func TestScoredV51LegacyExactAndCancellation(t *testing.T) {
	for _, alpha := range []float64{0, 1. / 150, 1} {
		cfg := compactConfigV48{Prior: []float64{.8, .1, .1}, Hazard: alpha, Trials: 32, Pending: 32}
		legacy, err := newCompactV48(cfg, 1)
		if err != nil {
			t.Fatal(err)
		}
		m, err := newScoreV51(scoreConfigV51{Prior: cfg.Prior, Hazard: alpha, Trials: 32, Pending: 32, Mode: "log_mean"}, 1)
		if err != nil {
			t.Fatal(err)
		}
		var old [32]compactTicketV48
		var next [32]scoreTicketV51
		for j := range old {
			advice := []float64{.01 + float64(j%9)/10, .2, .8}
			old[j], err = legacy.Issue(advice, int64(j))
			if err != nil {
				t.Fatal(err)
			}
			next[j], err = m.Issue(advice, int64(j))
			if err != nil || old[j].Forecast() != next[j].Forecast() {
				t.Fatal("legacy issued law differs", err)
			}
		}
		for j := 0; j < 32; j++ {
			id := j * 13 % 32
			if id%7 == 0 {
				if err = legacy.Cancel(old[id], int64(32+j)); err != nil {
					t.Fatal(err)
				}
				if err = m.Cancel(next[id], int64(32+j)); err != nil {
					t.Fatal(err)
				}
			} else {
				_, err = legacy.Resolve(old[id], id%3 == 0, int64(32+j))
				if err != nil {
					t.Fatal(err)
				}
				_, err = m.Resolve(next[id], id%3 == 0, int64(32+j))
				if err != nil {
					t.Fatal(err)
				}
			}
			if legacy.nextLogs() != m.nextLogs() || legacy.pending != m.pending {
				t.Fatal("legacy state differs")
			}
		}
	}
}

func TestScoredV51LifecycleAndFutureFork(t *testing.T) {
	for _, mode := range scoreModesV51 {
		cfg := scoreConfigV51{Prior: []float64{.9, .1}, Hazard: 1. / 16, Trials: 16, Pending: 16, Mode: mode}
		a, err := newScoreV51(cfg, 1)
		if err != nil {
			t.Fatal(err)
		}
		b, err := newScoreV51(cfg, 1)
		if err != nil {
			t.Fatal(err)
		}
		var x, y [16]scoreTicketV51
		for j := range x {
			advice := []float64{.1 + float64(j%3)/10, .9 - float64(j%5)/10}
			x[j], err = a.Issue(advice, int64(j))
			if err != nil {
				t.Fatal(err)
			}
			y[j], err = b.Issue(advice, int64(j))
			if err != nil || x[j].Forecast() != y[j].Forecast() {
				t.Fatal("unrevealed future reached law", err)
			}
		}
		before := append([]scoreRowV51(nil), a.rows...)
		if _, err := a.Resolve(y[0], true, 16); err == nil || !reflect.DeepEqual(before, a.rows) {
			t.Fatal("foreign ticket changed state")
		}
		for j := 0; j < 15; j++ {
			if _, err = a.Resolve(x[j], j%2 == 0, int64(16+j)); err != nil {
				t.Fatal(err)
			}
			if _, err = b.Resolve(y[j], j%2 == 0, int64(16+j)); err != nil {
				t.Fatal(err)
			}
			if a.nextLogs() != b.nextLogs() {
				t.Fatal("unobserved final label changed prefix")
			}
		}
		if _, err = a.Resolve(x[15], true, 31); err != nil {
			t.Fatal(err)
		}
		if _, err = b.Resolve(y[15], false, 31); err != nil {
			t.Fatal(err)
		}
		if err = a.BeginEpoch(2, 32); err != nil || a.mode != mode || a.Pending() != 0 {
			t.Fatal("epoch mode or pending", err)
		}
		if _, err = a.Resolve(x[0], true, 33); err == nil {
			t.Fatal("stale ticket accepted")
		}
	}
	for _, mode := range []string{"", "brier", "LOG_MEAN"} {
		if _, err := newScoreV51(scoreConfigV51{Prior: []float64{.9, .1}, Hazard: 0, Trials: 1, Pending: 1, Mode: mode}, 1); err == nil {
			t.Fatal("undeclared score mode accepted")
		}
	}
}

func TestScoredV51HybridLegacyExactAndFencing(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 32, Pending: 32}
	for _, alpha := range []float64{0, 1. / 2400, 1. / 150} {
		old, err := NewMemoHybridV49([]float64{.3, .8}, cfg, 16, 1./2400, alpha, 1)
		if err != nil {
			t.Fatal(err)
		}
		next, err := NewScoredHybridV51([]float64{.3, .8}, cfg, 16, 1./2400, alpha, 1, "log_mean")
		if err != nil {
			t.Fatal(err)
		}
		var x [32]HybridTicket
		var y [32]ScoredHybridTicketV51
		for j := range x {
			x[j], err = old.Issue(j%2, int64(j))
			if err != nil {
				t.Fatal(err)
			}
			y[j], err = next.Issue(j%2, int64(j))
			if err != nil || x[j].Forecast() != y[j].Forecast() {
				t.Fatal("whole served law differs", err)
			}
		}
		for j := 0; j < 32; j++ {
			id := j * 13 % 32
			a, err := old.Resolve(x[id], id%3 != 0, int64(32+j))
			if err != nil {
				t.Fatal(err)
			}
			b, err := next.Resolve(y[id], id%3 != 0, int64(32+j))
			if err != nil || a != b {
				t.Fatal("whole original receipt differs", a, b, err)
			}
			for member := 0; member < 2; member++ {
				qa, err := old.Predict(member)
				if err != nil {
					t.Fatal(err)
				}
				qb, err := next.Predict(member)
				if err != nil || qa != qb {
					t.Fatal("whole delayed law differs", err)
				}
			}
		}
	}
	for _, mode := range scoreModesV51 {
		p, err := NewScoredHybridV51([]float64{.3, .8}, cfg, 16, 1./2400, 1./150, 1, mode)
		if err != nil {
			t.Fatal(err)
		}
		p.scope.used = len(p.scope.rows) // Inject unexpected partial admission failure.
		if _, err = p.Issue(0, 0); err == nil || !p.local.poison {
			t.Fatal("partial bundle not fenced")
		}
		if _, err = p.Predict(0); err == nil {
			t.Fatal("fenced prediction allowed")
		}
		if err = p.BeginEpoch(2, 1); err != nil || p.local.poison || p.scope.mode != mode {
			t.Fatal("fresh bundle mode recovery", err)
		}
	}
}
