package researchpublication

import (
	"strconv"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func BenchmarkRestoredPublicationCompatibility(b *testing.B) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, n := range []uint64{1, 128, 4096} {
		b.Run("versions-"+strconv.FormatUint(n, 10), func(b *testing.B) {
			current := model.Snapshot{RuntimeVersion: n, EvidenceEpoch: n}
			motion := make(map[uint64]time.Time, n)
			for version := uint64(1); version <= n; version++ {
				motion[version] = now.Add(time.Hour)
			}
			p, err := NewWithMotion(current, motion)
			if err != nil {
				b.Fatal(err)
			}
			old := model.Snapshot{}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !p.Compatible(old, now) {
					b.Fatal("valid restored motion rejected")
				}
			}
		})
	}
}
