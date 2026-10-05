package researchmemory

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type conditionalBatchLedger struct{ *researchledger.Ledger }

func (l conditionalBatchLedger) AppendBatch(ctx context.Context, r []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	return l.Ledger.AppendBatchConditionalInsert(ctx, r)
}

// OpenSourceOwnerConditionalInsert changes only the opt-in batch SQL strategy.
// Actual forecasting, source resolution, guards and acknowledgment stay intact.
func OpenSourceOwnerConditionalInsert(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	o, err := OpenSourceOwnerResolvedAdmissions(ctx, path, tenant, stream, epoch, seed)
	if err != nil {
		return nil, err
	}
	o.d.log = conditionalBatchLedger{o.log}
	return o, nil
}
