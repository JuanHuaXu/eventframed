package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type researchDurableMixedTrial struct {
	recallNS, feedbackNS, liveAgeNS []int64
	writes, overlaps                int
	replayNS                        int64
}

type researchDurableAgeOrigin uint8

const (
	researchDurableNoAge researchDurableAgeOrigin = iota
	researchDurablePostReturnAge
	researchDurableOfferedAge
)

func researchDurableLoadPercentile(values []int64, fraction float64) time.Duration {
	copyValues := append([]int64(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool { return copyValues[i] < copyValues[j] })
	index := int(math.Ceil(fraction*float64(len(copyValues)))) - 1
	return time.Duration(copyValues[index])
}

func researchDurableMixedLoadTrial(t *testing.T, futureWriter bool) researchDurableMixedTrial {
	return researchDurableMixedLoadTrialAt(t, futureWriter, researchDurableNoAge)
}

func researchDurableMixedLoadTrialAt(t *testing.T, futureWriter bool, ageOrigin researchDurableAgeOrigin) researchDurableMixedTrial {
	t.Helper()
	observeLive := ageOrigin != researchDurableNoAge
	s, _, now := temporalBridgeFixture(t, true)
	wrapped, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		t.Fatal(err)
	}
	s.store = wrapped
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	path := t.TempDir() + "/mixed.sqlite"
	durable, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "mixed", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer durable.Close()
	preview := researchmemory.New(1, 42)
	result := researchDurableMixedTrial{recallNS: make([]int64, 0, 64), feedbackNS: make([]int64, 0, 64)}
	type submittedLabel struct {
		id uint64
		at time.Time
	}
	type observedLabels struct {
		ages []int64
		err  error
	}
	var submissions chan submittedLabel
	var observerDone chan observedLabels
	if observeLive {
		observerCtx, stopObserver := context.WithCancel(ctx)
		submissions = make(chan submittedLabel, 64)
		observerDone = make(chan observedLabels, 1)
		go func() {
			outcome := observedLabels{ages: make([]int64, 0, 64)}
			for submitted := range submissions {
				if err := durable.WaitProcessed(observerCtx, submitted.id); err != nil {
					outcome.err = err
					break
				}
				completed, failed, _, _ := durable.Counts()
				if failed != 0 || completed < submitted.id {
					outcome.err = fmt.Errorf("live label %d not successfully published: completed=%d failed=%d", submitted.id, completed, failed)
					break
				}
				outcome.ages = append(outcome.ages, time.Since(submitted.at).Nanoseconds())
			}
			observerDone <- outcome
		}()
		defer func() {
			if submissions != nil {
				stopObserver()
				close(submissions)
				<-observerDone
			}
		}()
	}

	var active atomic.Bool
	type writerResult struct {
		writes, overlaps int
		err              error
	}
	var writerDone chan writerResult
	var writerCancel context.CancelFunc
	if futureWriter {
		writerCtx, stop := context.WithCancel(ctx)
		writerCancel = stop
		writerDone = make(chan writerResult, 1)
		started := make(chan struct{})
		go func() {
			var outcome writerResult
			defer func() { writerDone <- outcome }()
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for i := 0; i < 256; i++ {
				if i > 0 {
					select {
					case <-writerCtx.Done():
						return
					case <-ticker.C:
					}
				}
				ev := testutil.Event(fmt.Sprintf("future-%d", i), "public query fixture", now.Add(time.Hour))
				wasActive := active.Load()
				_, err := s.Observe(writerCtx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
				if err != nil {
					outcome.err = err
					if i == 0 {
						close(started)
					}
					return
				}
				outcome.writes++
				if wasActive {
					outcome.overlaps++
				}
				if i == 0 {
					close(started)
				}
			}
		}()
		<-started
		defer func() {
			if writerDone != nil {
				writerCancel()
				<-writerDone
			}
		}()
	}

	for i := 0; i < 64; i++ {
		queryAt := now.Add(time.Duration(2*i) * time.Second)
		request := model.RecallRequest{
			ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a",
			SessionID: fmt.Sprintf("mixed-%d", i), Query: "public query fixture",
			AsOf: queryAt, RecallK: 3, PackK: 2, TokenBudget: 1000,
		}
		active.Store(true)
		start := time.Now()
		if _, err := s.Recall(ctx, request); err != nil {
			t.Fatal(err)
		}
		result.recallNS = append(result.recallNS, time.Since(start).Nanoseconds())
		in, err := s.config.ResearchFrontier.Take(ctx)
		if err != nil || len(in.Candidates) != 1 || in.Candidates[0].EventID != "seed" {
			t.Fatal("future event entered the as-of frontier", in, err)
		}
		candidate := in.Candidates[0]
		prediction, err := preview.Predict(candidate.Features, candidate.Baseline, 1, in.AsOf)
		if err != nil || prediction.ID != uint64(i+1) {
			t.Fatal("preview identity", prediction, err)
		}
		envelope, err := preview.Record(prediction.ID)
		if err != nil {
			t.Fatal(err)
		}
		binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
		envelope.Binding = &binding
		var original researchmemory.RecordedPrediction
		err = s.WithValidatedResearchAsOfAdmission(ctx, envelope, request, func() error {
			_, retry, err := durable.AdmitBound(ctx, uint64(i+1), candidate.Features, candidate.Baseline, in.AsOf, binding)
			if err != nil || retry {
				return fmt.Errorf("durable admission retry=%v: %w", retry, err)
			}
			original, err = durable.Admission(ctx, uint64(i+1))
			if err != nil {
				return err
			}
			return s.ValidateResearchAdmission(ctx, original, request)
		})
		if err != nil || original.Binding == nil || *original.Binding != binding {
			t.Fatal("bound durable admission", err, original)
		}
		start = time.Now()
		if ageOrigin == researchDurableOfferedAge {
			submissions <- submittedLabel{id: uint64(i + 1), at: start}
		}
		err = s.WithValidatedResearchAsOfAdmission(ctx, original, request, func() error {
			_, err := durable.Feedback(ctx, uint64(i+1), i%3 != 0, queryAt.Add(time.Second))
			return err
		})
		result.feedbackNS = append(result.feedbackNS, time.Since(start).Nanoseconds())
		active.Store(false)
		if err != nil {
			t.Fatalf("guarded feedback %d: %v", i, err)
		}
		if ageOrigin == researchDurablePostReturnAge {
			submissions <- submittedLabel{id: uint64(i + 1), at: time.Now()}
		}
	}
	if observeLive {
		close(submissions)
		outcome := <-observerDone
		submissions = nil
		if outcome.err != nil || len(outcome.ages) != 64 {
			t.Fatal("live completion observer", outcome.err, len(outcome.ages))
		}
		result.liveAgeNS = outcome.ages
		if completed, failed, pending, queued := durable.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
			t.Fatalf("live completion accounting completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
		}
	}
	if futureWriter {
		writerCancel()
		outcome := <-writerDone
		writerDone = nil
		if outcome.err != nil && !errors.Is(outcome.err, context.Canceled) {
			t.Fatal("future writer", outcome.err)
		}
		result.writes, result.overlaps = outcome.writes, outcome.overlaps
		if result.writes == 0 || result.overlaps == 0 {
			t.Fatal("future writer did not overlap learning", result)
		}
	}
	if err := durable.Close(); err != nil {
		t.Fatal(err)
	}
	log, err := researchledger.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	rows, err := log.ReadAfter(ctx, 0, 256)
	if err != nil || len(rows) != 128 {
		t.Fatal("incomplete durable ledger", len(rows), err)
	}
	seen := make(map[string]bool, 64)
	for i := 0; i < 64; i++ {
		admit, terminal := rows[2*i], rows[2*i+1]
		if admit.Kind != "admit" || terminal.Kind != "feedback" || admit.Key != terminal.Key || admit.Key.Event != fmt.Sprint(i+1) {
			t.Fatalf("wrong terminal ordering at %d", i)
		}
		var original researchmemory.RecordedPrediction
		var label researchmemory.RecordedFeedback
		if err := json.Unmarshal(admit.Payload, &original); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(terminal.Payload, &label); err != nil {
			t.Fatal(err)
		}
		if original.Binding == nil || original.Binding.Tenant != "tenant-a" || original.Binding.EventID != "seed" || original.Binding.JournalID == "" || seen[original.Binding.JournalID] ||
			!original.At.Equal(now.Add(time.Duration(2*i)*time.Second)) || label.ID != uint64(i+1) || label.Useful != (i%3 != 0) || !label.Available.Equal(original.At.Add(time.Second)) {
			t.Fatalf("wrong persisted source or label at %d", i)
		}
		seen[original.Binding.JournalID] = true
	}
	start := time.Now()
	restored, err := researchmemory.ResumeBackground(ctx, log, "tenant-a", "mixed", 1, 42, 128)
	result.replayNS = time.Since(start).Nanoseconds()
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if completed, failed, pending, queued := restored.Counts(); completed != 64 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatalf("replay accounting completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
	}
	return result
}

func TestResearchGuardedDurableMixedWriteLoadV2(t *testing.T) {
	for _, writer := range []bool{false, true} {
		name := "quiet"
		if writer {
			name = "future-writer"
		}
		t.Run(name, func(t *testing.T) {
			var recall, feedback []int64
			writes, overlaps := 0, 0
			var replayNS int64
			for trial := 0; trial < 3; trial++ {
				r := researchDurableMixedLoadTrial(t, writer)
				recall = append(recall, r.recallNS...)
				feedback = append(feedback, r.feedbackNS...)
				writes += r.writes
				overlaps += r.overlaps
				replayNS += r.replayNS
			}
			t.Logf("arm=%s calls=%d writes=%d overlaps=%d recall_p50=%s recall_p95=%s recall_p99=%s recall_max=%s feedback_p50=%s feedback_p95=%s feedback_p99=%s feedback_max=%s replay_mean=%s",
				name, len(recall), writes, overlaps,
				researchDurableLoadPercentile(recall, .5), researchDurableLoadPercentile(recall, .95), researchDurableLoadPercentile(recall, .99), researchDurableLoadPercentile(recall, 1),
				researchDurableLoadPercentile(feedback, .5), researchDurableLoadPercentile(feedback, .95), researchDurableLoadPercentile(feedback, .99), researchDurableLoadPercentile(feedback, 1),
				time.Duration(replayNS/3))
			if writer && os.Getenv("EVENTFRAME_RESEARCH_PERF_GATE") == "1" && researchDurableLoadPercentile(recall, .99) >= 100*time.Millisecond {
				t.Error("predeclared finite writer-arm recall p99 gate failed")
			}
		})
	}
}
