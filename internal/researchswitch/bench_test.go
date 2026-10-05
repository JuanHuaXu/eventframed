package researchswitch

import "testing"

var benchmarkForecast float64

func BenchmarkSwitchIssue8(b *testing.B) {
	cfg := Config{Prior: []float64{.125, .125, .125, .125, .125, .125, .125, .125}, Hazard: 1. / 150, Trials: MaxTrials, Pending: MaxTrials}
	m, err := New(cfg, 1)
	if err != nil {
		b.Fatal(err)
	}
	q := []float64{.1, .2, .3, .4, .5, .6, .7, .8}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		if m.used == MaxTrials {
			b.StopTimer()
			m, err = New(cfg, m.epoch+1)
			if err != nil {
				b.Fatal(err)
			}
			b.StartTimer()
		}
		ticket, e := m.Issue(q, int64(m.used))
		if e != nil {
			b.Fatal(e)
		}
		benchmarkForecast = ticket.Forecast()
	}
}

// Restore only the original slot's pending status OUTSIDE timing. Replay from
// slot0 reconstructs every forward message from the prior and retained advice.
// This is a microbenchmark restore, not a supported production mutation API.
func BenchmarkSwitchOldestReplay8(b *testing.B) {
	cfg := Config{Prior: []float64{.125, .125, .125, .125, .125, .125, .125, .125}, Hazard: 1. / 150, Trials: MaxTrials, Pending: MaxTrials}
	m, err := New(cfg, 1)
	if err != nil {
		b.Fatal(err)
	}
	q := []float64{.1, .2, .3, .4, .5, .6, .7, .8}
	var first Ticket
	for j := 0; j < MaxTrials; j++ {
		ticket, e := m.Issue(q, int64(j))
		if e != nil {
			b.Fatal(e)
		}
		if j == 0 {
			first = ticket
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		r, e := m.Resolve(first, true, MaxTrials)
		if e != nil {
			b.Fatal(e)
		}
		benchmarkForecast = r.Forecast
		b.StopTimer()
		m.rows[0].status = 1
		m.known = m.known[:0]
		m.pending++
		b.StartTimer()
	}
}

func benchmarkBase150() []float64 {
	x := make([]float64, 150)
	for i := range x {
		x[i] = .25 + .675*float64(i)/149
	}
	return x
}
func BenchmarkPoolConstructor150(b *testing.B) {
	base := benchmarkBase150()
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 150, Trials: 2400, Pending: 2400}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		p, e := NewPool(base, cfg, 1)
		if e != nil {
			b.Fatal(e)
		}
		benchmarkForecast = p.mix.prior[0]
	}
}

func BenchmarkPoolLoop150x16(b *testing.B) {
	for _, delay := range []int{0, 150, 299} {
		b.Run(map[int]string{0: "immediate", 150: "fixed150", 299: "fixed299"}[delay], func(b *testing.B) {
			base := benchmarkBase150()
			cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 150, Trials: 2400, Pending: 2400}
			b.ReportAllocs()
			b.ResetTimer()
			for repeat := 0; repeat < b.N; repeat++ {
				p, e := NewPool(base, cfg, 1)
				if e != nil {
					b.Fatal(e)
				}
				tickets := make([]PoolTicket, 2400)
				for tick := 0; tick < 2400+delay; tick++ {
					if tick < 2400 {
						ticket, e := p.Issue(tick%150, int64(tick))
						if e != nil {
							b.Fatal(e)
						}
						tickets[tick] = ticket
					}
					if trial := tick - delay; trial >= 0 {
						y := (trial%150 < 75) != (trial >= 1200)
						r, e := p.Resolve(tickets[trial], y, int64(tick))
						if e != nil || r.Forecast != tickets[trial].Forecast() {
							b.Fatal("original score lost", e)
						}
					}
					if (tick < 2400 && tick%150 == 149) || tick == 2399+delay {
						for i := 0; i < 150; i++ {
							q, e := p.Predict(i)
							if e != nil {
								b.Fatal(e)
							}
							benchmarkForecast = q
						}
					}
				}
				if p.Pending() != 0 {
					b.Fatal("undrained benchmark")
				}
			}
		})
	}
}
