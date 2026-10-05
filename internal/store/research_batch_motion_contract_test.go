package store

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestResearchBatchMotionRequiresEveryAcceptedVersion(t *testing.T) {
	asOf := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	captured := model.Snapshot{RuntimeVersion: 10, EvidenceEpoch: 10}
	committed := model.Snapshot{RuntimeVersion: 12, EvidenceEpoch: 12}
	motion := map[uint64]time.Time{11: asOf.Add(time.Second), 12: asOf.Add(time.Second)}
	if !ResearchSnapshotCompatible(captured, committed, asOf, motion) {
		t.Fatal("complete future-only batch motion rejected")
	}
	if ResearchSnapshotCompatible(captured, committed, asOf.Add(time.Second), motion) {
		t.Fatal("visible batch motion accepted")
	}
	delete(motion, 11)
	if ResearchSnapshotCompatible(captured, committed, asOf, motion) {
		t.Fatal("final-version-only motion accepted")
	}
	motion[11] = asOf.Add(time.Second)
	committed.EvidenceEpoch++
	if ResearchSnapshotCompatible(captured, committed, asOf, motion) {
		t.Fatal("unmatched evidence epoch accepted")
	}
}
