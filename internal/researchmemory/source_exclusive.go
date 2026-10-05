package researchmemory

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

// OpenSourceOwnerExclusive retains the resolved-source prepared SQL strategy.
// It excludes external DB readers and is never selected by daemon defaults.
func OpenSourceOwnerExclusive(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	o, e := openSourceOwner(ctx, path, tenant, stream, epoch, seed, researchledger.OpenExclusiveResearch)
	if e != nil {
		return nil, e
	}
	o.batchReads = true
	o.reuseAdmissions = true
	return o, nil
}
