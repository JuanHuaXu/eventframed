package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchShadowLoadPilot(t *testing.T) {
	for repeat := 0; repeat < 3; repeat++ {
		for order := 0; order < 2; order++ {
			enabled := (order+repeat)%2 == 1
			t.Run(fmt.Sprintf("repeat%d-enabled%v", repeat, enabled), func(t *testing.T) {
				p := ResearchShadowPolicy{Enabled: enabled, Capacity: 16, MaxAge: 100 * time.Millisecond, Process: func(ctx context.Context, in ResearchShadowInput) (float64, error) {
					x := 0.
					for pass := 0; pass < 2048; pass++ {
						if ctx.Err() != nil {
							return 0, ctx.Err()
						}
						for i := 0; i < in.Count; i++ {
							x += in.Scores[i] * in.Scores[i]
						}
					}
					return x, nil
				}}
				s, now := shadowService(t, p)
				start := make(chan struct{})
				durations := make(chan time.Duration, 512)
				errs := make(chan error, 544)
				var wg sync.WaitGroup
				for reader := 0; reader < 4; reader++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-start
						for i := 0; i < 128; i++ {
							begin := time.Now()
							_, e := s.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000})
							durations <- time.Since(begin)
							if e != nil {
								errs <- e
							}
						}
					}()
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					for i := 0; i < 32; i++ {
						ev := testutil.Event(fmt.Sprintf("load-%d", i), "public load fact", now.Add(-time.Second))
						ev.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
						ev.EmbeddingModel = s.embedder.ModelKey()
						_, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
						if e != nil {
							errs <- e
						}
					}
				}()
				close(start)
				wg.Wait()
				close(durations)
				close(errs)
				var values []time.Duration
				for d := range durations {
					values = append(values, d)
				}
				sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
				failures := 0
				for range errs {
					failures++
				}
				if enabled {
					s.researchShadow.Close()
				}
				t.Logf("n=%d p50_us=%.3f p95_us=%.3f p99_us=%.3f errors=%d shadow=%+v", len(values), float64(values[255])/1000, float64(values[486])/1000, float64(values[506])/1000, failures, s.ResearchShadowStatus())
				if failures > 0 {
					t.Errorf("concurrent service returned %d errors", failures)
				}
			})
		}
	}
}
