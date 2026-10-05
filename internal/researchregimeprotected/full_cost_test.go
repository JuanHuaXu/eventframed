package researchregimeprotected

import (
	"fmt"
	"reflect"
	"testing"
)

// Setup uses one cold replay of unknown rows, not repeated Issue calls. It is
// outside timing and is NOT a complete ingestion/serving benchmark.
func coldFixture(t testing.TB, members, rows int) *Ledger {
	t.Helper()
	base := make([]float64, members)
	for i := range base {
		base[i] = .25 + .675*float64(i)/float64(max(1, members-1))
	}
	cfg := Config{1. / 16, .25, 36}
	m, e := New(base, cfg)
	if e != nil {
		t.Fatal(e)
	}
	journal := make([]Row, rows)
	for i := range journal {
		journal[i] = Row{i % members, -1, -1}
	}
	r, e := Run(base, cfg, journal)
	if e != nil {
		t.Fatal(e)
	}
	r, cache, e := m.rebuild(journal, r.Support)
	if e != nil {
		t.Fatal(e)
	}
	m.publish(journal, r, int64(rows), uint64(rows))
	m.cache = cache
	return m
}

func TestColdUnknownSetupMatchesIssuedPrefix(t *testing.T) {
	for _, members := range []int{2, 150, 200} {
		m := coldFixture(t, members, 32)
		n, e := New(m.base, m.cfg)
		if e != nil {
			t.Fatal(e)
		}
		for i := 0; i < 32; i++ {
			if _, e = n.Issue(i%members, int64(i)); e != nil {
				t.Fatal(e)
			}
		}
		if !reflect.DeepEqual(m.state.Support, n.state.Support) {
			t.Fatal("cold setup re-ranked a different support")
		}
		for i, p := range m.state.Forecast {
			near(t, p, n.state.Forecast[i])
		}
		near(t, m.state.LogEvidence, n.state.LogEvidence)
	}
}

func BenchmarkFullFrontier(b *testing.B) {
	for _, members := range []int{150, 200} {
		b.Run(fmt.Sprintf("M%d/T%d", members, members*16), func(b *testing.B) {
			m := coldFixture(b, members, members*16)
			rows := len(m.rows)
			b.Run("PendingFirst", func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					q, e := m.Pending(m.token, rows-1, 1)
					if e != nil {
						b.Fatal(e)
					}
					benchQuery = q
				}
			})
			b.Run("RevealFirst", func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					x := *m
					if e := x.Reveal(x.token, rows-1, 1, 1, int64(rows+1)); e != nil {
						b.Fatal(e)
					}
					benchLedger = &x
				}
			})
			b.Run("Issue", func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					x := *m
					r, e := x.Issue(0, int64(rows+1))
					if e != nil {
						b.Fatal(e)
					}
					benchReceipt = r
				}
			})
		})
	}
}
