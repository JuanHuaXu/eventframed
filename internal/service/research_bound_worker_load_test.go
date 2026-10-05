package service

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type motionLoadTrial struct {
	offerNS, callNS, queueNS []int64
	admitNS, feedbackNS      []int64
	ageNS                    []int64
	probeOverlap, writes     int
	writeOverlap             int
	before, after            model.Snapshot
}

type motionLoadOffer struct {
	id uint64
	at time.Time
}

type motionLoadProbeResult struct {
	offerNS, callNS, queueNS []int64
	overlap                  int
	err                      error
}

func motionLoadProbe(ctx context.Context, probe *Service, now time.Time, interval time.Duration, active *atomic.Bool) <-chan motionLoadProbeResult {
	return motionLoadProbeBudgeted(ctx, probe, now, interval, active, 3, 2)
}

func motionLoadProbeBudgeted(ctx context.Context, probe *Service, now time.Time, interval time.Duration, active *atomic.Bool, recallK, packK int) <-chan motionLoadProbeResult {
	return motionLoadProbeBudgetedExpected(ctx, probe, now, interval, active, recallK, packK, nil)
}

func motionLoadProbeBudgetedExpected(ctx context.Context, probe *Service, now time.Time, interval time.Duration, active *atomic.Bool, recallK, packK int, expected map[string]bool) <-chan motionLoadProbeResult {
	done := make(chan motionLoadProbeResult, 1)
	go func() {
		type job struct {
			i       int
			offered time.Time
		}
		type sample struct {
			offerNS, callNS, queueNS int64
			overlap                  bool
			err                      error
		}
		jobs := make(chan job, 192)
		samples := make(chan sample, 192)
		var workers sync.WaitGroup
		for range 4 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for work := range jobs {
					request := motionExitRequest(probe, now.Add(time.Duration(2*work.i)*time.Second))
					request.SessionID = fmt.Sprintf("probe-%d", work.i)
					request.RecallK = recallK
					request.PackK = packK
					request.Embedding = nil
					request.EmbeddingModel = ""
					inWindow := active.Load()
					start := time.Now()
					packet, err := probe.Recall(ctx, request)
					if err == nil {
						for _, candidate := range packet.Candidates {
							if candidate.Event.AvailableAt.After(request.AsOf) {
								err = fmt.Errorf("future event %s entered as-of packet", candidate.Event.ID)
								break
							}
						}
						if expected != nil {
							seen := make(map[string]bool, len(expected))
							for _, decision := range packet.BayesianShadow.Decisions {
								if !expected[decision.EventID] || seen[decision.EventID] {
									err = fmt.Errorf("unexpected or duplicate nominated event %s", decision.EventID)
									break
								}
								seen[decision.EventID] = true
							}
							if err == nil && len(seen) != len(expected) {
								err = fmt.Errorf("as-of nomination has %d of %d expected events", len(seen), len(expected))
							}
						}
					}
					end := time.Now()
					samples <- sample{offerNS: end.Sub(work.offered).Nanoseconds(), callNS: end.Sub(start).Nanoseconds(), queueNS: start.Sub(work.offered).Nanoseconds(), overlap: inWindow, err: err}
				}
			}()
		}
		ticker := time.NewTicker(interval)
	dispatch:
		for i := range 192 {
			if i > 0 {
				select {
				case <-ctx.Done():
					break dispatch
				case <-ticker.C:
				}
			}
			jobs <- job{i: i, offered: time.Now()}
		}
		ticker.Stop()
		close(jobs)
		workers.Wait()
		close(samples)
		result := motionLoadProbeResult{offerNS: make([]int64, 0, 192), callNS: make([]int64, 0, 192), queueNS: make([]int64, 0, 192)}
		for s := range samples {
			if s.err != nil && result.err == nil {
				result.err = s.err
			}
			result.offerNS = append(result.offerNS, s.offerNS)
			result.callNS = append(result.callNS, s.callNS)
			result.queueNS = append(result.queueNS, s.queueNS)
			if s.overlap {
				result.overlap++
			}
		}
		done <- result
	}()
	return done
}

