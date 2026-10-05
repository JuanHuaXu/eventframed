package researchmemory

import (
	"context"
	"errors"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type conditionalFaultLog struct {
	conditionalBatchLedger
	commit bool
}

func (l conditionalFaultLog) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	if l.commit {
		if _, err := l.conditionalBatchLedger.AppendBatch(ctx, r); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("injected uncertain acknowledgment")
}

func TestConditionalOwnerUncertainCommit(t *testing.T) {
	for _, commit := range []bool{false, true} {
		ctx := context.Background()
		path := t.TempDir() + "/uncertain.sqlite"
		o, err := OpenSourceOwnerConditionalInsert(ctx, path, "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		o.d.log = conditionalFaultLog{o.d.log.(conditionalBatchLedger), commit}
		req := sourceRequest()
		got, err := o.Admit(ctx, []SourceAdmissionRequest{req})
		if err == nil || got != nil || !o.d.stopped {
			t.Fatal("uncertain owner continued", got, err)
		}
		if _, err = o.Admit(ctx, []SourceAdmissionRequest{req}); err == nil {
			t.Fatal("stopped owner resumed")
		}
		if err = o.Close(); err != nil {
			t.Fatal(err)
		}
		o, err = OpenSourceOwnerConditionalInsert(ctx, path, "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		got, err = o.Admit(ctx, []SourceAdmissionRequest{req})
		if err != nil || len(got) != 1 || got[0].Retry != commit || got[0].Record.Prediction.ID != 1 {
			t.Fatal("recovery mismatch", got, err)
		}
		if err = o.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
