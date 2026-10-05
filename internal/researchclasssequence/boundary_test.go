package researchclasssequence

import "testing"

func TestFullJournalRareOpposition(t *testing.T) {
	for _, mode := range []string{"shared", "hybrid"} {
		base := make([]float64, 150)
		for i := range base {
			base[i] = .25 + .675*float64(i)/149
		}
		m, e := New(base, 1, 2800, Config{Mode: mode, Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		var last Ticket
		for n := 0; n < len(base)*MaxTrials; n++ {
			last, e = m.Issue(n%len(base), int64(n))
			if e != nil {
				t.Fatal(mode, n, e)
			}
			if _, e = m.Resolve(last, true, int64(n)); e != nil {
				t.Fatal(mode, n, e)
			}
		}
		audit, e := m.RequestSecond(last, int64(len(base)*MaxTrials))
		if e != nil {
			t.Fatal(mode, e)
		}
		if _, e = m.Resolve(audit, false, int64(len(base)*MaxTrials+1)); e != nil {
			t.Fatal("positive-probability rare opposition must remain assimilable", mode, e)
		}
		for i := range base {
			if _, _, e = m.Predict(i); e != nil {
				t.Fatal(e)
			}
		}
	}
}
