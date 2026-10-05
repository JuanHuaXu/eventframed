package researchanchor

import "testing"

func benchModel(b *testing.B) (*Model, []Ticket) {
	b.Helper()
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	m, e := New(base, 1, 2800, Config{Depth: 7, Window: 600})
	if e != nil {
		b.Fatal(e)
	}
	origins := make([]Ticket, 150)
	for k := 0; k < 450; k++ {
		a, e := m.Issue(k%150, int64(k))
		if e != nil {
			b.Fatal(e)
		}
		if _, e = m.Resolve(a, k%5 < 3, int64(k)); e != nil {
			b.Fatal(e)
		}
		if k >= 300 {
			origins[k%150] = a
		}
	}
	return m, origins
}
func BenchmarkPredict(b *testing.B) {
	m, _ := benchModel(b)
	b.ReportAllocs()
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		if _, _, e := m.Predict(k % 150); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkCandidateBatch(b *testing.B) {
	m, ts := benchModel(b)
	b.ReportAllocs()
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		for _, a := range ts {
			if _, e := m.Query(a, "falsification"); e != nil {
				b.Fatal(e)
			}
		}
	}
}
func BenchmarkPredictiveBatch(b *testing.B) {
	m, ts := benchModel(b)
	weights := make([]float64, 150)
	for i := range weights {
		weights[i] = 1. / 150
	}
	b.ReportAllocs()
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		if _, e := m.PredictionValues(ts, weights); e != nil {
			b.Fatal(e)
		}
	}
}
