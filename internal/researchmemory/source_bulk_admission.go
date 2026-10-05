package researchmemory

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type bulkBatchLedger struct{ *researchledger.Ledger }

func (l bulkBatchLedger) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	return l.Ledger.AppendBatchBulkAdmissions(ctx, r)
}

// OpenSourceOwnerBulkAdmissions opts into bounded SQL grouping only. Forecast
// creation, source identity, snapshot guards and durable acknowledgments stay.
func OpenSourceOwnerBulkAdmissions(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	o, e := OpenSourceOwnerResolvedAdmissions(ctx, path, tenant, stream, epoch, seed)
	if e != nil {
		return nil, e
	}
	o.d.log = bulkBatchLedger{o.log}
	return o, nil
}
