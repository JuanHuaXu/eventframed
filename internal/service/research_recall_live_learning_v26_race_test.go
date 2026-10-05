package service

import (
	"os"
	"testing"
)

func TestResearchRecallLiveLearningV26FocusedRace(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_LIVE_V26_RACE") != "1" {
		t.Skip("opt-in focused live-learning race check")
	}
	r := runResearchRecallLiveV26(t, true)
	if len(r.offer) != 192 || len(r.frontierAge) != 64 || len(r.feedbackAge) != 64 || r.labels != 64 || r.dropped != 0 {
		t.Fatalf("incomplete focused race trial: offers=%d ages=%d/%d labels=%d drops=%d", len(r.offer), len(r.frontierAge), len(r.feedbackAge), r.labels, r.dropped)
	}
}