func motionLoadWriter(ctx context.Context, s *Service, now time.Time, active *atomic.Bool) <-chan struct {
	writes, overlap int
	err             error
} {
	result := make(chan struct {
		writes, overlap int
		err             error
	}, 1)
	go func() {
		var out struct {
			writes, overlap int
			err             error
		}
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for i := range 256 {
			if i > 0 {
				select {
				case <-ctx.Done():
					out.err = ctx.Err()
					result <- out
					return
				case <-ticker.C:
				}
			}
			event := testutil.Event(fmt.Sprintf("future-load-%d", i), "public query fixture", now.Add(time.Hour))
			inWindow := active.Load()
			if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
				out.err = err
				result <- out
				return
			}
			out.writes++
			if inWindow {
				out.overlap++
			}
		}
		result <- out
	}()
	return result
}

func runMotionBoundWorkerLoadTrial(t *testing.T, writer bool, interval time.Duration) motionLoadTrial {
	return runMotionBoundWorkerLoadTrialSized(t, writer, interval, 1)
}

func runMotionBoundWorkerLoadTrialSized(t *testing.T, writer bool, interval time.Duration, liveCount int) motionLoadTrial {
	t.Helper()
	if liveCount != 1 && liveCount != 50 && liveCount != 200 {
		t.Fatal("unsupported frozen live corpus size")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	s, tap := motionExitService(t, root, true)
	defer s.Close()
	defer tap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	expected := map[string]bool{"seed": true}
	for i := 1; i < liveCount; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	recallK, packK := 3, 2
	if liveCount >= 50 {
		recallK, packK = liveCount, 10
	}
	probe, err := New(s.store, s.embedder, Config{DefaultRecallK: recallK, DefaultPackK: packK, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	source, err := researchmemory.OpenDurable(ctx, root+"/source.sqlite", "tenant-a", "source-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	path := root + "/motion.sqlite"
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: now, KeyID: "motion-load", Key: []byte("0123456789abcdef0123456789abcdef")}
	prepared, err := s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path)
	if err != nil {
		t.Fatal(err)
	}
	openCtx, release := context.WithTimeout(ctx, 2*time.Second)
	worker, err := s.OpenPreparedResearchBoundWorker(openCtx, prepared, path)
	release()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = worker.Close() }()
	trial := motionLoadTrial{before: s.store.Snapshot(ctx), admitNS: make([]int64, 0, 64), feedbackNS: make([]int64, 0, 64)}
	var active atomic.Bool
	active.Store(true)
	probeDone := motionLoadProbeBudgeted(ctx, probe, now, interval, &active, recallK, packK)
	var writerDone <-chan struct {
		writes, overlap int
		err             error
	}
	if writer {
		writerDone = motionLoadWriter(ctx, s, now, &active)
	}
	submissions := make(chan motionLoadOffer, 64)
	observerDone := make(chan struct {
		ages []int64
		err  error
	}, 1)
	go func() {
		out := struct {
			ages []int64
			err  error
		}{ages: make([]int64, 0, 64)}
		for offered := range submissions {
			if err := worker.WaitProcessed(ctx, offered.id); err != nil {
				out.err = err
				break
			}
			completed, failed, _, _ := worker.Counts()
			if failed != 0 || completed < offered.id {
				out.err = fmt.Errorf("label %d not applied: completed=%d failed=%d", offered.id, completed, failed)
				break
			}
			out.ages = append(out.ages, time.Since(offered.at).Nanoseconds())
		}
		observerDone <- out
	}()
	closed := false
	defer func() {
		cancel()
		if !closed {
			close(submissions)
		}
		if observerDone != nil {
			<-observerDone
		}
		if probeDone != nil {
			<-probeDone
		}
		if writerDone != nil {
			<-writerDone
		}
	}()
	for i := range 64 {
		at := now.Add(time.Duration(2*i) * time.Second)
		request := motionExitRequest(s, at)
		request.SessionID = fmt.Sprintf("learner-%d", i)
		request.RecallK = recallK
		request.PackK = packK
		if _, err := s.Recall(ctx, request); err != nil {
			t.Fatal(err)
		}
		observed, err := tap.Take(ctx)
		if err != nil || len(observed.Candidates) != liveCount {
			t.Fatalf("as-of frontier changed: %+v err=%v", observed, err)
		}
		seen := make(map[string]bool, liveCount)
		for _, got := range observed.Candidates {
			if !expected[got.EventID] || seen[got.EventID] {
				t.Fatalf("unexpected or duplicate as-of event %s", got.EventID)
			}
			seen[got.EventID] = true
		}
		candidate := motionExitCandidate(t, observed)
		admitCtx, stop := context.WithTimeout(ctx, 2*time.Second)
		start := time.Now()
		prediction, retry, err := worker.Admit(admitCtx, request, observed, candidate)
		trial.admitNS = append(trial.admitNS, time.Since(start).Nanoseconds())
		stop()
		if err != nil || retry || prediction.ID != uint64(i+1) {
			t.Fatalf("guarded admit %d: %+v retry=%v err=%v", i, prediction, retry, err)
		}
		feedbackCtx, stop := context.WithTimeout(ctx, 2*time.Second)
		start = time.Now()
		submissions <- motionLoadOffer{id: prediction.ID, at: start}
		_, err = worker.Feedback(feedbackCtx, request, observed.JournalID, "seed", i%3 != 0, at.Add(time.Second))
		trial.feedbackNS = append(trial.feedbackNS, time.Since(start).Nanoseconds())
		stop()
		if err != nil {
			t.Fatalf("guarded feedback %d: %v", i, err)
		}
	}
	active.Store(false)
	close(submissions)
	closed = true
	observed := <-observerDone
	observerDone = nil
	if observed.err != nil || len(observed.ages) != 64 {
		t.Fatalf("live completion: count=%d err=%v", len(observed.ages), observed.err)
	}
	trial.ageNS = observed.ages
	probes := <-probeDone
	probeDone = nil
	if probes.err != nil || len(probes.offerNS) != 192 || probes.overlap == 0 {
		t.Fatalf("probe completion or overlap: count=%d overlap=%d err=%v", len(probes.offerNS), probes.overlap, probes.err)
	}
	trial.offerNS, trial.callNS, trial.queueNS, trial.probeOverlap = probes.offerNS, probes.callNS, probes.queueNS, probes.overlap
	if writer {
		writes := <-writerDone
		writerDone = nil
		if writes.err != nil || writes.writes != 256 || writes.overlap == 0 {
			t.Fatalf("future writer: writes=%d overlap=%d err=%v", writes.writes, writes.overlap, writes.err)
		}
		trial.writes, trial.writeOverlap = writes.writes, writes.overlap
	}
	trial.after = s.store.Snapshot(ctx)
	if completed, failed, pending, queued := worker.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatalf("worker counts: %d/%d/%d/%d", completed, failed, pending, queued)
	}
	labels, _, err := worker.durable.BoundLabelsWithSequence(ctx, now.Add(200*time.Second))
	if err != nil || len(labels) != 64 {
		t.Fatalf("incomplete durable label replay: count=%d err=%v", len(labels), err)
	}
	for i, label := range labels {
		if label.Prediction.Binding == nil || label.Prediction.Witness == nil || label.Prediction.Binding.EventID != "seed" || label.Feedback.ID != uint64(i+1) {
			t.Fatalf("unbound label %d: %+v", i, label)
		}
	}
	if err := worker.Close(); err != nil {
		t.Fatal(err)
	}
	plan.Target = trial.after
	prepared, err = s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path)
	if err != nil {
		t.Fatal(err)
	}
	openCtx, release = context.WithTimeout(ctx, 5*time.Second)
	worker, err = s.OpenPreparedResearchBoundWorker(openCtx, prepared, path)
	release()
	if err != nil {
		t.Fatal(err)
	}
	if completed, failed, pending, queued := worker.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatalf("post-load reopen counts: %d/%d/%d/%d", completed, failed, pending, queued)
	}
	return trial
}

