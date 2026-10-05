package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchadmission"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// Reuse both sealed components; only configure the scheduler before requests.
// Both arms use joined publication, isolating its interaction with scheduling.
func attachCombinedV26(t *testing.T, f *witnessFixtureV23, combined bool) *joinedWitnessV25 {
	t.Helper()
	s := attachJoinedV25(t, f, true)
	s.mode = combined
	return s
}

func TestResearchCombinedWitnessLifecycleV26(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_COMBINED_WITNESS_V26") != "1" {
		t.Skip("isolated combined lifecycle")
	}
	ctx := context.Background()
	for _, visible := range []bool{false, true} {
		t.Run(fmt.Sprint("visible-", visible), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachCombinedV26(t, f, true)
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			response, err := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "first", false)
			if err != nil {
				t.Fatal(err)
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite("combined-event", at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err := s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			after, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "after"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong transport", err)
			}
			p, err := s.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if err != nil || p.EvidenceEpoch != response.Posterior.EvidenceEpoch {
				t.Fatal("retagged source", err)
			}
			for _, alter := range []string{"query", "vector", "selection", "old"} {
				r := f.request(time.Now().UTC(), "changed")
				switch alter {
				case "query":
					r.Query = "different"
				case "vector":
					r.Embedding = denseRowV6(f.query, 11, .1)
				case "selection":
					r.PackK = 9
				case "old":
					r.AsOf = p.UpdatedAt.Add(-time.Nanosecond)
				}
				other, _, err := s.recall(ctx, f.svc, r)
				if err != nil || witnessBeliefsV23(other) != 0 {
					t.Fatal("request overreach", alter, err)
				}
			}
			f.reopen(t)
			s = attachCombinedV26(t, f, true)
			if s.poison.Load() {
				t.Fatal("valid reopen poisoned")
			}
			after, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong reopened transport", err)
			}
			if st := s.scheduler.Snapshot(); st.ActiveReaders != 0 || st.ActiveWriter || st.Queued != ([3]int{}) {
				t.Fatal("lease leak", st)
			}
		})
	}
	t.Run("runtime-gap", func(t *testing.T) {
		f := createWitnessFixtureV23(t, true)
		defer f.close()
		s := attachCombinedV26(t, f, true)
		w := pinnedWrite("unaccounted", f.origin.Add(time.Hour))
		w.Vector = denseRowV6(f.query, 3, .0001)
		if err := appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); !errors.Is(err, store.ErrStaleSnapshot) {
			t.Fatal("gap served", err)
		}
		f.reopen(t)
		s = attachCombinedV26(t, f, true)
		if !s.poison.Load() {
			t.Fatal("gap repaired")
		}
	})
}

func TestResearchCombinedWitnessCancellationV26(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_COMBINED_WITNESS_V26") != "1" {
		t.Skip("isolated combined cancellation")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachCombinedV26(t, f, true)
	entered, release := make(chan struct{}), make(chan struct{})
	s.barrier = func() { close(entered); <-release }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "cancel")); done <- err }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("no handoff")
	}
	cancel()
	select {
	case err := <-done:
		close(release)
		t.Fatal("early terminal", err)
	case <-time.After(10 * time.Millisecond):
	}
	if st := s.scheduler.Snapshot(); st.ActiveReaders != 1 || st.ActiveWriter {
		close(release)
		t.Fatal("early release", st)
	}
	close(release)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handoff hung")
	}
	if st := s.scheduler.Snapshot(); st.ActiveReaders != 0 || st.ActiveWriter || st.Queued != ([3]int{}) {
		t.Fatal("lease leak", st)
	}
	if len(s.state.Journals) != 1 {
		t.Fatal("binding lost")
	}
	f.reopen(t)
	s = attachCombinedV26(t, f, true)
	if s.poison.Load() || len(s.state.Journals) != 1 {
		t.Fatal("reopened handoff lost")
	}
	// Queued cancellation must not submit a journal or release the live owner.
	lease, err := s.scheduler.Acquire(context.Background(), researchadmission.Outcome)
	if err != nil {
		t.Fatal(err)
	}
	c, cancelQueued := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelQueued()
	_, _, err = s.recall(c, f.svc, f.request(time.Now().UTC(), "queued"))
	if !errors.Is(err, context.DeadlineExceeded) {
		lease.Release()
		t.Fatal(err)
	}
	if len(s.state.Journals) != 1 || !s.scheduler.Snapshot().ActiveWriter {
		lease.Release()
		t.Fatal("queued cancellation changed state")
	}
	lease.Release()
}
