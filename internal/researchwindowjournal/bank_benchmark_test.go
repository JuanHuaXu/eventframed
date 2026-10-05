package researchwindowjournal

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

var bankSink *Bank
var bankForecastSink float64

func benchmarkBase(members int) []float64 {
	out := make([]float64, members)
	for i := range out {
		out[i] = .25 + .675*float64(i)/float64(members-1)
	}
	return out
}
func BenchmarkBankConstructor150(b *testing.B) {
	base := benchmarkBase(150)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var e error
		bankSink, e = NewBank(base, 1, 2800, 7, [3]int{600, 1200, 2400})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkBankConstructor200(b *testing.B) {
	base := benchmarkBase(200)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var e error
		bankSink, e = NewBank(base, 1, 2800, 7, [3]int{600, 1200, 2400})
		if e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkBankPredict150(b *testing.B) {
	m, e := NewBank(benchmarkBase(150), 1, 2800, 7, [3]int{600, 1200, 2400})
	if e != nil {
		b.Fatal(e)
	}
	for n := 0; n < 900; n++ {
		x, e := m.Issue(n%150, int64(n))
		if e != nil {
			b.Fatal(e)
		}
		if _, e = m.Resolve(x, n%3 != 0, int64(n)); e != nil {
			b.Fatal(e)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q, _, e := m.Predict(i % 150)
		if e != nil {
			b.Fatal(e)
		}
		bankForecastSink = q
	}
}
func BenchmarkBankBounded2400Loop(b *testing.B) {
	base := benchmarkBase(150)
	b.ReportAllocs()
	b.ResetTimer()
	for repeat := 0; repeat < b.N; repeat++ {
		m, e := NewBank(base, 1, 2800, 7, [3]int{600, 1200, 2400})
		if e != nil {
			b.Fatal(e)
		}
		for n := 0; n < 2400; n++ {
			x, e := m.Issue(n%150, int64(n))
			if e != nil {
				b.Fatal(e)
			}
			bankForecastSink = x.Forecast()
			if _, e = m.Resolve(x, n%3 != 0, int64(n)); e != nil {
				b.Fatal(e)
			}
			// Exactly25 paired packets/150-frame round. This fixed nomination ablation
			// measures the core loop, NOT adaptive acquisition or a quality experiment.
			if n%6 == 0 {
				second, e := m.RequestSecond(x, int64(n))
				if e != nil {
					b.Fatal(e)
				}
				if _, e = m.Resolve(second, n%5 == 0, int64(n)); e != nil {
					b.Fatal(e)
				}
			}
			if n%150 == 149 {
				for i := 0; i < 150; i++ {
					q, _, e := m.Predict(i)
					if e != nil {
						b.Fatal(e)
					}
					bankForecastSink = q
				}
			}
		}
		if m.Pending() != 0 {
			b.Fatal("undrained loop")
		}
		bankSink = m
	}
}
func TestBankAllocation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BANK_V63_ALLOCATION")
	if path == "" {
		t.Skip("explicit isolated artifact")
	}
	allocation := func(members int, shared bool) uint64 {
		base := benchmarkBase(members)
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		if shared {
			var e error
			bankSink, e = NewBank(base, 1, 2800, 7, [3]int{600, 1200, 2400})
			if e != nil {
				t.Fatal(e)
			}
		} else {
			x := &Bank{}
			for k, w := range [3]int{600, 1200, 2400} {
				var e error
				x.models[k], e = New(base, 1, 2800, Config{Depth: 7, Window: w})
				if e != nil {
					t.Fatal(e)
				}
			}
			var e error
			x.selector, e = newSelector(members, 1)
			if e != nil {
				t.Fatal(e)
			}
			bankSink = x
		}
		runtime.ReadMemStats(&after)
		return after.TotalAlloc - before.TotalAlloc
	}
	values := map[string]uint64{"shared150": allocation(150, true), "unshared150": allocation(150, false), "shared200": allocation(200, true), "unshared200": allocation(200, false)}
	raw, e := json.MarshalIndent(struct {
		AllocatedBytes map[string]uint64
		CapBytes       uint64
		NotRSS         bool
	}{values, 8 * 1024 * 1024, true}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	file, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer file.Close()
	if _, e = file.Write(append(raw, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = file.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log(values)
}
