package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// Cold previews are detached computation, not durable reservations. An abandoned
// preview must not force a gap in the durable stream. This test does not authorize
// relabeling IDs of an asynchronously trained model's original forecasts.
func TestResearchColdPreviewAfterAbandonment(t *testing.T) {
	_, _, in := bridgeFixture(t)
	c := in.Candidates[0]
	a := researchmemory.New(1, 42)
	first, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	a.Discard(first.ID)
	next, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := a.Record(next.ID)
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != 2 {
		t.Fatal("fixture did not abandon an ID")
	}
	preview.Prediction.ID = 1
	preview.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
	ctx := context.Background()
	path := t.TempDir() + "/cold-preview.sqlite"
	d, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "load", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if _, _, err = d.AdmitBound(ctx, 1, c.Features, c.Baseline, in.AsOf, *preview.Binding); err != nil {
		t.Fatal(err)
	}
	saved, err := d.Admission(ctx, 1)
	if err != nil || !reflect.DeepEqual(saved, preview) {
		t.Fatal("cold original mismatch", err)
	}
	if _, err = d.Discard(ctx, 1, in.AsOf); err != nil {
		t.Fatal(err)
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "load", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	saved, err = reopened.Admission(ctx, 1)
	if err != nil || !reflect.DeepEqual(saved, preview) {
		t.Fatal("reopen changed cold original", err)
	}
	if _, err = reopened.Feedback(ctx, 1, true, in.AsOf); err == nil {
		t.Fatal("discard became evidence")
	}
	if _, _, err = reopened.AdmitBound(ctx, 2, c.Features, c.Baseline, in.AsOf, *preview.Binding); err != nil {
		t.Fatal("stream failed contiguous continuation", err)
	}
}
