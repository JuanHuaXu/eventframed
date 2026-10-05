package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type recallProfileKey struct{}

type recallProfileSpan struct {
	start, end time.Duration
}

type recallProfileTrace struct {
	origin time.Time
	mu     sync.Mutex
	spans  map[string][]recallProfileSpan
}

func (tr *recallProfileTrace) record(name string, start, end time.Time) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.spans[name] = append(tr.spans[name], recallProfileSpan{start.Sub(tr.origin), end.Sub(tr.origin)})
}

type recallProfileJournalSink struct {
	mu      sync.Mutex
	records map[string][]byte
}

func (s *recallProfileJournalSink) put(entry model.BayesianJournalEntry) error {
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if prior, exists := s.records[entry.ID]; exists && !bytes.Equal(prior, encoded) {
		return store.ErrJournalConflict
	}
	s.records[entry.ID] = encoded
	return nil
}

func (s *recallProfileJournalSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

type recallProfileStore struct {
	store.EventStore
	graphCache   *model.PredictiveGraph
	journalSink  *recallProfileJournalSink
	journalGuard interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	}
}

func (s recallProfileStore) record(ctx context.Context, name string, start time.Time) {
	if tr, ok := ctx.Value(recallProfileKey{}).(*recallProfileTrace); ok {
		tr.record(name, start, time.Now())
	}
}

func (s recallProfileStore) Search(ctx context.Context, tenant string, vector []float32, asOf time.Time, limit int) ([]store.SearchResult, error) {
	start := time.Now()
	defer s.record(ctx, "search", start)
	return s.EventStore.Search(ctx, tenant, vector, asOf, limit)
}

func (s recallProfileStore) Snapshot(ctx context.Context) model.Snapshot {
	start := time.Now()
	defer s.record(ctx, "snapshot", start)
	return s.EventStore.Snapshot(ctx)
}

func (s recallProfileStore) GetPredictiveGraph(ctx context.Context, tenant string) (model.PredictiveGraph, error) {
	start := time.Now()
	defer s.record(ctx, "graph", start)
	if s.graphCache != nil {
		return *s.graphCache, nil
	}
	return s.EventStore.GetPredictiveGraph(ctx, tenant)
}

func (s recallProfileStore) GetSelectionCertificate(ctx context.Context, tenant string) (model.SelectionSupportCertificate, error) {
	start := time.Now()
	defer s.record(ctx, "selection", start)
	return s.EventStore.GetSelectionCertificate(ctx, tenant)
}

func (s recallProfileStore) GetOmittedInfluenceCertificate(ctx context.Context, tenant string) (model.OmittedInfluenceCertificate, error) {
	start := time.Now()
	defer s.record(ctx, "omitted", start)
	return s.EventStore.GetOmittedInfluenceCertificate(ctx, tenant)
}

func (s recallProfileStore) GetAntiPigeonCertificate(ctx context.Context, tenant string, ids []string) (model.AntiPigeonCertificate, error) {
	start := time.Now()
	defer s.record(ctx, "anti_pigeon", start)
	return s.EventStore.GetAntiPigeonCertificate(ctx, tenant, ids)
}

func (s recallProfileStore) GetBayesianPosterior(ctx context.Context, tenant, key string) (model.BayesianPosterior, error) {
	start := time.Now()
	defer s.record(ctx, "posterior", start)
	return s.EventStore.GetBayesianPosterior(ctx, tenant, key)
}

func (s recallProfileStore) GetResidualCandidates(ctx context.Context, tenant, action, general string) (model.ResidualCandidates, error) {
	start := time.Now()
	defer s.record(ctx, "residual", start)
	return s.EventStore.GetResidualCandidates(ctx, tenant, action, general)
}

func (s recallProfileStore) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	start := time.Now()
	defer s.record(ctx, "journal", start)
	if s.journalSink != nil {
		if s.journalGuard != nil {
			return s.journalGuard.WithResearchAsOfSnapshotWait(ctx, entry.Snapshot, entry.AsOf, func() error {
				return s.journalSink.put(entry)
			})
		}
		return s.journalSink.put(entry)
	}
	return s.EventStore.PutBayesianJournal(ctx, entry)
}

