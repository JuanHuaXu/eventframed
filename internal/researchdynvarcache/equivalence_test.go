package researchdynvarcache

import (
	"fmt"
	"reflect"
	"testing"

	uncached "github.com/JuanHuaXu/eventframed/internal/researchdynvariance"
)

func TestBitwiseUncachedEquivalence(t *testing.T) {
	checks := 0
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, family := range []string{"learn", "baseline", "current", "free"} {
			for _, hazard := range []float64{0, 1. / 16, 1} {
				t.Run(fmt.Sprintf("%s/%s/%g", mode, family, hazard), func(t *testing.T) {
					base := []float64{.25, .47, .7, .925}
					m, e := New(base, 1, 512, Config{mode, family, hazard})
					if e != nil {
						t.Fatal(e)
					}
					r, e := uncached.New(base, 1, 512, uncached.Config{Mode: mode, Family: family, Hazard: hazard})
					if e != nil {
						t.Fatal(e)
					}
					compare := func() {
						w, e := m.ModelWeights()
						if e != nil {
							t.Fatal(e)
						}
						v, e := r.ModelWeights()
						if e != nil {
							t.Fatal(e)
						}
						if w != v {
							t.Fatal("bitwise model weights", w, v)
						}
						checks += 3
						for i := range base {
							q, o, e := m.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							u, v, e := r.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							if q != u || o != v {
								t.Fatalf("bitwise law member%d %.17g/%.17g vs %.17g/%.17g", i, q, o, u, v)
							}
							checks += 2
						}
						if m.Pending() != r.Pending() {
							t.Fatal("pending")
						}
					}
					var a [256]Ticket
					var b [256]uncached.Ticket
					for n := 0; n < 256; n++ {
						compare()
						a[n], e = m.Issue(n%4, int64(n))
						if e != nil {
							t.Fatal(e)
						}
						b[n], e = r.Issue(n%4, int64(n))
						if e != nil {
							t.Fatal(e)
						}
						if a[n].Forecast() != b[n].Forecast() {
							t.Fatal("ticket")
						}
						if n >= 11 {
							j := n - 11
							p, e := m.Resolve(a[j], j%7 < 3, int64(n))
							if e != nil {
								t.Fatal(e)
							}
							q, e := r.Resolve(b[j], j%7 < 3, int64(n))
							if e != nil {
								t.Fatal(e)
							}
							if p.Member != q.Member || p.Ordinal != q.Ordinal || p.Measurement != q.Measurement || p.Forecast != q.Forecast || p.ArrivedAt != q.ArrivedAt {
								t.Fatal("first receipt")
							}
						}
					}
					at := int64(256)
					for j := 245; j < 256; j++ {
						if _, e = m.Resolve(a[j], j%7 < 3, at); e != nil {
							t.Fatal(e)
						}
						if _, e = r.Resolve(b[j], j%7 < 3, at); e != nil {
							t.Fatal(e)
						}
						at++
					}
					for j := 0; j < 8; j++ {
						compare()
						before := snapshot(m)
						for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
							x, e := m.Query(a[j], query)
							if e != nil {
								t.Fatal(e)
							}
							y, e := r.Query(b[j], query)
							if e != nil {
								t.Fatal(e)
							}
							if x.Observed != y.Observed || x.Uncertainty != y.Uncertainty || x.Information != y.Information || x.EdgeCut != y.EdgeCut || x.Value != y.Value || x.ClassGain != y.ClassGain {
								t.Fatal("bitwise query", query, x, y)
							}
							checks += 6
						}
						if !reflect.DeepEqual(before, snapshot(m)) {
							t.Fatal("hypothetical cache publication")
						}
						x, e := m.RequestSecond(a[j], at)
						if e != nil {
							t.Fatal(e)
						}
						y, e := r.RequestSecond(b[j], at)
						if e != nil {
							t.Fatal(e)
						}
						if x.Forecast() != y.Forecast() {
							t.Fatal("pair ticket")
						}
						if j%3 == 1 {
							if e = m.Cancel(x, at); e != nil {
								t.Fatal(e)
							}
							if e = r.Cancel(y, at); e != nil {
								t.Fatal(e)
							}
						} else {
							p, e := m.Resolve(x, !(j%7 < 3), at)
							if e != nil {
								t.Fatal(e)
							}
							q, e := r.Resolve(y, !(j%7 < 3), at)
							if e != nil {
								t.Fatal(e)
							}
							if p.Forecast != q.Forecast || p.Ordinal != q.Ordinal {
								t.Fatal("pair receipt")
							}
						}
						compare()
						at++
					}
					if e = m.BeginEpoch(2, at); e != nil {
						t.Fatal(e)
					}
					if e = r.BeginEpoch(2, at); e != nil {
						t.Fatal(e)
					}
					compare()
				})
			}
		}
	}
	t.Logf("%d counted bitwise scalar comparisons across 36 full-journal configurations", checks)
}