func TestResearchMotionBoundWorkerLoadedScreenV11(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_LOAD_V11") != "1" {
		t.Skip("opt-in frozen load screen; known negative result")
	}
	all := map[bool]*motionLoadTrial{false: {}, true: {}}
	for trial := range 3 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, writer := range order {
			r := runMotionBoundWorkerLoadTrial(t, writer, time.Millisecond)
			arm := all[writer]
			arm.offerNS = append(arm.offerNS, r.offerNS...)
			arm.callNS = append(arm.callNS, r.callNS...)
			arm.queueNS = append(arm.queueNS, r.queueNS...)
			arm.admitNS = append(arm.admitNS, r.admitNS...)
			arm.feedbackNS = append(arm.feedbackNS, r.feedbackNS...)
			arm.ageNS = append(arm.ageNS, r.ageNS...)
			arm.probeOverlap += r.probeOverlap
			arm.writes += r.writes
			arm.writeOverlap += r.writeOverlap
			name := "quiet"
			if writer {
				name = "writer"
			}
			t.Logf("trial=%d arm=%s probes=%d labels=%d writes=%d probe_overlap=%d write_overlap=%d versions=%d->%d offer_p99=%s call_p99=%s queue_p99=%s age_p99=%s", trial, name, len(r.offerNS), len(r.ageNS), r.writes, r.probeOverlap, r.writeOverlap, r.before.RuntimeVersion, r.after.RuntimeVersion, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), researchDurableLoadPercentile(r.ageNS, .99))
		}
	}
	for _, writer := range []bool{false, true} {
		arm := all[writer]
		name := "quiet"
		if writer {
			name = "writer"
		}
		t.Logf("arm=%s probes=%d labels=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s admit_p99=%s feedback_p99=%s age_p50=%s age_p95=%s age_p99=%s age_max=%s", name, len(arm.offerNS), len(arm.ageNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
			researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), researchDurableLoadPercentile(arm.offerNS, .99), researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99), researchDurableLoadPercentile(arm.admitNS, .99), researchDurableLoadPercentile(arm.feedbackNS, .99), researchDurableLoadPercentile(arm.ageNS, .5), researchDurableLoadPercentile(arm.ageNS, .95), researchDurableLoadPercentile(arm.ageNS, .99), researchDurableLoadPercentile(arm.ageNS, 1))
	}
	if os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" {
		writer := all[true]
		if researchDurableLoadPercentile(writer.offerNS, .99) >= 100*time.Millisecond || researchDurableLoadPercentile(writer.ageNS, .99) >= 250*time.Millisecond {
			t.Error("frozen ordinary-build writer-arm latency or freshness gate failed")
		}
	}
}

