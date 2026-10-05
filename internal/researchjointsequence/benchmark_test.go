package researchjointsequence

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

var sink *Model
var forecastSink float64

func baseVector(n int) []float64 {
	x := make([]float64, n)
	for i := range x {
		x[i] = .25 + .675*float64(i)/float64(n-1)
	}
	return x
}
func TestConstructorAllocation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_JOINT_V66_ALLOCATION")
	if path == "" {
		t.Skip("explicit isolated artifact")
	}
	values := map[int]uint64{}
	for _, n := range []int{150, 200} {
		base := baseVector(n)
		runtime.GC()
		var a, b runtime.MemStats
		runtime.ReadMemStats(&a)
		var e error
		sink, e = New(base, 1, 2800, Config{Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		runtime.ReadMemStats(&b)
		values[n] = b.TotalAlloc - a.TotalAlloc
	}
	b, e := json.MarshalIndent(struct {
		AllocatedBytes map[int]uint64
		CapBytes       uint64
		NotRSS         bool
	}{values, 8 * 1024 * 1024, true}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log(values)
}
func BenchmarkConstructor150(b *testing.B) {
	base := baseVector(150)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var e error
		sink, e = New(base, 1, 2800, Config{Hazard: 1. / 16})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkConstructor200(b *testing.B) {
	base := baseVector(200)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var e error
		sink, e = New(base, 1, 2800, Config{Hazard: 1. / 16})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkPredict150(b *testing.B) {
	m, e := New(baseVector(150), 1, 2800, Config{Hazard: 1. / 16})
	if e != nil {
		b.Fatal(e)
	}
	for n := 0; n < 900; n++ {
		x, e := m.Issue(n%150, int64(n))
		if e != nil {
			b.Fatal(e)
		}
		if _, e = m.Resolve(x, n%3 == 0, int64(n)); e != nil {
			b.Fatal(e)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, _, e := m.Predict(n % 150)
		if e != nil {
			b.Fatal(e)
		}
		forecastSink = q
	}
}
func BenchmarkPredictiveQuery150(b *testing.B) {
	m, e := New(baseVector(150), 1, 2800, Config{Hazard: 1. / 16})
	if e != nil {
		b.Fatal(e)
	}
	var target Ticket
	for n := 0; n < 900; n++ {
		x, e := m.Issue(n%150, int64(n))
		if e != nil {
			b.Fatal(e)
		}
		if _, e = m.Resolve(x, n%3 == 0, int64(n)); e != nil {
			b.Fatal(e)
		}
		if n == 450 {
			target = x
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, e := m.Query(target, "predictive")
		if e != nil {
			b.Fatal(e)
		}
		forecastSink = q.Value
	}
}
