package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type publishedRecallLoadStoreV15 struct {
	store.EventStore
	gate        *incrementalSortGate
	owner       sync.Mutex
	current     atomic.Pointer[publishedRecallViewV8]
	searches    atomic.Int64
	stales      atomic.Int64
	queueMu     sync.Mutex
	closed      bool
	journalJobs chan *batchJournalJobV15
	workerEnd   chan struct{}
	statsMu     sync.Mutex
	batchSizes  []int
	batchNS     []int64
	gateNS      []int64
	ownerNS     []int64
	queueNS     []int64
}

type batchJournalJobV15 struct {
	entry    model.BayesianJournalEntry
	enqueued time.Time
	done     chan error
}

func (s *publishedRecallLoadStoreV15) publishLocked(ctx context.Context) error {
	lsn, ready := s.gate.capture(ctx)
	if !ready {
		s.current.Store(nil)
		return fmt.Errorf("journaled EventFrame view is not READY")
	}
	s.current.Store(&publishedRecallViewV8{lsn: lsn, snapshot: s.gate.store.Snapshot(ctx), at: time.Now()})
	return nil
}

func (s *publishedRecallLoadStoreV15) appendBatch(ctx context.Context, writes []ResearchEventWrite) (model.Snapshot, error) {
	s.owner.Lock()
	defer s.owner.Unlock()
	if err := appendDenseV6(ctx, s.gate, writes); err != nil {
		s.current.Store(nil)
		return model.Snapshot{}, err
	}
	if err := s.publishLocked(ctx); err != nil {
		return model.Snapshot{}, err
	}
	return s.current.Load().snapshot, nil
}

func (s *publishedRecallLoadStoreV15) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	if tenantID != "tenant-a" {
		return nil, fmt.Errorf("unexpected tenant %q", tenantID)
	}
	pin, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if pin == nil {
		return nil, fmt.Errorf("Recall has no per-call published view pin")
	}
	view := s.current.Load()
	if view == nil || time.Since(view.at) >= 250*time.Millisecond {
		return nil, fmt.Errorf("published view unavailable or expired")
	}
	results, err := queryDensePublishedV6(ctx, s.gate, view.lsn, availableBy, vector, limit)
	if err != nil {
		return nil, err
	}
	if time.Since(view.at) >= 250*time.Millisecond || s.current.Load() == nil {
		return nil, fmt.Errorf("published view expired during Search")
	}
	pin.view = view
	s.searches.Add(1)
	return results, nil
}

func (s *publishedRecallLoadStoreV15) Snapshot(ctx context.Context) model.Snapshot {
	if pin, ok := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8); ok && pin.view != nil {
		return pin.view.snapshot
	}
	return s.gate.store.Snapshot(ctx)
}

