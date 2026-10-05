package researchmemory

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type faultyBatchLog struct {
	durableLog
	commit, panicInstead bool
}

func (l faultyBatchLog) AppendBatch(ctx context.Context, entries []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	if l.commit {
		if _, err := l.durableLog.(batchDurableLog).AppendBatch(ctx, entries); err != nil {
			return nil, err
		}
	}
	if l.panicInstead {
		panic("batch write uncertainty")
	}
	return nil, errInjectedWrite
}

func TestDurableBatchDiscardParity(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	path := t.TempDir() + "/batch.sqlite"
	d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	requests := make([]DiscardRequest, 256)
	originals := make([]RecordedPrediction, 256)
	for i := range requests {
		id := uint64(i + 1)
		if _, _, err = d.Admit(ctx, id, uint16(i), .6, now); err != nil {
			t.Fatal(err)
		}
		originals[i], err = d.Admission(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		requests[i] = DiscardRequest{id, now.Add(time.Second)}
	}
	clock := d.worker.lastAvailable
	retries, err := d.DiscardBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range retries {
		if r {
			t.Fatal("new discard reported retry")
		}
	}
	if n, f, p, q := d.worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 || d.worker.lastAvailable != clock {
		t.Fatal("discard changed learning", n, f, p, q)
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	retries, err = d.DiscardBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range retries {
		if !r {
			t.Fatal("reopened retry duplicated terminal")
		}
		saved, err := d.Admission(ctx, uint64(i+1))
		if err != nil || !reflect.DeepEqual(saved, originals[i]) {
			t.Fatal("original changed", i, err)
		}
	}
	if _, err = d.Feedback(ctx, 1, true, now.Add(time.Second)); err == nil {
		t.Fatal("discard accepted label")
	}
}

func TestDurableBatchDiscardFailureRecovery(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, panics := range []bool{false, true} {
			t.Run(fmt.Sprintf("commit=%t/panic=%t", commit, panics), func(t *testing.T) {
				ctx := context.Background()
				now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
				path := t.TempDir() + "/failure.sqlite"
				d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				requests := []DiscardRequest{{1, now}, {2, now}}
				for id := uint64(1); id <= 2; id++ {
					if _, _, err = d.Admit(ctx, id, 1, .6, now); err != nil {
						t.Fatal(err)
					}
				}
				d.log = faultyBatchLog{d.log, commit, panics}
				func() {
					defer func() {
						r := recover()
						if (r != nil) != panics {
							t.Error("panic mismatch", r)
						}
					}()
					var result []bool
					result, err = d.DiscardBatch(ctx, requests)
					if result != nil || !errors.Is(err, errInjectedWrite) {
						t.Error("unexpected failure", result, err)
					}
				}()
				if !d.stopped {
					t.Fatal("uncertain owner remained active")
				}
				if _, err = d.DiscardBatch(ctx, requests); err == nil {
					t.Fatal("continued after uncertainty")
				}
				if err = d.Close(); err != nil {
					t.Fatal(err)
				}
				d, err = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				defer d.Close()
				_, _, pending, _ := d.worker.Counts()
				want := 2
				if commit {
					want = 0
				}
				if pending != want {
					t.Fatal("wrong replay", pending, want)
				}
				retries, err := d.DiscardBatch(ctx, requests)
				if err != nil {
					t.Fatal(err)
				}
				for _, r := range retries {
					if r != commit {
						t.Fatal("wrong retry outcome")
					}
				}
			})
		}
	}
}

func TestDurableBatchDiscardPreflight(t *testing.T) {
	for _, kind := range []string{"empty", "count", "duplicate", "missing", "early", "labeled", "unsupported", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			d, err := OpenDurable(ctx, t.TempDir()+"/reject.sqlite", "tenant", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			for id := uint64(1); id <= 2; id++ {
				if _, _, err = d.Admit(ctx, id, 1, .6, now); err != nil {
					t.Fatal(err)
				}
			}
			requests := []DiscardRequest{{1, now}, {2, now}}
			switch kind {
			case "empty":
				requests = nil
			case "count":
				requests = make([]DiscardRequest, 257)
			case "duplicate":
				requests[1].ID = 1
			case "missing":
				requests[1].ID = 3
			case "early":
				requests[1].Available = now.Add(-time.Second)
			case "labeled":
				if _, err = d.Feedback(ctx, 1, true, now); err != nil {
					t.Fatal(err)
				}
			case "unsupported":
				d.log = singleOnlyLog{d.log}
			case "cancel":
				cancel()
			}
			if result, err := d.DiscardBatch(ctx, requests); err == nil || result != nil {
				t.Fatal("bad discard accepted", err)
			}
			if d.stopped {
				t.Fatal("preflight stopped unmodified owner")
			}
			if _, err = d.worker.Record(2); err != nil {
				t.Fatal("valid pending member discarded", err)
			}
		})
	}
}