type recallProfileSample struct {
	offer, call, queue int64
	spans              map[string][]recallProfileSpan
	err                error
}

type recallProfileRollup struct {
	counts, callDurations, requestSums, requestSpans []int64
}

type recallProfileTrial struct {
	offer, call, queue []int64
	parts              map[string]*recallProfileRollup
	writes, overlap    int
	batchSizes         []int
	guardRejects       int
	err                error
}

var recallProfileNames = []string{"search", "snapshot", "graph", "selection", "omitted", "anti_pigeon", "posterior", "residual", "journal", "sqlite_encode", "sqlite_lookup", "sqlite_guard_wait", "sqlite_guard_total", "sqlite_insert"}

func (r *recallProfileTrial) add(sample recallProfileSample) {
	r.offer = append(r.offer, sample.offer)
	r.call = append(r.call, sample.call)
	r.queue = append(r.queue, sample.queue)
	if sample.err != nil && r.err == nil {
		r.err = sample.err
	}
	for _, name := range recallProfileNames {
		part := r.parts[name]
		spans := sample.spans[name]
		part.counts = append(part.counts, int64(len(spans)))
		var sum, first, last int64
		for i, span := range spans {
			start, end := span.start.Nanoseconds(), span.end.Nanoseconds()
			part.callDurations = append(part.callDurations, end-start)
			sum += end - start
			if i == 0 || start < first {
				first = start
			}
			if i == 0 || end > last {
				last = end
			}
		}
		part.requestSums = append(part.requestSums, sum)
		part.requestSpans = append(part.requestSpans, last-first)
	}
}

func newRecallProfileTrial() recallProfileTrial {
	r := recallProfileTrial{parts: make(map[string]*recallProfileRollup, len(recallProfileNames))}
	for _, name := range recallProfileNames {
		r.parts[name] = &recallProfileRollup{}
	}
	return r
}

func (r *recallProfileTrial) merge(other recallProfileTrial) {
	r.offer = append(r.offer, other.offer...)
	r.call = append(r.call, other.call...)
	r.queue = append(r.queue, other.queue...)
	r.writes += other.writes
	r.overlap += other.overlap
	r.batchSizes = append(r.batchSizes, other.batchSizes...)
	r.guardRejects += other.guardRejects
	for _, name := range recallProfileNames {
		dst, src := r.parts[name], other.parts[name]
		dst.counts = append(dst.counts, src.counts...)
		dst.callDurations = append(dst.callDurations, src.callDurations...)
		dst.requestSums = append(dst.requestSums, src.requestSums...)
		dst.requestSpans = append(dst.requestSpans, src.requestSpans...)
	}
}

func runRecallProfileTrial(t *testing.T, writer bool, liveCount int) recallProfileTrial {
	return runRecallProfileTrialMode(t, writer, liveCount, "native")
}

