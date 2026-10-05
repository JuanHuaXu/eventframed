package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var privateInsertionBenchmarkSink PrivateInsertion

func BenchmarkPrivateInsert(b *testing.B) {
	path := os.Getenv("RESEARCH_INSERT_CAPTURE")
	if path == "" {
		b.Skip("explicit insertion capture")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		b.Fatal(e)
	}
	var rows []layerRow
	if e = json.Unmarshal(raw, &rows); e != nil {
		b.Fatal(e)
	}
	cases := 0
	for _, row := range rows {
		if row.Kind != "insert" {
			continue
		}
		var newID uint32
		var newNode layerCaptureNode
		for id, r := range row.After.Nodes {
			if _, ok := row.Before.Nodes[id]; !ok {
				newID, newNode = id, r
			}
		}
		if newNode.ID != "new-0" && newNode.ID != "new-7" {
			continue
		}
		cases++
		b.Run(fmt.Sprintf("N%d/%s", row.N, newNode.ID), func(b *testing.B) {
			var edits []LayeredEdit
			var summary EntrySummary
			for id, r := range row.Before.Nodes {
				edits = append(edits, LayeredEdit{id, layerInput(b, r)})
				summary, e = summary.With(id, r.Level)
				if e != nil {
					b.Fatal(e)
				}
			}
			ctx := context.Background()
			s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, row.Before.Global, LayeredLimits{len(edits), 768, 32, 1024})
			if e != nil {
				b.Fatal(e)
			}
			vector := layerVector(newNode.ID)
			run := func() (PrivateInsertion, error) {
				return PreparePrivateInsertion(ctx, s, summary, newID, newNode.ID, vector, newNode.Level, len(edits), 16, 200, 100000, 100000, 128, 1)
			}
			probe, e := run()
			if e != nil {
				b.Fatal(e)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				got, e := run()
				if e != nil {
					b.Fatal(e)
				}
				privateInsertionBenchmarkSink = got
			}
			b.StopTimer()
			b.ReportMetric(float64(probe.Evaluations), "evals/op")
			b.ReportMetric(float64(probe.PairCalls), "pairs/op")
		})
	}
	if cases != 4 {
		b.Fatal("expected four fixed cases", cases)
	}
}
