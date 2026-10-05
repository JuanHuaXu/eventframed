package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type researchOrderedFrontierV28 struct {
	observation ResearchFrontierObservation
	sessionID   string
	takenAt     time.Time
}

type researchOrderedAdmissionV28 struct {
	id       uint64
	frontier researchOrderedFrontierV28
	request  model.RecallRequest
	original researchmemory.RecordedPrediction
}

func consumeResearchLiveOrderedV28(parent context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	admitted := make(chan researchOrderedAdmissionV28, 16)
	feedbackDone := make(chan researchLiveConsumerV26, 1)
	go func() {
		var result researchLiveConsumerV26
		for item := range admitted {
			labelAt := time.Now()
			err := s.WithValidatedResearchAsOfAdmission(ctx, item.original, item.request, func() error {
				_, err := durable.Feedback(ctx, item.id, item.id%3 != 0, item.frontier.observation.AsOf.Add(time.Second))
				return err
			})
			if err != nil {
				result.err = fmt.Errorf("ordered feedback %d: %w", item.id, err)
				cancel()
				break
			}
			feedbackReturned := time.Now()
			if err := durable.WaitProcessed(ctx, item.id); err != nil {
				result.err = fmt.Errorf("ordered publication %d: %w", item.id, err)
				cancel()
				break
			}
			publishedAt := time.Now()
			completed, failed, _, _ := durable.Counts()
			if failed != 0 || completed < item.id {
				result.err = fmt.Errorf("ordered label %d not published: completed=%d failed=%d", item.id, completed, failed)
				cancel()
				break
			}
			result.admitted++
			result.frontierAge = append(result.frontierAge, publishedAt.Sub(item.frontier.observation.queuedAt).Nanoseconds())
			result.feedbackAge = append(result.feedbackAge, publishedAt.Sub(labelAt).Nanoseconds())
			result.tapWait = append(result.tapWait, item.frontier.takenAt.Sub(item.frontier.observation.queuedAt).Nanoseconds())
			result.preFeedback = append(result.preFeedback, labelAt.Sub(item.frontier.takenAt).Nanoseconds())
			result.feedbackGuard = append(result.feedbackGuard, feedbackReturned.Sub(labelAt).Nanoseconds())
			result.publicationWait = append(result.publicationWait, publishedAt.Sub(feedbackReturned).Nanoseconds())
		}
		feedbackDone <- result
	}()

	var out researchLiveConsumerV26
	pending := make(map[int]researchOrderedFrontierV28)
	seenIndices := make(map[int]bool, 192)
	baseAt := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	readOne := func() error {
		in, err := tap.Take(ctx)
		if err != nil {
			return err
		}
		takenAt := time.Now()
		journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
		if err != nil {
			return err
		}
		var index int
		n, err := fmt.Sscanf(journal.SessionID, "profile-%d", &index)
		if err != nil || n != 1 || index < 0 || index >= 192 || journal.SessionID != fmt.Sprintf("profile-%d", index) || !in.AsOf.Equal(baseAt.Add(time.Duration(2*index)*time.Second)) || seenIndices[index] {
			return fmt.Errorf("invalid or duplicate committed frontier: session=%q at=%s", journal.SessionID, in.AsOf.Format(time.RFC3339Nano))
		}
		seenIndices[index] = true
		out.seen++
		if index%3 == 0 {
			pending[index] = researchOrderedFrontierV28{in, journal.SessionID, takenAt}
			if len(pending) > 32 {
				return fmt.Errorf("bounded reorder capacity exceeded: %d", len(pending))
			}
		}
		return nil
	}

	for index := 0; index < 192; index += 3 {
		for {
			if _, ok := pending[index]; ok {
				break
			}
			if err := readOne(); err != nil {
				out.err = fmt.Errorf("await selected frontier %d: %w", index, err)
				break
			}
		}
		if out.err != nil {
			break
		}
		frontier := pending[index]
		delete(pending, index)
		in := frontier.observation
		if len(in.Candidates) != 200 {
			out.err = fmt.Errorf("frontier %d has %d candidates", index, len(in.Candidates))
			break
		}
		var candidate ResearchFrontierCandidate
		found := false
		for _, c := range in.Candidates {
			if c.EventID == "seed" {
				candidate, found = c, true
				break
			}
		}
		if !found {
			out.err = fmt.Errorf("frontier %d lacks seed", index)
			break
		}
		request := motionExitRequest(s, in.AsOf)
		request.SessionID = frontier.sessionID
		request.RecallK, request.PackK = 200, 10
		binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
		id := uint64(index/3 + 1)
		var original researchmemory.RecordedPrediction
		err := s.WithValidatedResearchCandidateAsOfAdmission(ctx, binding, in.AsOf, candidate.Features, candidate.Baseline, request, func() error {
			_, retry, err := durable.AdmitBound(ctx, id, candidate.Features, candidate.Baseline, in.AsOf, binding)
			if err != nil || retry {
				return fmt.Errorf("ordered durable admission retry=%v: %w", retry, err)
			}
			original, err = durable.Admission(ctx, id)
			if err != nil {
				return err
			}
			return s.ValidateResearchAdmission(ctx, original, request)
		})
		if err != nil {
			out.err = fmt.Errorf("ordered admission %d at=%s: %w", id, in.AsOf.Format(time.RFC3339Nano), err)
			break
		}
		if id == 1 {
			out.firstRequest, out.firstOriginal = request, original
		}
		select {
		case admitted <- researchOrderedAdmissionV28{id, frontier, request, original}:
		case <-ctx.Done():
			out.err = ctx.Err()
		}
		if out.err != nil {
			break
		}
	}
	if out.err == nil {
		for out.seen < 192 {
			if err := readOne(); err != nil {
				out.err = fmt.Errorf("drain committed frontiers: %w", err)
				break
			}
		}
	}
	if out.err != nil {
		cancel()
	}
	close(admitted)
	completed := <-feedbackDone
	if out.err == nil {
		out.err = completed.err
	}
	out.admitted = completed.admitted
	out.frontierAge, out.feedbackAge = completed.frontierAge, completed.feedbackAge
	out.tapWait, out.preFeedback = completed.tapWait, completed.preFeedback
	out.feedbackGuard, out.publicationWait = completed.feedbackGuard, completed.publicationWait
	if out.err == nil && (len(pending) != 0 || len(seenIndices) != 192 || out.admitted != 64) {
		out.err = fmt.Errorf("ordered completion mismatch: buffered=%d seen=%d labels=%d", len(pending), len(seenIndices), out.admitted)
	}
	return out
}