func runRecallProfileTrialMode(t *testing.T, writer bool, liveCount int, mode string) recallProfileTrial {
	t.Helper()
	if mode != "native" && mode != "graph" && mode != "journal" && mode != "both" && mode != "guarded" && mode != "guarded-graph" && mode != "sqlite" && mode != "sqlite-graph" && mode != "group" && mode != "group-graph" {
		t.Fatal("unsupported phase-removal mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	s, tap := motionExitService(t, root, true)
	defer s.Close()
	defer tap.Close()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	expected := make(map[string]bool, liveCount)
	putTemporalFixture(t, s, "seed", now.Add(-time.Minute))
	expected["seed"] = true
	for i := 1; i < liveCount; i++ {
		id := fmt.Sprintf("live-%03d", i)
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
		expected[id] = true
	}
	profileStore := recallProfileStore{EventStore: s.store}
	var sqliteJournal *researchSQLiteJournalStore
	var groupJournal *researchSQLiteGroupJournalStore
	var journalPath string
	var err error
	if mode == "sqlite" || mode == "sqlite-graph" || mode == "group" || mode == "group-graph" {
		journalPath = filepath.Join(root, "journals.sqlite")
		sqliteJournal, err = openResearchSQLiteJournalStore(s.store, journalPath, "profile-owner", true)
		if err != nil {
			t.Fatal(err)
		}
		defer sqliteJournal.closeJournal()
		profileStore.EventStore = sqliteJournal
		if mode == "group" || mode == "group-graph" {
			groupJournal, err = newResearchSQLiteGroupJournalStore(sqliteJournal, 8, 8*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			defer groupJournal.closeGroup()
			profileStore.EventStore = groupJournal
		}
	}
	graphVersion := s.store.Snapshot(ctx).GraphVersion
	if mode == "graph" || mode == "both" || mode == "guarded-graph" || mode == "sqlite-graph" || mode == "group-graph" {
		graph, graphErr := s.store.GetPredictiveGraph(ctx, "tenant-a")
		if graphErr != nil {
			t.Fatal(graphErr)
		}
		profileStore.graphCache = &graph
	}
	if mode == "journal" || mode == "both" || mode == "guarded" || mode == "guarded-graph" {
		profileStore.journalSink = &recallProfileJournalSink{records: make(map[string][]byte, 192)}
	}
	if mode == "guarded" || mode == "guarded-graph" {
		guard, ok := s.store.(interface {
			WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
		})
		if !ok {
			t.Fatal("fixture lacks owned as-of publication guard")
		}
		profileStore.journalGuard = guard
	}
	probe, err := New(profileStore, s.embedder, Config{DefaultRecallK: liveCount, DefaultPackK: 10, DefaultTokenBudget: 1000})
	if err != nil {
		t.Fatal(err)
	}
	var active atomic.Bool
	active.Store(true)
	type job struct {
		i       int
		offered time.Time
	}
	jobs := make(chan job, 192)
	samples := make(chan recallProfileSample, 192)
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for work := range jobs {
				request := motionExitRequest(probe, now.Add(time.Duration(2*work.i)*time.Second))
				request.SessionID = fmt.Sprintf("profile-%d", work.i)
				request.RecallK, request.PackK = liveCount, 10
				start := time.Now()
				trace := &recallProfileTrace{origin: start, spans: make(map[string][]recallProfileSpan)}
				packet, callErr := probe.Recall(context.WithValue(ctx, recallProfileKey{}, trace), request)
				end := time.Now()
				if callErr == nil {
					seen := make(map[string]bool, liveCount)
					for _, decision := range packet.BayesianShadow.Decisions {
						if !expected[decision.EventID] || seen[decision.EventID] {
							callErr = fmt.Errorf("unexpected or duplicate nomination %s", decision.EventID)
							break
						}
						seen[decision.EventID] = true
					}
					if callErr == nil && len(seen) != liveCount {
						callErr = fmt.Errorf("nominated %d of %d live events", len(seen), liveCount)
					}
					for _, candidate := range packet.Candidates {
						if candidate.Event.AvailableAt.After(request.AsOf) {
							callErr = fmt.Errorf("future event %s in packet", candidate.Event.ID)
						}
					}
				}
				samples <- recallProfileSample{offer: end.Sub(work.offered).Nanoseconds(), call: end.Sub(start).Nanoseconds(), queue: start.Sub(work.offered).Nanoseconds(), spans: trace.spans, err: callErr}
			}
		}()
	}
	probeDone := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(8 * time.Millisecond)
		defer ticker.Stop()
		for i := range 192 {
			if i > 0 {
				select {
				case <-ctx.Done():
					close(jobs)
					workers.Wait()
					close(samples)
					probeDone <- struct{}{}
					return
				case <-ticker.C:
				}
			}
			jobs <- job{i: i, offered: time.Now()}
		}
		close(jobs)
		workers.Wait()
		close(samples)
		probeDone <- struct{}{}
	}()
	var writerDone <-chan struct {
		writes, overlap int
		err             error
	}
	if writer {
		writerDone = motionLoadWriter(ctx, s, now, &active)
	}
	r := newRecallProfileTrial()
	for sample := range samples {
		r.add(sample)
	}
	<-probeDone
	active.Store(false)
	if writer {
		got := <-writerDone
		r.writes, r.overlap = got.writes, got.overlap
		if got.err != nil && r.err == nil {
			r.err = got.err
		}
	}
	if r.err != nil || len(r.offer) != 192 || (writer && (r.writes != 256 || r.overlap == 0)) {
		t.Fatalf("profile incomplete: offers=%d writes=%d overlap=%d err=%v", len(r.offer), r.writes, r.overlap, r.err)
	}
	if profileStore.graphCache != nil && s.store.Snapshot(ctx).GraphVersion != graphVersion {
		t.Fatal("graph version changed during cached-graph diagnostic")
	}
	if profileStore.journalSink != nil && profileStore.journalSink.count() != 192 {
		t.Fatalf("journal sink retained %d of 192 entries", profileStore.journalSink.count())
	}
	if sqliteJournal != nil {
		if groupJournal != nil {
			if err := groupJournal.closeGroup(); err != nil {
				t.Fatal(err)
			}
			r.batchSizes = groupJournal.batchSizes()
			r.guardRejects = groupJournal.guardRejections()
		} else if err := sqliteJournal.closeJournal(); err != nil {
			t.Fatal(err)
		}
		reopened, openErr := openResearchSQLiteJournalStore(s.store, journalPath, "profile-owner", false)
		if openErr != nil {
			t.Fatal(openErr)
		}
		count, countErr := reopened.count(ctx)
		closeErr := reopened.closeJournal()
		if countErr != nil || closeErr != nil || count != 192 {
			t.Fatalf("durable journal reopen: count=%d read=%v close=%v", count, countErr, closeErr)
		}
	}
	return r
}

