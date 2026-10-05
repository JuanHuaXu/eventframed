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

type researchWorkerTrialV25 struct {
	recallProfileTrial
	gaps []int64
	rate float64
}

func runResearchRecallWorkersV25(t *testing.T, writer bool, workerCount int) researchWorkerTrialV25 {
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
	sqlite, err := openResearchSQLiteJournalStore(s.store, path, "workers-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.closeJournal()
	probe, err := New(recallProfileStore{EventStore: sqlite}, s.embedder, Config{DefaultRecallK: 200, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}

	type job struct {
		i       int
		offered time.Time
	}
	jobs := make(chan job, 192)
	offered := make(chan time.Time, 192)
	samples := make(chan recallProfileSample, 192)
	var workers sync.WaitGroup
	for range workerCount {
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
					close(offered)
					probeDone <- struct{}{}
					return
				case <-ticker.C:
				}
			}
			at := time.Now()
			offered <- at
			jobs <- job{i: i, offered: at}
		}
		close(jobs)
		workers.Wait()
		close(samples)
		close(offered)
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
	r := researchWorkerTrialV25{recallProfileTrial: newRecallProfileTrial()}
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
	var times []time.Time
	for at := range offered {
		times = append(times, at)
	}
	if r.err != nil || len(r.offer) != 192 || len(times) != 192 || (writer && (r.writes != 256 || r.overlap == 0)) {
		t.Fatalf("incomplete trial: workers=%d offers=%d timestamps=%d writes=%d overlap=%d err=%v", workerCount, len(r.offer), len(times), r.writes, r.overlap, r.err)
	}
	for i := 1; i < len(times); i++ {
		gap := times[i].Sub(times[i-1]).Nanoseconds()
		if gap <= 0 {
			t.Fatal("nonmonotonic offer clock")
		}
		r.gaps = append(r.gaps, gap)
	}
	r.rate = float64(len(r.gaps)) / times[len(times)-1].Sub(times[0]).Seconds()
	if err := sqlite.closeJournal(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openResearchSQLiteJournalStore(s.store, path, "workers-owner", false)
	if err != nil {
		t.Fatal(err)
	}
	count, countErr := reopened.count(ctx)
	closeErr := reopened.closeJournal()
	if countErr != nil || closeErr != nil || count != 192 {
		t.Fatalf("durable reopen count=%d read=%v close=%v", count, countErr, closeErr)
	}
	return r
}

func TestResearchRecallWorkersV25(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_WORKERS_V25") != "1" {
		t.Skip("opt-in bounded Recall worker scaling")
	}
	counts := []int{4, 8, 16}
	type aggregate struct {
		combined recallProfileTrial
		gaps     []int64
		trials   []researchWorkerTrialV25
	}
	all := make(map[int]map[bool]*aggregate, len(counts))
	for _, count := range counts {
		all[count] = make(map[bool]*aggregate, 2)
		for _, writer := range []bool{false, true} {
			all[count][writer] = &aggregate{combined: newRecallProfileTrial()}
		}
	}
	for trial := range 3 {
		for offset := range counts {
			count := counts[(trial+offset)%len(counts)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runResearchRecallWorkersV25(t, writer, count)
				a := all[count][writer]
				a.combined.merge(r.recallProfileTrial)
				a.gaps = append(a.gaps, r.gaps...)
				a.trials = append(a.trials, r)
				t.Logf("trial=%d workers=%d writer=%v offers=192 writes=%d rate=%.2f/s gap_p50=%s gap_p99=%s mean_call=%s offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s guard_wait_p99=%s", trial, count, writer, r.writes, r.rate,
					researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.gaps, .99), researchCapacityMeanV24(r.call),
					researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99),
					researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99), researchDurableLoadPercentile(r.parts["sqlite_guard_wait"].requestSpans, .99))
			}
		}
	}
	for _, count := range counts {
		for _, writer := range []bool{false, true} {
			a := all[count][writer]
			if len(a.combined.offer) != 576 || len(a.gaps) != 573 || (writer && a.combined.writes != 768) {
				t.Fatalf("incomplete aggregate: workers=%d writer=%v offers=%d gaps=%d writes=%d", count, writer, len(a.combined.offer), len(a.gaps), a.combined.writes)
			}
			t.Logf("pooled workers=%d writer=%v offers=576 writes=%d gap_p50=%s gap_p99=%s mean_call=%s offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s guard_wait_p99=%s", count, writer, a.combined.writes,
				researchDurableLoadPercentile(a.gaps, .5), researchDurableLoadPercentile(a.gaps, .99), researchCapacityMeanV24(a.combined.call),
				researchDurableLoadPercentile(a.combined.offer, .99), researchDurableLoadPercentile(a.combined.call, .99), researchDurableLoadPercentile(a.combined.queue, .99),
				researchDurableLoadPercentile(a.combined.parts["journal"].requestSpans, .99), researchDurableLoadPercentile(a.combined.parts["sqlite_guard_wait"].requestSpans, .99))
		}
	}
	quiet4 := researchDurableLoadPercentile(all[4][false].combined.offer, .99)
	for _, count := range []int{8, 16} {
		quiet := researchDurableLoadPercentile(all[count][false].combined.offer, .99)
		loaded := researchDurableLoadPercentile(all[count][true].combined.offer, .99)
		pass := quiet <= quiet4+10*time.Millisecond && loaded < 100*time.Millisecond
		for _, r := range all[count][true].trials {
			pass = pass && researchDurableLoadPercentile(r.offer, .99) < 100*time.Millisecond
		}
		if !pass {
			t.Errorf("frozen v25 latency component failed: workers=%d quiet=%s quiet4=%s loaded=%s", count, quiet, quiet4, loaded)
		}
	}
}
