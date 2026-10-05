//go:build researchpriority

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

type writeSample struct {
	NS          int64
	Error       string
	ActiveReads int64
}
type monitoredStore struct {
	*memorystore.Store
	attempts, stale atomic.Int64
}

func (s *monitoredStore) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	s.attempts.Add(1)
	err := s.Store.PutBayesianJournal(ctx, e)
	if errors.Is(err, store.ErrStaleSnapshot) {
		s.stale.Add(1)
	}
	return err
}

type sample struct {
	FutureRecords                       int
	NS                                  int64
	Error                               string
	Nominated, Packed, ExplanationBytes int
	JournalMatched                      bool
}
type arm struct {
	Mode                             string
	Writes                           []writeSample
	JournalAttempts, StaleRejections int64
	ReadWallNS                       int64
	N, Workers, Repeat               int
	Enabled                          bool
	Samples                          []sample
	WallNS                           int64
	AllocatedBytes                   uint64
}

func run(n, workers, repeat int, enabled bool, mode string, facts []string) arm {
	memory := &monitoredStore{Store: memorystore.New()}
	var active atomic.Int64
	em, err := embed.NewHashEmbedder(32)
	if err != nil {
		panic(err)
	}
	policy := packing.DefaultPolicy()
	policy.AdaptiveEnabled = true
	policy.DiversityEnabled = true
	s, err := service.New(memory, em, service.Config{DefaultRecallK: n, DefaultPackK: 10, DefaultTokenBudget: 10000, PackingPolicy: policy, ResearchTemporalPriority: enabled})
	if err != nil {
		panic(err)
	}
	defer s.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("public-%03d", i)
		_, err := s.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "seed", Sequence: uint64(i + 1), UserText: facts[i%len(facts)], AssistantText: "Recorded.", OccurredAt: now, ObservedAt: now, AvailableAt: now}})
		if err != nil {
			panic(err)
		}
	}
	call := func(i int) sample {
		query := "Which record refutes the claim that Curiosity landed after 2020?"
		c := researchcalendar.Bind(ctx, "research", query)
		request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: fmt.Sprintf("request-%d", i), Query: query, AsOf: now.Add(time.Minute), RecallK: n, PackK: 10, TokenBudget: 10000}
		active.Add(1)
		start := time.Now()
		packet, err := s.Recall(c, request)
		ns := time.Since(start).Nanoseconds()
		active.Add(-1)
		result := sample{NS: ns, Nominated: packet.BayesianShadow.Nominated, Packed: packet.Packed, ExplanationBytes: len(packet.TemporalPriority)}
		if err != nil {
			result.Error = err.Error()
			return result
		}
		j, err := memory.GetBayesianJournal(ctx, "research", packet.BayesianShadow.JournalID)
		result.JournalMatched = err == nil && j.SessionID == request.SessionID && j.TemporalPriority == packet.TemporalPriority
		if result.Nominated != n {
			result.Error = fmt.Sprintf("frontier mismatch: got%d want%d", result.Nominated, n)
		}
		if !result.JournalMatched {
			result.Error = "journal mismatch"
		}
		ids := make([]string, 0, len(packet.BayesianShadow.Decisions))
		for _, d := range packet.BayesianShadow.Decisions {
			ids = append(ids, d.EventID)
		}
		events, loadErr := memory.GetEvents(ctx, "research", ids, now.Add(10*time.Minute))
		if loadErr != nil || len(events) != len(ids) {
			result.Error = "frontier availability validation failed"
		} else {
			for _, e := range events {
				if e.AvailableAt.After(request.AsOf) {
					result.FutureRecords++
				}
			}
		}
		if result.FutureRecords > 0 {
			result.Error = "future evidence entered frontier"
		}
		return result
	}
	for i := -2; i < 0; i++ {
		if r := call(i); r.Error != "" {
			panic(r.Error)
		}
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	attemptsBefore, staleBefore := memory.attempts.Load(), memory.stale.Load()
	out := arm{Mode: mode, Writes: make([]writeSample, 16), N: n, Workers: workers, Repeat: repeat, Enabled: enabled, Samples: make([]sample, 32)}
	start := time.Now()
	gate := make(chan struct{})
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		<-gate
		for i := 0; i < 16; i++ {
			select {
			case <-time.After(5 * time.Millisecond):
			case <-ctx.Done():
				out.Writes[i].Error = ctx.Err().Error()
				return
			}
			stamp := now
			if mode == "future" {
				stamp = now.Add(2 * time.Minute)
			}
			id := fmt.Sprintf("writer-%03d", i)
			count := active.Load()
			start := time.Now()
			_, err := s.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: "writer", Sequence: uint64(i + 1), UserText: facts[i%len(facts)], AssistantText: "Recorded.", OccurredAt: stamp, ObservedAt: stamp, AvailableAt: stamp}})
			out.Writes[i] = writeSample{NS: time.Since(start).Nanoseconds(), ActiveReads: count}
			if err != nil {
				out.Writes[i].Error = err.Error()
			}
		}
	}()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			<-gate
			for i := w; i < 32; i += workers {
				out.Samples[i] = call(i)
			}
		}(w)
	}
	close(gate)
	wg.Wait()
	out.ReadWallNS = time.Since(start).Nanoseconds()
	<-writerDone
	out.JournalAttempts = memory.attempts.Load() - attemptsBefore
	out.StaleRejections = memory.stale.Load() - staleBefore
	out.WallNS = time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&after)
	out.AllocatedBytes = after.TotalAlloc - before.TotalAlloc
	return out
}
func main() {
	if len(os.Args) != 2 {
		panic("NEW output JSON required")
	}
	if _, e := os.Stat(os.Args[1]); !os.IsNotExist(e) {
		panic("output exists or inaccessible")
	}
	runtime.GOMAXPROCS(4)
	var corpus []struct{ Text string }
	b, e := os.ReadFile("research/public-task-pilot/landing-transfer-v1/corpus.json")
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, &corpus); e != nil {
		panic(e)
	}
	facts := []string{}
	for _, f := range corpus {
		facts = append(facts, f.Text)
	}
	if len(facts) != 4 {
		panic("expected four public facts")
	}
	rows := []arm{}
	for repeat := 0; repeat < 2; repeat++ {
		for _, n := range []int{50, 200} {
			for _, workers := range []int{4} {
				order := []bool{false, true}
				if repeat == 1 {
					order = []bool{true, false}
				}
				for _, mode := range []string{"future", "in-window"} {
					for _, enabled := range order {
						a := run(n, workers, repeat, enabled, mode, facts)
						rows = append(rows, a)
						fmt.Println("completed", repeat, n, workers, enabled, mode)
					}
				}
			}
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"cmd/research-task-lexical-churn/main.go", "research/public-task-pilot/TASK_LEXICAL_CHURN_PROTOCOL.md", "research/public-task-pilot/landing-transfer-v1/corpus.json", "research/public-task-pilot/task-lexical-overlay-v1/source-0.go.txt", "research/public-task-pilot/task-lexical-overlay-v1/source-1.go.txt", "research/public-task-pilot/task-lexical-overlay-v1/source-2.go.txt", "internal/researchcalendar/task_plan.go", "internal/researchcalendar/lexical.go", "internal/researchcalendar/packing_lexical.go", "internal/researchcalendar/packing_task.go", "internal/service/calendar_task_lexical_overlay.go", "internal/packing/packing.go", "internal/store/store.go", "internal/store/memorystore/store.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	out := struct {
		Go, OS, Arch string
		Procs        int
		Hashes       map[string]string
		Arms         []arm
	}{runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), hashes, rows}
	f, e := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if e = enc.Encode(out); e != nil {
		panic(e)
	}
	if e = f.Close(); e != nil {
		panic(e)
	}
}
