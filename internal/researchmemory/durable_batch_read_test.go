package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type readProbeLog struct {
	*researchledger.Ledger
	gets, batches int
	corrupt       string
}

func (l *readProbeLog) Get(ctx context.Context, key researchledger.Key, kind string) (researchledger.Entry, error) {
	l.gets++
	return l.Ledger.Get(ctx, key, kind)
}
func (l *readProbeLog) GetBatch(ctx context.Context, requests []researchledger.LookupRequest) ([]researchledger.LookupResult, error) {
	l.batches++
	r, err := l.Ledger.GetBatch(ctx, requests)
	if err != nil {
		return nil, err
	}
	switch l.corrupt {
	case "count":
		return r[:len(r)-1], nil
	case "key":
		r[0].Entry.Key.Tenant = "other"
	case "kind":
		r[0].Entry.Kind = "other"
	case "missing-shape":
		r[0].Found = false
		r[0].Entry.Sequence = 1
	case "payload":
		r[0].Found = true
		r[0].Entry.Sequence = 1
		r[0].Entry.Payload = json.RawMessage(`{`)
	}
	return r, nil
}

func TestDurableSnapshotReadParity(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	owners := make([]*Durable, 2)
	for i := range owners {
		d, err := OpenDurable(ctx, t.TempDir()+"/reads.sqlite", "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		owners[i] = d
		defer d.Close()
		for id := uint64(1); id <= 32; id++ {
			if _, _, err = d.Admit(ctx, id, uint16(id), .6, now); err != nil {
				t.Fatal(err)
			}
			if _, err = d.Feedback(ctx, id, id%3 == 0, now); err != nil {
				t.Fatal(err)
			}
		}
		wait, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = d.worker.WaitProcessed(wait, 32)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
	requests := make([]AdmissionRequest, 64)
	ids := make([]uint64, 64)
	discards := make([]DiscardRequest, 64)
	for i := range requests {
		ids[i] = uint64(i + 33)
		requests[i] = AdmissionRequest{ID: ids[i], Features: uint16(i), Baseline: .6, At: now}
		discards[i] = DiscardRequest{ids[i], now}
	}
	want, err := owners[0].AdmitBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	probe := &readProbeLog{Ledger: owners[1].log.(*researchledger.Ledger)}
	owners[1].log = probe
	got, err := owners[1].AdmitBatchWithSnapshotReads(ctx, requests)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("warm admission parity", err)
	}
	stored, err := owners[1].Admissions(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range stored {
		if !reflect.DeepEqual(r, got[i].Record) {
			t.Fatal("readback mismatch", i)
		}
	}
	a, err := owners[0].DiscardBatch(ctx, discards)
	if err != nil {
		t.Fatal(err)
	}
	b, err := owners[1].DiscardBatchWithSnapshotReads(ctx, discards)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("discard parity", err)
	}
	if probe.gets != 0 || probe.batches != 3 {
		t.Fatal("reads not amortized", probe.gets, probe.batches)
	}
	if _, err = owners[1].Admissions(ctx, []uint64{999}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing record not rejected", err)
	}
}

func TestDurableSnapshotReadMalformedResponse(t *testing.T) {
	for _, kind := range []string{"count", "key", "kind", "missing-shape", "payload", "unsupported"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			d, err := OpenDurable(ctx, t.TempDir()+"/bad.sqlite", "tenant", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			if kind == "unsupported" {
				// Keep batch writes available so only the missing read capability rejects.
				d.log = faultyBatchLog{d.log, false, false}
			} else {
				d.log = &readProbeLog{Ledger: d.log.(*researchledger.Ledger), corrupt: kind}
			}
			if r, err := d.AdmitBatchWithSnapshotReads(ctx, []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}}); err == nil || r != nil {
				t.Fatal("malformed response accepted")
			}
			if d.stopped || d.worker.next != 0 || len(d.worker.pending) != 0 {
				t.Fatal("invalid read mutated worker")
			}
		})
	}
}

type failingReadBatchLog struct{ faultyBatchLog }

func (l failingReadBatchLog) GetBatch(ctx context.Context, requests []researchledger.LookupRequest) ([]researchledger.LookupResult, error) {
	return l.durableLog.(batchReadDurableLog).GetBatch(ctx, requests)
}

func TestDurableSnapshotReadCommitRecovery(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		open := OpenDurable
		if prepared {
			open = OpenDurablePreparedBatches
		}
		for _, kind := range []string{"admit", "discard"} {
			for _, commit := range []bool{false, true} {
				t.Run(fmt.Sprintf("prepared=%t/%s/commit=%t", prepared, kind, commit), func(t *testing.T) {
					ctx := context.Background()
					now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
					path := t.TempDir() + "/failure.sqlite"
					d, err := open(ctx, path, "tenant", "stream", 1, 42)
					if err != nil {
						t.Fatal(err)
					}
					requests := []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}}
					if kind == "discard" {
						if _, err = d.AdmitBatch(ctx, requests); err != nil {
							t.Fatal(err)
						}
					}
					d.log = failingReadBatchLog{faultyBatchLog{d.log, commit, false}}
					if kind == "admit" {
						_, err = d.AdmitBatchWithSnapshotReads(ctx, requests)
					} else {
						_, err = d.DiscardBatchWithSnapshotReads(ctx, []DiscardRequest{{1, now}})
					}
					if !errors.Is(err, errInjectedWrite) || !d.stopped {
						t.Fatal("commit uncertainty escaped", err)
					}
					if err = d.Close(); err != nil {
						t.Fatal(err)
					}
					d, err = open(ctx, path, "tenant", "stream", 1, 42)
					if err != nil {
						t.Fatal(err)
					}
					defer d.Close()
					if kind == "admit" {
						r, err := d.AdmitBatchWithSnapshotReads(ctx, requests)
						if err != nil || r[0].Retry != commit {
							t.Fatal("admit recovery", err)
						}
					} else {
						r, err := d.DiscardBatchWithSnapshotReads(ctx, []DiscardRequest{{1, now}})
						if err != nil || r[0] != commit {
							t.Fatal("discard recovery", err)
						}
					}
				})
			}
		}
	}
}
