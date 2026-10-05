package researchregimelogsummary

import (
	"math"
	"testing"
)

func TestAnalyticEnvelopeAgainstIndependentUnrestrictedDense(t *testing.T) {
	for _, cfg := range []Config{{1. / 16, .25, 9}, {.5, 0, 18}, {1, .25, 36}, {0, 1, 9}, {1. / 16, .25, 0}} {
		m, e := New([]float64{.3, .8}, cfg)
		if e != nil {
			t.Fatal(e)
		}
		check := func() {
			if (cfg.Cap == 0 || cfg.Reset == 0 || cfg.Reset == 1) && (m.state.Discard != 0 || m.state.Envelope != 0) {
				t.Fatal("phantom truncation", cfg, m.state.Discard, m.state.Envelope)
			}
			full, _ := dense(m.base, m.cfg, m.rows, false)
			u, e := m.state.Joint(2)
			if e != nil {
				t.Fatal(e)
			}
			tv := 0.
			for i, p := range full {
				tv += math.Abs(p - u[i])
			}
			tv /= 2
			if tv > m.state.Envelope+2e-11 {
				t.Fatal("analytic truncation envelope violated", cfg, tv, m.state.Envelope)
			}
			report.EnvelopeOracleChecks++
			if m.state.Envelope < 1 {
				report.NonvacuousEnvelopes++
			}
			report.MaxEnvelopeActualTV = math.Max(report.MaxEnvelopeActualTV, tv)
		}
		for i := 0; i < 40; i++ {
			if _, e = m.Issue(i%2, int64(i)); e != nil {
				t.Fatal(e)
			}
			check()
			if i >= 3 && i%5 == 3 {
				target := i - 3
				if e = m.Reveal(m.token, target, 1, target%2, int64(i)); e != nil {
					t.Fatal(e)
				}
				check()
				if e = m.Reveal(m.token, target, 2, 1-target%2, int64(i)); e != nil {
					t.Fatal(e)
				}
				check()
			}
		}
	}
}