func TestResearchRecallStorePhaseProfileV17(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PROFILE_V17") != "1" {
		t.Skip("opt-in frozen v17 store-phase profile")
	}
	for _, liveCount := range []int{50, 200} {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all := map[bool]*recallProfileTrial{false: &quiet, true: &writing}
		for trial := range 3 {
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runRecallProfileTrial(t, writer, liveCount)
				all[writer].merge(r)
				t.Logf("live=%d writer=%v trial=%d offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s", liveCount, writer, trial, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99))
			}
		}
		for _, writer := range []bool{false, true} {
			arm := all[writer]
			t.Logf("live=%d writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s", liveCount, writer, len(arm.offer), arm.writes, researchDurableLoadPercentile(arm.offer, .99), researchDurableLoadPercentile(arm.call, .99), researchDurableLoadPercentile(arm.queue, .99))
			for _, name := range recallProfileNames {
				part := arm.parts[name]
				sort.Slice(part.counts, func(i, j int) bool { return part.counts[i] < part.counts[j] })
				if len(part.callDurations) == 0 {
					t.Logf("live=%d writer=%v phase=%s calls=0", liveCount, writer, name)
					continue
				}
				t.Logf("live=%d writer=%v phase=%s calls=%d count_p50=%d count_p99=%d individual_p99=%s span_p99=%s sum_p99=%s", liveCount, writer, name, len(part.callDurations), part.counts[len(part.counts)/2], part.counts[len(part.counts)-1], researchDurableLoadPercentile(part.callDurations, .99), researchDurableLoadPercentile(part.requestSpans, .99), researchDurableLoadPercentile(part.requestSums, .99))
			}
		}
	}
}

func TestResearchRecallPhaseRemovalV18(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_REMOVAL_V18") != "1" {
		t.Skip("opt-in frozen v18 phase-removal diagnostic")
	}
	arms := []string{"native", "graph", "journal", "both"}
	all := make(map[string]map[bool]*recallProfileTrial, len(arms))
	for _, arm := range arms {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all[arm] = map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runRecallProfileTrialMode(t, writer, 200, arm)
				all[arm][writer].merge(r)
				t.Logf("trial=%d arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s", trial, arm, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99))
			}
		}
	}
	for _, arm := range arms {
		for _, writer := range []bool{false, true} {
			r := all[arm][writer]
			if len(r.offer) != 576 || (writer && r.writes != 768) {
				t.Fatalf("arm=%s writer=%v incomplete: offers=%d writes=%d", arm, writer, len(r.offer), r.writes)
			}
			t.Logf("arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s search_p99=%s graph_p99=%s journal_p99=%s", arm, writer, len(r.offer), r.writes,
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99),
				researchDurableLoadPercentile(r.parts["search"].requestSpans, .99), researchDurableLoadPercentile(r.parts["graph"].requestSpans, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
		}
	}
}

