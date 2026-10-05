package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type concurrentPinKeyV1 struct{}

type concurrentPinStateV1 struct {
	mu           sync.Mutex
	pinned       model.Snapshot
	pending      bool
	injected     bool
	writeVersion uint64
	searches     int
	inject       func() (uint64, error)
}

type concurrentPinStoreV1 struct {
	store.EventStore
	pin bool
}

func (s *concurrentPinStoreV1) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	state, ok := ctx.Value(concurrentPinKeyV1{}).(*concurrentPinStateV1)
	if !ok {
		return nil, errors.New("research Recall context lacks pin state")
	}
	if s.pin {
		pinned := s.EventStore.Snapshot(ctx)
		state.mu.Lock()
		state.pinned, state.pending = pinned, true
		state.mu.Unlock()
	}
	results, err := s.EventStore.Search(ctx, tenantID, vector, availableBy, limit)
	state.mu.Lock()
	state.searches++
	first := !state.injected
	if first {
		state.injected = true
	}
	state.mu.Unlock()
	if err != nil || !first {
		return results, err
	}
	version, err := state.inject()
	state.mu.Lock()
	state.writeVersion = version
	state.mu.Unlock()
	return results, err
}

func (s *concurrentPinStoreV1) Snapshot(ctx context.Context) model.Snapshot {
	state, ok := ctx.Value(concurrentPinKeyV1{}).(*concurrentPinStateV1)
	if !ok || !s.pin {
		return s.EventStore.Snapshot(ctx)
	}
	state.mu.Lock()
	if state.pending {
		state.pending = false
		pinned := state.pinned
		state.mu.Unlock()
		return pinned
	}
	state.mu.Unlock()
	return s.EventStore.Snapshot(ctx)
}

type concurrentPinResultV1 struct {
	callNS, offeredNS int64
	searches          int
	injected          bool
	writeVersion      uint64
	success, stale    bool
	omission, future  bool
	journalMismatch   bool
	otherErr          error
}

type concurrentPinStatsV1 struct {
	calls, success, stale, other, searches int
	omissions, future, journalMismatch     int
	missingWrites                          int
	callNS, offeredNS, offerGapsNS         []int64
}

func runConcurrentPinArmV1(t *testing.T, block int, pin bool) concurrentPinStatsV1 {
	t.Helper()
	ctx := context.Background()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	db, err := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/concurrent-pin.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	wrapper := &concurrentPinStoreV1{EventStore: db, pin: pin}
	s, err := New(wrapper, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer s.Close()
	arm := "current"
	if pin {
		arm = "pin"
	}
	put := func(ctx context.Context, tenant, id string, at time.Time) (uint64, error) {
		event := testutil.Event(id, "public query", at)
		event.TenantID = tenant
		response, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event})
		return response.Snapshot.RuntimeVersion, err
	}
	const count = 64
	for i := 0; i < count; i++ {
		tenant := fmt.Sprintf("tenant-%d-%s-%02d", block, arm, i)
		if _, err := put(ctx, tenant, fmt.Sprintf("seed-%d-%02d", block, i), asOf.Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
		if _, err := put(ctx, tenant, fmt.Sprintf("future-%d-%02d", block, i), asOf.Add(time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	type job struct {
		index int
		at    time.Time
	}
	jobs := make(chan job, count)
	results := make(chan concurrentPinResultV1, count)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range jobs {
				i := work.index
				tenant := fmt.Sprintf("tenant-%d-%s-%02d", block, arm, i)
				seedID := fmt.Sprintf("seed-%d-%02d", block, i)
				newID := fmt.Sprintf("new-%d-%02d", block, i)
				state := &concurrentPinStateV1{}
				requestCtx := context.WithValue(ctx, concurrentPinKeyV1{}, state)
				state.inject = func() (uint64, error) { return put(requestCtx, tenant, newID, asOf.Add(-time.Minute)) }
				start := time.Now()
				packet, recallErr := s.Recall(requestCtx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: tenant, SessionID: tenant, Query: "public query", AsOf: asOf, RecallK: 50, PackK: 50, TokenBudget: 10000})
				completed := time.Now()
				state.mu.Lock()
				result := concurrentPinResultV1{callNS: completed.Sub(start).Nanoseconds(), offeredNS: completed.Sub(work.at).Nanoseconds(), searches: state.searches, injected: state.injected, writeVersion: state.writeVersion}
				state.mu.Unlock()
				if recallErr != nil {
					result.stale = errors.Is(recallErr, store.ErrStaleSnapshot)
					if !result.stale {
						result.otherErr = recallErr
					}
					results <- result
					continue
				}
				result.success = true
				journal, journalErr := db.GetBayesianJournal(requestCtx, tenant, packet.BayesianShadow.JournalID)
				result.journalMismatch = journalErr != nil || journal.Snapshot != packet.Snapshot
				seenSeed, seenNew := false, false
				for _, candidate := range packet.Candidates {
					seenSeed = seenSeed || candidate.Event.ID == seedID
					seenNew = seenNew || candidate.Event.ID == newID
					result.future = result.future || candidate.Event.AvailableAt.After(asOf) || candidate.Event.TenantID != tenant
				}
				result.omission = !seenSeed || (packet.Snapshot.RuntimeVersion >= result.writeVersion && !seenNew)
				results <- result
			}
		}()
	}
	startOffers := time.Now()
	var offered []time.Time
	for i := 0; i < count; i++ {
		if wait := time.Until(startOffers.Add(time.Duration(i) * 4 * time.Millisecond)); wait > 0 {
			time.Sleep(wait)
		}
		at := time.Now()
		offered = append(offered, at)
		jobs <- job{index: i, at: at}
	}
	close(jobs)
	wg.Wait()
	close(results)
	var out concurrentPinStatsV1
	for i := 1; i < len(offered); i++ {
		out.offerGapsNS = append(out.offerGapsNS, offered[i].Sub(offered[i-1]).Nanoseconds())
	}
	for result := range results {
		out.calls++
		out.searches += result.searches
		out.callNS = append(out.callNS, result.callNS)
		out.offeredNS = append(out.offeredNS, result.offeredNS)
		if !result.injected || result.writeVersion == 0 {
			out.missingWrites++
		}
		if result.success {
			out.success++
		}
		if result.stale {
			out.stale++
		}
		if result.otherErr != nil {
			out.other++
			t.Logf("block=%d arm=%s other_error=%v", block, arm, result.otherErr)
		}
		if result.omission {
			out.omissions++
		}
		if result.future {
			out.future++
		}
		if result.journalMismatch {
			out.journalMismatch++
		}
	}
	return out
}

