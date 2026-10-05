package researchindex

import "testing"

// These isolate cache bookkeeping, not HNSW distance work or its pair locality.
// Real construction remains the adoption gate even if this kernel is cheaper.
func BenchmarkBuildPairBookkeeping(b *testing.B) {
	for _, parallel := range []bool{false, true} {
		mode := "serial"
		if parallel {
			mode = "parallel"
		}
		for _, hit := range []bool{false, true} {
			pattern := "miss"
			if hit {
				pattern = "hit"
			}
			for _, kind := range []string{"map", "table"} {
				b.Run(mode+"/"+pattern+"/"+kind, func(b *testing.B) {
					cache, _ := NewBuildPairCache(65536)
					table, _ := NewBuildPairTable(65536)
					compute := func() float32 { return 0.5 }
					if hit {
						cache.Distance(1, 2, compute)
						table.Store(1, 2, compute())
					} else {
						// Fill the non-evicting map with other keys. The table's
						// misses intentionally do not store: lookup-only cost.
						for i := uint32(0); i < 65536; i++ {
							cache.Distance(3, i, compute)
						}
					}
					op := func() {
						if kind == "map" {
							cache.Distance(1, 2, compute)
						} else {
							table.Lookup(1, 2)
						}
					}
					b.ReportAllocs()
					b.ResetTimer()
					if parallel {
						b.RunParallel(func(pb *testing.PB) {
							for pb.Next() {
								op()
							}
						})
					} else {
						for i := 0; i < b.N; i++ {
							op()
						}
					}
				})
			}
		}
	}
}
