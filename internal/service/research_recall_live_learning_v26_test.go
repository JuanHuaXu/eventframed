package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type researchBoundSQLiteV26 struct{ *researchSQLiteJournalStore }

type researchGuardTimingKeyV35 struct{}
type researchGuardTimingV35 struct{ enteredAt time.Time }

func (s researchBoundSQLiteV26) WithResearchAsOfSnapshotWait(ctx context.Context, snapshot model.Snapshot, at time.Time, work func() error) error {
	if timing, ok := ctx.Value(researchGuardTimingKeyV35{}).(*researchGuardTimingV35); ok {
		return s.guard.WithResearchAsOfSnapshotWait(ctx, snapshot, at, func() error {
			timing.enteredAt = time.Now()
			return work()
		})
	}
	return s.guard.WithResearchAsOfSnapshotWait(ctx, snapshot, at, work)
}

func (s researchBoundSQLiteV26) ResearchPublicationCompatible(ctx context.Context, snapshot model.Snapshot, at time.Time) bool {
	proof, ok := s.researchSQLiteJournalStore.EventStore.(interface {
		ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
	})
	return ok && proof.ResearchPublicationCompatible(ctx, snapshot, at)
}

type researchLiveConsumerV26 struct {
	seen, admitted                 int
	frontierAge, feedbackAge       []int64
	tapWait, preFeedback           []int64
	preFeedbackPartsV32            []researchPreFeedbackPartsV32
	admissionPartsV34              []researchAdmissionPartsV34
	feedbackGuard, publicationWait []int64
	firstRequest                   model.RecallRequest
	firstOriginal                  researchmemory.RecordedPrediction
	err                            error
}

func consumeResearchLiveV26(ctx context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
	var out researchLiveConsumerV26
	var previousAvailable time.Time
	for {
		in, err := tap.Take(ctx)
		if err != nil {
			if ctx.Err() != nil {
				out.err = ctx.Err()
			}
			return out
		}
		takenAt := time.Now()
		out.seen++
		if out.seen%3 != 1 {
			continue
		}
		var candidate ResearchFrontierCandidate
		found := false
		for _, c := range in.Candidates {
			if c.EventID == "seed" {
				candidate, found = c, true
				break
			}
		}
		if !found || len(in.Candidates) != 200 {
			out.err = fmt.Errorf("frontier %d missing exact seed or 200 candidates", out.seen)
			return out
		}
		journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
		if err != nil {
			out.err = err
			return out
		}
		request := motionExitRequest(s, in.AsOf)
		request.SessionID = journal.SessionID
		request.RecallK, request.PackK = 200, 10
		binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
		id := uint64(out.admitted + 1)
		var original researchmemory.RecordedPrediction
		err = s.WithValidatedResearchCandidateAsOfAdmission(ctx, binding, in.AsOf, candidate.Features, candidate.Baseline, request, func() error {
			_, retry, err := durable.AdmitBound(ctx, id, candidate.Features, candidate.Baseline, in.AsOf, binding)
			if err != nil || retry {
				return fmt.Errorf("durable admission retry=%v: %w", retry, err)
			}
			original, err = durable.Admission(ctx, id)
			if err != nil {
				return err
			}
			return s.ValidateResearchAdmission(ctx, original, request)
		})
		if err != nil {
			out.err = fmt.Errorf("admission %d at=%s previous_available=%s: %w", id, in.AsOf.Format(time.RFC3339Nano), previousAvailable.Format(time.RFC3339Nano), err)
			return out
		}
		labelAt := time.Now()
		err = s.WithValidatedResearchAsOfAdmission(ctx, original, request, func() error {
			_, err := durable.Feedback(ctx, id, id%3 != 0, in.AsOf.Add(time.Second))
			return err
		})
		if err != nil {
			out.err = fmt.Errorf("feedback %d: %w", id, err)
			return out
		}
		feedbackReturned := time.Now()
		if err := durable.WaitProcessed(ctx, id); err != nil {
			out.err = fmt.Errorf("publication %d: %w", id, err)
			return out
		}
		publishedAt := time.Now()
		completed, failed, _, _ := durable.Counts()
		if failed != 0 || completed < id {
			out.err = fmt.Errorf("label %d not successfully published: completed=%d failed=%d", id, completed, failed)
			return out
		}
		out.admitted++
		out.frontierAge = append(out.frontierAge, publishedAt.Sub(in.queuedAt).Nanoseconds())
		out.feedbackAge = append(out.feedbackAge, publishedAt.Sub(labelAt).Nanoseconds())
		out.tapWait = append(out.tapWait, takenAt.Sub(in.queuedAt).Nanoseconds())
		out.preFeedback = append(out.preFeedback, labelAt.Sub(takenAt).Nanoseconds())
		out.feedbackGuard = append(out.feedbackGuard, feedbackReturned.Sub(labelAt).Nanoseconds())
		out.publicationWait = append(out.publicationWait, publishedAt.Sub(feedbackReturned).Nanoseconds())
		previousAvailable = in.AsOf.Add(time.Second)
		if id == 1 {
			out.firstRequest, out.firstOriginal = request, original
		}
	}
}

