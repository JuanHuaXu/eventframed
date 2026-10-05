package libravdbstore

import (
	"os"
	"testing"
	"time"
)

func TestResearchRollingLSNReaderV3(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ROLLING_LSN_READER_V3") != "1" {
		t.Skip("opt-in sortable-key rolling LSN screen")
	}
	type totals struct {
		call, offer, lags []int64
		writer            time.Duration
	}
	all := map[bool]*totals{false: {}, true: {}}
	for trial := 0; trial < 3; trial++ {
		order := []bool{false, true}
		if trial%2 == 1 {
			order = []bool{true, false}
		}
		for _, pinned := range order {
			r := runRollingReaderTrialWithKey(t, pinned, true, true)
			a := all[pinned]
			a.call = append(a.call, r.call...)
			a.offer = append(a.offer, r.offer...)
			a.lags = append(a.lags, r.lags...)
			a.writer += r.writerTotal
			t.Logf("trial=%d pinned=%v call_p50=%s call_p99=%s offer_p99=%s gap_p50=%s gap_p99=%s writer_total=%s write_ack_p99=%s lag_p99=%s max_lag=%s retained_bytes=%d",
				trial, pinned, pinnedPercentile(r.call, .5), pinnedPercentile(r.call, .99), pinnedPercentile(r.offer, .99),
				pinnedPercentile(r.gaps, .5), pinnedPercentile(r.gaps, .99), r.writerTotal, pinnedPercentile(r.writeAcks, .99),
				pinnedPercentile(r.lags, .99), r.maxLag, r.retainedBytes)
		}
	}
	current, candidate := all[false], all[true]
	readP99 := pinnedPercentile(candidate.call, .99)
	readRatio := float64(readP99) / float64(pinnedPercentile(current.call, .99))
	writerRatio := float64(candidate.writer) / float64(current.writer)
	lagP99 := pinnedPercentile(candidate.lags, .99)
	t.Logf("pooled current_call_p99=%s rolling_call_p99=%s read_ratio=%.3f current_offer_p99=%s rolling_offer_p99=%s writer_total_ratio=%.3f current_lag_p99=%s rolling_lag_p99=%s",
		pinnedPercentile(current.call, .99), readP99, readRatio, pinnedPercentile(current.offer, .99),
		pinnedPercentile(candidate.offer, .99), writerRatio, pinnedPercentile(current.lags, .99), lagP99)
	if len(current.call) != 576 || len(candidate.call) != 576 || len(candidate.lags) != 48 || readRatio > 1.10 ||
		readP99 > 100*time.Millisecond || writerRatio > 1.25 || lagP99 > 25*time.Millisecond {
		t.Error("rolling-LSN v3 failed frozen sortable-key screen")
	}
}

func TestResearchRollingLSNReaderV3Race(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ROLLING_LSN_READER_V3_RACE") != "1" {
		t.Skip("opt-in v3 correctness under race instrumentation; no latency gate")
	}
	for _, pinned := range []bool{false, true} {
		r := runRollingReaderTrialWithKey(t, pinned, true, false)
		if len(r.call) != 192 {
			t.Fatal("incomplete v3 race diagnostic", pinned, len(r.call))
		}
		t.Logf("race diagnostic pinned=%v reads=%d post_ack_batches_sampled=%d", pinned, len(r.call), len(r.lags))
	}
}
