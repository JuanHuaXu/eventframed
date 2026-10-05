package service

import (
	"os"
	"testing"
	"time"
)

func checkResearchGuardOnly(t *testing.T, r researchGuardLoadResult) {
	t.Helper()
	if r.Mode != "group4" || !r.UnionReads || r.SourceOwner || r.PreparedWrites || r.Admits != 0 || r.Discards != 0 || r.Accepted == 0 || r.Validated != 50*r.Accepted {
		t.Fatal("guard-only contract mismatch")
	}
	for _, p := range r.Phases {
		if p.DurableAdmitNS != 0 || p.DurableVerifyNS != 0 || p.DurableDiscardNS != 0 || p.VerifiedDiscardNS != 0 || p.PostGuardNS != 0 || (p.Accepted && (!p.Entered || p.CallbackNS <= 0)) {
			t.Fatal("guard-only phase mismatch")
		}
	}
}

func TestResearchGuardOnlyAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadScheduledArm(t, "group4", 0, 8, writes, true, false, true, false, false, false, false, &researchLoadSchedule{5 * time.Millisecond, 10 * time.Millisecond})
		checkQueuedGuardLoad(t, r, writes)
		if err := researchCheckOffered(r, writes); err != nil {
			t.Fatal(err)
		}
		checkResearchGuardOnly(t, r)
	}
}

func TestResearchGuardOnlyExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_GUARD_ONLY_ARTIFACT")
	if path == "" {
		t.Skip("opt-in guarded validation only")
	}
	runResearchGuardOnlySourceLoad(t, path, true, true, true, true, true, true)
}