func TestResearchMotionProbeOnlyDiagnosticV11(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_LOAD_V11") != "1" {
		t.Skip("opt-in sibling diagnostic for frozen load screen")
	}
	for trial := range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		root := t.TempDir()
		s, tap := motionExitService(t, root, true)
		now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
		putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
		probe, err := New(s.store, s.embedder, Config{DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
		if err != nil {
			t.Fatal(err)
		}
		var inactive atomic.Bool
		result := <-motionLoadProbe(ctx, probe, now, time.Millisecond, &inactive)
		if result.err != nil || len(result.offerNS) != 192 {
			t.Fatalf("probe-only trial %d: calls=%d err=%v", trial, len(result.offerNS), result.err)
		}
		t.Logf("probe_only_trial=%d calls=%d offer_p99=%s call_p99=%s queue_p99=%s", trial, len(result.offerNS), researchDurableLoadPercentile(result.offerNS, .99), researchDurableLoadPercentile(result.callNS, .99), researchDurableLoadPercentile(result.queueNS, .99))
		cancel()
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		tap.Close()
	}
}

func TestResearchMotionBoundWorkerRateMatrixV12(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_RATE_V12") != "1" {
		t.Skip("opt-in frozen v12 rate matrix")
	}
	for _, interval := range []time.Duration{time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond, 8 * time.Millisecond} {
		all := map[bool]*motionLoadTrial{false: {}, true: {}}
		for trial := range 3 {
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runMotionBoundWorkerLoadTrial(t, writer, interval)
				arm := all[writer]
				arm.offerNS = append(arm.offerNS, r.offerNS...)
				arm.callNS = append(arm.callNS, r.callNS...)
				arm.queueNS = append(arm.queueNS, r.queueNS...)
				arm.ageNS = append(arm.ageNS, r.ageNS...)
				arm.admitNS = append(arm.admitNS, r.admitNS...)
				arm.feedbackNS = append(arm.feedbackNS, r.feedbackNS...)
				arm.writes += r.writes
				arm.writeOverlap += r.writeOverlap
				arm.probeOverlap += r.probeOverlap
				name := "quiet"
				if writer {
					name = "writer"
				}
				t.Logf("rate_ms=%d trial=%d arm=%s offer_p99=%s call_p99=%s queue_p99=%s age_p99=%s writes=%d probe_overlap=%d write_overlap=%d versions=%d->%d", interval/time.Millisecond, trial, name, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), researchDurableLoadPercentile(r.ageNS, .99), r.writes, r.probeOverlap, r.writeOverlap, r.before.RuntimeVersion, r.after.RuntimeVersion)
			}
		}
		for _, writer := range []bool{false, true} {
			arm := all[writer]
			name := "quiet"
			if writer {
				name = "writer"
			}
			offerP99 := researchDurableLoadPercentile(arm.offerNS, .99)
			ageP99 := researchDurableLoadPercentile(arm.ageNS, .99)
			t.Logf("rate_ms=%d arm=%s probes=%d labels=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s admit_p99=%s feedback_p99=%s age_p50=%s age_p95=%s age_p99=%s age_max=%s", interval/time.Millisecond, name, len(arm.offerNS), len(arm.ageNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
				researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), offerP99, researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99), researchDurableLoadPercentile(arm.admitNS, .99), researchDurableLoadPercentile(arm.feedbackNS, .99), researchDurableLoadPercentile(arm.ageNS, .5), researchDurableLoadPercentile(arm.ageNS, .95), ageP99, researchDurableLoadPercentile(arm.ageNS, 1))
			if writer && interval >= 4*time.Millisecond && os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" && (offerP99 >= 100*time.Millisecond || ageP99 >= 250*time.Millisecond) {
				t.Errorf("frozen v12 rate_ms=%d writer-arm latency or freshness gate failed", interval/time.Millisecond)
			}
		}
	}
}

