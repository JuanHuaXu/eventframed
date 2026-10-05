package researchmeanratio

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

var keepModel *Model
var keepValue float64

func benchmarkModel(t testing.TB, size int) (*Model, Ticket) {
	t.Helper()
	base := make([]float64, size)
	for i := range base {
		base[i] = .25 + .675*float64(i)/float64(size-1)
	}
	m, e := New(base, 1, 2*size*MaxTrials, Config{"noise", "learn", "learn", 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	var first Ticket
	for n := 0; n < 16; n++ {
		for i := range base {
			x, e := m.Issue(i, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if n == 0 && i == 0 {
				first = x
			}
			if _, e = m.Resolve(x, (i+n)%3 != 0, int64(n)); e != nil {
				t.Fatal(e)
			}
		}
	}
	return m, first
}
func TestConstructorAllocation(t *testing.T) {
	report := struct {
		AllocatedBytes map[int]uint64
		CapBytes       uint64
	}{map[int]uint64{}, 8 * 1024 * 1024}
	for _, size := range []int{150, 200} {
		base := make([]float64, size)
		for i := range base {
			base[i] = .5
		}
		runtime.GC()
		var a, b runtime.MemStats
		runtime.ReadMemStats(&a)
		m, e := New(base, 1, 2*size*MaxTrials, Config{"noise", "learn", "learn", 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		runtime.ReadMemStats(&b)
		runtime.KeepAlive(m)
		allocated := b.TotalAlloc - a.TotalAlloc
		report.AllocatedBytes[size] = allocated
		t.Logf("members=%d constructor allocated bytes=%d; not RSS or loaded serving", size, allocated)
		if allocated > 8*1024*1024 {
			t.Fatal("constructor allocation cap", size, allocated)
		}
	}
	if path := os.Getenv("EVENTFRAME_MEANRATIO_V76_ALLOCATION"); path != "" {
		f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.NewEncoder(f).Encode(report); e != nil {
			f.Close()
			t.Fatal(e)
		}
		if e = f.Sync(); e != nil {
			f.Close()
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
}
func BenchmarkPredict(b *testing.B) {
	m, _ := benchmarkModel(b, 150)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, _, e := m.Predict(n % 150)
		if e != nil {
			b.Fatal(e)
		}
		keepValue = q
	}
}
func BenchmarkQuery(b *testing.B) {
	for _, mode := range []string{"model_class", "noise_class", "predictive"} {
		b.Run(mode, func(b *testing.B) {
			m, x := benchmarkModel(b, 150)
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				o, e := m.Query(x, mode)
				if e != nil {
					b.Fatal(e)
				}
				keepValue = o.ClassGain + o.Value
			}
		})
	}
}

// Most live nominations target the latest revealed row. The old-row benchmark
// remains above so the fallback is not hidden by this favorable lifecycle phase.
func BenchmarkAnchoredQuery(b *testing.B) {
	for _, mode := range []string{"model_class", "noise_class", "predictive"} {
		b.Run(mode, func(b *testing.B) {
			m, _ := benchmarkModel(b, 150)
			x := Ticket{owner: m, epoch: m.epoch, slot: m.members[0].slots[15]}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				o, e := m.Query(x, mode)
				if e != nil {
					b.Fatal(e)
				}
				keepValue = o.ClassGain + o.Value
			}
		})
	}
}
