package researchregimefrozen

import (
	"math"
	"reflect"
	"testing"

	original "github.com/JuanHuaXu/eventframed/internal/researchregime"
)

type hidden struct{ state, start int }

// Independent dense trajectory with last-reset index and explicit state masks.
func restrictedDense(base []float64, cfg Config, rows []Row, support [][]Key) ([]float64, float64) {
	prior := oraclePrior(base)
	width := 1 << len(base)
	p := map[hidden]float64{}
	for s, w := range prior {
		p[hidden{s, -1}] = w
	}
	logZ := 0.
	for t, row := range rows {
		next := map[hidden]float64{}
		for from, w := range p {
			for to, pi := range prior {
				if cfg.Reset > 0 {
					next[hidden{to, t}] += w * cfg.Reset * pi
				}
				fh, fm, th, tm := from.state/width, from.state%width, to/width, to%width
				if cfg.Reset < 1 && fh == th && fm&^(1<<row.Member) == tm&^(1<<row.Member) {
					x := oracleHigh(base[row.Member], th)
					if tm&(1<<row.Member) == 0 {
						x = 1 - x
					}
					q := cfg.Hazard * x
					if fm == tm {
						q += 1 - cfg.Hazard
					}
					next[hidden{to, from.start}] += w * (1 - cfg.Reset) * q
				}
			}
		}
		allowed := map[Key]bool{}
		for _, k := range support[t] {
			allowed[k] = true
		}
		z := 0.
		for s, w := range next {
			if !allowed[Key{s.state / width, s.start}] {
				delete(next, s)
				continue
			}
			w *= oracleFactor(s.state, len(base), row)
			next[s] = w
			z += w
		}
		for s := range next {
			next[s] /= z
		}
		logZ += math.Log(z)
		p = next
	}
	out := make([]float64, len(prior))
	for s, w := range p {
		out[s.state] += w
	}
	return out, logZ
}

func frozenCompare(t *testing.T, base []float64, cfg Config, rows []Row, support [][]Key) Result {
	t.Helper()
	r, e := Conditional(base, cfg, rows, support)
	if e != nil {
		t.Fatal(e)
	}
	p, z := restrictedDense(base, cfg, rows, support)
	q, e := r.Joint(len(base))
	if e != nil {
		t.Fatal(e)
	}
	for i := range p {
		near(t, p[i], q[i])
		report.FrozenChecks++
		report.MaxFrozenOracleDefect = math.Max(report.MaxFrozenOracleDefect, math.Abs(p[i]-q[i]))
	}
	near(t, z, r.LogEvidence)
	report.FrozenChecks++
	report.MaxFrozenOracleDefect = math.Max(report.MaxFrozenOracleDefect, math.Abs(z-r.LogEvidence))
	return r
}

func TestFrozenSupportPairConditioningAndTower(t *testing.T) {
	base := []float64{.3, .8}
	rows := []Row{{0, 1, -1}, {1, 0, -1}, {0, -1, -1}, {1, 1, -1}, {0, 0, -1}}
	for _, kappa := range []float64{0, 1. / 16, .5, 1} {
		for _, cap := range []int{9, 18, 36} {
			cfg := Config{kappa, .25, cap}
			before, e := Run(base, cfg, rows)
			if e != nil {
				t.Fatal(e)
			}
			frozenCompare(t, base, cfg, rows, before.Support)
			originalSupport := make([][]Key, len(before.Support))
			for i, ks := range before.Support {
				originalSupport[i] = append([]Key(nil), ks...)
			}
			for _, target := range []int{0, 4} {
				var branches [2]Result
				var prob [2]float64
				for value := 0; value < 2; value++ {
					x := append([]Row(nil), rows...)
					x[target].Second = value
					branches[value] = frozenCompare(t, base, cfg, x, before.Support)
					prob[value] = math.Exp(branches[value].LogEvidence - before.LogEvidence)
				}
				defect := math.Abs(prob[0] + prob[1] - 1)
				for i, q := range before.Forecast {
					weighted := prob[0]*branches[0].Forecast[i] + prob[1]*branches[1].Forecast[i]
					defect = math.Max(defect, math.Abs(weighted-q))
				}
				near(t, defect, 0)
				report.FrozenChecks++
				report.MaxFrozenTowerDefect = math.Max(report.MaxFrozenTowerDefect, defect)
			}
			if !reflect.DeepEqual(before.Support, originalSupport) {
				t.Fatal("conditional query mutated frozen support")
			}
		}
	}
	t.Logf("%d frozen-support reference checks, max oracle %g, tower %g", report.FrozenChecks, report.MaxFrozenOracleDefect, report.MaxFrozenTowerDefect)
}

func TestOriginalPosteriorPreservedButEvidenceAccountingChanges(t *testing.T) {
	base := []float64{.3, .8}
	rows := []Row{{0, 1, -1}, {1, 0, -1}, {0, -1, -1}, {1, 1, -1}, {0, 0, -1}}
	oldRows := make([]original.Row, len(rows))
	for i, r := range rows {
		oldRows[i] = original.Row{Member: r.Member, First: r.First, Second: r.Second}
	}
	changed := false
	for _, cap := range []int{0, 9, 18, 36} {
		x, e := Run(base, Config{.5, .25, cap}, rows)
		if e != nil {
			t.Fatal(e)
		}
		y, e := original.Run(base, original.Config{Reset: .5, Hazard: .25, Cap: cap}, oldRows)
		if e != nil {
			t.Fatal(e)
		}
		for i := range base {
			near(t, x.Forecast[i], y.Forecast[i])
		}
		if cap == 0 {
			near(t, x.LogEvidence, y.LogEvidence)
		} else if math.Abs(x.LogEvidence-y.LogEvidence) > 1e-4 {
			changed = true
		}
	}
	if !changed {
		t.Fatal("retained-path evidence accounting test was vacuous")
	}
}

func TestFrozenSupportBindingsFailClosed(t *testing.T) {
	base := []float64{.3, .8}
	cfg := Config{.5, .25, 9}
	rows := []Row{{0, 1, -1}, {1, 0, -1}}
	r, e := Run(base, cfg, rows)
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][][]Key{
		nil, {nil, r.Support[1]},
		{{{Class: 0, Start: 5}}, r.Support[1]},
		{{{Class: 9, Start: -1}}, r.Support[1]},
		{{{Class: 0, Start: -1}, {Class: 0, Start: -1}}, r.Support[1]},
	} {
		if _, e = Conditional(base, cfg, rows, bad); e == nil {
			t.Fatal("invalid frozen support accepted", bad)
		}
	}
	if _, e = Conditional(base, cfg, []Row{{0, 1, 0}}, [][]Key{{{Class: 0, Start: 0}}}); e == nil {
		t.Fatal("unsupported branch reopened or invented a likelihood")
	}
}