func TestResearchMotionBoundWorkerCorpusBreadthV13(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_CORPUS_V13") != "1" {
		t.Skip("opt-in frozen v13 corpus-breadth screen")
	}
	for _, liveCount := range []int{1, 50} {
		all := map[bool]*motionLoadTrial{false: {}, true: {}}
		for trial := range 3 {
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runMotionBoundWorkerLoadTrialSized(t, writer, 8*time.Millisecond, liveCount)
				arm := all[writer]
				arm.offerNS = append(arm.offerNS, r.offerNS...)
				arm.callNS = append(arm.callNS, r.callNS...)
				arm.queueNS = append(arm.queueNS, r.queueNS...)
				arm.ageNS = append(arm.ageNS, r.ageNS...)
				arm.admitNS = append(arm.admitNS, r.admitNS...)
				arm.feedbackNS = append(arm.feedbackNS, r.feedbackNS...)
				arm.writes += r.writes
				arm.writeOverlap += r.writeOverlap
				arm.probeOverlap += r.probeOverlap
				name := "quiet"
				if writer {
					name = "writer"
				}
				t.Logf("live=%d trial=%d arm=%s probes=%d labels=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s age_p99=%s versions=%d->%d", liveCount, trial, name, len(r.offerNS), len(r.ageNS), r.writes, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), researchDurableLoadPercentile(r.ageNS, .99), r.before.RuntimeVersion, r.after.RuntimeVersion)
			}
		}
		for _, writer := range []bool{false, true} {
			arm := all[writer]
			name := "quiet"
			if writer {
				name = "writer"
			}
			offerP99 := researchDurableLoadPercentile(arm.offerNS, .99)
			ageP99 := researchDurableLoadPercentile(arm.ageNS, .99)
			t.Logf("live=%d arm=%s probes=%d labels=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s admit_p99=%s feedback_p99=%s age_p50=%s age_p95=%s age_p99=%s age_max=%s", liveCount, name, len(arm.offerNS), len(arm.ageNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
				researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), offerP99, researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99), researchDurableLoadPercentile(arm.admitNS, .99), researchDurableLoadPercentile(arm.feedbackNS, .99), researchDurableLoadPercentile(arm.ageNS, .5), researchDurableLoadPercentile(arm.ageNS, .95), ageP99, researchDurableLoadPercentile(arm.ageNS, 1))
			if writer && os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" && (offerP99 >= 100*time.Millisecond || ageP99 >= 250*time.Millisecond) {
				t.Errorf("frozen v13 live=%d writer-arm latency or freshness gate failed", liveCount)
			}
		}
	}
}

