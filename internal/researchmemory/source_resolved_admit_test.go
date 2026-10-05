package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func openResolvedSourceTest(t *testing.T, path string) *SourceOwner {
	t.Helper()
	o, err := OpenSourceOwnerResolvedAdmissions(context.Background(), path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { o.Close() })
	return o
}

type admissionReadCountLog struct {
	preparedBatchLedger
	reads, appends int
}

func (l *admissionReadCountLog) Get(ctx context.Context, key researchledger.Key, kind string) (researchledger.Entry, error) {
	l.reads++
	return l.preparedBatchLedger.Get(ctx, key, kind)
}

func (l *admissionReadCountLog) GetBatch(ctx context.Context, r []researchledger.LookupRequest) ([]researchledger.LookupResult, error) {
	l.reads++
	return l.preparedBatchLedger.GetBatch(ctx, r)
}

func (l *admissionReadCountLog) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	l.appends++
	return l.preparedBatchLedger.AppendBatch(ctx, r)
}

func TestResolvedSourceAdmissionReadsAndRetries(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		t.Run(fmt.Sprint(reuse), func(t *testing.T) {
			ctx := context.Background()
			path := t.TempDir() + "/source.sqlite"
			open := openBatchSourceTest
			if reuse {
				open = openResolvedSourceTest
			}
			o := open(t, path)
			log := &admissionReadCountLog{preparedBatchLedger: o.d.log.(preparedBatchLedger)}
			o.d.log = log
			a, b := sourceRequest(), sourceRequest()
			b.Binding.EventID = "second"
			first, err := o.Admit(ctx, []SourceAdmissionRequest{a})
			if err != nil {
				t.Fatal(err)
			}
			// Equal instants are retries even with a different caller encoding.
			a.At = a.At.In(time.FixedZone("caller", 7200))
			mixed, err := o.Admit(ctx, []SourceAdmissionRequest{b, a})
			if err != nil || len(mixed) != 2 || mixed[0].Retry || !mixed[1].Retry || mixed[0].Record.Prediction.ID != 2 || !reflect.DeepEqual(mixed[1].Record, first[0].Record) {
				t.Fatal(mixed, err)
			}
			wantReads := 2
			if reuse {
				wantReads = 0
			}
			if log.reads != wantReads || log.appends != 2 {
				t.Fatal("preflight/atomic append counts", log.reads, log.appends)
			}
			if err = o.Close(); err != nil {
				t.Fatal(err)
			}
			o = openResolvedSourceTest(t, path)
			got, err := o.Admit(ctx, []SourceAdmissionRequest{b, a})
			if err != nil || len(got) != 2 || !got[0].Retry || !got[1].Retry {
				t.Fatal(got, err)
			}
			for i := range got {
				if !reflect.DeepEqual(got[i].Record, mixed[i].Record) {
					t.Fatal("reopen changed original")
				}
			}
		})
	}
}

func TestResolvedSourceAdmissionPreflight(t *testing.T) {
	for _, kind := range []string{"snapshot", "features", "baseline", "time", "tenant", "duplicate", "invalid", "cancel", "capacity"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			o := openResolvedSourceTest(t, t.TempDir()+"/source.sqlite")
			a := sourceRequest()
			if _, err := o.Admit(ctx, []SourceAdmissionRequest{a}); err != nil {
				t.Fatal(err)
			}
			fresh, bad := a, a
			fresh.Binding.EventID = "new"
			switch kind {
			case "snapshot":
				bad.Binding.Snapshot.PolicyVersion++
			case "features":
				bad.Features++
			case "baseline":
				bad.Baseline = .7
			case "time":
				bad.At = bad.At.Add(time.Nanosecond)
			case "tenant":
				bad.Binding.Tenant = "other"
			case "duplicate":
				bad = fresh
			case "invalid":
				bad.Binding.EventID = "invalid"
				bad.Baseline = math.NaN()
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "capacity":
				fill := make([]SourceAdmissionRequest, 255)
				for i := range fill {
					fill[i] = a
					fill[i].Binding.EventID = fmt.Sprint(i)
				}
				if _, err := o.Admit(ctx, fill); err != nil {
					t.Fatal(err)
				}
			}
			next := o.d.worker.next
			if got, err := o.Admit(ctx, []SourceAdmissionRequest{fresh, bad}); err == nil || got != nil {
				t.Fatal("bad batch accepted", got, err)
			}
			if o.d.worker.next != next || o.d.stopped {
				t.Fatal("preflight staged or stopped")
			}
			if _, err := o.Lookup(context.Background(), "journal", "new"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("partial original", err)
			}
		})
	}
}

