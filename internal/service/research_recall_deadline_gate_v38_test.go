package service

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

type researchDeadlineKeyV38 struct{}

type researchDeadlineWaiterV38 struct {
	deadline time.Time
	sequence uint64
}

type researchDeadlineGateV38 struct {
	mu      sync.Mutex
	cond    *sync.Cond
	active  bool
	next    uint64
	waiting []*researchDeadlineWaiterV38
}

func newResearchDeadlineGateV38() *researchDeadlineGateV38 {
	g := &researchDeadlineGateV38{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

func (g *researchDeadlineGateV38) earliest() *researchDeadlineWaiterV38 {
	var best *researchDeadlineWaiterV38
	for _, item := range g.waiting {
		if best == nil || item.deadline.Before(best.deadline) || (item.deadline.Equal(best.deadline) && item.sequence < best.sequence) {
			best = item
		}
	}
	return best
}

func (g *researchDeadlineGateV38) remove(target *researchDeadlineWaiterV38) {
	for i, item := range g.waiting {
		if item == target {
			g.waiting = append(g.waiting[:i], g.waiting[i+1:]...)
			return
		}
	}
	panic("deadline waiter disappeared")
}

func (g *researchDeadlineGateV38) enter(ctx context.Context) error {
	deadline, ok := ctx.Value(researchDeadlineKeyV38{}).(time.Time)
	if !ok || deadline.IsZero() {
		return errors.New("missing research scheduling deadline")
	}
	g.mu.Lock()
	g.next++
	item := &researchDeadlineWaiterV38{deadline: deadline, sequence: g.next}
	g.waiting = append(g.waiting, item)
	stop := context.AfterFunc(ctx, func() {
		g.mu.Lock()
		g.cond.Broadcast()
		g.mu.Unlock()
	})
	defer stop()
	for {
		if err := ctx.Err(); err != nil {
			g.remove(item)
			g.cond.Broadcast()
			g.mu.Unlock()
			return err
		}
		if !g.active && g.earliest() == item {
			g.remove(item)
			g.active = true
			g.mu.Unlock()
			return nil
		}
		g.cond.Wait()
	}
}

func (g *researchDeadlineGateV38) leave() {
	g.mu.Lock()
	g.active = false
	g.cond.Broadcast()
	g.mu.Unlock()
}

func (g *researchDeadlineGateV38) withAdmission(ctx context.Context, work func() error) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.leave()
	return work()
}

func (g *researchDeadlineGateV38) withWriter(ctx context.Context, work func() error) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.leave()
	return work()
}

func TestResearchDeadlineGateV38OrdersByDeadline(t *testing.T) {
	for _, admissionFirst := range []bool{false, true} {
		g := newResearchDeadlineGateV38()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		blocker := context.WithValue(ctx, researchDeadlineKeyV38{}, time.Now().Add(time.Hour))
		if err := g.enter(blocker); err != nil {
			cancel()
			t.Fatal(err)
		}
		order := make(chan string, 2)
		done := make(chan error, 2)
		admitDeadline := time.Now().Add(200 * time.Millisecond)
		if admissionFirst {
			admitDeadline = time.Now().Add(-time.Millisecond)
		}
		admitCtx := context.WithValue(ctx, researchDeadlineKeyV38{}, admitDeadline)
		writerCtx := context.WithValue(ctx, researchDeadlineKeyV38{}, time.Now().Add(30*time.Millisecond))
		go func() { done <- g.withAdmission(admitCtx, func() error { order <- "admission"; return nil }) }()
		go func() { done <- g.withWriter(writerCtx, func() error { order <- "writer"; return nil }) }()
		limit := time.Now().Add(200 * time.Millisecond)
		for {
			g.mu.Lock()
			waiting := len(g.waiting)
			g.mu.Unlock()
			if waiting == 2 {
				break
			}
			if time.Now().After(limit) {
				g.leave()
				cancel()
				t.Fatal("deadline waiters did not queue")
			}
			time.Sleep(time.Millisecond)
		}
		g.leave()
		want := "writer"
		if admissionFirst {
			want = "admission"
		}
		select {
		case first := <-order:
			if first != want {
				cancel()
				t.Fatalf("first deadline winner=%s want=%s", first, want)
			}
		case <-ctx.Done():
			cancel()
			t.Fatal("deadline arbitration timed out")
		}
		for range 2 {
			if err := <-done; err != nil {
				cancel()
				t.Fatal(err)
			}
		}
		cancel()
	}
}

func TestResearchRecallDeadlineGateV38(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_DEADLINE_V38") != "1" {
		t.Skip("opt-in deadline-aware writer/admission screen")
	}
	type arm struct {
		offer, frontierAge, feedbackAge []int64
		writerOffer, writerCall         []int64
		writerTotal                     int64
		writerStartedOverlap            int
	}
	all := map[bool]*arm{false: {}, true: {}}
	for trial := range 2 {
		order := []bool{true, false}
		if trial == 1 {
			order = []bool{false, true}
		}
		for _, deadlineScheduled := range order {
			var gate researchSchedulingGateV38
			consume := consumeResearchLiveDecoupledV29
			if deadlineScheduled {
				gate = newResearchDeadlineGateV38()
				consume = func(ctx context.Context, s *Service, tap *ResearchFrontierTap, durable *researchmemory.Durable) researchLiveConsumerV26 {
					return consumeResearchLiveDecoupledWithGateV37(ctx, s, tap, durable, gate)
				}
			}
			r := runResearchRecallLiveWithPolicyV37(t, true, consume, true, 4*time.Millisecond, true, gate, true)
			checkResearchLivePhasesV27(t, r)
			if len(r.writerOffer) != 256 || len(r.writerCall) != 256 || r.writerTotal <= 0 {
				t.Fatalf("missing writer timing: offered=%d call=%d total=%s", len(r.writerOffer), len(r.writerCall), time.Duration(r.writerTotal))
			}
			t.Logf("trial=%d deadline=%v offers=%d labels=%d drops=%d writes=%d offered_overlap=%d started_overlap=%d gap_p50=%s offer_p99=%s frontier_age_p99=%s feedback_age_p99=%s writer_offer_p99=%s writer_call_p99=%s writer_total=%s", trial, deadlineScheduled, len(r.offer), r.labels, r.dropped, r.writes, r.overlap, r.writerStartedOverlap,
				researchDurableLoadPercentile(r.gaps, .5), researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.frontierAge, .99), researchDurableLoadPercentile(r.feedbackAge, .99),
				researchDurableLoadPercentile(r.writerOffer, .99), researchDurableLoadPercentile(r.writerCall, .99), time.Duration(r.writerTotal))
			if deadlineScheduled && researchDurableLoadPercentile(r.offer, .99) >= 100*time.Millisecond {
				t.Errorf("frozen trial serving gate failed: trial=%d", trial)
			}
			a := all[deadlineScheduled]
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
		t.Error("frozen v38 deadline-gate component failed")
	}
}