func TestResearchMotionBoundWorkerFrontierCeilingV14(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_CORPUS_V14") != "1" {
		t.Skip("opt-in frozen v14 frontier-ceiling screen")
	}
	for _, liveCount := range []int{50, 200} {
		t.Run(fmt.Sprintf("live-%d", liveCount), func(t *testing.T) {
			all := map[bool]*motionLoadTrial{false: {}, true: {}}
			for trial := range 3 {
				order := []bool{false, true}
				if trial == 1 {
					order = []bool{true, false}
				}
				for _, writer := range order {
					r := runMotionBoundWorkerLoadTrialSized(t, writer, 8*time.Millisecond, liveCount)
					arm := all[writer]
					arm.offerNS = append(arm.offerNS, r.offerNS...)
					arm.callNS = append(arm.callNS, r.callNS...)
					arm.queueNS = append(arm.queueNS, r.queueNS...)
					arm.ageNS = append(arm.ageNS, r.ageNS...)
					arm.admitNS = append(arm.admitNS, r.admitNS...)
					arm.feedbackNS = append(arm.feedbackNS, r.feedbackNS...)
					arm.writes += r.writes
					arm.writeOverlap += r.writeOverlap
					arm.probeOverlap += r.probeOverlap
					name := "quiet"
					if writer {
						name = "writer"
					}
					t.Logf("live=%d trial=%d arm=%s probes=%d labels=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s age_p99=%s versions=%d->%d", liveCount, trial, name, len(r.offerNS), len(r.ageNS), r.writes, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), researchDurableLoadPercentile(r.ageNS, .99), r.before.RuntimeVersion, r.after.RuntimeVersion)
				}
			}
			for _, writer := range []bool{false, true} {
				arm := all[writer]
				name := "quiet"
				if writer {
					name = "writer"
				}
				offerP99 := researchDurableLoadPercentile(arm.offerNS, .99)
				ageP99 := researchDurableLoadPercentile(arm.ageNS, .99)
				t.Logf("live=%d arm=%s probes=%d labels=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s admit_p99=%s feedback_p99=%s age_p50=%s age_p95=%s age_p99=%s age_max=%s", liveCount, name, len(arm.offerNS), len(arm.ageNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
					researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), offerP99, researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99), researchDurableLoadPercentile(arm.admitNS, .99), researchDurableLoadPercentile(arm.feedbackNS, .99), researchDurableLoadPercentile(arm.ageNS, .5), researchDurableLoadPercentile(arm.ageNS, .95), ageP99, researchDurableLoadPercentile(arm.ageNS, 1))
				if writer && os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" && (offerP99 >= 100*time.Millisecond || ageP99 >= 250*time.Millisecond) {
					t.Errorf("frozen v14 live=%d writer-arm latency or freshness gate failed", liveCount)
				}
			}
		})
	}
}

