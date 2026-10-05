package researchbatch

import (
	"os"
	"testing"
	"time"
)

func TestBatchCapV2(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_CAP_V2") != "1" {
		t.Skip("opt-in batch-size reader-protection screen")
	}
	type totals struct {
		ack, read []int64
		writer    time.Duration
	}
	all := map[int]*totals{1: {}, 4: {}, 8: {}}
	orders := [][]int{{1, 4, 8}, {8, 1, 4}, {4, 8, 1}}
	for trial, order := range orders {
		for _, cap := range order {
			r := runBatchAckLoad(t, cap)
			a := all[cap]
			a.ack = append(a.ack, r.ackOffer...)
			a.read = append(a.read, r.readCall...)
			a.writer += r.writerTotal
			t.Logf("trial=%d cap=%d writes=%d reads=%d batches=%d size_counts_1_to_16=%v ack_p99=%s read_call_p99=%s writer_total=%s write_gap_p50=%s write_gap_p99=%s", trial, cap, r.writes, r.readCount, len(r.batchSizes), batchSizeDistribution(r.batchSizes)[1:], percentileDuration(r.ackOffer, .99), percentileDuration(r.readCall, .99), r.writerTotal, percentileDuration(r.writeGaps, .5), percentileDuration(r.writeGaps, .99))
		}
	}
	control := all[1]
	controlAck := percentileDuration(control.ack, .99)
	controlRead := percentileDuration(control.read, .99)
	advanced := false
	for _, cap := range []int{4, 8} {
		candidate := all[cap]
		ackRatio := float64(percentileDuration(candidate.ack, .99)) / float64(controlAck)
		readRatio := float64(percentileDuration(candidate.read, .99)) / float64(controlRead)
		writerRatio := float64(candidate.writer) / float64(control.writer)
		pass := len(candidate.ack) == 768 && len(candidate.read) == 576 && writerRatio < 1 && ackRatio <= 1.25 && readRatio <= 1.10
		t.Logf("pooled cap=%d control_ack_p99=%s candidate_ack_p99=%s ack_ratio=%.3f control_read_p99=%s candidate_read_p99=%s read_ratio=%.3f writer_total_ratio=%.3f design_screen_pass=%v", cap, controlAck, percentileDuration(candidate.ack, .99), ackRatio, controlRead, percentileDuration(candidate.read, .99), readRatio, writerRatio, pass)
		advanced = advanced || pass
	}
	if len(control.ack) != 768 || len(control.read) != 576 || !advanced {
		t.Error("no cap passed the frozen batch reader-protection design screen")
	}
}
