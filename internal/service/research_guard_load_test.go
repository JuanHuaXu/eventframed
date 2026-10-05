package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type researchGuardLoadResult struct {
	ResearchGroupCap                                             int
	FourNativeWriters                                            bool
	IdlePublicationWrapper                                       bool
	OfferedReadEveryNS, OfferedWriteEveryNS                      int64
	ReadSchedule, WriteSchedule                                  []researchScheduledCall
	ResolvedSourceAdmissions                                     bool
	CombinedSourceCleanup                                        bool
	BatchSourceReads                                             bool
	SourceOwner                                                  bool
	Mode                                                         string
	Trial, Requests, Overlap                                     int
	ReadNS, WriteNS, GuardNS, AgeNS                              []int64
	Attempts, Accepted, Busy, Stale, Validated, Admits, Discards uint64
	Dropped                                                      uint64
	WaitExpired                                                  uint64
	Errors                                                       []string
	Phases                                                       []researchGuardPhase
	GroupSizes                                                   []int
	PreparedWrites                                               bool
	UnionReads                                                   bool
}

// Durations use monotonic clocks. Entry includes guard waiting and its native
// snapshot check; it is not a pure mutex-wait measurement.
type researchGuardPhase struct {
	VerifiedDiscardNS                                 int64
	QueueNS, PrefetchNS, EntryNS, CallbackNS, TotalNS int64
	DurableAdmitNS, DurableVerifyNS, DurableDiscardNS int64
	PostGuardNS                                       int64
	Entered, Accepted                                 bool
}

func researchGuardLoadArm(t *testing.T, mode string, trial, requests int) researchGuardLoadResult {
	return researchGuardLoadProbeArm(t, mode, trial, requests, requests/2, true)
}

// The v19 entry point retains its original defaults; later diagnostics vary
// only writer count and journal-prefetch placement through this helper.
func researchGuardLoadProbeArm(t *testing.T, mode string, trial, requests, writes int, prefetch bool) researchGuardLoadResult {
	return researchGuardLoadStorageArm(t, mode, trial, requests, writes, prefetch, false)
}

func researchGuardLoadStorageArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared bool) researchGuardLoadResult {
	return researchGuardLoadValidationArm(t, mode, trial, requests, writes, prefetch, prepared, false)
}

func researchGuardLoadValidationArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union bool) researchGuardLoadResult {
	return researchGuardLoadSourceArm(t, mode, trial, requests, writes, prefetch, prepared, union, false)
}

func researchGuardLoadSourceArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source bool) researchGuardLoadResult {
	return researchGuardLoadSourceReadArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, false)
}

func researchGuardLoadSourceReadArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource bool) researchGuardLoadResult {
	return researchGuardLoadCleanupArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, false)
}

func researchGuardLoadCleanupArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined bool) researchGuardLoadResult {
	return researchGuardLoadResolvedArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, false)
}

func researchGuardLoadResolvedArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool) researchGuardLoadResult {
	return researchGuardLoadScheduledArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, resolved, nil)
}

func researchGuardLoadScheduledArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool, schedule *researchLoadSchedule) researchGuardLoadResult {
	return researchGuardLoadWrapperArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, resolved, schedule, false)
}

func researchGuardLoadWrapperArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool, schedule *researchLoadSchedule, idleWrapper bool) researchGuardLoadResult {
	return researchGuardLoadWritersArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, resolved, schedule, idleWrapper, false)
}

func researchGuardLoadWritersArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool, schedule *researchLoadSchedule, idleWrapper, fourWriters bool) researchGuardLoadResult {
	return researchGuardLoadGroupCapArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, resolved, schedule, idleWrapper, fourWriters, 0)
}

func researchGuardLoadGroupCapArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool, schedule *researchLoadSchedule, idleWrapper, fourWriters bool, groupCap int) researchGuardLoadResult {
	return researchGuardLoadMeasuredArm(t, mode, trial, requests, writes, prefetch, prepared, union, source, batchSource, combined, resolved, schedule, idleWrapper, fourWriters, groupCap, nil)
}