func TestResearchRecallProfileInstrumentationConcurrent(t *testing.T) {
	sink := &recallProfileJournalSink{records: make(map[string][]byte)}
	trace := &recallProfileTrace{origin: time.Now(), spans: make(map[string][]recallProfileSpan)}
	const workers, perWorker = 16, 16
	var group sync.WaitGroup
	errorsOut := make(chan error, workers*perWorker)
	for worker := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for i := range perWorker {
				id := fmt.Sprintf("journal-%d-%d", worker, i)
				entry := model.BayesianJournalEntry{ID: id, TenantID: "tenant-a", SessionID: id}
				start := time.Now()
				if err := sink.put(entry); err != nil {
					errorsOut <- err
				}
				trace.record("journal", start, time.Now())
			}
		}()
	}
	group.Wait()
	close(errorsOut)
	for err := range errorsOut {
		t.Fatal(err)
	}
	if sink.count() != workers*perWorker || len(trace.spans["journal"]) != workers*perWorker {
		t.Fatalf("incomplete concurrent sink/trace: journals=%d spans=%d", sink.count(), len(trace.spans["journal"]))
	}
	prior := model.BayesianJournalEntry{ID: "journal-0-0", TenantID: "tenant-a", SessionID: "journal-0-0"}
	if err := sink.put(prior); err != nil {
		t.Fatalf("identical retry: %v", err)
	}
	prior.SessionID = "changed"
	if err := sink.put(prior); !errors.Is(err, store.ErrJournalConflict) {
		t.Fatalf("changed retry: got %v, want conflict", err)
	}
}

func TestResearchRecallJournalGuardCostV19(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_GUARD_V19") != "1" {
		t.Skip("opt-in frozen v19 journal-guard diagnostic")
	}
	arms := []string{"native", "journal", "guarded", "guarded-graph"}
	all := make(map[string]map[bool]*recallProfileTrial, len(arms))
	for _, arm := range arms {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all[arm] = map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runRecallProfileTrialMode(t, writer, 200, arm)
				all[arm][writer].merge(r)
				t.Logf("trial=%d arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s", trial, arm, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
			}
		}
	}
	for _, arm := range arms {
		for _, writer := range []bool{false, true} {
			r := all[arm][writer]
			if len(r.offer) != 576 || (writer && r.writes != 768) {
				t.Fatalf("arm=%s writer=%v incomplete: offers=%d writes=%d", arm, writer, len(r.offer), r.writes)
			}
			t.Logf("arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s search_p99=%s graph_p99=%s journal_p99=%s", arm, writer, len(r.offer), r.writes,
				researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99),
				researchDurableLoadPercentile(r.parts["search"].requestSpans, .99), researchDurableLoadPercentile(r.parts["graph"].requestSpans, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
		}
	}
}

func TestResearchRecallSQLiteJournalLoadV20(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_SQLITE_V20") != "1" {
		t.Skip("opt-in frozen v20 SQLite journal load screen")
	}
	arms := []string{"native", "sqlite", "sqlite-graph"}
	all := make(map[string]map[bool]*recallProfileTrial, len(arms))
	for _, arm := range arms {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all[arm] = map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runRecallProfileTrialMode(t, writer, 200, arm)
				all[arm][writer].merge(r)
				t.Logf("trial=%d arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s", trial, arm, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
			}
		}
	}
	for _, arm := range arms {
		for _, writer := range []bool{false, true} {
			r := all[arm][writer]
			if len(r.offer) != 576 || (writer && r.writes != 768) {
				t.Fatalf("arm=%s writer=%v incomplete: offers=%d writes=%d", arm, writer, len(r.offer), r.writes)
			}
			offerP99 := researchDurableLoadPercentile(r.offer, .99)
			t.Logf("arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s search_p99=%s graph_p99=%s journal_p99=%s", arm, writer, len(r.offer), r.writes,
				offerP99, researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99),
				researchDurableLoadPercentile(r.parts["search"].requestSpans, .99), researchDurableLoadPercentile(r.parts["graph"].requestSpans, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
			if writer && arm != "native" && offerP99 >= 100*time.Millisecond {
				t.Errorf("frozen v20 diagnostic latency screen failed: arm=%s offer_p99=%s", arm, offerP99)
			}
		}
	}
}

