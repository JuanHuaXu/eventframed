package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallAdmissionWaitV33(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_ADMISSION_WAIT_V33") != "1" {
		t.Skip("opt-in admission-wait attribution")
	}
	for _, gap := range []time.Duration{6 * time.Millisecond, 4 * time.Millisecond} {
		t.Run(gap.String(), func(t *testing.T) {
			r := runResearchRecallLiveCadenceV30(t, true, consumeResearchLiveDecoupledV29, true, gap)
			checkResearchLivePhasesV27(t, r)
			if len(r.preFeedbackPartsV32) != 64 {
				t.Fatalf("incomplete admission-wait parts: %d", len(r.preFeedbackPartsV32))
			}
			lookup, handoff, ordered := make([]int64, 0, 64), make([]int64, 0, 64), make([]int64, 0, 64)
			for i, parts := range r.preFeedbackPartsV32 {
				if parts.journalLookup < 0 || parts.selectedHandoff < 0 || parts.orderedWait < 0 || parts.journalLookup+parts.selectedHandoff+parts.orderedWait != parts.beforeAdmission || parts.beforeAdmission+parts.admissionGuard+parts.feedbackQueue != r.preFeedback[i] {
					t.Fatalf("admission-wait conservation failed at label %d: parts=%+v total=%s", i+1, parts, time.Duration(r.preFeedback[i]))
				}
				lookup = append(lookup, parts.journalLookup)
				handoff = append(handoff, parts.selectedHandoff)
				ordered = append(ordered, parts.orderedWait)
			}
			t.Logf("gap=%s offers=%d labels=%d drops=%d actual_gap_p50=%s offer_p99=%s frontier_age_p99=%s", gap, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99))
			t.Logf("stage journal_lookup_p50=%s journal_lookup_p99=%s selected_handoff_p50=%s selected_handoff_p99=%s ordered_wait_p50=%s ordered_wait_p99=%s admission_guard_p50=%s admission_guard_p99=%s",
				researchDurableLoadPercentile(lookup, .5), researchDurableLoadPercentile(lookup, .99),
				researchDurableLoadPercentile(handoff, .5), researchDurableLoadPercentile(handoff, .99),
				researchDurableLoadPercentile(ordered, .5), researchDurableLoadPercentile(ordered, .99),
				researchDurableLoadPercentile(extractAdmissionGuardV33(r.preFeedbackPartsV32), .5), researchDurableLoadPercentile(extractAdmissionGuardV33(r.preFeedbackPartsV32), .99))
			indices := make([]int, 64)
			for i := range indices {
				indices[i] = i
			}
			sort.Slice(indices, func(i, j int) bool { return r.frontierAge[indices[i]] > r.frontierAge[indices[j]] })
			for _, index := range indices[:5] {
				parts := r.preFeedbackPartsV32[index]
				t.Logf("oldest label=%d age=%s journal_lookup=%s selected_handoff=%s ordered_wait=%s admission_guard=%s feedback_queue=%s", index+1, time.Duration(r.frontierAge[index]),
					time.Duration(parts.journalLookup), time.Duration(parts.selectedHandoff), time.Duration(parts.orderedWait), time.Duration(parts.admissionGuard), time.Duration(parts.feedbackQueue))
			}
		})
	}
}

func extractAdmissionGuardV33(parts []researchPreFeedbackPartsV32) []int64 {
	out := make([]int64, len(parts))
	for i, part := range parts {
		out[i] = part.admissionGuard
	}
	return out
}
