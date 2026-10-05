//go:build researchpriority

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type sample struct {
	NS                                  int64
	Error                               string
	Nominated, Packed, ExplanationBytes int
	JournalMatched                      bool
}
type arm struct {
	TraceFile, BeforeHeap, AfterHeap string
	N, Workers, Repeat               int
	Enabled                          bool
	Samples                          []sample
	WallNS                           int64
	AllocatedBytes                   uint64
}

func heapSnapshot(path string) {
	runtime.GC()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	if err = pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
}
func run(n, workers, repeat int, enabled bool, facts []string) arm {
	em, err := embed.NewHashEmbedder(32)
	if err != nil {
		panic(err)
	}
	memory, err := libravdbstore.Open(libravdbstore.Config{Path: fmt.Sprintf("%s.stores/enabled%t.libravdb", os.Args[1], enabled), Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
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
		start := time.Now()
		packet, err := s.Recall(c, request)
		ns := time.Since(start).Nanoseconds()
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
		return result
	}
	for i := -2; i < 0; i++ {
		if r := call(i); r.Error != "" {
			panic(r.Error)
		}
	}
	profilePath := fmt.Sprintf("%s.%t.trace", os.Args[1], enabled)
	beforeHeap, afterHeap := profilePath+".before.heap", profilePath+".after.heap"
	heapSnapshot(beforeHeap)
	profile, err := os.OpenFile(profilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	if err := trace.Start(profile); err != nil {
		panic(err)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	out := arm{TraceFile: profilePath, BeforeHeap: beforeHeap, AfterHeap: afterHeap, N: n, Workers: workers, Repeat: repeat, Enabled: enabled, Samples: make([]sample, 128)}
	start := time.Now()
	gate := make(chan struct{})
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			<-gate
			for i := w; i < 128; i += workers {
				out.Samples[i] = call(i)
			}
		}(w)
	}
	close(gate)
	wg.Wait()
	out.WallNS = time.Since(start).Nanoseconds()
	trace.Stop()
	if err := profile.Close(); err != nil {
		panic(err)
	}
	runtime.ReadMemStats(&after)
	out.AllocatedBytes = after.TotalAlloc - before.TotalAlloc
	heapSnapshot(afterHeap)
	return out
}
func main() {
	if len(os.Args) != 2 {
		panic("NEW output JSON required")
	}
	if _, e := os.Stat(os.Args[1]); !os.IsNotExist(e) {
		panic("output exists or inaccessible")
	}
	if err := os.Mkdir(os.Args[1]+".stores", 0700); err != nil {
		panic(err)
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
	for repeat := 0; repeat < 1; repeat++ {
		for _, n := range []int{200} {
			for _, workers := range []int{1} {
				order := []bool{false, true}
				if repeat == 1 {
					order = []bool{true, false}
				}
				for _, enabled := range order {
					a := run(n, workers, repeat, enabled, facts)
					rows = append(rows, a)
					fmt.Println("completed", repeat, n, workers, enabled)
				}
			}
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"internal/store/libravdbstore/store.go", "internal/packing/research_incremental_diversity.go", "research/public-task-pilot/incremental-overlay-v1/source-3.go.txt", "cmd/research-task-durable-trace/main.go", "research/public-task-pilot/DURABLE_TRACE_PROTOCOL.md", "research/public-task-pilot/landing-transfer-v1/corpus.json", "research/public-task-pilot/incremental-overlay-v1/source-0.go.txt", "research/public-task-pilot/incremental-overlay-v1/source-1.go.txt", "research/public-task-pilot/incremental-overlay-v1/source-2.go.txt", "internal/researchcalendar/task_plan.go", "internal/researchcalendar/lexical.go", "internal/researchcalendar/packing_lexical.go", "internal/researchcalendar/packing_task.go", "internal/service/calendar_task_lexical_overlay.go", "internal/packing/packing.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	for _, a := range rows {
		for _, path := range []string{a.TraceFile, a.BeforeHeap, a.AfterHeap} {
			b, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}
			h := sha256.Sum256(b)
			hashes[path] = hex.EncodeToString(h[:])
		}
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
