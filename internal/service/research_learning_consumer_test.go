package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type learningConsumerResult struct {
	Capacity, Requests                   int
	AgeNS                                []int64
	AdmitNS, FeedbackNS, WaitNS, Polls   []int64
	Trial                                int
	Enabled                              bool
	ReadNS, WriteNS                      []int64
	Errors                               []string
	Overlap                              int
	Admitted, Dropped, Completed, Failed uint64
	P99                                  int64
}

func learningConsumerArm(t *testing.T, trial, capacity, requests int) learningConsumerResult {
	t.Helper()
	enabled := capacity > 0
	r := learningConsumerResult{Trial: trial, Enabled: enabled, Capacity: capacity, Requests: requests}
	em, e := embed.NewHashEmbedder(32)
	if e != nil {
		t.Fatal(e)
	}
	db, e := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/learning.libravdb", Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if e != nil {
		t.Fatal(e)
	}
	tap, e := NewResearchFrontierTap("tenant-a", max(1, capacity))
	if e != nil {
		t.Fatal(e)
	}
	defer tap.Close()
	config := Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000}
	if enabled {
		config.ResearchFrontier = tap
	}
	s, e := New(db, em, config)
	if e != nil {
		db.Close()
		t.Fatal(e)
	}
	defer s.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	put := func(i int) error {
		at := now.Add(-time.Second)
		if i >= 50 {
			at = now.Add(time.Hour)
		}
		ev := testutil.Event(fmt.Sprintf("load-%d", i), "public load fixture", at)
		_, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
		return e
	}
	for i := 0; i < 50; i++ {
		if e := put(i); e != nil {
			t.Fatal(e)
		}
	}
	if enabled {
		s.store, e = researchpublicationstore.Wrap(db)
		if e != nil {
			t.Fatal(e)
		}
	}
	var mu sync.Mutex
	consumerDone := make(chan struct{})
	var bridge *ResearchFeedbackBridge
	if enabled {
		bridge, e = NewResearchTemporalFeedbackBridge(s, "tenant-a", 42)
		if e != nil {
			t.Fatal(e)
		}
		defer bridge.Close()
		go func() {
			defer close(consumerDone)
			var expected uint64
			for {
				in, e := tap.Take(context.Background())
				if e != nil {
					return
				}
				stageStart := time.Now()
				_, e = bridge.Admit(context.Background(), in)
				r.AdmitNS = append(r.AdmitNS, time.Since(stageStart).Nanoseconds())
				if e != nil {
					mu.Lock()
					r.Errors = append(r.Errors, "admit: "+e.Error())
					mu.Unlock()
					continue
				}
				mu.Lock()
				r.Admitted++
				mu.Unlock()
				stageStart = time.Now()
				for _, c := range in.Candidates {
					var id int
					if _, e = fmt.Sscanf(c.EventID, "load-%d", &id); e != nil {
						mu.Lock()
						r.Errors = append(r.Errors, "fixture: "+e.Error())
						mu.Unlock()
						continue
					}
					if e = bridge.Feedback(context.Background(), in.JournalID, c.EventID, id%2 == 0, now); e != nil {
						mu.Lock()
						r.Errors = append(r.Errors, "feedback: "+e.Error())
						mu.Unlock()
					} else {
						expected++
					}
				}
				r.FeedbackNS = append(r.FeedbackNS, time.Since(stageStart).Nanoseconds())
				stageStart = time.Now()
				var polls int64
				deadline := time.Now().Add(5 * time.Second)
				for {
					n, failed, _, _ := bridge.worker.Counts()
					if n+failed >= expected {
						break
					}
					if time.Now().After(deadline) {
						mu.Lock()
						r.Errors = append(r.Errors, "worker timeout")
						mu.Unlock()
						return
					}
					polls++
					time.Sleep(time.Millisecond)
				}
				r.WaitNS = append(r.WaitNS, time.Since(stageStart).Nanoseconds())
				r.Polls = append(r.Polls, polls)
				r.AgeNS = append(r.AgeNS, time.Since(in.queuedAt).Nanoseconds())
			}
		}()
	} else {
		close(consumerDone)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var readers atomic.Int64
	readers.Store(4)
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			defer readers.Add(-1)
			<-start
			for i := 0; i < requests/4; i++ {
				begin := time.Now()
				_, e := s.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: fmt.Sprintf("read-%d-%d", w, i), Query: "public load fixture", AsOf: now, RecallK: 50, PackK: 10, TokenBudget: 10000})
				ns := time.Since(begin).Nanoseconds()
				mu.Lock()
				r.ReadNS = append(r.ReadNS, ns)
				if e != nil {
					r.Errors = append(r.Errors, "read: "+e.Error())
				}
				mu.Unlock()
			}
		}(w)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 50; i < 50+requests/2; i++ {
			begin := time.Now()
			e := put(i)
			ns := time.Since(begin).Nanoseconds()
			mu.Lock()
			r.WriteNS = append(r.WriteNS, ns)
			if e != nil {
				r.Errors = append(r.Errors, "write: "+e.Error())
			}
			if readers.Load() > 0 {
				r.Overlap++
			}
			mu.Unlock()
			time.Sleep(2 * time.Millisecond)
		}
	}()
	close(start)
	wg.Wait()
	tap.Close()
	<-consumerDone
	if enabled {
		r.Completed, r.Failed, _, _ = bridge.worker.Counts()
		r.Dropped = tap.Dropped()
	}
	ordered := append([]int64(nil), r.ReadNS...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	r.P99 = ordered[int(math.Ceil(.99*float64(len(ordered))))-1]
	return r
}

