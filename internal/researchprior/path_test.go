package researchprior

import (
	"math"
	"math/rand"
	"testing"
)

func TestExplicitThreeStatePaths(t *testing.T) {
	for _, c := range []Config{{"raw", 2, 0, false}, {"inverse", 4, 1. / 16, true}, {"raw", 4, .2, true}} {
		m, _ := New([]float64{.32, .85}, 1, 128, c)
		tickets := make([]Ticket, 3)
		for j := range tickets {
			tickets[j], _ = m.Issue(0, int64(j))
		}
		if _, e := m.Resolve(tickets[2], true, 3); e != nil {
			t.Fatal(e)
		}
		if _, e := m.Resolve(tickets[0], false, 4); e != nil {
			t.Fatal(e)
		}
		numerator, denominator := 0., 0.
		mean := .32
		if c.Center == "inverse" {
			mean = (.9*mean - .05) / .8
		}
		for h, p := range []float64{mean, 1 - mean, .5} {
			var pri [21]float64
			sum, average := 0., 0.
			for z := range pri {
				q := (float64(z) + .5) / 21
				pri[z] = math.Exp((c.Strength*p-1)*math.Log(q) + (c.Strength*(1-p)-1)*math.Log1p(-q))
				sum += pri[z]
			}
			for z := range pri {
				pri[z] /= sum
				average += pri[z] * (float64(z) + .5) / 21
			}
			for a := 0; a < 21; a++ {
				for b := 0; b < 21; b++ {
					for d := 0; d < 21; d++ {
						ab := c.Hazard * pri[b]
						if a == b {
							ab += 1 - c.Hazard
						}
						bd := c.Hazard * pri[d]
						if b == d {
							bd += 1 - c.Hazard
						}
						q0, q2 := (float64(a)+.5)/21, (float64(d)+.5)/21
						mass := []float64{.8, .1, .1}[h] * pri[a] * ab * bd * (1 - q0) * q2
						denominator += mass
						numerator += mass * ((1-c.Hazard)*q2 + c.Hazard*average)
					}
				}
			}
		}
		got, e := m.Predict(0)
		if e != nil {
			t.Fatal(e)
		}
		near(t, got, numerator/denominator)
	}
}
func TestInterleavedPrefixAndCensoring(t *testing.T) {
	for _, shared := range []bool{false, true} {
		cfg := Config{"inverse", 2, 1. / 16, shared}
		base := []float64{.29, .59, .91}
		m, _ := New(base, 1, 192, cfg)
		n := make([]int, 3)
		known, yes := make([]uint64, 3), make([]uint64, 3)
		type pending struct {
			ticket Ticket
			i, j   int
		}
		queue := []pending{}
		rng := rand.New(rand.NewSource(670021))
		for at := 0; at < 300; at++ {
			if at < 100 && len(queue) < 10 {
				i := rng.Intn(3)
				ticket, e := m.Issue(i, int64(at))
				if e != nil {
					t.Fatal(e)
				}
				queue = append(queue, pending{ticket, i, n[i]})
				n[i]++
			}
			if len(queue) > 0 && (at >= 100 || rng.Intn(2) == 1) {
				k := rng.Intn(len(queue))
				p := queue[k]
				queue = append(queue[:k], queue[k+1:]...)
				if rng.Intn(5) == 0 {
					if e := m.Cancel(p.ticket, int64(at)); e != nil {
						t.Fatal(e)
					}
				} else {
					y := rng.Intn(2) == 1
					if _, e := m.Resolve(p.ticket, y, int64(at)); e != nil {
						t.Fatal(e)
					}
					known[p.i] |= 1 << p.j
					if y {
						yes[p.i] |= 1 << p.j
					}
				}
			}
			want := reference(base, cfg, n, known, yes)
			for i, v := range want {
				got, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(t, got, v)
			}
		}
		if m.Pending() != 0 {
			t.Fatal("pending drain")
		}
	}
}

// Fully observed63-position suffix, unlike V39's one-label/unit-emission
// benchmark. Restoring the original private checkpoint occurs OUTSIDE timing.
func BenchmarkPriorReplay64FullyObserved(b *testing.B) {
	m, _ := New([]float64{.3, .8}, 1, 128, Config{"raw", 4, 1. / 16, true})
	tickets := make([]Ticket, 64)
	for j := range tickets {
		tickets[j], _ = m.Issue(0, int64(j))
	}
	for j := 1; j < 64; j++ {
		if _, e := m.Resolve(tickets[j], j%2 == 0, 64); e != nil {
			b.Fatal(e)
		}
	}
	rows := append([]row(nil), m.forward...)
	logs := append([][3]float64(nil), m.evidence...)
	weights, global, private := m.weights, m.global, m.private[0]
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		b.StopTimer()
		copy(m.forward, rows)
		copy(m.evidence, logs)
		m.weights, m.global, m.private[0] = weights, global, private
		m.trials[0].status = 1
		m.trials[0].useful = false
		m.pending = 1
		m.clock = 64
		b.StartTimer()
		if _, e := m.Resolve(tickets[0], true, 65); e != nil {
			b.Fatal(e)
		}
	}
}
