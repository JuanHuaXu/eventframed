package libravdbstore

import "testing"

var packetV23BenchmarkSink float64

func BenchmarkResearchPacketInnovationV23(b *testing.B) {
	b.ReportAllocs()
	var sum float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < 32; j++ {
			alpha, beta := 1.0, 2.0
			if j%2 == 0 {
				alpha, beta = 2, 1
			}
			flat, contextual := packetV23Innovation(.925, alpha, beta)
			sum += flat + contextual
		}
	}
	packetV23BenchmarkSink = sum
}
