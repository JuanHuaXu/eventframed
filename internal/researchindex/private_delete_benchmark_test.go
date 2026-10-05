package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var privateDeletionBenchmarkSink PrivateDeletion

// Component benchmark: source parsing/construction are outside the timer;
// discovery, metric work and all intermediate/final immutable edits are inside.
func BenchmarkPrivateDelete(b *testing.B) {
	path := os.Getenv("RESEARCH_DELETE_CAPTURE")
	if path == "" {
		b.Skip("explicit deletion capture")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		b.Fatal(e)
	}
	var rows []struct {
		N             int
		ID, Kind      string
		Before, After layerCapture
	}
	if e = json.Unmarshal(raw, &rows); e != nil {
		b.Fatal(e)
	}
	cases := 0
	for _, r := range rows {
		if r.Kind != "delete" || r.ID != "seed-0" && r.Before.Global == r.After.Global {
			continue
		}
		cases++
		b.Run(fmt.Sprintf("N%d/%s", r.N, r.ID), func(b *testing.B) {
			var initial []LayeredEdit
			var summary EntrySummary
			var target uint32
			found := false
			for id, n := range r.Before.Nodes {
				initial = append(initial, LayeredEdit{id, layerInput(b, n)})
				summary, e = summary.With(id, n.Level)
				if e != nil {
					b.Fatal(e)
				}
				if n.ID == r.ID {
					target = id
					found = true
				}
			}
			if !found {
				b.Fatal("target absent")
			}
			source, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, initial, r.Before.Global, LayeredLimits{len(initial), 768, 32, 1024})
			if e != nil {
				b.Fatal(e)
			}
			probe, e := PreparePrivateDeletion(context.Background(), source, summary, target, 16, 4096, 256, 128, 65536)
			if e != nil {
				b.Fatal(e)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p, e := PreparePrivateDeletion(context.Background(), source, summary, target, 16, 4096, 256, 128, 65536)
				if e != nil {
					b.Fatal(e)
				}
				privateDeletionBenchmarkSink = p
			}
			b.StopTimer()
			b.ReportMetric(float64(probe.Changed), "records/op")
			b.ReportMetric(float64(probe.PairCalls), "pairs/op")
		})
	}
	if cases != 4 {
		b.Fatal("expected four fixed cases", cases)
	}
}
