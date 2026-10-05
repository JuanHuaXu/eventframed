package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallPreFeedbackV32(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PREFEEDBACK_V32") != "1" {
		t.Skip("opt-in pre-feedback phase attribution")
	}
	for _, gap := range []time.Duration{6 * time.Millisecond, 4 * time.Millisecond} {
		t.Run(gap.String(), func(t *testing.T) {
			r := runResearchRecallLiveCadenceV30(t, true, consumeResearchLiveDecoupledV29, true, gap)
			checkResearchLivePhasesV27(t, r)
			if len(r.preFeedbackPartsV32) != 64 {
				t.Fatalf("incomplete pre-feedback parts: %d", len(r.preFeedbackPartsV32))
			}
			before := make([]int64, 0, 64)
			admission := make([]int64, 0, 64)
			queue := make([]int64, 0, 64)
			for i, parts := range r.preFeedbackPartsV32 {
				if parts.beforeAdmission < 0 || parts.admissionGuard < 0 || parts.feedbackQueue < 0 || parts.beforeAdmission+parts.admissionGuard+parts.feedbackQueue != r.preFeedback[i] {
					t.Fatalf("pre-feedback conservation failed at label %d: parts=%+v total=%s", i+1, parts, time.Duration(r.preFeedback[i]))
				}
				before = append(before, parts.beforeAdmission)
				admission = append(admission, parts.admissionGuard)
				queue = append(queue, parts.feedbackQueue)
			}
			t.Logf("gap=%s offers=%d labels=%d drops=%d actual_gap_p50=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s", gap, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99), researchDurableLoadPercentile(r.feedbackAge, .99))
			t.Logf("stages before_admission_p50=%s before_admission_p99=%s admission_guard_p50=%s admission_guard_p99=%s feedback_queue_p50=%s feedback_queue_p99=%s",
				researchDurableLoadPercentile(before, .5), researchDurableLoadPercentile(before, .99),
				researchDurableLoadPercentile(admission, .5), researchDurableLoadPercentile(admission, .99),
				researchDurableLoadPercentile(queue, .5), researchDurableLoadPercentile(queue, .99))
			indices := make([]int, 64)
			for i := range indices {
				indices[i] = i
			}
			sort.Slice(indices, func(i, j int) bool { return r.frontierAge[indices[i]] > r.frontierAge[indices[j]] })
			for _, index := range indices[:5] {
				parts := r.preFeedbackPartsV32[index]
				t.Logf("oldest label=%d age=%s before_admission=%s admission_guard=%s feedback_queue=%s", index+1, time.Duration(r.frontierAge[index]),
					time.Duration(parts.beforeAdmission), time.Duration(parts.admissionGuard), time.Duration(parts.feedbackQueue))
			}
		})
	}
}
