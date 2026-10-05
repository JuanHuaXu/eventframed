package libravdbstore

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
)

func TestResearchBatchArchiveRetryV33(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_ARCHIVE_V33") != "1" {
		t.Skip("isolated archival retry controls")
	}
	ctx := context.Background()
	t.Run("sealed-v31-rejects-identical-retry", func(t *testing.T) {
		f := createWitnessFixtureV23(t, true)
		defer f.close()
		s := attachArchiveV31(t, f)
		defer s.Close()
		r := f.request(time.Now().UTC(), "first")
		first, _, err := s.recall(ctx, f.svc, r)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "other")); err != nil {
			t.Fatal(err)
		}
		_, _, err = s.recall(ctx, f.svc, r)
		if !errors.Is(err, store.ErrJournalConflict) {
			t.Fatal("expected V31 capture-root conflict", err)
		}
		t.Logf("confirmed V31 retry conflict after journal-only publication; original journal=%s", first.BayesianShadow.JournalID)
	})
	t.Run("batch-retains-original-provenance", func(t *testing.T) {
		f := createWitnessFixtureV23(t, true)
		defer f.close()
		s := attachBatchArchiveV33(t, f, 128)
		defer s.Close()
		r := f.request(time.Now().UTC(), "first")
		first, _, err := s.recall(ctx, f.svc, r)
		if err != nil {
			t.Fatal(err)
		}
		var original string
		if err := s.gate.sidecar.QueryRow("SELECT root FROM archive_capture_v29 WHERE journal_id=?", first.BayesianShadow.JournalID).Scan(&original); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "other")); err != nil {
			t.Fatal(err)
		}
		retry, _, err := s.recall(ctx, f.svc, r)
		if err != nil || retry.BayesianShadow.JournalID != first.BayesianShadow.JournalID {
			t.Fatal("identical retry rejected/changed", err)
		}
		var after string
		_ = s.gate.sidecar.QueryRow("SELECT root FROM archive_capture_v29 WHERE journal_id=?", first.BayesianShadow.JournalID).Scan(&after)
		if original != after || len(s.state.Journals) != 2 {
			t.Fatal("retry rewrote provenance")
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		f.reopen(t)
		s = attachBatchArchiveV33(t, f, 128)
		defer s.Close()
		if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen")); err != nil {
			t.Fatal(err)
		}
		if _, err := s.gate.sidecar.Exec("UPDATE archive_capture_v29 SET root='forged' WHERE journal_id=?", first.BayesianShadow.JournalID); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.recall(ctx, f.svc, r); err == nil {
			t.Fatal("forged original capture root accepted")
		}
	})
}

func TestResearchBatchArchiveLifecycleV33(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_ARCHIVE_V33") != "1" {
		t.Skip("isolated bounded queue/cancel/close controls")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachBatchArchiveV33(t, f, 3)
	defer s.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	var first sync.Once
	s.barrier = func() { first.Do(func() { close(entered) }); <-release }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 3)
	base := time.Now().UTC().Add(-time.Millisecond)
	for i := 0; i < 3; i++ {
		go func(i int) {
			_, _, err := s.recall(ctx, f.svc, f.request(base.Add(time.Duration(i)*time.Nanosecond), "held"))
			done <- err
		}(i)
	}
	remaining := 3
	defer func() {
		unblock()
		for remaining > 0 {
			select {
			case <-done:
				remaining--
			case <-time.After(5 * time.Second):
				t.Error("live caller did not drain")
				return
			}
		}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not enter barrier")
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		s.mu.Lock()
		accepted := s.stats.Accepted
		s.mu.Unlock()
		if accepted == 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("queue not filled")
		}
		time.Sleep(time.Millisecond)
	}
	if _, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "overload")); !errors.Is(err, errArchiveCapacityV33) {
		t.Fatal("cap did not reject", err)
	}
	cancel()
	closing := make(chan error, 1)
	go func() { closing <- s.Close() }()
	select {
	case err := <-closing:
		t.Fatal("close returned before accepted durability", err)
	case <-time.After(5 * time.Millisecond):
	}
	select {
	case err := <-done:
		remaining--
		t.Fatal("canceled accepted handoff returned early", err)
	default:
	}
	unblock()
	for remaining > 0 {
		select {
		case err := <-done:
			remaining--
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("accepted cancellation did not finish")
		}
	}
	select {
	case err := <-closing:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("close did not finish")
	}
	s.mu.Lock()
	stats := s.stats
	s.mu.Unlock()
	if !stats.Closed || stats.Accepted != 3 || stats.Finished != 3 || stats.Active != 0 || stats.Peak > 3 || stats.Rejected != 1 || len(s.state.Journals) != 3 {
		t.Fatal("queue/drain conservation", stats)
	}
	if _, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "closed")); err == nil {
		t.Fatal("closed adapter accepted request")
	}
	t.Logf("capacity/reject/cancel/close stats=%+v", stats)
}

func TestResearchBatchArchiveInterruptionV33(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_ARCHIVE_V33") != "1" {
		t.Skip("isolated native/sidecar interruptions")
	}
	for _, stop := range []string{"after_db", "before_witness", "after_sqlite"} {
		t.Run(stop, func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachBatchArchiveV33(t, f, 128)
			defer s.Close()
			s.stopAt = stop
			if _, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "stop")); err == nil {
				t.Fatal("interruption acked")
			}
			if !s.poison.Load() || s.current.Load() != nil {
				t.Fatal("failure not blocked")
			}
			_ = s.Close()
			if stop == "after_sqlite" {
				f.reopen(t)
				s = attachBatchArchiveV33(t, f, 128)
				defer s.Close()
				if s.poison.Load() {
					t.Fatal("complete joint commit lost")
				}
				if _, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "reopen")); err != nil {
					t.Fatal(err)
				}
			} else {
				f.close()
				g, err := openDenseOutcomeGateV17(f.root)
				if err != nil {
					t.Fatal(err)
				}
				defer g.close()
				if _, ready := g.capture(context.Background()); ready {
					t.Fatal("incomplete publication became READY")
				}
			}
		})
	}
}
