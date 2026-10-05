package researchclasssequence

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

var modelSink *Model
var numberSink float64

func baseline(n int) []float64 {
	b := make([]float64, n)
	for i := range b {
		b[i] = .25 + .675*float64(i)/float64(n-1)
	}
	return b
}
func TestConstructorAllocation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_CLASS_V69_ALLOCATION")
	if path == "" {
		t.Skip("explicit isolated artifact")
	}
	values := map[int]uint64{}
	for _, n := range []int{150, 200} {
		base := baseline(n)
		runtime.GC()
		var a, b runtime.MemStats
		runtime.ReadMemStats(&a)
		var e error
		modelSink, e = New(base, 1, 2800, Config{Mode: "hybrid", Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		runtime.ReadMemStats(&b)
		values[n] = b.TotalAlloc - a.TotalAlloc
	}
	data, e := json.MarshalIndent(struct {
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
	if _, e = f.Write(append(data, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log(values)
}
func BenchmarkConstructor150(b *testing.B) {
	base := baseline(150)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var e error
		modelSink, e = New(base, 1, 2800, Config{Mode: "hybrid", Hazard: 1. / 16})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkConstructor200(b *testing.B) {
	base := baseline(200)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var e error
		modelSink, e = New(base, 1, 2800, Config{Mode: "hybrid", Hazard: 1. / 16})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func populated(b *testing.B) (*Model, Ticket) {
	m, e := New(baseline(150), 1, 2800, Config{Mode: "hybrid", Hazard: 1. / 16})
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
		if n == 750 {
			target = x
		}
	}
	return m, target
}
func BenchmarkPredict150(b *testing.B) {
	m, _ := populated(b)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, _, e := m.Predict(n % 150)
		if e != nil {
			b.Fatal(e)
		}
		numberSink = q
	}
}
func BenchmarkPredictiveQuery150(b *testing.B) {
	m, t := populated(b)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, e := m.Query(t, "predictive")
		if e != nil {
			b.Fatal(e)
		}
		numberSink = q.Value
	}
}

func BenchmarkModelClassQuery150(b *testing.B) {
	m, t := populated(b)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		x, e := m.Query(t, "model_class")
		if e != nil {
			b.Fatal(e)
		}
		numberSink = x.ClassGain
	}
}
func BenchmarkNoiseClassQuery150(b *testing.B) {
	m, t := populated(b)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		x, e := m.Query(t, "noise_class")
		if e != nil {
			b.Fatal(e)
		}
		numberSink = x.ClassGain
	}
}
