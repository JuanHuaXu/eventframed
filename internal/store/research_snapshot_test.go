package store

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"testing"
	"time"
)

func TestResearchTemporalCompatibility(t *testing.T) {
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	a := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	b := a
	b.RuntimeVersion++
	b.EvidenceEpoch++
	motion := map[uint64]time.Time{2: now.Add(time.Second)}
	if !ResearchSnapshotCompatible(a, b, now, motion) {
		t.Fatal("future insertion rejected")
	}
	for _, mutate := range []func(*model.Snapshot){
		func(s *model.Snapshot) { s.GraphVersion++ }, func(s *model.Snapshot) { s.PosteriorVersion++ },
		func(s *model.Snapshot) { s.PolicyVersion++ }, func(s *model.Snapshot) { s.ResidualVersion++ },
		func(s *model.Snapshot) { s.AbstractionVersion++ }, func(s *model.Snapshot) { s.AgencyVersion++ },
		func(s *model.Snapshot) { s.ContractVersion++ }, func(s *model.Snapshot) { s.EvidenceEpoch++ },
	} {
		c := b
		mutate(&c)
		if ResearchSnapshotCompatible(a, c, now, motion) {
			t.Fatal("semantic motion accepted")
		}
	}
	if ResearchSnapshotCompatible(a, b, now, nil) || ResearchSnapshotCompatible(a, b, time.Time{}, motion) {
		t.Fatal("unaccounted motion accepted")
	}
	motion[2] = now
	if ResearchSnapshotCompatible(a, b, now, motion) {
		t.Fatal("backfill accepted")
	}
	a.RuntimeVersion = ^uint64(0)
	if !ResearchSnapshotCompatible(a, a, now, nil) {
		t.Fatal("exact snapshot rejected")
	}
	if ResearchSnapshotCompatible(b, a, now, motion) {
		t.Fatal("overflow boundary accepted")
	}
}
