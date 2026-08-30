package store

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestJournalSnapshotCompatibilityAllowsOnlyIngestionMotion(t *testing.T) {
	captured := model.Snapshot{RuntimeVersion: 10, EvidenceEpoch: 10, PolicyVersion: 2, ContractVersion: 12, GraphVersion: 3, PosteriorVersion: 4, ResidualVersion: 5, AbstractionVersion: 6, AgencyVersion: 7}
	current := captured
	current.RuntimeVersion++
	current.EvidenceEpoch++
	if !JournalSnapshotCompatible(captured, current, time.Unix(100, 0), map[uint64]time.Time{11: time.Unix(101, 0)}) {
		t.Fatal("ingestion-only motion invalidated a fixed-as-of journal")
	}
	if JournalSnapshotCompatible(captured, current, time.Unix(100, 0), map[uint64]time.Time{11: time.Unix(99, 0)}) {
		t.Fatal("backfilled ingestion was treated as future-only motion")
	}
	if JournalSnapshotCompatible(captured, current, time.Unix(100, 0), nil) {
		t.Fatal("unclassified runtime motion was accepted")
	}
	for _, mutate := range []func(*model.Snapshot){
		func(value *model.Snapshot) { value.PolicyVersion++ },
		func(value *model.Snapshot) { value.ContractVersion++ },
		func(value *model.Snapshot) { value.GraphVersion++ },
		func(value *model.Snapshot) { value.PosteriorVersion++ },
		func(value *model.Snapshot) { value.ResidualVersion++ },
		func(value *model.Snapshot) { value.AbstractionVersion++ },
		func(value *model.Snapshot) { value.AgencyVersion++ },
	} {
		changed := captured
		mutate(&changed)
		if JournalSnapshotCompatible(captured, changed, time.Unix(100, 0), nil) {
			t.Fatalf("semantic snapshot motion was accepted: captured=%+v current=%+v", captured, changed)
		}
	}
}