func runMotionProbeWriterOnlyTrial(t *testing.T, writer bool, liveCount int) motionLoadTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	s, tap := motionExitService(t, t.TempDir(), true)
	defer s.Close()
	defer tap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	expected := map[string]bool{"seed": true}
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	for i := 1; i < liveCount; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	probe, err := New(s.store, s.embedder, Config{DefaultRecallK: liveCount, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	var active atomic.Bool
	active.Store(true)
	trial := motionLoadTrial{before: s.store.Snapshot(ctx)}
	probeDone := motionLoadProbeBudgetedExpected(ctx, probe, now, 8*time.Millisecond, &active, liveCount, 10, expected)
	var writerDone <-chan struct {
		writes, overlap int
		err             error
	}
	if writer {
		writerDone = motionLoadWriter(ctx, s, now, &active)
	}
	if writer {
		w := <-writerDone
		if w.err != nil || w.writes != 256 || w.overlap == 0 {
			t.Fatalf("writer completion: writes=%d overlap=%d err=%v", w.writes, w.overlap, w.err)
		}
		trial.writes, trial.writeOverlap = w.writes, w.overlap
	}
	p := <-probeDone
	active.Store(false)
	if p.err != nil || len(p.offerNS) != 192 || p.overlap == 0 {
		t.Fatalf("probe completion: calls=%d overlap=%d err=%v", len(p.offerNS), p.overlap, p.err)
	}
	trial.offerNS, trial.callNS, trial.queueNS = p.offerNS, p.callNS, p.queueNS
	trial.probeOverlap = p.overlap
	trial.after = s.store.Snapshot(ctx)
	return trial
}

func TestResearchMotionProbeWriterOnlyV15(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_MOTION_PROBE_V15") != "1" {
		t.Skip("opt-in frozen v15 no-learner probe/writer control")
	}
	for _, liveCount := range []int{50, 200} {
		t.Run(fmt.Sprintf("live-%d", liveCount), func(t *testing.T) {
			all := map[bool]*motionLoadTrial{false: {}, true: {}}
			for trial := range 3 {
				order := []bool{false, true}
				if trial == 1 {
					order = []bool{true, false}
				}
				for _, writer := range order {
					r := runMotionProbeWriterOnlyTrial(t, writer, liveCount)
					arm := all[writer]
					arm.offerNS = append(arm.offerNS, r.offerNS...)
					arm.callNS = append(arm.callNS, r.callNS...)
					arm.queueNS = append(arm.queueNS, r.queueNS...)
					arm.writes += r.writes
					arm.writeOverlap += r.writeOverlap
					arm.probeOverlap += r.probeOverlap
					name := "quiet"
					if writer {
						name = "writer"
					}
					t.Logf("live=%d trial=%d arm=%s probes=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s versions=%d->%d", liveCount, trial, name, len(r.offerNS), r.writes, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), r.before.RuntimeVersion, r.after.RuntimeVersion)
				}
			}
			for _, writer := range []bool{false, true} {
				arm := all[writer]
				name := "quiet"
				if writer {
					name = "writer"
				}
				t.Logf("live=%d arm=%s probes=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s", liveCount, name, len(arm.offerNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
					researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), researchDurableLoadPercentile(arm.offerNS, .99), researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99))
			}
		})
	}
}

func motionRawSearchProbe(ctx context.Context, s *Service, now time.Time, active *atomic.Bool, expected map[string]bool) <-chan motionLoadProbeResult {
	done := make(chan motionLoadProbeResult, 1)
	go func() {
		type job struct {
			i       int
			offered time.Time
		}
		type sample struct {
			offerNS, callNS, queueNS int64
			overlap                  bool
			err                      error
		}
		jobs := make(chan job, 192)
		samples := make(chan sample, 192)
		var workers sync.WaitGroup
		for range 4 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for work := range jobs {
					asOf := now.Add(time.Duration(2*work.i) * time.Second)
					inWindow := active.Load()
					start := time.Now()
					results, err := s.store.Search(ctx, "tenant-a", []float32{1, 0, 0, 0, 0, 0, 0, 0}, asOf, len(expected))
					if err == nil {
						seen := make(map[string]bool, len(expected))
						for _, result := range results {
							id := result.Event.ID
							if !expected[id] || seen[id] || result.Event.AvailableAt.After(asOf) {
								err = fmt.Errorf("unexpected raw as-of search event %s", id)
								break
							}
							seen[id] = true
						}
						if err == nil && len(seen) != len(expected) {
							err = fmt.Errorf("raw as-of search has %d of %d expected events", len(seen), len(expected))
						}
					}
					end := time.Now()
					samples <- sample{offerNS: end.Sub(work.offered).Nanoseconds(), callNS: end.Sub(start).Nanoseconds(), queueNS: start.Sub(work.offered).Nanoseconds(), overlap: inWindow, err: err}
				}
			}()
		}
		ticker := time.NewTicker(8 * time.Millisecond)
	dispatch:
		for i := range 192 {
			if i > 0 {
				select {
				case <-ctx.Done():
					break dispatch
				case <-ticker.C:
				}
			}
			jobs <- job{i: i, offered: time.Now()}
		}
		ticker.Stop()
		close(jobs)
		workers.Wait()
		close(samples)
		out := motionLoadProbeResult{offerNS: make([]int64, 0, 192), callNS: make([]int64, 0, 192), queueNS: make([]int64, 0, 192)}
		for got := range samples {
			if got.err != nil && out.err == nil {
				out.err = got.err
			}
			out.offerNS = append(out.offerNS, got.offerNS)
			out.callNS = append(out.callNS, got.callNS)
			out.queueNS = append(out.queueNS, got.queueNS)
			if got.overlap {
				out.overlap++
			}
		}
		done <- out
	}()
	return done
}

