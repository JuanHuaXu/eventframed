package researchwindowjournal_test

import (
	old "github.com/JuanHuaXu/eventframed/internal/researchwindowbank"
	new "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"
	"testing"
)

// Compare public laws, pending counts, issued/requested and received metadata.
// Raw child metadata intentionally changes to the bank law; it is not public.
func TestV62CompleteExternalEquivalence(t *testing.T) {
	checks := 0
	for _, depth := range []int{0, 2, 7} {
		for _, delay := range []int{0, 3, 11} {
			base := []float64{.25, .3, .47, .5, .6, .78, .8, .9, .925}
			a, e := old.NewBank(base, 1, 256, depth, [3]int{2, 9, 32})
			if e != nil {
				t.Fatal(e)
			}
			b, e := new.NewBank(base, 1, 256, depth, [3]int{2, 9, 32})
			if e != nil {
				t.Fatal(e)
			}
			type pair struct {
				a old.BankTicket
				b new.BankTicket
			}
			type scheduled struct {
				at, n                    int
				second, available, value bool
			}
			origins := []pair{}
			queue := []scheduled{}
			verify := func() {
				if a.Pending() != b.Pending() {
					t.Fatal("pending mismatch")
				}
				for i := range base {
					q, o, e := a.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					r, p, e := b.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					if q != r || o != p {
						t.Fatal("bitwise law mismatch", depth, delay, i, q, r, o, p)
					}
					checks += 2
				}
			}
			deliver := func(at int) {
				for j := 0; j < len(queue); {
					event := queue[j]
					if event.at > at {
						j++
						continue
					}
					queue = append(queue[:j], queue[j+1:]...)
					x := origins[event.n]
					if !event.available {
						if e := a.Cancel(x.a, int64(at)); e != nil {
							t.Fatal(e)
						}
						if e := b.Cancel(x.b, int64(at)); e != nil {
							t.Fatal(e)
						}
					} else {
						r, e := a.Resolve(x.a, event.value, int64(at))
						if e != nil {
							t.Fatal(e)
						}
						s, e := b.Resolve(x.b, event.value, int64(at))
						if e != nil {
							t.Fatal(e)
						}
						// Named types differ; compare each exported field directly.
						if r.Member != s.Member || r.Ordinal != s.Ordinal || r.Measurement != s.Measurement || r.Epoch != s.Epoch || r.IssuedAt != s.IssuedAt || r.ArrivedAt != s.ArrivedAt || r.Forecast != s.Forecast || r.Value != s.Value {
							t.Fatal("receipt mismatch")
						}
						checks++
						if !event.second && event.n%5 == 0 {
							q, e := a.QuerySecond(x.a)
							if e != nil {
								t.Fatal(e)
							}
							r, e := b.QuerySecond(x.b)
							if e != nil {
								t.Fatal(e)
							}
							if q != r {
								t.Fatal("query mismatch")
							}
							u, e := a.RequestSecond(x.a, int64(at))
							if e != nil {
								t.Fatal(e)
							}
							v, e := b.RequestSecond(x.b, int64(at))
							if e != nil {
								t.Fatal(e)
							}
							if u.Forecast() != v.Forecast() {
								t.Fatal("request mismatch")
							}
							origins[event.n] = pair{u, v}
							queue = append(queue, scheduled{at + 13, event.n, true, event.n%15 != 0, event.n%4 == 0})
						}
					}
					verify()
				}
			}
			for n := 0; n < 72; n++ {
				at := n * 10
				deliver(at)
				x, e := a.Issue(n%len(base), int64(at))
				if e != nil {
					t.Fatal(e)
				}
				y, e := b.Issue(n%len(base), int64(at))
				if e != nil {
					t.Fatal(e)
				}
				if x.Forecast() != y.Forecast() {
					t.Fatal("issued law mismatch")
				}
				origins = append(origins, pair{x, y})
				queue = append(queue, scheduled{at + delay, n, false, n%17 != 0, n%3 != 0})
				deliver(at)
				verify()
			}
			deliver(10000)
			if e := a.BeginEpoch(2, 10001); e != nil {
				t.Fatal(e)
			}
			if e := b.BeginEpoch(2, 10001); e != nil {
				t.Fatal(e)
			}
			verify()
			if _, e := a.Resolve(origins[0].a, true, 10002); e == nil {
				t.Fatal("old epoch V62")
			}
			if _, e := b.Resolve(origins[0].b, true, 10002); e == nil {
				t.Fatal("old epoch V63")
			}
		}
	}
	t.Log("bitwise V62 public comparisons", checks)
}
