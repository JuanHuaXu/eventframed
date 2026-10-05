package service

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestResearchRecallCadenceV30(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_CADENCE_V30") != "1" {
		t.Skip("opt-in full Recall cadence envelope")
	}
	for _, trial := range []struct {
		name  string
		gap   time.Duration
		order [2]bool
	}{
		{"8ms", 8 * time.Millisecond, [2]bool{false, true}},
		{"6ms", 6 * time.Millisecond, [2]bool{true, false}},
		{"4ms", 4 * time.Millisecond, [2]bool{false, true}},
	} {
		t.Run(trial.name, func(t *testing.T) {
			var off, on researchRecallLiveTrialV26
			for _, enabled := range trial.order {
				r := runResearchRecallLiveCadenceV30(t, enabled, consumeResearchLiveDecoupledV29, true, trial.gap)
				medianGap := researchDurableLoadPercentile(r.gaps, .5)
				t.Logf("enabled=%v offers=%d labels=%d drops=%d gap_p50=%s gap_p99=%s offer_p99=%s queue_p99=%s frontier_age_p99=%s feedback_age_p99=%s", enabled, len(r.offer), r.labels, r.dropped,
					medianGap, researchDurableLoadPercentile(r.gaps, .99),
					researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.queue, .99), researchLiveAgeV26(r.frontierAge), researchLiveAgeV26(r.feedbackAge))
				if medianGap < 3*trial.gap/4 || medianGap > 5*trial.gap/4 {
					t.Errorf("offered cadence outside frozen tolerance: target=%s measured=%s", trial.gap, medianGap)
				}
				if enabled {
					on = r
					checkResearchLivePhasesV27(t, r)
				} else {
					off = r
				}
			}
			offP99 := researchDurableLoadPercentile(off.offer, .99)
			onP99 := researchDurableLoadPercentile(on.offer, .99)
			frontierP99 := researchDurableLoadPercentile(on.frontierAge, .99)
			feedbackP99 := researchDurableLoadPercentile(on.feedbackAge, .99)
			t.Logf("rate=%s off_p99=%s on_p99=%s ratio=%.3f frontier_age_p99=%s feedback_age_p99=%s", trial.gap, offP99, onP99, float64(onP99)/float64(offP99), frontierP99, feedbackP99)
			if len(off.offer) != 192 || len(on.offer) != 192 || on.labels != 64 || on.dropped != 0 || onP99 >= 100*time.Millisecond || float64(onP99) > 1.10*float64(offP99) || frontierP99 >= 250*time.Millisecond || feedbackP99 >= 250*time.Millisecond {
				t.Error("frozen v30 cadence screen failed")
			}
		})
	}
}

func TestResearchRecallDecoupledDrainV29ClosedTap(t *testing.T) {
	tap, err := NewResearchFrontierTap("tenant-a", 64)
	if err != nil {
		t.Fatal(err)
	}
	tap.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := consumeResearchLiveDecoupledV29(ctx, nil, tap, nil)
	if result.err == nil || result.seen != 0 || result.admitted != 0 || ctx.Err() != nil {
		t.Fatalf("missing frontier not rejected promptly: seen=%d labels=%d err=%v context=%v", result.seen, result.admitted, result.err, ctx.Err())
	}
}
