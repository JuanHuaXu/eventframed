package researchmemory

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func sourceDiscardFixture(t *testing.T) (*SourceOwner, []SourceAdmissionRequest, []SourceDiscardRequest, string) {
	t.Helper()
	path := t.TempDir() + "/source.sqlite"
	o := openSourceTest(t, path)
	a, b := sourceRequest(), sourceRequest()
	b.Binding.EventID = "second"
	requests := []SourceAdmissionRequest{a, b}
	if _, err := o.Admit(context.Background(), requests); err != nil {
		t.Fatal(err)
	}
	return o, requests, []SourceDiscardRequest{{a.Binding.JournalID, a.Binding.EventID, a.At}, {b.Binding.JournalID, b.Binding.EventID, b.At}}, path
}

func TestSourceDiscardBatchPreflight(t *testing.T) {
	for _, kind := range []string{"missing", "duplicate", "early", "zero", "changed-terminal", "cancel", "empty", "cap"} {
		t.Run(kind, func(t *testing.T) {
			o, requests, discard, _ := sourceDiscardFixture(t)
			ctx := context.Background()
			wantPending := 2
			switch kind {
			case "missing":
				discard[1].EventID = "unknown"
			case "duplicate":
				discard[1] = discard[0]
			case "early":
				discard[1].Available = discard[1].Available.Add(-time.Second)
			case "zero":
				discard[1].Available = time.Time{}
			case "changed-terminal":
				if _, err := o.Discard(ctx, discard[1].JournalID, discard[1].EventID, discard[1].Available); err != nil {
					t.Fatal(err)
				}
				wantPending = 1
				discard[1].Available = discard[1].Available.Add(time.Second)
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "empty":
				discard = nil
			case "cap":
				discard = make([]SourceDiscardRequest, 257)
			}
			if r, err := o.DiscardBatch(ctx, discard); err == nil || r != nil {
				t.Fatal("accepted bad batch", r, err)
			}
			if n, f, p, q := o.d.worker.Counts(); n != 0 || f != 0 || p != wantPending || q != 0 || o.d.stopped {
				t.Fatal(n, f, p, q, o.d.stopped)
			}
			good := SourceDiscardRequest{requests[0].Binding.JournalID, requests[0].Binding.EventID, requests[0].At}
			if r, err := o.DiscardBatch(context.Background(), []SourceDiscardRequest{good}); err != nil || r[0] {
				t.Fatal("partial first terminal", r, err)
			}
		})
	}
}

func TestSourceDiscardBatchMixedRetryAndReopen(t *testing.T) {
	o, requests, discard, path := sourceDiscardFixture(t)
	ctx := context.Background()
	original, err := o.Lookup(ctx, discard[0].JournalID, discard[0].EventID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = o.Discard(ctx, discard[0].JournalID, discard[0].EventID, discard[0].Available); err != nil {
		t.Fatal(err)
	}
	if r, err := o.DiscardBatch(ctx, discard); err != nil || !reflect.DeepEqual(r, []bool{true, false}) {
		t.Fatal(r, err)
	}
	if err = o.Close(); err != nil {
		t.Fatal(err)
	}
	o = openSourceTest(t, path)
	if r, err := o.DiscardBatch(ctx, discard); err != nil || !reflect.DeepEqual(r, []bool{true, true}) {
		t.Fatal(r, err)
	}
	if r, err := o.Admit(ctx, requests); err != nil || !r[0].Retry || !r[1].Retry || !reflect.DeepEqual(r[0].Record, original) {
		t.Fatal(r, err)
	}
	if n, f, p, q := o.d.worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 {
		t.Fatal(n, f, p, q)
	}
}

func TestSourceDiscardBatchUncertainCommit(t *testing.T) {
	for _, commit := range []bool{false, true} {
		t.Run(map[bool]string{false: "before", true: "after"}[commit], func(t *testing.T) {
			o, _, discard, path := sourceDiscardFixture(t)
			ctx := context.Background()
			o.d.log = sourceFaultLog{o.d.log.(preparedBatchLedger), commit}
			if r, err := o.DiscardBatch(ctx, discard); err == nil || r != nil {
				t.Fatal(r, err)
			}
			if _, err := o.DiscardBatch(ctx, discard); err == nil {
				t.Fatal("continued uncertain owner")
			}
			if err := o.Close(); err != nil {
				t.Fatal(err)
			}
			o = openSourceTest(t, path)
			if r, err := o.DiscardBatch(ctx, discard); err != nil || !reflect.DeepEqual(r, []bool{commit, commit}) {
				t.Fatal(r, err)
			}
		})
	}
}
