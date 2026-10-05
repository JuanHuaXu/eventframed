package researchregimefrozen

import (
	original "github.com/JuanHuaXu/eventframed/internal/researchregime"
	"runtime"
	"testing"
)

var sink Result
var originalSink original.Result

func BenchmarkMatchedReplay(b *testing.B) {
	for _, members := range []int{2, 150, 200} {
		base := make([]float64, members)
		for i := range base {
			base[i] = .25 + .675*float64(i)/float64(members-1)
		}
		rows := make([]Row, 32)
		oldRows := make([]original.Row, 32)
		for i := range rows {
			rows[i] = Row{i % members, i % 2, -1}
			if i%4 == 0 {
				rows[i].Second = 1 - rows[i].First
			}
			oldRows[i] = original.Row{Member: rows[i].Member, First: rows[i].First, Second: rows[i].Second}
		}
		b.Run(stringName(members)+"/v77", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, e := original.Run(base, original.Config{Reset: 1. / 16, Hazard: .25, Cap: 36}, oldRows)
				if e != nil {
					b.Fatal(e)
				}
				originalSink = r
			}
		})
		b.Run(stringName(members)+"/v78", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, e := Run(base, Config{1. / 16, .25, 36}, rows)
				if e != nil {
					b.Fatal(e)
				}
				sink = r
			}
		})
	}
}

func BenchmarkRegimeReplay(b *testing.B) {
	for _, members := range []int{2, 150, 200} {
		base := make([]float64, members)
		for i := range base {
			base[i] = .25 + .675*float64(i)/float64(members-1)
		}
		rows := make([]Row, 32)
		for i := range rows {
			rows[i] = Row{i % members, i % 2, -1}
			if i%4 == 0 {
				rows[i].Second = 1 - rows[i].First
			}
		}
		b.Run(stringName(members), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, e := Run(base, Config{1. / 16, .25, 36}, rows)
				if e != nil {
					b.Fatal(e)
				}
				sink = r
			}
		})
	}
}

func stringName(members int) string {
	switch members {
	case 2:
		return "members2_cap36_rows32"
	case 150:
		return "members150_cap36_rows32"
	default:
		return "members200_cap36_rows32"
	}
}

func TestConstructorAllocation(t *testing.T) {
	for _, members := range []int{150, 200} {
		base := make([]float64, members)
		for i := range base {
			base[i] = .5
		}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		m, e := New(base, Config{1. / 16, .25, 36})
		runtime.ReadMemStats(&after)
		if e != nil {
			t.Fatal(e)
		}
		runtime.KeepAlive(m)
		t.Logf("members=%d constructor allocated=%d; does NOT reserve a full journal or measure replay/retained peak/RSS/serving", members, after.TotalAlloc-before.TotalAlloc)
	}
}
