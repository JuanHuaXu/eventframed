package researchdispersion

import "testing"

var benchmarkModelV34 *Model

func BenchmarkConstruct150(b *testing.B) {
	base := bases(150)
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		var err error
		benchmarkModelV34, err = New(base, "adaptive")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepeatedUpdate16(b *testing.B) {
	m, _ := New(bases(150), "adaptive")
	for round := 1; round <= 15; round++ {
		for i := 0; i < 150; i++ {
			if err := m.Observe(i, round, (i+round)%3 != 0); err != nil {
				b.Fatal(err)
			}
		}
	}
	original := *m
	n, s := m.n[0], m.success[0]
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		m.logs, m.w, m.n[0], m.success[0] = original.logs, original.w, n, s
		if err := m.Observe(0, 16, j%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
}
