package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallAdmissionGuardV34(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_ADMISSION_GUARD_V34") != "1" {
		t.Skip("opt-in guarded admission sub-operation profile")
	}
	for _, gap := range []time.Duration{6 * time.Millisecond, 4 * time.Millisecond} {
		t.Run(gap.String(), func(t *testing.T) {
			r := runResearchRecallLiveCadenceV30(t, true, consumeResearchLiveDecoupledV29, true, gap)
			checkResearchLivePhasesV27(t, r)
			if len(r.preFeedbackPartsV32) != 64 || len(r.admissionPartsV34) != 64 {
				t.Fatalf("incomplete admission profile: outer=%d inner=%d", len(r.preFeedbackPartsV32), len(r.admissionPartsV34))
			}
			source, admit, readback, post, exit := make([]int64, 0, 64), make([]int64, 0, 64), make([]int64, 0, 64), make([]int64, 0, 64), make([]int64, 0, 64)
			for i, parts := range r.admissionPartsV34 {
				if parts.sourceValidation < 0 || parts.durableAdmit < 0 || parts.durableRead < 0 || parts.postRecordValidation < 0 || parts.guardExit < 0 || parts.sourceValidation+parts.durableAdmit+parts.durableRead+parts.postRecordValidation+parts.guardExit != r.preFeedbackPartsV32[i].admissionGuard {
					t.Fatalf("admission sub-operation conservation failed at label %d: parts=%+v total=%s", i+1, parts, time.Duration(r.preFeedbackPartsV32[i].admissionGuard))
				}
				source = append(source, parts.sourceValidation)
				admit = append(admit, parts.durableAdmit)
				readback = append(readback, parts.durableRead)
				post = append(post, parts.postRecordValidation)
				exit = append(exit, parts.guardExit)
			}
			t.Logf("gap=%s offers=%d labels=%d drops=%d actual_gap_p50=%s offer_p99=%s frontier_age_p99=%s", gap, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99))
			t.Logf("stage source_validation_p50=%s source_validation_p99=%s durable_admit_p50=%s durable_admit_p99=%s durable_read_p50=%s durable_read_p99=%s post_record_validation_p50=%s post_record_validation_p99=%s guard_exit_p50=%s guard_exit_p99=%s",
				researchDurableLoadPercentile(source, .5), researchDurableLoadPercentile(source, .99),
				researchDurableLoadPercentile(admit, .5), researchDurableLoadPercentile(admit, .99),
				researchDurableLoadPercentile(readback, .5), researchDurableLoadPercentile(readback, .99),
				researchDurableLoadPercentile(post, .5), researchDurableLoadPercentile(post, .99),
				researchDurableLoadPercentile(exit, .5), researchDurableLoadPercentile(exit, .99))
			indices := make([]int, 64)
			for i := range indices {
				indices[i] = i
			}
			sort.Slice(indices, func(i, j int) bool { return r.frontierAge[indices[i]] > r.frontierAge[indices[j]] })
			for _, index := range indices[:5] {
				parts := r.admissionPartsV34[index]
				t.Logf("oldest label=%d age=%s source_validation=%s durable_admit=%s durable_read=%s post_record_validation=%s guard_exit=%s", index+1, time.Duration(r.frontierAge[index]),
					time.Duration(parts.sourceValidation), time.Duration(parts.durableAdmit), time.Duration(parts.durableRead), time.Duration(parts.postRecordValidation), time.Duration(parts.guardExit))
			}
		})
	}
}
