// Read-only cross-representation diagnosis; no learner/gate tuning.
package main

import (
	"fmt"
	"math"

	l "github.com/JuanHuaXu/eventframed/internal/researchregimelog"
	v "github.com/JuanHuaXu/eventframed/internal/researchregimeprotected"
)

func main() {
	for _, cfg := range []l.Config{{1. / 16, .25, 9}, {.5, 0, 18}, {1, .25, 36}, {0, 1, 9}, {1. / 16, .25, 0}} {
		m, e := l.New([]float64{.3, .8}, cfg)
		if e != nil {
			panic(e)
		}
		n := 320
		if cfg.Cap == 0 {
			n = 64
		}
		for i := 0; i < n; i++ {
			if _, e = m.Issue(i%2, int64(i)); e != nil {
				panic(e)
			}
		}
		targets := []int{255, 64, 63, 128, 319, 65, 0, 256}
		if cfg.Cap == 0 {
			targets = []int{0}
		}
		for _, target := range targets {
			if e = m.Reveal(m.Snapshot().Token, target, 1, target%2, int64(n)); e != nil {
				panic(e)
			}
			s := m.Snapshot()
			rows := make([]v.Row, len(s.Rows))
			keys := make([][]v.Key, len(s.Support))
			for i, r := range s.Rows {
				rows[i] = v.Row{Member: r.Member, First: r.First, Second: r.Second}
			}
			for i, ks := range s.Support {
				for _, k := range ks {
					keys[i] = append(keys[i], v.Key{Class: k.Class, Start: k.Start})
				}
			}
			r, e := v.Conditional([]float64{.3, .8}, v.Config{Reset: cfg.Reset, Hazard: cfg.Hazard, Cap: cfg.Cap}, rows, keys)
			if e != nil {
				panic(e)
			}
			cold, e := l.Conditional([]float64{.3, .8}, cfg, s.Rows, s.Support)
			if e != nil {
				panic(e)
			}
			max := 0.
			for i, p := range s.NextClean {
				max = math.Max(max, math.Abs(p-r.Forecast[i]))
			}
			fmt.Printf("cfg=%+v target=%d candidateEnvelope=%g oldEnvelope=%g coldLogEnvelope=%g maxForecastDefect=%g logEvidenceDefect=%g candidateDiscard=%g oldDiscard=%g\n", cfg, target, s.TVEnvelope, r.Envelope, cold.Envelope, max, math.Abs(s.LogEvidence-r.LogEvidence), cold.Discard, r.Discard)
		}
	}
}
