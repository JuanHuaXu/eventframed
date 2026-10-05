package researchmemory

import (
	"context"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type corruptTerminalRead struct{ durableLog }

func (l corruptTerminalRead) ReadAfter(ctx context.Context, after int64, limit int) ([]researchledger.Entry, error) {
	rows, err := l.durableLog.ReadAfter(ctx, after, limit)
	for i := range rows {
		if rows[i].Kind == "feedback" {
			rows[i].Payload = []byte(`{}`)
		}
	}
	return rows, err
}

func TestDurableBoundLabelsUsesAllTerminalEvidenceByCutoff(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "labels.sqlite")
	d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	binding, digest, event, key := sourceWitnessFixture()
	now := event.AvailableAt
	for id := uint64(1); id <= 4; id++ {
		copyEvent := event
		copyBinding := binding
		copyEvent.ID = "event-" + strconv.FormatUint(id, 10)
		copyBinding.EventID = copyEvent.ID
		witness, err := NewSourceWitness("lab-v1", key, copyBinding, digest, copyEvent, 7, .6)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := d.AdmitBoundWithWitness(ctx, id, 7, .6, now, copyBinding, witness); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.Feedback(ctx, 1, true, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	before, sequence, err := d.BoundLabelsWithSequence(ctx, now.Add(2*time.Second))
	if err != nil || len(before) != 1 || before[0].Prediction.Prediction.ID != 1 {
		t.Fatalf("first cutoff read: labels=%d err=%v", len(before), err)
	}
	called := false
	if err := d.WithUnchangedSequence(ctx, sequence, func() error { called = true; return nil }); err != nil || !called {
		t.Fatalf("unchanged durable stream rejected: %v", err)
	}
	if _, err := d.Discard(ctx, 2, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Feedback(ctx, 3, false, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	called = false
	if err := d.WithUnchangedSequence(ctx, sequence, func() error { called = true; return nil }); err == nil || called {
		t.Fatal("new terminal failed to invalidate prepared sequence")
	}
	after, err := d.BoundLabels(ctx, now.Add(2*time.Second))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("future label or discard changed earlier cutoff: err=%v", err)
	}
	all, err := d.BoundLabels(ctx, now.Add(4*time.Second))
	if err != nil || len(all) != 2 || all[0].Prediction.Prediction.ID != 1 || all[1].Prediction.Prediction.ID != 3 || all[1].Feedback.Useful {
		t.Fatalf("complete cutoff omitted or converted a terminal: labels=%+v err=%v", all, err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	replayed, err := d.BoundLabels(ctx, now.Add(4*time.Second))
	if err != nil || !reflect.DeepEqual(all, replayed) {
		t.Fatalf("durable bound labels changed after reopen: err=%v", err)
	}
}

func TestDurableBoundLabelsRejectsUnwitnessedTerminal(t *testing.T) {
	ctx := context.Background()
	d, err := OpenDurable(ctx, filepath.Join(t.TempDir(), "labels.sqlite"), "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	binding, _, event, _ := sourceWitnessFixture()
	now := event.AvailableAt
	if _, _, err := d.AdmitBound(ctx, 1, 7, .6, now, binding); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Feedback(ctx, 1, true, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if labels, err := d.BoundLabels(ctx, now.Add(2*time.Second)); err == nil || labels != nil {
		t.Fatal("unwitnessed historical label became transferable")
	}
}

func TestDurableBoundLabelsRejectsMalformedTerminal(t *testing.T) {
	ctx := context.Background()
	d, err := OpenDurable(ctx, filepath.Join(t.TempDir(), "labels.sqlite"), "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	binding, digest, event, key := sourceWitnessFixture()
	witness, err := NewSourceWitness("lab-v1", key, binding, digest, event, 7, .6)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := d.AdmitBoundWithWitness(ctx, 1, 7, .6, event.AvailableAt, binding, witness); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Feedback(ctx, 1, false, event.AvailableAt.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	d.log = corruptTerminalRead{d.log}
	if labels, err := d.BoundLabels(ctx, event.AvailableAt.Add(2*time.Second)); err == nil || labels != nil {
		t.Fatal("malformed terminal became a negative or discard")
	}
}
