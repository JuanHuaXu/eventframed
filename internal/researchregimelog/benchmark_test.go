package researchregimelog

import (
	"fmt"
	"testing"
)

var benchLedger *Ledger
var benchQuery Query
var benchReceipt Receipt

func fixture(b *testing.B, members, rows int) *Ledger {
	b.Helper()
	base := make([]float64, members)
	for i := range base {
		base[i] = .25 + .675*float64(i)/float64(max(1, members-1))
	}
	m, e := New(base, Config{1. / 16, .25, 36})
	if e != nil {
		b.Fatal(e)
	}
	for i := 0; i < rows; i++ {
		if _, e = m.Issue(i%members, int64(i)); e != nil {
			b.Fatal(e)
		}
	}
	for _, i := range []int{0, rows / 2, rows - 1} {
		if e = m.Reveal(m.token, i, 1, i%2, int64(rows)); e != nil {
			b.Fatal(e)
		}
	}
	return m
}

func BenchmarkLedgerOperations(b *testing.B) {
	for _, members := range []int{2, 150, 200} {
		for _, rows := range []int{32, 64} {
			b.Run(fmt.Sprintf("M%d/T%d", members, rows), func(b *testing.B) {
				m := fixture(b, members, rows)
				name := "PendingSecond"
				impossible := latestPairImpossible(m)
				if impossible {
					name = "PendingSecondAbstention"
				}
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for n := 0; n < b.N; n++ {
						q, e := m.Pending(m.token, rows-1, 2)
						if impossible {
							if e == nil {
								b.Fatal("unsupported query accepted")
							}
							continue
						}
						if e != nil {
							b.Fatal(e)
						}
						benchQuery = q
					}
				})
				b.Run("PendingFirst", func(b *testing.B) {
					b.ReportAllocs()
					for n := 0; n < b.N; n++ {
						q, e := m.Pending(m.token, rows-2, 1)
						if e != nil {
							b.Fatal(e)
						}
						benchQuery = q
					}
				})
				b.Run("Reveal", func(b *testing.B) {
					b.ReportAllocs()
					for n := 0; n < b.N; n++ {
						x := *m
						if e := x.Reveal(x.token, rows-1, 2, (rows-1)%2, int64(rows+1)); e != nil {
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
}

func BenchmarkConstructor(b *testing.B) {
	for _, members := range []int{150, 200} {
		b.Run(fmt.Sprintf("M%d", members), func(b *testing.B) {
			base := make([]float64, members)
			for i := range base {
				base[i] = .5
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m, e := New(base, Config{1. / 16, .25, 36})
				if e != nil {
					b.Fatal(e)
				}
				benchLedger = m
			}
		})
	}
}