func researchGuardLoadMeasuredArm(t *testing.T, mode string, trial, requests, writes int, prefetch, prepared, union, source, batchSource, combined, resolved bool, schedule *researchLoadSchedule, idleWrapper, fourWriters bool, groupCap int, recorder *researchpublicationstore.GateRecorder) researchGuardLoadResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	r := researchGuardLoadResult{Mode: mode, Trial: trial, Requests: requests, PreparedWrites: prepared, UnionReads: union}
	r.SourceOwner = source
	r.BatchSourceReads = batchSource
	r.CombinedSourceCleanup = combined
	r.ResolvedSourceAdmissions = resolved
	r.IdlePublicationWrapper = idleWrapper
	r.FourNativeWriters = fourWriters
	r.ResearchGroupCap = groupCap
	if groupCap < 0 || groupCap > 4 || (groupCap != 0 && (!resolved || mode != "group4postverify")) {
		t.Fatal("group cap override requires resolved source and cap1-4")
	}
	if idleWrapper && mode != "off" {
		t.Fatal("idle wrapper requires disabled consumer")
	}
	if schedule != nil {
		if schedule.ReadEvery <= 0 || schedule.WriteEvery <= 0 || schedule.ReadEvery > time.Second || schedule.WriteEvery > time.Second || requests <= 0 || requests%4 != 0 || writes < 0 {
			t.Fatal("invalid offered schedule")
		}
		r.OfferedReadEveryNS = int64(schedule.ReadEvery)
		r.OfferedWriteEveryNS = int64(schedule.WriteEvery)
	}
	if resolved && !combined {
		t.Fatal("resolved admission study requires combined source cleanup")
	}
	if combined && !batchSource {
		t.Fatal("verified cleanup study requires batch source owner")
	}
	if batchSource && !source {
		t.Fatal("batch source reads require source owner")
	}
	if source && (mode != "group4postverify" || !prepared || !union) {
		t.Fatal("source owner requires prepared union postverify")
	}
	if union && mode != "group4postverify" && mode != "group4" {
		t.Fatal("union study requires postverify mode")
	}
	if prepared && mode != "group4postverify" {
		t.Fatal("prepared study requires postverify mode")
	}
	em, err := embed.NewHashEmbedder(32)
	if err != nil {
		t.Fatal(err)
	}
	open := libravdbstore.Open
	if fourWriters {
		open = libravdbstore.OpenResearchFourWriters
	}
	db, err := open(libravdbstore.Config{Path: t.TempDir() + "/load.libravdb", Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	tap, err := NewResearchFrontierTap("tenant-a", 64)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer tap.Close()
	config := Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000}
	if mode != "off" {
		config.ResearchFrontier = tap
	}
	s, err := New(db, em, config)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	put := func(i int) error {
		at := now.Add(-time.Second)
		if i >= 50 {
			at = now.Add(time.Hour)
		}
		ev := testutil.Event(fmt.Sprintf("guard-load-%d", i), "public load fixture", at)
		_, e := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev})
		return e
	}
	for i := 0; i < 50; i++ {
		if err = put(i); err != nil {
			t.Fatal(err)
		}
	}
	if mode != "off" || idleWrapper {
		if recorder == nil {
			s.store, err = researchpublicationstore.Wrap(db)
		} else {
			s.store, err = researchpublicationstore.WrapMeasured(db, recorder)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	var durable *researchmemory.Durable
	var sourceOwner *researchmemory.SourceOwner
	if source {
		open := researchmemory.OpenSourceOwner
		if batchSource {
			open = researchmemory.OpenSourceOwnerBatchReads
		}
		if resolved {
			open = researchmemory.OpenSourceOwnerResolvedAdmissions
		}
		sourceOwner, err = open(ctx, t.TempDir()+"/source.sqlite", "tenant-a", "load", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		defer sourceOwner.Close()
	} else if mode == "durable" || mode == "group4durable" || mode == "group4transactions" || mode == "group4reads" || mode == "group4postdiscard" || mode == "group4postverify" {
		open := researchmemory.OpenDurable
		if prepared {
			open = researchmemory.OpenDurablePreparedBatches
		}
		durable, err = open(ctx, t.TempDir()+"/load.sqlite", "tenant-a", "load", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		defer durable.Close()
	}
	var mu sync.Mutex
	addError := func(e error) { mu.Lock(); r.Errors = append(r.Errors, e.Error()); mu.Unlock() }
	consumerDone := make(chan struct{})
	if mode == "off" {
		close(consumerDone)
	} else if mode == "group1" || mode == "group4" || mode == "group4durable" || mode == "group4transactions" || mode == "group4reads" || mode == "group4postdiscard" || mode == "group4postverify" {
		limit := 1
		if mode == "group4" || mode == "group4durable" || mode == "group4transactions" || mode == "group4reads" || mode == "group4postdiscard" || mode == "group4postverify" {
			limit = 4
		}
		if groupCap != 0 {
			limit = groupCap
		}
		go func() {
			defer close(consumerDone)
			researchGroupedSourceLoadConsumer(ctx, s, tap, now, limit, durable, sourceOwner, &r, addError)
		}()
	} else {
		go func() {
			defer close(consumerDone)
			a := researchmemory.New(1, 42)
			guard := s.store.(interface {
				WithResearchSnapshot(context.Context, model.Snapshot, func() error) error
			})
			guardCall := guard.WithResearchSnapshot
			if mode == "queued" {
				queued := s.store.(interface {
					WithResearchSnapshotWait(context.Context, model.Snapshot, func() error) error
				})
				guardCall = func(parent context.Context, snapshot model.Snapshot, work func() error) error {
					// This budget limits entry waiting; callback I/O retains the
					// enclosing fixture context rather than a new timeout policy.
					waitCtx, cancelWait := context.WithTimeout(parent, 20*time.Millisecond)
					defer cancelWait()
					return queued.WithResearchSnapshotWait(waitCtx, snapshot, work)
				}
			}
			if mode == "asof" || mode == "batch" {
				asOfGuard := s.store.(interface {
					WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
				})
				guardCall = func(parent context.Context, snapshot model.Snapshot, work func() error) error {
					waitCtx, cancelWait := context.WithTimeout(parent, 20*time.Millisecond)
					defer cancelWait()
					return asOfGuard.WithResearchAsOfSnapshotWait(waitCtx, snapshot, now, work)
				}
			}
			for {
				in, e := tap.Take(ctx)
				if e != nil {
					if ctx.Err() != nil {
						addError(e)
					}
					return
				}
				r.Attempts++
				phase := researchGuardPhase{QueueNS: time.Since(in.queuedAt).Nanoseconds()}
				prefetchStart := time.Now()
				var j model.BayesianJournalEntry
				if prefetch {
					j, e = s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
					if e != nil {
						addError(e)
						return
					}
				}
				request := model.RecallRequest{TenantID: "tenant-a", SessionID: j.SessionID, Query: "public load fixture", AsOf: now}
				phase.PrefetchNS = time.Since(prefetchStart).Nanoseconds()
				start := time.Now()
				e = guardCall(ctx, in.Snapshot, func() error {
					phase.Entered = true
					phase.EntryNS = time.Since(start).Nanoseconds()
					callbackStart := time.Now()
					defer func() { phase.CallbackNS = time.Since(callbackStart).Nanoseconds() }()
					if !prefetch {
						var readErr error
						j, readErr = s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
						if readErr != nil {
							return readErr
						}
						request.SessionID = j.SessionID
					}
					if len(in.Candidates) != 50 {
						return fmt.Errorf("unexpected frontier %d", len(in.Candidates))
					}
					var batch []researchmemory.RecordedPrediction
					if mode == "batch" {
						batch = make([]researchmemory.RecordedPrediction, 0, len(in.Candidates))
					}
					for _, c := range in.Candidates {
						p, e := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
						if e != nil {
							return e
						}
						record, e := a.Record(p.ID)
						if e != nil {
							return e
						}
						record.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
						if mode == "batch" {
							batch = append(batch, record)
							continue
						}
						if e = s.ValidateResearchAdmission(ctx, record, request); e != nil {
							return e
						}
						r.Validated++
						if durable != nil {
							if _, _, e = durable.AdmitBound(ctx, p.ID, c.Features, c.Baseline, in.AsOf, *record.Binding); e != nil {
								return e
							}
							r.Admits++
							if _, e = durable.Discard(ctx, p.ID, in.AsOf); e != nil {
								return e
							}
							r.Discards++
						}
						a.Discard(p.ID)
					}
					if mode == "batch" {
						if e := s.ValidateResearchAdmissionBatch(ctx, batch, request); e != nil {
							return e
						}
						r.Validated += uint64(len(batch))
						for _, record := range batch {
							a.Discard(record.Prediction.ID)
						}
					}
					return nil
				})
				phase.TotalNS = time.Since(start).Nanoseconds()
				phase.Accepted = e == nil
				r.Phases = append(r.Phases, phase)
				r.GuardNS = append(r.GuardNS, phase.TotalNS)
				switch {
				case e == nil:
					r.Accepted++
					r.AgeNS = append(r.AgeNS, time.Since(in.queuedAt).Nanoseconds())
				case strings.Contains(e.Error(), "guard busy"):
					r.Busy++
				case strings.Contains(e.Error(), "stale or quarantined"):
					r.Stale++
				case errors.Is(e, context.DeadlineExceeded) && ctx.Err() == nil:
					r.WaitExpired++
				default:
					addError(e)
					return
				}
			}
		}()
	}
	start := make(chan struct{})
	var loadStart time.Time // Published to all producers by closing start.
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
				index := i*4 + w
				var due time.Time
				if schedule != nil {
					due = loadStart.Add(time.Duration(index) * schedule.ReadEvery)
					if e := researchAwaitDue(ctx, due); e != nil {
						addError(e)
						return
					}
				}
				begin := time.Now()
				_, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: fmt.Sprintf("read-%d-%d", w, i), Query: "public load fixture", AsOf: now, RecallK: 50, PackK: 10, TokenBudget: 10000})
				end := time.Now()
				ns := end.Sub(begin).Nanoseconds()
				mu.Lock()
				r.ReadNS = append(r.ReadNS, ns)
				if schedule != nil {
					r.ReadSchedule = append(r.ReadSchedule, researchScheduledCall{index, due.Sub(loadStart).Nanoseconds(), begin.Sub(loadStart).Nanoseconds(), end.Sub(loadStart).Nanoseconds()})
				}
				mu.Unlock()
				if e != nil {
					addError(e)
				}
			}
		}(w)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 50; i < 50+writes; i++ {
			var due time.Time
			if schedule != nil {
				due = loadStart.Add(time.Duration(i-50) * schedule.WriteEvery)
				if e := researchAwaitDue(ctx, due); e != nil {
					addError(e)
					return
				}
			}
			begin := time.Now()
			e := put(i)
			end := time.Now()
			ns := end.Sub(begin).Nanoseconds()
			mu.Lock()
			r.WriteNS = append(r.WriteNS, ns)
			if schedule != nil {
				r.WriteSchedule = append(r.WriteSchedule, researchScheduledCall{i - 50, due.Sub(loadStart).Nanoseconds(), begin.Sub(loadStart).Nanoseconds(), end.Sub(loadStart).Nanoseconds()})
			}
			if readers.Load() > 0 {
				r.Overlap++
			}
			mu.Unlock()
			if e != nil {
				addError(e)
			}
			if schedule == nil {
				time.Sleep(2 * time.Millisecond)
			}
		}
	}()
	loadStart = time.Now()
	close(start)
	wg.Wait()
	tap.Close()
	<-consumerDone
	r.Dropped = tap.Dropped()
	return r
}

