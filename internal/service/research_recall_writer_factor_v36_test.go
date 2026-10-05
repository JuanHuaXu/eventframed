package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallWriterFactorV36(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_WRITER_FACTOR_V36") != "1" {
		t.Skip("opt-in writer-factor guard diagnostic")
	}
	for _, cell := range []struct {
		gap   time.Duration
		order [2]bool
	}{
		{6 * time.Millisecond, [2]bool{true, false}},
		{4 * time.Millisecond, [2]bool{false, true}},
	} {
		t.Run(cell.gap.String(), func(t *testing.T) {
			for _, withWriter := range cell.order {
				name := "writer_off"
				if withWriter {
					name = "writer_on"
				}
				t.Run(name, func(t *testing.T) {
					r := runResearchRecallLiveWithWritesV36(t, true, consumeResearchLiveDecoupledV29, true, cell.gap, withWriter)
					checkResearchLivePhasesV27(t, r)
					if len(r.admissionPartsV34) != 64 {
						t.Fatalf("incomplete writer-factor profile: %d", len(r.admissionPartsV34))
					}
					entry, candidate := make([]int64, 0, 64), make([]int64, 0, 64)
					for i, parts := range r.admissionPartsV34 {
						if parts.guardEntry < 0 || parts.candidateValidation < 0 || parts.guardEntry+parts.candidateValidation != parts.sourceValidation {
							t.Fatalf("writer-factor conservation failed at label %d: %+v", i+1, parts)
						}
						entry = append(entry, parts.guardEntry)
						candidate = append(candidate, parts.candidateValidation)
					}
					t.Logf("gap=%s writer=%v offers=%d writes=%d overlap=%d labels=%d drops=%d actual_gap_p50=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s guard_entry_p50=%s guard_entry_p99=%s candidate_validation_p50=%s candidate_validation_p99=%s", cell.gap, withWriter, len(r.offer), r.writes, r.overlap, r.labels, r.dropped,
						researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99), researchDurableLoadPercentile(r.feedbackAge, .99),
						researchDurableLoadPercentile(entry, .5), researchDurableLoadPercentile(entry, .99), researchDurableLoadPercentile(candidate, .5), researchDurableLoadPercentile(candidate, .99))
					indices := make([]int, 64)
					for i := range indices {
						indices[i] = i
					}
					sort.Slice(indices, func(i, j int) bool { return r.frontierAge[indices[i]] > r.frontierAge[indices[j]] })
					for _, index := range indices[:5] {
						t.Logf("oldest label=%d age=%s guard_entry=%s candidate_validation=%s", index+1, time.Duration(r.frontierAge[index]),
							time.Duration(r.admissionPartsV34[index].guardEntry), time.Duration(r.admissionPartsV34[index].candidateValidation))
					}
				})
			}
		})
	}
}