func TestResearchLearningConsumerDiagnostic(t *testing.T) {
	path := os.Getenv("EVENTFRAME_LEARNING_CONSUMER_ARTIFACT")
	if path == "" {
		t.Skip("opt-in bounded buffer study")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources := map[string]string{}
	hashes := map[string]string{}
	for _, name := range []string{"internal/service/research_learning_consumer_test.go", "internal/service/research_feedback.go", "internal/service/research_frontier.go", "internal/researchmemory/background.go", "internal/researchmemory/adapter.go", "internal/researchpublicationstore/store.go", "internal/researchpublicationstore/mutations.go", "internal/researchpublicationstore/capabilities.go", "internal/researchpublication/publication.go", "docs/experiments/mmm-consumer-v6-protocol.md"} {
		content, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(content)
		h := sha256.Sum256(content)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"kind": "header", "Sources": sources, "Hashes": hashes, "expected_arms": 18}); e != nil {
		t.Fatal(e)
	}
	for _, requests := range []int{64, 192} {
		for trial := 0; trial < 3; trial++ {
			results := map[int]learningConsumerResult{}
			for order := 0; order < 3; order++ {
				capacity := []int{0, 16, 64}[(trial+order)%3]
				r := learningConsumerArm(t, trial, capacity, requests)
				if e = enc.Encode(r); e != nil {
					t.Fatal(e)
				}
				if e = f.Sync(); e != nil {
					t.Fatal(e)
				}
				results[capacity] = r
				t.Logf("n%d trial%d cap%d p99_ms %.3f admitted%d drops%d errors%d", requests, trial, capacity, float64(r.P99)/1e6, r.Admitted, r.Dropped, len(r.Errors))
			}
			off := results[0]
			for _, capacity := range []int{16, 64} {
				on := results[capacity]
				age := append([]int64(nil), on.AgeNS...)
				sort.Slice(age, func(i, j int) bool { return age[i] < age[j] })
				p95 := int64(0)
				if len(age) > 0 {
					p95 = age[int(math.Ceil(.95*float64(len(age))))-1]
				}
				t.Logf("n%d trial%d cap%d completion_p95_ms %.3f", requests, trial, capacity, float64(p95)/1e6)
				if len(off.Errors)+len(on.Errors) > 0 || off.Overlap == 0 || on.Overlap == 0 || float64(on.P99) > 1.1*float64(off.P99) || float64(on.Admitted) < .8*float64(requests) || on.Completed < on.Admitted*50 || on.Failed != 0 || p95 > int64(250*time.Millisecond) {
					t.Errorf("n%d trial%d cap%d FAILED frozen screen", requests, trial, capacity)
				}
			}
		}
	}
}
