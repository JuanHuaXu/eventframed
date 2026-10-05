package service

import (
	"os"
	"testing"
	"time"
)

func researchCapacityMeanV24(values []int64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	var total int64
	for _, value := range values {
		total += value
	}
	return time.Duration(total / int64(len(values)))
}

func researchCapacityPhaseV24(r recallProfileTrial, name string) string {
	part := r.parts[name]
	if part == nil || len(part.callDurations) == 0 {
		return "not_captured"
	}
	return researchCapacityMeanV24(part.requestSums).String()
}

func TestResearchRecallCapacityV24(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_CAPACITY_V24") != "1" {
		t.Skip("opt-in v24 full Recall capacity diagnostic")
	}
	arms := []string{"sqlite", "group8ms", "group1ns"}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			var r recallProfileTrial
			switch arm {
			case "sqlite":
				r = runRecallProfileTrialMode(t, true, 200, "sqlite")
			case "group8ms":
				r = runRecallProfileTrialMode(t, true, 200, "group")
			case "group1ns":
				r = runResearchRecallDwellTrialV23(t, true, time.Nanosecond)
			}
			if len(r.call) != 192 || len(r.queue) != 192 || r.writes != 256 || r.overlap == 0 || r.guardRejects != 0 {
				t.Fatalf("incomplete capacity sample arm=%s trial=%d calls=%d writes=%d overlap=%d rejects=%d", arm, trial, len(r.call), r.writes, r.overlap, r.guardRejects)
			}
			meanCall := researchCapacityMeanV24(r.call)
			firstQueue := researchDurableLoadPercentile(r.queue[:48], .5)
			lastQueue := researchDurableLoadPercentile(r.queue[144:], .5)
			t.Logf("trial=%d arm=%s mean_call=%s nominal_load=%.3f call_p99=%s queue_p99=%s first_completed_queue_p50=%s last_completed_queue_p50=%s journal_mean=%s guard_wait_mean=%s sqlite_insert_mean=%s",
				trial, arm, meanCall, float64(meanCall)/float64(32*time.Millisecond),
				researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), firstQueue, lastQueue,
				researchCapacityPhaseV24(r, "journal"), researchCapacityPhaseV24(r, "sqlite_guard_wait"), researchCapacityPhaseV24(r, "sqlite_insert"))
		}
	}
}
