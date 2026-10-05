package researchregimefrozen

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
)

type evidence struct {
	Checks, Configurations int
	MaxExactDefect         float64
	MaxProjectionTV        float64
	MaxBeamTV              float64
	MaxBeamEnvelope        float64
	VacuousBeamEnvelopes   int
	BeamCases              int
	IndependentPathChecks  int
	MaxExactTowerDefect    float64
	MaxBeamTowerDefect     float64
	FrozenChecks           int
	MaxFrozenTowerDefect   float64
	MaxFrozenOracleDefect  float64
	WholeGoalValidation    bool
}

var report evidence

func TestMain(m *testing.M) {
	code := m.Run()
	if p := os.Getenv("EVENTFRAME_REGIME_V78_REPORT"); p != "" {
		b, e := json.MarshalIndent(report, "", "  ")
		if e == nil {
			e = os.WriteFile(p, append(b, '\n'), 0600)
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			code = 1
		}
	}
	os.Exit(code)
}

func near(t *testing.T, a, b float64) {
	t.Helper()
	if !finite(a) || !finite(b) || math.Abs(a-b) > 2e-11 {
		t.Fatalf("scalar disagreement %.17g %.17g", a, b)
	}
	report.Checks++
	report.MaxExactDefect = math.Max(report.MaxExactDefect, math.Abs(a-b))
}

func tv(a, b []float64) float64 {
	s := 0.
	for i, p := range a {
		s += math.Abs(p - b[i])
	}
	return s / 2
}

func compareDense(t *testing.T, base []float64, cfg Config, rows []Row) Result {
	t.Helper()
	r, e := Run(base, cfg, rows)
	if e != nil {
		t.Fatal(e)
	}
	p, z := dense(base, cfg, rows, false)
	u, e := r.Joint(len(base))
	if e != nil {
		t.Fatal(e)
	}
	for i := range p {
		near(t, u[i], p[i])
	}
	near(t, r.LogEvidence, z)
	for i, q := range r.Forecast {
		near(t, q, denseForecast(base, cfg, p, i))
	}
	return r
}

func TestIndependentTransitionsAndLatentPaths(t *testing.T) {
	base := []float64{.25, .925}
	for _, kappa := range []float64{0, 1. / 16, .5, 1} {
		for _, lambda := range []float64{0, .25, 1} {
			cfg := Config{kappa, lambda, 0}
			prior := oraclePrior(base)
			for from := range prior {
				s := 0.
				for to := range prior {
					s += oracleTransition(from, to, 0, base, cfg, prior)
				}
				near(t, s, 1)
			}
			rows := []Row{{0, 1, 0}, {1, 0, -1}}
			p, z := enumerate(base, cfg, rows)
			q, w := dense(base, cfg, rows, false)
			for i := range p {
				near(t, p[i], q[i])
				report.IndependentPathChecks++
			}
			near(t, z, w)
			compareDense(t, base, cfg, rows)
		}
	}
}

func TestExactDelayedJointAndProjectionControl(t *testing.T) {
	all := []float64{.25, .47, .7, .925}
	for members := 1; members <= 4; members++ {
		base := all[:members]
		for _, kappa := range []float64{0, 1. / 16, .5, 1} {
			for _, lambda := range []float64{0, .25, 1} {
				cfg := Config{kappa, lambda, 0}
				report.Configurations++
				rows := make([]Row, 12)
				for i := range rows {
					rows[i] = Row{i % members, -1, -1}
				}
				compareDense(t, base, cfg, rows)
				// Original issue order remains unchanged as labels arrive in reverse.
				for _, i := range []int{8, 0, 5, 2, 11, 4, 1, 9} {
					rows[i].First = i % 2
					compareDense(t, base, cfg, rows)
				}
				for _, i := range []int{0, 8, 11} {
					rows[i].Second = (i + 1) % 2
					compareDense(t, base, cfg, rows)
				}
				p, _ := dense(base, cfg, rows, false)
				q, _ := dense(base, cfg, rows, true)
				d := tv(p, q)
				if members == 1 || kappa == 0 || kappa == 1 {
					near(t, d, 0)
				} else {
					report.MaxProjectionTV = math.Max(report.MaxProjectionTV, d)
				}
			}
		}
	}
	if report.MaxProjectionTV < 1e-5 {
		t.Fatal("projection dependence negative control was vacuous")
	}
	t.Logf("%d configurations: max conditional-product TV defect %g", report.Configurations, report.MaxProjectionTV)
}

func TestBeamErrorEnvelopeAndCaps(t *testing.T) {
	base := []float64{.25, .7, .925}
	for _, kappa := range []float64{1. / 16, .5, 1} {
		rows := []Row{}
		for n := 0; n < 16; n++ {
			r := Row{n % len(base), n % 2, -1}
			if n%4 == 0 {
				r.Second = 1 - r.First
			}
			rows = append(rows, r)
			p, _ := dense(base, Config{kappa, .25, 0}, rows, false)
			for _, cap := range []int{9, 18, 36} {
				r, e := Run(base, Config{kappa, .25, cap}, rows)
				if e != nil || r.Components > cap || r.MaxComponents > cap+Classes {
					t.Fatal("beam cap/support", r.Components, r.MaxComponents, e)
				}
				q, e := r.Joint(len(base))
				if e != nil {
					t.Fatal(e)
				}
				d, sum := tv(p, q), 0.
				for _, w := range q {
					if !finite(w) || w < 0 {
						t.Fatal("beam invalid mass")
					}
					sum += w
				}
				near(t, sum, 1)
				if d > r.Envelope+2e-11 {
					t.Fatal("beam envelope did not cover actual joint TV", d, r.Envelope)
				}
				report.BeamCases++
				report.MaxBeamTV = math.Max(report.MaxBeamTV, d)
				report.MaxBeamEnvelope = math.Max(report.MaxBeamEnvelope, r.Envelope)
				if r.Envelope >= 1 {
					report.VacuousBeamEnvelopes++
				}
			}
		}
	}
	t.Logf("%d beam cases: max TV %g, vacuous envelopes %d", report.BeamCases, report.MaxBeamTV, report.VacuousBeamEnvelopes)
}

