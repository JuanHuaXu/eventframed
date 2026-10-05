package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type researchIndexedFrontierV29 struct {
	index        int
	frontier     researchOrderedFrontierV28
	lookupDoneAt time.Time
}

type researchSelectedV33 struct {
	frontier                 researchOrderedFrontierV28
	lookupDoneAt, receivedAt time.Time
}

type researchDrainResultV29 struct {
	seen int
	err  error
}

type researchPreFeedbackPartsV32 struct {
	beforeAdmission, admissionGuard, feedbackQueue int64
	journalLookup, selectedHandoff, orderedWait    int64
}

type researchAdmissionPartsV34 struct {
	sourceValidation, durableAdmit, durableRead int64
	postRecordValidation, guardExit             int64
	guardEntry, candidateValidation             int64
}

type researchAdmissionV32 struct {
	researchOrderedAdmissionV28
	lookupDoneAt, receivedAt      time.Time
	admissionStart, admissionDone time.Time
	admissionPartsV34             researchAdmissionPartsV34
}

func drainResearchLiveV29(ctx context.Context, s *Service, tap *ResearchFrontierTap, selected chan<- researchIndexedFrontierV29) researchDrainResultV29 {
	defer close(selected)
	var out researchDrainResultV29
	seenIndices := make(map[int]bool, 192)
	baseAt := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for out.seen < 192 {
		in, err := tap.Take(ctx)
		if err != nil {
			out.err = fmt.Errorf("take committed frontier %d: %w", out.seen, err)
			return out
		}
		takenAt := time.Now()
		journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
		if err != nil {
			out.err = fmt.Errorf("lookup committed frontier: %w", err)
			return out
		}
		var index int
		n, err := fmt.Sscanf(journal.SessionID, "profile-%d", &index)
		if err != nil || n != 1 || index < 0 || index >= 192 || journal.SessionID != fmt.Sprintf("profile-%d", index) || !in.AsOf.Equal(baseAt.Add(time.Duration(2*index)*time.Second)) || seenIndices[index] {
			out.err = fmt.Errorf("invalid or duplicate committed frontier: session=%q at=%s", journal.SessionID, in.AsOf.Format(time.RFC3339Nano))
			return out
		}
		seenIndices[index] = true
		out.seen++
		if index%3 == 0 {
			item := researchIndexedFrontierV29{index: index, frontier: researchOrderedFrontierV28{in, journal.SessionID, takenAt}, lookupDoneAt: time.Now()}
			select {
			case selected <- item:
			case <-ctx.Done():
				out.err = ctx.Err()
				return out
			}
		}
	}
	return out
}

func consumeResearchLiveDecoupledV29(parent context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
	return consumeResearchLiveDecoupledWithGateV37(parent, s, tap, durable, nil)
}