func checkGuardLoadAccounting(t *testing.T, r researchGuardLoadResult) {
	t.Helper()
	if len(r.Errors) > 0 || len(r.ReadNS) != r.Requests || len(r.WriteNS) != r.Requests/2 {
		t.Errorf("load errors/accounting: %+v", r)
		return
	}
	if r.Mode == "off" {
		return
	}
	if r.Attempts+r.Dropped != uint64(r.Requests) || r.Attempts != r.Accepted+r.Busy+r.Stale || r.Validated != 50*r.Accepted {
		t.Errorf("guard accounting: %+v", r)
	}
	if r.Mode == "durable" && (r.Admits != r.Validated || r.Discards != r.Admits) {
		t.Errorf("ledger accounting: %+v", r)
	}
}

func TestResearchGuardLoadAccounting(t *testing.T) {
	for _, mode := range []string{"off", "validate", "durable"} {
		t.Run(mode, func(t *testing.T) { checkGuardLoadAccounting(t, researchGuardLoadArm(t, mode, 0, 8)) })
	}
}

func TestResearchGuardLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_GUARD_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in guarded admission load")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/service/service.go", "internal/service/research_guard_load_test.go", "internal/service/research_durable_validation.go", "internal/service/research_frontier.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchpublicationstore/mutations.go", "internal/researchpublicationstore/capabilities.go", "internal/researchpublication/publication.go", "internal/researchmemory/adapter.go", "internal/researchmemory/durable.go", "internal/researchmemory/durable_discard.go", "internal/researchmemory/background.go", "internal/researchmemory/record.go", "internal/researchledger/ledger.go", "docs/experiments/mmm-guard-load-v19-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"kind": "header", "expected_arms": 9, "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "NumCPU": runtime.NumCPU(), "GOMAXPROCS": runtime.GOMAXPROCS(0)}); e != nil {
		t.Fatal(e)
	}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 3; order++ {
			mode := []string{"off", "validate", "durable"}[(trial+order)%3]
			r := researchGuardLoadArm(t, mode, trial, 192)
			if e = enc.Encode(r); e != nil {
				t.Fatal(e)
			}
			if e = f.Sync(); e != nil {
				t.Fatal(e)
			}
			checkGuardLoadAccounting(t, r)
			ordered := append([]int64(nil), r.ReadNS...)
			sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
			t.Logf("trial%d %s accepted%d busy%d stale%d dropped%d read_p99_ms%.3f", trial, mode, r.Accepted, r.Busy, r.Stale, r.Dropped, float64(ordered[int(math.Ceil(.99*float64(len(ordered))))-1])/1e6)
		}
	}
}
