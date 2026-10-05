package researchmemory

import (
	"context"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type preparedBatchLedger struct{ *researchledger.Ledger }

func (l preparedBatchLedger) AppendBatch(ctx context.Context, requests []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	return l.Ledger.AppendBatchPrepared(ctx, requests)
}

// OpenDurablePreparedBatches opts into transaction-scoped statement reuse for
// batch writes only. Replay, individual writes, ownership and acknowledgment
// semantics are unchanged. This is a research owner, not a production default.
func OpenDurablePreparedBatches(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*Durable, error) {
	d, err := OpenDurable(ctx, path, tenant, stream, epoch, seed)
	if err != nil {
		return nil, err
	}
	// Replace the private log before exposing this owner to any caller. Its
	// background worker uses learner state, not this durable log interface.
	d.log = preparedBatchLedger{d.log.(*researchledger.Ledger)}
	return d, nil
}