func consumeResearchLiveDecoupledWithGateV37(parent context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable, gate researchSchedulingGateV38) researchLiveConsumerV26 {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	selected := make(chan researchIndexedFrontierV29, 32)
	drainDone := make(chan researchDrainResultV29, 1)
	go func() { drainDone <- drainResearchLiveV29(ctx, s, tap, selected) }()

	admitted := make(chan researchAdmissionV32, 16)
	feedbackDone := make(chan researchLiveConsumerV26, 1)
	go func() {
		var result researchLiveConsumerV26
		for item := range admitted {
			labelAt := time.Now()
			feedback := func() error {
				return s.WithValidatedResearchAsOfAdmission(ctx, item.original, item.request, func() error {
					_, err := durable.Feedback(ctx, item.id, item.id%3 != 0, item.frontier.observation.AsOf.Add(time.Second))
					return err
				})
			}
			var err error
			if gate != nil {
				gateCtx := context.WithValue(ctx, researchDeadlineKeyV38{}, item.frontier.observation.queuedAt.Add(200*time.Millisecond))
				err = gate.withAdmission(gateCtx, feedback)
			} else {
				err = feedback()
			}
			if err != nil {
				result.err = fmt.Errorf("decoupled feedback %d: %w", item.id, err)
				cancel()
				break
			}
			feedbackReturned := time.Now()
			if err := durable.WaitProcessed(ctx, item.id); err != nil {
				result.err = fmt.Errorf("decoupled publication %d: %w", item.id, err)
				cancel()
				break
			}
			publishedAt := time.Now()
			completed, failed, _, _ := durable.Counts()
			if failed != 0 || completed < item.id {
				result.err = fmt.Errorf("decoupled label %d not published: completed=%d failed=%d", item.id, completed, failed)
				cancel()
				break
			}
			result.admitted++
			result.frontierAge = append(result.frontierAge, publishedAt.Sub(item.frontier.observation.queuedAt).Nanoseconds())
			result.feedbackAge = append(result.feedbackAge, publishedAt.Sub(labelAt).Nanoseconds())
			result.tapWait = append(result.tapWait, item.frontier.takenAt.Sub(item.frontier.observation.queuedAt).Nanoseconds())
			result.preFeedback = append(result.preFeedback, labelAt.Sub(item.frontier.takenAt).Nanoseconds())
			result.preFeedbackPartsV32 = append(result.preFeedbackPartsV32, researchPreFeedbackPartsV32{
				beforeAdmission: labelDurationV32(item.admissionStart, item.frontier.takenAt),
				admissionGuard:  labelDurationV32(item.admissionDone, item.admissionStart),
				feedbackQueue:   labelDurationV32(labelAt, item.admissionDone),
				journalLookup:   labelDurationV32(item.lookupDoneAt, item.frontier.takenAt),
				selectedHandoff: labelDurationV32(item.receivedAt, item.lookupDoneAt),
				orderedWait:     labelDurationV32(item.admissionStart, item.receivedAt),
			})
			result.admissionPartsV34 = append(result.admissionPartsV34, item.admissionPartsV34)
			result.feedbackGuard = append(result.feedbackGuard, feedbackReturned.Sub(labelAt).Nanoseconds())
			result.publicationWait = append(result.publicationWait, publishedAt.Sub(feedbackReturned).Nanoseconds())
		}
		feedbackDone <- result
	}()

	var out researchLiveConsumerV26
	pending := make(map[int]researchSelectedV33)
	for index := 0; index < 192; index += 3 {
		for {
			if _, ok := pending[index]; ok {
				break
			}
			item, ok := <-selected
			if !ok {
				out.err = fmt.Errorf("await selected frontier %d: selected drainer closed", index)
				break
			}
			pending[item.index] = researchSelectedV33{frontier: item.frontier, lookupDoneAt: item.lookupDoneAt, receivedAt: time.Now()}
			if len(pending) > 32 {
				out.err = fmt.Errorf("bounded reorder capacity exceeded: %d", len(pending))
				break
			}
		}
		if out.err != nil {
			break
		}
		selectedFrontier := pending[index]
		delete(pending, index)
		frontier := selectedFrontier.frontier
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
		admissionStart := time.Now()
		guardTiming := &researchGuardTimingV35{}
		admissionCtx := context.WithValue(ctx, researchGuardTimingKeyV35{}, guardTiming)
		var callbackEntryAt, admitEndAt, readEndAt, validateEndAt time.Time
		admit := func() error {
			return s.WithValidatedResearchCandidateAsOfAdmission(admissionCtx, binding, in.AsOf, candidate.Features, candidate.Baseline, request, func() error {
				callbackEntryAt = time.Now()
				_, retry, err := durable.AdmitBound(ctx, id, candidate.Features, candidate.Baseline, in.AsOf, binding)
				admitEndAt = time.Now()
				if err != nil || retry {
					return fmt.Errorf("decoupled durable admission retry=%v: %w", retry, err)
				}
				original, err = durable.Admission(ctx, id)
				readEndAt = time.Now()
				if err != nil {
					return err
				}
				err = s.ValidateResearchAdmission(ctx, original, request)
				validateEndAt = time.Now()
				return err
			})
		}
		var err error
		if gate != nil {
			gateCtx := context.WithValue(ctx, researchDeadlineKeyV38{}, in.queuedAt.Add(200*time.Millisecond))
			err = gate.withAdmission(gateCtx, admit)
		} else {
			err = admit()
		}
		if err != nil {
			out.err = fmt.Errorf("decoupled admission %d at=%s: %w", id, in.AsOf.Format(time.RFC3339Nano), err)
			break
		}
		admissionDone := time.Now()
		admissionParts := researchAdmissionPartsV34{
			sourceValidation:     labelDurationV32(callbackEntryAt, admissionStart),
			guardEntry:           labelDurationV32(guardTiming.enteredAt, admissionStart),
			candidateValidation:  labelDurationV32(callbackEntryAt, guardTiming.enteredAt),
			durableAdmit:         labelDurationV32(admitEndAt, callbackEntryAt),
			durableRead:          labelDurationV32(readEndAt, admitEndAt),
			postRecordValidation: labelDurationV32(validateEndAt, readEndAt),
			guardExit:            labelDurationV32(admissionDone, validateEndAt),
		}
		if id == 1 {
			out.firstRequest, out.firstOriginal = request, original
		}
		select {
		case admitted <- researchAdmissionV32{researchOrderedAdmissionV28{id, frontier, request, original}, selectedFrontier.lookupDoneAt, selectedFrontier.receivedAt, admissionStart, admissionDone, admissionParts}:
		case <-ctx.Done():
			out.err = ctx.Err()
		}
		if out.err != nil {
			break
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
	out.preFeedbackPartsV32 = completed.preFeedbackPartsV32
	out.admissionPartsV34 = completed.admissionPartsV34
	out.feedbackGuard, out.publicationWait = completed.feedbackGuard, completed.publicationWait
	if out.err == nil {
		for range selected {
			out.err = fmt.Errorf("unexpected selected frontier after final admission")
			cancel()
			break
		}
	}
	drain := <-drainDone
	out.seen = drain.seen
	if out.err == nil {
		out.err = drain.err
	}
	if out.err == nil && (len(pending) != 0 || out.seen != 192 || out.admitted != 64) {
		out.err = fmt.Errorf("decoupled completion mismatch: buffered=%d seen=%d labels=%d", len(pending), out.seen, out.admitted)
	}
	return out
}

func labelDurationV32(end, start time.Time) int64 { return end.Sub(start).Nanoseconds() }

func TestResearchRecallDecoupledDrainV29(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_DECOUPLED_V29") != "1" {
		t.Skip("opt-in decoupled live-learning pipeline")
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
			r := runResearchRecallLiveWithConsumerV28(t, enabled, consumeResearchLiveDecoupledV29, true)
			all[enabled].offers = append(all[enabled].offers, r.offer...)
			all[enabled].frontierAge = append(all[enabled].frontierAge, r.frontierAge...)
			all[enabled].feedbackAge = append(all[enabled].feedbackAge, r.feedbackAge...)
			t.Logf("trial=%d enabled=%v offers=%d labels=%d drops=%d gap_p50=%s gap_p99=%s offer_p99=%s queue_p99=%s frontier_age_p99=%s feedback_age_p99=%s", trial, enabled, len(r.offer), r.labels, r.dropped,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.gaps, .99),
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.queue, .99), researchLiveAgeV26(r.frontierAge), researchLiveAgeV26(r.feedbackAge))
			if enabled {
				checkResearchLivePhasesV27(t, r)
				if researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
					t.Errorf("frozen v29 trial serving gate failed: trial=%d", trial)
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
		t.Error("frozen v29 decoupled-drain component failed")
	}
}

func TestResearchRecallDecoupledDrainV29FocusedRace(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_DECOUPLED_V29_RACE") != "1" {
		t.Skip("opt-in decoupled-drain race check")
	}
	r := runResearchRecallLiveWithConsumerV28(t, true, consumeResearchLiveDecoupledV29, true)
	checkResearchLivePhasesV27(t, r)
}
