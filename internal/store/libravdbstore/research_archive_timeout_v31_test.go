package libravdbstore

import (
	"context"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
)

// V30 exposed queue wait being mischarged to the database-operation timeout.
// The total request is still timed from offer and has no new adoption slack.
type archiveWitnessV31 struct{ *archiveWitnessV29 }

func attachArchiveV31(t *testing.T, f *witnessFixtureV23) *archiveWitnessV31 {
	t.Helper()
	s := &archiveWitnessV31{attachArchiveV29(t, f)}
	before := s.gate.store.Snapshot(context.Background())
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != s.gate.store.Snapshot(context.Background()) {
		t.Fatal("revision retagged source")
	}
	f.svc = svc
	return s
}
func (s *archiveWitnessV31) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c, err := s.capture(ctx, e)
	if err != nil {
		return err
	}
	l := ctx.Value(archiveLeaseKeyV29{}).(*archiveLeaseV29)
	l.finish()
	if s.afterCapture != nil {
		s.afterCapture(c)
	}
	// Queue/owner wait is visible end-to-end, not inside a native-operation timer.
	// As in V29, accepted capture cancellation waits for terminal persistence.
	if err := s.commitV31(context.WithoutCancel(ctx), c); err != nil {
		return err
	}
	return ctx.Err()
}
