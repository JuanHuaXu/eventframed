package researchmemory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func verifiedDiscardFixture(t *testing.T) (*SourceOwner, []VerifiedSourceDiscardRequest, string) {
	t.Helper()
	path := t.TempDir() + "/verified.sqlite"
	o := openBatchSourceTest(t, path)
	a, b := sourceRequest(), sourceRequest()
	b.Binding.EventID = "second"
	r, err := o.Admit(context.Background(), []SourceAdmissionRequest{a, b})
	if err != nil {
		t.Fatal(err)
	}
	return o, []VerifiedSourceDiscardRequest{{r[0].Record, a.At}, {r[1].Record, b.At}}, path
}

func TestVerifiedSourceDiscardPreflight(t *testing.T) {
	for _, kind := range []string{"missing", "duplicate", "nil-binding", "tenant", "id", "seed", "epoch", "probability", "inner", "snapshot", "time", "early", "zero", "changed-terminal", "cancel", "empty", "cap"} {
		t.Run(kind, func(t *testing.T) {
			o, requests, _ := verifiedDiscardFixture(t)
			ctx := context.Background()
			wantPending := 2
			bad := append([]VerifiedSourceDiscardRequest(nil), requests...)
			binding := *bad[1].Original.Binding
			bad[1].Original.Binding = &binding
			switch kind {
			case "missing":
				binding.EventID = "unknown"
			case "duplicate":
				bad[1] = bad[0]
			case "nil-binding":
				bad[1].Original.Binding = nil
			case "tenant":
				binding.Tenant = "other"
			case "id":
				bad[1].Original.Prediction.ID = 1
			case "seed":
				bad[1].Original.Seed++
			case "epoch":
				bad[1].Original.Prediction.Epoch++
			case "probability":
				bad[1].Original.Prediction.Probability = .7
				bad[1].Original.Outer[0] = .7
			case "inner":
				bad[1].Original.Inner[0] = .4
			case "snapshot":
				binding.Snapshot.PolicyVersion++
			case "time":
				bad[1].Original.At = bad[1].Original.At.Add(-time.Nanosecond)
			case "early":
				bad[1].Available = bad[1].Available.Add(-time.Second)
			case "zero":
				bad[1].Available = time.Time{}
			case "changed-terminal":
				if _, err := o.Discard(ctx, "journal", "second", bad[1].Available); err != nil {
					t.Fatal(err)
				}
				wantPending = 1
				bad[1].Available = bad[1].Available.Add(time.Second)
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "empty":
				bad = nil
			case "cap":
				bad = make([]VerifiedSourceDiscardRequest, 257)
			}
			if got, err := o.VerifyAndDiscardBatch(ctx, bad); err == nil || got != nil {
				t.Fatal("bad expected record accepted", got, err)
			}
			if n, f, p, q := o.d.worker.Counts(); n != 0 || f != 0 || p != wantPending || q != 0 || o.d.stopped {
				t.Fatal("partial cleanup", n, f, p, q, o.d.stopped)
			}
			if _, err := o.d.log.Get(context.Background(), o.d.key(1), "feedback"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("partial durable terminal", err)
			}
			if got, err := o.VerifyAndDiscardBatch(context.Background(), requests[:1]); err != nil || got[0] {
				t.Fatal("preflight damaged owner", got, err)
			}
		})
	}
}

