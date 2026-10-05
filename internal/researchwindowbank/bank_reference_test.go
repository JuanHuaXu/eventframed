package researchwindowbank_test

import (
	"fmt"
	frozen "github.com/JuanHuaXu/eventframed/internal/researchretention"
	law "github.com/JuanHuaXu/eventframed/internal/researchretentionlaw"
	bank "github.com/JuanHuaXu/eventframed/internal/researchwindowbank"
	reference "github.com/JuanHuaXu/eventframed/internal/researchwindowbankref"
	"sort"
	"testing"
)

type bankEvent struct {
	at                 int64
	trial, measurement int
	value, available   bool
}
type origin struct {
	candidate                 bank.BankTicket
	selector                  frozen.Ticket
	member, ordinal, position int
	audited                   [3]bool
}

func TestIndependentBankDelayedIssueJointAndExpiry(t *testing.T) {
	checks := 0
	for _, depth := range []int{0, 2, 7} {
		for _, delay := range []int64{0, 3, 11} {
			base := []float64{.25, .3, .47, .5, .6, .78, .8, .9, .925}
			windows := [3]int{2, 9, 32}
			candidate, e := bank.NewBank(base, 1, 256, depth, windows)
			if e != nil {
				t.Fatal(e)
			}
			var refs [3]*reference.Reference
			for k, w := range windows {
				refs[k], e = reference.New(base, bank.Config{Depth: depth, Window: w})
				if e != nil {
					t.Fatal(e)
				}
			}
			selector, e := frozen.New(len(base), 1)
			if e != nil {
				t.Fatal(e)
			}
			origins := []origin{}
			queue := []bankEvent{}
			issued := make([]int, len(base))
			verify := func() {
				for i := range base {
					var js [3]law.Joint
					for k, r := range refs {
						j, e := r.LatentJoint(i)
						if e != nil {
							t.Fatal(e)
						}
						js[k] = j
					}
					fs, e := law.Forecasts(js)
					if e != nil {
						t.Fatal(e)
					}
					ws, e := selector.Weights(i)
					if e != nil {
						t.Fatal(e)
					}
					q, o := 0., 0.
					for k, w := range ws {
						q += w * fs[k].Clean
						o += w * (fs[k].Joint[2] + fs[k].Joint[3])
					}
					actual, observed, e := candidate.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					close(t, q, actual)
					close(t, o, observed)
					checks += 2
				}
			}
			var deliver func(int64)
			deliver = func(at int64) {
				for {
					sort.SliceStable(queue, func(i, j int) bool {
						if queue[i].at != queue[j].at {
							return queue[i].at < queue[j].at
						}
						if queue[i].trial != queue[j].trial {
							return queue[i].trial < queue[j].trial
						}
						return queue[i].measurement < queue[j].measurement
					})
					if len(queue) == 0 || queue[0].at > at {
						return
					}
					event := queue[0]
					queue = queue[1:]
					x := &origins[event.trial]
					if !event.available {
						if e := candidate.Cancel(x.candidate, event.at); e != nil {
							t.Fatal(e)
						}
						if e := selector.Cancel(x.selector, event.at); e != nil {
							t.Fatal(e)
						}
						verify()
						continue
					}
					actual, e := candidate.Resolve(x.candidate, event.value, event.at)
					if e != nil {
						t.Fatal(e)
					}
					want, e := selector.Resolve(x.selector, event.value, event.at)
					if e != nil {
						t.Fatal(e)
					}
					close(t, actual.Forecast, want.Forecast)
					if actual.Ordinal != want.Ordinal || actual.Measurement != want.Measurement || actual.IssuedAt != want.IssuedAt || actual.ArrivedAt != want.ArrivedAt {
						t.Fatal("receipt contract mismatch")
					}
					for k, r := range refs {
						if event.measurement == 1 || x.audited[k] {
							if e := r.Observe(x.member, x.ordinal, event.measurement, event.value); e != nil {
								t.Fatal(e)
							}
						}
					}
					verify()
					if event.measurement == 1 && event.trial%5 == 0 {
						a, e := candidate.QuerySecond(x.candidate)
						if e != nil {
							t.Fatal(e)
						}
						z, e := selector.QuerySecond(x.selector)
						if e != nil {
							t.Fatal(e)
						}
						close(t, a, z)
						second, e := candidate.RequestSecond(x.candidate, event.at)
						if e != nil {
							t.Fatal(e)
						}
						ss, e := selector.RequestSecond(x.selector, event.at)
						if e != nil {
							t.Fatal(e)
						}
						close(t, second.Forecast(), ss.Forecast())
						x.candidate, x.selector = second, ss
						for k, w := range windows {
							if len(origins)-x.position <= w {
								if e := refs[k].Audit(x.member, x.ordinal); e != nil {
									t.Fatal(e)
								}
								x.audited[k] = true
							}
						}
						queue = append(queue, bankEvent{event.at + 13, event.trial, 2, event.trial%4 == 0, event.trial%15 != 0})
					}
				}
			}
			for n := 0; n < 72; n++ {
				at := int64(10 * n)
				deliver(at)
				i := n % len(base)
				var js [3]law.Joint
				for k, r := range refs {
					j, e := r.LatentJoint(i)
					if e != nil {
						t.Fatal(e)
					}
					js[k] = j
				}
				f, e := law.Forecasts(js)
				if e != nil {
					t.Fatal(e)
				}
				actual, e := candidate.Issue(i, at)
				if e != nil {
					t.Fatal(e)
				}
				ss, e := selector.Issue(i, at, f)
				if e != nil {
					t.Fatal(e)
				}
				close(t, actual.Forecast(), ss.Forecast())
				for _, r := range refs {
					if e := r.Issue(i); e != nil {
						t.Fatal(e)
					}
				}
				issued[i]++
				origins = append(origins, origin{candidate: actual, selector: ss, member: i, ordinal: issued[i], position: n})
				queue = append(queue, bankEvent{at + delay, n, 1, n%3 != 0, n%17 != 0})
				deliver(at)
				verify()
			}
			deliver(10000)
			if candidate.Pending() != 0 {
				t.Fatal("pending not drained")
			}
			for _, r := range refs {
				if e := r.RebuildAll(); e != nil {
					t.Fatal(e)
				}
			}
			verify()
		}
	}
	t.Log("independent bank forecast comparisons", checks)
}
func TestIntegratedFutureForkVisibleBoundary(t *testing.T) {
	run := func(change bool) []float64 {
		b, e := bank.NewBank([]float64{.3, .7}, 1, 128, 1, [3]int{2, 6, 12})
		if e != nil {
			t.Fatal(e)
		}
		out := []float64{}
		var pending bank.BankTicket
		for n := 0; n < 24; n++ {
			if n > 0 {
				value := (n-1)%3 == 0
				if change && n-1 >= 12 {
					value = !value
				}
				if _, e = b.Resolve(pending, value, int64(n)); e != nil {
					t.Fatal(e)
				}
			}
			pending, e = b.Issue(n%2, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			out = append(out, pending.Forecast())
		}
		return out
	}
	a, z := run(false), run(true)
	changed := false
	for n := range a {
		if n <= 12 && a[n] != z[n] {
			t.Fatal("unreceived future affected issue", n)
		}
		if n > 12 && a[n] != z[n] {
			changed = true
		}
	}
	if !changed {
		t.Fatal(fmt.Sprint("vacuous integrated fork"))
	}
}
