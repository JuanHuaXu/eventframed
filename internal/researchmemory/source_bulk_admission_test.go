package researchmemory

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"testing"
)

type bulkFaultLog struct {
	bulkBatchLedger
	commit bool
}

func (l bulkFaultLog) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	if l.commit {
		if _, e := l.bulkBatchLedger.AppendBatch(ctx, r); e != nil {
			return nil, e
		}
	}
	return nil, errors.New("uncertain bulk acknowledgment")
}

func TestBulkOwnerRecovery(t *testing.T) {
	for _, commit := range []bool{false, true} {
		ctx := context.Background()
		path := t.TempDir() + "/recovery.sqlite"
		o, e := OpenSourceOwnerBulkAdmissions(ctx, path, "tenant", "stream", 1, 42)
		if e != nil {
			t.Fatal(e)
		}
		o.d.log = bulkFaultLog{o.d.log.(bulkBatchLedger), commit}
		r := []SourceAdmissionRequest{sourceRequest()}
		ack, e := o.Admit(ctx, r)
		if e == nil || ack != nil || !o.d.stopped {
			t.Fatal("uncertain owner continued", ack, e)
		}
		if _, e = o.Admit(ctx, r); e == nil {
			t.Fatal("stopped owner resumed")
		}
		if e = o.Close(); e != nil {
			t.Fatal(e)
		}
		o, e = OpenSourceOwnerBulkAdmissions(ctx, path, "tenant", "stream", 1, 42)
		if e != nil {
			t.Fatal(e)
		}
		ack, e = o.Admit(ctx, r)
		if e != nil || len(ack) != 1 || ack[0].Retry != commit || ack[0].Record.Prediction.ID != 1 {
			t.Fatal("bulk recovery", ack, e)
		}
		if e = o.Close(); e != nil {
			t.Fatal(e)
		}
	}
}
