package observationlearners_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	learners "github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type pacedFitTrace struct {
	FromOfferMS                                  float64
	ID                                           int
	EntryRemainingMS, ComputeMS, ExitRemainingMS float64
	ContextError, FitError                       string
}
type pacedFitArm struct {
	Rate                 int
	ScheduledReadNS      []int64
	Traces               []pacedFitTrace
	Trial                int
	Arm                  string
	Distinct             bool
	Fits, Hits, Checked  uint64
	ReadNS, WriteNS      []int64
	Errors               []string
	Overlap              int
	Status               service.ResearchShadowStatus
	P99                  int64
	Started, Interrupted uint64
}

func runPacedFitArm(t *testing.T, trial int, arm string, rate int) pacedFitArm {
	t.Helper()
	r := pacedFitArm{Trial: trial, Arm: arm, Distinct: true, Rate: rate}
	em, e := embed.NewHashEmbedder(32)
	if e != nil {
		t.Fatal(e)
	}
	db, e := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/fit.libravdb", Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if e != nil {
		t.Fatal(e)
	}
	var started, interrupted atomic.Uint64
	var traceMu sync.Mutex
	var epoch time.Time
	spacing := time.Second / time.Duration(rate)
	process, counts, err := learners.ResearchTableChangingFixture(true, true)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	policy := service.ResearchShadowPolicy{Enabled: arm != "off", TemporalReuse: true, Capacity: 16, MaxAge: 100 * time.Millisecond, Process: func(ctx context.Context, in service.ResearchShadowInput) (float64, error) {
		started.Add(1)
		id := int(in.AsOf.Sub(now) / time.Microsecond)
		deadline, ok := ctx.Deadline()
		if !ok {
			return 0, fmt.Errorf("missing processor deadline")
		}
		begin := time.Now()
		v, e := process(ctx, id)
		end := time.Now()
		trace := pacedFitTrace{FromOfferMS: float64(end.Sub(epoch.Add(time.Duration(id)*spacing))) / 1e6, ID: id, EntryRemainingMS: float64(deadline.Sub(begin)) / 1e6, ComputeMS: float64(end.Sub(begin)) / 1e6, ExitRemainingMS: float64(deadline.Sub(end)) / 1e6}
		if e != nil {
			trace.FitError = e.Error()
		}
		if err := ctx.Err(); err != nil {
			trace.ContextError = err.Error()
		}
		traceMu.Lock()
		r.Traces = append(r.Traces, trace)
		traceMu.Unlock()
		if ctx.Err() != nil {
			interrupted.Add(1)
		}
		return v, e
	}}
	s, e := service.New(db, em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, ResearchShadow: policy})
	if e != nil {
		db.Close()
		t.Fatal(e)
	}
	defer s.Close()
	put := func(i int) error {
		at := now.Add(-time.Second)
		if i >= 50 {
			at = now.Add(time.Hour)
		}
		ev := testutil.Event(fmt.Sprintf("fit-%d", i), "public load fixture", at)
		_, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
		return e
	}
	for i := 0; i < 50; i++ {
		if e := put(i); e != nil {
			t.Fatal(e)
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var readers atomic.Int64
	readers.Store(4)
	start := make(chan struct{})
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			defer readers.Add(-1)
			<-start
			for i := 0; i < 16; i++ {
				id := 4*i + w
				scheduled := epoch.Add(time.Duration(id) * spacing)
				time.Sleep(time.Until(scheduled))
				begin := time.Now()
				_, e := s.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: fmt.Sprintf("read-%d-%d", w, i), Query: "public load fixture", AsOf: now.Add(time.Duration(4*i+w) * time.Microsecond), RecallK: 50, PackK: 10, TokenBudget: 10000})
				ns := time.Since(begin).Nanoseconds()
				mu.Lock()
				r.ReadNS = append(r.ReadNS, ns)
				r.ScheduledReadNS = append(r.ScheduledReadNS, time.Since(scheduled).Nanoseconds())
				if e != nil {
					r.Errors = append(r.Errors, e.Error())
				}
				mu.Unlock()
			}
		}(w)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 50; i < 66; i++ {
			begin := time.Now()
			e := put(i)
			ns := time.Since(begin).Nanoseconds()
			mu.Lock()
			r.WriteNS = append(r.WriteNS, ns)
			if e != nil {
				r.Errors = append(r.Errors, e.Error())
			}
			if readers.Load() > 0 {
				r.Overlap++
			}
			mu.Unlock()
			time.Sleep(2 * time.Millisecond)
		}
	}()
	epoch = time.Now()
	close(start)
	wg.Wait()
	deadline := time.Now().Add(5 * time.Second)
	for {
		r.Status = s.ResearchShadowStatus()
		v := r.Status
		if v.Completed+v.Stale+v.Failed+v.Cancelled == v.Accepted {
			break
		}
		if time.Now().After(deadline) {
			r.Errors = append(r.Errors, "drain timeout")
			break
		}
		time.Sleep(time.Millisecond)
	}
	ordered := append([]int64(nil), r.ReadNS...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	r.P99 = ordered[63]
	r.Started, r.Interrupted = started.Load(), interrupted.Load()
	r.Fits, r.Hits, r.Checked = counts()
	return r
}

func TestPacedFitExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PACED_FIT_OUT")
	if path == "" {
		t.Skip("opt-in load experiment")
	}
	old := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(old)
	if _, e := learners.ResearchTransformFixture(context.Background()); e != nil {
		t.Fatal(e)
	}
	var arms []pacedFitArm
	for _, rate := range []int{10, 40} {
		for trial := 0; trial < 3; trial++ {
			for order := 0; order < 2; order++ {
				arm := []string{"off", "table"}[(trial+order)%2]
				arms = append(arms, runPacedFitArm(t, trial, arm, rate))
			}
		}
	}
	hashes := map[string]string{}
	for _, file := range []string{"segment_paced_fit_test.go", "segment_transform_export_test.go", "segment_batch_research_test.go", "segment_table_research_test.go", "segment_table_service_helper_test.go", "segment_transform_memo_test.go", "segment_transform_context_test.go", "segment_transform_fullpair_test.go", "../service/research_shadow.go", "../service/service.go", "../../research/segment-paced-fit-protocol.md"} {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		hashes[file] = hex.EncodeToString(h[:])
	}
	data, e := json.MarshalIndent(struct {
		Go     string
		Procs  int
		Hashes map[string]string
		Arms   []pacedFitArm
	}{runtime.Version(), 4, hashes, arms}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	_, we := f.Write(append(data, '\n'))
	ce := f.Close()
	if we != nil {
		t.Fatal(we)
	}
	if ce != nil {
		t.Fatal(ce)
	}
	for _, r := range arms {
		t.Logf("trial%d arm%v p99_ms %.3f overlap%d completed%d stale%d drops%d errors%d", r.Trial, r.Arm, float64(r.P99)/1e6, r.Overlap, r.Status.Completed, r.Status.Stale, r.Status.Dropped, len(r.Errors))
	}
}