func TestResearchRecallPresearchConcurrentV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PRESEARCH_CONCURRENT_V1") != "1" {
		t.Skip("opt-in concurrent pre-search snapshot screen")
	}
	stats := map[string]concurrentPinStatsV1{}
	for block := 0; block < 2; block++ {
		for offset := 0; offset < 2; offset++ {
			pin := (block+offset)%2 == 1
			arm := "current"
			if pin {
				arm = "pin"
			}
			run := runConcurrentPinArmV1(t, block, pin)
			pooled := stats[arm]
			pooled.calls += run.calls
			pooled.success += run.success
			pooled.stale += run.stale
			pooled.other += run.other
			pooled.searches += run.searches
			pooled.omissions += run.omissions
			pooled.future += run.future
			pooled.journalMismatch += run.journalMismatch
			pooled.missingWrites += run.missingWrites
			pooled.callNS = append(pooled.callNS, run.callNS...)
			pooled.offeredNS = append(pooled.offeredNS, run.offeredNS...)
			pooled.offerGapsNS = append(pooled.offerGapsNS, run.offerGapsNS...)
			stats[arm] = pooled
		}
	}
	for _, arm := range []string{"current", "pin"} {
		r := stats[arm]
		t.Logf("arm=%s calls=%d success=%d stale=%d other=%d searches=%d omissions=%d future=%d journal_mismatch=%d missing_writes=%d offer_gap_p50=%s call_p50=%s call_p95=%s call_p99=%s offer_to_done_p50=%s offer_to_done_p95=%s offer_to_done_p99=%s",
			arm, r.calls, r.success, r.stale, r.other, r.searches, r.omissions, r.future, r.journalMismatch, r.missingWrites,
			researchDurableLoadPercentile(r.offerGapsNS, .5),
			researchDurableLoadPercentile(r.callNS, .5), researchDurableLoadPercentile(r.callNS, .95), researchDurableLoadPercentile(r.callNS, .99),
			researchDurableLoadPercentile(r.offeredNS, .5), researchDurableLoadPercentile(r.offeredNS, .95), researchDurableLoadPercentile(r.offeredNS, .99))
		if r.calls != 128 || r.missingWrites != 0 || r.future != 0 || r.journalMismatch != 0 {
			t.Errorf("arm=%s lost offered writes or violated availability/journal integrity", arm)
		}
	}
	good := stats["pin"]
	if good.omissions != 0 || good.success < 127 || researchDurableLoadPercentile(good.offeredNS, .99) >= 100*time.Millisecond {
		t.Errorf("pin arm failed frozen correctness, completion or latency gate: success=%d omissions=%d offer_p99=%s",
			good.success, good.omissions, researchDurableLoadPercentile(good.offeredNS, .99))
	}
}
