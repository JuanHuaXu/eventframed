package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type persistentShadowTrial struct {
	Trial             int
	Enabled           bool
	FutureIngest      bool
	TemporalReuse     bool
	ReadNS, WriteNS   []int64
	Errors            []string
	OverlappingWrites int64
	Before, After     ResearchShadowStatus
	P99               int64
}

func TestResearchShadowPersistentPilot(t *testing.T) {
	if os.Getenv("EVENTFRAME_RESEARCH_ARTIFACT") == "" {
		t.Skip("opt-in timing experiment; set a new EVENTFRAME_RESEARCH_ARTIFACT path to preserve failures")
	}
	var results []persistentShadowTrial
	future := os.Getenv("EVENTFRAME_RESEARCH_FUTURE_INGEST") == "1"
	temporal := os.Getenv("EVENTFRAME_RESEARCH_TEMPORAL_REUSE") == "1"
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 2; order++ {
			enabled := (trial+order)%2 == 1
			r := persistentShadowTrial{Trial: trial, Enabled: enabled, FutureIngest: future, TemporalReuse: temporal}
			em, e := embed.NewHashEmbedder(32)
			if e != nil {
				t.Fatal(e)
			}
			db, e := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/shadow.libravdb", Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
			if e != nil {
				t.Fatal(e)
			}
			p := ResearchShadowPolicy{Enabled: enabled, TemporalReuse: temporal, Capacity: 16, MaxAge: 250 * time.Millisecond, Process: func(ctx context.Context, in ResearchShadowInput) (float64, error) {
				x := 0.
				for pass := 0; pass < 2048; pass++ {
					if e := ctx.Err(); e != nil {
						return 0, e
					}
					for i := 0; i < in.Count; i++ {
						x += in.Scores[i] * in.Scores[i]
					}
				}
				return x, nil
			}}
			s, e := New(db, em, Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, ResearchShadow: p})
			if e != nil {
				db.Close()
				t.Fatal(e)
			}
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			observe := func(i int) error {
				at := now.Add(-time.Second)
				if future && i >= 50 {
					at = now.Add(time.Second)
				}
				ev := testutil.Event(fmt.Sprintf("persistent-%d", i), "public load placeholder", at)
				_, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
				return e
			}
			for i := 0; i < 50; i++ {
				if e := observe(i); e != nil {
					s.Close()
					t.Fatal(e)
				}
			}
			var mu sync.Mutex
			var remaining atomic.Int64
			remaining.Store(4)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for worker := 0; worker < 4; worker++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer remaining.Add(-1)
					<-start
					for i := 0; i < 64; i++ {
						begin := time.Now()
						_, e := s.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "research-read", Query: "public load placeholder", AsOf: now, RecallK: 50, PackK: 10, TokenBudget: 10000})
						elapsed := time.Since(begin).Nanoseconds()
						mu.Lock()
						r.ReadNS = append(r.ReadNS, elapsed)
						if e != nil {
							r.Errors = append(r.Errors, "read: "+e.Error())
						}
						mu.Unlock()
					}
				}()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				for i := 0; i < 64; i++ {
					begin := time.Now()
					e := observe(50 + i)
					elapsed := time.Since(begin).Nanoseconds()
					mu.Lock()
					r.WriteNS = append(r.WriteNS, elapsed)
					if e != nil {
						r.Errors = append(r.Errors, "write: "+e.Error())
					}
					if remaining.Load() > 0 {
						r.OverlappingWrites++
					}
					mu.Unlock()
					time.Sleep(2 * time.Millisecond)
				}
			}()
			close(start)
			wg.Wait()
			r.Before = s.ResearchShadowStatus()
			s.Close()
			r.After = s.ResearchShadowStatus()
			ordered := append([]int64(nil), r.ReadNS...)
			sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
			r.P99 = ordered[(len(ordered)-1)*99/100]
			results = append(results, r)
			t.Logf("trial%d enabled%v p99_ms=%.3f errors%d overlap%d completed%d stale%d", trial, enabled, float64(r.P99)/1e6, len(r.Errors), r.OverlappingWrites, r.After.Completed, r.After.Stale)
		}
	}
	if path := os.Getenv("EVENTFRAME_RESEARCH_ARTIFACT"); path != "" {
		f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		e = json.NewEncoder(f).Encode(results)
		closeErr := f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	for trial := 0; trial < 3; trial++ {
		var off, on persistentShadowTrial
		for _, r := range results {
			if r.Trial == trial {
				if r.Enabled {
					on = r
				} else {
					off = r
				}
			}
		}
		if len(off.Errors)+len(on.Errors) > 0 || off.OverlappingWrites == 0 || on.OverlappingWrites == 0 || float64(on.P99) > 1.1*float64(off.P99) || on.After.Completed == 0 || on.After.HasResult || on.After.Accepted != on.After.Completed+on.After.Stale+on.After.Failed+on.After.Cancelled {
			t.Errorf("trial%d failed frozen screen: off=%d on=%d errors=%d completed=%d", trial, off.P99, on.P99, len(off.Errors)+len(on.Errors), on.After.Completed)
		}
	}
}
