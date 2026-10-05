package service

import (
	"os"
	"testing"
	"time"
)

func TestResearchIdleWrapperAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadWrapperArm(t, "off", 0, 8, writes, true, false, false, false, false, false, false, &researchLoadSchedule{5 * time.Millisecond, 10 * time.Millisecond}, true)
		checkQueuedGuardLoad(t, r, writes)
		if err := researchCheckOffered(r, writes); err != nil {
			t.Fatal(err)
		}
		if !r.IdlePublicationWrapper || r.SourceOwner || r.Accepted != 0 || r.Attempts != 0 || r.Admits != 0 || r.Discards != 0 || r.Dropped != 0 || len(r.Phases) != 0 {
			t.Fatal("idle arm accidentally enabled observation")
		}
	}
}

func TestResearchIdleWrapperExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_IDLE_WRAPPER_ARTIFACT")
	if path == "" {
		t.Skip("opt-in idle wrapper diagnostic")
	}
	runResearchWrapperSourceLoad(t, path, true, true, true, true, true)
}
