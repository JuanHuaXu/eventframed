package researchretention

import "testing"

func BenchmarkWeightsAtCap(b *testing.B) {
	s, _ := New(MaxMembers, 1)
	f := joints([3]float64{.2, .5, .8}, [3]float64{.05, .1, .2})
	for n := 0; n < MaxTrials; n++ {
		t, _ := s.Issue(0, int64(2*n), f)
		_, _ = s.Resolve(t, n%2 == 0, int64(2*n+1))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := s.Weights(0); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkOriginSmoothingAtCap(b *testing.B) {
	s, _ := New(MaxMembers, 1)
	f := joints([3]float64{.2, .5, .8}, [3]float64{.05, .1, .2})
	var first Ticket
	for n := 0; n < MaxTrials; n++ {
		t, _ := s.Issue(0, int64(2*n), f)
		if n == 0 {
			first = t
		}
		_, _ = s.Resolve(t, n%2 == 0, int64(2*n+1))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := s.QuerySecond(first); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkReplayOriginalAtCap(b *testing.B) {
	s, _ := New(MaxMembers, 1)
	f := joints([3]float64{.2, .5, .8}, [3]float64{.05, .1, .2})
	var first Ticket
	for n := 0; n < MaxTrials; n++ {
		t, _ := s.Issue(0, int64(2*n), f)
		if n == 0 {
			first = t
		}
		_, _ = s.Resolve(t, n%2 == 0, int64(2*n+1))
	}
	row := s.members[0].rows[first.ordinal]
	row.second = 2
	row.w2 = false
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := s.prepare(0, 0, row); e != nil {
			b.Fatal(e)
		}
	}
}

var allocatedSelector *Selector

func BenchmarkConstructorAtCap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var e error
		allocatedSelector, e = New(MaxMembers, 1)
		if e != nil {
			b.Fatal(e)
		}
	}
}
