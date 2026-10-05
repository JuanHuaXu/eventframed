package researchregimelog

import (
	"math"
	"testing"

	old "github.com/JuanHuaXu/eventframed/internal/researchregimeprotected"
)

func TestMemberOddsSurviveEndpointAndEvidenceReversal(t *testing.T) {
	base := []float64{.3, .8}
	cfg := Config{0, 0, 9}
	rows := make([]Row, 400)
	for i := range rows {
		v := 1
		if i >= 200 {
			v = 0
		}
		rows[i] = Row{0, v, -1}
	}
	u, e := Run(base, cfg, rows[:200])
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, c := range u.parts {
		if c.h%3 == 0 {
			if rate(c.odds[0]) != 1 || !finite(c.odds[0]) || c.odds[0] < 700 {
				t.Fatal("endpoint witness", c.odds[0])
			}
			found = true
		}
	}
	if !found {
		t.Fatal("vacuous endpoint test")
	}
	r, e := Run(base, cfg, rows)
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range r.parts {
		near(t, c.odds[0], logit(oracleHigh(base[0], c.h)))
		near(t, rate(c.odds[1]), oracleHigh(base[1], c.h))
	}
	// Independent closed form: balanced measurements have equal likelihood
	// under both rate atoms, so the member posterior returns to its prior.
	ls := make([]float64, 9)
	for h := range ls {
		eta := []float64{0, .1, .2}[h%3]
		q := eta + (1-2*eta)*.98
		pi := []float64{.1, .8, .1}[h/3] * []float64{.8, .1, .1}[h%3]
		ls[h] = math.Log(pi) + 200*(math.Log(q)+math.Log1p(-q))
	}
	logZ := logSum(ls)
	if d := math.Abs(r.LogEvidence - logZ); d > 5e-8 {
		t.Fatal("balanced log oracle", d)
	}
	for i := range base {
		expected := 0.
		for h, lp := range ls {
			mu := []float64{base[i], (9*base[i] - .5) / 8, 1 - base[i]}[h/3]
			expected += math.Exp(lp-logZ) * mu
		}
		near(t, r.Forecast[i], expected)
	}
	rr := make([]old.Row, len(rows))
	for i, v := range rows {
		rr[i] = old.Row{Member: v.Member, First: v.First, Second: v.Second}
	}
	v, e := old.Run(base, old.Config{Reset: 0, Hazard: 0, Cap: 9}, rr)
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(v.Forecast[0]-r.Forecast[0]) < .1 {
		t.Fatal("linear-endpoint negative was vacuous", v.Forecast[0], r.Forecast[0])
	}
	report.MemberOddsRescue = true
	t.Logf("member endpoint reversal: V81 %.9f, log/closed-form %.9f", v.Forecast[0], r.Forecast[0])
}

func TestTinyTransitionsAndBranchMassMetadata(t *testing.T) {
	for _, cfg := range []Config{{math.SmallestNonzeroFloat64, 0, 9}, {math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, 36}, {1, 1, 9}, {0, 0, 9}, {.5, .25, 36}} {
		m, e := New([]float64{.25, .925}, cfg)
		if e != nil {
			t.Fatal(e)
		}
		for i := 0; i < 64; i++ {
			if _, e = m.Issue(i%2, int64(i)); e != nil {
				t.Fatal(e)
			}
			if e = m.Reveal(m.token, i, 1, i%2, int64(i)); e != nil {
				t.Fatal(e)
			}
		}
		for _, i := range []int{0, 32, 63} {
			q, e := m.Pending(m.token, i, 2)
			if e != nil {
				t.Fatal("tiny transition query", cfg, i, e)
			}
			if math.Abs(logAdd(q.Branches[0].LogProbability, q.Branches[1].LogProbability)) > 2e-11 {
				t.Fatal("log branch mass")
			}
			for _, b := range q.Branches {
				if !finite(b.LogProbability) || b.Underflow != (b.Probability == 0) {
					t.Fatal("branch metadata")
				}
			}
			if e = m.Reveal(m.token, i, 2, 1-i%2, 64); e != nil {
				t.Fatal("tiny supported observation", e)
			}
		}
	}
}
