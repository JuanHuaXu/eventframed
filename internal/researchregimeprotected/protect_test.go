package researchregimeprotected

import (
	"math"
	"testing"
)

func TestProtectedKeysAndRestartOnlyNegativeControl(t *testing.T) {
	for _, reset := range []float64{0, 1. / 16, .5, 1} {
		for _, cap := range []int{9, 18, 36} {
			m, e := New([]float64{.3, .8}, Config{reset, .25, cap})
			if e != nil {
				t.Fatal(e)
			}
			prior, e := Run(m.base, Config{1, .25, 9}, nil)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 24; i++ {
				r, e := m.Issue(i%2, int64(i))
				if e != nil {
					t.Fatal(e)
				}
				keys := m.state.Support[i]
				if len(keys) > cap {
					t.Fatal("protected cap exceeded")
				}
				for h := 0; h < Classes; h++ {
					start := i
					if reset == 0 {
						start = -1
					}
					found := false
					for _, k := range keys {
						if k.Class == h && k.Start == start {
							found = true
						}
					}
					if !found {
						t.Fatal("fresh model class lost", reset, cap, i, h)
					}
					report.ProtectedChecks++
				}
				if cap == 9 && reset > 0 {
					near(t, r.Clean, prior.Forecast[i%2])
				}
				if e = m.Reveal(m.token, i, 1, i%2, int64(i)); e != nil {
					t.Fatal(e)
				}
				q, e := m.Pending(m.token, i, 2)
				if e != nil {
					t.Fatal(e)
				}
				if q.Branches[0].Probability <= 0 || q.Branches[1].Probability <= 0 {
					t.Fatal("protected evidence lacks a branch")
				}
				if e = m.Reveal(m.token, i, 2, 1-i%2, int64(i)); e != nil {
					t.Fatal("supported contradictory pair", e)
				}
			}
		}
	}
}

func TestProtectedOldAndLatestContradictions(t *testing.T) {
	for _, reset := range []float64{1. / 16, .5, 1} {
		for _, cap := range []int{9, 18, 36} {
			m, e := New([]float64{.3, .8}, Config{reset, .25, cap})
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				if _, e = m.Issue(i%2, int64(i)); e != nil {
					t.Fatal(e)
				}
			}
			for _, i := range []int{63, 0, 32, 17} {
				if e = m.Reveal(m.token, i, 1, i%2, 64); e != nil {
					t.Fatal(e)
				}
				if _, ok := queried(t, m, i, 2); !ok {
					t.Fatal("old/latest protected query abstained")
				}
				if e = m.Reveal(m.token, i, 2, 1-i%2, 64); e != nil {
					t.Fatal(e)
				}
				inspect(t, m)
			}
		}
	}
}

func logSum(values []float64) float64 {
	maxLog := math.Inf(-1)
	for _, v := range values {
		maxLog = math.Max(maxLog, v)
	}
	sum := 0.
	for _, v := range values {
		sum += math.Exp(v - maxLog)
	}
	return maxLog + math.Log(sum)
}

// Independent closed form: with lambda1,kappa0 all events are independent
// conditional on H. The contradictory pair has likelihood eta*(1-eta).
func TestNoResetLongUnderflowNegativeControl(t *testing.T) {
	base := []float64{.925}
	cfg := Config{0, 1, 9}
	rows := make([]Row, MaxRows)
	for i := range rows {
		rows[i] = Row{0, 1, -1}
	}
	r, e := Run(base, cfg, rows)
	if e != nil {
		t.Fatal(e)
	}
	first, pair := make([]float64, Classes), make([]float64, Classes)
	for h := 0; h < Classes; h++ {
		mu := []float64{.925, (9*.925 - .5) / 8, 1 - .925}[h/3]
		eta := []float64{0, .1, .2}[h%3]
		prior := []float64{.1, .8, .1}[h/3] * []float64{.8, .1, .1}[h%3]
		p := eta + (1-2*eta)*mu
		first[h] = math.Log(prior) + float64(MaxRows)*math.Log(p)
		pair[h] = math.Log(prior) + float64(MaxRows-1)*math.Log(p) + math.Log(eta*(1-eta))
	}
	logP := logSum(pair) - logSum(first)
	if !finite(logP) || logP >= math.Log(math.SmallestNonzeroFloat64) {
		t.Fatal("long rare-evidence witness was vacuous", logP)
	}
	rows[MaxRows-1].Second = 0
	if _, e = Conditional(base, cfg, rows, r.Support); e == nil {
		t.Fatal("expected documented linear-weight underflow not reproduced")
	}
	report.NoResetUnderflow = true
	report.ExactLongPairLogProbability = logP
	t.Logf("NEGATIVE: real pair probability is positive, logp=%g; linear filtering rejects through underflow, not true impossible evidence", logP)
}
