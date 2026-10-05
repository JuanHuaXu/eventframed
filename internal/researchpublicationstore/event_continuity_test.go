package researchpublicationstore

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchEventContinuityCoversRetentionAndComposition(t *testing.T) {
	ctx := context.Background()
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	put := func(id string, at time.Time) {
		t.Helper()
		if _, err := s.Put(ctx, testutil.Event(id, "public", at), []float32{1, 0, 0, 0}, id); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"a", "b"} {
		put(id, now.Add(-2*time.Minute))
	}
	from := s.Snapshot(ctx)
	if !s.writer.TryAcquire(1) {
		t.Fatal("unable to hold research mutation gate")
	}
	busy, busyErr := s.ResearchEventUnchanged(ctx, from, from, "tenant-a", "a")
	s.writer.Release(1)
	if busyErr == nil || busy {
		t.Fatal("busy mutation gate returned positive continuity")
	}
	if r, err := s.Delete(ctx, "tenant-a", "absent"); err != nil || r.Deleted || s.Snapshot(ctx) != from {
		t.Fatalf("no-op delete changed snapshot: %v", err)
	}
	for _, id := range []string{"a", "b"} {
		if same, err := s.ResearchEventUnchanged(ctx, from, s.Snapshot(ctx), "tenant-a", id); err != nil || !same {
			t.Fatalf("no-op delete touched %s: same=%v err=%v", id, same, err)
		}
	}
	retained, err := s.DeleteBefore(ctx, "tenant-a", now.Add(-time.Minute), 2)
	if err != nil || len(retained.DeletedIDs) != 2 {
		t.Fatalf("retention did not delete both sources: %+v err=%v", retained, err)
	}
	for _, id := range []string{"a", "b"} {
		if same, err := s.ResearchEventUnchanged(ctx, from, retained.Snapshot, "tenant-a", id); err != nil || same {
			t.Fatalf("retained deletion missed %s: same=%v err=%v", id, same, err)
		}
	}
	for _, id := range []string{"c", "d"} {
		put(id, now.Add(-time.Minute))
	}
	base := s.Snapshot(ctx)
	if _, err := s.PublishAntiPigeonCertificate(ctx, model.AntiPigeonCertificate{ID: "ap", TenantID: "tenant-a", MemberEventIDs: []string{"c", "d"}, GraphVersion: base.GraphVersion, EvidenceEpoch: base.EvidenceEpoch}); err != nil {
		t.Fatal(err)
	}
	base = s.Snapshot(ctx)
	macro := testutil.Event("macro", "public mission", now)
	macro.Kind = model.HigherOrderEventKind
	macro.Provenance.SourceEventIDs = []string{"c", "d"}
	macro.Composition = &model.Composition{MemberEventIDs: []string{"c", "d"}, RepresentativeEventID: "c", RuleID: "stages", Resolution: "mission", Confidence: .9, AntiPigeonCertificateID: "ap", EvidenceEpoch: base.EvidenceEpoch}
	if _, err := s.PutComposition(ctx, macro, []float32{1, 0, 0, 0}, "macro", base); err != nil {
		t.Fatal(err)
	}
	fromMacro := s.Snapshot(ctx)
	deleted, err := s.DeleteComposition(ctx, "tenant-a", "macro", "fixture", now.Add(time.Second))
	if err != nil || !deleted.Deleted {
		t.Fatalf("composition deletion failed: %+v err=%v", deleted, err)
	}
	if same, err := s.ResearchEventUnchanged(ctx, fromMacro, deleted.Snapshot, "tenant-a", "macro"); err != nil || same {
		t.Fatalf("deleted macro retained: same=%v err=%v", same, err)
	}
	if same, err := s.ResearchEventUnchanged(ctx, fromMacro, deleted.Snapshot, "tenant-a", "c"); err != nil || !same {
		t.Fatalf("untouched member revoked: same=%v err=%v", same, err)
	}
}

func TestResearchEventContinuityDuplicateQuarantinesWithoutTouch(t *testing.T) {
	ctx := context.Background()
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	event := testutil.Event("a", "public", now)
	if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, "a"); err != nil {
		t.Fatal(err)
	}
	from := s.Snapshot(ctx)
	if _, err := s.Put(ctx, event, []float32{1, 0, 0, 0}, "a"); err == nil {
		t.Fatal("duplicate ingestion unexpectedly retained publication authority")
	}
	if s.Snapshot(ctx) != from || s.eventTouches[eventTouchKey{"tenant-a", "a"}] != from.RuntimeVersion {
		t.Fatal("duplicate ingestion recorded a source mutation")
	}
	if same, err := s.ResearchEventUnchanged(ctx, from, from, "tenant-a", "a"); err == nil || same {
		t.Fatal("quarantined publication granted event continuity")
	}
}