type researchRecallLiveTrialV26 struct {
	offer, call, queue, gaps                             []int64
	frontierAge, feedbackAge                             []int64
	tapWait, preFeedback, feedbackGuard, publicationWait []int64
	preFeedbackPartsV32                                  []researchPreFeedbackPartsV32
	admissionPartsV34                                    []researchAdmissionPartsV34
	writerOffer, writerCall                              []int64
	writerTotal                                          int64
	writerStartedOverlap                                 int
	writes, overlap, labels                              int
	dropped                                              uint64
}

func runResearchRecallLiveV26(t *testing.T, enabled bool) researchRecallLiveTrialV26 {
	return runResearchRecallLiveWithConsumerV28(t, enabled, consumeResearchLiveV26, false)
}

func runResearchRecallLiveWithConsumerV28(t *testing.T, enabled bool, consume func(context.Context, *Service, *ResearchFrontierTap, *researchmemory.Durable) researchLiveConsumerV26, allowInterleavedLedger bool) researchRecallLiveTrialV26 {
	return runResearchRecallLiveCadenceV30(t, enabled, consume, allowInterleavedLedger, 8*time.Millisecond)
}

func runResearchRecallLiveCadenceV30(t *testing.T, enabled bool, consume func(context.Context, *Service, *ResearchFrontierTap, *researchmemory.Durable) researchLiveConsumerV26, allowInterleavedLedger bool, offerGap time.Duration) researchRecallLiveTrialV26 {
	return runResearchRecallLiveWithWritesV36(t, enabled, consume, allowInterleavedLedger, offerGap, true)
}

func runResearchRecallLiveWithWritesV36(t *testing.T, enabled bool, consume func(context.Context, *Service, *ResearchFrontierTap, *researchmemory.Durable) researchLiveConsumerV26, allowInterleavedLedger bool, offerGap time.Duration, withWriter bool) researchRecallLiveTrialV26 {
	return runResearchRecallLiveWithPolicyV37(t, enabled, consume, allowInterleavedLedger, offerGap, withWriter, nil, false)
}

