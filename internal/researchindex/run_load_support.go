package researchindex

import "context"

type RunWriterStatus struct {
	Revision    uint64
	Delta, Runs int
}

func (w *DurableRunWriter) Status(ctx context.Context) (RunWriterStatus, error) {
	if err := w.lock(ctx); err != nil {
		return RunWriterStatus{}, err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return RunWriterStatus{}, ErrRecoveryRequired
	}
	return RunWriterStatus{w.current.revision, len(w.core.current.Load().delta), len(w.runs)}, nil
}
func (l *RunLease) Revision() uint64 { l.mu.Lock(); defer l.mu.Unlock(); return l.snapshot.revision }

// CloseRunsAfterReaders is for isolated harness shutdown only. The caller must
// stop/join writes and builders and successfully drain the associated lease pool
// first. It invalidates new writer operations, then closes retained current runs.
func (w *DurableRunWriter) CloseRunsAfterReaders(ctx context.Context) error {
	if err := w.lock(ctx); err != nil {
		return err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.core.compacting.Load() {
		return ErrCompactionBusy
	}
	w.failed = true
	for _, r := range w.runs {
		if err := r.Close(); err != nil {
			return err
		}
	}
	return nil
}
