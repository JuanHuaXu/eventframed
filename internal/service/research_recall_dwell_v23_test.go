package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func runResearchRecallDwellTrialV23(t *testing.T, writer bool, wait time.Duration) recallProfileTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	s, tap := motionExitService(t, root, true)
	defer s.Close()
	defer tap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	expected := make(map[string]bool, 200)
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	expected["seed"] = true
	for i := 1; i < 200; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	path := filepath.Join(root, "journals.sqlite")
	sqlite, err := openResearchSQLiteJournalStore(s.store, path, "dwell-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	group, err := newResearchSQLiteGroupJournalStore(sqlite, 8, wait)
	if err != nil {
		t.Fatal(err)
	}
	defer group.closeGroup()
	probe, err := New(recallProfileStore{EventStore: group}, s.embedder, Config{DefaultRecallK: 200, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}

	type job struct {
		i       int
		offered time.Time
	}
	jobs := make(chan job, 192)
	samples := make(chan recallProfileSample, 192)
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for work := range jobs {
				request := motionExitRequest(probe, now.Add(time.Duration(2*work.i)*time.Second))
				request.SessionID = fmt.Sprintf("profile-%d", work.i)
				request.RecallK, request.PackK = 200, 10
				start := time.Now()
				trace := &recallProfileTrace{origin: start, spans: make(map[string][]recallProfileSpan)}
				packet, callErr := probe.Recall(context.WithValue(ctx, recallProfileKey{}, trace), request)
				end := time.Now()
				if callErr == nil {
					seen := make(map[string]bool, 200)
					for _, decision := range packet.BayesianShadow.Decisions {
						if !expected[decision.EventID] || seen[decision.EventID] {
							callErr = fmt.Errorf("unexpected or duplicate nomination %s", decision.EventID)
							break
						}
						seen[decision.EventID] = true
					}
					if callErr == nil && len(seen) != 200 {
						callErr = fmt.Errorf("nominated %d of 200 live events", len(seen))
					}
					for _, candidate := range packet.Candidates {
						if candidate.Event.AvailableAt.After(request.AsOf) {
							callErr = fmt.Errorf("future event %s in packet", candidate.Event.ID)
						}
					}
				}
				samples <- recallProfileSample{offer: end.Sub(work.offered).Nanoseconds(), call: end.Sub(start).Nanoseconds(), queue: start.Sub(work.offered).Nanoseconds(), spans: trace.spans, err: callErr}
			}
		}()
	}
	probeDone := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(8 * time.Millisecond)
		defer ticker.Stop()
		for i := range 192 {
			if i > 0 {
				select {
				case <-ctx.Done():
					close(jobs)
					workers.Wait()
					close(samples)
					probeDone <- struct{}{}
					return
				case <-ticker.C:
				}
			}
			jobs <- job{i: i, offered: time.Now()}
		}
		close(jobs)
		workers.Wait()
		close(samples)
		probeDone <- struct{}{}
	}()
	var active atomic.Bool
	active.Store(true)
	var writerDone <-chan struct {
		writes, overlap int
		err             error
	}
	if writer {
		writerDone = motionLoadWriter(ctx, s, now, &active)
	}
	r := newRecallProfileTrial()
	for sample := range samples {
		r.add(sample)
	}
	<-probeDone
	active.Store(false)
	if writer {
		got := <-writerDone
		r.writes, r.overlap = got.writes, got.overlap
		if got.err != nil && r.err == nil {
			r.err = got.err
		}
	}
	if r.err != nil || len(r.offer) != 192 || (writer && (r.writes != 256 || r.overlap == 0)) {
		t.Fatalf("incomplete trial: offers=%d writes=%d overlap=%d err=%v", len(r.offer), r.writes, r.overlap, r.err)
	}
	if err := group.closeGroup(); err != nil {
		t.Fatal(err)
	}
	r.batchSizes, r.guardRejects = group.batchSizes(), group.guardRejections()
	reopened, err := openResearchSQLiteJournalStore(s.store, path, "dwell-owner", false)
	if err != nil {
		t.Fatal(err)
	}
	count, countErr := reopened.count(ctx)
	closeErr := reopened.closeJournal()
	if countErr != nil || closeErr != nil || count != 192 {
		t.Fatalf("durable reopen count=%d read=%v close=%v", count, countErr, closeErr)
	}
	total := 0
	for _, size := range r.batchSizes {
		if size < 1 || size > 8 {
			t.Fatalf("invalid batch size %d", size)
		}
		total += size
	}
	if total != 192 || r.guardRejects != 0 {
		t.Fatalf("batch rows=%d rejects=%d", total, r.guardRejects)
	}
	return r
}

func TestResearchRecallGroupDwellV23(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_DWELL_V23") != "1" {
		t.Skip("opt-in v23 research dwell diagnostic")
	}
	arms := []string{"sqlite", "group8ms", "group1ns"}
	all := make(map[string]map[bool]*recallProfileTrial, len(arms))
	for _, arm := range arms {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all[arm] = map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				var r recallProfileTrial
				if arm == "sqlite" {
					r = runRecallProfileTrialMode(t, writer, 200, "sqlite")
				} else {
					wait := 8 * time.Millisecond
					if arm == "group1ns" {
						wait = time.Nanosecond
					}
					r = runResearchRecallDwellTrialV23(t, writer, wait)
				}
				all[arm][writer].merge(r)
				t.Logf("trial=%d arm=%s writer=%v offers=%d writes=%d offer_p99=%s queue_p99=%s", trial, arm, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.queue, .99))
			}
		}
	}
	for _, arm := range arms {
		for _, writer := range []bool{false, true} {
			r := all[arm][writer]
			if len(r.offer) != 576 || (writer && r.writes != 768) {
				t.Fatalf("arm=%s writer=%v incomplete offers=%d writes=%d", arm, writer, len(r.offer), r.writes)
			}
			var histogram [9]int
			for _, size := range r.batchSizes {
				histogram[size]++
			}
			t.Logf("arm=%s writer=%v offers=%d writes=%d offer_p50=%s offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s batches=%v rejects=%d", arm, writer, len(r.offer), r.writes,
				researchDurableLoadPercentile(r.offer, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99),
				researchDurableLoadPercentile(r.queue, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99), histogram, r.guardRejects)
		}
	}
	noWait := researchDurableLoadPercentile(all["group1ns"][true].offer, .99)
	single := researchDurableLoadPercentile(all["sqlite"][true].offer, .99)
	if noWait >= 100*time.Millisecond || noWait > single {
		t.Errorf("frozen v23 Goal 6 candidate gate failed: no-wait=%s single=%s", noWait, single)
	}
}
