package researchbatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchadmission"
)

func TestRunAdmittedDistinguishesEntry(t *testing.T) {
	g, _ := researchadmission.New(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := RunAdmitted(ctx, g.Write, func(context.Context) error { called = true; return nil })
	if called || !NeverEntered(err) || !errors.Is(err, context.Canceled) {
		t.Fatal("non-entry not recorded", err)
	}
	err = RunAdmitted(context.Background(), g.Write, func(context.Context) error { return context.DeadlineExceeded })
	if NeverEntered(err) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("entered error misclassified", err)
	}
	if NeverEntered(errors.New("batch admission did not enter: context canceled")) {
		t.Fatal("trusted text")
	}
}

// This test is enabled only in the phase-aware research overlay.
func TestPhaseAwareQueueContinuesAfterNonEntry(t *testing.T) {
	if !phaseAwareTestEnabled {
		t.Skip("requires phase-aware overlay")
	}
	g, _ := researchadmission.New(1)
	calls := 0
	q, _ := New(4, 1, time.Second, func(ctx context.Context, values []int) ([]int, error) {
		calls++
		if calls == 1 {
			c, cancel := context.WithCancel(ctx)
			cancel()
			return nil, RunAdmitted(c, g.Write, func(context.Context) error { t.Error("entered canceled admission"); return nil })
		}
		return values, nil
	})
	if _, err := q.Put(context.Background(), 1); !NeverEntered(err) {
		t.Fatal(err)
	}
	if value, err := q.Put(context.Background(), 2); err != nil || value != 2 {
		t.Fatal("safe rejection poisoned queue", value, err)
	}
	q.Close()
}

const phaseAwareTestEnabled = false
