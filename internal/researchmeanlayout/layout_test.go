package researchmeanlayout

import (
	"encoding/json"
	"math"
	"os"
	"runtime"
	"testing"
)

func close(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsInf(a, 0) || math.Abs(a-b) > 2e-11 {
		t.Fatalf("%.17g vs %.17g", a, b)
	}
}

// Independent urn dynamic programming, not adjacent-ratio construction.
func dp(p, strength float64) (out [21]float64) {
	out[0] = 1
	for n := 0; n < 20; n++ {
		var next [21]float64
		for z := 0; z <= n; z++ {
			next[z+1] += out[z] * (strength*p + float64(z)) / (strength + float64(n))
			next[z] += out[z] * (strength*(1-p) + float64(n-z)) / (strength + float64(n))
		}
		out = next
	}
	return
}

func TestFullGridPriorLayout(t *testing.T) {
	checks := 0
	for _, size := range []int{2, 4, 150, 200} {
		base := make([]float64, size)
		for i := range base {
			base[i] = .25 + .675*float64(i)/float64(size-1)
		}
		m, e := New(base)
		if e != nil {
			t.Fatal(e)
		}
		if len(m.rows) != size*64 || len(m.members) != size {
			t.Fatal("full caps")
		}
		for i, x := range m.members {
			q := 0.
			for t0, p := range x.mean {
				q += meanPrior[t0] * p
				for a := 1; a <= 2; a++ {
					want := dp(p, float64(a))
					for h := 0; h < 3; h++ {
						mass, mean := 0., 0.
						if a == 1 {
							mass, mean = x.current[t0][h][0], x.current[t0][h][0]*p
						}
						for z := range want {
							v := x.free[t0][h][z]
							w := want[z]
							if a == 1 {
								v, w = x.current[t0][h][z+1], .2*w
							}
							close(t, v, w)
							checks++
							mass += v
							mean += v * float64(z) / 20
						}
						close(t, mass, 1)
						close(t, mean, p)
						checks += 2
					}
				}
			}
			close(t, q, base[i])
			checks++
		}
		prior := 0.
		for _, p := range m.joint {
			prior += p
		}
		close(t, prior, 1)
		checks++
	}
	t.Logf("%d counted prior/layout scalar checks; no posterior learning tested", checks)
}

func TestCompletePlannedLayoutAllocation(t *testing.T) {
	report := struct {
		AllocatedBytes map[int]uint64
		CapBytes       uint64
		Scope          string
	}{map[int]uint64{}, 8 * 1024 * 1024, "initialization with reserved full journal and caches; NOT fitted-learner or replay allocation"}
	for _, size := range []int{150, 200} {
		base := make([]float64, size)
		for i := range base {
			base[i] = .5
		}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		m, e := New(base)
		if e != nil {
			t.Fatal(e)
		}
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(m)
		report.AllocatedBytes[size] = after.TotalAlloc - before.TotalAlloc
		t.Logf("members %d: %d bytes, initialization only", size, report.AllocatedBytes[size])
		if report.AllocatedBytes[size] > report.CapBytes {
			t.Fatal("allocation cap", size)
		}
	}
	if p := os.Getenv("EVENTFRAME_MEAN_LAYOUT_V73_ALLOCATION"); p != "" {
		f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.NewEncoder(f).Encode(report); e != nil {
			f.Close()
			t.Fatal(e)
		}
		if e = f.Sync(); e != nil {
			f.Close()
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
}

func TestRejectInvalidLayout(t *testing.T) {
	for _, b := range [][]float64{nil, {.5}, make([]float64, 201), {.5, math.NaN()}, {.5, math.Inf(1)}, {.2, .5}, {.5, .94}} {
		if _, e := New(b); e == nil {
			t.Fatal("invalid input accepted")
		}
	}
}