func (s *publishedRecallLoadStoreV15) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	job := &batchJournalJobV15{entry: entry, enqueued: time.Now(), done: make(chan error, 1)}
	s.queueMu.Lock()
	if s.closed {
		s.queueMu.Unlock()
		return fmt.Errorf("journal batch worker is closed")
	}
	select {
	case s.journalJobs <- job:
	case <-ctx.Done():
		s.queueMu.Unlock()
		return ctx.Err()
	}
	s.queueMu.Unlock()
	select {
	case err := <-job.done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *publishedRecallLoadStoreV15) runJournalWorker() {
	defer close(s.workerEnd)
	for first := range s.journalJobs {
		batch := []*batchJournalJobV15{first}
		timer := time.NewTimer(time.Millisecond)
		closed := false
	collect:
		for len(batch) < 4 {
			select {
			case next, ok := <-s.journalJobs:
				if !ok {
					closed = true
					break collect
				}
				batch = append(batch, next)
			case <-timer.C:
				break collect
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		started := time.Now()
		entries := make([]model.BayesianJournalEntry, len(batch))
		for index, job := range batch {
			entries[index] = job.entry
		}
		workCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		ownerStarted := time.Now()
		s.owner.Lock()
		ownerWait := time.Since(ownerStarted)
		gateStarted := time.Now()
		err := s.gate.appendJournalBatchV15(workCtx, entries, "", 0)
		gateElapsed := time.Since(gateStarted)
		if err == nil {
			err = s.publishLocked(workCtx)
		} else {
			s.current.Store(nil)
		}
		s.owner.Unlock()
		cancel()
		if errors.Is(err, store.ErrStaleSnapshot) {
			s.stales.Add(int64(len(batch)))
		}
		s.statsMu.Lock()
		s.batchSizes = append(s.batchSizes, len(batch))
		s.batchNS = append(s.batchNS, time.Since(started).Nanoseconds())
		s.gateNS = append(s.gateNS, gateElapsed.Nanoseconds())
		s.ownerNS = append(s.ownerNS, ownerWait.Nanoseconds())
		for _, job := range batch {
			s.queueNS = append(s.queueNS, started.Sub(job.enqueued).Nanoseconds())
		}
		s.statsMu.Unlock()
		for _, job := range batch {
			job.done <- err
		}
		if closed {
			return
		}
	}
}

func (s *publishedRecallLoadStoreV15) Close() error {
	s.queueMu.Lock()
	if !s.closed {
		s.closed = true
		close(s.journalJobs)
	}
	s.queueMu.Unlock()
	<-s.workerEnd
	return nil
}

type recallLoadOfferV15 struct {
	index   int
	offered time.Time
}

type recallLoadResultV15 struct {
	index                   int
	offered, started, ended time.Time
	admissionWait           time.Duration
	view                    *publishedRecallViewV8
	packet                  model.ContextPacket
	err                     error
}

type recallLoadAckV15 struct {
	id      string
	offered time.Time
	acked   time.Time
	version uint64
}

type recallLoadWriterV15 struct {
	acks           []recallLoadAckV15
	groups         []int
	leaseWaitNanos []int64
	err            error
}

func runPublishedRecallLoadV15(t *testing.T, raceCorrectnessOnly bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createUnpublishedDenseGateV8(t, t.TempDir(), query)
	defer gate.close()
	adapter := &publishedRecallLoadStoreV15{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV15, 128), workerEnd: make(chan struct{})}
	var admission sync.RWMutex
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	go adapter.runJournalWorker()
	defer svc.Close()
	if err := gate.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := adapter.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	allRows := []denseOracleRowV7{{"past100", 0}, {"at120", 0}}
	for start := 0; start < 198; start += 16 {
		batch := make([]ResearchEventWrite, 0, min(16, 198-start))
		for i := start; i < min(198, start+16); i++ {
			angle := 0.005 * float64(i+1)
			id := fmt.Sprintf("eligible-%03d", i)
			write := pinnedWrite(id, second.Add(100*time.Millisecond))
			write.Vector = denseRowV6(query, i, angle)
			batch = append(batch, write)
			allRows = append(allRows, denseOracleRowV7{id, angle})
		}
		if _, err := adapter.appendBatch(ctx, batch); err != nil {
			t.Fatal("seed eligible", err)
		}
	}
	for start := 0; start < 16; start += 16 {
		batch := make([]ResearchEventWrite, 0, 16)
		for i := start; i < start+16; i++ {
			write := pinnedWrite(fmt.Sprintf("future-%03d", i), asOf.Add(5*time.Millisecond))
			write.Vector = denseRowV6(query, i, 0.001*float64(i+1))
			batch = append(batch, write)
		}
		if _, err := adapter.appendBatch(ctx, batch); err != nil {
			t.Fatal("seed future", err)
		}
	}
	seedVersion := adapter.current.Load().snapshot.RuntimeVersion
	oracleByVersion := map[uint64][]denseOracleRowV7{seedVersion: sortedDenseOracleV7(allRows)}
	writeJobs := make(chan sortLoadWriteJob, 128)
	readJobs := make(chan recallLoadOfferV15, 128)
	writeOffers, readOffers := make(chan []time.Time, 1), make(chan []time.Time, 1)
	writerDone := make(chan recallLoadWriterV15, 1)
	readReports := make(chan recallLoadResultV15, 128)
	start := make(chan struct{})
	go func() {
		<-start
		defer close(writeJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 128)
		defer func() { writeOffers <- times }()
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			offered := time.Now()
			times = append(times, offered)
			write := pinnedWrite(fmt.Sprintf("visible-%03d", i), asOf)
			write.Vector = denseRowV6(query, i, 0.00213*float64(i+1))
			select {
			case writeJobs <- sortLoadWriteJob{write: write, offered: offered}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		defer close(readJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 128)
		defer func() { readOffers <- times }()
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			offered := time.Now()
			times = append(times, offered)
			select {
			case readJobs <- recallLoadOfferV15{index: i, offered: offered}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		report := recallLoadWriterV15{}
		for {
			group, closed := sortLoadTakeGroup(writeJobs, 16, 16*time.Millisecond)
			if len(group) == 0 {
				writerDone <- report
				return
			}
			batch := make([]ResearchEventWrite, len(group))
			for i, job := range group {
				batch[i] = job.write
			}
			waiting := time.Now()
			admission.Lock()
			leaseWait := time.Since(waiting)
			snapshot, err := adapter.appendBatch(ctx, batch)
			admission.Unlock()
			report.leaseWaitNanos = append(report.leaseWaitNanos, leaseWait.Nanoseconds())
			if err != nil {
				report.err = err
				writerDone <- report
				return
			}
			for _, job := range group {
				var index int
				if _, err := fmt.Sscanf(strings.TrimPrefix(job.write.Event.ID, "visible-"), "%d", &index); err != nil {
					report.err = err
					writerDone <- report
					return
				}
				allRows = append(allRows, denseOracleRowV7{job.write.Event.ID, 0.00213 * float64(index+1)})
			}
			oracleByVersion[snapshot.RuntimeVersion] = sortedDenseOracleV7(allRows)
			acked := time.Now()
			report.groups = append(report.groups, len(group))
			for _, job := range group {
				report.acks = append(report.acks, recallLoadAckV15{job.write.Event.ID, job.offered, acked, snapshot.RuntimeVersion})
			}
			if closed {
				writerDone <- report
				return
			}
		}
	}()
	var readers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			for offer := range readJobs {
				began := time.Now()
				pin := &publishedRecallPinV8{}
				callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
				request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion,
					TenantID: "tenant-a", SessionID: fmt.Sprintf("lab-%03d", offer.index),
					Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
					AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 10000}
				waiting := time.Now()
				admission.RLock()
				leaseWait := time.Since(waiting)
				packet, err := svc.Recall(callCtx, request)
				admission.RUnlock()
				readReports <- recallLoadResultV15{index: offer.index, offered: offer.offered, started: began,
					ended: time.Now(), admissionWait: leaseWait, view: pin.view, packet: packet, err: err}
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	close(start)
	reports := make([]recallLoadResultV15, 0, 128)
	callNS, offerNS, viewAgeNS := make([]int64, 0, 128), make([]int64, 0, 128), make([]int64, 0, 128)
	readLeaseNS := make([]int64, 0, 128)
	for report := range readReports {
		reports = append(reports, report)
		callNS = append(callNS, report.ended.Sub(report.started).Nanoseconds())
		offerNS = append(offerNS, report.ended.Sub(report.offered).Nanoseconds())
		readLeaseNS = append(readLeaseNS, report.admissionWait.Nanoseconds())
		if report.view != nil {
			viewAgeNS = append(viewAgeNS, report.ended.Sub(report.view.at).Nanoseconds())
		}
	}
	writer := <-writerDone
	writeTimes, readTimes := <-writeOffers, <-readOffers
	violations := make([]string, 0)
	if writer.err != nil {
		violations = append(violations, "writer: "+writer.err.Error())
	}
	if len(writer.acks) != 128 || len(reports) != 128 || len(writeTimes) != 128 || len(readTimes) != 128 {
		violations = append(violations, fmt.Sprintf("counts ack=%d recall=%d write_offer=%d read_offer=%d", len(writer.acks), len(reports), len(writeTimes), len(readTimes)))
	}
	for _, report := range reports {
		if report.err != nil {
			violations = append(violations, fmt.Sprintf("recall %d: %v", report.index, report.err))
			continue
		}
		if report.view == nil || report.packet.Snapshot != report.view.snapshot {
			violations = append(violations, fmt.Sprintf("recall %d lost per-call pinned snapshot", report.index))
			continue
		}
		journal, err := gate.store.GetBayesianJournal(ctx, "tenant-a", report.packet.BayesianShadow.JournalID)
		if err != nil || journal.Snapshot != report.packet.Snapshot || journal.Report.JournalID != report.packet.BayesianShadow.JournalID {
			violations = append(violations, fmt.Sprintf("recall %d missing/mismatched durable journal: %v", report.index, err))
			continue
		}
		expected, ok := oracleByVersion[report.packet.Snapshot.RuntimeVersion]
		if !ok || len(expected) < 150 || report.packet.Recalled != 150 || len(journal.Report.Decisions) != 150 {
			violations = append(violations, fmt.Sprintf("recall %d missing 150-row version oracle/frontier", report.index))
			continue
		}
		want := make(map[string]bool, 150)
		for _, row := range expected[:150] {
			want[row.id] = true
		}
		for _, decision := range journal.Report.Decisions {
			if !want[decision.EventID] || strings.HasPrefix(decision.EventID, "future") {
				violations = append(violations, fmt.Sprintf("recall %d bad nominated event %s", report.index, decision.EventID))
			}
			delete(want, decision.EventID)
		}
		if len(want) != 0 {
			violations = append(violations, fmt.Sprintf("recall %d omitted %d oracle nominees", report.index, len(want)))
		}
		for _, candidate := range report.packet.Candidates {
			if candidate.Event.AvailableAt.After(asOf) || strings.HasPrefix(candidate.Event.ID, "future") {
				violations = append(violations, fmt.Sprintf("recall %d packed future event %s", report.index, candidate.Event.ID))
			}
		}
		latestBeforeOffer := seedVersion
		for _, ack := range writer.acks {
			if !ack.acked.After(report.offered) && ack.version > latestBeforeOffer {
				latestBeforeOffer = ack.version
			}
		}
		if report.packet.Snapshot.RuntimeVersion < latestBeforeOffer {
			violations = append(violations, fmt.Sprintf("recall %d omitted an acknowledged-before-offer version", report.index))
		}
	}
	final := adapter.current.Load()
	if final == nil || final.snapshot.RuntimeVersion != seedVersion+128 {
		violations = append(violations, "final published snapshot did not include all visible writes")
	}
	if _, ready := gate.capture(ctx); !ready {
		violations = append(violations, "final journal publication gate is not READY")
	}
	if stale := adapter.stales.Load(); stale != 0 {
		violations = append(violations, fmt.Sprintf("read admission still allowed %d stale-snapshot rejections", stale))
	}
	writeAgeNS := make([]int64, 0, len(writer.acks))
	for _, ack := range writer.acks {
		writeAgeNS = append(writeAgeNS, ack.acked.Sub(ack.offered).Nanoseconds())
	}
	writeP99 := pinnedPercentile(writeAgeNS, .99)
	callP99 := pinnedPercentile(callNS, .99)
	offerP99 := pinnedPercentile(offerNS, .99)
	maxViewAge := pinnedPercentile(viewAgeNS, 1)
	adapter.statsMu.Lock()
	batchSizes := append([]int(nil), adapter.batchSizes...)
	batchNS := append([]int64(nil), adapter.batchNS...)
	gateNS := append([]int64(nil), adapter.gateNS...)
	ownerNS := append([]int64(nil), adapter.ownerNS...)
	queueNS := append([]int64(nil), adapter.queueNS...)
	adapter.statsMu.Unlock()
	var batchHistogram [5]int
	var batchedEntries int
	for _, size := range batchSizes {
		if size < 1 || size > 4 {
			violations = append(violations, fmt.Sprintf("invalid journal batch size %d", size))
			continue
		}
		batchHistogram[size]++
		batchedEntries += size
	}
	if batchedEntries != 128 || len(queueNS) != 128 {
		violations = append(violations, fmt.Sprintf("journal batch accounting entries=%d queue_samples=%d", batchedEntries, len(queueNS)))
	}
	t.Logf("journal_batch_histogram_1_to_4=%v batches=%d queue_wait_p50=%s queue_wait_p99=%s owner_wait_p50=%s owner_wait_p99=%s gate_cost_p50=%s gate_cost_p99=%s batch_cost_p50=%s batch_cost_p99=%s",
		batchHistogram[1:], len(batchNS), pinnedPercentile(queueNS, .5), pinnedPercentile(queueNS, .99),
		pinnedPercentile(ownerNS, .5), pinnedPercentile(ownerNS, .99),
		pinnedPercentile(gateNS, .5), pinnedPercentile(gateNS, .99),
		pinnedPercentile(batchNS, .5), pinnedPercentile(batchNS, .99))
	if !raceCorrectnessOnly && (writeP99 >= 250*time.Millisecond || callP99 >= 100*time.Millisecond ||
		offerP99 >= 100*time.Millisecond || maxViewAge >= 250*time.Millisecond) {
		violations = append(violations, fmt.Sprintf("normal latency/freshness gate write_p99=%s call_p99=%s offer_p99=%s view_age_max=%s", writeP99, callP99, offerP99, maxViewAge))
	}
	t.Logf("mode_race_correctness_only=%v writes=%d groups=%d recalls=%d violations=%d searches=%d stale_rejects=%d write_gap_p50=%s write_gap_p99=%s read_gap_p50=%s read_gap_p99=%s writer_lease_p50=%s writer_lease_p99=%s reader_lease_p50=%s reader_lease_p99=%s write_age_p50=%s write_age_p99=%s call_p50=%s call_p99=%s offer_p50=%s offer_p99=%s view_age_p50=%s view_age_max=%s",
		raceCorrectnessOnly, len(writer.acks), len(writer.groups), len(reports), len(violations),
		adapter.searches.Load(), adapter.stales.Load(),
		pinnedPercentile(sortLoadOfferGaps(writeTimes), .5), pinnedPercentile(sortLoadOfferGaps(writeTimes), .99),
		pinnedPercentile(sortLoadOfferGaps(readTimes), .5), pinnedPercentile(sortLoadOfferGaps(readTimes), .99),
		pinnedPercentile(writer.leaseWaitNanos, .5), pinnedPercentile(writer.leaseWaitNanos, .99),
		pinnedPercentile(readLeaseNS, .5), pinnedPercentile(readLeaseNS, .99),
		pinnedPercentile(writeAgeNS, .5), writeP99,
		pinnedPercentile(callNS, .5), callP99, pinnedPercentile(offerNS, .5), offerP99,
		pinnedPercentile(viewAgeNS, .5), maxViewAge)
	if len(violations) > 0 {
		for i, violation := range violations {
			if i >= 12 {
				break
			}
			t.Error(violation)
		}
		t.Fatalf("loaded Service Recall failed with %d violations", len(violations))
	}
}

func TestResearchPublishedRecallLoadV15(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V15") != "1" {
		t.Skip("opt-in full Service Recall published-LSN load screen")
	}
	raceCorrectnessOnly := os.Getenv("EVENTFRAME_RACE_CORRECTNESS_ONLY") == "1"
	for trial := 1; trial <= 2; trial++ {
		t.Run(fmt.Sprintf("trial-%d", trial), func(t *testing.T) {
			runPublishedRecallLoadV15(t, raceCorrectnessOnly)
		})
	}
}

func TestResearchBatchJournalDrainV15(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_BATCH_JOURNAL_V15") != "1" {
		t.Skip("opt-in native batch-journal Close drain")
	}
	ctx := context.Background()
	root := t.TempDir()
	gate := createIncrementalSortGate(t, root)
	adapter := &publishedRecallLoadStoreV15{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV15, 4), workerEnd: make(chan struct{})}
	if err := adapter.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC).Add(120 * time.Millisecond)
	jobs := make([]*batchJournalJobV15, 4)
	for index := range jobs {
		entry := researchSortJournal(gate, fmt.Sprintf("drain-journal-%d", index), asOf)
		jobs[index] = &batchJournalJobV15{entry: entry, enqueued: time.Now(), done: make(chan error, 1)}
		adapter.journalJobs <- jobs[index]
	}
	go adapter.runJournalWorker()
	if err := adapter.Close(); err != nil {
		t.Fatal("drain batch worker", err)
	}
	if len(adapter.batchSizes) != 1 || adapter.batchSizes[0] != 4 {
		t.Fatal("Close did not flush one four-journal batch", adapter.batchSizes)
	}
	for _, job := range jobs {
		if err := <-job.done; err != nil {
			t.Fatal("queued journal was not acknowledged after commit", err)
		}
	}
	if err := gate.close(); err != nil {
		t.Fatal(err)
	}
	gate, err := openIncrementalSortGate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer gate.close()
	if _, ready := gate.capture(ctx); !ready {
		t.Fatal("drained batch lost READY after reopen")
	}
	for _, job := range jobs {
		if _, err := gate.store.GetBayesianJournal(ctx, job.entry.TenantID, job.entry.ID); err != nil {
			t.Fatal("drained journal missing after reopen", job.entry.ID, err)
		}
	}
}
