package researchpublication

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestCommittedSnapshotAgreement(t *testing.T) {
	base := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	p := New(base)
	if !p.CommittedSnapshotMatches(base) {
		t.Fatal("initial agreement rejected")
	}
	next := base
	next.RuntimeVersion++
	next.EvidenceEpoch++
	if p.CommittedSnapshotMatches(next) {
		t.Fatal("mismatch accepted")
	}
	ticket, err := p.Begin(Ingestion, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if p.CommittedSnapshotMatches(base) {
		t.Fatal("pending state accepted")
	}
	if err = p.Commit(ticket, next); err != nil {
		t.Fatal(err)
	}
	if !p.CommittedSnapshotMatches(next) || p.CommittedSnapshotMatches(base) {
		t.Fatal("commit agreement incorrect")
	}
	ticket, err = p.Begin(General, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Commit(ticket, next); err == nil {
		t.Fatal("invalid transition accepted")
	}
	if p.CommittedSnapshotMatches(next) {
		t.Fatal("quarantine accepted")
	}
}
