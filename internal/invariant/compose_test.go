package invariant_test

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/invariant"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestComposeRetainsOnlySharedInvariantsAndTimeEnvelope(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	left := testutil.Event("left", "left stage", now.Add(-time.Hour))
	right := testutil.Event("right", "right stage", now.Add(-time.Minute))
	left.Who.Value, right.Who.Value = "shared team", "SHARED  team"
	left.Where.Value, right.Where.Value = "Earth", "Moon"
	request := model.ComposeInvariantRequest{ID: "macro", TenantID: "tenant-a", SessionID: "session-a", MemberEventIDs: []string{"left", "right"}, RepresentativeEventID: "left", Label: "mission", RuleID: "stages", Resolution: "mission", Confidence: .8, AntiPigeonCertificateID: "ap", PublishedAt: now, BaseSnapshot: model.Snapshot{EvidenceEpoch: 7}}
	composite := invariant.Compose(request, []model.Event{left, right})
	if composite.Who.Value != "shared team" || composite.Where.Value != "" {
		t.Fatalf("shared/non-shared fields = who:%+v where:%+v", composite.Who, composite.Where)
	}
	if composite.OccurredAt != left.OccurredAt || composite.ObservedAt != right.ObservedAt || composite.Composition.EvidenceEpoch != 7 {
		t.Fatalf("time/provenance envelope = %+v", composite)
	}
}
