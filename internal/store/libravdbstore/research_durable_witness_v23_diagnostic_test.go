package libravdbstore

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// Post-hoc falsifier for the v23 collector predicate, not a change to its sealed
// tape or adoption gate. Private, json:"-" grouping keys are not wire evidence.
func TestResearchDurableWitnessWireDiagnosticV23(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_WITNESS_DIAGNOSTIC_V23") != "1" {
		t.Skip("isolated post-hoc journal boundary diagnostic")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	ctx := context.Background()
	packet, _, err := f.adapter.recallV23(ctx, f.svc, f.request(time.Now().UTC(), "diagnostic"))
	if err != nil {
		t.Fatal(err)
	}
	saved, err := f.adapter.gate.store.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
	if err != nil || saved.Snapshot != packet.Snapshot {
		t.Fatal("snapshot/identity mismatch", err)
	}
	got, err := json.Marshal(saved.Report.Decisions)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(packet.BayesianShadow.Decisions)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("wire mismatch", err)
	}
	if reflect.DeepEqual(saved.Report.Decisions, packet.BayesianShadow.Decisions) {
		t.Fatal("original predicate did not reproduce")
	}
	copyDecisions := append([]model.BayesianDecision(nil), packet.BayesianShadow.Decisions...)
	for i := range copyDecisions {
		copyDecisions[i].EvidenceGroupKey = ""
	}
	if !reflect.DeepEqual(saved.Report.Decisions, copyDecisions) {
		t.Fatal("difference extends beyond private grouping key")
	}
	t.Logf("sealed collector false positive reproduced: %d wire decisions match, private grouping keys are intentionally absent after decode", len(copyDecisions))
}