func TestResolvedSourceAdmissionCommitRecovery(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, crash := range []bool{false, true} {
			t.Run(fmt.Sprintf("commit%t/panic%t", commit, crash), func(t *testing.T) {
				ctx := context.Background()
				path := t.TempDir() + "/source.sqlite"
				o := openResolvedSourceTest(t, path)
				req := []SourceAdmissionRequest{sourceRequest()}
				fault := sourceFaultLog{o.d.log.(preparedBatchLedger), commit}
				o.d.log = fault
				if crash {
					o.d.log = verifiedPanicLog{fault}
					func() {
						defer func() {
							if recover() == nil {
								t.Error("missing panic")
							}
						}()
						_, _ = o.Admit(ctx, req)
					}()
				} else if got, err := o.Admit(ctx, req); err == nil || got != nil {
					t.Fatal("uncertain commit acknowledged", got, err)
				}
				if _, err := o.Admit(ctx, req); err == nil {
					t.Fatal("stopped owner continued")
				}
				if err := o.Close(); err != nil {
					t.Fatal(err)
				}
				o = openResolvedSourceTest(t, path)
				got, err := o.Admit(ctx, req)
				if err != nil || len(got) != 1 || got[0].Retry != commit || got[0].Record.Prediction.ID != 1 {
					t.Fatal(got, err)
				}
			})
		}
	}
}

func TestResolvedSourceAdmissionConcurrentRetry(t *testing.T) {
	o := openResolvedSourceTest(t, t.TempDir()+"/source.sqlite")
	var wg sync.WaitGroup
	results := make(chan bool, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := o.Admit(context.Background(), []SourceAdmissionRequest{sourceRequest()})
			if err != nil || len(got) != 1 || got[0].Record.Prediction.ID != 1 {
				t.Error(got, err)
				return
			}
			results <- got[0].Retry
		}()
	}
	wg.Wait()
	close(results)
	newCount, retries := 0, 0
	for retry := range results {
		if retry {
			retries++
		} else {
			newCount++
		}
	}
	if newCount != 1 || retries != 7 {
		t.Fatal(newCount, retries)
	}
}

type unexpectedAdmissionLog struct{ preparedBatchLedger }

func (l unexpectedAdmissionLog) AppendBatch(ctx context.Context, requests []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	// Deliberately violate private ownership to exercise the last storage guard.
	// This is not a supported external writer or a source of trusted evidence.
	var other RecordedPrediction
	if err := json.Unmarshal(requests[0].Payload, &other); err != nil {
		return nil, err
	}
	other.Binding.EventID = "unexpected-source"
	raw, err := json.Marshal(other)
	if err != nil {
		return nil, err
	}
	if _, _, err = l.preparedBatchLedger.Append(ctx, requests[0].Key, "admit", raw); err != nil {
		return nil, err
	}
	return l.preparedBatchLedger.AppendBatch(ctx, requests)
}

func TestResolvedSourceAdmissionUnexpectedWrite(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/source.sqlite"
	o := openResolvedSourceTest(t, path)
	o.d.log = unexpectedAdmissionLog{o.d.log.(preparedBatchLedger)}
	a, b := sourceRequest(), sourceRequest()
	b.Binding.EventID = "second"
	requests := []SourceAdmissionRequest{a, b}
	if got, err := o.Admit(ctx, requests); err == nil || got != nil || !o.d.stopped {
		t.Fatal("unexpected write acknowledged", got, err)
	}
	rows, err := o.log.ReadAfter(ctx, 0, 128)
	if err != nil || len(rows) != 1 {
		t.Fatal("candidate batch partly persisted", rows, err)
	}
	if _, err = o.Lookup(ctx, "journal", "unexpected-source"); err == nil {
		t.Fatal("stopped owner served")
	}
	if err = o.Close(); err != nil {
		t.Fatal(err)
	}
	o = openResolvedSourceTest(t, path)
	got, err := o.Admit(ctx, requests)
	if err != nil || len(got) != 2 || got[0].Retry || got[1].Retry || got[0].Record.Prediction.ID != 2 || got[1].Record.Prediction.ID != 3 {
		t.Fatal("replay did not restore ID ownership", got, err)
	}
}