func TestResearchRecallSQLiteJournalPhasesV21(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_SQLITE_PHASES_V21") != "1" {
		t.Skip("opt-in frozen v21 SQLite journal phase profile")
	}
	quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
	all := map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	for trial := range 3 {
		order := []bool{false, true}
		if trial == 1 {
			order = []bool{true, false}
		}
		for _, writer := range order {
			r := runRecallProfileTrialMode(t, writer, 200, "sqlite")
			all[writer].merge(r)
			t.Logf("trial=%d writer=%v offers=%d writes=%d offer_p99=%s journal_p99=%s", trial, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
		}
	}
	for _, writer := range []bool{false, true} {
		r := all[writer]
		if len(r.offer) != 576 || (writer && r.writes != 768) {
			t.Fatalf("writer=%v incomplete: offers=%d writes=%d", writer, len(r.offer), r.writes)
		}
		t.Logf("writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s", writer, len(r.offer), r.writes,
			researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99))
		for _, name := range []string{"sqlite_encode", "sqlite_lookup", "sqlite_guard_wait", "sqlite_guard_total", "sqlite_insert"} {
			part := r.parts[name]
			if len(part.callDurations) != 576 {
				t.Fatalf("writer=%v phase=%s has %d spans, want 576", writer, name, len(part.callDurations))
			}
			t.Logf("writer=%v phase=%s calls=%d p50=%s p99=%s", writer, name, len(part.callDurations), researchDurableLoadPercentile(part.requestSpans, .5), researchDurableLoadPercentile(part.requestSpans, .99))
		}
	}
}

func TestResearchRecallGroupJournalLoadV22(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_GROUP_V22") != "1" {
		t.Skip("opt-in frozen v22 group-journal load screen")
	}
	arms := []string{"native", "sqlite", "group", "group-graph"}
	all := make(map[string]map[bool]*recallProfileTrial, len(arms))
	for _, arm := range arms {
		quiet, writing := newRecallProfileTrial(), newRecallProfileTrial()
		all[arm] = map[bool]*recallProfileTrial{false: &quiet, true: &writing}
	}
	for trial := range 3 {
		for offset := range arms {
			arm := arms[(trial+offset)%len(arms)]
			order := []bool{false, true}
			if trial == 1 {
				order = []bool{true, false}
			}
			for _, writer := range order {
				r := runRecallProfileTrialMode(t, writer, 200, arm)
				all[arm][writer].merge(r)
				t.Logf("trial=%d arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s batches=%d guard_rejects=%d", trial, arm, writer, len(r.offer), r.writes, researchDurableLoadPercentile(r.offer, .99), researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), len(r.batchSizes), r.guardRejects)
			}
		}
	}
	for _, arm := range arms {
		for _, writer := range []bool{false, true} {
			r := all[arm][writer]
			if len(r.offer) != 576 || (writer && r.writes != 768) {
				t.Fatalf("arm=%s writer=%v incomplete: offers=%d writes=%d", arm, writer, len(r.offer), r.writes)
			}
			offerP99 := researchDurableLoadPercentile(r.offer, .99)
			t.Logf("arm=%s writer=%v offers=%d writes=%d offer_p99=%s call_p99=%s queue_p99=%s journal_p99=%s guard_rejects=%d", arm, writer, len(r.offer), r.writes,
				offerP99, researchDurableLoadPercentile(r.call, .99), researchDurableLoadPercentile(r.queue, .99), researchDurableLoadPercentile(r.parts["journal"].requestSpans, .99), r.guardRejects)
			if arm == "group" || arm == "group-graph" {
				var histogram [9]int
				total := 0
				for _, size := range r.batchSizes {
					if size < 1 || size > 8 {
						t.Fatalf("invalid group size %d", size)
					}
					histogram[size]++
					total += size
				}
				if total != 576 || r.guardRejects != 0 {
					t.Fatalf("group journal incomplete: arm=%s writer=%v rows=%d rejects=%d", arm, writer, total, r.guardRejects)
				}
				t.Logf("arm=%s writer=%v group_sizes=%v", arm, writer, histogram)
			}
			if writer && (arm == "group" || arm == "group-graph") && offerP99 >= 100*time.Millisecond {
				t.Errorf("frozen v22 writer latency gate failed: arm=%s offer_p99=%s", arm, offerP99)
			}
		}
	}
}
