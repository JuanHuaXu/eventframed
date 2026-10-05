package researchmemory

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDurableCompletionObservation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := OpenDurable(ctx, t.TempDir()+"/completion.sqlite", "tenant-a", "completion", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := d.WaitProcessed(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if completed, failed, pending, queued := d.Counts(); completed != 0 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatal("unexpected initial progress")
	}
	closedCtx, closeCtx := context.WithCancel(ctx)
	closeCtx()
	if err := d.WaitProcessed(closedCtx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled wait did not stop", err)
	}
	queryAt := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	if _, retry, err := d.Admit(ctx, 1, 3, .6, queryAt); err != nil || retry {
		t.Fatal("admission", retry, err)
	}
	if _, err := d.Feedback(ctx, 1, true, queryAt.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := d.WaitProcessed(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if completed, failed, pending, queued := d.Counts(); completed != 1 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatalf("wrong live progress: completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := d.WaitProcessed(ctx, 2); err == nil {
		t.Fatal("closed worker waited for an impossible completion")
	}
}
