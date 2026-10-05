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

type researchPriorityGateV37 struct {
	mu                sync.Mutex
	cond              *sync.Cond
	active            bool
	waitingAdmissions int
}

type researchSchedulingGateV38 interface {
	withAdmission(context.Context, func() error) error
	withWriter(context.Context, func() error) error
}

func newResearchPriorityGateV37() *researchPriorityGateV37 {
	g := &researchPriorityGateV37{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

func (g *researchPriorityGateV37) enter(ctx context.Context, admission bool) error {
	g.mu.Lock()
	if admission {
		g.waitingAdmissions++
	}
	stop := context.AfterFunc(ctx, func() {
		g.mu.Lock()
		g.cond.Broadcast()
		g.mu.Unlock()
	})
	defer stop()
	for g.active || (!admission && g.waitingAdmissions > 0) {
		if err := ctx.Err(); err != nil {
			if admission {
				g.waitingAdmissions--
			}
			g.cond.Broadcast()
			g.mu.Unlock()
			return err
		}
		g.cond.Wait()
	}
	if err := ctx.Err(); err != nil {
		if admission {
			g.waitingAdmissions--
		}
		g.cond.Broadcast()
		g.mu.Unlock()
		return err
	}
	if admission {
		g.waitingAdmissions--
	}
	g.active = true
	g.mu.Unlock()
	return nil
}

func (g *researchPriorityGateV37) leave() {
	g.mu.Lock()
	g.active = false
	g.cond.Broadcast()
	g.mu.Unlock()
}

func (g *researchPriorityGateV37) withAdmission(ctx context.Context, work func() error) error {
	if err := g.enter(ctx, true); err != nil {
		return err
	}
	defer g.leave()
	return work()
}

func (g *researchPriorityGateV37) withWriter(ctx context.Context, work func() error) error {
	if err := g.enter(ctx, false); err != nil {
		return err
	}
	defer g.leave()
	return work()
}

type researchWriterTimingV37 struct {
	writes, offeredOverlap, startedOverlap int
	offerNS, callNS                        []int64
	totalNS                                int64
	err                                    error
}

func motionLoadWriterTimedV37(ctx context.Context, s *Service, now time.Time, active *atomic.Bool, gate researchSchedulingGateV38) <-chan researchWriterTimingV37 {
	done := make(chan researchWriterTimingV37, 1)
	go func() {
		var result researchWriterTimingV37
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		var firstOffer time.Time
		for i := range 256 {
			if i > 0 {
				select {
				case <-ctx.Done():
					result.err = ctx.Err()
					done <- result
					return
				case <-ticker.C:
				}
			}
			event := testutil.Event(fmt.Sprintf("future-load-%d", i), "public query fixture", now.Add(time.Hour))
			offeredAt := time.Now()
			if i == 0 {
				firstOffer = offeredAt
			}
			if active.Load() {
				result.offeredOverlap++
			}
			work := func() error {
				startedAt := time.Now()
				if active.Load() {
					result.startedOverlap++
				}
				_, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event})
				result.callNS = append(result.callNS, time.Since(startedAt).Nanoseconds())
				return err
			}
			var err error
			if gate != nil {
				gateCtx := context.WithValue(ctx, researchDeadlineKeyV38{}, offeredAt.Add(30*time.Millisecond))
				err = gate.withWriter(gateCtx, work)
			} else {
				err = work()
			}
			result.offerNS = append(result.offerNS, time.Since(offeredAt).Nanoseconds())
			if err != nil {
				result.err = err
				done <- result
				return
			}
			result.writes++
		}
		result.totalNS = time.Since(firstOffer).Nanoseconds()
		done <- result
	}()
	return done
}

func TestResearchPriorityGateV37OrdersAdmissionFirst(t *testing.T) {
	g := newResearchPriorityGateV37()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := g.enter(ctx, false); err != nil {
		t.Fatal(err)
	}
	order := make(chan string, 2)
	admissionDone := make(chan error, 1)
	go func() {
		admissionDone <- g.withAdmission(ctx, func() error { order <- "admission"; return nil })
	}()
	deadline := time.Now().Add(200 * time.Millisecond)
	for {
		g.mu.Lock()
		waiting := g.waitingAdmissions
		g.mu.Unlock()
		if waiting == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("admission did not queue")
		}
		time.Sleep(time.Millisecond)
	}
	writerDone := make(chan error, 1)
	go func() {
		writerDone <- g.withWriter(ctx, func() error { order <- "writer"; return nil })
	}()
	g.leave()
	select {
	case first := <-order:
		if first != "admission" {
			t.Fatalf("writer overtook queued admission: %s", first)
		}
	case <-ctx.Done():
		t.Fatal("admission priority timed out")
	}
	if err := <-admissionDone; err != nil {
		t.Fatal(err)
	}
	select {
	case second := <-order:
		if second != "writer" {
			t.Fatalf("writer did not resume: %s", second)
		}
	case <-ctx.Done():
		t.Fatal("writer resume timed out")
	}
	if err := <-writerDone; err != nil {
		t.Fatal(err)
	}
}

