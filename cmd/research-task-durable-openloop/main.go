//go:build researchpriority

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/researchadmission"
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
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type writeSample struct {
	EventID     string
	DispatchNS  int64
	Entered     bool
	WaitNS      int64
	NS          int64
	Error       string
	ActiveReads int64
}
type monitoredStore struct {
	*libravdbstore.Store
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
	JournalID, JournalDigest            string
	DispatchNS                          int64
	Entered                             bool
	WaitNS                              int64
	FutureRecords                       int
	NS                                  int64
	Error                               string
	Nominated, Packed, ExplanationBytes int
	JournalMatched                      bool
}
type arm struct {
	DatabasePath                     string
	Warmups                          []sample
	ReopenErrors                     []string
	ReopenedJournals, ReopenedWrites int
	CloseNS, ReopenNS                int64
	ReadIntervalMS                   int
	Lease                            bool
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

func run(n, workers, repeat int, enabled bool, mode string, lease bool, intervalMS int, facts []string) arm {
	var active atomic.Int64
	permits, err := researchadmission.New(int64(workers))
	if err != nil {
		panic(err)
	}
	em, err := embed.NewHashEmbedder(32)
	if err != nil {
		panic(err)
	}
	dbPath := fmt.Sprintf("%s.stores/r%d-%s-gap%d-lease%t-enabled%t.libravdb", os.Args[1], repeat, mode, intervalMS, lease, enabled)
	dbConfig := libravdbstore.Config{Path: dbPath, Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true}
	db, err := libravdbstore.Open(dbConfig)
	if err != nil {
		panic(err)
	}
	memory := &monitoredStore{Store: db}
	policy := packing.DefaultPolicy()
	policy.AdaptiveEnabled = true
	policy.DiversityEnabled = true
	s, err := service.New(memory, em, service.Config{DefaultRecallK: n, DefaultPackK: 10, DefaultTokenBudget: 10000, PackingPolicy: policy, ResearchTemporalPriority: enabled})
	if err != nil {
		panic(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = s.Close()
		}
	}()
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
	call := func(i int, due time.Time) sample {
		ctx, cancel := context.WithDeadline(ctx, due.Add(100*time.Millisecond))
		defer cancel()
		dispatch := time.Since(due).Nanoseconds()
		query := "Which record refutes the claim that Curiosity landed after 2020?"
		c := researchcalendar.Bind(ctx, "research", query)
		request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: fmt.Sprintf("request-%d", i), Query: query, AsOf: now.Add(time.Minute), RecallK: n, PackK: 10, TokenBudget: 10000}
		start := due
		wait := int64(0)
		entered := false
		var packet model.ContextPacket
		invoke := func(context.Context) error {
			entered = true
			wait = time.Since(start).Nanoseconds()
			active.Add(1)
			defer active.Add(-1)
			var callErr error
			packet, callErr = s.Recall(c, request)
			return callErr
		}
		var err error
		if lease {
			err = permits.Read(ctx, invoke)
		} else {
			err = invoke(ctx)
		}
		ns := time.Since(start).Nanoseconds()
		if !entered {
			wait = ns
		}
		result := sample{DispatchNS: dispatch, Entered: entered, WaitNS: wait, NS: ns, Nominated: packet.BayesianShadow.Nominated, Packed: packet.Packed, ExplanationBytes: len(packet.TemporalPriority)}
		if err != nil {
			result.Error = err.Error()
			return result
		}
		j, err := memory.GetBayesianJournal(ctx, "research", packet.BayesianShadow.JournalID)
		result.JournalMatched = err == nil && j.SessionID == request.SessionID && j.TemporalPriority == packet.TemporalPriority
		result.JournalID = packet.BayesianShadow.JournalID
		if err == nil {
			encoded, e := json.Marshal(j)
			if e != nil {
				panic(e)
			}
			h := sha256.Sum256(encoded)
			result.JournalDigest = hex.EncodeToString(h[:])
		}
		upper := n
		if mode == "in-window" {
			upper += 16
		}
		if result.Nominated < n || result.Nominated > upper {
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
	warmups := []sample{}
	for i := -2; i < 0; i++ {
		warmups = append(warmups, call(i, time.Now()))
	}

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	attemptsBefore, staleBefore := memory.attempts.Load(), memory.stale.Load()
	out := arm{DatabasePath: dbPath, Warmups: warmups, ReadIntervalMS: intervalMS, Lease: lease, Mode: mode, Writes: make([]writeSample, 16), N: n, Workers: workers, Repeat: repeat, Enabled: enabled, Samples: make([]sample, 32)}
	start := time.Now().Add(20 * time.Millisecond)
	var readers, writers sync.WaitGroup
	// Each arrival has its own timer; completions never schedule new arrivals.
	for i := 0; i < 16; i++ {
		writers.Add(1)
		go func(i int) {
			defer writers.Done()
			due := start.Add(time.Duration(2*i+1) * time.Duration(intervalMS) * time.Millisecond)
			time.Sleep(time.Until(due))
			dispatch := time.Since(due).Nanoseconds()
			c, cancel := context.WithDeadline(ctx, due.Add(100*time.Millisecond))
			defer cancel()
			stamp := now
			if mode == "future" {
				stamp = now.Add(2 * time.Minute)
			}
			id := fmt.Sprintf("writer-%03d", i)
			count := active.Load()
			wait := int64(0)
			entered := false
			eventID := ""
			invoke := func(c context.Context) error {
				entered = true
				wait = time.Since(due).Nanoseconds()
				capture, err := s.CaptureTurn(c, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: model.TurnCapture{ID: id, TenantID: "research", SessionID: id, Sequence: 1, UserText: facts[i%len(facts)], AssistantText: "Recorded.", OccurredAt: stamp, ObservedAt: stamp, AvailableAt: stamp}})
				eventID = capture.EventID
				return err
			}
			var err error
			if lease {
				err = permits.Write(c, invoke)
			} else {
				err = invoke(c)
			}
			ns := time.Since(due).Nanoseconds()
			if !entered {
				wait = ns
			}
			out.Writes[i] = writeSample{EventID: eventID, DispatchNS: dispatch, Entered: entered, WaitNS: wait, NS: ns, ActiveReads: count}
			if err != nil {
				out.Writes[i].Error = err.Error()
			}
		}(i)
	}
	for i := 0; i < 32; i++ {
		readers.Add(1)
		go func(i int) {
			defer readers.Done()
			due := start.Add(time.Duration(i) * time.Duration(intervalMS) * time.Millisecond)
			time.Sleep(time.Until(due))
			out.Samples[i] = call(i, due)
		}(i)
	}
	readers.Wait()
	out.ReadWallNS = time.Since(start).Nanoseconds()
	writers.Wait()
	out.JournalAttempts = memory.attempts.Load() - attemptsBefore
	out.StaleRejections = memory.stale.Load() - staleBefore
	out.WallNS = time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&after)
	out.AllocatedBytes = after.TotalAlloc - before.TotalAlloc
	closeStart := time.Now()
	closeErr := s.Close()
	closed = true
	out.CloseNS = time.Since(closeStart).Nanoseconds()
	if closeErr != nil {
		out.ReopenErrors = append(out.ReopenErrors, "close: "+closeErr.Error())
	}
	reopenStart := time.Now()
	reopened, reopenErr := libravdbstore.Open(dbConfig)
	if reopenErr != nil {
		out.ReopenErrors = append(out.ReopenErrors, "reopen: "+reopenErr.Error())
		return out
	}
	for _, r := range out.Samples {
		if r.Error != "" {
			continue
		}
		j, err := reopened.GetBayesianJournal(ctx, "research", r.JournalID)
		encoded, encodeErr := json.Marshal(j)
		h := sha256.Sum256(encoded)
		if err != nil || encodeErr != nil || hex.EncodeToString(h[:]) != r.JournalDigest {
			out.ReopenErrors = append(out.ReopenErrors, "journal mismatch: "+r.JournalID)
		} else {
			out.ReopenedJournals++
		}
	}
	for _, w := range out.Writes {
		if w.Error != "" {
			continue
		}
		events, err := reopened.GetEvents(ctx, "research", []string{w.EventID}, now.Add(10*time.Minute))
		if err != nil || len(events) != 1 || events[0].ID != w.EventID {
			out.ReopenErrors = append(out.ReopenErrors, "write missing: "+w.EventID)
		} else {
			out.ReopenedWrites++
		}
	}
	if err := reopened.Close(); err != nil {
		out.ReopenErrors = append(out.ReopenErrors, "reclosed: "+err.Error())
	}
	out.ReopenNS = time.Since(reopenStart).Nanoseconds()
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
	for repeat := 0; repeat < 2; repeat++ {
		for _, n := range []int{200} {
			for _, workers := range []int{4} {
				order := []bool{false, true}
				if repeat == 1 {
					order = []bool{true, false}
				}
				for _, mode := range []string{"future", "in-window"} {
					for _, intervalMS := range []int{20, 5} {
						for _, lease := range order {
							for _, enabled := range order {
								a := run(n, workers, repeat, enabled, mode, lease, intervalMS, facts)
								rows = append(rows, a)
								fmt.Println("completed", repeat, n, workers, enabled, mode, lease, intervalMS)
							}
						}
					}
				}
			}
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"internal/packing/research_incremental_diversity.go", "research/public-task-pilot/incremental-overlay-v1/source-3.go.txt", "research/public-task-pilot/build-incremental-overlay.mjs", "internal/researchadmission/gate.go", "cmd/research-task-durable-openloop/main.go", "research/public-task-pilot/DURABLE_OPENLOOP_PROTOCOL.md", "research/public-task-pilot/landing-transfer-v1/corpus.json", "research/public-task-pilot/incremental-overlay-v1/source-0.go.txt", "research/public-task-pilot/incremental-overlay-v1/source-1.go.txt", "research/public-task-pilot/incremental-overlay-v1/source-2.go.txt", "internal/researchcalendar/task_plan.go", "internal/researchcalendar/lexical.go", "internal/researchcalendar/packing_lexical.go", "internal/researchcalendar/packing_task.go", "internal/service/calendar_task_lexical_overlay.go", "internal/packing/packing.go", "internal/store/store.go", "internal/store/libravdbstore/store.go"} {
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
