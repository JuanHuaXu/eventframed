package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

type researchLoadSchedule struct{ ReadEvery, WriteEvery time.Duration }
type researchScheduledCall struct {
	Index                 int
	DueNS, StartNS, EndNS int64
}

// Offered work stays in four fixed read lanes and one writer lane. Late calls
// keep their original due time; no unbounded goroutine fan-out or dropped due
// times can conceal overload. Context cancellation joins the normal fixture exit.
func researchAwaitDue(ctx context.Context, due time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if wait := time.Until(due); wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	return ctx.Err()
}

func researchCheckOffered(r researchGuardLoadResult, writes int) error {
	for _, stream := range []struct {
		samples []researchScheduledCall
		calls   []int64
		count   int
		every   int64
	}{
		{r.ReadSchedule, r.ReadNS, r.Requests, r.OfferedReadEveryNS},
		{r.WriteSchedule, r.WriteNS, writes, r.OfferedWriteEveryNS},
	} {
		if stream.every <= 0 || len(stream.samples) != stream.count || len(stream.calls) != stream.count {
			return errors.New("offered sample count")
		}
		seen := make(map[int]bool, stream.count)
		for i, v := range stream.samples {
			if v.Index < 0 || v.Index >= stream.count || seen[v.Index] || v.DueNS != int64(v.Index)*stream.every || v.StartNS < v.DueNS || v.EndNS < v.StartNS || v.EndNS-v.StartNS != stream.calls[i] {
				return errors.New("offered timing or identity mismatch")
			}
			seen[v.Index] = true
		}
	}
	return nil
}

func TestResearchOfferedTimingChecks(t *testing.T) {
	r := researchGuardLoadResult{Requests: 1, OfferedReadEveryNS: 10, OfferedWriteEveryNS: 20, ReadNS: []int64{5}, ReadSchedule: []researchScheduledCall{{0, 0, 3, 8}}}
	if err := researchCheckOffered(r, 0); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []researchScheduledCall{{1, 0, 3, 8}, {0, 1, 3, 8}, {0, 0, -1, 4}, {0, 0, 8, 3}, {0, 0, 3, 9}} {
		r.ReadSchedule = []researchScheduledCall{bad}
		if err := researchCheckOffered(r, 0); err == nil {
			t.Fatal("bad timing accepted", bad)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := researchAwaitDue(ctx, time.Now().Add(time.Hour)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestResearchOfferedLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		for arm := 0; arm < 3; arm++ {
			mode := "group4postverify"
			if arm == 0 {
				mode = "off"
			}
			schedule := &researchLoadSchedule{2 * time.Millisecond, 4 * time.Millisecond}
			r := researchGuardLoadScheduledArm(t, mode, 0, 8, writes, true, arm > 0, arm > 0, arm == 2, arm == 2, arm == 2, arm == 2, schedule)
			checkQueuedGuardLoad(t, r, writes)
			if err := researchCheckOffered(r, writes); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestResearchOfferedLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_OFFERED_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in matched offered-load diagnostic")
	}
	runResearchOfferedSourceLoad(t, path, true, true, true, true)
}
