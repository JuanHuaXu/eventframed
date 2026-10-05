package observationlearners

import "testing"

// Entire four-arm learned fixture, including fits, publication and flush.
// Timing uses consumed compatibility seeds and is not extra quality evidence.
func BenchmarkArrivalV106Fixture(b *testing.B) {
	for schedule, name := range []string{"immediate", "delay_missing"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, err := arrivalV106Run(0, 3, 0, schedule, 2100110400)
				if err != nil || r.Final[2].Pending != 0 || r.FinalSelector[2] != uint64(r.Arrived) {
					b.Fatal(err)
				}
			}
		})
	}
}