func validateResearchLiveInterleavedV28(t *testing.T, rows []researchledger.Entry) {
	t.Helper()
	admissions := make(map[uint64]researchmemory.RecordedPrediction, 64)
	keys := make(map[uint64]researchledger.Key, 64)
	var lastAdmit, lastFeedback uint64
	var lastAt time.Time
	for _, row := range rows {
		id, err := strconv.ParseUint(row.Key.Event, 10, 64)
		if err != nil || id < 1 || id > 64 {
			t.Fatalf("invalid ledger event key %q", row.Key.Event)
		}
		switch row.Kind {
		case "admit":
			if id != lastAdmit+1 {
				t.Fatalf("nonmonotonic or duplicate admission %d after %d", id, lastAdmit)
			}
			var original researchmemory.RecordedPrediction
			if err := json.Unmarshal(row.Payload, &original); err != nil {
				t.Fatal(err)
			}
			if original.Binding == nil || original.Binding.EventID != "seed" || !original.At.After(lastAt) {
				t.Fatalf("bad source or backdated admission %d", id)
			}
			admissions[id], keys[id] = original, row.Key
			lastAt, lastAdmit = original.At, id
		case "feedback":
			original, ok := admissions[id]
			if !ok || id != lastFeedback+1 || keys[id] != row.Key {
				t.Fatalf("unbound or nonmonotonic feedback %d after %d", id, lastFeedback)
			}
			var label researchmemory.RecordedFeedback
			if err := json.Unmarshal(row.Payload, &label); err != nil {
				t.Fatal(err)
			}
			if label.ID != id || label.Useful != (id%3 != 0) || !label.Available.Equal(original.At.Add(time.Second)) {
				t.Fatalf("wrong label %d", id)
			}
			lastFeedback = id
		default:
			t.Fatalf("unexpected ledger kind %q", row.Kind)
		}
	}
	if lastAdmit != 64 || lastFeedback != 64 || len(admissions) != 64 {
		t.Fatalf("incomplete ledger: admissions=%d terminal=%d", lastAdmit, lastFeedback)
	}
}

func TestResearchRecallOrderedPipelineV28(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_ORDERED_V28") != "1" {
		t.Skip("opt-in ordered live-learning pipeline")
	}
	type arm struct {
		offers, frontierAge, feedbackAge []int64
	}
	all := map[bool]*arm{false: {}, true: {}}
	for trial := range 3 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, enabled := range order {
			r := runResearchRecallLiveWithConsumerV28(t, enabled, consumeResearchLiveOrderedV28, true)
			all[enabled].offers = append(all[enabled].offers, r.offer...)
			all[enabled].frontierAge = append(all[enabled].frontierAge, r.frontierAge...)
			all[enabled].feedbackAge = append(all[enabled].feedbackAge, r.feedbackAge...)
			t.Logf("trial=%d enabled=%v offers=%d labels=%d drops=%d gap_p50=%s gap_p99=%s offer_p99=%s queue_p99=%s frontier_age_p99=%s feedback_age_p99=%s", trial, enabled, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.gaps, .99),
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.queue, .99), researchLiveAgeV26(r.frontierAge), researchLiveAgeV26(r.feedbackAge))
			if enabled {
				checkResearchLivePhasesV27(t, r)
				if researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
					t.Errorf("frozen v28 trial serving gate failed: trial=%d", trial)
				}
			}
		}
	}
	offP99 := researchDurableLoadPercentile(all[false].offers, .99)
	onP99 := researchDurableLoadPercentile(all[true].offers, .99)
	frontierP99 := researchDurableLoadPercentile(all[true].frontierAge, .99)
	feedbackP99 := researchDurableLoadPercentile(all[true].feedbackAge, .99)
	t.Logf("pooled off_p99=%s on_p99=%s ratio=%.3f labels=%d frontier_age_p99=%s feedback_age_p99=%s", offP99, onP99, float64(onP99)/float64(offP99), len(all[true].frontierAge), frontierP99, feedbackP99)
	if len(all[true].frontierAge) != 192 || onP99 >= 100*time.Millisecond || float64(onP99) > 1.10*float64(offP99) || frontierP99 >= 250*time.Millisecond || feedbackP99 >= 250*time.Millisecond {
		t.Error("frozen v28 ordered-pipeline component failed")
	}
}

func TestResearchRecallOrderedPipelineV28FocusedRace(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_ORDERED_V28_RACE") != "1" {
		t.Skip("opt-in ordered-pipeline race check")
	}
	r := runResearchRecallLiveWithConsumerV28(t, true, consumeResearchLiveOrderedV28, true)
	checkResearchLivePhasesV27(t, r)
}
