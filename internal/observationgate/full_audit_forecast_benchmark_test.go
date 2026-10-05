package observationgate

import "testing"

var fullAuditForecastSink float64

// Measures only the two extra probability lookups on already-acquired bits.
// No retrieval, fitting, gate update, queue or persistence cost is included.
func BenchmarkFullAuditForecastPair(b *testing.B) {
	base, _, err := arrivalSwitchSetup(1, 3, 0)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sum float64
	for i := 0; i < b.N; i++ {
		p, err := base.ForecastObserved(511, uint16(i&511))
		if err != nil {
			b.Fatal(err)
		}
		q, err := base.ForecastObserved(511, uint16((i+137)&511))
		if err != nil {
			b.Fatal(err)
		}
		sum += p + q
	}
	fullAuditForecastSink = sum
}
