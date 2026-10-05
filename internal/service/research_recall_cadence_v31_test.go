package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallCadenceV31ConfirmSixMillis(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_CADENCE_V31") != "1" {
		t.Skip("opt-in six-millisecond cadence confirmation")
	}
	type arm struct {
		offer, frontierAge, feedbackAge []int64
	}
	all := map[bool]*arm{false: {}, true: {}}
	for trial := range 3 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, enabled := range order {
			r := runResearchRecallLiveCadenceV30(t, enabled, consumeResearchLiveDecoupledV29, true, 6*time.Millisecond)
			medianGap := researchDurableLoadPercentile(r.gaps, .5)
			t.Logf("trial=%d enabled=%v offers=%d labels=%d drops=%d gap_p50=%s gap_p99=%s offer_p99=%s queue_p99=%s frontier_age_p99=%s feedback_age_p99=%s", trial, enabled, len(r.offer), r.labels, r.dropped,
				medianGap, researchDurableLoadPercentile(r.gaps, .99),
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.queue, .99), researchLiveAgeV26(r.frontierAge), researchLiveAgeV26(r.feedbackAge))
			if medianGap < 4500*time.Microsecond || medianGap > 7500*time.Microsecond {
				t.Errorf("trial=%d enabled=%v cadence outside frozen tolerance: %s", trial, enabled, medianGap)
			}
			if enabled {
				checkResearchLivePhasesV27(t, r)
				if researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
					t.Errorf("trial=%d serving p99 failed", trial)
				}
			}
			all[enabled].offer = append(all[enabled].offer, r.offer...)
			all[enabled].frontierAge = append(all[enabled].frontierAge, r.frontierAge...)
			all[enabled].feedbackAge = append(all[enabled].feedbackAge, r.feedbackAge...)
		}
	}
	offP99 := researchDurableLoadPercentile(all[false].offer, .99)
	onP99 := researchDurableLoadPercentile(all[true].offer, .99)
	frontierP99 := researchDurableLoadPercentile(all[true].frontierAge, .99)
	feedbackP99 := researchDurableLoadPercentile(all[true].feedbackAge, .99)
	t.Logf("pooled off_p99=%s on_p99=%s ratio=%.3f labels=%d frontier_age_p99=%s feedback_age_p99=%s", offP99, onP99, float64(onP99)/float64(offP99), len(all[true].frontierAge), frontierP99, feedbackP99)
	if len(all[false].offer) != 576 || len(all[true].offer) != 576 || len(all[true].frontierAge) != 192 || onP99 >= 100*time.Millisecond || float64(onP99) > 1.10*float64(offP99) || frontierP99 >= 250*time.Millisecond || feedbackP99 >= 250*time.Millisecond {
		t.Error("frozen v31 six-millisecond confirmation failed")
	}
}

func TestResearchRecallCadenceV31TraceFourMillis(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_CADENCE_V31_TRACE") != "1" {
		t.Skip("opt-in four-millisecond phase trace")
	}
	r := runResearchRecallLiveCadenceV30(t, true, consumeResearchLiveDecoupledV29, true, 4*time.Millisecond)
	checkResearchLivePhasesV27(t, r)
	t.Logf("offers=%d labels=%d drops=%d gap_p50=%s gap_p99=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s", len(r.offer), r.labels, r.dropped,
		researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.gaps, .99),
		researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99), researchDurableLoadPercentile(r.feedbackAge, .99))
	t.Logf("phase tap_wait_p50=%s tap_wait_p99=%s pre_feedback_p50=%s pre_feedback_p99=%s feedback_guard_p50=%s feedback_guard_p99=%s publication_wait_p50=%s publication_wait_p99=%s",
		researchDurableLoadPercentile(r.tapWait, .5), researchDurableLoadPercentile(r.tapWait, .99),
		researchDurableLoadPercentile(r.preFeedback, .5), researchDurableLoadPercentile(r.preFeedback, .99),
		researchDurableLoadPercentile(r.feedbackGuard, .5), researchDurableLoadPercentile(r.feedbackGuard, .99),
		researchDurableLoadPercentile(r.publicationWait, .5), researchDurableLoadPercentile(r.publicationWait, .99))
	indices := make([]int, len(r.frontierAge))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return r.frontierAge[indices[i]] > r.frontierAge[indices[j]]
	})
	for _, index := range indices[:5] {
		t.Logf("oldest label=%d total=%s tap_wait=%s pre_feedback=%s feedback_guard=%s publication_wait=%s", index+1,
			time.Duration(r.frontierAge[index]), time.Duration(r.tapWait[index]), time.Duration(r.preFeedback[index]), time.Duration(r.feedbackGuard[index]), time.Duration(r.publicationWait[index]))
	}
}
