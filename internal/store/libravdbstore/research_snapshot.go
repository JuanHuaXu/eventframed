package libravdbstore

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"time"
)

func (s *Store) ResearchSnapshotCompatible(ctx context.Context, captured model.Snapshot, asOf time.Time) bool {
	s.writeMu.RLock()
	defer s.writeMu.RUnlock()
	return ctx.Err() == nil && store.ResearchSnapshotCompatible(captured, s.snapshot, asOf, s.ingestMotion)
}
