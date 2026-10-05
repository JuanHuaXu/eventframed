package researchmemory

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func sourceRequest() SourceAdmissionRequest {
	return SourceAdmissionRequest{Features: 3, Baseline: .6, At: time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), Binding: ServiceBinding{Tenant: "tenant", JournalID: "journal", EventID: "event", Snapshot: model.Snapshot{RuntimeVersion: 1, ContractVersion: 1}}}
}

func openSourceTest(t *testing.T, path string) *SourceOwner {
	t.Helper()
	o, err := OpenSourceOwner(context.Background(), path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { o.Close() })
	return o
}

func TestSourceOwnerOriginalRetryAndTerminal(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/source.sqlite"
	o := openSourceTest(t, path)
	req := sourceRequest()
	r, err := o.Admit(ctx, []SourceAdmissionRequest{req})
	if err != nil || len(r) != 1 || r[0].Retry || r[0].Record.Prediction.ID != 1 {
		t.Fatal(r, err)
	}
	original := r[0].Record
	binding := *original.Binding
	original.Binding = &binding
	r[0].Record.Binding.EventID = "mutated output"
	if retry, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err != nil || !retry[0].Retry || !reflect.DeepEqual(retry[0].Record, original) {
		t.Fatal(retry, err)
	}
	if retry, err := o.Discard(ctx, "journal", "event", req.At); err != nil || retry {
		t.Fatal(retry, err)
	}
	if err = o.Close(); err != nil {
		t.Fatal(err)
	}
	o = openSourceTest(t, path)
	if retry, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err != nil || !retry[0].Retry || !reflect.DeepEqual(retry[0].Record, original) {
		t.Fatal(retry, err)
	}
	if retry, err := o.Discard(ctx, "journal", "event", req.At); err != nil || !retry {
		t.Fatal(retry, err)
	}
	if n, f, p, q := o.d.worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 {
		t.Fatal(n, f, p, q)
	}
	a, b := req, req
	a.Binding.JournalID = "other"
	b.Binding.EventID = "other"
	r, err = o.Admit(ctx, []SourceAdmissionRequest{a, req, b})
	if err != nil || r[0].Record.Prediction.ID != 2 || r[1].Record.Prediction.ID != 1 || !r[1].Retry || r[2].Record.Prediction.ID != 3 {
		t.Fatal(r, err)
	}
}

func TestSourceOwnerPreflight(t *testing.T) {
	for _, kind := range []string{"snapshot", "features", "baseline", "time", "tenant", "duplicate", "invalid", "utf8", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			o := openSourceTest(t, t.TempDir()+"/source.sqlite")
			req := sourceRequest()
			if _, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err != nil {
				t.Fatal(err)
			}
			fresh, bad := req, req
			fresh.Binding.EventID = "new"
			switch kind {
			case "snapshot":
				bad.Binding.Snapshot.PolicyVersion++
			case "features":
				bad.Features++
			case "baseline":
				bad.Baseline = .7
			case "time":
				bad.At = bad.At.Add(time.Second)
			case "tenant":
				bad.Binding.Tenant = "other"
			case "duplicate":
				bad = fresh
			case "invalid":
				bad.Binding.EventID = "invalid"
				bad.Baseline = math.NaN()
			case "utf8":
				bad.Binding.EventID = string([]byte{255})
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if got, err := o.Admit(ctx, []SourceAdmissionRequest{fresh, bad}); err == nil || got != nil {
				t.Fatal("accepted invalid batch", got, err)
			}
			if o.d.worker.next != 1 || o.d.stopped {
				t.Fatal("preflight staged or stopped")
			}
			if _, err := o.Lookup(context.Background(), "journal", "new"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal(err)
			}
			if r, err := o.Admit(context.Background(), []SourceAdmissionRequest{fresh}); err != nil || r[0].Record.Prediction.ID != 2 {
				t.Fatal(r, err)
			}
		})
	}
}

func TestSourceOwnerConcurrentRetry(t *testing.T) {
	o := openSourceTest(t, t.TempDir()+"/source.sqlite")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := o.Admit(context.Background(), []SourceAdmissionRequest{sourceRequest()})
			if err != nil || len(r) != 1 || r[0].Record.Prediction.ID != 1 {
				t.Error(r, err)
			}
		}()
	}
	wg.Wait()
	rows, err := o.log.ReadAfter(context.Background(), 0, 128)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
}

func TestSourceOwnerRejectsLegacyHistory(t *testing.T) {
	for _, bound := range []bool{false, true} {
		t.Run(map[bool]string{false: "unbound", true: "duplicate"}[bound], func(t *testing.T) {
			ctx := context.Background()
			path := t.TempDir() + "/source.sqlite"
			req := sourceRequest()
			d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			for id := uint64(1); id <= 2; id++ {
				if bound {
					_, _, err = d.AdmitBound(ctx, id, req.Features, req.Baseline, req.At, req.Binding)
				} else {
					_, _, err = d.Admit(ctx, id, req.Features, req.Baseline, req.At)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = d.Close(); err != nil {
				t.Fatal(err)
			}
			if o, err := OpenSourceOwner(ctx, path, "tenant", "stream", 1, 42); err == nil {
				o.Close()
				t.Fatal("legacy history accepted")
			}
			log, err := researchledger.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer log.Close()
			rows, err := log.ReadAfter(ctx, 0, 128)
			if err != nil || len(rows) != 2 {
				t.Fatal(rows, err)
			}
			_, err = log.GetServiceAdmission(ctx, researchledger.ServiceIdentity{Tenant: "tenant", Stream: "stream", Contract: RecordContract, Journal: "journal", Event: "event"})
			if err == nil || errors.Is(err, sql.ErrNoRows) {
				t.Fatal("failed activation left index", err)
			}
		})
	}
}

type sourceFaultLog struct {
	preparedBatchLedger
	commit bool
}

func (l sourceFaultLog) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	if l.commit {
		if _, err := l.preparedBatchLedger.AppendBatch(ctx, r); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("uncertain source acknowledgment")
}

func TestSourceOwnerUncertainCommit(t *testing.T) {
	for _, commit := range []bool{false, true} {
		t.Run(map[bool]string{false: "before", true: "after"}[commit], func(t *testing.T) {
			ctx := context.Background()
			path := t.TempDir() + "/source.sqlite"
			o := openSourceTest(t, path)
			req := sourceRequest()
			o.d.log = sourceFaultLog{o.d.log.(preparedBatchLedger), commit}
			if r, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err == nil || r != nil {
				t.Fatal(r, err)
			}
			if _, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err == nil {
				t.Fatal("uncertain owner continued")
			}
			if _, err := o.Lookup(ctx, "journal", "event"); err == nil {
				t.Fatal("uncertain owner served")
			}
			if err := o.Close(); err != nil {
				t.Fatal(err)
			}
			o = openSourceTest(t, path)
			r, err := o.Admit(ctx, []SourceAdmissionRequest{req})
			if err != nil || r[0].Retry != commit || r[0].Record.Prediction.ID != 1 {
				t.Fatal(r, err)
			}
		})
	}
}