func TestVerifiedSourceDiscardRetryReopenAndClock(t *testing.T) {
	o, requests, path := verifiedDiscardFixture(t)
	ctx := context.Background()
	if _, err := o.Discard(ctx, "journal", "event", requests[0].Available); err != nil {
		t.Fatal(err)
	}
	if got, err := o.VerifyAndDiscardBatch(ctx, requests); err != nil || !reflect.DeepEqual(got, []bool{true, false}) {
		t.Fatal(got, err)
	}
	if err := o.Close(); err != nil {
		t.Fatal(err)
	}
	o = openBatchSourceTest(t, path)
	if got, err := o.VerifyAndDiscardBatch(ctx, requests); err != nil || !reflect.DeepEqual(got, []bool{true, true}) {
		t.Fatal(got, err)
	}
	req := sourceRequest()
	req.Binding.EventID = "clock"
	req.At = time.Now()
	original, err := o.Admit(ctx, []SourceAdmissionRequest{req})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := o.VerifyAndDiscardBatch(ctx, []VerifiedSourceDiscardRequest{{original[0].Record, req.At}}); err != nil || got[0] {
		t.Fatal("monotonic clock mismatch", got, err)
	}
	clock := original[0].Record
	clock.At = clock.At.In(time.FixedZone("alternate", 7200))
	if got, err := o.VerifyAndDiscardBatch(ctx, []VerifiedSourceDiscardRequest{{clock, req.At.In(time.FixedZone("retry", 7200))}}); err != nil || !got[0] {
		t.Fatal("same instant rejected", got, err)
	}
	stored, err := o.Lookup(ctx, "journal", "clock")
	if err != nil || !stored.At.Equal(req.At) {
		t.Fatal("original changed", err)
	}
}

func TestVerifiedSourceDiscardPreservesLabel(t *testing.T) {
	o, requests, _ := verifiedDiscardFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Private test-only label, not authorization exposed by SourceOwner.
	if _, err := o.d.Feedback(ctx, 2, true, requests[1].Available); err != nil {
		t.Fatal(err)
	}
	if err := o.d.worker.WaitProcessed(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if got, err := o.VerifyAndDiscardBatch(ctx, requests); err == nil || got != nil {
		t.Fatal("label erased", got, err)
	}
	if n, f, p, q := o.d.worker.Counts(); n != 1 || f != 0 || p != 1 || q != 0 {
		t.Fatal(n, f, p, q)
	}
	if got, err := o.VerifyAndDiscardBatch(ctx, requests[:1]); err != nil || got[0] {
		t.Fatal("first record partially discarded", got, err)
	}
}

type verifiedPanicLog struct{ sourceFaultLog }

func (l verifiedPanicLog) AppendBatch(ctx context.Context, entries []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	_, _ = l.sourceFaultLog.AppendBatch(ctx, entries)
	panic("verified discard uncertainty")
}

func TestVerifiedSourceDiscardCommitRecovery(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, crash := range []bool{false, true} {
			t.Run(fmt.Sprintf("commit%t/panic%t", commit, crash), func(t *testing.T) {
				o, requests, path := verifiedDiscardFixture(t)
				ctx := context.Background()
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
						_, _ = o.VerifyAndDiscardBatch(ctx, requests)
					}()
				} else if got, err := o.VerifyAndDiscardBatch(ctx, requests); err == nil || got != nil {
					t.Fatal("uncertain commit acknowledged", got, err)
				}
				if _, err := o.VerifyAndDiscardBatch(ctx, requests); err == nil {
					t.Fatal("stopped owner continued")
				}
				if err := o.Close(); err != nil {
					t.Fatal(err)
				}
				o = openBatchSourceTest(t, path)
				if got, err := o.VerifyAndDiscardBatch(ctx, requests); err != nil || !reflect.DeepEqual(got, []bool{commit, commit}) {
					t.Fatal(got, err)
				}
			})
		}
	}
}

func TestVerifiedSourceDiscardConcurrentRetry(t *testing.T) {
	o, requests, _ := verifiedDiscardFixture(t)
	var wg sync.WaitGroup
	results := make(chan []bool, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := o.VerifyAndDiscardBatch(context.Background(), requests)
			if err != nil {
				t.Error(err)
				return
			}
			results <- got
		}()
	}
	wg.Wait()
	close(results)
	fresh, retries := 0, 0
	for got := range results {
		if reflect.DeepEqual(got, []bool{false, false}) {
			fresh++
		} else if reflect.DeepEqual(got, []bool{true, true}) {
			retries++
		} else {
			t.Fatal("partial receipt", got)
		}
	}
	if fresh != 1 || retries != 7 {
		t.Fatal(fresh, retries)
	}
}