func TestResearchRecallPriorityGateV37(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PRIORITY_V37") != "1" {
		t.Skip("opt-in admission-priority writer scheduling screen")
	}
	type arm struct {
		offer, frontierAge, feedbackAge []int64
		writerOffer, writerCall         []int64
		writerTotal                     int64
		writerStartedOverlap            int
	}
	all := map[bool]*arm{false: {}, true: {}}
	for trial := range 2 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, prioritized := range order {
			var gate *researchPriorityGateV37
			consume := consumeResearchLiveDecoupledV29
			if prioritized {
				gate = newResearchPriorityGateV37()
				consume = func(ctx context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
					return consumeResearchLiveDecoupledWithGateV37(ctx, s, tap, durable, gate)
				}
			}
			r := runResearchRecallLiveWithPolicyV37(t, true, consume, true, 4*time.Millisecond, true, gate, true)
			checkResearchLivePhasesV27(t, r)
			if len(r.writerOffer) != 256 || len(r.writerCall) != 256 || r.writerTotal <= 0 {
				t.Fatalf("missing writer timing: offered=%d call=%d total=%s", len(r.writerOffer), len(r.writerCall), time.Duration(r.writerTotal))
			}
			t.Logf("trial=%d prioritized=%v offers=%d labels=%d drops=%d writes=%d offered_overlap=%d started_overlap=%d gap_p50=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s writer_offer_p99=%s writer_call_p99=%s writer_total=%s", trial, prioritized, len(r.offer), r.labels, r.dropped, r.writes, r.overlap, r.writerStartedOverlap,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99), researchDurableLoadPercentile(r.feedbackAge, .99),
				researchDurableLoadPercentile(r.writerOffer, .99), researchDurableLoadPercentile(r.writerCall, .99), time.Duration(r.writerTotal))
			if prioritized && researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
				t.Errorf("frozen trial serving gate failed: trial=%d", trial)
			}
			a := all[prioritized]
			a.offer = append(a.offer, r.offer...)
			a.frontierAge = append(a.frontierAge, r.frontierAge...)
			a.feedbackAge = append(a.feedbackAge, r.feedbackAge...)
			a.writerOffer = append(a.writerOffer, r.writerOffer...)
			a.writerCall = append(a.writerCall, r.writerCall...)
			a.writerTotal += r.writerTotal
			a.writerStartedOverlap += r.writerStartedOverlap
		}
	}
	control, candidate := all[false], all[true]
	controlOffer := researchDurableLoadPercentile(control.offer, .99)
	candidateOffer := researchDurableLoadPercentile(candidate.offer, .99)
	frontierAge := researchDurableLoadPercentile(candidate.frontierAge, .99)
	feedbackAge := researchDurableLoadPercentile(candidate.feedbackAge, .99)
	controlWriter := researchDurableLoadPercentile(control.writerOffer, .99)
	candidateWriter := researchDurableLoadPercentile(candidate.writerOffer, .99)
	t.Logf("pooled control_offer_p99=%s candidate_offer_p99=%s offer_ratio=%.3f candidate_frontier_age_p99=%s candidate_feedback_age_p99=%s control_writer_offer_p99=%s candidate_writer_offer_p99=%s writer_p99_ratio=%.3f writer_total_ratio=%.3f started_overlap_control=%d candidate=%d", controlOffer, candidateOffer, float64(candidateOffer)/float64(controlOffer), frontierAge, feedbackAge,
		controlWriter, candidateWriter, float64(candidateWriter)/float64(controlWriter), float64(candidate.writerTotal)/float64(control.writerTotal), control.writerStartedOverlap, candidate.writerStartedOverlap)
	if len(control.offer) != 384 || len(candidate.offer) != 384 || len(candidate.frontierAge) != 128 || len(candidate.writerOffer) != 512 || candidateOffer >= 100*time.Millisecond || float64(candidateOffer) > 1.10*float64(controlOffer) || frontierAge >= 250*time.Millisecond || feedbackAge >= 250*time.Millisecond || float64(candidateWriter) > 1.25*float64(controlWriter) || float64(candidate.writerTotal) > 1.25*float64(control.writerTotal) || candidate.writerStartedOverlap*5 < control.writerStartedOverlap*4 {
		t.Error("frozen v37 priority-gate component failed")
	}
}
