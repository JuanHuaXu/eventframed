package store

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"time"
)

// ResearchSnapshotCompatible reuses journal temporal semantics for read-only
// diagnostics. Callers lock snapshot and motion together. Unknown history fails
// closed; finite bounds also avoid wrapping the journal version iteration.
func ResearchSnapshotCompatible(captured, current model.Snapshot, asOf time.Time, motion map[uint64]time.Time) bool {
	if captured == current {
		return true
	}
	if asOf.IsZero() || current.RuntimeVersion <= captured.RuntimeVersion || current.RuntimeVersion == ^uint64(0) || current.RuntimeVersion-captured.RuntimeVersion > uint64(len(motion)) {
		return false
	}
	if current.EvidenceEpoch < captured.EvidenceEpoch || current.EvidenceEpoch-captured.EvidenceEpoch != current.RuntimeVersion-captured.RuntimeVersion {
		return false
	}
	return JournalSnapshotCompatible(captured, current, asOf, motion)
}
