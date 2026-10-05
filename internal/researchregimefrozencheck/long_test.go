package researchregimefrozencheck

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	model "github.com/JuanHuaXu/eventframed/internal/researchregimefrozen"
)

type Config = model.Config
type Row = model.Row
type Key = model.Key
type hidden struct{ state, start int }

func restrictedDense(base []float64, cfg Config, rows []Row, support [][]Key) ([]float64, float64) {
	prior, width := oraclePrior(base), 1<<len(base)
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
			if !allowed[Key{Class: s.state / width, Start: s.start}] {
				delete(next, s)
				continue
			}
			w *= oracleFactor(s.state, len(base), row)
			next[s] = w
			z += w
		}
		if z == 0 {
			return nil, math.Inf(-1)
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

func TestFull64FrozenSupportBranches(t *testing.T) {
	base := []float64{.25, .925}
	checks, cases, zeros := 0, 0, 0
	maxOracle, maxTower := 0., 0.
	near := func(a, b float64) {
		checks++
		if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) || math.Abs(a-b) > 2e-11 {
			t.Fatalf("comparison %g %g", a, b)
		}
		maxOracle = math.Max(maxOracle, math.Abs(a-b))
	}
	for _, kappa := range []float64{0, 1. / 16, .5, 1} {
		for _, cap := range []int{9, 18, 36} {
			cfg := Config{Reset: kappa, Hazard: .25, Cap: cap}
			rows := make([]Row, 64)
			for i := range rows {
				rows[i] = Row{Member: i % 2, First: -1, Second: -1}
			}
			// The revealing labels arrive out of original order; 58 others stay
			// unknown and still advance the original issue clock.
			for _, i := range []int{63, 0, 32, 17, 48, 1} {
				rows[i].First = i % 2
			}
			before, e := model.Run(base, cfg, rows)
			if e != nil {
				t.Fatal(e)
			}
			for _, target := range []int{0, 32, 63} {
				cases++
				var prob [2]float64
				var branches [2]model.Result
				for value := 0; value < 2; value++ {
					x := append([]Row(nil), rows...)
					x[target].Second = value
					p, z := restrictedDense(base, cfg, x, before.Support)
					branches[value], e = model.Conditional(base, cfg, x, before.Support)
					if math.IsInf(z, -1) {
						if e == nil {
							t.Fatal("dense-impossible branch accepted")
						}
						zeros++
						continue
					}
					if e != nil {
						t.Fatal("supported dense branch rejected", e)
					}
					q, e := branches[value].Joint(len(base))
					if e != nil {
						t.Fatal(e)
					}
					for i := range p {
						near(q[i], p[i])
					}
					near(branches[value].LogEvidence, z)
					prob[value] = math.Exp(z - before.LogEvidence)
				}
				defect := math.Abs(prob[0] + prob[1] - 1)
				for i, q := range before.Forecast {
					weighted := 0.
					for value, p := range prob {
						if p > 0 {
							weighted += p * branches[value].Forecast[i]
						}
					}
					defect = math.Max(defect, math.Abs(weighted-q))
				}
				near(defect, 0)
				maxTower = math.Max(maxTower, defect)
			}
		}
	}
	t.Logf("%d full64 pair cases; %d scalar checks; max oracle %g; tower %g; independently verified zero branches %d", cases, checks, maxOracle, maxTower, zeros)
	if path := os.Getenv("EVENTFRAME_REGIME_V78_REPORT"); path != "" {
		b, e := json.MarshalIndent(map[string]any{"cases": cases, "checks": checks, "maxOracle": maxOracle, "maxTower": maxTower, "zeroBranches": zeros, "wholeGoalValidation": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, append(b, '\n'), 0600); e != nil {
			t.Fatal(e)
		}
	}
}

var sink model.Result

func BenchmarkFull64Conditional(b *testing.B) {
	base := []float64{.25, .925}
	cfg := Config{Reset: 1. / 16, Hazard: .25, Cap: 36}
	rows := make([]Row, 64)
	for i := range rows {
		rows[i] = Row{Member: i % 2, First: -1, Second: -1}
	}
	for _, i := range []int{63, 0, 32, 17, 48, 1} {
		rows[i].First = i % 2
	}
	r, e := model.Run(base, cfg, rows)
	if e != nil {
		b.Fatal(e)
	}
	rows[0].Second = 1
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		x, e := model.Conditional(base, cfg, rows, r.Support)
		if e != nil {
			b.Fatal(e)
		}
		sink = x
	}
}
