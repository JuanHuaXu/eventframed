package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchSortableReceiptCostV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORTABLE_RECEIPT_COST_V1") != "1" {
		t.Skip("opt-in private receipt writer cost comparison")
	}
	const warmup, measured = 16, 128
	type armResult struct {
		name      string
		latencies []int64
	}
	run := func(block int, useReceipt bool) armResult {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		name := "control"
		if useReceipt {
			name = "receipt"
		}
		s, err := Open(Config{Path: t.TempDir() + "/" + name + ".libravdb", Dimension: 4,
			EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		collection := collectionName("tenant-a", "research:d4")
		col, err := s.db.CreateCollection(ctx, collection, libra.WithDimension(4),
			libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
			libra.WithMemoryMapping(true),
			libra.WithMetadataSchema(libra.MetadataSchema{
				"available_at": libra.StringField, "available_at_sort": libra.StringField,
			}))
		if err != nil {
			t.Fatal(err)
		}
		base := s.Snapshot(ctx)
		beforeLSN, err := s.db.LatestCommitLSN(ctx)
		if err != nil {
			t.Fatal(err)
		}
		latencies := make([]int64, 0, measured)
		second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		for i := 0; i < warmup+measured; i++ {
			write := pinnedWrite(fmt.Sprintf("block%d-%s-%03d", block, name, i), second.Add(time.Duration(i)*time.Millisecond))
			start := time.Now()
			var receipt uint64
			if useReceipt {
				_, receipt, err = s.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write})
			} else {
				_, err = s.PutResearchEventBatchSortable(ctx, []ResearchEventWrite{write})
			}
			elapsed := time.Since(start).Nanoseconds()
			if err != nil {
				t.Fatal("writer call", block, name, i, err)
			}
			if i >= warmup {
				latencies = append(latencies, elapsed)
			}
			latest, err := s.db.LatestCommitLSN(ctx)
			if err != nil || latest <= beforeLSN || useReceipt && receipt != latest {
				t.Fatal("commit boundary", block, name, i, receipt, latest, beforeLSN, err)
			}
			beforeLSN = latest
		}
		rows, err := col.ListAll(ctx)
		if err != nil || len(rows) != warmup+measured ||
			s.Snapshot(ctx).RuntimeVersion != base.RuntimeVersion+warmup+measured {
			t.Fatal("writer lost committed rows", block, name, len(rows), err)
		}
		return armResult{name: name, latencies: latencies}
	}
	var controlAll, receiptAll []int64
	passed := true
	for block := 0; block < 3; block++ {
		order := []bool{false, true}
		if block == 1 {
			order = []bool{true, false}
		}
		results := make(map[string]armResult)
		for _, withReceipt := range order {
			result := run(block, withReceipt)
			results[result.name] = result
		}
		control := results["control"].latencies
		receipt := results["receipt"].latencies
		controlAll = append(controlAll, control...)
		receiptAll = append(receiptAll, receipt...)
		p50Ratio := float64(pinnedPercentile(receipt, .5)) / float64(pinnedPercentile(control, .5))
		p99Ratio := float64(pinnedPercentile(receipt, .99)) / float64(pinnedPercentile(control, .99))
		t.Logf("block=%d control_p50=%s receipt_p50=%s p50_ratio=%.3f control_p99=%s receipt_p99=%s p99_ratio=%.3f",
			block, pinnedPercentile(control, .5), pinnedPercentile(receipt, .5), p50Ratio,
			pinnedPercentile(control, .99), pinnedPercentile(receipt, .99), p99Ratio)
		if p50Ratio > 1.10 || p99Ratio > 1.20 {
			passed = false
		}
	}
	t.Logf("pooled control_p50=%s receipt_p50=%s control_p99=%s receipt_p99=%s",
		pinnedPercentile(controlAll, .5), pinnedPercentile(receiptAll, .5),
		pinnedPercentile(controlAll, .99), pinnedPercentile(receiptAll, .99))
	if !passed {
		t.Error("receipt cost exceeded frozen paired-block ratios")
	}
}