func runResearchRecallLiveWithPolicyV37(t *testing.T, enabled bool, consume func(context.Context, *Service, *ResearchFrontierTap, *researchmemory.Durable) researchLiveConsumerV26, allowInterleavedLedger bool, offerGap time.Duration, withWriter bool, gate researchSchedulingGateV38, measureWriter bool) researchRecallLiveTrialV26 {
	t.Helper()
	if offerGap <= 0 {
		t.Fatal("nonpositive research offer gap")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	s, oldTap := motionExitService(t, root, true)
	defer s.Close()
	oldTap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	expected := make(map[string]bool, 200)
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	expected["seed"] = true
	for i := 1; i < 200; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	journalPath := filepath.Join(root, "journals.sqlite")
	sqlite, err := openResearchSQLiteJournalStore(s.store, journalPath, "live-owner", true)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.closeJournal()
	s.store = researchBoundSQLiteV26{sqlite}
	s.config.DefaultRecallK, s.config.DefaultPackK = 200, 10
	var tap *ResearchFrontierTap
	var durable *researchmemory.Durable
	var consumerDone chan researchLiveConsumerV26
	memoryPath := filepath.Join(root, "memory.sqlite")
	if enabled {
		tap, err = NewResearchFrontierTap("tenant-a", 64)
		if err != nil {
			t.Fatal(err)
		}
		s.config.ResearchFrontier = tap
		durable, err = researchmemory.OpenDurable(ctx, memoryPath, "tenant-a", "live", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		defer durable.Close()
		consumerDone = make(chan researchLiveConsumerV26, 1)
		go func() { consumerDone <- consume(ctx, s, tap, durable) }()
	} else {
		s.config.ResearchFrontier = nil
	}

	type job struct {
		i       int
		offered time.Time
	}
	type sample struct {
		offer, call, queue int64
		err                error
	}
	jobs := make(chan job, 192)
	offered := make(chan time.Time, 192)
	samples := make(chan sample, 192)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for work := range jobs {
				request := motionExitRequest(s, now.Add(time.Duration(2*work.i)*time.Second))
				request.SessionID = fmt.Sprintf("profile-%d", work.i)
				request.RecallK, request.PackK = 200, 10
				start := time.Now()
				packet, callErr := s.Recall(ctx, request)
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
				samples <- sample{offer: end.Sub(work.offered).Nanoseconds(), call: end.Sub(start).Nanoseconds(), queue: start.Sub(work.offered).Nanoseconds(), err: callErr}
			}
		}()
	}
	probeDone := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(offerGap)
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
	var timedWriterDone <-chan researchWriterTimingV37
	if withWriter {
		if measureWriter {
			timedWriterDone = motionLoadWriterTimedV37(ctx, s, now, &active, gate)
		} else {
			writerDone = motionLoadWriter(ctx, s, now, &active)
		}
	}
	var result researchRecallLiveTrialV26
	for item := range samples {
		if item.err != nil {
			t.Fatalf("full Recall: %v", item.err)
		}
		result.offer = append(result.offer, item.offer)
		result.call = append(result.call, item.call)
		result.queue = append(result.queue, item.queue)
	}
	<-probeDone
	active.Store(false)
	if withWriter {
		if measureWriter {
			writer := <-timedWriterDone
			if writer.err != nil {
				t.Fatal("timed future writer", writer.err)
			}
			result.writes, result.overlap = writer.writes, writer.offeredOverlap
			result.writerStartedOverlap = writer.startedOverlap
			result.writerOffer, result.writerCall = writer.offerNS, writer.callNS
			result.writerTotal = writer.totalNS
		} else {
			writer := <-writerDone
			if writer.err != nil {
				t.Fatal("future writer", writer.err)
			}
			result.writes, result.overlap = writer.writes, writer.overlap
		}
	}
	var times []time.Time
	for at := range offered {
		times = append(times, at)
	}
	expectedWrites := 0
	if withWriter {
		expectedWrites = 256
	}
	if len(result.offer) != 192 || len(times) != 192 || result.writes != expectedWrites || (withWriter && result.overlap == 0) || (!withWriter && result.overlap != 0) {
		t.Fatalf("incomplete serving offers=%d timestamps=%d writes=%d overlap=%d", len(result.offer), len(times), result.writes, result.overlap)
	}
	for i := 1; i < len(times); i++ {
		gap := times[i].Sub(times[i-1]).Nanoseconds()
		if gap <= 0 {
			t.Fatal("nonmonotonic offer clock")
		}
		result.gaps = append(result.gaps, gap)
	}
	if enabled {
		tap.Close()
		consumer := <-consumerDone
		result.dropped = tap.Dropped()
		result.labels = consumer.admitted
		result.frontierAge, result.feedbackAge = consumer.frontierAge, consumer.feedbackAge
		result.tapWait, result.preFeedback = consumer.tapWait, consumer.preFeedback
		result.preFeedbackPartsV32 = consumer.preFeedbackPartsV32
		result.admissionPartsV34 = consumer.admissionPartsV34
		result.feedbackGuard, result.publicationWait = consumer.feedbackGuard, consumer.publicationWait
		if consumer.err != nil || consumer.seen != 192 || consumer.admitted != 64 || result.dropped != 0 {
			t.Fatalf("live consumer seen=%d labels=%d dropped=%d err=%v", consumer.seen, consumer.admitted, result.dropped, consumer.err)
		}
		if completed, failed, pending, queued := durable.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
			t.Fatalf("live completion completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
		}
		putTemporalFixture(t, s, "visible", now)
		called := false
		err := s.WithValidatedResearchAsOfAdmission(ctx, consumer.firstOriginal, consumer.firstRequest, func() error {
			called = true
			return nil
		})
		if err == nil || called {
			t.Fatal("visible mutation failed to invalidate old admission")
		}
		if err := durable.Close(); err != nil {
			t.Fatal(err)
		}
		log, err := researchledger.Open(memoryPath)
		if err != nil {
			t.Fatal(err)
		}
		rows, readErr := log.ReadAfter(ctx, 0, 256)
		if readErr != nil || len(rows) != 128 {
			t.Fatalf("durable ledger rows=%d read=%v", len(rows), readErr)
		}
		if allowInterleavedLedger {
			validateResearchLiveInterleavedV28(t, rows)
		} else {
			for i := 0; i < 64; i++ {
				admit, terminal := rows[2*i], rows[2*i+1]
				if admit.Kind != "admit" || terminal.Kind != "feedback" || admit.Key != terminal.Key || admit.Key.Event != fmt.Sprint(i+1) {
					t.Fatalf("wrong ledger order at %d", i)
				}
				var original researchmemory.RecordedPrediction
				var feedback researchmemory.RecordedFeedback
				if err := json.Unmarshal(admit.Payload, &original); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(terminal.Payload, &feedback); err != nil {
					t.Fatal(err)
				}
				if original.Binding == nil || original.Binding.EventID != "seed" || feedback.ID != uint64(i+1) || feedback.Useful != (uint64(i+1)%3 != 0) {
					t.Fatalf("wrong persisted binding or label at %d", i)
				}
			}
		}
		replay, err := researchmemory.ResumeBackground(ctx, log, "tenant-a", "live", 1, 42, 128)
		if err != nil {
			t.Fatal(err)
		}
		if completed, failed, pending, queued := replay.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
			t.Fatalf("replay completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
		}
		replay.Close()
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := sqlite.closeJournal(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openResearchSQLiteJournalStore(sqlite.EventStore, journalPath, "live-owner", false)
	if err != nil {
		t.Fatal(err)
	}
	count, countErr := reopened.count(ctx)
	closeErr := reopened.closeJournal()
	if countErr != nil || closeErr != nil || count != 192 {
		t.Fatalf("durable Recall journal count=%d read=%v close=%v", count, countErr, closeErr)
	}
	return result
}

func TestResearchRecallLiveLearningV26(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_LIVE_V26") != "1" {
		t.Skip("opt-in full Recall plus durable live learning")
	}
	type arm struct {
		offer, frontierAge, feedbackAge []int64
	}
	all := map[bool]*arm{false: {}, true: {}}
	for trial := range 3 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, enabled := range order {
			r := runResearchRecallLiveV26(t, enabled)
			a := all[enabled]
			a.offer = append(a.offer, r.offer...)
			a.frontierAge = append(a.frontierAge, r.frontierAge...)
			a.feedbackAge = append(a.feedbackAge, r.feedbackAge...)
			t.Logf("trial=%d enabled=%v offers=%d writes=%d labels=%d dropped=%d gap_p50=%s gap_p99=%s offer_p99=%s call_p99=%s queue_p99=%s frontier_age_p99=%s feedback_age_p99=%s", trial, enabled, len(r.offer), r.writes, r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.gaps, .99),
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99),
				researchLiveAgeV26(r.frontierAge), researchLiveAgeV26(r.feedbackAge))
			if enabled && researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
				t.Errorf("frozen v26 trial latency gate failed: trial=%d", trial)
			}
		}
	}
	if len(all[false].offer) != 576 || len(all[true].offer) != 576 || len(all[true].frontierAge) != 192 || len(all[true].feedbackAge) != 192 {
		t.Fatal("incomplete aggregate")
	}
	offP99 := researchDurableLoadPercentile(all[false].offer, .99)
	onP99 := researchDurableLoadPercentile(all[true].offer, .99)
	frontierP99 := researchDurableLoadPercentile(all[true].frontierAge, .99)
	feedbackP99 := researchDurableLoadPercentile(all[true].feedbackAge, .99)
	t.Logf("pooled off_p99=%s on_p99=%s ratio=%.3f labels=192 frontier_age_p99=%s feedback_age_p99=%s", offP99, onP99, float64(onP99)/float64(offP99), frontierP99, feedbackP99)
	if onP99 >= 100*time.Millisecond || float64(onP99) > 1.10*float64(offP99) || frontierP99 >= 250*time.Millisecond || feedbackP99 >= 250*time.Millisecond {
		t.Error("frozen v26 full integration component failed")
	}
}

func researchLiveAgeV26(values []int64) string {
	if len(values) == 0 {
		return "not_applicable"
	}
	return researchDurableLoadPercentile(values, .99).String()
}
