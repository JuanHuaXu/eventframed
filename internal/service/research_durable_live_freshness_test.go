package service

import (
	"os"
	"testing"
	"time"
)

func researchGuardedDurableLiveFreshness(t *testing.T, ageOrigin researchDurableAgeOrigin) {
	t.Helper()
	for _, writer := range []bool{false, true} {
		name := "quiet"
		if writer {
			name = "future-writer"
		}
		t.Run(name, func(t *testing.T) {
			var recall, feedback, age []int64
			writes, overlaps := 0, 0
			for trial := 0; trial < 3; trial++ {
				r := researchDurableMixedLoadTrialAt(t, writer, ageOrigin)
				recall = append(recall, r.recallNS...)
				feedback = append(feedback, r.feedbackNS...)
				age = append(age, r.liveAgeNS...)
				writes += r.writes
				overlaps += r.overlaps
			}
			t.Logf("arm=%s calls=%d live_labels=%d writes=%d overlaps=%d recall_p50=%s recall_p95=%s recall_p99=%s feedback_p99=%s live_age_p50=%s live_age_p95=%s live_age_p99=%s live_age_max=%s",
				name, len(recall), len(age), writes, overlaps,
				researchDurableLoadPercentile(recall, .5), researchDurableLoadPercentile(recall, .95), researchDurableLoadPercentile(recall, .99),
				researchDurableLoadPercentile(feedback, .99),
				researchDurableLoadPercentile(age, .5), researchDurableLoadPercentile(age, .95), researchDurableLoadPercentile(age, .99), researchDurableLoadPercentile(age, 1))
			if writer && os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" {
				if researchDurableLoadPercentile(recall, .99) >= 100*time.Millisecond {
					t.Error("predeclared finite writer-arm recall p99 gate failed")
				}
				if researchDurableLoadPercentile(age, .99) >= 250*time.Millisecond {
					t.Error("predeclared finite live-completion p99 gate failed")
				}
			}
		})
	}
}

func TestResearchGuardedDurableLiveFreshnessV1(t *testing.T) {
	researchGuardedDurableLiveFreshness(t, researchDurablePostReturnAge)
}

func TestResearchGuardedDurableLiveFreshnessV2(t *testing.T) {
	researchGuardedDurableLiveFreshness(t, researchDurableOfferedAge)
}
