package service

import (
	"os"
	"testing"
)

func checkResearchLivePhasesV27(t *testing.T, r researchRecallLiveTrialV26) {
	t.Helper()
	if len(r.frontierAge) != 64 || len(r.feedbackAge) != 64 || len(r.tapWait) != 64 || len(r.preFeedback) != 64 || len(r.feedbackGuard) != 64 || len(r.publicationWait) != 64 {
		t.Fatalf("incomplete phase accounting frontier=%d feedback=%d tap=%d pre=%d guarded=%d published=%d", len(r.frontierAge), len(r.feedbackAge), len(r.tapWait), len(r.preFeedback), len(r.feedbackGuard), len(r.publicationWait))
	}
	for i := range r.frontierAge {
		if r.frontierAge[i] != r.tapWait[i]+r.preFeedback[i]+r.feedbackGuard[i]+r.publicationWait[i] || r.feedbackAge[i] != r.feedbackGuard[i]+r.publicationWait[i] {
			t.Fatalf("phase conservation failed at label %d", i+1)
		}
	}
}

func TestResearchRecallLivePhasesV27(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_LIVE_PHASES_V27") != "1" {
		t.Skip("opt-in full Recall live-age phase diagnostic")
	}
	var pooled researchRecallLiveTrialV26
	for trial := range 3 {
		r := runResearchRecallLiveV26(t, true)
		checkResearchLivePhasesV27(t, r)
		pooled.offer = append(pooled.offer, r.offer...)
		pooled.frontierAge = append(pooled.frontierAge, r.frontierAge...)
		pooled.feedbackAge = append(pooled.feedbackAge, r.feedbackAge...)
		pooled.tapWait = append(pooled.tapWait, r.tapWait...)
		pooled.preFeedback = append(pooled.preFeedback, r.preFeedback...)
		pooled.feedbackGuard = append(pooled.feedbackGuard, r.feedbackGuard...)
		pooled.publicationWait = append(pooled.publicationWait, r.publicationWait...)
		t.Logf("trial=%d offer_p99=%s frontier_p99=%s tap_wait_p50=%s tap_wait_p99=%s pre_feedback_p50=%s pre_feedback_p99=%s feedback_guard_p50=%s feedback_guard_p99=%s publication_wait_p50=%s publication_wait_p99=%s", trial,
			researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99),
			researchDurableLoadPercentile(r.tapWait, .5), researchDurableLoadPercentile(r.tapWait, .99),
			researchDurableLoadPercentile(r.preFeedback, .5), researchDurableLoadPercentile(r.preFeedback, .99),
			researchDurableLoadPercentile(r.feedbackGuard, .5), researchDurableLoadPercentile(r.feedbackGuard, .99),
			researchDurableLoadPercentile(r.publicationWait, .5), researchDurableLoadPercentile(r.publicationWait, .99))
	}
	t.Logf("pooled offers=%d labels=%d offer_p99=%s frontier_p99=%s tap_wait_p50=%s tap_wait_p99=%s pre_feedback_p50=%s pre_feedback_p99=%s feedback_guard_p50=%s feedback_guard_p99=%s publication_wait_p50=%s publication_wait_p99=%s", len(pooled.offer), len(pooled.frontierAge),
		researchDurableLoadPercentile(pooled.offer, .99), researchDurableLoadPercentile(pooled.frontierAge, .99),
		researchDurableLoadPercentile(pooled.tapWait, .5), researchDurableLoadPercentile(pooled.tapWait, .99),
		researchDurableLoadPercentile(pooled.preFeedback, .5), researchDurableLoadPercentile(pooled.preFeedback, .99),
		researchDurableLoadPercentile(pooled.feedbackGuard, .5), researchDurableLoadPercentile(pooled.feedbackGuard, .99),
		researchDurableLoadPercentile(pooled.publicationWait, .5), researchDurableLoadPercentile(pooled.publicationWait, .99))
}

func TestResearchRecallLivePhasesV27FocusedRace(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_LIVE_PHASES_V27_RACE") != "1" {
		t.Skip("opt-in race-stressed as-of order diagnostic")
	}
	r := runResearchRecallLiveV26(t, true)
	checkResearchLivePhasesV27(t, r)
}
