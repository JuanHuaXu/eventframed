package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

func TestResearchRecallDeadlineReplicationV39(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_DEADLINE_V39") != "1" {
		t.Skip("opt-in unchanged deadline-scheduler replication")
	}
	type arm struct {
		offer, frontierAge, feedbackAge []int64
		writerOffer, writerCall         []int64
		writerTotal                     int64
		writerStartedOverlap            int
	}
	all := map[bool]*arm{false: {}, true: {}}
	allFrontierTrialsPass := true
	for trial := range 4 {
		order := []bool{false, true}
		if trial%2 == 1 {
			order = []bool{true, false}
		}
		var pair [2]time.Duration
		for _, deadlineScheduled := range order {
			var gate researchSchedulingGateV38
			consume := consumeResearchLiveDecoupledV29
			if deadlineScheduled {
				gate = newResearchDeadlineGateV38()
				consume = func(ctx context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
					return consumeResearchLiveDecoupledWithGateV37(ctx, s, tap, durable, gate)
				}
			}
			r := runResearchRecallLiveWithPolicyV37(t, true, consume, true, 4*time.Millisecond, true, gate, true)
			checkResearchLivePhasesV27(t, r)
			if len(r.writerOffer) != 256 || len(r.writerCall) != 256 || r.writerTotal <= 0 {
				t.Fatalf("missing writer timing: offered=%d call=%d total=%s", len(r.writerOffer), len(r.writerCall), time.Duration(r.writerTotal))
			}
			frontierP99 := researchDurableLoadPercentile(r.frontierAge, .99)
			writerP99 := researchDurableLoadPercentile(r.writerOffer, .99)
			if deadlineScheduled {
				pair[1] = writerP99
				if frontierP99 >= 250*time.Millisecond {
					allFrontierTrialsPass = false
				}
				if researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
					t.Errorf("frozen trial serving gate failed: trial=%d", trial)
				}
			} else {
				pair[0] = writerP99
			}
			t.Logf("trial=%d deadline=%v offers=%d labels=%d drops=%d writes=%d offered_overlap=%d started_overlap=%d gap_p50=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s writer_offer_p99=%s writer_call_p99=%s writer_total=%s", trial, deadlineScheduled, len(r.offer), r.labels, r.dropped, r.writes, r.overlap, r.writerStartedOverlap,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), frontierP99, researchDurableLoadPercentile(r.feedbackAge, .99), writerP99, researchDurableLoadPercentile(r.writerCall, .99), time.Duration(r.writerTotal))
			a := all[deadlineScheduled]
			a.offer = append(a.offer, r.offer...)
			a.frontierAge = append(a.frontierAge, r.frontierAge...)
			a.feedbackAge = append(a.feedbackAge, r.feedbackAge...)
			a.writerOffer = append(a.writerOffer, r.writerOffer...)
			a.writerCall = append(a.writerCall, r.writerCall...)
			a.writerTotal += r.writerTotal
			a.writerStartedOverlap += r.writerStartedOverlap
		}
		t.Logf("trial=%d matched_writer_p99_ratio=%.3f", trial, float64(pair[1])/float64(pair[0]))
	}
	control, candidate := all[false], all[true]
	controlOffer := researchDurableLoadPercentile(control.offer, .99)
	candidateOffer := researchDurableLoadPercentile(candidate.offer, .99)
	frontierAge := researchDurableLoadPercentile(candidate.frontierAge, .99)
	feedbackAge := researchDurableLoadPercentile(candidate.feedbackAge, .99)
	controlWriter := researchDurableLoadPercentile(control.writerOffer, .99)
	candidateWriter := researchDurableLoadPercentile(candidate.writerOffer, .99)
	t.Logf("pooled control_offer_p99=%s candidate_offer_p99=%s offer_ratio=%.3f candidate_frontier_age_p99=%s candidate_feedback_age_p99=%s control_writer_offer_p99=%s candidate_writer_offer_p99=%s writer_p99_ratio=%.3f writer_total_ratio=%.3f started_overlap_control=%d candidate=%d every_candidate_frontier_trial_pass=%v", controlOffer, candidateOffer, float64(candidateOffer)/float64(controlOffer), frontierAge, feedbackAge,
		controlWriter, candidateWriter, float64(candidateWriter)/float64(controlWriter), float64(candidate.writerTotal)/float64(control.writerTotal), control.writerStartedOverlap, candidate.writerStartedOverlap, allFrontierTrialsPass)
	if len(control.offer) != 768 || len(candidate.offer) != 768 || len(candidate.frontierAge) != 256 || len(candidate.writerOffer) != 1024 || candidateOffer >= 100*time.Millisecond || float64(candidateOffer) > 1.10*float64(controlOffer) || frontierAge >= 250*time.Millisecond || feedbackAge >= 250*time.Millisecond || float64(candidateWriter) > 1.25*float64(controlWriter) || float64(candidate.writerTotal) > 1.25*float64(control.writerTotal) || candidate.writerStartedOverlap*5 < control.writerStartedOverlap*4 {
		t.Error("frozen v39 deadline replication component failed")
	}
}