func TestActualPairBranchesAndTower(t *testing.T) {
	base := []float64{.3, .8}
	rows := []Row{{0, 1, -1}, {1, 0, -1}, {0, -1, -1}, {1, 1, -1}, {0, 0, -1}}
	for _, kappa := range []float64{0, 1. / 16, .5, 1} {
		for _, cap := range []int{0, 9, 18} {
			cfg := Config{kappa, .25, cap}
			before, e := Run(base, cfg, rows)
			if e != nil {
				t.Fatal(e)
			}
			for _, target := range []int{0, 4} {
				var branches [2]Result
				var probabilities [2]float64
				for value := 0; value < 2; value++ {
					copyRows := append([]Row(nil), rows...)
					copyRows[target].Second = value
					branches[value], e = Run(base, cfg, copyRows)
					if e != nil {
						t.Fatal(e)
					}
					probabilities[value] = math.Exp(branches[value].LogEvidence - before.LogEvidence)
				}
				mass := probabilities[0] + probabilities[1]
				defect := math.Abs(mass - 1)
				for i, q := range before.Forecast {
					after := 0.
					for value, branch := range branches {
						after += probabilities[value] / mass * branch.Forecast[i]
					}
					defect = math.Max(defect, math.Abs(after-q))
				}
				if cap == 0 {
					near(t, defect, 0)
					report.MaxExactTowerDefect = math.Max(report.MaxExactTowerDefect, defect)
				} else {
					report.MaxBeamTowerDefect = math.Max(report.MaxBeamTowerDefect, defect)
				}
			}
		}
	}
	t.Logf("exact tower %g; capped tower/evidence-mass defect %g (not an exact scored-law certificate)", report.MaxExactTowerDefect, report.MaxBeamTowerDefect)
}

func TestJournalFuturePhasesAndAtomicFailures(t *testing.T) {
	base := []float64{.3, .8}
	cfg := Config{1. / 16, .25, 18}
	a, e := New(base, cfg)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := New(base, cfg)
	for tick := 0; tick < 12; tick++ {
		_, x, e := a.Issue(tick%2, int64(tick))
		if e != nil {
			t.Fatal(e)
		}
		_, y, e := b.Issue(tick%2, int64(tick))
		if e != nil {
			t.Fatal(e)
		}
		near(t, x, y)
	}
	for _, ordinal := range []int{5, 0, 9} {
		if e = a.Reveal(ordinal, 1, ordinal%2, 12); e != nil {
			t.Fatal(e)
		}
		if e = b.Reveal(ordinal, 1, ordinal%2, 12); e != nil {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("identical as-of histories differ")
	}
	copyState := fmt.Sprintf("%#v%#v%#v", a.rows, a.state, a.clock)
	for _, bad := range [][4]int{{5, 1, 1, 12}, {1, 2, 1, 12}, {12, 1, 1, 12}, {1, 1, 2, 12}, {1, 1, 1, 11}} {
		if e = a.Reveal(bad[0], bad[1], bad[2], int64(bad[3])); e == nil || copyState != fmt.Sprintf("%#v%#v%#v", a.rows, a.state, a.clock) {
			t.Fatal("failed reveal published or accepted", bad, e)
		}
	}
	// Hypothetical branch calculations cannot modify journal or input rows.
	hypothetical := append([]Row(nil), a.rows...)
	hypothetical[5].Second = 1
	if _, e = Run(base, cfg, hypothetical); e != nil || copyState != fmt.Sprintf("%#v%#v%#v", a.rows, a.state, a.clock) {
		t.Fatal("hypothetical publication", e)
	}
	if e = a.Reveal(1, 1, 0, 13); e != nil {
		t.Fatal(e)
	}
	if e = b.Reveal(1, 1, 1, 13); e != nil {
		t.Fatal(e)
	}
	if reflect.DeepEqual(a.state.Forecast, b.state.Forecast) {
		t.Fatal("future-label fork was vacuous after reveal")
	}
	// Zero-noise opposite measurements have zero likelihood for SAME Y.
	for bit := 0; bit < 2; bit++ {
		near(t, likelihood(atom(bit), 0, Row{0, 1, 0}), 0)
		if oracleFactor(bit, 1, Row{0, 1, 0}) != 0 {
			t.Fatal("oracle impossible pair")
		}
	}
	for _, invalid := range []Config{{math.NaN(), .25, 9}, {.1, math.Inf(1), 9}, {.1, .25, 1}, {.1, .25, 257}} {
		if _, e = New(base, invalid); e == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}

func TestFullJournalDelayedPairs(t *testing.T) {
	base := []float64{.25, .925}
	cfg := Config{1. / 16, .25, 0}
	rows := make([]Row, 64)
	for i := range rows {
		rows[i] = Row{i % 2, -1, -1}
	}
	for _, i := range []int{63, 0, 32, 17, 48, 1} {
		rows[i].First = i % 2
		compareDense(t, base, cfg, rows)
	}
	for _, i := range []int{0, 32, 63} {
		rows[i].Second = 1 - rows[i].First
		compareDense(t, base, cfg, rows)
	}
	t.Logf("%d counted exact scalar checks; max defect %g", report.Checks, report.MaxExactDefect)
}