func runMotionRawSearchTrial(t *testing.T, writer bool, liveCount int) motionLoadTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	s, tap := motionExitService(t, t.TempDir(), true)
	defer s.Close()
	defer tap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	expected := map[string]bool{"seed": true}
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	for i := 1; i < liveCount; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	var active atomic.Bool
	active.Store(true)
	trial := motionLoadTrial{before: s.store.Snapshot(ctx)}
	probeDone := motionRawSearchProbe(ctx, s, now, &active, expected)
	var writerDone <-chan struct {
		writes, overlap int
		err             error
	}
	if writer {
		writerDone = motionLoadWriter(ctx, s, now, &active)
		w := <-writerDone
		if w.err != nil || w.writes != 256 || w.overlap == 0 {
			t.Fatalf("raw-search writer: writes=%d overlap=%d err=%v", w.writes, w.overlap, w.err)
		}
		trial.writes, trial.writeOverlap = w.writes, w.overlap
	}
	p := <-probeDone
	active.Store(false)
	if p.err != nil || len(p.offerNS) != 192 || p.overlap == 0 {
		t.Fatalf("raw search completion: calls=%d overlap=%d err=%v", len(p.offerNS), p.overlap, p.err)
	}
	trial.offerNS, trial.callNS, trial.queueNS = p.offerNS, p.callNS, p.queueNS
	trial.probeOverlap = p.overlap
	trial.after = s.store.Snapshot(ctx)
	return trial
}

func TestResearchRawSearchFutureWriterV16(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RAW_SEARCH_V16") != "1" {
		t.Skip("opt-in frozen v16 raw-search writer diagnostic")
	}
	for _, liveCount := range []int{50, 200} {
		t.Run(fmt.Sprintf("live-%d", liveCount), func(t *testing.T) {
			all := map[bool]*motionLoadTrial{false: {}, true: {}}
			for trial := range 3 {
				order := []bool{false, true}
				if trial == 1 {
					order = []bool{true, false}
				}
				for _, writer := range order {
					r := runMotionRawSearchTrial(t, writer, liveCount)
					arm := all[writer]
					arm.offerNS = append(arm.offerNS, r.offerNS...)
					arm.callNS = append(arm.callNS, r.callNS...)
					arm.queueNS = append(arm.queueNS, r.queueNS...)
					arm.writes += r.writes
					arm.writeOverlap += r.writeOverlap
					arm.probeOverlap += r.probeOverlap
					name := "quiet"
					if writer {
						name = "writer"
					}
					t.Logf("live=%d trial=%d arm=%s probes=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s versions=%d->%d", liveCount, trial, name, len(r.offerNS), r.writes, researchDurableLoadPercentile(r.offerNS, .99), researchDurableLoadPercentile(r.callNS, .99), researchDurableLoadPercentile(r.queueNS, .99), r.before.RuntimeVersion, r.after.RuntimeVersion)
				}
			}
			for _, writer := range []bool{false, true} {
				arm := all[writer]
				name := "quiet"
				if writer {
					name = "writer"
				}
				t.Logf("live=%d arm=%s probes=%d writes=%d probe_overlap=%d write_overlap=%d offer_p50=%s offer_p95=%s offer_p99=%s offer_max=%s call_p99=%s queue_p99=%s", liveCount, name, len(arm.offerNS), arm.writes, arm.probeOverlap, arm.writeOverlap,
					researchDurableLoadPercentile(arm.offerNS, .5), researchDurableLoadPercentile(arm.offerNS, .95), researchDurableLoadPercentile(arm.offerNS, .99), researchDurableLoadPercentile(arm.offerNS, 1), researchDurableLoadPercentile(arm.callNS, .99), researchDurableLoadPercentile(arm.queueNS, .99))
			}
		})
	}
}
