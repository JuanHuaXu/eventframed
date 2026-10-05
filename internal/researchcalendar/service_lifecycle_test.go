package researchcalendar_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

var asOf = time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)

func runtimeFixture(t *testing.T, memory *memorystore.Store, ranker retrieval.CandidateRanker) *service.Service {
	t.Helper()
	em, err := embed.NewHashEmbedder(32)
	if err != nil {
		t.Fatal(err)
	}
	s, err := service.New(memory, em, service.Config{DefaultRecallK: 50, DefaultPackK: 1, DefaultTokenBudget: 10000, CandidateRanker: ranker, CandidateRankerRequired: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	for _, tenant := range []string{"alpha", "beta"} {
		for i, text := range []string{"Rosetta launched on 2004-03-02.", "Rosetta arrived at its comet on 2014-08-06."} {
			id := fmt.Sprintf("%s-%d", tenant, i)
			_, err := s.CaptureTurn(context.Background(), model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: model.TurnCapture{ID: id, TenantID: tenant, SessionID: "seed", Sequence: uint64(i + 1), UserText: text, AssistantText: "Recorded.", OccurredAt: asOf, ObservedAt: asOf, AvailableAt: asOf}})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return s
}

func request(ctx context.Context, tenant, session string, early bool) (context.Context, model.RecallRequest) {
	q := "Which retained Rosetta event occurred on the other date, rather than 2 March 2004?"
	if early {
		q = "Which retained Rosetta event occurred on the other date, rather than 6 August 2014?"
	}
	return researchcalendar.Bind(ctx, tenant, q), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: tenant, SessionID: session, Query: researchcalendar.Focus(q), AsOf: asOf.Add(time.Minute), RecallK: 50, PackK: 1, TokenBudget: 10000}
}

func TestServiceConcurrentOriginalQueries(t *testing.T) {
	s := runtimeFixture(t, memorystore.New(), researchcalendar.Ranker{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	failures := make(chan error, 64)
	start := make(chan struct{})
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			tenant := "alpha"
			if i%2 == 1 {
				tenant = "beta"
			}
			early := i%4 < 2
			c, r := request(ctx, tenant, fmt.Sprintf("request-%d", i), early)
			p, err := s.Recall(c, r)
			if err != nil {
				failures <- err
				return
			}
			want := tenant + "-1"
			if early {
				want = tenant + "-0"
			}
			if len(p.Candidates) != 1 || p.Candidates[0].Event.ID != want || p.Candidates[0].Event.TenantID != tenant {
				failures <- fmt.Errorf("request%d wrong binding/winner: %+v", i, p.Candidates)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

// Mutation/barrier wrappers are test instruments, not behavior of the ranker.
type barrierRanker struct {
	entered, release chan struct{}
	calls            atomic.Int32
}

func (*barrierRanker) ContractName() string { return "research/calendar-barrier-test" }
func (b *barrierRanker) RankCandidates(ctx context.Context, r retrieval.RankRequest) ([]retrieval.Candidate, error) {
	out, err := (researchcalendar.Ranker{}).RankCandidates(ctx, r)
	if err != nil {
		return nil, err
	}
	if b.calls.Add(1) == 1 {
		close(b.entered)
		select {
		case <-b.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return out, nil
}

type outcome struct {
	packet model.ContextPacket
	err    error
}

func pending(t *testing.T, s *service.Service, b *barrierRanker, ctx context.Context) <-chan outcome {
	t.Helper()
	done := make(chan outcome, 1)
	c, r := request(ctx, "alpha", "barrier", true)
	go func() { p, e := s.Recall(c, r); done <- outcome{p, e} }()
	select {
	case <-b.entered:
	case <-ctx.Done():
		t.Fatal("ranker barrier not reached", ctx.Err())
	}
	return done
}

func TestServiceRetriesStaleCalendarRanking(t *testing.T) {
	m := memorystore.New()
	b := &barrierRanker{entered: make(chan struct{}), release: make(chan struct{})}
	s := runtimeFixture(t, m, b)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := pending(t, s, b, ctx)
	before := m.Snapshot(ctx)
	after, err := m.BindBayesianPolicy(ctx, "forced-version-change")
	if err != nil || after == before {
		t.Fatal(after, err)
	}
	close(b.release)
	select {
	case got := <-done:
		if got.err != nil || b.calls.Load() != 2 || got.packet.Snapshot != after || len(got.packet.Candidates) != 1 || got.packet.Candidates[0].Event.ID != "alpha-0" {
			t.Fatal(got.err, b.calls.Load(), got.packet)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestServiceCancelledCalendarBarrier(t *testing.T) {
	b := &barrierRanker{entered: make(chan struct{}), release: make(chan struct{})}
	s := runtimeFixture(t, memorystore.New(), b)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := pending(t, s, b, ctx)
	cancel()
	select {
	case got := <-done:
		if !errors.Is(got.err, context.Canceled) || len(got.packet.Candidates) != 0 {
			t.Fatal(got)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled request did not return")
	}
}

type churnRanker struct {
	memory *memorystore.Store
	calls  atomic.Int32
}

func (*churnRanker) ContractName() string { return "research/calendar-churn-test" }
func (b *churnRanker) RankCandidates(ctx context.Context, r retrieval.RankRequest) ([]retrieval.Candidate, error) {
	out, err := (researchcalendar.Ranker{}).RankCandidates(ctx, r)
	if err != nil {
		return nil, err
	}
	n := b.calls.Add(1)
	_, err = b.memory.BindBayesianPolicy(ctx, fmt.Sprintf("churn-%d", n))
	return out, err
}

func TestServiceCalendarChurnExhaustsRetries(t *testing.T) {
	m := memorystore.New()
	b := &churnRanker{memory: m}
	s := runtimeFixture(t, m, b)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, r := request(ctx, "alpha", "churn", true)
	p, err := s.Recall(c, r)
	if !errors.Is(err, store.ErrStaleSnapshot) || b.calls.Load() != 5 || len(p.Candidates) != 0 {
		t.Fatal(err, b.calls.Load(), p)
	}
}
