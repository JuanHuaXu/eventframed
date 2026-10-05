package service

import (
	"os"
	"sort"
	"testing"
	"time"
)

func TestResearchRecallGuardEntryV35(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_GUARD_ENTRY_V35") != "1" {
		t.Skip("opt-in guard-entry versus source-validation profile")
	}
	for _, gap := range []time.Duration{6 * time.Millisecond, 4 * time.Millisecond} {
		t.Run(gap.String(), func(t *testing.T) {
			r := runResearchRecallLiveCadenceV30(t, true, consumeResearchLiveDecoupledV29, true, gap)
			checkResearchLivePhasesV27(t, r)
			if len(r.admissionPartsV34) != 64 {
				t.Fatalf("incomplete guard-entry profile: %d", len(r.admissionPartsV34))
			}
			entry, candidate := make([]int64, 0, 64), make([]int64, 0, 64)
			for i, parts := range r.admissionPartsV34 {
				if parts.guardEntry < 0 || parts.candidateValidation < 0 || parts.guardEntry+parts.candidateValidation != parts.sourceValidation {
					t.Fatalf("guard-entry conservation failed at label %d: parts=%+v", i+1, parts)
				}
				entry = append(entry, parts.guardEntry)
				candidate = append(candidate, parts.candidateValidation)
			}
			t.Logf("gap=%s offers=%d labels=%d drops=%d actual_gap_p50=%s offer_p99=%s frontier_age_p99=%s", gap, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99))
			t.Logf("stage guard_entry_p50=%s guard_entry_p99=%s candidate_validation_p50=%s candidate_validation_p99=%s",
				researchDurableLoadPercentile(entry, .5), researchDurableLoadPercentile(entry, .99),
				researchDurableLoadPercentile(candidate, .5), researchDurableLoadPercentile(candidate, .99))
			indices := make([]int, 64)
			for i := range indices {
				indices[i] = i
			}
			sort.Slice(indices, func(i, j int) bool { return r.frontierAge[indices[i]] > r.frontierAge[indices[j]] })
			for _, index := range indices[:5] {
				parts := r.admissionPartsV34[index]
				t.Logf("oldest label=%d age=%s guard_entry=%s candidate_validation=%s", index+1, time.Duration(r.frontierAge[index]), time.Duration(parts.guardEntry), time.Duration(parts.candidateValidation))
			}
		})
	}
}
