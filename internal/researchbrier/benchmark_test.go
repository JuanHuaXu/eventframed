package researchbrier

import "testing"

var benchmarkForecast float64

func BenchmarkSubstitution3(b *testing.B) {
	benchmarkSubstitution(b, 3)
}

func BenchmarkSubstitution8(b *testing.B) {
	benchmarkSubstitution(b, 8)
}

func benchmarkSubstitution(b *testing.B, n int) {
	var logs, advice [MaxExperts]float64
	for k := 0; k < n; k++ {
		logs[k] = -float64(k)
		advice[k] = float64(k+1) / float64(n+1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	q := 0.
	for j := 0; j < b.N; j++ {
		var err error
		q, err = Forecast(logs[:n], advice[:n])
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if q < 0 || q > 1 {
		b.Fatal("invalid forecast")
	}
	benchmarkForecast = q
}

// Includes issue, original-advice ownership, substitution, resolved receipt,
// normalization and a prior reset every 4,096 admitted observations. This is
// an immediate isolated component, not durable/loaded daemon serving.
func BenchmarkImmediateCycle3(b *testing.B) {
	m, err := New([]float64{.8, .1, .1}, 1, MaxTrials)
	if err != nil {
		b.Fatal(err)
	}
	advice := []float64{.2, .5, .8}
	epoch := uint64(1)
	b.ReportAllocs()
	b.ResetTimer()
	q := 0.
	for j := 0; j < b.N; j++ {
		at := int64(j) * 2
		if j > 0 && j%MaxTrials == 0 {
			epoch++
			if err := m.BeginEpoch(epoch, at); err != nil {
				b.Fatal(err)
			}
		}
		x, err := m.Issue(advice, at)
		if err != nil {
			b.Fatal(err)
		}
		r, err := m.Resolve(x, j%3 != 0, at+1)
		if err != nil {
			b.Fatal(err)
		}
		q = r.Forecast
	}
	b.StopTimer()
	benchmarkForecast = q
}
