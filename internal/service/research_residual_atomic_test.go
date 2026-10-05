package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestResearchResidualAtomicParity(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, 9, 50, 200} {
		c, d, idx := strideFixture(n)
		s := &Service{store: &strideStore{Store: memorystore.New(), get: strideValues}}
		request := model.RecallRequest{TenantID: "research"}
		want, err := s.loadResidualCandidates(t.Context(), request, "query", c, d, idx)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s.researchLoadResidualAtomic(t.Context(), request, "query", c, d, idx)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("n%d err%v", n, err)
		}
	}
}

func TestResearchResidualAtomicCancellationJoins(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var active atomic.Int64
		s := &Service{store: &strideStore{Store: memorystore.New(), get: func(ctx context.Context, _, _, _ string) (model.ResidualCandidates, error) {
			active.Add(1)
			defer active.Add(-1)
			<-ctx.Done()
			return model.ResidualCandidates{}, ctx.Err()
		}}}
		c, d, idx := strideFixture(200)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := s.researchLoadResidualAtomic(ctx, model.RecallRequest{}, "q", c, d, idx)
			done <- err
		}()
		synctest.Wait()
		if active.Load() != 8 {
			t.Fatal("expected eight active workers", active.Load())
		}
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if active.Load() != 0 {
			t.Fatal("returned with live workers")
		}
	})
}

func TestResearchResidualAtomicErrorsAndDisabled(t *testing.T) {
	sentinel := errors.New("lookup failed")
	var calls atomic.Int64
	s := &Service{store: &strideStore{Store: memorystore.New(), get: func(context.Context, string, string, string) (model.ResidualCandidates, error) {
		calls.Add(1)
		return model.ResidualCandidates{}, sentinel
	}}}
	c, d, idx := strideFixture(50)
	got, err := s.researchLoadResidualAtomic(t.Context(), model.RecallRequest{}, "q", c, d, idx)
	if got != nil || err != sentinel {
		t.Fatal("lost lookup error")
	}
	calls.Store(0)
	delete(idx, c[0].Event.ID)
	if _, err := s.researchLoadResidualAtomic(t.Context(), model.RecallRequest{}, "q", c, d, idx); err == nil || calls.Load() != 0 {
		t.Fatal("missing frontier validation")
	}
	s.config.ResidualMode = ResidualModeDisabled
	got, err = s.researchLoadResidualAtomic(t.Context(), model.RecallRequest{}, "q", c, d, idx)
	if err != nil || len(got) != 50 || calls.Load() != 0 {
		t.Fatal("disabled path changed")
	}
}

func BenchmarkResearchResidualAtomic(b *testing.B) {
	c, d, idx := strideFixture(200)
	s := &Service{store: &strideStore{Store: memorystore.New(), get: strideValues}}
	for _, fast := range []bool{false, true} {
		b.Run(fmt.Sprint(fast), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if fast {
					s.researchLoadResidualAtomic(b.Context(), model.RecallRequest{TenantID: "research"}, "q", c, d, idx)
				} else {
					s.loadResidualCandidates(b.Context(), model.RecallRequest{TenantID: "research"}, "q", c, d, idx)
				}
			}
		})
	}
}

func TestResearchResidualAtomicSkewCounterexample(t *testing.T) {
	// Dynamic allocation must avoid the previously demonstrated stride imbalance.
	synctest.Test(t, func(t *testing.T) {
		c, d, idx := strideFixture(200)
		slow := map[string]bool{}
		for i := 0; i < len(c); i += 8 {
			slow[residualActionKey("q", c[i].Event.ID, model.RetrievalUsefulnessHorizon)] = true
		}
		s := &Service{store: &strideStore{Store: memorystore.New(), get: func(ctx context.Context, t, a, g string) (model.ResidualCandidates, error) {
			if slow[a] {
				time.Sleep(time.Millisecond)
			}
			return strideValues(ctx, t, a, g)
		}}}
		start := time.Now()
		want, err := s.loadResidualCandidates(t.Context(), model.RecallRequest{}, "q", c, d, idx)
		if err != nil {
			t.Fatal(err)
		}
		original := time.Since(start)
		start = time.Now()
		got, err := s.researchLoadResidualAtomic(t.Context(), model.RecallRequest{}, "q", c, d, idx)
		if err != nil {
			t.Fatal(err)
		}
		strided := time.Since(start)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("results changed")
		}
		if strided > original {
			t.Fatalf("load balancing regressed: original%s atomic%s", original, strided)
		}
		t.Logf("virtual-time skew: dispatcher=%s atomic=%s", original, strided)
	})
}
